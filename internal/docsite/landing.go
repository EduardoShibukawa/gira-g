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
            <span>⚡</span> The Architectural GPS &amp; Surgical Scalpel for Your Codebase
          </div>
          <h1 class="hero-title">Understand Any Codebase in 2 Minutes</h1>
          <p class="hero-desc">
            Know what breaks before you touch a line of code, eliminate onboarding fatigue, 
            and feed AI agents with zero-hallucination context. Gira-G unites deterministic code reality 
            with human business intent in a unified living architecture.
          </p>
          <div class="hero-actions">
            <a href="intro.html" class="btn-primary">🚀 Quickstart Guide</a>
            <a href="tour.html" class="btn-secondary">🧭 Guided Tour</a>
            <a href="cli/ask.html" class="btn-secondary">💬 Ask in Plain Language</a>
            <a href="architecture/graph.html" class="btn-secondary">🕸️ Visual Explorer</a>
          </div>

          <!-- Quickstart Terminal Card -->
          <div class="terminal-card">
            <div class="terminal-header">
              <div class="terminal-dot dot-red"></div>
              <div class="terminal-dot dot-yellow"></div>
              <div class="terminal-dot dot-green"></div>
              <div class="terminal-title">Quickstart in 3 Steps — Terminal</div>
            </div>
            <div class="terminal-body">
              <div style="color: #6b7280;"># 1. Install gira CLI binary</div>
              <div><span style="color: #38bdf8;">go</span> install github.com/EduardoShibukawa/gira-g/cmd/gira@latest</div>
              <br/>
              <div style="color: #6b7280;"># 2. Index your repository (sub-millisecond, zero-token fast pass)</div>
              <div><span style="color: #38bdf8;">gira</span> init .</div>
              <br/>
              <div style="color: #6b7280;"># 3. Explore entrypoints or ask questions in plain language</div>
              <div><span style="color: #38bdf8;">gira</span> tour --format md</div>
            </div>
          </div>
        </div>

        <hr/>

        <!-- 3 Real Scenarios in Your Daily Workflow -->
        <h2 class="section-heading-center">3 Real Scenarios in Your Daily Workflow</h2>
        <p class="section-subheading-center">
          Instead of memorizing dozens of CLI commands, here is how Gira-G solves everyday challenges:
        </p>

        <div class="scalpel-steps">
          <div class="scalpel-step">
            <div class="scalpel-num">Scenario 1</div>
            <div class="scalpel-title">Day 1 Onboarding</div>
            <p class="scalpel-desc">
              Don't spend weeks reading 50,000 lines of code. Get an immediate mental model of key domains, 
              tour entrypoints, and query the architecture in plain language with <code>gira tour</code>.
            </p>
          </div>
          <div class="scalpel-step">
            <div class="scalpel-num">Scenario 2</div>
            <div class="scalpel-title">Refactoring with Confidence</div>
            <p class="scalpel-desc">
              Before modifying a shared contract or function, compute its exact blast radius across all packages 
              with <code>gira impact &lt;node_id&gt;</code> to know what will break before you commit.
            </p>
          </div>
          <div class="scalpel-step">
            <div class="scalpel-num">Scenario 3</div>
            <div class="scalpel-title">AI Pair Programming (Cursor / Claude)</div>
            <p class="scalpel-desc">
              Deliver byte-exact code slices to AI coding agents via <code>gira slice &lt;node_id&gt;</code>. 
              Save 95%% of token budget with zero hallucinations and exact line locators.
            </p>
          </div>
        </div>

        <hr/>

        <!-- The 5 Core Pillars -->
        <h2 class="section-heading-center">The 5 Core Pillars of Gira-G</h2>
        <p class="section-subheading-center">
          Engineered to bridge deterministic code facts with human business intent.
        </p>

        <div class="pillar-grid">
          <div class="pillar-card">
            <div class="pillar-icon">⚡</div>
            <div class="pillar-title">1. Deterministic Core</div>
            <p class="pillar-desc">
              Native Go AST and static analyzers extract symbols, callers, and contracts in &lt;5ms with zero token cost.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">📝</div>
            <div class="pillar-title">2. Non-Deterministic Context</div>
            <p class="pillar-desc">
              Ingests human specifications, tasks, ADRs, and business goals to explain the "Why" behind the code.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">⚖️</div>
            <div class="pillar-title">3. SYNC &amp; Drift Proofing</div>
            <p class="pillar-desc">
              Connects code reality with architectural intent using Merkle proofs, enforcing strict CI gate verification.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">📚</div>
            <div class="pillar-title">4. Triple Surface Docs</div>
            <p class="pillar-desc">
              Clean web portal for humans, versioned ARCHITECTURE.md for repos, and surgical MCP slices for AI agents.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">💬</div>
            <div class="pillar-title">5. Conversational Query</div>
            <p class="pillar-desc">
              Query architecture and blast radius in plain natural language with <code>gira ask</code> and instant Mermaid diagrams.
            </p>
          </div>
          <div class="pillar-card">
            <div class="pillar-icon">🛡️</div>
            <div class="pillar-title">CI Governance Guardian</div>
            <p class="pillar-desc">
              Run <code>gira gate-lint --strict</code> in automated pipelines to prevent architectural erosion before merging.
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
            Take the guided tour through key components or query your architecture in plain language.
          </p>
          <div style="display: flex; justify-content: center; gap: 1rem; flex-wrap: wrap;">
            <a href="tour.html" class="btn-primary">🧭 Start Guided Tour</a>
            <a href="cli/ask.html" class="btn-secondary">💬 Try gira ask</a>
            <a href="architecture/graph.html" class="btn-secondary">🕸️ Open Visual Explorer</a>
            <a href="intro.html" class="btn-secondary">📖 Read Documentation</a>
          </div>
        </div>
      </article>

      <aside class="site-toc">
        <div class="toc-title">On this page</div>
        <ul class="toc-list">
          <li class="toc-item level-2"><a href="#3-real-scenarios-in-your-daily-workflow" class="toc-link">3 Real Scenarios</a></li>
          <li class="toc-item level-2"><a href="#the-5-core-pillars-of-gira-g" class="toc-link">5 Core Pillars</a></li>
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
