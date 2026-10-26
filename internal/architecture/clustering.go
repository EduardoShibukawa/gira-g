package architecture

import (
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// DefaultDomainInvariants returns the foundational invariants for each domain.
func DefaultDomainInvariants() map[DomainID][]string {
	return map[DomainID][]string{
		DomainCore: {
			"Zero-Token Fast Pass: Engine must resolve bounded slices and subgraphs in <5ms without LLM invocation.",
			"Deterministic Projection: Given the same MacroMap and query parameters, slices and tours emit byte-identical representations.",
			"Explicit Unknowns: Unsupported files or missing joins are never silently omitted; they are projected under KnownUnknowns.",
		},
		DomainManipulator: {
			"LLM-Free Surgical Edits: Code symbol retrieval and replacement are strictly syntactic and deterministic.",
			"Safe Brace Balancing: Code blocks are tokenized with string/comment/regex evasion to prevent corrupted AST syntax.",
			"Exact Line Preservation: Surgical replacements do not reformat or alter unedited regions of the file.",
		},
		DomainExtractors: {
			"Three-Layer Hierarchy: Layer 1 (Repo Data) -> Layer 2 (Declarative Configs/Infra) -> Layer 3 (Language ASTs).",
			"Discrete Provenance: Every extracted fact carries an EvidenceRecord with Git commit SHA and line-range locator.",
			"Deterministic Multi-Pass: Extractors run in dependency order and resolve cross-file relations without nondeterminism.",
		},
		DomainEntrypoints: {
			"Dual-Surface Parity: CLI and MCP adapters expose equivalent capabilities over the same core engine data.",
			"Standard IO Separation: Machine-readable JSON and human-friendly Markdown write to stdout; warnings and staleness to stderr.",
			"Fail-Closed Execution: Corrupt index files or uncommitted mutations without authorization exit with nonzero exit codes.",
		},
		DomainGovernance: {
			"Architectural Drift Prevention: Gate checks block circular dependencies, untracked contracts, and deleted symbols.",
			"Finding Accounting: Findings are recorded deterministically with rule IDs and explicit verification criteria.",
			"Objective Scorecard: Metric calculations (discovery, precision, zero-token fast pass) are verified against goldens.",
		},
		DomainStorage: {
			"Append-Only Merkle Integrity: MacroMap source identities are derived via cryptographic SHA256 hashes of input trees.",
			"Cross-Platform Safe Locking: File locks prevent concurrent writer races across Unix and Windows runtimes.",
			"Byte-Identical Serialization: Keys and arrays are sorted prior to JSON encoding to prevent spurious diffs.",
		},
	}
}

// DefaultDomainDescriptions returns descriptions for each architectural domain.
func DefaultDomainDescriptions() map[DomainID]string {
	return map[DomainID]string{
		DomainCore:        "The central graph engine responsible for context slicing, impact calculations, guided tours, and deterministic querying over the MacroMap topology.",
		DomainManipulator: "Syntactic AST symbol query and surgical manipulation engine supporting Go, TypeScript, and Python without external runtimes.",
		DomainExtractors:  "Multi-layer static analysis pipeline parsing manifests, Dockerfiles, compose files, OpenAPI schemas, and source ASTs into canonical nodes and mediated edges.",
		DomainEntrypoints: "Human CLI commands (gira) and agent communication interfaces (Model Context Protocol / stdio JSON-RPC).",
		DomainGovernance:  "Quality gates, drift linting, architectural policy validation, and deterministic scorecard accounting.",
		DomainStorage:     "Cryptographic source-identity hashing, cross-platform locking, and deterministic JSON persistence for architectural maps.",
	}
}

// ClusterMacroMap partitions a domain.MacroMap into architectural communities.
func ClusterMacroMap(mm *domain.MacroMap) *ClusteredArchitecture {
	if mm == nil {
		mm = domain.NewMacroMap("empty")
	}

	inDegree := make(map[string]int)
	outDegree := make(map[string]int)

	for _, rel := range mm.Relationships {
		outDegree[rel.SourceID]++
		inDegree[rel.TargetID]++
	}

	descriptions := DefaultDomainDescriptions()
	invariants := DefaultDomainInvariants()

	domainOrder := []DomainID{
		DomainEntrypoints,
		DomainCore,
		DomainManipulator,
		DomainExtractors,
		DomainGovernance,
		DomainStorage,
		DomainExternal,
	}

	communities := make(map[DomainID]*Community)
	for _, did := range domainOrder {
		communities[did] = &Community{
			ID:               did,
			Title:            string(did),
			Description:      descriptions[did],
			Invariants:       invariants[did],
			CoreAbstractions: make([]ClusteredNode, 0),
			Utilities:        make([]ClusteredNode, 0),
			CrossDomainEdges: make([]DomainRelation, 0),
		}
	}

	nodesByID := make(map[string]ClusteredNode)

	for _, n := range mm.Nodes {
		did := classifyNodeDomain(n)
		inD := inDegree[n.ID]
		outD := outDegree[n.ID]
		centrality := inD*2 + outD

		role := RoleUtilityLeaf
		if inD >= 2 || centrality >= 4 || n.Kind == domain.NodeContract {
			role = RoleCoreAbstraction
		} else if inD >= 1 || outD >= 1 {
			role = RoleMediatorService
		}

		var loc domain.Locator
		if len(n.Evidence) > 0 {
			loc = n.Evidence[0].Locator
		}

		var methods []string
		if n.Metadata != nil {
			if ms, ok := n.Metadata["methods"].([]string); ok {
				methods = ms
			}
		}

		cNode := ClusteredNode{
			ID:          n.ID,
			Name:        n.Name,
			Kind:        n.Kind,
			Domain:      did,
			Role:        role,
			InDegree:    inD,
			OutDegree:   outD,
			Centrality:  centrality,
			Description: n.Description,
			FileLocator: loc,
			Methods:     methods,
			Metadata:    n.Metadata,
		}

		nodesByID[n.ID] = cNode

		if comm, ok := communities[did]; ok {
			if role == RoleCoreAbstraction {
				comm.CoreAbstractions = append(comm.CoreAbstractions, cNode)
			} else {
				comm.Utilities = append(comm.Utilities, cNode)
			}
		}
	}

	// Calculate cross-domain mediated edges
	for _, rel := range mm.Relationships {
		srcNode, srcOk := nodesByID[rel.SourceID]
		tgtNode, tgtOk := nodesByID[rel.TargetID]
		if !srcOk || !tgtOk {
			continue
		}
		if srcNode.Domain != tgtNode.Domain {
			citation := ""
			if len(rel.Evidence) > 0 {
				citation = rel.Evidence[0].Locator.String()
			}
			dr := DomainRelation{
				SourceDomain: srcNode.Domain,
				TargetDomain: tgtNode.Domain,
				Verb:         string(rel.Verb),
				MediatorNode: rel.TargetID,
				Citation:     citation,
			}
			if comm, ok := communities[srcNode.Domain]; ok {
				comm.CrossDomainEdges = append(comm.CrossDomainEdges, dr)
			}
		}
	}

	// Sort nodes deterministically within each community
	for _, comm := range communities {
		sort.Slice(comm.CoreAbstractions, func(i, j int) bool {
			if comm.CoreAbstractions[i].Centrality != comm.CoreAbstractions[j].Centrality {
				return comm.CoreAbstractions[i].Centrality > comm.CoreAbstractions[j].Centrality
			}
			return comm.CoreAbstractions[i].ID < comm.CoreAbstractions[j].ID
		})
		sort.Slice(comm.Utilities, func(i, j int) bool {
			if comm.Utilities[i].Centrality != comm.Utilities[j].Centrality {
				return comm.Utilities[i].Centrality > comm.Utilities[j].Centrality
			}
			return comm.Utilities[i].ID < comm.Utilities[j].ID
		})
	}

	return &ClusteredArchitecture{
		TargetSystem: mm.TargetSystem,
		Communities:  communities,
		NodesByID:    nodesByID,
		TotalNodes:   len(mm.Nodes),
		TotalEdges:   len(mm.Relationships),
		Domains:      domainOrder,
	}
}

// ClusterLocalRepository clusters the scanned packages of a local Go repository.
func ClusterLocalRepository(packages []PackageInfo) *ClusteredArchitecture {
	descriptions := DefaultDomainDescriptions()
	invariants := DefaultDomainInvariants()

	domainOrder := []DomainID{
		DomainEntrypoints,
		DomainCore,
		DomainManipulator,
		DomainExtractors,
		DomainGovernance,
		DomainStorage,
	}

	communities := make(map[DomainID]*Community)
	for _, did := range domainOrder {
		communities[did] = &Community{
			ID:               did,
			Title:            string(did),
			Description:      descriptions[did],
			Invariants:       invariants[did],
			CoreAbstractions: make([]ClusteredNode, 0),
			Utilities:        make([]ClusteredNode, 0),
			CrossDomainEdges: make([]DomainRelation, 0),
		}
	}

	nodesByID := make(map[string]ClusteredNode)

	// Calculate import centrality among packages
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)

	for _, p := range packages {
		outDegree[p.Path] = len(p.Imports)
		for _, imp := range p.Imports {
			inDegree[imp]++
		}
	}

	totalEdges := 0
	for _, p := range packages {
		for _, imp := range p.Imports {
			totalEdges++
			srcDomain := p.Domain
			tgtDomain := ClassifyDomain(imp)
			if srcDomain != tgtDomain {
				if comm, ok := communities[srcDomain]; ok {
					comm.CrossDomainEdges = append(comm.CrossDomainEdges, DomainRelation{
						SourceDomain: srcDomain,
						TargetDomain: tgtDomain,
						Verb:         "imports",
						MediatorNode: imp,
					})
				}
			}
		}

		inD := inDegree[p.Path]
		outD := outDegree[p.Path]
		pkgCentrality := inD*2 + outD

		// For each package, register it and its primary types
		pkgNodeID := "package:" + p.Path
		role := RoleUtilityLeaf
		if inD >= 2 || pkgCentrality >= 4 || p.Path == "internal/domain" || p.Path == "internal/slices" || p.Path == "internal/manipulator" {
			role = RoleCoreAbstraction
		} else if inD >= 1 {
			role = RoleMediatorService
		}

		var firstLoc domain.Locator
		if len(p.Files) > 0 {
			firstLoc = domain.Locator{Path: p.Files[0], StartLine: 1, EndLine: 1}
		}

		var typeNames []string
		for _, t := range p.Types {
			typeNames = append(typeNames, t.Name)
		}

		node := ClusteredNode{
			ID:          pkgNodeID,
			Name:        p.Path,
			Kind:        domain.NodeLibrary,
			Domain:      p.Domain,
			Role:        role,
			InDegree:    inD,
			OutDegree:   outD,
			Centrality:  pkgCentrality,
			Description: p.Doc,
			FileLocator: firstLoc,
			Methods:     typeNames,
		}

		nodesByID[pkgNodeID] = node
		if comm, ok := communities[p.Domain]; ok {
			if role == RoleCoreAbstraction {
				comm.CoreAbstractions = append(comm.CoreAbstractions, node)
			} else {
				comm.Utilities = append(comm.Utilities, node)
			}
		}
	}

	for _, comm := range communities {
		sort.Slice(comm.CoreAbstractions, func(i, j int) bool {
			if comm.CoreAbstractions[i].Centrality != comm.CoreAbstractions[j].Centrality {
				return comm.CoreAbstractions[i].Centrality > comm.CoreAbstractions[j].Centrality
			}
			return comm.CoreAbstractions[i].ID < comm.CoreAbstractions[j].ID
		})
		sort.Slice(comm.Utilities, func(i, j int) bool {
			if comm.Utilities[i].Centrality != comm.Utilities[j].Centrality {
				return comm.Utilities[i].Centrality > comm.Utilities[j].Centrality
			}
			return comm.Utilities[i].ID < comm.Utilities[j].ID
		})
	}

	return &ClusteredArchitecture{
		TargetSystem: "gira-g",
		Communities:  communities,
		NodesByID:    nodesByID,
		TotalNodes:   len(nodesByID),
		TotalEdges:   totalEdges,
		Domains:      domainOrder,
	}
}

func classifyNodeDomain(n domain.Node) DomainID {
	if len(n.Evidence) > 0 {
		path := n.Evidence[0].Locator.Path
		return ClassifyDomain(path)
	}

	id := strings.ToLower(n.ID)
	switch {
	case strings.Contains(id, "mcp") || strings.Contains(id, "cli"):
		return DomainEntrypoints
	case strings.Contains(id, "manipulat") || strings.Contains(id, "ast"):
		return DomainManipulator
	case strings.Contains(id, "extractor") || strings.Contains(id, "proto") || strings.Contains(id, "schema"):
		return DomainExtractors
	case strings.Contains(id, "gate") || strings.Contains(id, "eval") || strings.Contains(id, "scorecard"):
		return DomainGovernance
	case strings.Contains(id, "store") || strings.Contains(id, "cache") || n.Kind == domain.NodeInfrastructure:
		return DomainStorage
	default:
		return DomainCore
	}
}
