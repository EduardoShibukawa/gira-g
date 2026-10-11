package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// NodeType represents the concrete type of an architectural unit.
type NodeType string

const (
	NodeService        NodeType = "Service"
	NodeLibrary        NodeType = "Library"
	NodeInfrastructure NodeType = "Infrastructure"
	NodeContract       NodeType = "Contract"
)

// RelationVerb defines the closed set of directional relationship verbs.
type RelationVerb string

const (
	VerbProvides RelationVerb = "provides"
	VerbConsumes RelationVerb = "consumes"
	VerbImports  RelationVerb = "imports"
	VerbAccesses RelationVerb = "accesses"
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
	DeploymentProfile string           `json:"deployment_profile,omitempty"`
	Status            FreshnessStatus  `json:"status"`
	DiscoveredAt      time.Time        `json:"discovered_at"`
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
	ID       string           `json:"id"`
	SourceID string           `json:"source_id"`
	TargetID string           `json:"target_id"`
	Verb     RelationVerb     `json:"verb"`
	Evidence []EvidenceRecord `json:"evidence"`
}

// UnsupportedBoundary explicitly records unhandled languages, files, or constructs.
type UnsupportedBoundary struct {
	Path    string  `json:"path"`
	Reason  string  `json:"reason"`
	Locator Locator `json:"locator"`
}

// MacroMap represents the complete architectural graph for a Target System.
type MacroMap struct {
	SchemaVersion string                `json:"schema_version"`
	TargetSystem  string                `json:"target_system"`
	Nodes         []Node                `json:"nodes"`
	Relationships []Relationship        `json:"relationships"`
	Unsupported   []UnsupportedBoundary `json:"unsupported,omitempty"`
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

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
