package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// NodeType represents the concrete type of an architectural unit.
type NodeType string

const (
	NodeService        NodeType = "Service"
	NodeLibrary        NodeType = "Library"
	NodeInfrastructure NodeType = "Infrastructure"
	NodeContract       NodeType = "Contract"
	NodePolicy         NodeType = "Policy"
	NodeSkill          NodeType = "Skill"
	NodeTask           NodeType = "Task"
)

// RelationVerb defines the closed set of directional relationship verbs.
type RelationVerb string

const (
	VerbProvides   RelationVerb = "provides"
	VerbConsumes   RelationVerb = "consumes"
	VerbImports    RelationVerb = "imports"
	VerbAccesses   RelationVerb = "accesses"
	VerbImplements RelationVerb = "implements"
	VerbVerifies   RelationVerb = "verifies"
	VerbGoverns    RelationVerb = "governs"
)

// DerivationOrigin designates the provenance classification of a discovered fact.
type DerivationOrigin string

const (
	OriginDeclared  DerivationOrigin = "declared"
	OriginExtracted DerivationOrigin = "extracted"
	OriginInferred  DerivationOrigin = "inferred"
)

// FreshnessStatus indicates whether evidence aligns with current repository HEAD.
type FreshnessStatus string

const (
	StatusCurrent    FreshnessStatus = "current"
	StatusStale      FreshnessStatus = "stale"
	StatusUnverified FreshnessStatus = "unverified"
)

// Locator identifies the exact file location and line range within a repository commit.
type Locator struct {
	Repo      string `json:"repo"`
	CommitSHA string `json:"commit_sha"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

// String returns a canonical locator string: repo@commit:path#Lstart-Lend
func (l Locator) String() string {
	base := fmt.Sprintf("%s@%s:%s", l.Repo, l.CommitSHA, l.Path)
	if l.StartLine > 0 && l.EndLine > 0 {
		return fmt.Sprintf("%s#L%d-L%d", base, l.StartLine, l.EndLine)
	} else if l.StartLine > 0 {
		return fmt.Sprintf("%s#L%d", base, l.StartLine)
	}
	return base
}

// EvidenceRecord documents the provenance, origin, and deployment context of a fact.
type EvidenceRecord struct {
	Locator           Locator          `json:"locator"`
	Origin            DerivationOrigin `json:"origin"`
	RuleID            string           `json:"rule_id,omitempty"`
	Rationale         string           `json:"rationale,omitempty"`
	DeploymentProfile string           `json:"deployment_profile,omitempty"`
	Status            FreshnessStatus  `json:"status"`
	// DiscoveredAt is the commit time of the repository the locator cites (or the
	// latest commit time of the Target System for declarative overrides), not the
	// wall-clock time of indexing, so the same commits always index to the same
	// bytes (spec gira-agent-skills-and-eval SC4; owner exception, 2026-10-13).
	DiscoveredAt time.Time `json:"discovered_at"`
}

// Node represents an architectural entity in the Macro Map.
type Node struct {
	ID          string           `json:"id"`
	Kind        NodeType         `json:"kind"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
	Evidence    []EvidenceRecord `json:"evidence"`
}

// Relationship represents a directional, evidence-backed connection between two nodes.
type Relationship struct {
	ID       string       `json:"id"`
	SourceID string       `json:"source_id"`
	TargetID string       `json:"target_id"`
	Verb     RelationVerb `json:"verb"`
	// Metadata qualifies the edge, e.g. protocol "http" and the routes called.
	Metadata map[string]any   `json:"metadata,omitempty"`
	Evidence []EvidenceRecord `json:"evidence"`
}

// UnsupportedBoundary explicitly records unhandled languages, files, or constructs.
type UnsupportedBoundary struct {
	Path    string  `json:"path"`
	Reason  string  `json:"reason"`
	Locator Locator `json:"locator"`
	// UnknownJoins lists the Contracts the unsupported code names. Whether it
	// provides or consumes them is unknown: no edge is ever drawn from it.
	UnknownJoins []string `json:"unknown_joins,omitempty"`
}

// UnresolvedReference records a relationship or override that cites a non-existent
// node ID, preserving provenance without fabricating a phantom node in MacroMap.Nodes.
type UnresolvedReference struct {
	ID         string           `json:"id"`
	RelationID string           `json:"relation_id,omitempty"`
	SourceID   string           `json:"source_id,omitempty"`
	TargetID   string           `json:"target_id,omitempty"`
	MissingID  string           `json:"missing_id"`
	Verb       RelationVerb     `json:"verb,omitempty"`
	Evidence   []EvidenceRecord `json:"evidence"`
}

// ConflictRecord documents a deterministic conflict between declared/override facts
// and extracted/inferred facts for a node or relationship.
type ConflictRecord struct {
	ID          string           `json:"id"`
	SubjectID   string           `json:"subject_id"`
	Kind        string           `json:"kind"`
	Field       string           `json:"field,omitempty"`
	DeclaredVal string           `json:"declared_value"`
	PreviousVal string           `json:"previous_value"`
	Evidence    []EvidenceRecord `json:"evidence"`
}

// ExtractorIdentity captures the identifier and SemVer version of an architectural extractor.
type ExtractorIdentity struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// ExtractorList is a slice of ExtractorIdentity supporting flexible unmarshaling
// from both JSON arrays of objects and JSON key-value maps.
type ExtractorList []ExtractorIdentity

func (el *ExtractorList) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		*el = nil
		return nil
	}
	if trimmed[0] == '[' {
		var list []ExtractorIdentity
		if err := json.Unmarshal(trimmed, &list); err != nil {
			return err
		}
		*el = list
		return nil
	}
	if trimmed[0] == '{' {
		var m map[string]string
		if err := json.Unmarshal(trimmed, &m); err != nil {
			return err
		}
		list := make([]ExtractorIdentity, 0, len(m))
		for k, v := range m {
			list = append(list, ExtractorIdentity{ID: k, Version: v})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].ID < list[j].ID
		})
		*el = list
		return nil
	}
	return fmt.Errorf("invalid extractors JSON: expected array or object")
}

// SourceRepoIdentity captures the cryptographic provenance of a source repository
// declared in the manifest, without leaking raw clone URLs.
type SourceRepoIdentity struct {
	RepoID    string `json:"repo_id"`
	CommitSHA string `json:"commit_sha,omitempty"`
	SHA       string `json:"sha,omitempty"`
	PathScope string `json:"path_scope,omitempty"`
	URLSHA256 string `json:"url_sha256"`
}

func (s *SourceRepoIdentity) UnmarshalJSON(data []byte) error {
	type Alias SourceRepoIdentity
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*s = SourceRepoIdentity(aux)
	if s.CommitSHA == "" && s.SHA != "" {
		s.CommitSHA = s.SHA
	}
	if s.SHA == "" && s.CommitSHA != "" {
		s.SHA = s.CommitSHA
	}
	return nil
}

// SourceIdentity defines the self-contained cryptographic provenance metadata
// embedded within a Schema 2 MacroMap index or detached snapshot.
type SourceIdentity struct {
	SchemaVersion  int                  `json:"schema_version"`
	ManifestSHA256 string               `json:"manifest_sha256"`
	Extractors     ExtractorList        `json:"extractors"`
	Sources        []SourceRepoIdentity `json:"sources"`
}

func (si *SourceIdentity) ExtractorsMap() map[string]string {
	if si == nil || len(si.Extractors) == 0 {
		return nil
	}
	m := make(map[string]string, len(si.Extractors))
	for _, ext := range si.Extractors {
		m[ext.ID] = ext.Version
	}
	return m
}

// MacroMap represents the complete architectural graph for a Target System.
type MacroMap struct {
	SchemaVersion         any                   `json:"schema_version"`
	TargetSystem          string                `json:"target_system"`
	Nodes                 []Node                `json:"nodes"`
	Relationships         []Relationship        `json:"relationships"`
	Unsupported           []UnsupportedBoundary `json:"unsupported,omitempty"`
	// UnsupportedConstructs lists the paths whose Unsupported row is an
	// unsupported construct inside a file an Extractor did parse, as opposed
	// to a file no Extractor parses. The row stays in Unsupported (briefings
	// keep the explicit unknown), but it is not an unsupported file.
	UnsupportedConstructs []string              `json:"unsupported_constructs,omitempty"`
	UnresolvedRefs        []UnresolvedReference `json:"unresolved_refs,omitempty"`
	Conflicts             []ConflictRecord      `json:"conflicts,omitempty"`
	SourceIdentity        *SourceIdentity       `json:"source_identity,omitempty"`
}

func (m *MacroMap) UnmarshalJSON(data []byte) error {
	type Alias MacroMap
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&aux); err != nil {
		return err
	}
	if num, ok := m.SchemaVersion.(json.Number); ok {
		if intVal, err := num.Int64(); err == nil {
			m.SchemaVersion = int(intVal)
		} else {
			m.SchemaVersion = num.String()
		}
	} else if f, ok := m.SchemaVersion.(float64); ok && f == float64(int(f)) {
		m.SchemaVersion = int(f)
	}
	return nil
}

// IsUnsupportedConstruct reports whether the Unsupported row at path is an
// unsupported construct inside a parsed file rather than an unsupported file.
func (m *MacroMap) IsUnsupportedConstruct(path string) bool {
	for _, p := range m.UnsupportedConstructs {
		if p == path {
			return true
		}
	}
	return false
}

// UnsupportedFiles returns the Unsupported rows that are whole files no
// Extractor parses: the Unsupported Boundaries proper.
func (m *MacroMap) UnsupportedFiles() []UnsupportedBoundary {
	out := make([]UnsupportedBoundary, 0, len(m.Unsupported))
	for _, ub := range m.Unsupported {
		if !m.IsUnsupportedConstruct(ub.Path) {
			out = append(out, ub)
		}
	}
	return out
}

// UnsupportedConstructRows returns the Unsupported rows that are constructs
// inside parsed files.
func (m *MacroMap) UnsupportedConstructRows() []UnsupportedBoundary {
	out := make([]UnsupportedBoundary, 0)
	for _, ub := range m.Unsupported {
		if m.IsUnsupportedConstruct(ub.Path) {
			out = append(out, ub)
		}
	}
	return out
}

// NewMacroMap initializes an empty MacroMap with canonical schema version.
func NewMacroMap(targetSystem string) *MacroMap {
	return &MacroMap{
		SchemaVersion: "1.0.0",
		TargetSystem:  targetSystem,
		Nodes:         make([]Node, 0),
		Relationships: make([]Relationship, 0),
		Unsupported:   make([]UnsupportedBoundary, 0),
	}
}

// NewMacroMapV2 initializes an empty Schema 2 MacroMap with embedded SourceIdentity.
func NewMacroMapV2(targetSystem string, si *SourceIdentity) *MacroMap {
	return &MacroMap{
		SchemaVersion:  2,
		TargetSystem:   targetSystem,
		Nodes:          make([]Node, 0),
		Relationships:  make([]Relationship, 0),
		Unsupported:    make([]UnsupportedBoundary, 0),
		SourceIdentity: si,
	}
}

// Validate checks structural invariants of the MacroMap so corrupted or incomplete
// index files cannot be loaded or published as valid state.
func (m *MacroMap) Validate() error {
	if m == nil {
		return fmt.Errorf("macro map is nil")
	}
	if m.SchemaVersion == nil {
		return fmt.Errorf("macro map missing schema_version")
	}
	switch v := m.SchemaVersion.(type) {
	case int:
		if v <= 0 {
			return fmt.Errorf("macro map invalid integer schema_version %d", v)
		}
	case float64:
		if v != float64(int(v)) || v <= 0 {
			return fmt.Errorf("macro map invalid numeric schema_version %v", v)
		}
	case json.Number:
		n, err := v.Int64()
		if err != nil || n <= 0 {
			return fmt.Errorf("macro map invalid json.Number schema_version %v", v)
		}
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("macro map missing schema_version")
		}
	default:
		return fmt.Errorf("macro map unsupported schema_version type %T", m.SchemaVersion)
	}
	if m.TargetSystem == "" {
		return fmt.Errorf("macro map missing target_system")
	}
	if m.SourceIdentity != nil {
		if m.SourceIdentity.SchemaVersion != 2 {
			return fmt.Errorf("source_identity schema_version must be 2, got %d", m.SourceIdentity.SchemaVersion)
		}
		if m.SourceIdentity.ManifestSHA256 == "" {
			return fmt.Errorf("source_identity missing manifest_sha256")
		}
		for _, s := range m.SourceIdentity.Sources {
			if s.RepoID == "" {
				return fmt.Errorf("source_identity source has empty repo_id")
			}
			if s.URLSHA256 == "" {
				return fmt.Errorf("source_identity source %q has empty url_sha256", s.RepoID)
			}
		}
	}
	nodeIDs := make(map[string]bool, len(m.Nodes))
	for _, n := range m.Nodes {
		if n.ID == "" {
			return fmt.Errorf("macro map contains node with empty id")
		}
		if n.Kind == "" {
			return fmt.Errorf("macro map node %q missing kind", n.ID)
		}
		if nodeIDs[n.ID] {
			return fmt.Errorf("macro map contains duplicate node id %q", n.ID)
		}
		nodeIDs[n.ID] = true
	}
	relIDs := make(map[string]bool, len(m.Relationships))
	for _, r := range m.Relationships {
		if r.ID == "" {
			return fmt.Errorf("macro map contains relationship with empty id")
		}
		if r.SourceID == "" || r.TargetID == "" || r.Verb == "" {
			return fmt.Errorf("macro map relationship %q has incomplete endpoints or verb", r.ID)
		}
		if !nodeIDs[r.SourceID] {
			return fmt.Errorf("macro map relationship %q cites missing source node %q", r.ID, r.SourceID)
		}
		if !nodeIDs[r.TargetID] {
			return fmt.Errorf("macro map relationship %q cites missing target node %q", r.ID, r.TargetID)
		}
		if relIDs[r.ID] {
			return fmt.Errorf("macro map contains duplicate relationship id %q", r.ID)
		}
		relIDs[r.ID] = true
	}
	return nil
}

// AddNode inserts a node into the MacroMap, ensuring IDs remain unique.
func (m *MacroMap) AddNode(node Node) error {
	for _, n := range m.Nodes {
		if n.ID == node.ID {
			return fmt.Errorf("duplicate node id: %s", node.ID)
		}
	}
	m.Nodes = append(m.Nodes, node)
	return nil
}

// AddRelationship inserts a relationship into the MacroMap.
func (m *MacroMap) AddRelationship(rel Relationship) error {
	var srcFound, tgtFound bool
	for _, n := range m.Nodes {
		if n.ID == rel.SourceID {
			srcFound = true
		}
		if n.ID == rel.TargetID {
			tgtFound = true
		}
	}
	if !srcFound {
		return fmt.Errorf("relationship source not found: %s", rel.SourceID)
	}
	if !tgtFound {
		return fmt.Errorf("relationship target not found: %s", rel.TargetID)
	}
	m.Relationships = append(m.Relationships, rel)
	return nil
}

// CanonicalJSON serializes the MacroMap to deterministic, formatted JSON with sorted keys.
func (m *MacroMap) CanonicalJSON() ([]byte, error) {
	// Sort nodes deterministically by ID
	sortedNodes := make([]Node, len(m.Nodes))
	copy(sortedNodes, m.Nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].ID < sortedNodes[j].ID
	})

	// Sort relationships deterministically by ID
	sortedRels := make([]Relationship, len(m.Relationships))
	copy(sortedRels, m.Relationships)
	sort.Slice(sortedRels, func(i, j int) bool {
		return sortedRels[i].ID < sortedRels[j].ID
	})

	// Sort unsupported entries deterministically by path
	sortedUnsupported := make([]UnsupportedBoundary, len(m.Unsupported))
	copy(sortedUnsupported, m.Unsupported)
	sort.Slice(sortedUnsupported, func(i, j int) bool {
		return sortedUnsupported[i].Path < sortedUnsupported[j].Path
	})

	normalized := MacroMap{
		SchemaVersion: m.SchemaVersion,
		TargetSystem:  m.TargetSystem,
		Nodes:         sortedNodes,
		Relationships: sortedRels,
		Unsupported:   sortedUnsupported,
	}
	if f, ok := normalized.SchemaVersion.(float64); ok && f == float64(int(f)) {
		normalized.SchemaVersion = int(f)
	}
	if len(m.UnsupportedConstructs) > 0 {
		constructs := append([]string(nil), m.UnsupportedConstructs...)
		sort.Strings(constructs)
		normalized.UnsupportedConstructs = constructs
	}
	if len(m.UnresolvedRefs) > 0 {
		unresolved := make([]UnresolvedReference, len(m.UnresolvedRefs))
		copy(unresolved, m.UnresolvedRefs)
		sort.Slice(unresolved, func(i, j int) bool {
			return unresolved[i].ID < unresolved[j].ID
		})
		normalized.UnresolvedRefs = unresolved
	}
	if len(m.Conflicts) > 0 {
		conflicts := make([]ConflictRecord, len(m.Conflicts))
		copy(conflicts, m.Conflicts)
		sort.Slice(conflicts, func(i, j int) bool {
			return conflicts[i].ID < conflicts[j].ID
		})
		normalized.Conflicts = conflicts
	}
	if m.SourceIdentity != nil {
		siCopy := &SourceIdentity{
			SchemaVersion:  m.SourceIdentity.SchemaVersion,
			ManifestSHA256: m.SourceIdentity.ManifestSHA256,
		}
		if len(m.SourceIdentity.Extractors) > 0 {
			exts := make([]ExtractorIdentity, len(m.SourceIdentity.Extractors))
			copy(exts, m.SourceIdentity.Extractors)
			sort.Slice(exts, func(i, j int) bool {
				return exts[i].ID < exts[j].ID
			})
			siCopy.Extractors = exts
		}
		if len(m.SourceIdentity.Sources) > 0 {
			sources := make([]SourceRepoIdentity, len(m.SourceIdentity.Sources))
			copy(sources, m.SourceIdentity.Sources)
			sort.Slice(sources, func(i, j int) bool {
				return sources[i].RepoID < sources[j].RepoID
			})
			siCopy.Sources = sources
		}
		normalized.SourceIdentity = siCopy
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

