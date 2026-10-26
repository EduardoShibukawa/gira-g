package architecture

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// GenerateArchitectureMarkdown generates a high-density, Spring Framework Reference Overview
// standard architecture document for Gira-G or any analyzed target system.
func GenerateArchitectureMarkdown(repoRoot string, arch *ClusteredArchitecture, pkgs []PackageInfo) string {
	var sb strings.Builder

	systemName := "Gira-G Architecture Harness"
	if arch != nil && arch.TargetSystem != "" && arch.TargetSystem != "gira-g" && arch.TargetSystem != "empty" {
		systemName = arch.TargetSystem
	}

	// 1. Title & Executive Overview
	sb.WriteString(fmt.Sprintf("# %s — Architectural Reference Overview\n\n", systemName))
	sb.WriteString("> **Living Architecture Standard**: Deterministic, self-generating architecture and semantic engine reference. Designed for cognitive density, machine readability by AI agents, and complete navigational parity for software engineers.\n\n")

	sb.WriteString("## 1. Executive Overview & Core Philosophy\n\n")
	sb.WriteString("Gira-G is an architectural intelligence harness and deterministic semantic engine. It provides high-speed, structural codebase comprehension, multi-layer graph extraction, surgical AST editing, and governance gates without relying on multi-gigabyte language servers or heavy JavaScript runtimes.\n\n")

	sb.WriteString("### 1.1 Fundamental Design Invariants\n\n")
	sb.WriteString("| Invariant | Architectural Guarantee | Enforcement Mechanism |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	sb.WriteString("| **Zero-Token Fast Pass** | Structural queries (impact, slices, docs, flows) resolve in <5ms without LLM token expenditure. | Pre-indexed in-memory graph traversals (`internal/slices:Engine`). |\n")
	sb.WriteString("| **Deterministic AST Manipulation** | Code symbol reading, writing, and member addition preserve surrounding formatting and indentations. | Native AST parsers + `SafeBraceCounter` token balancing (`internal/manipulator`). |\n")
	sb.WriteString("| **Append-Only Merkle Integrity** | Source maps are versioned, reproducible, and verifiable via cryptographic SHA256 hashes. | Manifest and repository tree hashing (`internal/store:BuildSourceIdentity`). |\n")
	sb.WriteString("| **Fail-Closed Boundary Modeling** | Unhandled languages and unresolvable peers are explicitly recorded, never silently dropped. | `domain.UnsupportedBoundary` records throughout all extractor layers. |\n")
	sb.WriteString("| **Dual-Surface Parity** | Human CLI (`cmd/gira`) and AI Agent Protocol (`internal/mcp`) query the identical deterministic facts. | Shared `slices.Engine` data structures and budget envelopes. |\n\n")

	// 2. Layered Architecture Block Diagram
	sb.WriteString("## 2. Layered Architecture Block Diagram\n\n")
	sb.WriteString("The system is structured in six strictly decoupled layers, enforcing unidirectional dependency flows:\n\n")
	sb.WriteString("```mermaid\ngraph TD\n")
	sb.WriteString("  subgraph L1[\"Layer 1: Entrypoints & Communication Interfaces\"]\n")
	sb.WriteString("    CLI[\"CLI Interface<br/>(cmd/gira)\"]\n")
	sb.WriteString("    MCP[\"MCP JSON-RPC Server<br/>(internal/mcp)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  subgraph L2[\"Layer 2: Core Engine, Slices & Budgeting\"]\n")
	sb.WriteString("    ENG[\"Slices Engine<br/>(internal/slices:Engine)\"]\n")
	sb.WriteString("    IDX[\"Indexer Orchestrator<br/>(internal/indexer)\"]\n")
	sb.WriteString("    TOK[\"Token Budgeting<br/>(internal/tokens)\"]\n")
	sb.WriteString("    ARCH[\"Living Architecture Engine<br/>(internal/architecture)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  subgraph L3[\"Layer 3: Semantic Extraction & AST Manipulation\"]\n")
	sb.WriteString("    MANIP[\"AST Manipulator<br/>(internal/manipulator)\"]\n")
	sb.WriteString("    EXT1[\"Layer 1 Data Extractor<br/>(internal/extractors:Layer1Data)\"]\n")
	sb.WriteString("    EXT2[\"Layer 2 Config/Infra<br/>(Compose, Docker, Helm, Mise)\"]\n")
	sb.WriteString("    EXT3[\"Layer 3 Language ASTs<br/>(Go, TS, Python, Protobuf)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  subgraph L4[\"Layer 4: Domain & Knowledge Graph\"]\n")
	sb.WriteString("    DOM[\"MacroMap & Evidence Records<br/>(internal/domain:MacroMap)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  subgraph L5[\"Layer 5: Governance, Quality Gates & Eval\"]\n")
	sb.WriteString("    GATE[\"Architecture Gate & Lint<br/>(internal/gate)\"]\n")
	sb.WriteString("    EVAL[\"Scorecard & Findings Register<br/>(internal/eval)\"]\n")
	sb.WriteString("    DIFF[\"Arch Diff Engine<br/>(internal/archdiff)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  subgraph L6[\"Layer 6: Storage, Caching & Cryptographic Integrity\"]\n")
	sb.WriteString("    STORE[\"Deterministic JSON Store<br/>(internal/store)\"]\n")
	sb.WriteString("    CACHE[\"Bare Git Cache Manager<br/>(internal/gitcache)\"]\n")
	sb.WriteString("  end\n\n")
	sb.WriteString("  CLI --> ENG\n")
	sb.WriteString("  CLI --> MANIP\n")
	sb.WriteString("  CLI --> GATE\n")
	sb.WriteString("  CLI --> ARCH\n")
	sb.WriteString("  MCP --> ENG\n")
	sb.WriteString("  MCP --> GATE\n")
	sb.WriteString("  ENG --> DOM\n")
	sb.WriteString("  ENG --> TOK\n")
	sb.WriteString("  IDX --> EXT1\n")
	sb.WriteString("  IDX --> EXT2\n")
	sb.WriteString("  IDX --> EXT3\n")
	sb.WriteString("  IDX --> STORE\n")
	sb.WriteString("  EXT1 --> DOM\n")
	sb.WriteString("  EXT2 --> DOM\n")
	sb.WriteString("  EXT3 --> DOM\n")
	sb.WriteString("  GATE --> DOM\n")
	sb.WriteString("  GATE --> EVAL\n")
	sb.WriteString("  GATE --> DIFF\n")
	sb.WriteString("  STORE --> DOM\n")
	sb.WriteString("  STORE --> CACHE\n")
	sb.WriteString("```\n\n")

	// 3. Architectural Communities & Module Breakdown
	sb.WriteString("## 3. Architectural Communities & Module Breakdown\n\n")
	sb.WriteString("AST nodes, contracts, and packages are clustered into cohesive architectural communities. Within each community, components are classified by graph centrality as **Core Abstractions** (high in-degree hubs) or **Utilities / Leaf Nodes**.\n\n")

	if arch != nil && len(arch.Communities) > 0 {
		commIdx := 1
		for _, did := range arch.Domains {
			comm, ok := arch.Communities[did]
			if !ok || (len(comm.CoreAbstractions) == 0 && len(comm.Utilities) == 0) {
				continue
			}

			sb.WriteString(fmt.Sprintf("### 3.%d %s\n\n", commIdx, comm.Title))
			commIdx++
			sb.WriteString(fmt.Sprintf("**Mission**: %s\n\n", comm.Description))

			if len(comm.Invariants) > 0 {
				sb.WriteString("#### Architectural Invariants\n\n")
				for _, inv := range comm.Invariants {
					sb.WriteString(fmt.Sprintf("- %s\n", inv))
				}
				sb.WriteString("\n")
			}

			// Core Abstractions table
			sb.WriteString(fmt.Sprintf("#### Core Abstractions (%d)\n\n", len(comm.CoreAbstractions)))
			if len(comm.CoreAbstractions) == 0 {
				sb.WriteString("- *(None declared)*\n\n")
			} else {
				sb.WriteString("| Component / Abstraction | Kind | Centrality | In/Out Degree | Code Citation |\n")
				sb.WriteString("| :--- | :---: | :---: | :---: | :--- |\n")
				for _, n := range comm.CoreAbstractions {
					cit := formatCitationLink(repoRoot, n.FileLocator)
					locText := n.FileLocator.Path
					if locText == "" {
						locText = "locator"
					}
					sb.WriteString(fmt.Sprintf("| [`%s`](%s) | `%s` | **%d** | in:%d / out:%d | [`%s`](%s) |\n",
						n.Name, cit, n.Kind, n.Centrality, n.InDegree, n.OutDegree, locText, cit))
				}
				sb.WriteString("\n")
			}

			// Utilities table
			if len(comm.Utilities) > 0 {
				sb.WriteString(fmt.Sprintf("#### Utilities & Leaf Nodes (%d)\n\n", len(comm.Utilities)))
				sb.WriteString("| Component / Package | Role | Centrality | Description |\n")
				sb.WriteString("| :--- | :---: | :---: | :--- |\n")
				for _, u := range comm.Utilities {
					desc := u.Description
					if desc == "" {
						desc = "-"
					}
					sb.WriteString(fmt.Sprintf("| [`%s`](%s) | `%s` | %d | %s |\n",
						u.Name, formatCitationLink(repoRoot, u.FileLocator), u.Role, u.Centrality, desc))
				}
				sb.WriteString("\n")
			}

			// Cross-Domain relations
			if len(comm.CrossDomainEdges) > 0 {
				sb.WriteString("#### Cross-Domain Mediation\n\n")
				seenRels := make(map[string]bool)
				for _, rel := range comm.CrossDomainEdges {
					key := fmt.Sprintf("%s->%s->%s(%s)", rel.SourceDomain, rel.Verb, rel.TargetDomain, rel.MediatorNode)
					if seenRels[key] {
						continue
					}
					seenRels[key] = true
					sb.WriteString(fmt.Sprintf("- `%s` &rarr; **%s** &rarr; `%s` (via `%s`)\n",
						rel.SourceDomain, rel.Verb, rel.TargetDomain, rel.MediatorNode))
				}
				sb.WriteString("\n")
			}
		}
	}

	// 4. Critical Lifecycles & Intent Flows
	sb.WriteString("## 4. Critical Lifecycles & Intent Flows\n\n")
	sb.WriteString("Execution flows in Gira-G follow deterministic lifecycles that travel predictably from entrypoints to storage:\n\n")

	lifecycles := DefaultLifecycles()
	for i, lc := range lifecycles {
		sb.WriteString(fmt.Sprintf("### 4.%d %s (`intent: %s`)\n\n", i+1, lc.Title, lc.Intent))
		sb.WriteString(fmt.Sprintf("%s\n\n", lc.Description))

		if lc.MermaidSeq != "" {
			sb.WriteString("```mermaid\n")
			sb.WriteString(lc.MermaidSeq)
			sb.WriteString("\n```\n\n")
		}

		sb.WriteString("#### Step-by-Step Flow\n\n")
		sb.WriteString("| Step | Component | Domain | Action | Source Citation |\n")
		sb.WriteString("| :---: | :--- | :--- | :--- | :--- |\n")
		for _, s := range lc.Steps {
			sb.WriteString(fmt.Sprintf("| %d | `%s` | %s | %s | [%s](%s) |\n",
				s.Order, s.Component, s.Domain, s.Action, s.Component, s.CodeLocator))
		}
		sb.WriteString("\n")
	}

	// 5. Codebase Packages & Public Contracts Catalog
	sb.WriteString("## 5. API Contracts & Module Directory\n\n")
	sb.WriteString("Direct reference to all core Go packages, their primary types, and essential exported contracts:\n\n")

	if len(pkgs) > 0 {
		for _, p := range pkgs {
			pkgLink := formatFileLink(repoRoot, p.Path)
			sb.WriteString(fmt.Sprintf("### Package [`%s`](%s)\n\n", p.Path, pkgLink))
			if p.Doc != "" {
				sb.WriteString(fmt.Sprintf("> %s\n\n", p.Doc))
			}
			sb.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", p.Domain))
			sb.WriteString(fmt.Sprintf("- **Internal Dependencies**: `%s`\n\n", strings.Join(p.Imports, ", ")))

			if len(p.Types) > 0 {
				sb.WriteString("#### Exported Types & Contracts\n\n")
				sb.WriteString("| Type | Kind | Description | Locator Citation |\n")
				sb.WriteString("| :--- | :---: | :--- | :--- |\n")
				for _, t := range p.Types {
					tCit := formatCitationLink(repoRoot, t.Locator)
					tDoc := t.Doc
					if tDoc == "" {
						tDoc = "-"
					}
					locText := t.Locator.Path
					if locText == "" {
						locText = "locator"
					}
					sb.WriteString(fmt.Sprintf("| [`%s`](%s) | `%s` | %s | [`%s`](%s) |\n",
						t.Name, tCit, t.Kind, tDoc, locText, tCit))
				}
				sb.WriteString("\n")
			}

			if len(p.Functions) > 0 {
				sb.WriteString("#### Key Exported Functions\n\n")
				sb.WriteString("```go\n")
				for _, f := range p.Functions {
					if f.Exported {
						sb.WriteString(fmt.Sprintf("%s\n", f.Signature))
					}
				}
				sb.WriteString("```\n\n")
			}
		}
	}

	return sb.String()
}

const githubRepoBase = "https://github.com/EduardoShibukawa/gira-g/blob/main"

func formatCitationLink(repoRoot string, loc domain.Locator) string {
	if loc.Path == "" {
		return "#"
	}
	cleanPath := filepath.ToSlash(loc.Path)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if (loc.StartLine > 1 || loc.EndLine > 1) && loc.StartLine > 0 && loc.EndLine >= loc.StartLine {
		if loc.StartLine == loc.EndLine {
			return fmt.Sprintf("%s/%s#L%d", githubRepoBase, cleanPath, loc.StartLine)
		}
		return fmt.Sprintf("%s/%s#L%d-L%d", githubRepoBase, cleanPath, loc.StartLine, loc.EndLine)
	} else if loc.StartLine > 1 {
		return fmt.Sprintf("%s/%s#L%d", githubRepoBase, cleanPath, loc.StartLine)
	}
	return fmt.Sprintf("%s/%s", githubRepoBase, cleanPath)
}

func formatFileLink(repoRoot, relPath string) string {
	cleanPath := filepath.ToSlash(relPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	return fmt.Sprintf("%s/%s", githubRepoBase, cleanPath)
}
