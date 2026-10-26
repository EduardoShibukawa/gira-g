package docsite

import (
	"html/template"
)

// Page represents a rendered documentation page.
type Page struct {
	ID          string // Relative file path without ext, e.g. "cli/slice"
	Title       string // Page title
	Description string // Optional page description
	Section     string // Section name (e.g. "CLI Reference", "Concepts", "Architecture")
	SectionIcon string // SVG or emoji icon
	URL         string // Relative URL from site root, e.g. "cli/slice.html"
	SourcePath  string // Path to source markdown file
	HTMLContent template.HTML
	RawContent  string
	TOC         []TOCItem
	Prev        *NavItem
	Next        *NavItem
	Order       int
}

// TOCItem represents a heading in the table of contents.
type TOCItem struct {
	ID    string
	Text  string
	Level int // 2 for h2, 3 for h3
}

// NavItem represents a link in the sidebar navigation.
type NavItem struct {
	Title string
	URL   string
	Icon  string
	Order int
}

// NavGroup represents a collapsible section in the sidebar.
type NavGroup struct {
	Title string
	Icon  string
	Items []NavItem
	Order int
}

// SearchDoc is indexed for client-side fuzzy search.
type SearchDoc struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Section  string `json:"section"`
	URL      string `json:"url"`
	Excerpt  string `json:"excerpt"`
	Keywords string `json:"keywords"`
}

// SiteConfig holds site-wide configuration.
type SiteConfig struct {
	Title       string
	Tagline     string
	Version     string
	GitHubURL   string
	BaseURL     string
	RepoRoot    string
	OutputDir   string
	IndexPath   string
}
