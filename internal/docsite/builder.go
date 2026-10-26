package docsite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/architecture"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

type Builder struct {
	Config    *SiteConfig
	Pages     []*Page
	NavGroups []NavGroup
	MacroMap  *domain.MacroMap
	Arch      *architecture.ClusteredArchitecture
	Packages  []architecture.PackageInfo
}

func NewBuilder(cfg *SiteConfig) *Builder {
	if cfg.Title == "" {
		cfg.Title = "Gira-G"
	}
	if cfg.Tagline == "" {
		cfg.Tagline = "Deterministic Code Intelligence & Living Architecture"
	}
	if cfg.Version == "" {
		cfg.Version = "v0.3.0"
	}
	if cfg.GitHubURL == "" {
		cfg.GitHubURL = "https://github.com/EduardoShibukawa/gira-g"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "build"
	}
	return &Builder{Config: cfg}
}

// Build compiles all markdown sources, graphs, search index, and assets into OutputDir.
func (b *Builder) Build() error {
	repoRoot := b.Config.RepoRoot
	if repoRoot == "" {
		repoRoot = "."
	}

	// 1. Scan Go packages and load/generate ClusteredArchitecture
	pkgs, _ := architecture.ScanRepository(repoRoot)
	b.Packages = pkgs

	if b.Config.IndexPath != "" {
		st := store.NewStore(filepath.Dir(b.Config.IndexPath))
		if mm, loadErr := st.LoadFile(b.Config.IndexPath); loadErr == nil && mm != nil && len(mm.Nodes) > 0 {
			b.MacroMap = mm
			b.Arch = architecture.ClusterMacroMap(mm)
		}
	}
	if b.Arch == nil {
		b.Arch = architecture.ClusterLocalRepository(pkgs)
	}

	// 2. Scan & parse markdown files
	if err := b.scanPages(repoRoot); err != nil {
		return fmt.Errorf("failed to scan pages: %w", err)
	}

	// 3. Build navigation groups
	b.buildNavGroups()

	// 4. Connect prev/next navigation
	b.connectPagination()

	// 5. Ensure output directories exist
	if err := os.MkdirAll(b.Config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "assets"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "architecture"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "cli"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "concepts"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "guides"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "adr"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(b.Config.OutputDir, "prior-art"), 0755); err != nil {
		return err
	}

	// 6. Write static assets (CSS, JS)
	cssPath := filepath.Join(b.Config.OutputDir, "assets", "style.css")
	if err := os.WriteFile(cssPath, []byte(CSSBundle), 0644); err != nil {
		return err
	}
	jsPath := filepath.Join(b.Config.OutputDir, "assets", "app.js")
	if err := os.WriteFile(jsPath, []byte(JSBundle), 0644); err != nil {
		return err
	}

	// 7. Render and write each page
	for _, page := range b.Pages {
		htmlStr, err := b.renderPageHTML(page)
		if err != nil {
			return fmt.Errorf("failed to render page %s: %w", page.URL, err)
		}
		targetFile := filepath.Join(b.Config.OutputDir, filepath.FromSlash(page.URL))
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(targetFile, []byte(htmlStr), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", targetFile, err)
		}
	}

	// 8. Generate and write Modern Landing Page (index.html)
	landingHTML := GenerateLandingPageHTML(b.Config, b.NavGroups)
	indexPath := filepath.Join(b.Config.OutputDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(landingHTML), 0644); err != nil {
		return fmt.Errorf("failed to write landing index.html: %w", err)
	}

	// 9. Generate and write Understand-Anything style Guided Tour (tour.html)
	tourHTML := GenerateTourPageHTML(b.Config, b.NavGroups)
	tourPath := filepath.Join(b.Config.OutputDir, "tour.html")
	if err := os.WriteFile(tourPath, []byte(tourHTML), 0644); err != nil {
		return fmt.Errorf("failed to write tour.html: %w", err)
	}

	// 10. Generate and write Interactive Cytoscape Visual Explorer
	graphHTML, err := GenerateInteractiveGraphHTML(b.Config, b.MacroMap, b.Arch, b.Packages, b.NavGroups)
	if err != nil {
		return fmt.Errorf("failed to generate graph HTML: %w", err)
	}
	graphFile := filepath.Join(b.Config.OutputDir, "architecture", "graph.html")
	if err := os.WriteFile(graphFile, []byte(graphHTML), 0644); err != nil {
		return fmt.Errorf("failed to write graph HTML: %w", err)
	}

	// 11. Generate and write search index
	if err := b.writeSearchIndex(); err != nil {
		return fmt.Errorf("failed to write search index: %w", err)
	}

	return nil
}

func (b *Builder) scanPages(repoRoot string) error {
	var pages []*Page

	// 1. Overview Page (from README.md)
	readmePath := filepath.Join(repoRoot, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil {
		page, err := ParseMarkdown(data, "Gira-G Architecture Harness")
		if err == nil {
			page.ID = "overview"
			page.URL = "overview.html"
			page.Section = "Overview"
			page.SourcePath = readmePath
			pages = append(pages, page)
		}
	}

	// 2. Intro Page (docs/intro.md)
	introPath := filepath.Join(repoRoot, "docs", "intro.md")
	if data, err := os.ReadFile(introPath); err == nil {
		page, err := ParseMarkdown(data, "Introduction")
		if err == nil {
			page.ID = "intro"
			page.URL = "intro.html"
			page.Section = "Getting Started"
			page.SourcePath = introPath
			pages = append(pages, page)
		}
	}

	// 3. Living Architecture Reference (ARCHITECTURE.md)
	archPath := filepath.Join(repoRoot, "ARCHITECTURE.md")
	if data, err := os.ReadFile(archPath); err == nil {
		page, err := ParseMarkdown(data, "Living Architecture Reference")
		if err == nil {
			page.ID = "architecture/index"
			page.URL = "architecture/index.html"
			page.Section = "Architecture"
			page.SourcePath = archPath
			pages = append(pages, page)
		}
	}

	// 4. Scan subdirectories in docs/
	dirs := []struct {
		SubDir  string
		Section string
	}{
		{"guides", "Guides & Workflows"},
		{"cli", "CLI Reference"},
		{"concepts", "Core Concepts"},
		{"adr", "Architecture Decisions (ADR)"},
		{"prior-art", "Prior Art & Research"},
	}

	for _, d := range dirs {
		dirPath := filepath.Join(repoRoot, "docs", d.SubDir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			fpath := filepath.Join(dirPath, e.Name())
			data, err := os.ReadFile(fpath)
			if err != nil {
				continue
			}
			baseName := strings.TrimSuffix(e.Name(), ".md")
			fallbackTitle := strings.ReplaceAll(baseName, "-", " ")
			fallbackTitle = strings.Title(fallbackTitle)

			page, err := ParseMarkdown(data, fallbackTitle)
			if err != nil {
				continue
			}
			page.ID = d.SubDir + "/" + baseName
			page.URL = d.SubDir + "/" + baseName + ".html"
			page.Section = d.Section
			page.SourcePath = fpath
			pages = append(pages, page)
		}
	}

	b.Pages = pages
	return nil
}

func (b *Builder) buildNavGroups() {
	groupMap := make(map[string][]NavItem)
	sectionOrders := map[string]int{
		"Overview":                     1,
		"Getting Started":              2,
		"Guides & Workflows":           3,
		"CLI Reference":                4,
		"Core Concepts":                5,
		"Architecture":                 6,
		"Architecture Decisions (ADR)": 7,
		"Prior Art & Research":         8,
	}

	// Always add Home to Overview
	groupMap["Overview"] = append(groupMap["Overview"], NavItem{
		Title: "Home",
		URL:   "index.html",
		Order: 0,
	})

	// Always add Guided Tour to Getting Started
	groupMap["Getting Started"] = append(groupMap["Getting Started"], NavItem{
		Title: "Guided Tour (Topological)",
		URL:   "tour.html",
		Order: 1,
	})

	// Always add the Interactive Visual Explorer to Architecture
	groupMap["Architecture"] = append(groupMap["Architecture"], NavItem{
		Title: "Interactive Graph Explorer",
		URL:   "architecture/graph.html",
		Order: 99,
	})

	for _, p := range b.Pages {
		sec := p.Section
		if sec == "" {
			sec = "General"
		}
		groupMap[sec] = append(groupMap[sec], NavItem{
			Title: p.Title,
			URL:   p.URL,
			Order: p.Order,
		})
	}

	var groups []NavGroup
	for sec, items := range groupMap {
		// Sort items within each group
		sort.Slice(items, func(i, j int) bool {
			if items[i].Order != items[j].Order && items[i].Order > 0 && items[j].Order > 0 {
				return items[i].Order < items[j].Order
			}
			return items[i].Title < items[j].Title
		})
		groups = append(groups, NavGroup{
			Title: sec,
			Items: items,
			Order: sectionOrders[sec],
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Order < groups[j].Order
	})

	b.NavGroups = groups
}

func (b *Builder) connectPagination() {
	var linear []*Page
	// Linear order: Overview -> Getting Started -> Architecture -> CLI -> Concepts -> Guides -> ADRs
	for _, g := range b.NavGroups {
		for _, item := range g.Items {
			for _, p := range b.Pages {
				if p.URL == item.URL {
					linear = append(linear, p)
					break
				}
			}
		}
	}

	for i := 0; i < len(linear); i++ {
		if i > 0 {
			linear[i].Prev = &NavItem{Title: linear[i-1].Title, URL: linear[i-1].URL}
		}
		if i < len(linear)-1 {
			linear[i].Next = &NavItem{Title: linear[i+1].Title, URL: linear[i+1].URL}
		}
	}
}

func (b *Builder) renderPageHTML(p *Page) (string, error) {
	prefix := ""
	depth := strings.Count(p.URL, "/")
	for i := 0; i < depth; i++ {
		prefix += "../"
	}
	if prefix == "" {
		prefix = "./"
	}

	navHTML := renderSidebarNav(b.NavGroups, p.URL)
	tocHTML := renderTOCHTML(p.TOC)
	breadcrumbsHTML := fmt.Sprintf(`<a href="%sindex.html">Docs</a> &rsaquo; <span>%s</span> &rsaquo; <span>%s</span>`, prefix, p.Section, p.Title)

	paginationHTML := `<div class="pagination-nav">`
	if p.Prev != nil {
		paginationHTML += fmt.Sprintf(`<a href="%s%s" class="pagination-link"><span class="pagination-sub">&larr; Previous</span><span class="pagination-title">%s</span></a>`, prefix, p.Prev.URL, p.Prev.Title)
	} else {
		paginationHTML += `<div></div>`
	}
	if p.Next != nil {
		paginationHTML += fmt.Sprintf(`<a href="%s%s" class="pagination-link pagination-next"><span class="pagination-sub">Next &rarr;</span><span class="pagination-title">%s</span></a>`, prefix, p.Next.URL, p.Next.Title)
	}
	paginationHTML += `</div>`

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s — %s</title>
  <meta name="description" content="%s">
  <link rel="stylesheet" href="%sassets/style.css">
  <script src="https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js"></script>
</head>
<body>
  <header class="site-header">
    <div style="display: flex; align-items: center; gap: 1rem;">
      <a href="%sindex.html" class="brand">
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
      <a href="%sintro.html" class="header-link">Docs</a>
      <a href="%stour.html" class="header-link">Guided Tour</a>
      <a href="%sarchitecture/index.html" class="header-link">Architecture</a>
      <a href="%sarchitecture/graph.html" class="header-link">Visual Explorer</a>
      <a href="%scli/index.html" class="header-link">CLI</a>
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
          %s
        </div>
        %s
        %s
      </article>

      <aside class="site-toc">
        %s
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

  <script src="%sassets/app.js"></script>
</body>
</html>`,
		p.Title,
		b.Config.Title,
		p.Description,
		prefix,
		prefix,
		b.Config.Title,
		b.Config.Version,
		prefix,
		prefix,
		prefix,
		prefix,
		prefix,
		b.Config.GitHubURL,
		navHTML,
		breadcrumbsHTML,
		p.HTMLContent,
		paginationHTML,
		tocHTML,
		prefix,
	)

	return html, nil
}

func (b *Builder) writeSearchIndex() error {
	var index []SearchDoc

	// Add Landing page to search index
	index = append(index, SearchDoc{
		ID:       "index",
		Title:    "Home & Overview",
		Section:  "Overview",
		URL:      "index.html",
		Excerpt:  "Deterministic Architecture Intelligence & Living Architecture. Zero hallucinations, sub-millisecond AST queries, and surgical editing interfaces.",
		Keywords: "overview home quickstart install scalpel fast-pass",
	})

	// Add Guided Tour to search index
	index = append(index, SearchDoc{
		ID:       "tour",
		Title:    "Guided Architectural Tour",
		Section:  "Getting Started",
		URL:      "tour.html",
		Excerpt:  "A curated topological walkthrough of Gira-G's architecture in 6 steps from entrypoints down to storage and governance.",
		Keywords: "tour guided walkthrough tutorial architecture",
	})

	for _, p := range b.Pages {
		index = append(index, SearchDoc{
			ID:       p.ID,
			Title:    p.Title,
			Section:  p.Section,
			URL:      p.URL,
			Excerpt:  extractExcerpt(p.RawContent),
			Keywords: p.ID + " " + p.Section + " " + p.Title,
		})
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(b.Config.OutputDir, "search-index.json")
	return os.WriteFile(indexPath, data, 0644)
}

func extractExcerpt(raw string) string {
	if len(raw) <= 200 {
		return raw
	}
	return raw[:200] + "..."
}

func renderSidebarNav(groups []NavGroup, currentURL string) string {
	var sb strings.Builder
	prefix := ""
	depth := strings.Count(currentURL, "/")
	for i := 0; i < depth; i++ {
		prefix += "../"
	}
	if prefix == "" {
		prefix = "./"
	}

	for _, g := range groups {
		isOpen := " open"
		if g.Title == "Architecture Decisions (ADR)" || g.Title == "Prior Art & Research" {
			isOpen = ""
			for _, item := range g.Items {
				if item.URL == currentURL {
					isOpen = " open"
					break
				}
			}
		}
		sb.WriteString(fmt.Sprintf(`<details class="nav-group-collapsible"%s>`, isOpen))
		sb.WriteString(fmt.Sprintf(`<summary class="nav-group-summary"><span>%s</span><span class="nav-group-chevron">▶</span></summary>`, g.Title))
		sb.WriteString(`<ul class="nav-list">`)
		for _, item := range g.Items {
			activeCls := ""
			if item.URL == currentURL {
				activeCls = " active"
			}
			targetURL := prefix + item.URL
			sb.WriteString(fmt.Sprintf(`<li class="nav-item%s"><a href="%s">%s</a></li>`, activeCls, targetURL, item.Title))
		}
		sb.WriteString(`</ul></details>`)
	}
	return sb.String()
}

func renderTOCHTML(toc []TOCItem) string {
	if len(toc) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<div class="toc-title">On this page</div>`)
	sb.WriteString(`<ul class="toc-list">`)
	for _, item := range toc {
		levelCls := "level-2"
		if item.Level == 3 {
			levelCls = "level-3"
		}
		sb.WriteString(fmt.Sprintf(`<li class="toc-item %s"><a href="#%s" class="toc-link">%s</a></li>`, levelCls, item.ID, item.Text))
	}
	sb.WriteString(`</ul>`)
	return sb.String()
}
