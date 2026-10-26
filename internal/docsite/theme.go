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
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
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
  background: rgba(11, 15, 25, 0.85);
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

/* Tables */
table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
  font-size: 0.9rem;
}
th, td {
  padding: 0.65rem 0.9rem;
  border: 1px solid var(--border-color);
  text-align: left;
}
th {
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-weight: 600;
}
tr:nth-child(even) {
  background: rgba(31, 41, 55, 0.4);
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

/* Mermaid */
.mermaid-diagram {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1.5rem 0;
  display: flex;
  justify-content: center;
  overflow-x: auto;
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
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
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
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
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

/* Responsive */
@media (max-width: 960px) {
  .site-sidebar { transform: translateX(-100%); transition: transform 0.25s ease; }
  .site-sidebar.open { transform: translateX(0); }
  .main-wrapper { margin-left: 0; padding: 1.5rem 1rem; }
  .header-center { display: none; }
}
`

// JSBundle contains interactive behavior: search modal, TOC scrollspy, code copying, and Mermaid init.
const JSBundle = `
(function() {
  // 1. Mermaid Initialization
  if (typeof mermaid !== 'undefined') {
    mermaid.initialize({
      startOnLoad: true,
      theme: 'dark',
      themeVariables: {
        darkMode: true,
        background: '#111827',
        primaryColor: '#1f2937',
        primaryTextColor: '#f9fafb',
        primaryBorderColor: '#38bdf8',
        lineColor: '#38bdf8',
        secondaryColor: '#1e293b',
        tertiaryColor: '#0f172a'
      }
    });
  }

  // 2. Code Block Copy Buttons
  document.querySelectorAll('pre').forEach(function(pre) {
    if (pre.classList.contains('mermaid')) return;
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

  // 3. Search Modal & Fuzzy Search
  var searchModal = document.getElementById('search-modal');
  var searchInput = document.getElementById('search-input');
  var searchResults = document.getElementById('search-results');
  var searchIndex = null;

  function loadSearchIndex() {
    if (!searchIndex) {
      fetch('search-index.json')
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
    if (e.key === 'Escape' && searchModal && searchModal.classList.contains('open')) {
      closeSearch();
    }
  });

  if (searchInput) {
    searchInput.addEventListener('input', function() {
      var query = searchInput.value.trim().toLowerCase();
      if (!query || !searchIndex) {
        searchResults.innerHTML = '';
        return;
      }
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
        return '<a href="' + item.url + '" class="search-result-item">' +
               '<div class="search-res-section">' + item.section + '</div>' +
               '<div class="search-res-title">' + item.title + '</div>' +
               '<div class="search-res-excerpt">' + (item.excerpt || '') + '</div>' +
               '</a>';
      }).join('');
    });
  }

  // 4. TOC ScrollSpy
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
