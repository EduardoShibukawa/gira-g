package extractors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Pipeline manages extractors ordered by layer and processes target files.
type Pipeline struct {
	extractors []Extractor
	mu         sync.RWMutex
}

// NewPipeline creates a new Pipeline with the specified extractors.
// Extractors are sorted deterministically by Layer and then by Name.
func NewPipeline(extractors ...Extractor) *Pipeline {
	p := &Pipeline{}
	for _, ext := range extractors {
		p.Register(ext)
	}
	return p
}

// DefaultPipeline initializes a Pipeline with Wave 1, Wave 2, and Layer 3 extractors pre-registered.
func DefaultPipeline() *Pipeline {
	return NewPipeline(
		NewDataFormatExtractor(),
		NewDockerfileExtractor(),
		NewComposeExtractor(),
		NewEnvFileExtractor(),
		NewHelmExtractor(),
		NewGoModExtractor(),
		NewGitHubActionsExtractor(),
		NewOpenAPIExtractor(),
		NewMiseExtractor(),
		NewMarkdownExtractor(),
		NewProtobufExtractor(),
		NewGoASTExtractor(),
		NewPythonASTExtractor(),
		NewTypeScriptASTExtractor(),
	)
}

// Register adds an extractor to the pipeline, keeping the list sorted.
func (p *Pipeline) Register(ext Extractor) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.extractors = append(p.extractors, ext)
	sort.Slice(p.extractors, func(i, j int) bool {
		if p.extractors[i].Layer() != p.extractors[j].Layer() {
			return p.extractors[i].Layer() < p.extractors[j].Layer()
		}
		return p.extractors[i].Name() < p.extractors[j].Name()
	})
}

// Extractors returns a snapshot copy of the registered extractors.
func (p *Pipeline) Extractors() []Extractor {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]Extractor, len(p.extractors))
	copy(out, p.extractors)
	return out
}

// ProcessFiles processes the given target files and builds a consolidated domain.MacroMap.
// For any file where no registered extractor returns CanExtract == true, an UnsupportedBoundary
// record is deterministically emitted.
func (p *Pipeline) ProcessFiles(ctx context.Context, targetSystem string, files []TargetFile) (*domain.MacroMap, error) {
	macroMap := domain.NewMacroMap(targetSystem)
	extractors := p.Extractors()

	// Internal deduplication maps
	nodeMap := make(map[string]*domain.Node)
	relMap := make(map[string]*domain.Relationship)
	unsupportedMap := make(map[string]domain.UnsupportedBoundary)
	unsupportedContent := make(map[string][]byte)
	constructPaths := make(map[string]bool)
	var hints []Hint

	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Logical paths are slash-form on every OS: filepath.WalkDir yields
		// backslash paths on Windows, which would otherwise leak into node
		// IDs, locators and golden comparisons. No-op on Linux.
		file.Path = filepath.ToSlash(file.Path)

		var matched []Extractor
		for _, ext := range extractors {
			if ext.CanExtract(file.Path) {
				matched = append(matched, ext)
			}
		}

		if len(matched) == 0 {
			// No extractor registered for file: emit explicit UnsupportedBoundary
			ub := domain.UnsupportedBoundary{
				Path:   file.Path,
				Reason: "no extractor registered for file extension",
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
				},
			}
			unsupportedMap[file.Path] = ub
			unsupportedContent[file.Path] = file.Content
			continue
		}

		// Run all matching extractors (ordered by layer)
		for _, ext := range matched {
			nodes, rels, unsupporteds, err := ext.Extract(ctx, file)
			if err != nil && ext.Layer() == LayerDataFormat {
				// A malformed data file must not abort the whole index: skip it
				// with a warning and an explicit UnsupportedBoundary diagnostic.
				fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", file.Path, err)
				unsupportedMap[file.Path] = domain.UnsupportedBoundary{
					Path:    file.Path,
					Reason:  "malformed data file: " + err.Error(),
					Locator: domain.Locator{Repo: file.Repo, CommitSHA: file.CommitSHA, Path: file.Path},
				}
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("extractor %s failed on file %s: %w", ext.Name(), file.Path, err)
			}

			// Aggregate nodes (merging evidence and metadata on duplicate node ID)
			for _, n := range nodes {
				existing, found := nodeMap[n.ID]
				if !found {
					nCopy := n
					nodeMap[n.ID] = &nCopy
				} else {
					// Merge evidence
					existing.Evidence = append(existing.Evidence, n.Evidence...)
					// Merge metadata
					if existing.Metadata == nil && len(n.Metadata) > 0 {
						existing.Metadata = make(map[string]any)
					}
					for k, v := range n.Metadata {
						if _, has := existing.Metadata[k]; !has {
							existing.Metadata[k] = v
						}
					}
					// Update description if previous was empty
					if existing.Description == "" && n.Description != "" {
						existing.Description = n.Description
					}
				}
			}

			// Aggregate relationships
			for _, r := range rels {
				existing, found := relMap[r.ID]
				if !found {
					rCopy := r
					relMap[r.ID] = &rCopy
				} else {
					existing.Evidence = append(existing.Evidence, r.Evidence...)
				}
			}

			// Aggregate unsupported boundaries. The file was parsed by an
			// Extractor, so each row is an unsupported construct inside a
			// supported file, not an unsupported file (F7).
			for _, u := range unsupporteds {
				unsupportedMap[u.Path] = u
				constructPaths[u.Path] = true
				unsupportedContent[u.Path] = file.Content
			}

			if he, ok := ext.(HintEmitter); ok {
				hs, err := he.Hints(ctx, file)
				if err != nil {
					return nil, fmt.Errorf("extractor %s failed on hints for %s: %w", ext.Name(), file.Path, err)
				}
				hints = append(hints, hs...)
			}
		}
	}

	// Cross-file resolution step (ADR 0013).
	resolve(&resolution{nodes: nodeMap, rels: relMap, hints: hints, modules: goModulesByRepo(files), renamed: map[string]string{}})
	annotateUnknownJoins(unsupportedMap, unsupportedContent, nodeMap)
	markIntraModuleLibraries(nodeMap, files)

	// Transfer nodes to MacroMap in deterministic order
	nodeIDs := make([]string, 0, len(nodeMap))
	for id := range nodeMap {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	for _, id := range nodeIDs {
		n := nodeMap[id]
		if err := macroMap.AddNode(*n); err != nil {
			return nil, fmt.Errorf("failed adding node %s: %w", n.ID, err)
		}
	}

	// Transfer relationships to MacroMap in deterministic order; record unresolved
	// references instead of fabricating phantom dummy nodes when endpoints are missing.
	relIDs := make([]string, 0, len(relMap))
	for id := range relMap {
		relIDs = append(relIDs, id)
	}
	sort.Strings(relIDs)
	for _, id := range relIDs {
		r := relMap[id]
		_, srcOK := nodeMap[r.SourceID]
		_, tgtOK := nodeMap[r.TargetID]
		if !srcOK || !tgtOK {
			evs := make([]domain.EvidenceRecord, len(r.Evidence))
			copy(evs, r.Evidence)
			if !srcOK {
				macroMap.UnresolvedRefs = append(macroMap.UnresolvedRefs, domain.UnresolvedReference{
					ID:         fmt.Sprintf("unresolved:%s:missing:%s", r.ID, r.SourceID),
					RelationID: r.ID,
					SourceID:   r.SourceID,
					TargetID:   r.TargetID,
					MissingID:  r.SourceID,
					Verb:       r.Verb,
					Evidence:   evs,
				})
			}
			if !tgtOK && r.TargetID != r.SourceID {
				evsCopy := make([]domain.EvidenceRecord, len(r.Evidence))
				copy(evsCopy, r.Evidence)
				macroMap.UnresolvedRefs = append(macroMap.UnresolvedRefs, domain.UnresolvedReference{
					ID:         fmt.Sprintf("unresolved:%s:missing:%s", r.ID, r.TargetID),
					RelationID: r.ID,
					SourceID:   r.SourceID,
					TargetID:   r.TargetID,
					MissingID:  r.TargetID,
					Verb:       r.Verb,
					Evidence:   evsCopy,
				})
			}
			continue
		}

		if err := macroMap.AddRelationship(*r); err != nil {
			return nil, fmt.Errorf("failed adding relationship %s: %w", r.ID, err)
		}
	}

	// Transfer unsupported boundaries
	for _, u := range unsupportedMap {
		macroMap.Unsupported = append(macroMap.Unsupported, u)
	}
	for p := range constructPaths {
		macroMap.UnsupportedConstructs = append(macroMap.UnsupportedConstructs, p)
	}
	sort.Strings(macroMap.UnsupportedConstructs)

	return macroMap, nil
}

// markIntraModuleLibraries tags Library nodes whose import path lives inside
// one of the Go modules declared by go.mod files in the same file batch.
// Single-repo fidelity: own packages keep the Library kind (CONTEXT.md) but
// become distinguishable from external dependencies via
// metadata["origin_detail"] = "intra-module". Matching is scoped to the
// repositories cited in each node's own evidence.
func markIntraModuleLibraries(nodeMap map[string]*domain.Node, files []TargetFile) {
	modulesByRepo := make(map[string]map[string]bool)
	for _, f := range files {
		if !strings.HasSuffix(f.Path, "go.mod") {
			continue
		}
		for _, line := range strings.Split(string(f.Content), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "module ") {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) < 2 {
				continue
			}
			if modulesByRepo[f.Repo] == nil {
				modulesByRepo[f.Repo] = make(map[string]bool)
			}
			modulesByRepo[f.Repo][strings.Trim(fields[1], `"`)] = true
			break
		}
	}
	if len(modulesByRepo) == 0 {
		return
	}
	for _, n := range nodeMap {
		if n.Kind != domain.NodeLibrary {
			continue
		}
		importPath, _ := n.Metadata["import_path"].(string)
		if importPath == "" {
			continue
		}
		repos := make(map[string]bool)
		for _, ev := range n.Evidence {
			if ev.Locator.Repo != "" {
				repos[ev.Locator.Repo] = true
			}
		}
		for repo := range repos {
			for module := range modulesByRepo[repo] {
				if importPath == module || strings.HasPrefix(importPath, module+"/") {
					if n.Metadata == nil {
						n.Metadata = make(map[string]any)
					}
					n.Metadata["origin_detail"] = "intra-module"
					break
				}
			}
			if n.Metadata["origin_detail"] == "intra-module" {
				break
			}
		}
	}
}

// OriginDeclaredOverride identifies provenance from declarative YAML overlays.
const OriginDeclaredOverride domain.DerivationOrigin = "declared:override"

// ApplyOverrides applies declarative manifest overrides to the given MacroMap.
// Relationships added or updated receive origin "declared:override" and a Locator pointing to manifestPath.
// Missing endpoints are recorded in mm.UnresolvedRefs (never fabricated as dummy nodes), and
// contradictions with existing extracted facts are recorded in mm.Conflicts.
func ApplyOverrides(mm *domain.MacroMap, overrides *manifest.OverridesDeclaration, manifestPath string) error {
	if mm == nil {
		return fmt.Errorf("macro map cannot be nil")
	}
	if overrides == nil {
		return nil
	}
	hasOverrides := len(overrides.Relationships) > 0 || len(overrides.Metadata) > 0
	if hasOverrides && strings.TrimSpace(manifestPath) == "" {
		return manifest.ErrManifestPathRequired
	}

	now := time.Now().UTC()

	// 1. Apply metadata overrides in deterministic nodeID and key order
	metaNodeIDs := make([]string, 0, len(overrides.Metadata))
	for nodeID := range overrides.Metadata {
		metaNodeIDs = append(metaNodeIDs, nodeID)
	}
	sort.Strings(metaNodeIDs)

	for _, nodeID := range metaNodeIDs {
		meta := overrides.Metadata[nodeID]
		var targetNode *domain.Node
		for i := range mm.Nodes {
			if mm.Nodes[i].ID == nodeID {
				targetNode = &mm.Nodes[i]
				break
			}
		}

		loc := domain.Locator{Path: manifestPath}
		if overrides.MetadataLines != nil {
			if line := overrides.MetadataLines[nodeID]; line > 0 {
				loc.StartLine = line
				loc.EndLine = line
			}
		}
		overrideEv := domain.EvidenceRecord{
			Locator:      loc,
			Origin:       OriginDeclaredOverride,
			RuleID:       "rule:override-metadata",
			Rationale:    "declarative metadata override from manifest",
			Status:       domain.StatusCurrent,
			DiscoveredAt: now,
		}

		if targetNode == nil {
			mm.UnresolvedRefs = append(mm.UnresolvedRefs, domain.UnresolvedReference{
				ID:        fmt.Sprintf("unresolved:override:metadata:%s", nodeID),
				MissingID: nodeID,
				Evidence:  []domain.EvidenceRecord{overrideEv},
			})
			continue
		}

		if targetNode.Metadata == nil {
			targetNode.Metadata = make(map[string]any)
		}
		keys := make([]string, 0, len(meta))
		for k := range meta {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := meta[k]
			if prevVal, exists := targetNode.Metadata[k]; exists && fmt.Sprint(prevVal) != fmt.Sprint(v) {
				conflictEvs := make([]domain.EvidenceRecord, 0, len(targetNode.Evidence)+1)
				conflictEvs = append(conflictEvs, targetNode.Evidence...)
				conflictEvs = append(conflictEvs, overrideEv)
				mm.Conflicts = append(mm.Conflicts, domain.ConflictRecord{
					ID:          fmt.Sprintf("conflict:metadata:%s:%s", nodeID, k),
					SubjectID:   nodeID,
					Kind:        "metadata_value",
					Field:       k,
					DeclaredVal: fmt.Sprint(v),
					PreviousVal: fmt.Sprint(prevVal),
					Evidence:    conflictEvs,
				})
			}
			targetNode.Metadata[k] = v
		}
		targetNode.Evidence = append(targetNode.Evidence, overrideEv)
	}

	// Build node existence index for relationship endpoint verification
	nodeExists := make(map[string]bool, len(mm.Nodes))
	for _, n := range mm.Nodes {
		nodeExists[n.ID] = true
	}

	// 2. Apply relationship overrides
	for _, relOverride := range overrides.Relationships {
		loc := domain.Locator{
			Path:      manifestPath,
			StartLine: relOverride.StartLine,
			EndLine:   relOverride.EndLine,
		}
		overrideEvidence := domain.EvidenceRecord{
			Locator:      loc,
			Origin:       OriginDeclaredOverride,
			RuleID:       "rule:override-relationship",
			Rationale:    "declarative relationship override from manifest",
			Status:       domain.StatusCurrent,
			DiscoveredAt: now,
		}

		relID := fmt.Sprintf("rel:%s-%s-%s", relOverride.SourceID, relOverride.Verb, relOverride.TargetID)
		srcOK := nodeExists[relOverride.SourceID]
		tgtOK := nodeExists[relOverride.TargetID]
		if !srcOK || !tgtOK {
			if !srcOK {
				mm.UnresolvedRefs = append(mm.UnresolvedRefs, domain.UnresolvedReference{
					ID:         fmt.Sprintf("unresolved:%s:missing:%s", relID, relOverride.SourceID),
					RelationID: relID,
					SourceID:   relOverride.SourceID,
					TargetID:   relOverride.TargetID,
					MissingID:  relOverride.SourceID,
					Verb:       domain.RelationVerb(relOverride.Verb),
					Evidence:   []domain.EvidenceRecord{overrideEvidence},
				})
			}
			if !tgtOK && relOverride.TargetID != relOverride.SourceID {
				mm.UnresolvedRefs = append(mm.UnresolvedRefs, domain.UnresolvedReference{
					ID:         fmt.Sprintf("unresolved:%s:missing:%s", relID, relOverride.TargetID),
					RelationID: relID,
					SourceID:   relOverride.SourceID,
					TargetID:   relOverride.TargetID,
					MissingID:  relOverride.TargetID,
					Verb:       domain.RelationVerb(relOverride.Verb),
					Evidence:   []domain.EvidenceRecord{overrideEvidence},
				})
			}
			continue
		}

		// Check if a relationship between source and target already exists
		var matchedRel *domain.Relationship
		for i := range mm.Relationships {
			if mm.Relationships[i].SourceID == relOverride.SourceID && mm.Relationships[i].TargetID == relOverride.TargetID {
				matchedRel = &mm.Relationships[i]
				break
			}
		}

		if matchedRel != nil {
			if matchedRel.Verb != domain.RelationVerb(relOverride.Verb) {
				conflictEvs := make([]domain.EvidenceRecord, 0, len(matchedRel.Evidence)+1)
				conflictEvs = append(conflictEvs, matchedRel.Evidence...)
				conflictEvs = append(conflictEvs, overrideEvidence)
				mm.Conflicts = append(mm.Conflicts, domain.ConflictRecord{
					ID:          fmt.Sprintf("conflict:relationship:%s:%s->%s", matchedRel.ID, matchedRel.Verb, relOverride.Verb),
					SubjectID:   matchedRel.ID,
					Kind:        "relationship_verb",
					Field:       "verb",
					DeclaredVal: relOverride.Verb,
					PreviousVal: string(matchedRel.Verb),
					Evidence:    conflictEvs,
				})
			}
			matchedRel.Verb = domain.RelationVerb(relOverride.Verb)
			matchedRel.Evidence = append(matchedRel.Evidence, overrideEvidence)
		} else {
			newRel := domain.Relationship{
				ID:       relID,
				SourceID: relOverride.SourceID,
				TargetID: relOverride.TargetID,
				Verb:     domain.RelationVerb(relOverride.Verb),
				Evidence: []domain.EvidenceRecord{overrideEvidence},
			}
			if err := mm.AddRelationship(newRel); err != nil {
				return fmt.Errorf("failed to add override relationship %s: %w", relID, err)
			}
		}
	}

	return nil
}

// ApplyOverrides applies declarative manifest overrides to the given MacroMap.
func (p *Pipeline) ApplyOverrides(mm *domain.MacroMap, overrides *manifest.OverridesDeclaration, manifestPath string) error {
	return ApplyOverrides(mm, overrides, manifestPath)
}

