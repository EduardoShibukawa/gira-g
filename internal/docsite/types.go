package docsite

import (
	"html/template"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
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
	Title      string
	Tagline    string
	Version    string
	GitHubURL  string
	LogoLetter string
	TargetName string
	TargetDesc string
	BaseURL    string
	RepoRoot   string
	OutputDir  string
	IndexPath  string
	MacroMap   *domain.MacroMap
	// Docsite carries the target repo's `docsite:` section from gira.yaml.
	// When nil, tour and landing render the generic no-special-case path.
	Docsite *manifest.DocsiteConfig
}

// siteTitle resolves the human-facing site title: the explicit site Title
// first, then the target manifest's docsite title. With neither set it uses
// a name-agnostic phrase — per-project naming happens where the title is
// configured (the `docsite:` section or the `gira docs -title` flag), never
// by special-casing the target's identifier.
func siteTitle(config *SiteConfig) string {
	if config == nil {
		return "the project"
	}
	if config.Title != "" {
		return config.Title
	}
	if config.Docsite != nil && config.Docsite.Title != "" {
		return config.Docsite.Title
	}
	return "the project"
}

// headerTitle layers the explicit site Title over the manifest docsite title
// without any name formatting fallback, so empty stays empty for targets
// that carry no configuration at all.
func headerTitle(config *SiteConfig) string {
	if config == nil {
		return ""
	}
	if config.Title != "" {
		return config.Title
	}
	if config.Docsite != nil {
		return config.Docsite.Title
	}
	return ""
}
