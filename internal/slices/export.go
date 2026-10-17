package slices

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// ExportMarkdown produces a complete, standalone architectural Markdown document
// for a MacroMap with discrete origin tags, topology hierarchy, and unsupported boundaries.
func ExportMarkdown(mm *domain.MacroMap) string {
	if mm == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Target System Architecture: %s\n\n", mm.TargetSystem))
	sb.WriteString(fmt.Sprintf("- Schema Version: `%s`\n", mm.SchemaVersion))
	sb.WriteString(fmt.Sprintf("- Total Nodes: %d\n", len(mm.Nodes)))
	sb.WriteString(fmt.Sprintf("- Total Relationships: %d\n", len(mm.Relationships)))
	sb.WriteString(fmt.Sprintf("- Unsupported Boundaries: %d\n\n", len(mm.UnsupportedFiles())))

	// Group nodes by Kind
	services := make([]domain.Node, 0)
	infrastructures := make([]domain.Node, 0)
	contracts := make([]domain.Node, 0)
	libraries := make([]domain.Node, 0)

	for _, n := range mm.Nodes {
		switch n.Kind {
		case domain.NodeService:
			services = append(services, n)
		case domain.NodeInfrastructure:
			infrastructures = append(infrastructures, n)
		case domain.NodeContract:
			contracts = append(contracts, n)
		case domain.NodeLibrary:
			libraries = append(libraries, n)
		}
	}

	// Deterministic sorting
	sortNodesByID(services)
	sortNodesByID(infrastructures)
	sortNodesByID(contracts)
	sortNodesByID(libraries)

	// Summary by Kind
	sb.WriteString("## Architectural Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Services (%d)**: Top-level runnable microservices and processing units.\n", len(services)))
	sb.WriteString(fmt.Sprintf("- **Infrastructure (%d)**: Datastores, brokers, caches, and storage infrastructure.\n", len(infrastructures)))
	sb.WriteString(fmt.Sprintf("- **Contracts (%d)**: Explicit communication interfaces (Protobuf, gRPC, OpenAPI, message schemas).\n", len(contracts)))
	sb.WriteString(fmt.Sprintf("- **Libraries (%d)**: Internal shared utility packages and modules.\n\n", len(libraries)))

	// Section 1: Services
	sb.WriteString(fmt.Sprintf("## Services (%d)\n\n", len(services)))
	if len(services) == 0 {
		sb.WriteString("- (none declared or extracted)\n\n")
	} else {
		for _, s := range services {
			renderNodeMarkdownItem(&sb, s)
		}
		sb.WriteString("\n")
	}

	// Section 2: Infrastructure
	sb.WriteString(fmt.Sprintf("## Infrastructure (%d)\n\n", len(infrastructures)))
	if len(infrastructures) == 0 {
		sb.WriteString("- (none declared or extracted)\n\n")
	} else {
		for _, inf := range infrastructures {
			renderNodeMarkdownItem(&sb, inf)
		}
		sb.WriteString("\n")
	}

	// Section 3: Contracts
	sb.WriteString(fmt.Sprintf("## Contracts (%d)\n\n", len(contracts)))
	if len(contracts) == 0 {
		sb.WriteString("- (none declared or extracted)\n\n")
	} else {
		for _, c := range contracts {
			renderContractMarkdown(&sb, c)
		}
		sb.WriteString("\n")
	}

	// Section 4: Libraries
	sb.WriteString(fmt.Sprintf("## Libraries (%d)\n\n", len(libraries)))
	if len(libraries) == 0 {
		sb.WriteString("- (none declared or extracted)\n\n")
	} else {
		for _, l := range libraries {
			renderNodeMarkdownItem(&sb, l)
		}
		sb.WriteString("\n")
	}

	// Section 5: Mediated Relationships Table
	sortedRels := make([]domain.Relationship, len(mm.Relationships))
	copy(sortedRels, mm.Relationships)
	sort.Slice(sortedRels, func(i, j int) bool { return sortedRels[i].ID < sortedRels[j].ID })

	sb.WriteString(fmt.Sprintf("## Mediated Relationships (%d)\n\n", len(sortedRels)))
	if len(sortedRels) == 0 {
		sb.WriteString("- (no mediated relationships found)\n\n")
	} else {
		sb.WriteString("| Source | Verb | Target | Derivation Origin | Freshness | Locator Citation |\n")
		sb.WriteString("| :--- | :---: | :--- | :---: | :---: | :--- |\n")
		for _, r := range sortedRels {
			origin := "-"
			status := "-"
			loc := "-"
			if len(r.Evidence) > 0 {
				origin = string(r.Evidence[0].Origin)
				status = string(r.Evidence[0].Status)
				if r.Evidence[0].Locator.Path != "" {
					loc = fmt.Sprintf("`%s`", r.Evidence[0].Locator.String())
				}
			}
			sb.WriteString(fmt.Sprintf("| `%s` | **%s** | `%s` | `%s` | `%s` | %s |\n",
				r.SourceID, r.Verb, r.TargetID, origin, status, loc))
		}
		sb.WriteString("\n")
	}

	// Section 6: Unsupported Boundaries & Known Unknowns (Mandatory per ADR 0008)
	sortedUnsupported := mm.UnsupportedFiles()
	sort.Slice(sortedUnsupported, func(i, j int) bool { return sortedUnsupported[i].Path < sortedUnsupported[j].Path })

	sb.WriteString(fmt.Sprintf("## Unsupported Boundaries & Known Unknowns (%d)\n\n", len(sortedUnsupported)))
	if len(sortedUnsupported) == 0 {
		sb.WriteString("- (no unhandled languages or unsupported boundaries detected)\n")
	} else {
		for _, ub := range sortedUnsupported {
			locStr := ""
			if ub.Locator.Path != "" {
				locStr = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("> ⚠️ UNSUPPORTED: `%s` - %s%s\n", ub.Path, ub.Reason, locStr))
		}
	}

	// Section 7: unsupported constructs inside parsed files (F7). Not
	// Unsupported Boundaries, but still explicit unknowns, never silent.
	if constructs := sortedConstructs(mm); len(constructs) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Unsupported Constructs in Parsed Files (%d)\n\n", len(constructs)))
		for _, ub := range constructs {
			locStr := ""
			if ub.Locator.Path != "" {
				locStr = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("> ⚠️ UNKNOWN: `%s` - %s%s\n", ub.Path, ub.Reason, locStr))
		}
	}

	return sb.String()
}

// sortedConstructs returns the unsupported constructs inside parsed files,
// sorted by path.
func sortedConstructs(mm *domain.MacroMap) []domain.UnsupportedBoundary {
	rows := mm.UnsupportedConstructRows()
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	return rows
}

func sortNodesByID(nodes []domain.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
}

func renderNodeMarkdownItem(sb *strings.Builder, n domain.Node) {
	loc := primaryLocatorCitation(n.Evidence)
	origin := ""
	if len(n.Evidence) > 0 {
		origin = fmt.Sprintf(" [origin: %s, %s]", n.Evidence[0].Origin, n.Evidence[0].Status)
	}
	sb.WriteString(fmt.Sprintf("- **%s** (`%s`)%s%s\n", n.Name, n.ID, origin, loc))
	if n.Description != "" {
		sb.WriteString(fmt.Sprintf("  - Description: %s\n", n.Description))
	}
	if hasStaleEvidence(n.Evidence) {
		sb.WriteString(fmt.Sprintf("  > ⚠️ NOTICE: Stale evidence detected for `%s`\n", n.ID))
	}
}

// ExportHTML produces a standalone, zero-dependency static HTML document with embedded SVG
// for the topological graph, discrete derivation tags, task-oriented exploration controls,
// guided tours/flows, and no external CSS/JS/fonts.
func ExportHTML(mm *domain.MacroMap) string {
	if mm == nil {
		mm = domain.NewMacroMap("empty")
	}

	var sb strings.Builder

	// Prepare data
	services := make([]domain.Node, 0)
	infrastructures := make([]domain.Node, 0)
	contracts := make([]domain.Node, 0)
	libraries := make([]domain.Node, 0)

	for _, n := range mm.Nodes {
		switch n.Kind {
		case domain.NodeService:
			services = append(services, n)
		case domain.NodeInfrastructure:
			infrastructures = append(infrastructures, n)
		case domain.NodeContract:
			contracts = append(contracts, n)
		case domain.NodeLibrary:
			libraries = append(libraries, n)
		}
	}
	sortNodesByID(services)
	sortNodesByID(infrastructures)
	sortNodesByID(contracts)
	sortNodesByID(libraries)

	sortedRels := make([]domain.Relationship, len(mm.Relationships))
	copy(sortedRels, mm.Relationships)
	sort.Slice(sortedRels, func(i, j int) bool { return sortedRels[i].ID < sortedRels[j].ID })

	sortedUnsupported := mm.UnsupportedFiles()
	sort.Slice(sortedUnsupported, func(i, j int) bool { return sortedUnsupported[i].Path < sortedUnsupported[j].Path })

	eng := NewEngine(mm)
	tour, _ := eng.GetTour("", 0)

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Gira-G Architecture Export: ` + html.EscapeString(mm.TargetSystem) + `</title>
<style>
:root {
  --bg: #0f141c;
  --surface: #18202c;
  --surface-border: #283548;
  --text: #e2e8f0;
  --text-dim: #94a3b8;
  --accent: #38bdf8;
  --accent-contract: #c084fc;
  --accent-infra: #f59e0b;
  --accent-lib: #10b981;
  --tag-declared: #3b82f6;
  --tag-extracted: #10b981;
  --tag-inferred: #f59e0b;
  --tag-stale: #ef4444;
  --tag-current: #10b981;
  --code-bg: #0b0f17;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background-color: var(--bg);
  color: var(--text);
  line-height: 1.5;
  padding: 2rem;
}
.container { max-width: 1200px; margin: 0 auto; }
header {
  border-bottom: 2px solid var(--surface-border);
  padding-bottom: 1.5rem;
  margin-bottom: 2rem;
}
h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 0.5rem; }
.meta-bar {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
  font-size: 0.875rem;
  color: var(--text-dim);
}
.meta-item strong { color: var(--text); }
h2 {
  font-size: 1.35rem;
  margin: 2rem 0 1rem 0;
  color: #f1f5f9;
  border-bottom: 1px solid var(--surface-border);
  padding-bottom: 0.4rem;
}
.badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.badge-declared { background: rgba(59, 130, 246, 0.2); color: var(--tag-declared); border: 1px solid var(--tag-declared); }
.badge-extracted { background: rgba(16, 185, 129, 0.2); color: var(--tag-extracted); border: 1px solid var(--tag-extracted); }
.badge-inferred { background: rgba(245, 158, 11, 0.2); color: var(--tag-inferred); border: 1px solid var(--tag-inferred); }
.badge-current { background: rgba(16, 185, 129, 0.15); color: var(--tag-current); }
.badge-stale { background: rgba(239, 68, 68, 0.2); color: var(--tag-stale); border: 1px solid var(--tag-stale); }
.badge-unverified { background: rgba(245, 158, 11, 0.2); color: var(--tag-inferred); border: 1px solid var(--tag-inferred); }
.badge-profile { background: rgba(148, 163, 184, 0.15); color: var(--text-dim); border: 1px solid var(--surface-border); }

.controls-bar {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 0.85rem 1rem;
  margin-bottom: 1.5rem;
}
.controls-bar input, .controls-bar select, .controls-bar button {
  background: var(--code-bg);
  color: var(--text);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  padding: 0.45rem 0.75rem;
  font-size: 0.85rem;
}
.controls-bar input { flex: 1; min-width: 220px; }
.controls-bar button { cursor: pointer; }
.controls-bar button:hover { border-color: var(--accent); }

.inspector-panel {
  background: var(--surface);
  border: 1px solid var(--accent);
  border-radius: 6px;
  padding: 1.25rem;
  margin-bottom: 2rem;
}
.inspector-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
  margin-top: 0.75rem;
}
.inspector-col {
  background: var(--code-bg);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  padding: 0.85rem;
}
.inspector-col h4 { color: var(--accent); font-size: 0.85rem; text-transform: uppercase; margin-bottom: 0.5rem; }
.inspector-col ul { padding-left: 1.2rem; font-size: 0.85rem; }

.summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}
.card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 1.25rem;
}
.card h3 { font-size: 0.9rem; text-transform: uppercase; color: var(--text-dim); margin-bottom: 0.25rem; }
.card .count { font-size: 2rem; font-weight: bold; color: var(--accent); }

.node-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.node-card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 1rem;
  cursor: pointer;
  transition: border-color 0.15s ease;
}
.node-card:hover, .node-card.selected { border-color: var(--accent); }
.node-card.dimmed, tr.dimmed, g.svg-node.dimmed, g.svg-edge.dimmed { opacity: 0.25; }
.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}
.node-title { font-size: 1.1rem; font-weight: bold; color: #fff; }
.node-id { font-family: monospace; font-size: 0.8rem; color: var(--text-dim); word-break: break-all; }
.locator {
  font-family: monospace;
  font-size: 0.75rem;
  background: var(--code-bg);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  margin-top: 0.35rem;
  color: #38bdf8;
  word-break: break-all;
  cursor: pointer;
  display: inline-block;
}
.evidence-list { margin-top: 0.5rem; font-size: 0.8rem; }
.evidence-item { margin-top: 0.3rem; padding: 0.3rem 0.45rem; background: rgba(0,0,0,0.2); border-radius: 4px; }
.stale-warning {
  margin-top: 0.45rem;
  padding: 0.35rem 0.5rem;
  background: rgba(239, 68, 68, 0.15);
  border-left: 3px solid var(--tag-stale);
  color: #fca5a5;
  font-size: 0.8rem;
}
.methods-list {
  margin-top: 0.5rem;
  padding-left: 1.2rem;
  font-size: 0.85rem;
  color: var(--text-dim);
}
.methods-list li { margin-bottom: 0.2rem; }

.table-wrap {
  overflow-x: auto;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  margin-bottom: 2rem;
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  text-align: left;
}
th, td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--surface-border);
  vertical-align: top;
}
th {
  background: rgba(0,0,0,0.2);
  color: var(--text-dim);
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.75rem;
}
tr:last-child td { border-bottom: none; }
.verb { font-weight: bold; color: var(--accent-contract); }

.graph-container {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  overflow-x: auto;
}
.svg-node { cursor: pointer; }

.tour-panel {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}
.tour-panel ol { padding-left: 1.4rem; }
.tour-panel li { margin-bottom: 0.65rem; }

.unsupported-banner {
  background: rgba(239, 68, 68, 0.1);
  border-left: 4px solid var(--tag-stale);
  padding: 0.85rem 1rem;
  margin-bottom: 0.75rem;
  border-radius: 0 4px 4px 0;
  font-size: 0.875rem;
}
.unsupported-path { font-family: monospace; font-weight: bold; color: #fca5a5; }
</style>
</head>
<body>
<div class="container">
<header>
  <h1>Target System: ` + html.EscapeString(mm.TargetSystem) + `</h1>
  <div class="meta-bar">
    <div class="meta-item">Schema: <strong>` + html.EscapeString(mm.SchemaVersion) + `</strong></div>
    <div class="meta-item">Nodes: <strong>` + fmt.Sprintf("%d", len(mm.Nodes)) + `</strong></div>
    <div class="meta-item">Relationships: <strong>` + fmt.Sprintf("%d", len(mm.Relationships)) + `</strong></div>
    <div class="meta-item">Unsupported Boundaries: <strong>` + fmt.Sprintf("%d", len(sortedUnsupported)) + `</strong></div>
  </div>
</header>

<div class="controls-bar" role="region" aria-label="Task-oriented architecture filters">
  <input id="gira-search" type="search" placeholder="Search nodes, contracts, methods, or locators..." oninput="giraApplyFilters()">
  <select id="gira-filter-kind" onchange="giraApplyFilters()">
    <option value="">All Kinds</option>
    <option value="service">Services</option>
    <option value="infrastructure">Infrastructure</option>
    <option value="contract">Contracts</option>
    <option value="library">Libraries</option>
  </select>
  <select id="gira-filter-origin" onchange="giraApplyFilters()">
    <option value="">All Origins</option>
    <option value="declared">Declared</option>
    <option value="extracted">Extracted</option>
    <option value="inferred">Inferred</option>
  </select>
  <select id="gira-filter-status" onchange="giraApplyFilters()">
    <option value="">All Freshness</option>
    <option value="current">Current</option>
    <option value="stale">Stale (⚠️)</option>
    <option value="unverified">Unverified (⚠️)</option>
  </select>
  <button type="button" onclick="giraResetSelection()">Reset Selection</button>
</div>

<div id="gira-inspector" class="inspector-panel">
  <h3 id="gira-inspector-title">Interactive Component Inspector (click any node in graph or cards)</h3>
  <div class="inspector-grid">
    <div class="inspector-col">
      <h4>O que é isto? (Identity &amp; Evidence)</h4>
      <div id="gira-inspector-what">Select a node to inspect its kind, contracts, and discrete evidence locators.</div>
    </div>
    <div class="inspector-col">
      <h4>Quem consome? (Downstream Consumers)</h4>
      <div id="gira-inspector-who">Select a node to highlight downstream consumers and mediating contracts.</div>
    </div>
    <div class="inspector-col">
      <h4>O que muda? (Upstream &amp; Impact)</h4>
      <div id="gira-inspector-change">Select a node to inspect upstream dependencies and blast radius.</div>
    </div>
  </div>
</div>

<div class="summary-cards">
  <div class="card">
    <h3>Services</h3>
    <div class="count">` + fmt.Sprintf("%d", len(services)) + `</div>
  </div>
  <div class="card">
    <h3>Infrastructure</h3>
    <div class="count" style="color: var(--accent-infra);">` + fmt.Sprintf("%d", len(infrastructures)) + `</div>
  </div>
  <div class="card">
    <h3>Contracts</h3>
    <div class="count" style="color: var(--accent-contract);">` + fmt.Sprintf("%d", len(contracts)) + `</div>
  </div>
  <div class="card">
    <h3>Libraries</h3>
    <div class="count" style="color: var(--accent-lib);">` + fmt.Sprintf("%d", len(libraries)) + `</div>
  </div>
</div>
`)

	// Topological Graph (SVG)
	sb.WriteString("<h2>Topological Graph (Mediated Relations)</h2>\n")
	sb.WriteString("<div class=\"graph-container\">\n")
	renderSVGGraph(&sb, mm)
	sb.WriteString("</div>\n")

	// Guided Tour & Flows Section
	sb.WriteString("<h2>Guided Tour &amp; Flows</h2>\n")
	sb.WriteString("<div class=\"tour-panel\" id=\"gira-tour-flows\">\n")
	if tour == nil || len(tour.Steps) == 0 {
		sb.WriteString("<p style=\"color: var(--text-dim);\">(no guided tour entry points declared)</p>\n")
	} else {
		sb.WriteString(fmt.Sprintf("<h3>Guided Reading Tour (%d steps)</h3>\n<ol>\n", len(tour.Steps)))
		for _, step := range tour.Steps {
			nodeFocus := ""
			if len(step.NodeIDs) > 0 {
				nodeFocus = step.NodeIDs[0]
			}
			sb.WriteString(fmt.Sprintf("<li><strong onclick=\"giraSelectNode('%s')\" style=\"cursor:pointer;color:var(--accent);\">%d. %s</strong> (<code>%s</code>)<br><span>%s</span>",
				html.EscapeString(nodeFocus), step.Order, html.EscapeString(step.Title),
				html.EscapeString(strings.Join(step.NodeIDs, ", ")), html.EscapeString(step.Why)))
			if step.WhyLocator != "" {
				sb.WriteString(fmt.Sprintf(" <span class=\"locator\" onclick=\"giraCopyLocator(this)\" title=\"Click to copy locator\">[%s]</span>", html.EscapeString(step.WhyLocator)))
			}
			if len(step.Read) > 0 {
				reads := make([]string, len(step.Read))
				for i, r := range step.Read {
					reads[i] = fmt.Sprintf("<span class=\"locator\" onclick=\"giraCopyLocator(this)\">%s</span>", html.EscapeString(r))
				}
				sb.WriteString("<br><small>Read: " + strings.Join(reads, " ") + "</small>")
			}
			sb.WriteString("</li>\n")
		}
		sb.WriteString("</ol>\n")
	}
	if len(services) > 0 {
		hasAnyFlow := false
		for _, svc := range services {
			fls, err := eng.GetFlows(svc.ID)
			if err != nil || fls == nil || len(fls.Flows) == 0 {
				continue
			}
			if !hasAnyFlow {
				sb.WriteString("<h3 style=\"margin-top:1rem;\">Deterministic Flows</h3>\n<ul>\n")
				hasAnyFlow = true
			}
			for _, fl := range fls.Flows {
				hops := make([]string, len(fl.Steps))
				for i, h := range fl.Steps {
					if h.Locator != "" {
						hops[i] = fmt.Sprintf("<code>%s</code> <span class=\"locator\">[%s]</span>", html.EscapeString(h.NodeID), html.EscapeString(h.Locator))
					} else {
						hops[i] = fmt.Sprintf("<code>%s</code>", html.EscapeString(h.NodeID))
					}
				}
				sb.WriteString("<li>" + strings.Join(hops, " &rarr; ") + "</li>\n")
			}
		}
		if hasAnyFlow {
			sb.WriteString("</ul>\n")
		}
	}
	sb.WriteString("</div>\n")

	// Mediated Relationships Table
	sb.WriteString(fmt.Sprintf("<h2>Mediated Relationships (%d)</h2>\n", len(sortedRels)))
	sb.WriteString("<div class=\"table-wrap\">\n<table id=\"gira-rel-table\">\n<thead>\n<tr>\n")
	sb.WriteString("<th>Source</th><th>Verb</th><th>Target</th><th>Derivation</th><th>Freshness</th><th>Locator Citation</th>\n")
	sb.WriteString("</tr>\n</thead>\n<tbody>\n")
	if len(sortedRels) == 0 {
		sb.WriteString("<tr><td colspan=\"6\" style=\"text-align: center; color: var(--text-dim);\">No relationships defined.</td></tr>\n")
	} else {
		for _, r := range sortedRels {
			originsAttr, statusesAttr := collectEvidenceAttrs(r.Evidence)
			originCell := "-"
			statusCell := "-"
			locCell := "-"
			if len(r.Evidence) > 0 {
				var oBadges, sBadges, lItems []string
				for _, ev := range r.Evidence {
					oBadge := renderBadgeHTML(string(ev.Origin))
					if ev.DeploymentProfile != "" {
						oBadge += " " + fmt.Sprintf(`<span class="badge badge-profile">%s</span>`, html.EscapeString(ev.DeploymentProfile))
					}
					oBadges = append(oBadges, oBadge)
					sBadges = append(sBadges, renderBadgeHTML(string(ev.Status)))
					if ev.Locator.Path != "" {
						lItems = append(lItems, fmt.Sprintf(`<code class="locator" onclick="giraCopyLocator(this)" title="Click to copy locator">%s</code>`, html.EscapeString(ev.Locator.String())))
					}
				}
				originCell = strings.Join(oBadges, "<br>")
				statusCell = strings.Join(sBadges, "<br>")
				if len(lItems) > 0 {
					locCell = strings.Join(lItems, "<br>")
				}
			}
			sb.WriteString(fmt.Sprintf("<tr class=\"rel-row\" data-source=\"%s\" data-target=\"%s\" data-verb=\"%s\" data-origins=\"%s\" data-statuses=\"%s\"><td><code>%s</code></td><td class=\"verb\">%s</td><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
				html.EscapeString(r.SourceID),
				html.EscapeString(r.TargetID),
				html.EscapeString(string(r.Verb)),
				html.EscapeString(originsAttr),
				html.EscapeString(statusesAttr),
				html.EscapeString(r.SourceID),
				html.EscapeString(string(r.Verb)),
				html.EscapeString(r.TargetID),
				originCell,
				statusCell,
				locCell,
			))
		}
	}
	sb.WriteString("</tbody>\n</table>\n</div>\n")

	// Nodes by Kind
	renderNodeCategoryHTML(&sb, "Services", services)
	renderNodeCategoryHTML(&sb, "Infrastructure", infrastructures)
	renderNodeCategoryHTML(&sb, "Contracts", contracts)
	renderNodeCategoryHTML(&sb, "Libraries", libraries)

	// Unsupported Boundaries & Known Unknowns
	sb.WriteString(fmt.Sprintf("<h2>Unsupported Boundaries & Known Unknowns (%d)</h2>\n", len(sortedUnsupported)))
	if len(sortedUnsupported) == 0 {
		sb.WriteString("<p style=\"color: var(--text-dim);\">No unsupported boundaries or unhandled languages recorded.</p>\n")
	} else {
		for _, ub := range sortedUnsupported {
			loc := ""
			if ub.Locator.Path != "" {
				loc = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("<div class=\"unsupported-banner\">⚠️ <span class=\"unsupported-path\">%s</span>: %s%s</div>\n",
				html.EscapeString(ub.Path),
				html.EscapeString(ub.Reason),
				html.EscapeString(loc),
			))
		}
	}

	// Unsupported constructs inside parsed files (F7): explicit unknowns,
	// not Unsupported Boundaries.
	if constructs := sortedConstructs(mm); len(constructs) > 0 {
		sb.WriteString(fmt.Sprintf("<h2>Unsupported Constructs in Parsed Files (%d)</h2>\n", len(constructs)))
		for _, ub := range constructs {
			loc := ""
			if ub.Locator.Path != "" {
				loc = fmt.Sprintf(" [%s]", ub.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("<div class=\"unsupported-banner\">⚠️ <span class=\"unsupported-path\">%s</span>: %s%s</div>\n",
				html.EscapeString(ub.Path),
				html.EscapeString(ub.Reason),
				html.EscapeString(loc),
			))
		}
	}

	sb.WriteString(`<script>
function giraCopyLocator(el) {
  if (!el) return;
  var txt = el.textContent.replace(/^\[|\]$/g, '');
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(txt);
  }
}
function giraApplyFilters() {
  var q = (document.getElementById('gira-search').value || '').toLowerCase();
  var kind = document.getElementById('gira-filter-kind').value;
  var origin = document.getElementById('gira-filter-origin').value;
  var status = document.getElementById('gira-filter-status').value;
  document.querySelectorAll('.node-card').forEach(function(card) {
    var matchQ = !q || card.textContent.toLowerCase().indexOf(q) !== -1;
    var matchK = !kind || card.getAttribute('data-kind') === kind;
    var matchO = !origin || (card.getAttribute('data-origins') || '').indexOf(origin) !== -1;
    var matchS = !status || (card.getAttribute('data-statuses') || '').indexOf(status) !== -1;
    card.style.display = (matchQ && matchK && matchO && matchS) ? '' : 'none';
  });
  document.querySelectorAll('.rel-row').forEach(function(row) {
    var matchQ = !q || row.textContent.toLowerCase().indexOf(q) !== -1;
    var matchO = !origin || (row.getAttribute('data-origins') || '').indexOf(origin) !== -1;
    var matchS = !status || (row.getAttribute('data-statuses') || '').indexOf(status) !== -1;
    row.style.display = (matchQ && matchO && matchS) ? '' : 'none';
  });
}
function giraSelectNode(nodeId) {
  if (!nodeId) return;
  var upstream = [];
  var downstream = [];
  var related = {};
  related[nodeId] = true;
  document.querySelectorAll('.rel-row').forEach(function(row) {
    var src = row.getAttribute('data-source');
    var tgt = row.getAttribute('data-target');
    var verb = row.getAttribute('data-verb');
    if (src === nodeId) {
      related[tgt] = true;
      upstream.push(verb + ' -> ' + tgt);
      row.classList.remove('dimmed');
    } else if (tgt === nodeId) {
      related[src] = true;
      downstream.push(src + ' (' + verb + ')');
      row.classList.remove('dimmed');
    } else {
      row.classList.add('dimmed');
    }
  });
  document.querySelectorAll('.node-card').forEach(function(card) {
    var id = card.getAttribute('data-node-id');
    card.classList.toggle('selected', id === nodeId);
    card.classList.toggle('dimmed', !related[id]);
    if (id === nodeId) {
      document.getElementById('gira-inspector-title').textContent = 'Component Inspector: ' + id;
      document.getElementById('gira-inspector-what').innerHTML = card.innerHTML;
    }
  });
  document.querySelectorAll('.svg-node').forEach(function(g) {
    var id = g.getAttribute('data-node-id');
    g.classList.toggle('dimmed', !related[id]);
  });
  document.querySelectorAll('.svg-edge').forEach(function(e) {
    var s = e.getAttribute('data-source-id');
    var t = e.getAttribute('data-target-id');
    e.classList.toggle('dimmed', s !== nodeId && t !== nodeId);
  });
  document.getElementById('gira-inspector-who').innerHTML = downstream.length
    ? '<ul><li>' + downstream.join('</li><li>') + '</li></ul>'
    : '<span>No direct consumers recorded.</span>';
  document.getElementById('gira-inspector-change').innerHTML = upstream.length
    ? '<ul><li>' + upstream.join('</li><li>') + '</li></ul>'
    : '<span>No outgoing dependencies recorded.</span>';
}
function giraResetSelection() {
  document.getElementById('gira-search').value = '';
  document.getElementById('gira-filter-kind').value = '';
  document.getElementById('gira-filter-origin').value = '';
  document.getElementById('gira-filter-status').value = '';
  giraApplyFilters();
  document.querySelectorAll('.dimmed, .selected').forEach(function(el) {
    el.classList.remove('dimmed');
    el.classList.remove('selected');
  });
}
</script>
</div>
</body>
</html>`)

	return sb.String()
}

func collectEvidenceAttrs(evs []domain.EvidenceRecord) (string, string) {
	oSet := make(map[string]bool)
	sSet := make(map[string]bool)
	var origins, statuses []string
	for _, ev := range evs {
		if o := string(ev.Origin); o != "" && !oSet[o] {
			oSet[o] = true
			origins = append(origins, o)
		}
		if s := string(ev.Status); s != "" && !sSet[s] {
			sSet[s] = true
			statuses = append(statuses, s)
		}
	}
	return strings.Join(origins, ","), strings.Join(statuses, ",")
}

func renderBadgeHTML(val string) string {
	switch val {
	case "declared":
		return `<span class="badge badge-declared">declared</span>`
	case "extracted":
		return `<span class="badge badge-extracted">extracted</span>`
	case "inferred":
		return `<span class="badge badge-inferred">inferred</span>`
	case "current":
		return `<span class="badge badge-current">current</span>`
	case "stale":
		return `<span class="badge badge-stale">⚠️ stale</span>`
	case "unverified":
		return `<span class="badge badge-unverified">⚠️ unverified</span>`
	default:
		return html.EscapeString(val)
	}
}

func renderNodeCategoryHTML(sb *strings.Builder, category string, nodes []domain.Node) {
	sb.WriteString(fmt.Sprintf("<h2>%s (%d)</h2>\n", category, len(nodes)))
	if len(nodes) == 0 {
		sb.WriteString("<p style=\"color: var(--text-dim); margin-bottom: 1.5rem;\">(none)</p>\n")
		return
	}

	sb.WriteString("<div class=\"node-grid\">\n")
	for _, n := range nodes {
		originsAttr, statusesAttr := collectEvidenceAttrs(n.Evidence)
		sb.WriteString(fmt.Sprintf("<div class=\"node-card\" data-node-id=\"%s\" data-kind=\"%s\" data-origins=\"%s\" data-statuses=\"%s\" onclick=\"giraSelectNode('%s')\">\n",
			html.EscapeString(n.ID),
			html.EscapeString(string(n.Kind)),
			html.EscapeString(originsAttr),
			html.EscapeString(statusesAttr),
			html.EscapeString(n.ID),
		))
		sb.WriteString("<div class=\"node-header\">\n")
		sb.WriteString(fmt.Sprintf("<span class=\"node-title\">%s</span>\n", html.EscapeString(n.Name)))
		if len(n.Evidence) > 0 {
			sb.WriteString("<span>" + renderBadgeHTML(string(n.Evidence[0].Origin)) + " " + renderBadgeHTML(string(n.Evidence[0].Status)) + "</span>")
		}
		sb.WriteString("</div>\n")
		sb.WriteString(fmt.Sprintf("<div class=\"node-id\">%s</div>\n", html.EscapeString(n.ID)))
		if n.Description != "" {
			sb.WriteString(fmt.Sprintf("<p style=\"font-size: 0.85rem; margin-top: 0.4rem;\">%s</p>\n", html.EscapeString(n.Description)))
		}
		if hasStaleEvidence(n.Evidence) {
			sb.WriteString(fmt.Sprintf("<div class=\"stale-warning\">⚠️ NOTICE: Stale evidence detected for <code>%s</code></div>\n", html.EscapeString(n.ID)))
		}
		if len(n.Evidence) > 0 {
			sb.WriteString("<div class=\"evidence-list\">\n")
			for _, ev := range n.Evidence {
				sb.WriteString("<div class=\"evidence-item\">")
				sb.WriteString(renderBadgeHTML(string(ev.Origin)) + " " + renderBadgeHTML(string(ev.Status)))
				if ev.DeploymentProfile != "" {
					sb.WriteString(fmt.Sprintf(` <span class="badge badge-profile">%s</span>`, html.EscapeString(ev.DeploymentProfile)))
				}
				if ev.Locator.Path != "" {
					sb.WriteString(fmt.Sprintf("<div class=\"locator\" onclick=\"event.stopPropagation();giraCopyLocator(this)\" title=\"Click to copy locator\">%s</div>", html.EscapeString(ev.Locator.String())))
				}
				sb.WriteString("</div>\n")
			}
			sb.WriteString("</div>\n")
		}

		// Contract methods/messages if present
		if n.Metadata != nil {
			if rDefs, ok := n.Metadata["rpc_definitions"].([]map[string]string); ok && len(rDefs) > 0 {
				sb.WriteString("<ul class=\"methods-list\">\n")
				for _, r := range rDefs {
					sb.WriteString(fmt.Sprintf("<li><code>%s(%s) &rarr; %s</code></li>\n",
						html.EscapeString(r["name"]), html.EscapeString(r["request"]), html.EscapeString(r["response"])))
				}
				sb.WriteString("</ul>\n")
			} else if methods, ok := n.Metadata["methods"].([]string); ok && len(methods) > 0 {
				sb.WriteString("<ul class=\"methods-list\">\n")
				for _, m := range methods {
					sb.WriteString(fmt.Sprintf("<li><code>%s</code></li>\n", html.EscapeString(m)))
				}
				sb.WriteString("</ul>\n")
			}
		}

		sb.WriteString("</div>\n")
	}
	sb.WriteString("</div>\n")
}

// renderSVGGraph generates an inline SVG topological visualization of nodes and relations
func renderSVGGraph(sb *strings.Builder, mm *domain.MacroMap) {
	if len(mm.Nodes) == 0 {
		sb.WriteString("<p style=\"color: var(--text-dim);\">Empty graph.</p>")
		return
	}

	// Calculate deterministic positions for nodes in columns based on kind:
	// Col 0: Services
	// Col 1: Contracts & Libraries
	// Col 2: Infrastructure / Consumers
	nodeCols := make(map[string]int)
	nodesByCol := map[int][]string{0: {}, 1: {}, 2: {}}

	for _, n := range mm.Nodes {
		col := 0
		switch n.Kind {
		case domain.NodeService:
			col = 0
		case domain.NodeContract, domain.NodeLibrary:
			col = 1
		case domain.NodeInfrastructure:
			col = 2
		}
		nodeCols[n.ID] = col
		nodesByCol[col] = append(nodesByCol[col], n.ID)
	}

	for col := 0; col < 3; col++ {
		sort.Strings(nodesByCol[col])
	}

	maxNodesInCol := 0
	for _, ids := range nodesByCol {
		if len(ids) > maxNodesInCol {
			maxNodesInCol = len(ids)
		}
	}
	if maxNodesInCol == 0 {
		maxNodesInCol = 1
	}

	colWidth := 280
	rowHeight := 65
	boxWidth := 200
	boxHeight := 40
	paddingLeft := 40
	paddingTop := 40

	width := paddingLeft*2 + colWidth*2 + boxWidth
	height := paddingTop*2 + maxNodesInCol*rowHeight
	if height < 200 {
		height = 200
	}

	type point struct{ x, y int }
	nodePos := make(map[string]point)

	for col := 0; col < 3; col++ {
		ids := nodesByCol[col]
		for row, id := range ids {
			x := paddingLeft + col*colWidth
			y := paddingTop + row*rowHeight
			nodePos[id] = point{x: x, y: y}
		}
	}

	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" width="100%%" height="%d" xmlns="http://www.w3.org/2000/svg" style="background:#0b0f17; border-radius:6px;">`, width, height, height))
	sb.WriteString(`<defs>
  <marker id="arrow" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 0 L 10 5 L 0 10 z" fill="#64748b" />
  </marker>
</defs>
`)

	// Render edges with relationship verb labels
	for _, rel := range mm.Relationships {
		src, ok1 := nodePos[rel.SourceID]
		tgt, ok2 := nodePos[rel.TargetID]
		if !ok1 || !ok2 {
			continue
		}

		x1 := src.x + boxWidth
		y1 := src.y + boxHeight/2
		x2 := tgt.x
		y2 := tgt.y + boxHeight/2

		// If arrow goes backwards or in same column, adjust anchor points
		if src.x >= tgt.x {
			x1 = src.x
			x2 = tgt.x + boxWidth
		}

		midX := (x1 + x2) / 2
		midY := (y1+y2)/2 - 4

		sb.WriteString(fmt.Sprintf(`<g class="svg-edge" data-source-id="%s" data-target-id="%s">`+"\n",
			html.EscapeString(rel.SourceID), html.EscapeString(rel.TargetID)))
		sb.WriteString(fmt.Sprintf(`  <path d="M %d %d C %d %d, %d %d, %d %d" stroke="#64748b" stroke-width="2" fill="none" marker-end="url(#arrow)" />`+"\n",
			x1, y1, midX, y1, midX, y2, x2, y2))
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" fill="#94a3b8" font-size="10" text-anchor="middle" font-family="monospace">%s</text>`+"\n",
			midX, midY, html.EscapeString(string(rel.Verb))))
		sb.WriteString("</g>\n")
	}

	// Render node boxes with full-name <title> tooltips and data-node-id
	for _, n := range mm.Nodes {
		pos, ok := nodePos[n.ID]
		if !ok {
			continue
		}

		color := "#38bdf8" // default service
		switch n.Kind {
		case domain.NodeContract:
			color = "#c084fc"
		case domain.NodeInfrastructure:
			color = "#f59e0b"
		case domain.NodeLibrary:
			color = "#10b981"
		}

		sb.WriteString(fmt.Sprintf(`<g class="svg-node" data-node-id="%s" transform="translate(%d, %d)" onclick="giraSelectNode('%s')">`+"\n",
			html.EscapeString(n.ID), pos.x, pos.y, html.EscapeString(n.ID)))
		sb.WriteString(fmt.Sprintf(`  <title>%s (%s)</title>`+"\n",
			html.EscapeString(n.Name), html.EscapeString(n.ID)))
		sb.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" rx="6" fill="#18202c" stroke="%s" stroke-width="1.5" />`+"\n",
			boxWidth, boxHeight, color))
		sb.WriteString(fmt.Sprintf(`  <text x="12" y="24" fill="#f8fafc" font-size="12" font-family="-apple-system, sans-serif" font-weight="600">%s</text>`+"\n",
			html.EscapeString(truncateString(n.Name, 22))))
		sb.WriteString("</g>\n")
	}

	sb.WriteString("</svg>\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
