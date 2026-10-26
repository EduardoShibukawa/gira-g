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
			Title:       "Repository Onboarding & AST Indexing",
			Domain:      "Getting Started",
			Description: "Initialize Gira-G on your codebase. It scans Go modules, parses AST symbols deterministically, and creates the pinned manifest and Macro Map knowledge graph in milliseconds without requiring language servers.",
			Invariants: []string{
				"Zero Probabilistic Guessing: All symbol bounds and package relations are exact byte locators.",
				"Sub-150ms Cold Execution: Deterministic scanning runs natively without JVM or background daemons.",
			},
			KeyFiles: []string{
				"gira.yaml",
				".gira/macro-map.json",
			},
			SampleCmd: "gira init .",
			Snippet: `# Initialize and index the current project
gira init .`,
		},
		{
			Number:      2,
			Title:       "Surgical Symbol Reading & AST Inspection",
			Domain:      "Deterministic Code Reading",
			Description: "Read exact functions, interfaces, or structs from your codebase without dumping whole files into prompts. Gira extracts precise AST boundaries with surrounding documentation.",
			Invariants: []string{
				"Token-Efficient Extraction: Return only the target AST declaration rather than multi-thousand-line source files.",
				"Line & Byte Fidelity: Exact coordinates preserved for automated tooling and human inspection.",
			},
			KeyFiles: []string{
				"cmd/gira/main.go",
				"internal/domain/models.go",
			},
			SampleCmd: "gira read cmd/gira/main.go main",
			Snippet: `# Read the exact 'main' function declaration
gira read cmd/gira/main.go main`,
		},
		{
			Number:      3,
			Title:       "Blast Radius & Impact Analysis",
			Domain:      "Change Safety",
			Description: "Before modifying a component, query its downstream consumers and callers. Gira calculates transitive call graphs and contractual impact briefings in <5ms.",
			Invariants: []string{
				"Transitive Graph Traversal: Surface callers, callees, and cross-domain dependencies before making edits.",
				"Zero Token Overhead: Graph queries resolve against the pre-computed in-memory Macro Map in <5ms.",
			},
			KeyFiles: []string{
				".gira/macro-map.json",
			},
			SampleCmd: "gira impact <node_id> --depth 2",
			Snippet: `# Discover indexed nodes with 'gira tour', then inspect blast radius
gira impact <node_id> --depth 2`,
		},
		{
			Number:      4,
			Title:       "Bounded Context Slicing",
			Domain:      "Agent & Human Context",
			Description: "Extract bounded context slices with strict token budgets (summary, contracts, or deep) to fit directly into AI agent prompts or developer reviews without context window overflows.",
			Invariants: []string{
				"Budget Enforcement: Slices return token-bounded envelopes to prevent LLM context overflows.",
				"Multi-Tier Views: Choose between summary, contracts, or full architectural views.",
			},
			KeyFiles: []string{
				"internal/slices/engine.go",
			},
			SampleCmd: "gira slice service:storage --tier contracts --budget 1000",
			Snippet: `# Extract contract-tier slice within 1000 tokens
gira slice service:storage --tier contracts --budget 1000`,
		},
		{
			Number:      5,
			Title:       "Architecture Governance & CI Gates",
			Domain:      "Drift Prevention",
			Description: "Enforce architectural contracts in your CI pipeline. Gira detects unauthorized cross-domain coupling, circular dependencies, and symbol drifts before pull requests merge.",
			Invariants: []string{
				"Fail-Closed CI Verification: Architectural boundary breaches exit with non-zero status in CI.",
				"Diff Verification: Compare Macro Maps across git commits to prevent structural decay.",
			},
			KeyFiles: []string{
				"gira.yaml",
				".github/workflows/ci.yml",
			},
			SampleCmd: "gira gate-lint --strict",
			Snippet: `# Enforce architecture contracts and fail CI on drift
gira gate-lint --strict`,
		},
		{
			Number:      6,
			Title:       "Living Documentation & Interactive Visuals",
			Domain:      "Living Architecture",
			Description: "Generate self-updating documentation with interactive Cytoscape dependency graphs, zoomable Mermaid diagrams, and searchable CLI reference.",
			Invariants: []string{
				"Self-Updating Parity: Documentation regenerates automatically from code facts without manual wiki syncing.",
				"Interactive Visual Exploration: Explore community clustering and cross-domain relations in Cytoscape.",
			},
			KeyFiles: []string{
				"build/index.html",
				"build/architecture/graph.html",
			},
			SampleCmd: "gira docs --output ./build",
			Snippet: `# Build full static documentation site with interactive explorer
gira docs --output ./build`,
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
