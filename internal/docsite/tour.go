package docsite

import (
	"fmt"
	"html"
	"strings"
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
	steps := generateTourSteps(config, navGroups)
	navHTML := renderSidebarNav(navGroups, "tour.html")
	headerNav := buildHeaderNav(config, navGroups)

	var stepsHTML string
	for _, s := range steps {
		invariantsList := ""
		for _, inv := range s.Invariants {
			invariantsList += fmt.Sprintf("<li><strong>%s</strong></li>", html.EscapeString(inv))
		}

		filesList := ""
		for _, f := range s.KeyFiles {
			filesList += fmt.Sprintf("<li><code>%s</code></li>", html.EscapeString(f))
		}

		stepsHTML += fmt.Sprintf(`
        <section id="step-%d" class="tour-card">
          <div class="tour-step-badge">Step %d of %d &bull; %s</div>
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
			len(steps),
			html.EscapeString(s.Domain),
			html.EscapeString(s.Title),
			html.EscapeString(s.Description),
			invariantsList,
			filesList,
			html.EscapeString(s.SampleCmd),
			html.EscapeString(s.Snippet),
			func() string {
				if s.Number > 1 {
					return fmt.Sprintf(`<a href="#step-%d" class="btn-secondary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">&larr; Previous Step (%d)</a>`, s.Number-1, s.Number-1)
				}
				return `<div></div>`
			}(),
			func() string {
				if s.Number < len(steps) {
					return fmt.Sprintf(`<a href="#step-%d" class="btn-primary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">Next: %s &rarr;</a>`, s.Number+1, html.EscapeString(steps[s.Number].Title))
				}
				return `<a href="architecture/graph.html" class="btn-primary" style="font-size: 0.85rem; padding: 0.4rem 0.85rem;">Explore Visual Graph &rarr;</a>`
			}(),
		)
	}

	var tocItems strings.Builder
	for _, s := range steps {
		tocItems.WriteString(fmt.Sprintf(`          <li class="toc-item level-2"><a href="#step-%d" class="toc-link">%d. %s</a></li>`+"\n", s.Number, s.Number, html.EscapeString(s.Title)))
	}

	target := siteTitle(config)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Guided Architectural Tour — %s</title>
  <meta name="description" content="A curated topological walkthrough of %s's architecture in %d steps.">
  <link rel="stylesheet" href="assets/style.css">
</head>
<body>
  <header class="site-header">
    <div style="display: flex; align-items: center; gap: 1rem;">
      <a href="index.html" class="brand">
        <div class="brand-logo">%s</div>
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
%s
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
          Welcome to the topological walkthrough of %s. This %d-step tour traverses the codebase 
          from entrypoints down to the storage and verification layers, highlighting design invariants 
          and key component implementations.
        </p>

        %s
      </article>

      <aside class="site-toc">
        <div class="toc-title">Tour Steps</div>
        <ul class="toc-list">
%s
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
		html.EscapeString(target),
		html.EscapeString(target),
		len(steps),
		logoLetter(config),
		html.EscapeString(target),
		config.Version,
		headerNav,
		navHTML,
		html.EscapeString(target),
		len(steps),
		stepsHTML,
		tocItems.String(),
	)

	return html
}

func generateTourSteps(config *SiteConfig, navGroups []NavGroup) []TourStep {
	// Configured tour from the target repo's `docsite:` section in gira.yaml.
	if config != nil && config.Docsite != nil && len(config.Docsite.Tour) > 0 {
		steps := make([]TourStep, 0, len(config.Docsite.Tour))
		for i, s := range config.Docsite.Tour {
			steps = append(steps, TourStep{
				Number:      i + 1,
				Title:       s.Title,
				Domain:      s.Domain,
				Description: s.Description,
				Invariants:  s.Invariants,
				KeyFiles:    s.KeyFiles,
				SampleCmd:   s.SampleCmd,
				Snippet:     s.Snippet,
			})
		}
		return steps
	}

	if isDogfoodOrSelf(config) {
		return []TourStep{
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
	}

	target := siteTitle(config)

	// Generic external project tour steps
	return []TourStep{
		{
			Number:      1,
			Title:       "System Overview & Entrypoints",
			Domain:      "Getting Started",
			Description: fmt.Sprintf("Entrypoints and foundational architecture of %s. Inspect primary manifests and runtime configurations.", target),
			Invariants: []string{
				"Deterministic Initialization: System entrypoints enforce contract validation upon boot.",
				"Explicit Scope: Project manifests define pinned dependencies and boundary scopes.",
			},
			KeyFiles: []string{
				"README.md",
				"gira.yaml",
			},
			SampleCmd: "cat README.md",
			Snippet: fmt.Sprintf(`# Review onboarding guide for %s
cat README.md`, target),
		},
		{
			Number:      2,
			Title:       "Core Domain Logic & Services",
			Domain:      "Core Engine",
			Description: fmt.Sprintf("The primary domain components and services responsible for business logic in %s.", target),
			Invariants: []string{
				"Modularity: Domain services preserve bounded responsibilities and clean interfaces.",
				"Explicit Contracts: State transitions are verified against defined data contracts.",
			},
			KeyFiles: []string{
				"docs/intro.md",
			},
			SampleCmd: "cat docs/intro.md",
			Snippet: `# Inspect core concepts and architecture
cat docs/intro.md`,
		},
		{
			Number:      3,
			Title:       "Quality Gates & Verification",
			Domain:      "Verification & Quality",
			Description: fmt.Sprintf("Continuous verification rules, automated test harnesses, and contract gates across %s.", target),
			Invariants: []string{
				"Fail-Closed CI: Any invariant violation or schema drift halts the pipeline immediately.",
				"Verifiable Proofs: Tests validate critical paths prior to production merging.",
			},
			KeyFiles: []string{
				".github/workflows",
			},
			SampleCmd: "ls .github/workflows",
			Snippet: `# Inspect CI automation workflows
ls .github/workflows`,
		},
		{
			Number:      4,
			Title:       "Living Architecture & Visual Graph",
			Domain:      "Living Architecture",
			Description: fmt.Sprintf("Continuous topological graph and visual dependency mapping for %s powered by Gira-G.", target),
			Invariants: []string{
				"Zero Drift: Architecture diagrams stay in sync with underlying code facts.",
				"Interactive Exploration: Visual graph surfaces coupling and centrality metrics.",
			},
			KeyFiles: []string{
				"build/architecture/graph.html",
				"build/architecture/index.html",
			},
			SampleCmd: "open build/architecture/graph.html",
			Snippet: `# Launch interactive visual architecture explorer
open build/architecture/graph.html`,
		},
	}
}
