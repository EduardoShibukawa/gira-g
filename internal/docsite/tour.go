package docsite

import (
	"fmt"
)

// TourStep defines a single stage in the Understand-Anything style guided architectural tour.
type TourStep struct {
	Number      int
	Title       string
	Domain      string
	Description string
	Invariants  []string
	KeyFiles    []string
	SampleCmd   string
	Snippet     string
}

// GenerateTourPageHTML creates a progressive architectural walkthrough page (tour.html).
func GenerateTourPageHTML(config *SiteConfig, navGroups []NavGroup) string {
	steps := []TourStep{
		{
			Number:      1,
			Title:       "Entrypoint & Protocol Dispatch",
			Domain:      "CLI & MCP Entrypoints",
			Description: "The system begins with strict separation between human terminal execution and autonomous AI agent communication. The CLI orchestrates commands while the MCP server exposes JSON-RPC tools.",
			Invariants: []string{
				"Dual-Surface Parity: Human CLI and Agent MCP resolve the exact same deterministic AST facts.",
				"Zero Uncaught Panics: All errors return structured exit codes and error envelopes.",
			},
			KeyFiles: []string{
				"cmd/gira/main.go",
				"internal/cli/root.go",
				"internal/mcp/server.go",
			},
			SampleCmd: "gira --help",
			Snippet: `// cmd/gira/main.go
func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}`,
		},
		{
			Number:      2,
			Title:       "Deterministic AST Parsing & Extraction",
			Domain:      "Extractors Pipeline",
			Description: "Gira-G extracts AST symbols without spinning up heavyweight language servers. It uses native Go parsers and tree-sitter to build precise function signatures, struct layouts, and imports.",
			Invariants: []string{
				"Zero Probabilistic Guessing: Symbol bounds and lines are exact byte locators.",
				"Fail-Closed Boundaries: Unsupported languages are explicitly recorded in UnsupportedBoundary.",
			},
			KeyFiles: []string{
				"internal/parser/parser.go",
				"internal/extractors/layer3_treesitter.go",
				"internal/domain/evidence.go",
			},
			SampleCmd: "gira read --file internal/parser/parser.go",
			Snippet: `// internal/parser/parser.go
func ParseFile(fset *token.FileSet, path string) (*ast.File, error) {
    return parser.ParseFile(fset, path, nil, parser.ParseComments)
}`,
		},
		{
			Number:      3,
			Title:       "MacroMap & Cryptographic Evidence Store",
			Domain:      "Storage & Knowledge Graph",
			Description: "Parsed symbols and relationships are collected into a unified MacroMap knowledge graph. The store computes cryptographic SHA256 hashes of repository trees to ensure cache reproducibility.",
			Invariants: []string{
				"Append-Only Merkle Integrity: Manifests track exact git tree SHAs.",
				"Zero-Token Cache Hits: Subsequent index reads load directly from serialized JSON.",
			},
			KeyFiles: []string{
				"internal/domain/macro_map.go",
				"internal/store/store.go",
				"internal/store/identity.go",
			},
			SampleCmd: "gira index . --output .gira/macro-map.json",
			Snippet: `// internal/domain/macro_map.go
type MacroMap struct {
    Version       string         ` + "`json:\"version\"`" + `
    Nodes         []Node         ` + "`json:\"nodes\"`" + `
    Relationships []Relationship ` + "`json:\"relationships\"`" + `
}`,
		},
		{
			Number:      4,
			Title:       "Architectural Clustering & Centrality",
			Domain:      "Core Engine",
			Description: "The clustering engine analyzes the import topology, computing in-degree, out-degree, and centrality scores. It groups packages into 6 resilient architectural domains with verified contracts.",
			Invariants: []string{
				"Bounded Cross-Domain Coupling: Cycles between architectural domains are flagged.",
				"Centrality-Aware Surfacing: Hub packages receive higher prominence in diagrams and searches.",
			},
			KeyFiles: []string{
				"internal/architecture/clustering.go",
				"internal/architecture/generator.go",
				"internal/architecture/types.go",
			},
			SampleCmd: "gira arch query \"storage engine\"",
			Snippet: `// internal/architecture/clustering.go
func ClusterMacroMap(mm *domain.MacroMap) *ClusteredArchitecture {
    // Calculates inDegree, outDegree and partitions into 6 domains
}`,
		},
		{
			Number:      5,
			Title:       "Surgical AST Manipulation & Slicing",
			Domain:      "AST Manipulator",
			Description: "The manipulator allows automated agents to add members, replace methods, or extract bounded sub-slices without disturbing surrounding code formatting or comments.",
			Invariants: []string{
				"Formatting Preservation: SafeBraceCounter balances braces while retaining original indentation.",
				"Budget Enforcement: Slices return token-bounded envelopes to prevent LLM context overflows.",
			},
			KeyFiles: []string{
				"internal/manipulator/manipulator.go",
				"internal/slices/engine.go",
				"internal/tokens/budget.go",
			},
			SampleCmd: "gira slice --symbol Engine --budget 500",
			Snippet: `// internal/manipulator/manipulator.go
func (m *Manipulator) InsertStructField(structName, fieldDecl string) error {
    // Exact token-balanced insertion
}`,
		},
		{
			Number:      6,
			Title:       "Governance Gates & Living Documentation",
			Domain:      "Governance & Gate",
			Description: "The final layer runs automated architecture rules, computes health scorecards, and compiles this very living documentation site with interactive Cytoscape graphs and Mermaid diagrams.",
			Invariants: []string{
				"Fail-Closed CI Verification: Architectural regressions fail pull requests immediately.",
				"Living Documentation Parity: Documentation self-updates upon every git commit.",
			},
			KeyFiles: []string{
				"internal/gate/gate.go",
				"internal/docsite/builder.go",
				"internal/docsite/graph.go",
			},
			SampleCmd: "gira docs --output build",
			Snippet: `// internal/gate/gate.go
func EvaluateArchitecture(arch *ClusteredArchitecture) (*Scorecard, error) {
    // Enforces architectural invariants
}`,
		},
	}

	navHTML := renderSidebarNav(navGroups, "tour.html")

	var stepsHTML string
	for _, s := range steps {
		invariantsList := ""
		for _, inv := range s.Invariants {
			invariantsList += fmt.Sprintf("<li><strong>%s</strong></li>", inv)
		}

		filesList := ""
		for _, f := range s.KeyFiles {
			filesList += fmt.Sprintf("<li><code>%s</code></li>", f)
		}

		stepsHTML += fmt.Sprintf(`
        <section id="step-%d" class="tour-card">
          <div class="tour-step-badge">Step %d of 6 &bull; %s</div>
          <h2 style="margin-top: 0; margin-bottom: 0.75rem;">%s</h2>
          <p style="font-size: 1.05rem; color: #d1d5db; line-height: 1.6;">%s</p>

          <div style="margin: 1.25rem 0;">
            <div style="font-size: 0.8rem; text-transform: uppercase; color: var(--accent); font-weight: 700; margin-bottom: 0.4rem;">Enforced Invariants</div>
            <ul style="padding-left: 1.25rem; font-size: 0.9rem; margin-bottom: 1rem;">
              %s
            </ul>
          </div>

          <div style="margin: 1.25rem 0;">
            <div style="font-size: 0.8rem; text-transform: uppercase; color: var(--text-muted); font-weight: 700; margin-bottom: 0.4rem;">Key Component Files</div>
            <ul style="padding-left: 1.25rem; font-size: 0.875rem; margin-bottom: 1rem;">
              %s
            </ul>
          </div>

          <div style="margin: 1.25rem 0;">
            <div style="font-size: 0.8rem; text-transform: uppercase; color: var(--text-muted); font-weight: 700; margin-bottom: 0.4rem;">Inspect via CLI</div>
            <pre><code>%s</code></pre>
          </div>

          <div style="margin: 1.25rem 0;">
            <div style="font-size: 0.8rem; text-transform: uppercase; color: var(--text-muted); font-weight: 700; margin-bottom: 0.4rem;">Code Snippet</div>
            <pre><code>%s</code></pre>
          </div>

          <div class="tour-nav-bar">
            %s
            %s
          </div>
        </section>
`,
			s.Number,
			s.Number,
			s.Domain,
			s.Title,
			s.Description,
			invariantsList,
			filesList,
			s.SampleCmd,
			s.Snippet,
			func() string {
				if s.Number > 1 {
					return fmt.Sprintf(`<a href="#step-%d" class="btn-secondary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">&larr; Previous Step (%d)</a>`, s.Number-1, s.Number-1)
				}
				return `<div></div>`
			}(),
			func() string {
				if s.Number < len(steps) {
					return fmt.Sprintf(`<a href="#step-%d" class="btn-primary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">Next: %s &rarr;</a>`, s.Number+1, steps[s.Number].Title)
				}
				return `<a href="architecture/graph.html" class="btn-primary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">Explore Visual Graph &rarr;</a>`
			}(),
		)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Guided Architectural Tour — %s</title>
  <meta name="description" content="A curated topological walkthrough of Gira-G's architecture in 6 steps.">
  <link rel="stylesheet" href="assets/style.css">
</head>
<body>
  <header class="site-header">
    <div style="display: flex; align-items: center; gap: 1rem;">
      <a href="index.html" class="brand">
        <div class="brand-logo">G</div>
        <span>%s</span>
      </a>
      <span class="version-badge">%s</span>
    </div>
    <div class="header-center">
      <button id="search-btn" class="search-btn">
        <span>Quick search docs...</span>
        <span class="search-kbd">⌘K</span>
      </button>
    </div>
    <nav class="header-nav">
      <a href="intro.html" class="header-link">Docs</a>
      <a href="tour.html" class="header-link active">Guided Tour</a>
      <a href="architecture/index.html" class="header-link">Architecture</a>
      <a href="architecture/graph.html" class="header-link">Visual Explorer</a>
      <a href="cli/index.html" class="header-link">CLI</a>
      <a href="%s" target="_blank" rel="noopener" class="header-link">GitHub ↗</a>
    </nav>
  </header>

  <div class="layout-container">
    <aside class="site-sidebar">
      %s
    </aside>

    <main class="main-wrapper">
      <article class="content-article">
        <div class="breadcrumbs">
          <a href="index.html">Docs</a> &rsaquo; <span>Getting Started</span> &rsaquo; <span>Guided Tour</span>
        </div>

        <h1 style="border: none; margin-bottom: 0.5rem;">Guided Architectural Tour</h1>
        <p style="font-size: 1.15rem; color: #9ca3af; margin-bottom: 2.5rem;">
          Welcome to the topological walkthrough of Gira-G. This 6-step tour traverses the codebase 
          from entrypoints down to the storage and governance layers, highlighting design invariants 
          and key component implementations.
        </p>

        %s
      </article>

      <aside class="site-toc">
        <div class="toc-title">Tour Steps</div>
        <ul class="toc-list">
          <li class="toc-item level-2"><a href="#step-1" class="toc-link">1. Entrypoint & Dispatch</a></li>
          <li class="toc-item level-2"><a href="#step-2" class="toc-link">2. AST Parsing</a></li>
          <li class="toc-item level-2"><a href="#step-3" class="toc-link">3. MacroMap Store</a></li>
          <li class="toc-item level-2"><a href="#step-4" class="toc-link">4. Clustering</a></li>
          <li class="toc-item level-2"><a href="#step-5" class="toc-link">5. AST Manipulation</a></li>
          <li class="toc-item level-2"><a href="#step-6" class="toc-link">6. Governance & Docs</a></li>
        </ul>
      </aside>
    </main>
  </div>

  <div id="search-modal" class="search-modal">
    <div class="search-box">
      <div class="search-input-wrap">
        <input type="text" id="search-input" class="search-input" placeholder="Search documentation, symbols, commands...">
      </div>
      <div id="search-results" class="search-results"></div>
    </div>
  </div>

  <script src="assets/app.js"></script>
</body>
</html>`,
		config.Title,
		config.Title,
		config.Version,
		config.GitHubURL,
		navHTML,
		stepsHTML,
	)

	return html
}
