package indexer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// RefreshClass classifies the scope of re-indexing required by a set of file changes.
type RefreshClass string

const (
	RefreshSkip    RefreshClass = "SKIP"
	RefreshPartial RefreshClass = "PARTIAL"
	RefreshArch    RefreshClass = "ARCH"
	RefreshFull    RefreshClass = "FULL"
)

// DefaultExtractorVersion is the version string stamped on fingerprints for Wave-1 extractors.
const DefaultExtractorVersion = "1.0.0"

// ErrSymbolLossGate is returned when an incremental or re-index run loses
// architectural nodes, contracts, or relationships without corresponding file deletion
// and without AllowShrink enabled.
var ErrSymbolLossGate = errors.New("symbol-loss gate blocked unexplained architectural shrinkage")

// FileFingerprint captures both raw content and normalized structural digests for a file.
type FileFingerprint struct {
	Repo             string `json:"repo"`
	Path             string `json:"path"`
	ContentSHA256    string `json:"content_sha256"`
	StructuralHash   string `json:"structural_hash"`
	ExtractorID      string `json:"extractor_id"`
	ExtractorVersion string `json:"extractor_version"`
}

// IncrementalState persists fingerprints and refresh classification across incremental runs.
type IncrementalState struct {
	SchemaVersion string                    `json:"schema_version"`
	Fingerprints  map[string]FileFingerprint `json:"fingerprints"`
	LastClass     RefreshClass              `json:"last_class"`
	PerFileClass  map[string]RefreshClass   `json:"per_file_class,omitempty"`
}

// SymbolLossReport details any nodes, contracts, or relationships present in a
// previous MacroMap that disappeared in the new MacroMap.
type SymbolLossReport struct {
	LostNodes                []string `json:"lost_nodes,omitempty"`
	LostContracts            []string `json:"lost_contracts,omitempty"`
	LostRelationships        []string `json:"lost_relationships,omitempty"`
	RemovedPhantomNodes      []string `json:"removed_phantom_nodes,omitempty"`
	ReclassifiedInvalidRefs  []string `json:"reclassified_invalid_refs,omitempty"`
	UnexplainedNodes         []string `json:"unexplained_nodes,omitempty"`
	UnexplainedContracts     []string `json:"unexplained_contracts,omitempty"`
	UnexplainedRelationships []string `json:"unexplained_relationships,omitempty"`
	Details                  []string `json:"details,omitempty"`
	AllowedByShrink          bool     `json:"allowed_by_shrink,omitempty"`
}

// HasUnexplainedLoss reports whether any lost symbol could not be attributed to a deleted file
// or to the cleanup of legacy phantom nodes / invalid references.
func (r *SymbolLossReport) HasUnexplainedLoss() bool {
	if r == nil {
		return false
	}
	return len(r.UnexplainedNodes) > 0 || len(r.UnexplainedContracts) > 0 || len(r.UnexplainedRelationships) > 0
}

// ComputeFileFingerprint computes the raw SHA-256 and comment/blank-line normalized
// structural hash for a target file.
func ComputeFileFingerprint(file extractors.TargetFile, pipe *extractors.Pipeline) FileFingerprint {
	if pipe == nil {
		pipe = extractors.DefaultPipeline()
	}
	normPath := filepath.ToSlash(file.Path)
	rawSum := sha256.Sum256(file.Content)

	var matchedIDs []string
	for _, ext := range pipe.Extractors() {
		if ext.CanExtract(normPath) {
			matchedIDs = append(matchedIDs, ext.Name())
		}
	}
	sort.Strings(matchedIDs)
	extID := "unsupported"
	if len(matchedIDs) > 0 {
		extID = strings.Join(matchedIDs, ",")
	}

	return FileFingerprint{
		Repo:             file.Repo,
		Path:             normPath,
		ContentSHA256:    hex.EncodeToString(rawSum[:]),
		StructuralHash:   computeStructuralHash(normPath, file.Content),
		ExtractorID:      extID,
		ExtractorVersion: DefaultExtractorVersion,
	}
}

// ComputeFingerprints computes a map of FileFingerprints for a slice of TargetFiles.
func ComputeFingerprints(files []extractors.TargetFile, pipe *extractors.Pipeline) map[string]FileFingerprint {
	out := make(map[string]FileFingerprint, len(files))
	multiRepo := hasMultipleRepos(files)
	for _, f := range files {
		fp := ComputeFileFingerprint(f, pipe)
		key := fp.Path
		if multiRepo && fp.Repo != "" {
			key = fp.Repo + ":" + fp.Path
		}
		out[key] = fp
	}
	return out
}

func hasMultipleRepos(files []extractors.TargetFile) bool {
	var first string
	for _, f := range files {
		if f.Repo == "" {
			continue
		}
		if first == "" {
			first = f.Repo
		} else if f.Repo != first {
			return true
		}
	}
	return false
}

// computeStructuralHash strips comments and blank lines so comment-only edits
// produce the exact same structural digest.
func computeStructuralHash(path string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	var normalized []byte
	switch {
	case ext == ".go":
		normalized = normalizeGoSource(content)
	case ext == ".proto" || base == "go.mod":
		normalized = normalizeSlashComments(content)
	case ext == ".yaml" || ext == ".yml" || ext == ".toml" ||
		base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") || strings.HasSuffix(base, ".dockerfile") ||
		base == ".env":
		normalized = normalizeHashComments(content)
	default:
		normalized = normalizeBlankLines(content)
	}

	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:])
}

func normalizeGoSource(content []byte) []byte {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(content))
	var s scanner.Scanner
	var scanErrors int
	// Mode 0 skips comments; Go's scanner still emits synthetic semicolons accurately.
	s.Init(file, content, func(_ token.Position, _ string) {
		scanErrors++
	}, 0)

	var buf bytes.Buffer
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		buf.WriteString(tok.String())
		if lit != "" {
			buf.WriteByte(':')
			buf.WriteString(lit)
		}
		buf.WriteByte('\n')
	}
	if scanErrors > 0 {
		return normalizeSlashComments(content)
	}
	return buf.Bytes()
}

// normalizeSlashComments strips // line comments and /* */ block comments outside
// string literals, and collapses whitespace and blank lines.
func normalizeSlashComments(content []byte) []byte {
	var out bytes.Buffer
	n := len(content)
	inString := byte(0)
	inBlockComment := false
	inLineComment := false

	for i := 0; i < n; i++ {
		c := content[i]
		if inLineComment {
			if c == '\n' {
				inLineComment = false
				out.WriteByte('\n')
			}
			continue
		}
		if inBlockComment {
			if c == '*' && i+1 < n && content[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inString != 0 {
			out.WriteByte(c)
			if c == '\\' && i+1 < n {
				i++
				out.WriteByte(content[i])
				continue
			}
			if c == inString {
				inString = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inString = c
			out.WriteByte(c)
			continue
		}
		if c == '/' && i+1 < n {
			if content[i+1] == '/' {
				inLineComment = true
				i++
				continue
			}
			if content[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}
		}
		out.WriteByte(c)
	}
	return normalizeBlankLines(out.Bytes())
}

// normalizeHashComments strips # comments outside quotes and removes blank lines.
func normalizeHashComments(content []byte) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		line := stripUnquotedHashComment(sc.Text())
		trimmed := strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		out.WriteString(trimmed)
		out.WriteByte('\n')
	}
	return out.Bytes()
}

func stripUnquotedHashComment(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return ""
	}
	inQuote := byte(0)
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inQuote != 0 {
			if c == '\\' && i+1 < len(line) {
				i++
				continue
			}
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = c
			continue
		}
		if c == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
			return line[:i]
		}
	}
	return line
}

func normalizeBlankLines(content []byte) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		trimmed := strings.TrimSpace(sc.Text())
		if trimmed == "" {
			continue
		}
		out.WriteString(trimmed)
		out.WriteByte('\n')
	}
	return out.Bytes()
}

// isArchOrContractFile reports whether a path is a Layer 2 manifest/contract/config
// file or a Layer 3 contract definition (.proto) whose change triggers RefreshArch.
func isArchOrContractFile(path string) bool {
	norm := filepath.ToSlash(path)
	base := strings.ToLower(filepath.Base(norm))
	ext := strings.ToLower(filepath.Ext(base))

	switch {
	case ext == ".proto":
		return true
	case base == "gira.yaml" || base == "gira.yml":
		return true
	case base == "go.mod":
		return true
	case base == ".env":
		return true
	case base == "docker-compose.yaml" || base == "docker-compose.yml" ||
		base == "compose.yaml" || base == "compose.yml" ||
		strings.HasPrefix(base, "docker-compose.") || strings.HasPrefix(base, "compose."):
		return true
	case base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") || strings.HasSuffix(base, ".dockerfile"):
		return true
	case base == "chart.yaml" || base == "chart.yml" ||
		base == "values.yaml" || base == "values.yml" ||
		strings.HasPrefix(base, "values-") || strings.HasPrefix(base, "values."):
		return true
	case strings.Contains(base, "openapi") || strings.Contains(base, "swagger"):
		return true
	case base == "mise.toml" || base == ".mise.toml":
		return true
	case strings.Contains(norm, ".github/workflows/"):
		return true
	default:
		return false
	}
}

func refreshSeverity(c RefreshClass) int {
	switch c {
	case RefreshFull:
		return 4
	case RefreshArch:
		return 3
	case RefreshPartial:
		return 2
	default:
		return 1
	}
}

// ClassifyRefresh compares previous and current file fingerprints and returns the
// overall RefreshClass along with the per-file RefreshClass breakdown.
//   - SKIP: all files have identical ContentSHA256 or identical StructuralHash.
//   - PARTIAL: only Layer 3 implementation files changed without manifest/contract/compose changes.
//   - ARCH: Layer 2/contract files changed or files were added/removed.
//   - FULL: schema version or extractor version changed, or no previous fingerprint state exists.
func ClassifyRefresh(prevFingerprints, currFingerprints map[string]FileFingerprint) (RefreshClass, map[string]RefreshClass) {
	perFile := make(map[string]RefreshClass, len(currFingerprints))
	if len(prevFingerprints) == 0 {
		for k := range currFingerprints {
			perFile[k] = RefreshFull
		}
		return RefreshFull, perFile
	}

	overall := RefreshSkip
	promote := func(c RefreshClass) {
		if refreshSeverity(c) > refreshSeverity(overall) {
			overall = c
		}
	}

	for k, curr := range currFingerprints {
		prev, exists := prevFingerprints[k]
		if !exists {
			perFile[k] = RefreshArch
			promote(RefreshArch)
			continue
		}
		if (prev.ExtractorVersion != "" && curr.ExtractorVersion != "" && prev.ExtractorVersion != curr.ExtractorVersion) ||
			(prev.ExtractorID != "" && curr.ExtractorID != "" && prev.ExtractorID != curr.ExtractorID) {
			perFile[k] = RefreshFull
			promote(RefreshFull)
			continue
		}
		if (curr.ContentSHA256 != "" && prev.ContentSHA256 == curr.ContentSHA256) ||
			(curr.StructuralHash != "" && prev.StructuralHash == curr.StructuralHash) {
			perFile[k] = RefreshSkip
			continue
		}
		filePath := curr.Path
		if filePath == "" {
			filePath = k
		}
		if isArchOrContractFile(filePath) {
			perFile[k] = RefreshArch
			promote(RefreshArch)
		} else {
			perFile[k] = RefreshPartial
			promote(RefreshPartial)
		}
	}

	for k := range prevFingerprints {
		if _, exists := currFingerprints[k]; !exists {
			perFile[k] = RefreshArch
			promote(RefreshArch)
		}
	}

	return overall, perFile
}

// CheckSymbolLossGate checks whether next lost any Node, Contract, or Relationship
// from prev whose evidence paths were NOT deleted in deletedPaths and that were NOT
// legacy artificial/phantom nodes or invalid references now classified in UnresolvedRefs.
// It fails closed with ErrSymbolLossGate unless every removed real symbol is accounted
// for by a deleted file or allowShrink is true.
func CheckSymbolLossGate(prev, next *domain.MacroMap, deletedPaths map[string]bool, allowShrink bool) (*SymbolLossReport, error) {
	report := &SymbolLossReport{
		AllowedByShrink: allowShrink,
	}
	if prev == nil || next == nil {
		return report, nil
	}

	isDeletedPath := func(loc domain.Locator) bool {
		if len(deletedPaths) == 0 {
			return false
		}
		p := filepath.ToSlash(loc.Path)
		if deletedPaths[p] {
			return true
		}
		if loc.Repo != "" && deletedPaths[loc.Repo+":"+p] {
			return true
		}
		return false
	}

	evidenceDeleted := func(evs []domain.EvidenceRecord) bool {
		if len(evs) == 0 {
			return false
		}
		for _, ev := range evs {
			if !isDeletedPath(ev.Locator) {
				return false
			}
		}
		return true
	}

	isPhantomNode := func(n domain.Node) bool {
		if len(n.Evidence) == 0 {
			return true
		}
		for _, ev := range n.Evidence {
			if ev.Origin != extractors.OriginDeclaredOverride || ev.Locator.Repo != "" {
				return false
			}
		}
		return true
	}

	summarizeEvidence := func(evs []domain.EvidenceRecord) string {
		if len(evs) == 0 {
			return "no-evidence"
		}
		parts := make([]string, 0, len(evs))
		for _, ev := range evs {
			parts = append(parts, fmt.Sprintf("%s(%s)", ev.Locator.String(), ev.Origin))
		}
		sort.Strings(parts)
		return strings.Join(parts, ", ")
	}

	nextNodes := make(map[string]bool, len(next.Nodes))
	for _, n := range next.Nodes {
		nextNodes[n.ID] = true
	}

	nextUnresolvedRels := make(map[string]bool, len(next.UnresolvedRefs))
	nextUnresolvedPairs := make(map[string]bool, len(next.UnresolvedRefs))
	for _, u := range next.UnresolvedRefs {
		if u.RelationID != "" {
			nextUnresolvedRels[u.RelationID] = true
		}
		if u.SourceID != "" && u.TargetID != "" {
			nextUnresolvedPairs[u.SourceID+"\x00"+string(u.Verb)+"\x00"+u.TargetID] = true
		}
	}

	explainedDeletedNodes := make(map[string]bool)
	phantomNodes := make(map[string]bool)

	for _, n := range prev.Nodes {
		if nextNodes[n.ID] {
			continue
		}
		if isPhantomNode(n) {
			report.RemovedPhantomNodes = append(report.RemovedPhantomNodes, n.ID)
			phantomNodes[n.ID] = true
			continue
		}
		if n.Kind == domain.NodeContract {
			report.LostContracts = append(report.LostContracts, n.ID)
			if evidenceDeleted(n.Evidence) {
				explainedDeletedNodes[n.ID] = true
			} else {
				report.UnexplainedContracts = append(report.UnexplainedContracts, n.ID)
				report.Details = append(report.Details, fmt.Sprintf("contract %s (kind=%s, evidence=[%s]) disappeared without source file deletion", n.ID, n.Kind, summarizeEvidence(n.Evidence)))
			}
		} else {
			report.LostNodes = append(report.LostNodes, n.ID)
			if evidenceDeleted(n.Evidence) {
				explainedDeletedNodes[n.ID] = true
			} else {
				report.UnexplainedNodes = append(report.UnexplainedNodes, n.ID)
				report.Details = append(report.Details, fmt.Sprintf("node %s (kind=%s, evidence=[%s]) disappeared without source file deletion", n.ID, n.Kind, summarizeEvidence(n.Evidence)))
			}
		}
	}

	nextRels := make(map[string]bool, len(next.Relationships))
	for _, r := range next.Relationships {
		nextRels[r.ID] = true
	}

	for _, r := range prev.Relationships {
		if nextRels[r.ID] {
			continue
		}
		pairKey := r.SourceID + "\x00" + string(r.Verb) + "\x00" + r.TargetID
		if nextUnresolvedRels[r.ID] || nextUnresolvedPairs[pairKey] || phantomNodes[r.SourceID] || phantomNodes[r.TargetID] {
			report.ReclassifiedInvalidRefs = append(report.ReclassifiedInvalidRefs, r.ID)
			continue
		}
		report.LostRelationships = append(report.LostRelationships, r.ID)
		if evidenceDeleted(r.Evidence) || explainedDeletedNodes[r.SourceID] || explainedDeletedNodes[r.TargetID] {
			continue
		}
		report.UnexplainedRelationships = append(report.UnexplainedRelationships, r.ID)
		report.Details = append(report.Details, fmt.Sprintf("relationship %s (%s -%s-> %s, evidence=[%s]) disappeared without source file deletion", r.ID, r.SourceID, r.Verb, r.TargetID, summarizeEvidence(r.Evidence)))
	}

	sort.Strings(report.LostNodes)
	sort.Strings(report.LostContracts)
	sort.Strings(report.LostRelationships)
	sort.Strings(report.RemovedPhantomNodes)
	sort.Strings(report.ReclassifiedInvalidRefs)
	sort.Strings(report.UnexplainedNodes)
	sort.Strings(report.UnexplainedContracts)
	sort.Strings(report.UnexplainedRelationships)
	sort.Strings(report.Details)

	if report.HasUnexplainedLoss() && !allowShrink {
		return report, fmt.Errorf("%w: unexplained lost nodes=%v contracts=%v relationships=%v; details: %s (pass --allow-symbol-loss to authorize intentional removal)",
			ErrSymbolLossGate, report.UnexplainedNodes, report.UnexplainedContracts, report.UnexplainedRelationships, strings.Join(report.Details, "; "))
	}

	return report, nil
}

// BuildIncremental computes file fingerprints, classifies the refresh scope
// (SKIP, PARTIAL, ARCH, FULL), reuses the prior MacroMap on SKIP, runs the
// symbol-loss gate, and guarantees exact canonical JSON equivalence with Build.
func BuildIncremental(
	ctx context.Context,
	prev *domain.MacroMap,
	prevState *IncrementalState,
	m *manifest.Manifest,
	cm *gitcache.CacheManager,
	opts Options,
) (*domain.MacroMap, *IncrementalState, error) {
	if m == nil {
		return nil, nil, fmt.Errorf("manifest cannot be nil")
	}
	if cm == nil {
		return nil, nil, fmt.Errorf("cache manager cannot be nil")
	}

	var overrideLocator string
	if m.HasDeclarativeOverrides() {
		var err error
		overrideLocator, err = manifest.ResolveManifestLocator(opts.ManifestPath)
		if err != nil {
			return nil, nil, err
		}
	}

	commitTimes := make(map[string]time.Time, len(m.Repositories))
	var latest time.Time
	var targetFiles []extractors.TargetFile

	for _, repo := range m.Repositories {
		if err := cm.EnsureRepo(ctx, repo, opts.Offline); err != nil {
			return nil, nil, fmt.Errorf("failed to ensure repository %q: %w", repo.ID, err)
		}

		ct, err := commitTime(ctx, cm.RepoDir(repo), repo.SHA)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read commit time for repository %q: %w", repo.ID, err)
		}
		commitTimes[repo.ID] = ct
		if ct.After(latest) {
			latest = ct
		}

		entries, err := cm.ListTree(ctx, repo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list tree for repository %q: %w", repo.ID, err)
		}

		for _, entry := range entries {
			content, err := cm.ReadFile(ctx, repo, entry.Path)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read file %q from repository %q: %w", entry.Path, repo.ID, err)
			}

			targetFiles = append(targetFiles, extractors.TargetFile{
				Repo:      repo.ID,
				CommitSHA: repo.SHA,
				Path:      entry.Path,
				Content:   content,
			})
		}
	}

	pipe := opts.Pipeline
	if pipe == nil {
		pipe = extractors.DefaultPipeline()
	}

	currFingerprints := ComputeFingerprints(targetFiles, pipe)

	var overallClass RefreshClass
	var perFileClass map[string]RefreshClass
	if prev == nil || prevState == nil || len(prevState.Fingerprints) == 0 ||
		(prevState.SchemaVersion != "" && prev.SchemaVersion != nil && fmt.Sprint(prevState.SchemaVersion) != fmt.Sprint(prev.SchemaVersion)) {
		overallClass = RefreshFull
		perFileClass = make(map[string]RefreshClass, len(currFingerprints))
		for k := range currFingerprints {
			perFileClass[k] = RefreshFull
		}
	} else {
		overallClass, perFileClass = ClassifyRefresh(prevState.Fingerprints, currFingerprints)
	}

	deletedPaths := make(map[string]bool)
	if prevState != nil {
		for k, fp := range prevState.Fingerprints {
			if _, exists := currFingerprints[k]; !exists {
				deletedPaths[k] = true
				if fp.Path != "" {
					deletedPaths[filepath.ToSlash(fp.Path)] = true
				}
				if fp.Repo != "" && fp.Path != "" {
					deletedPaths[fp.Repo+":"+filepath.ToSlash(fp.Path)] = true
				}
			}
		}
	}

	var next *domain.MacroMap
	if overallClass == RefreshSkip && prev != nil {
		cloned, err := cloneMacroMap(prev)
		if err != nil {
			return nil, nil, err
		}
		next = cloned
	} else {
		mm, err := pipe.ProcessFiles(ctx, m.TargetSystem.Name, targetFiles)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract files: %w", err)
		}
		if m.Overrides != nil {
			if err := pipe.ApplyOverrides(mm, m.Overrides, overrideLocator); err != nil {
				return nil, nil, fmt.Errorf("failed to apply overrides: %w", err)
			}
		}
		StampEvidence(mm, commitTimes, latest)
		next = mm
	}

	if prev != nil {
		if _, err := CheckSymbolLossGate(prev, next, deletedPaths, opts.AllowShrink); err != nil {
			return nil, nil, err
		}
	}

	nextState := &IncrementalState{
		SchemaVersion: fmt.Sprint(next.SchemaVersion),
		Fingerprints:  currFingerprints,
		LastClass:     overallClass,
		PerFileClass:  perFileClass,
	}
	return next, nextState, nil
}

func cloneMacroMap(mm *domain.MacroMap) (*domain.MacroMap, error) {
	if mm == nil {
		return nil, nil
	}
	out := &domain.MacroMap{
		SchemaVersion: mm.SchemaVersion,
		TargetSystem:  mm.TargetSystem,
		Nodes:         make([]domain.Node, len(mm.Nodes)),
		Relationships: make([]domain.Relationship, len(mm.Relationships)),
		Unsupported:   make([]domain.UnsupportedBoundary, len(mm.Unsupported)),
	}
	for i, n := range mm.Nodes {
		nCopy := n
		if len(n.Evidence) > 0 {
			nCopy.Evidence = append([]domain.EvidenceRecord(nil), n.Evidence...)
		}
		if n.Metadata != nil {
			nCopy.Metadata = make(map[string]any, len(n.Metadata))
			for k, v := range n.Metadata {
				nCopy.Metadata[k] = v
			}
		}
		out.Nodes[i] = nCopy
	}
	for i, r := range mm.Relationships {
		rCopy := r
		if len(r.Evidence) > 0 {
			rCopy.Evidence = append([]domain.EvidenceRecord(nil), r.Evidence...)
		}
		if r.Metadata != nil {
			rCopy.Metadata = make(map[string]any, len(r.Metadata))
			for k, v := range r.Metadata {
				rCopy.Metadata[k] = v
			}
		}
		out.Relationships[i] = rCopy
	}
	for i, u := range mm.Unsupported {
		uCopy := u
		if len(u.UnknownJoins) > 0 {
			uCopy.UnknownJoins = append([]string(nil), u.UnknownJoins...)
		}
		out.Unsupported[i] = uCopy
	}
	if len(mm.UnsupportedConstructs) > 0 {
		out.UnsupportedConstructs = append([]string(nil), mm.UnsupportedConstructs...)
	}
	if len(mm.UnresolvedRefs) > 0 {
		out.UnresolvedRefs = make([]domain.UnresolvedReference, len(mm.UnresolvedRefs))
		for i, unresolved := range mm.UnresolvedRefs {
			unresolvedCopy := unresolved
			if len(unresolved.Evidence) > 0 {
				unresolvedCopy.Evidence = append([]domain.EvidenceRecord(nil), unresolved.Evidence...)
			}
			out.UnresolvedRefs[i] = unresolvedCopy
		}
	}
	if len(mm.Conflicts) > 0 {
		out.Conflicts = make([]domain.ConflictRecord, len(mm.Conflicts))
		for i, conflict := range mm.Conflicts {
			conflictCopy := conflict
			if len(conflict.Evidence) > 0 {
				conflictCopy.Evidence = append([]domain.EvidenceRecord(nil), conflict.Evidence...)
			}
			out.Conflicts[i] = conflictCopy
		}
	}
	if mm.SourceIdentity != nil {
		siCopy := &domain.SourceIdentity{
			SchemaVersion:  mm.SourceIdentity.SchemaVersion,
			ManifestSHA256: mm.SourceIdentity.ManifestSHA256,
		}
		if len(mm.SourceIdentity.Extractors) > 0 {
			siCopy.Extractors = append([]domain.ExtractorIdentity(nil), mm.SourceIdentity.Extractors...)
		}
		if len(mm.SourceIdentity.Sources) > 0 {
			siCopy.Sources = append([]domain.SourceRepoIdentity(nil), mm.SourceIdentity.Sources...)
		}
		out.SourceIdentity = siCopy
	}
	return out, nil
}
