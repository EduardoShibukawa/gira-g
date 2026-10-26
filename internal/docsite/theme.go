package docsite

// CSSBundle contains modern, responsive styling for the documentation site.
const CSSBundle = `
:root {
  --bg-primary: #0b0f19;
  --bg-secondary: #111827;
  --bg-card: #1f2937;
  --bg-code: #030712;
  --border-color: #374151;
  --border-subtle: #1f2937;
  --text-primary: #f9fafb;
  --text-secondary: #9ca3af;
  --text-muted: #6b7280;
  --accent: #38bdf8;
  --accent-hover: #0ea5e9;
  --accent-bg: rgba(56, 189, 248, 0.1);
  --success: #34d399;
  --warning: #fbbf24;
  --danger: #f87171;
  --font-sans: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  --sidebar-w: 280px;
  --toc-w: 240px;
  --header-h: 64px;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: var(--font-sans);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  line-height: 1.65;
  -webkit-font-smoothing: antialiased;
}

a { color: var(--accent); text-decoration: none; transition: color 0.15s ease; }
a:hover { color: var(--accent-hover); text-decoration: underline; }

/* Header */
.site-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--header-h);
  background: rgba(11, 15, 25, 0.9);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  z-index: 50;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-primary);
  text-decoration: none;
}
.brand:hover { text-decoration: none; }
.brand-logo {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #38bdf8, #6366f1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 800;
  font-size: 1.1rem;
  box-shadow: 0 2px 8px rgba(56, 189, 248, 0.3);
}
.version-badge {
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--accent-bg);
  color: var(--accent);
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.3);
}

.header-center {
  flex: 1;
  max-width: 480px;
  margin: 0 2rem;
}
.search-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}
.search-btn:hover {
  border-color: var(--accent);
  color: var(--text-primary);
}
.search-kbd {
  background: var(--bg-card);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  font-size: 0.75rem;
  border: 1px solid var(--border-color);
  color: var(--text-muted);
}

.header-nav {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}
.header-link {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-weight: 500;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.header-link:hover { color: var(--accent); text-decoration: none; }
.header-link.active { color: var(--accent); }

/* Layout */
.layout-container {
  display: flex;
  margin-top: var(--header-h);
  min-height: calc(100vh - var(--header-h));
}

/* Sidebar */
.site-sidebar {
  width: var(--sidebar-w);
  position: fixed;
  top: var(--header-h);
  bottom: 0;
  left: 0;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  padding: 1.5rem 1rem 3rem 1.25rem;
  z-index: 40;
}
.nav-group {
  margin-bottom: 1.5rem;
}
.nav-group-collapsible {
  margin-bottom: 1.25rem;
}
.nav-group-summary {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  padding: 0.3rem 0.6rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  border-radius: 4px;
  transition: background 0.15s ease, color 0.15s ease;
}
.nav-group-summary:hover {
  color: var(--text-primary);
  background: var(--bg-secondary);
}
.nav-group-chevron {
  font-size: 0.65rem;
  transition: transform 0.2s ease;
}
.nav-group-collapsible[open] .nav-group-chevron {
  transform: rotate(90deg);
}
.nav-group-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  padding-left: 0.6rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.nav-list {
  list-style: none;
  padding-left: 0.25rem;
}
.nav-item a {
  display: flex;
  align-items: center;
  padding: 0.4rem 0.6rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
  border-radius: 6px;
  text-decoration: none;
  transition: all 0.15s ease;
}
.nav-item a:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
  text-decoration: none;
}
.nav-item.active a {
  background: var(--accent-bg);
  color: var(--accent);
  font-weight: 600;
}

/* Main Content Area */
.main-wrapper {
  margin-left: var(--sidebar-w);
  flex: 1;
  display: flex;
  justify-content: center;
  padding: 2.5rem 2rem 4rem 2rem;
  min-width: 0;
}
.content-article {
  max-width: 860px;
  width: 100%;
  min-width: 0;
}

/* Breadcrumbs */
.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}
.breadcrumbs a { color: var(--text-secondary); text-decoration: none; }
.breadcrumbs a:hover { color: var(--accent); }

/* Typography */
h1, h2, h3, h4, h5, h6 {
  color: var(--text-primary);
  font-weight: 700;
  line-height: 1.3;
  margin-top: 2rem;
  margin-bottom: 0.85rem;
  scroll-margin-top: calc(var(--header-h) + 1.5rem);
}
h1 { font-size: 2.25rem; margin-top: 0; margin-bottom: 1rem; border-bottom: 1px solid var(--border-color); padding-bottom: 0.75rem; }
h2 { font-size: 1.5rem; border-bottom: 1px solid var(--border-subtle); padding-bottom: 0.5rem; }
h3 { font-size: 1.25rem; }
h4 { font-size: 1.05rem; }

p { margin-bottom: 1.15rem; color: #d1d5db; font-size: 0.975rem; }
ul, ol { margin-bottom: 1.15rem; padding-left: 1.5rem; color: #d1d5db; }
li { margin-bottom: 0.35rem; }

hr {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 2rem 0;
}

/* Tables & Wrappers */
.table-wrapper {
  width: 100%;
  overflow-x: auto;
  margin: 1.75rem 0;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
  margin: 0;
}
th, td {
  padding: 0.75rem 1rem;
  border: 1px solid var(--border-color);
  text-align: left;
}
th {
  background: #1f2937;
  color: var(--text-primary);
  font-weight: 600;
}
tr:nth-child(even) {
  background: rgba(31, 41, 55, 0.35);
}

/* Code & Pre */
pre {
  background: var(--bg-code);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.1rem;
  overflow-x: auto;
  margin: 1.25rem 0;
  position: relative;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  line-height: 1.55;
}
code {
  font-family: var(--font-mono);
  font-size: 0.875em;
  background: rgba(55, 65, 81, 0.45);
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  color: #38bdf8;
}
pre code {
  background: transparent;
  padding: 0;
  border-radius: 0;
  color: #e5e7eb;
}
.copy-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  border-radius: 4px;
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
pre:hover .copy-btn { opacity: 1; }
.copy-btn:hover { background: var(--bg-secondary); color: #fff; }

/* Callouts / Alerts */
.callout {
  border-left: 4px solid var(--accent);
  background: var(--bg-secondary);
  border-radius: 0 8px 8px 0;
  padding: 1rem 1.25rem;
  margin: 1.5rem 0;
}
.callout-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}
.callout-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
}
.callout-body {
  font-size: 0.925rem;
  color: #d1d5db;
}
.callout-note { border-left-color: #38bdf8; }
.callout-note .callout-header { color: #38bdf8; }
.callout-tip { border-left-color: #34d399; }
.callout-tip .callout-header { color: #34d399; }
.callout-important { border-left-color: #a855f7; }
.callout-important .callout-header { color: #a855f7; }
.callout-warning { border-left-color: #fbbf24; }
.callout-warning .callout-header { color: #fbbf24; }
.callout-caution { border-left-color: #f87171; }
.callout-caution .callout-header { color: #f87171; }

/* Mermaid Diagrams & Interactive Toolbar */
.mermaid-wrapper {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin: 2rem 0;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}
.mermaid-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.9rem;
  background: rgba(31, 41, 55, 0.7);
  border-bottom: 1px solid var(--border-color);
  font-size: 0.8rem;
}
.mermaid-badge {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  color: var(--accent);
  text-transform: uppercase;
  font-size: 0.725rem;
  letter-spacing: 0.05em;
}
.mermaid-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.mermaid-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  border-radius: 6px;
  font-size: 0.75rem;
  padding: 0.25rem 0.55rem;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  transition: all 0.15s ease;
  user-select: none;
}
.mermaid-btn:hover {
  background: var(--accent-bg);
  color: var(--accent);
  border-color: var(--accent);
}
.mermaid-viewport {
  overflow-x: auto;
  overflow-y: hidden;
  padding: 2rem 1.5rem;
  width: 100%;
  display: block;
  background: #0f172a;
  text-align: left;
}
.mermaid-inner {
  display: inline-block;
  min-width: 880px;
  transform-origin: top left;
  transition: transform 0.15s ease-out;
}
.mermaid-inner svg {
  max-width: none !important;
  height: auto !important;
  display: block;
}

/* Lightbox Modal */
.mermaid-lightbox {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(3, 7, 18, 0.96);
  backdrop-filter: blur(12px);
  z-index: 200;
  flex-direction: column;
}
.mermaid-lightbox.open {
  display: flex;
}
.lightbox-header {
  height: 56px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  background: rgba(17, 24, 39, 0.9);
}
.lightbox-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.lightbox-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.lightbox-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0 0.5rem;
  line-height: 1;
}
.lightbox-close:hover {
  color: #fff;
}
.lightbox-body {
  flex: 1;
  overflow: auto;
  padding: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
.lightbox-content {
  transform-origin: center center;
  transition: transform 0.15s ease-out;
}
.lightbox-content svg {
  max-width: none !important;
  height: auto !important;
}

/* Right TOC */
.site-toc {
  width: var(--toc-w);
  position: sticky;
  top: calc(var(--header-h) + 2rem);
  height: calc(100vh - var(--header-h) - 4rem);
  overflow-y: auto;
  padding-left: 1.5rem;
  font-size: 0.85rem;
  display: none;
}
@media (min-width: 1280px) {
  .site-toc { display: block; }
}
.toc-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}
.toc-list { list-style: none; }
.toc-item { margin-bottom: 0.45rem; }
.toc-item.level-3 { padding-left: 0.75rem; font-size: 0.8rem; }
.toc-link { color: var(--text-muted); text-decoration: none; transition: color 0.15s; }
.toc-link:hover, .toc-link.active { color: var(--accent); text-decoration: none; }

/* Prev / Next Pagination */
.pagination-nav {
  display: flex;
  justify-content: space-between;
  margin-top: 3.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}
.pagination-link {
  display: flex;
  flex-direction: column;
  text-decoration: none;
}
.pagination-link:hover { text-decoration: none; }
.pagination-sub { font-size: 0.75rem; text-transform: uppercase; color: var(--text-muted); }
.pagination-title { font-size: 1rem; font-weight: 600; color: var(--accent); }
.pagination-next { margin-left: auto; text-align: right; }

/* Search Modal */
.search-modal {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(6px);
  z-index: 100;
  align-items: flex-start;
  justify-content: center;
  padding-top: 15vh;
}
.search-modal.open { display: flex; }
.search-box {
  width: 90%;
  max-width: 640px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6);
}
.search-input-wrap {
  display: flex;
  align-items: center;
  padding: 0.85rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
}
.search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: #fff;
  font-size: 1.1rem;
}
.search-results {
  max-height: 420px;
  overflow-y: auto;
  padding: 0.75rem;
}
.search-result-item {
  display: block;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  color: var(--text-primary);
  text-decoration: none;
  transition: background 0.15s;
}
.search-result-item:hover, .search-result-item.selected {
  background: var(--bg-card);
  text-decoration: none;
}
.search-res-title { font-weight: 600; font-size: 0.95rem; color: var(--accent); }
.search-res-section { font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.2rem; }
.search-res-excerpt { font-size: 0.825rem; color: var(--text-secondary); line-height: 1.4; }

/* Landing Page Hero & Modern Layout */
.hero-container {
  padding: 2.5rem 0 3.5rem 0;
  text-align: center;
  max-width: 860px;
  margin: 0 auto;
}
.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.9rem;
  border-radius: 9999px;
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.3);
  color: var(--accent);
  font-size: 0.825rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
}
.hero-title {
  font-size: 2.85rem;
  font-weight: 800;
  line-height: 1.15;
  margin-bottom: 1.25rem;
  background: linear-gradient(135deg, #ffffff 40%, #38bdf8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  border: none;
  padding: 0;
}
.hero-desc {
  font-size: 1.15rem;
  color: #9ca3af;
  line-height: 1.6;
  margin-bottom: 2rem;
  max-width: 720px;
  margin-left: auto;
  margin-right: auto;
}
.hero-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 3rem;
}
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: #38bdf8;
  color: #030712 !important;
  font-weight: 600;
  padding: 0.75rem 1.4rem;
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.2s ease;
  box-shadow: 0 4px 14px rgba(56, 189, 248, 0.3);
}
.btn-primary:hover {
  background: #0ea5e9;
  text-decoration: none;
  transform: translateY(-1px);
}
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-card);
  color: var(--text-primary);
  font-weight: 600;
  padding: 0.75rem 1.4rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  text-decoration: none;
  transition: all 0.2s ease;
}
.btn-secondary:hover {
  background: var(--bg-secondary);
  border-color: var(--accent);
  color: var(--accent);
  text-decoration: none;
}

/* Terminal Quickstart Card */
.terminal-card {
  background: var(--bg-code);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  text-align: left;
  margin: 2.5rem 0;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
}
.terminal-header {
  background: #111827;
  padding: 0.65rem 1rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.terminal-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
.dot-red { background: #ef4444; }
.dot-yellow { background: #f59e0b; }
.dot-green { background: #10b981; }
.terminal-title {
  margin-left: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}
.terminal-body {
  padding: 1.25rem;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  line-height: 1.7;
}

/* Pillars Grid */
.section-heading-center {
  text-align: center;
  margin: 3.5rem 0 0.75rem 0;
  border: none;
  font-size: 1.85rem;
}
.section-subheading-center {
  text-align: center;
  color: var(--text-muted);
  font-size: 1.05rem;
  margin-bottom: 2rem;
}
.pillar-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}
.pillar-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 1.5rem;
  transition: transform 0.2s ease, border-color 0.2s ease;
}
.pillar-card:hover {
  transform: translateY(-3px);
  border-color: var(--accent);
}
.pillar-icon {
  font-size: 1.75rem;
  margin-bottom: 0.75rem;
}
.pillar-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}
.pillar-desc {
  font-size: 0.9rem;
  color: #9ca3af;
  line-height: 1.55;
  margin-bottom: 0;
}

/* Mental Model Scalpel Section */
.scalpel-steps {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.25rem;
  margin: 2rem 0;
}
.scalpel-step {
  background: #111827;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.25rem;
}
.scalpel-num {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--accent);
  text-transform: uppercase;
  margin-bottom: 0.5rem;
}
.scalpel-title {
  font-size: 1.05rem;
  font-weight: 600;
  margin-bottom: 0.4rem;
}
.scalpel-desc {
  font-size: 0.85rem;
  color: #9ca3af;
  margin-bottom: 0;
}

/* Guided Tour Styles */
.tour-step-badge {
  display: inline-block;
  padding: 0.25rem 0.65rem;
  border-radius: 9999px;
  background: rgba(56, 189, 248, 0.15);
  color: var(--accent);
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  margin-bottom: 0.75rem;
}
.tour-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 1.75rem;
  margin: 1.75rem 0;
}
.tour-nav-bar {
  display: flex;
  justify-content: space-between;
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

/* Modern Landing Page Styles */
.landing-body {
  background: #0b0f19;
}
.landing-body .site-header {
  position: sticky;
  top: 0;
  z-index: 50;
  background: rgba(11, 15, 25, 0.88);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-color);
}
.landing-main {
  max-width: 1140px;
  margin: 0 auto;
  padding: 1.5rem;
}
.landing-hero {
  text-align: center;
  padding: 4.5rem 1rem 3rem 1rem;
  background: radial-gradient(ellipse 80% 50% at 50% -20%, rgba(56, 189, 248, 0.12), transparent 70%);
}
.landing-title {
  font-size: 3.25rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  margin: 1.25rem 0 1rem 0;
  color: #f8fafc;
  line-height: 1.15;
}
.landing-subtitle {
  font-size: 1.25rem;
  color: #94a3b8;
  max-width: 780px;
  margin: 0 auto 2.25rem auto;
  line-height: 1.6;
}
.landing-section {
  margin: 5rem 0;
}
.section-header-center {
  text-align: center;
  max-width: 700px;
  margin: 0 auto 2.5rem auto;
}
.section-title {
  font-size: 2.25rem;
  font-weight: 700;
  color: #f8fafc;
  margin-bottom: 0.5rem;
}
.section-desc {
  font-size: 1.05rem;
  color: #94a3b8;
}
.scenarios-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}
.scenario-card {
  background: #111827;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.75rem;
  transition: transform 0.2s, border-color 0.2s;
  display: flex;
  flex-direction: column;
}
.scenario-card:hover {
  border-color: #38bdf8;
  transform: translateY(-2px);
}
.scenario-badge {
  align-self: flex-start;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.1);
  padding: 0.25rem 0.6rem;
  border-radius: 9999px;
  margin-bottom: 0.75rem;
}
.scenario-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: #f8fafc;
}
.scenario-text {
  font-size: 0.95rem;
  color: #94a3b8;
  line-height: 1.55;
  flex-grow: 1;
  margin-bottom: 1.25rem;
}
.scenario-command {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  background: #0b0f19;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  color: #38bdf8;
}
.pillars-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}
.pillar-card {
  background: #111827;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.75rem;
  transition: border-color 0.2s;
}
.pillar-card:hover {
  border-color: #38bdf8;
}
.pillar-icon {
  font-size: 1.75rem;
  margin-bottom: 0.75rem;
}
.pillar-title {
  font-size: 1.15rem;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 0.5rem;
}
.pillar-text {
  font-size: 0.9rem;
  color: #94a3b8;
  line-height: 1.55;
}
.landing-cta {
  margin: 5rem 0 3rem 0;
  padding: 3.5rem 2rem;
  background: linear-gradient(180deg, #111827 0%, #0b0f19 100%);
  border-radius: 16px;
  border: 1px solid var(--border-color);
  text-align: center;
}
.site-footer {
  border-top: 1px solid var(--border-color);
  padding: 3rem 1.5rem;
  text-align: center;
  color: var(--text-muted);
}

/* Responsive */
@media (max-width: 960px) {
  .site-sidebar { transform: translateX(-100%); transition: transform 0.25s ease; }
  .site-sidebar.open { transform: translateX(0); }
  .main-wrapper { margin-left: 0; padding: 1.5rem 1rem; }
  .header-center { display: none; }
  .hero-title { font-size: 2.25rem; }
  .pillar-grid { grid-template-columns: 1fr; }
}
`

// JSBundle contains interactive behavior: search modal, TOC scrollspy, code copying,
// Mermaid initialization, zoom controls, and Fullscreen Lightbox.
const JSBundle = `
(function() {
  // 1. Mermaid Initialization with Calibrated Theme
  if (typeof mermaid !== 'undefined') {
    mermaid.initialize({
      startOnLoad: true,
      theme: 'dark',
      themeVariables: {
        darkMode: true,
        fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        background: '#0f172a',
        primaryColor: '#1e293b',
        primaryTextColor: '#f8fafc',
        primaryBorderColor: '#38bdf8',
        lineColor: '#38bdf8',
        secondaryColor: '#1e293b',
        tertiaryColor: '#111827',
        mainBkg: '#1e293b',
        nodeBorder: '#38bdf8',
        clusterBkg: '#162032',
        clusterBorder: '#334155',
        titleColor: '#38bdf8',
        edgeLabelBackground: '#0b0f19',
        actorBkg: '#1e293b',
        actorBorder: '#38bdf8',
        actorTextColor: '#f8fafc',
        actorLineColor: '#38bdf8',
        signalColor: '#38bdf8',
        signalTextColor: '#f8fafc',
        labelBoxBkgColor: '#1e293b',
        labelBoxBorderColor: '#38bdf8',
        labelTextColor: '#f8fafc',
        loopTextColor: '#f8fafc',
        noteBkgColor: '#1e293b',
        noteBorderColor: '#fbbf24',
        noteTextColor: '#f8fafc'
      }
    });
  }

  // 2. Lightbox Elements
  var lightbox = document.createElement('div');
  lightbox.className = 'mermaid-lightbox';
  lightbox.id = 'mermaid-lightbox';
  lightbox.innerHTML = 
    '<div class="lightbox-header">' +
      '<div class="lightbox-title"><span>⛶</span> Diagram Lightbox</div>' +
      '<div class="lightbox-controls">' +
        '<button class="mermaid-btn" id="lightbox-zoom-in">🔍 +</button>' +
        '<button class="mermaid-btn" id="lightbox-zoom-out">🔍 -</button>' +
        '<button class="mermaid-btn" id="lightbox-reset">↺ Reset</button>' +
        '<button class="lightbox-close" id="lightbox-close">&times;</button>' +
      '</div>' +
    '</div>' +
    '<div class="lightbox-body" id="lightbox-body">' +
      '<div class="lightbox-content" id="lightbox-content"></div>' +
    '</div>';
  document.body.appendChild(lightbox);

  var lightboxScale = 1.0;
  var lightboxContent = document.getElementById('lightbox-content');
  var lightboxCloseBtn = document.getElementById('lightbox-close');

  function openLightbox(svgHTML) {
    if (!lightboxContent) return;
    lightboxContent.innerHTML = svgHTML;
    lightboxScale = 1.0;
    lightboxContent.style.transform = 'scale(1)';
    lightbox.classList.add('open');
    document.body.style.overflow = 'hidden';
  }

  function closeLightbox() {
    lightbox.classList.remove('open');
    if (lightboxContent) lightboxContent.innerHTML = '';
    document.body.style.overflow = '';
  }

  if (lightboxCloseBtn) {
    lightboxCloseBtn.addEventListener('click', closeLightbox);
  }
  document.getElementById('lightbox-zoom-in')?.addEventListener('click', function() {
    lightboxScale *= 1.25;
    if (lightboxContent) lightboxContent.style.transform = 'scale(' + lightboxScale + ')';
  });
  document.getElementById('lightbox-zoom-out')?.addEventListener('click', function() {
    lightboxScale = Math.max(0.2, lightboxScale / 1.25);
    if (lightboxContent) lightboxContent.style.transform = 'scale(' + lightboxScale + ')';
  });
  document.getElementById('lightbox-reset')?.addEventListener('click', function() {
    lightboxScale = 1.0;
    if (lightboxContent) lightboxContent.style.transform = 'scale(1)';
  });
  lightbox.addEventListener('click', function(e) {
    if (e.target === lightbox || e.target === document.getElementById('lightbox-body')) {
      closeLightbox();
    }
  });

  // 3. Setup Mermaid Toolbars (Zoom, Fullscreen, Copy)
  document.querySelectorAll('.mermaid-wrapper').forEach(function(wrap) {
    var inner = wrap.querySelector('.mermaid-inner');
    var zoomLevel = 1.0;

    var zoomInBtn = wrap.querySelector('.mermaid-zoom-in');
    var zoomOutBtn = wrap.querySelector('.mermaid-zoom-out');
    var resetBtn = wrap.querySelector('.mermaid-zoom-reset');
    var fsBtn = wrap.querySelector('.mermaid-fullscreen-btn');
    var copyBtn = wrap.querySelector('.mermaid-copy-btn');

    if (zoomInBtn && inner) {
      zoomInBtn.addEventListener('click', function() {
        zoomLevel *= 1.25;
        inner.style.transform = 'scale(' + zoomLevel + ')';
      });
    }
    if (zoomOutBtn && inner) {
      zoomOutBtn.addEventListener('click', function() {
        zoomLevel = Math.max(0.4, zoomLevel / 1.25);
        inner.style.transform = 'scale(' + zoomLevel + ')';
      });
    }
    if (resetBtn && inner) {
      resetBtn.addEventListener('click', function() {
        zoomLevel = 1.0;
        inner.style.transform = 'scale(1)';
      });
    }
    if (fsBtn && inner) {
      fsBtn.addEventListener('click', function() {
        var svg = inner.querySelector('svg');
        if (svg) {
          openLightbox(svg.outerHTML);
        }
      });
    }
    if (copyBtn) {
      copyBtn.addEventListener('click', function() {
        var rawCode = wrap.getAttribute('data-mermaid-src') || '';
        if (!rawCode) {
          var pre = wrap.querySelector('pre.mermaid');
          if (pre) rawCode = pre.innerText;
        }
        if (rawCode) {
          navigator.clipboard.writeText(rawCode).then(function() {
            var orig = copyBtn.innerText;
            copyBtn.innerText = '✓ Copied!';
            setTimeout(function() { copyBtn.innerText = orig; }, 2000);
          });
        }
      });
    }
  });

  // 4. Code Block Copy Buttons
  document.querySelectorAll('pre').forEach(function(pre) {
    if (pre.classList.contains('mermaid') || pre.closest('.mermaid-wrapper')) return;
    var btn = document.createElement('button');
    btn.className = 'copy-btn';
    btn.innerText = 'Copy';
    btn.addEventListener('click', function() {
      var code = pre.querySelector('code') ? pre.querySelector('code').innerText : pre.innerText;
      navigator.clipboard.writeText(code).then(function() {
        btn.innerText = 'Copied!';
        setTimeout(function() { btn.innerText = 'Copy'; }, 2000);
      });
    });
    pre.appendChild(btn);
  });

  // 5. Search Modal & Path Resolution
  var searchModal = document.getElementById('search-modal');
  var searchInput = document.getElementById('search-input');
  var searchResults = document.getElementById('search-results');
  var searchIndex = null;

  function getBasePrefix() {
    var link = document.querySelector('link[rel="stylesheet"][href*="assets/style.css"]');
    if (link) {
      return link.getAttribute('href').replace('assets/style.css', '');
    }
    return '';
  }

  function loadSearchIndex() {
    if (!searchIndex) {
      var prefix = getBasePrefix();
      fetch(prefix + 'search-index.json')
        .then(function(res) { return res.json(); })
        .then(function(data) { searchIndex = data; })
        .catch(function(err) { console.error('Failed to load search index', err); });
    }
  }

  function openSearch() {
    loadSearchIndex();
    if (searchModal) {
      searchModal.classList.add('open');
      if (searchInput) searchInput.focus();
    }
  }

  function closeSearch() {
    if (searchModal) {
      searchModal.classList.remove('open');
      if (searchInput) searchInput.value = '';
      if (searchResults) searchResults.innerHTML = '';
    }
  }

  var searchBtn = document.getElementById('search-btn');
  if (searchBtn) searchBtn.addEventListener('click', openSearch);

  if (searchModal) {
    searchModal.addEventListener('click', function(e) {
      if (e.target === searchModal) closeSearch();
    });
  }

  document.addEventListener('keydown', function(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      if (searchModal && searchModal.classList.contains('open')) closeSearch();
      else openSearch();
    }
    if (e.key === 'Escape') {
      if (lightbox && lightbox.classList.contains('open')) closeLightbox();
      else if (searchModal && searchModal.classList.contains('open')) closeSearch();
    }
  });

  if (searchInput) {
    searchInput.addEventListener('input', function() {
      var query = searchInput.value.trim().toLowerCase();
      if (!query || !searchIndex) {
        searchResults.innerHTML = '';
        return;
      }
      var prefix = getBasePrefix();
      var matches = searchIndex.filter(function(item) {
        return item.title.toLowerCase().indexOf(query) !== -1 ||
               item.section.toLowerCase().indexOf(query) !== -1 ||
               (item.excerpt && item.excerpt.toLowerCase().indexOf(query) !== -1) ||
               (item.keywords && item.keywords.toLowerCase().indexOf(query) !== -1);
      }).slice(0, 10);

      if (matches.length === 0) {
        searchResults.innerHTML = '<div style="padding: 1.5rem; text-align: center; color: var(--text-muted);">No documents found for "' + query + '"</div>';
        return;
      }

      searchResults.innerHTML = matches.map(function(item) {
        var targetURL = prefix + item.url;
        return '<a href="' + targetURL + '" class="search-result-item">' +
               '<div class="search-res-section">' + item.section + '</div>' +
               '<div class="search-res-title">' + item.title + '</div>' +
               '<div class="search-res-excerpt">' + (item.excerpt || '') + '</div>' +
               '</a>';
      }).join('');
    });
  }

  // 6. TOC ScrollSpy
  var tocLinks = document.querySelectorAll('.toc-link');
  var headings = [];
  tocLinks.forEach(function(link) {
    var id = link.getAttribute('href');
    if (id && id.startsWith('#')) {
      var el = document.getElementById(id.substring(1));
      if (el) headings.push({ el: el, link: link });
    }
  });

  if (headings.length > 0) {
    window.addEventListener('scroll', function() {
      var scrollPos = window.scrollY + 100;
      var current = null;
      for (var i = 0; i < headings.length; i++) {
        if (headings[i].el.offsetTop <= scrollPos) {
          current = headings[i].link;
        } else {
          break;
        }
      }
      tocLinks.forEach(function(l) { l.classList.remove('active'); });
      if (current) current.classList.add('active');
    }, { passive: true });
  }
})();
`
