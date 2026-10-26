package docsite

import (
	"bytes"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	rendererHTML "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

var (
	frontmatterRe = regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---\r?\n`)
	headingH1Re   = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	headingTagRe  = regexp.MustCompile(`(?i)<h([23])\s+id="([^"]+)">([^<]+)</h[23]>`)
	codeFenceRe   = regexp.MustCompile("(?s)```.*?```")
	htmlTagRe     = regexp.MustCompile(`<[^>]+>`)
	calloutRe     = regexp.MustCompile(`(?m)^>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*(.*)$`)
)

type frontmatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Slug        string `yaml:"slug"`
	Position    int    `yaml:"sidebar_position"`
}

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		rendererHTML.WithHardWraps(),
		rendererHTML.WithXHTML(),
		rendererHTML.WithUnsafe(),
	),
)

// ParseMarkdown processes raw markdown bytes, extracting frontmatter, TOC, title,
// and compiling to HTML.
func ParseMarkdown(content []byte, fallbackTitle string) (*Page, error) {
	str := string(content)
	var fm frontmatter

	// 1. Extract and strip frontmatter
	if loc := frontmatterRe.FindStringSubmatchIndex(str); loc != nil {
		fmData := str[loc[2]:loc[3]]
		_ = yaml.Unmarshal([]byte(fmData), &fm)
		str = str[loc[1]:]
	}

	// 2. Extract title
	title := fm.Title
	if title == "" {
		if match := headingH1Re.FindStringSubmatch(str); len(match) > 1 {
			title = strings.TrimSpace(match[1])
			// Strip backticks or formatting from title
			title = strings.Trim(title, "`*")
		} else {
			title = fallbackTitle
		}
	}

	// 3. Transform GitHub-style alerts into styled callouts
	str = transformCallouts(str)

	// 4. Render to HTML via goldmark
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(str), &buf); err != nil {
		return nil, err
	}

	renderedHTML := buf.String()

	// 5. Post-process Mermaid diagrams if present:
	// goldmark renders ```mermaid as <pre><code class="language-mermaid">...</code></pre>
	// We convert it to <div class="mermaid">...</div> for crisp client-side rendering.
	renderedHTML = transformMermaidBlocks(renderedHTML)

	// 6. Extract Table of Contents from h2 and h3
	toc := extractTOC(renderedHTML)

	// 7. Plain text excerpt for search
	plainText := extractPlainText(str)
	excerpt := plainText
	if len(excerpt) > 280 {
		excerpt = excerpt[:280] + "..."
	}

	return &Page{
		Title:       title,
		Description: fm.Description,
		Order:       fm.Position,
		HTMLContent: template.HTML(renderedHTML),
		RawContent:  plainText,
		TOC:         toc,
	}, nil
}

func transformCallouts(input string) string {
	lines := strings.Split(input, "\n")
	var result []string
	inCallout := false
	var calloutType string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if match := calloutRe.FindStringSubmatch(line); len(match) > 1 {
			if inCallout {
				result = append(result, "</div></div>")
			}
			calloutType = strings.ToLower(match[1])
			title := strings.ToUpper(calloutType[:1]) + calloutType[1:]
			inCallout = true
			result = append(result, `<div class="callout callout-`+calloutType+`">`)
			result = append(result, `<div class="callout-header"><span class="callout-icon"></span><span class="callout-title">`+title+`</span></div>`)
			result = append(result, `<div class="callout-body">`)
			if len(match) > 2 && strings.TrimSpace(match[2]) != "" {
				result = append(result, match[2])
			}
			continue
		}

		if inCallout {
			if strings.HasPrefix(line, "> ") {
				result = append(result, strings.TrimPrefix(line, "> "))
				continue
			} else if line == ">" {
				result = append(result, "")
				continue
			} else {
				// End of callout
				result = append(result, "</div></div>")
				inCallout = false
			}
		}

		result = append(result, line)
	}

	if inCallout {
		result = append(result, "</div></div>")
	}

	return strings.Join(result, "\n")
}

var mermaidCodeRe = regexp.MustCompile(`(?s)<pre><code class="language-mermaid">(.*?)</code></pre>`)

func transformMermaidBlocks(htmlStr string) string {
	return mermaidCodeRe.ReplaceAllStringFunc(htmlStr, func(m string) string {
		sub := mermaidCodeRe.FindStringSubmatch(m)
		if len(sub) > 1 {
			// Unescape standard HTML entities because Mermaid parser needs raw text
			raw := html.UnescapeString(sub[1])
			return `<div class="mermaid-diagram"><pre class="mermaid">` + raw + `</pre></div>`
		}
		return m
	})
}

func extractTOC(htmlStr string) []TOCItem {
	var items []TOCItem
	matches := headingTagRe.FindAllStringSubmatch(htmlStr, -1)
	for _, m := range matches {
		if len(m) > 3 {
			level := 2
			if m[1] == "3" {
				level = 3
			}
			text := strings.TrimSpace(htmlTagRe.ReplaceAllString(m[3], ""))
			items = append(items, TOCItem{
				Level: level,
				ID:    m[2],
				Text:  text,
			})
		}
	}
	return items
}

func extractPlainText(markdown string) string {
	// Remove code fences
	cleaned := codeFenceRe.ReplaceAllString(markdown, "")
	// Remove HTML tags
	cleaned = htmlTagRe.ReplaceAllString(cleaned, "")
	// Remove markdown links [text](url) -> text
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	cleaned = linkRe.ReplaceAllString(cleaned, "$1")
	// Remove headers '#'
	headRe := regexp.MustCompile(`(?m)^#+\s+`)
	cleaned = headRe.ReplaceAllString(cleaned, "")
	// Normalize whitespace
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	return cleaned
}
