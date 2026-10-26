package architecture

import (
	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// DomainID identifies an architectural domain or community in Gira-G.
type DomainID string

const (
	DomainCore        DomainID = "Core Engine"
	DomainManipulator DomainID = "AST Manipulator"
	DomainExtractors  DomainID = "Extractors Pipeline"
	DomainEntrypoints DomainID = "CLI & MCP Entrypoints"
	DomainGovernance  DomainID = "Governance & Gate"
	DomainStorage     DomainID = "Storage & Persistence"
	DomainExternal    DomainID = "External & Infrastructure"
)

// NodeRole classifies an entity by its centrality and structural significance.
type NodeRole string

const (
	RoleCoreAbstraction NodeRole = "Core Abstraction"
	RoleMediatorService NodeRole = "Mediator / Service"
	RoleUtilityLeaf     NodeRole = "Utility / Leaf"
)

// ClusteredNode represents a node enriched with architectural domain and centrality metrics.
type ClusteredNode struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Kind        domain.NodeType `json:"kind"`
	Domain      DomainID        `json:"domain"`
	Role        NodeRole        `json:"role"`
	InDegree    int             `json:"in_degree"`
	OutDegree   int             `json:"out_degree"`
	Centrality  int             `json:"centrality"`
	Description string          `json:"description,omitempty"`
	FileLocator domain.Locator  `json:"file_locator,omitempty"`
	Methods     []string        `json:"methods,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// DomainRelation captures cross-domain mediation or dependencies.
type DomainRelation struct {
	SourceDomain DomainID `json:"source_domain"`
	TargetDomain DomainID `json:"target_domain"`
	Verb         string   `json:"verb"`
	MediatorNode string   `json:"mediator_node"`
	Citation     string   `json:"citation,omitempty"`
}

// Community represents a cohesive architectural cluster of components and contracts.
type Community struct {
	ID               DomainID         `json:"id"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Invariants       []string         `json:"invariants"`
	CoreAbstractions []ClusteredNode  `json:"core_abstractions"`
	Utilities        []ClusteredNode  `json:"utilities"`
	CrossDomainEdges []DomainRelation `json:"cross_domain_edges"`
}

// ClusteredArchitecture represents the complete clustered topology of a system.
type ClusteredArchitecture struct {
	TargetSystem string                   `json:"target_system"`
	Communities  map[DomainID]*Community  `json:"communities"`
	NodesByID    map[string]ClusteredNode `json:"nodes_by_id"`
	TotalNodes   int                      `json:"total_nodes"`
	TotalEdges   int                      `json:"total_edges"`
	Domains      []DomainID               `json:"domains"`
}

// SymbolDoc represents semantic documentation for an individual code symbol.
type SymbolDoc struct {
	Symbol     string         `json:"symbol"`
	Container  string         `json:"container,omitempty"`
	Kind       string         `json:"kind"`
	Signature  string         `json:"signature"`
	DocComment string         `json:"doc_comment,omitempty"`
	Package    string         `json:"package"`
	Domain     DomainID       `json:"domain"`
	Locator    domain.Locator `json:"locator"`
	Code       string         `json:"code,omitempty"`
}

// ModuleDoc represents semantic documentation for an architectural module or package.
type ModuleDoc struct {
	Package          string          `json:"package"`
	Path             string          `json:"path"`
	Domain           DomainID        `json:"domain"`
	Description      string          `json:"description"`
	CoreAbstractions []ClusteredNode `json:"core_abstractions"`
	Types            []string        `json:"types"`
	Functions        []string        `json:"functions"`
	Imports          []string        `json:"imports"`
	DependentModules []string        `json:"dependent_modules"`
}

// LifecycleStep represents one discrete stage in an architectural execution flow.
type LifecycleStep struct {
	Order       int    `json:"order"`
	Component   string `json:"component"`
	Domain      string `json:"domain"`
	Action      string `json:"action"`
	Artifact    string `json:"artifact,omitempty"`
	CodeLocator string `json:"code_locator,omitempty"`
}

// LifecycleFlow documents an end-to-end execution path across architectural layers.
type LifecycleFlow struct {
	Intent      string          `json:"intent"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Steps       []LifecycleStep `json:"steps"`
	MermaidSeq  string          `json:"mermaid_seq,omitempty"`
}
