package docsite

import (
	"fmt"
)

// GenerateLandingPageHTML renders a modern, human-centric landing page for index.html.
func GenerateLandingPageHTML(config *SiteConfig, navGroups []NavGroup) string {
	navHTML := renderSidebarNav(navGroups, "index.html")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s — %s</title>
  <meta name="description" content="%s">
  <link rel="stylesheet" href="assets/style.css">
  <script src="https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js"></script>
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
      <a href="tour.html" class="header-link">Guided Tour</a>
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
        <!-- Hero Section -->
        <div class="hero-container">
          <div class="hero-badge">
            <span>⚡</span> AST-Grounded Code Intelligence & Living Architecture
          </div>
          <h1 class="hero-title">Deterministic Architecture Intelligence</h1>
          <p class="hero-desc">
            Transform monolithic codebases into queryable AST graphs, living architectural contracts, 
            and surgical editing interfaces. Designed for sub-millisecond execution, zero hallucinations, 
            and complete parity between software engineers and autonomous AI agents.
          </p>
          <div class="hero-actions">
            <a href="intro.html" class="btn-primary">🚀 Get Started</a>
            <a href="tour.html" class="btn-secondary">🧭 Guided Tour</a>
            <a href="architecture/graph.html" class="btn-secondary">🕸️ Visual Explorer</a>
            <a href="guides/change-maker-workflow.html" class="btn-secondary">📖 Developer Workflows</a>
          </div>

          <!-- Quickstart Terminal Card -->
          <div class="terminal-card">
            <div class="terminal-header">
              <div class="terminal-dot dot-red"></div>
              <div class="terminal-dot dot-yellow"></div>
              <div class="terminal-dot dot-green"></div>
              <div class="terminal-title">Quickstart in 4 Steps — Terminal</div>
            </div>
            <div class="terminal-body">
              <div style="color: #6b7280;"># 1. Install gira CLI binary</div>
              <div><span style="color: #38bdf8;">go</span> install github.com/EduardoShibukawa/gira-g/cmd/gira@latest</div>
              <br/>
              <div style="color: #6b7280;"># 2. Initialize and index your repository</div>
              <div><span style="color: #38bdf8;">gira</span> init .</div>
              <br/>
              <div style="color: #6b7280;"># 3. Read symbol snippet or inspect impact deterministically</div>
              <div><span style="color: #38bdf8;">gira</span> read cmd/gira/main.go main</div>
              <br/>
              <div style="color: #6b7280;"># 4. Generate living documentation suite & interactive Cytoscape graph</div>
              <div><span style="color: #38bdf8;">gira</span> docs --output ./build</div>
            </div>
          </div>
        </div>

        <hr/>

        <!-- The Surgical Scalpel Mental Model -->
        <h2 class="section-heading-center">The "Surgical Scalpel" Paradigm</h2>
        <p class="section-subheading-center">
          Traditional tools dump whole files into LLM prompts or rely on imprecise regex heuristics. 
          Gira-G provides exact surgical precision across three distinct phases.
        </p>

        <div class="scalpel-steps">
          <div class="scalpel-step">
            <div class="scalpel-num">Phase 1</div>
            <div class="scalpel-title">Deterministic AST Extraction</div>
            <p class="scalpel-desc">
              Native Tree-sitter and Go AST parsers extract symbols, call hierarchies, and cross-file locators without relying on language servers.
            </p>
          </div>
          <div class="scalpel-step">
            <div class="scalpel-num">Phase 2</div>
            <div class="scalpel-title">Architectural Clustering</div>
            <p class="scalpel-desc">
              Import centrality scoring and graph algorithms partition packages into 6 resilient architectural domains with enforced invariants.
            </p>
          </div>
          <div class="scalpel-step">
            <div class="scalpel-num">Phase 3</div>
            <div class="scalpel-title">Surgical Action & Verification</div>
            <p class="scalpel-desc">
              Subagent context budgets, AST modifications preserving code indentation, and automated quality gates guarantee zero regressions.
            </p>
          </div>
        </div>

        <hr/>

        <!-- 4 Core Pillars -->
        <h2 class="section-heading-center">Architectural Pillars</h2>
        <p class="section-subheading-center">
          Engineered for cognitive density, machine readability, and navigational clarity.
        </p>

        <div class="pillar-grid">
          <div class="pillar-card">
            <div class="pillar-icon">⚡</div>
            <div class="pillar-title">Zero-Token Fast Pass</div>
            <p class="pillar-desc">
              Resolve dependencies, impact radius, callers, and code slices in under 5 milliseconds without expending LLM tokens.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">🕸️</div>
            <div class="pillar-title">Interactive Visual Explorer</div>
            <p class="pillar-desc">
              Cytoscape spring-physics engine visualizing package communities, import topology, and real-time caller/callee inspection drawers.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">🏛️</div>
            <div class="pillar-title">Living Architecture Suite</div>
            <p class="pillar-desc">
              Self-updating Mermaid sequence flows and dependency diagrams with responsive viewports, zoom controls, and Fullscreen Lightbox modals.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">🤖</div>
            <div class="pillar-title">Dual-Surface Agent Parity</div>
            <p class="pillar-desc">
              Human CLI and AI Model Context Protocol (MCP) share the identical AST facts, token budgets, and governance gate validations.
            </p>
          </div>
        </div>

        <hr/>

        <!-- Living Architecture Preview -->
        <h2>Living Architecture Overview</h2>
        <p>
          Gira-G models systems through strictly decoupled layers. The diagram below represents the active 
          runtime topology extracted directly from this repository:
        </p>

        <div class="mermaid-wrapper" data-mermaid-src="graph TD
  subgraph Entrypoints[&quot;Layer 1: Entrypoints&quot;]
    CLI[&quot;CLI Interface (cmd/gira)&quot;]
    MCP[&quot;MCP JSON-RPC Server (internal/mcp)&quot;]
  end

  subgraph Engine[&quot;Layer 2: Engine &amp; Slices&quot;]
    ENG[&quot;Slices Engine (internal/slices)&quot;]
    ARCH[&quot;Architecture Engine (internal/architecture)&quot;]
  end

  subgraph AST[&quot;Layer 3: AST &amp; Extraction&quot;]
    MANIP[&quot;AST Manipulator (internal/manipulator)&quot;]
    EXT[&quot;Language Extractors (internal/extractors)&quot;]
  end

  subgraph Store[&quot;Layer 4: Storage &amp; Persistence&quot;]
    DOM[&quot;MacroMap Graph (internal/domain)&quot;]
    ST[&quot;Deterministic Store (internal/store)&quot;]
  end

  CLI --&gt; ENG
  CLI --&gt; MANIP
  MCP --&gt; ENG
  ENG --&gt; DOM
  ARCH --&gt; DOM
  EXT --&gt; DOM
  ST --&gt; DOM">
          <div class="mermaid-toolbar">
            <span class="mermaid-badge">Architecture Diagram</span>
            <div class="mermaid-actions">
              <button class="mermaid-btn mermaid-fullscreen-btn" title="Fullscreen Lightbox">⛶ Fullscreen</button>
              <button class="mermaid-btn mermaid-zoom-in" title="Zoom In">+</button>
              <button class="mermaid-btn mermaid-zoom-out" title="Zoom Out">-</button>
              <button class="mermaid-btn mermaid-zoom-reset" title="Reset Zoom">↺</button>
              <button class="mermaid-btn mermaid-copy-btn" title="Copy Mermaid Source">📋 Copy</button>
            </div>
          </div>
          <div class="mermaid-viewport">
            <div class="mermaid-inner">
              <pre class="mermaid">graph TD
  subgraph Entrypoints["Layer 1: Entrypoints"]
    CLI["CLI Interface (cmd/gira)"]
    MCP["MCP JSON-RPC Server (internal/mcp)"]
  end

  subgraph Engine["Layer 2: Engine & Slices"]
    ENG["Slices Engine (internal/slices)"]
    ARCH["Architecture Engine (internal/architecture)"]
  end

  subgraph AST["Layer 3: AST & Extraction"]
    MANIP["AST Manipulator (internal/manipulator)"]
    EXT["Language Extractors (internal/extractors)"]
  end

  subgraph Store["Layer 4: Storage & Persistence"]
    DOM["MacroMap Graph (internal/domain)"]
    ST["Deterministic Store (internal/store)"]
  end

  CLI --> ENG
  CLI --> MANIP
  MCP --> ENG
  ENG --> DOM
  ARCH --> DOM
  EXT --> DOM
  ST --> DOM</pre>
            </div>
          </div>
        </div>

        <div style="margin-top: 3rem; padding: 2rem; background: var(--bg-secondary); border-radius: 12px; border: 1px solid var(--border-color); text-align: center;">
          <h3 style="margin-top: 0; margin-bottom: 0.75rem;">Ready to explore?</h3>
          <p style="color: var(--text-muted); margin-bottom: 1.5rem;">
            Take the guided tour through the 6 core architectural stages or launch the interactive Visual Explorer.
          </p>
          <div style="display: flex; justify-content: center; gap: 1rem; flex-wrap: wrap;">
            <a href="tour.html" class="btn-primary">🧭 Start Guided Tour</a>
            <a href="architecture/graph.html" class="btn-secondary">🕸️ Open Visual Explorer</a>
            <a href="intro.html" class="btn-secondary">📖 Read Documentation</a>
          </div>
        </div>
      </article>

      <aside class="site-toc">
        <div class="toc-title">On this page</div>
        <ul class="toc-list">
          <li class="toc-item level-2"><a href="#the-surgical-scalpel-paradigm" class="toc-link">The Surgical Scalpel</a></li>
          <li class="toc-item level-2"><a href="#architectural-pillars" class="toc-link">Architectural Pillars</a></li>
          <li class="toc-item level-2"><a href="#living-architecture-overview" class="toc-link">Living Architecture</a></li>
          <li class="toc-item level-2"><a href="#ready-to-explore" class="toc-link">Next Steps</a></li>
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
		config.Tagline,
		config.Tagline,
		config.Title,
		config.Version,
		config.GitHubURL,
		navHTML,
	)

	return html
}
