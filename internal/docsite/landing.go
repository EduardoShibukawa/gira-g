package docsite

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// GenerateLandingPageHTML renders a modern, human-centric, full-width landing page for index.html.
func GenerateLandingPageHTML(config *SiteConfig, navGroups []NavGroup) string {
	letter := logoLetter(config)
	heroTitle, heroDesc, terminalBody := buildHeroContent(config)
	mermaidSrc := buildMermaidArchitectureDiagram(config)

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
<body class="landing-body">
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
      <a href="intro.html" class="header-link">Docs</a>
      <a href="tour.html" class="header-link">Guided Tour</a>
      <a href="architecture/index.html" class="header-link">Architecture</a>
      <a href="architecture/graph.html" class="header-link">Visual Explorer</a>
      <a href="cli/index.html" class="header-link">CLI</a>
      <a href="%s" target="_blank" rel="noopener" class="header-link">GitHub ↗</a>
    </nav>
  </header>

  <main class="landing-main">
    <!-- Hero Section -->
    <section class="landing-hero">
      <div class="hero-badge">
        <span>⚡</span> The Architectural GPS &amp; Surgical Scalpel for Your Codebase
      </div>
      <h1 class="landing-title hero-title">%s</h1>
      <p class="landing-subtitle">
        %s
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
          <div class="terminal-title">Get Started in 3 Steps — Zero Egress, Local First</div>
        </div>
        <div class="terminal-body">
%s
        </div>
      </div>
    </section>

    <!-- 3 Real Scenarios in Your Daily Workflow -->
    <section class="landing-section">
      <div class="section-header-center">
        <h2 class="section-title">3 Real Scenarios in Your Daily Workflow</h2>
        <p class="section-desc">
          Instead of memorizing dozens of compiler commands, here is how Gira-G solves everyday engineering friction:
        </p>
      </div>

      <div class="scenarios-grid">
        <div class="scenario-card">
          <div class="scenario-badge">Scenario 1</div>
          <h3 class="scenario-title">Day 1 Onboarding</h3>
          <p class="scenario-text">
            Don't spend weeks reading 50,000 lines of unfamiliar code. Get an immediate mental model of key domains, 
            tour entrypoints, and query the architecture in plain language with <code>gira tour</code> and <code>gira ask</code>.
          </p>
          <div class="scenario-command">gira tour --format md</div>
        </div>

        <div class="scenario-card">
          <div class="scenario-badge">Scenario 2</div>
          <h3 class="scenario-title">Refactoring with Confidence</h3>
          <p class="scenario-text">
            Before touching shared code, calculate its exact blast radius across callers and downstream packages. 
            Know what will break before committing.
          </p>
          <div class="scenario-command">gira impact &lt;node_id&gt;</div>
        </div>

        <div class="scenario-card">
          <div class="scenario-badge">Scenario 3</div>
          <h3 class="scenario-title">AI Agent Pair Programming</h3>
          <p class="scenario-text">
            Give Cursor, Claude, or Hermes surgical AST slices instead of bloating context windows. 
            Save 85%% token budget with deterministic line locators.
          </p>
          <div class="scenario-command">gira ask "query" --format caveman</div>
        </div>
      </div>
    </section>

    <!-- The 5 Core Pillars -->
    <section class="landing-section">
      <div class="section-header-center">
        <h2 class="section-title">The 5 Core Pillars of Gira-G</h2>
        <p class="section-desc">
          Engineered to bridge deterministic code reality with human business intent.
        </p>
      </div>

      <div class="pillars-grid">
        <div class="pillar-card">
          <div class="pillar-icon">⚡</div>
          <h3 class="pillar-title">1. Deterministic Core</h3>
          <p class="pillar-text">
            Native Go AST analysis indexes symbols, callers, and contracts in &lt;5ms. Zero external dependencies, zero LLM hallucination risk.
          </p>
        </div>

        <div class="pillar-card">
          <div class="pillar-icon">🧠</div>
          <h3 class="pillar-title">2. Context Synthesis</h3>
          <p class="pillar-text">
            Ingests specifications, architecture decisions (ADRs), and business context to explain the structural "Why" behind the code.
          </p>
        </div>

        <div class="pillar-card">
          <div class="pillar-icon">⚖️</div>
          <h3 class="pillar-title">3. Drift Prevention</h3>
          <p class="pillar-text">
            Cryptographic SHA-256 Merkle proofs ensure architecture documentation never diverges from code reality. Strict CI gatekeeper (<code>gira gate-lint --strict</code>) blocks illegal cross-domain imports.
          </p>
        </div>

        <div class="pillar-card">
          <div class="pillar-icon">📚</div>
          <h3 class="pillar-title">4. Tri-Tier Documentation</h3>
          <p class="pillar-text">
            Living documentation tailored for every audience: public static portal for stakeholders, in-repo Markdown for developers, 
            and ultra-dense Caveman slices for AI agents.
          </p>
        </div>

        <div class="pillar-card">
          <div class="pillar-icon">🤖</div>
          <h3 class="pillar-title">5. Unified Human &amp; Agent Surface</h3>
          <p class="pillar-text">
            Natural language queries for humans via <code>gira ask</code> and standardized JSON-RPC MCP server for autonomous agents. 
            Both consume the exact same underlying MacroMap graph.
          </p>
        </div>
      </div>
    </section>

    <!-- Architecture Overview Diagram -->
    <section class="landing-section">
      <div class="section-header-center">
        <h2 class="section-title">Living Architecture Overview</h2>
        <p class="section-desc">
          Deterministic architectural topology extracted directly from this codebase.
        </p>
      </div>

      <div class="mermaid-container" style="max-width: 900px; margin: 0 auto;">
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
            <pre class="mermaid">%s</pre>
          </div>
        </div>
      </div>
    </section>

    <!-- Bottom CTA Banner -->
    <section class="landing-cta">
      <h2 style="font-size: 2rem; margin-top: 0; margin-bottom: 0.75rem;">Ready to understand your architecture?</h2>
      <p style="color: var(--text-muted); max-width: 600px; margin: 0 auto 1.5rem auto; font-size: 1.1rem;">
        Take the guided tour through key components, query your architecture in plain language, or launch the interactive visual explorer.
      </p>
      <div style="display: flex; justify-content: center; gap: 1rem; flex-wrap: wrap;">
        <a href="intro.html" class="btn-primary">🚀 Quickstart Guide</a>
        <a href="tour.html" class="btn-secondary">🧭 Guided Tour</a>
        <a href="architecture/graph.html" class="btn-secondary">🕸️ Open Visual Explorer</a>
        <a href="cli/index.html" class="btn-secondary">💻 CLI Reference</a>
      </div>
    </section>
  </main>

  <footer class="site-footer">
    <div class="footer-inner">
      <div>Gira-G Architecture Harness — Open Source under MIT License</div>
      <div style="margin-top: 0.5rem; color: var(--text-muted); font-size: 0.85rem;">
        Deterministic AST intelligence for humans and AI agents.
      </div>
    </div>
  </footer>

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
		letter,
		config.Title,
		config.Version,
		config.GitHubURL,
		heroTitle,
		heroDesc,
		terminalBody,
		mermaidSrc,
	)

	return html
}

func logoLetter(config *SiteConfig) string {
	if config != nil && config.LogoLetter != "" {
		return config.LogoLetter
	}
	return "G"
}

func buildHeroContent(config *SiteConfig) (string, string, string) {
	isDogfoodOrSelf := config == nil || config.TargetName == "" || config.TargetName == "gira-g" || strings.Contains(strings.ToLower(config.TargetName), "dogfood")
	if !isDogfoodOrSelf {
		title := config.Title
		if title == "" {
			title = config.TargetName
		}
		desc := config.Tagline
		if desc == "" {
			desc = "Living architecture and deterministic contract intelligence powered by Gira-G."
		} else {
			desc = fmt.Sprintf("%s. Living architecture and deterministic contract intelligence powered by Gira-G.", desc)
		}
		terminal := `          <div style="color: #6b7280;"># 1. Explore system entrypoints and architecture tour</div>
          <div><span style="color: #38bdf8;">gira</span> tour --format md</div>
          <br/>
          <div style="color: #6b7280;"># 2. Ask architectural questions in natural language</div>
          <div><span style="color: #38bdf8;">gira</span> ask &quot;How does this system operate?&quot;</div>
          <br/>
          <div style="color: #6b7280;"># 3. Verify architectural invariants and boundaries</div>
          <div><span style="color: #38bdf8;">gira</span> gate-lint --strict</div>`
		return title, desc, terminal
	}

	title := "Understand Any Codebase in 2 Minutes"
	desc := "Know what breaks before you touch a line of code, eliminate onboarding fatigue, and feed AI agents with zero-hallucination context. Gira-G unites deterministic AST reality with human business intent into a unified living architecture."
	terminal := `          <div style="color: #6b7280;"># 1. Install gira CLI binary</div>
          <div><span style="color: #38bdf8;">go</span> install github.com/EduardoShibukawa/gira-g/cmd/gira@latest</div>
          <br/>
          <div style="color: #6b7280;"># 2. Index your repository (sub-millisecond, zero-token local AST pass)</div>
          <div><span style="color: #38bdf8;">gira</span> init .</div>
          <br/>
          <div style="color: #6b7280;"># 3. Explore entrypoints or query architecture in plain language</div>
          <div><span style="color: #38bdf8;">gira</span> ask &quot;Qual o objetivo deste microsserviço?&quot; --format caveman</div>`
	return title, desc, terminal
}

func buildMermaidArchitectureDiagram(config *SiteConfig) string {
	isDogfoodOrSelf := config == nil || config.TargetName == "" || config.TargetName == "gira-g" || strings.Contains(strings.ToLower(config.TargetName), "dogfood")
	if isDogfoodOrSelf || config.MacroMap == nil || len(config.MacroMap.Nodes) == 0 {
		return `graph TD
  subgraph Entrypoints["Layer 1: Entrypoints"]
    CLI["CLI Interface (cmd/gira)"]
    MCP["MCP JSON-RPC Server (internal/mcp)"]
    ASK["Plain Language Ask (internal/ask)"]
  end

  subgraph Engine["Layer 2: Engine & Slices"]
    ENG["Slices Engine (internal/slices)"]
    ARCH["Architecture Engine (internal/architecture)"]
    JEV["Semantic Classifier (internal/jev)"]
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
  CLI --> ASK
  MCP --> ENG
  ASK --> JEV
  ASK --> DOM
  ENG --> DOM
  ARCH --> DOM
  EXT --> DOM
  ST --> DOM`
	}

	mm := config.MacroMap
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	groups := make(map[string][]domain.Node)
	for _, n := range mm.Nodes {
		grp := string(n.Kind)
		if len(n.Evidence) > 0 && n.Evidence[0].Locator.Path != "" {
			parts := strings.Split(filepath.ToSlash(n.Evidence[0].Locator.Path), "/")
			if len(parts) > 1 {
				grp = parts[0]
			}
		}
		groups[grp] = append(groups[grp], n)
	}

	groupKeys := make([]string, 0, len(groups))
	for k := range groups {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	sanitizedIDs := make(map[string]string)
	sanitizeID := func(raw string) string {
		if s, ok := sanitizedIDs[raw]; ok {
			return s
		}
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '_'
		}, raw)
		if len(clean) > 24 {
			clean = clean[:24]
		}
		sanitizedIDs[raw] = clean
		return clean
	}

	shownNodes := make(map[string]bool)
	maxGroups := 5
	if len(groupKeys) < maxGroups {
		maxGroups = len(groupKeys)
	}

	for i := 0; i < maxGroups; i++ {
		grp := groupKeys[i]
		nodes := groups[grp]
		grpID := sanitizeID("grp_" + grp)
		sb.WriteString(fmt.Sprintf("  subgraph %s[\"%s\"]\n", grpID, escapeMermaid(grp)))
		maxN := 4
		if len(nodes) < maxN {
			maxN = len(nodes)
		}
		for j := 0; j < maxN; j++ {
			node := nodes[j]
			nID := sanitizeID(node.ID)
			shownNodes[node.ID] = true
			label := node.Name
			if label == "" {
				label = node.ID
			}
			sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nID, escapeMermaid(label)))
		}
		sb.WriteString("  end\n\n")
	}

	edgeCount := 0
	for _, rel := range mm.Relationships {
		if shownNodes[rel.SourceID] && shownNodes[rel.TargetID] && rel.SourceID != rel.TargetID {
			sID := sanitizeID(rel.SourceID)
			tID := sanitizeID(rel.TargetID)
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", sID, tID))
			edgeCount++
			if edgeCount >= 15 {
				break
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

func escapeMermaid(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
