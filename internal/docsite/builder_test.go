package docsite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	md := `---
title: Test Page
sidebar_position: 2
---

# Test Heading

> [!NOTE]
> This is a crucial note for testing.

Some text with ` + "`code`" + ` and a table:

| Invariant | Value |
| :--- | :--- |
| Fast Pass | <5ms |

And a diagram:

` + "```mermaid\ngraph TD\n  A --> B\n```" + `

Link test: [Architecture](docs/intro.md)

## Section Two

More content.

### Subsection Three

Deep content.
`
	page, err := ParseMarkdown([]byte(md), "Fallback")
	if err != nil {
		t.Fatalf("ParseMarkdown error: %v", err)
	}

	if page.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", page.Title)
	}
	if page.Order != 2 {
		t.Errorf("expected order 2, got %d", page.Order)
	}
	contentStr := string(page.HTMLContent)
	if !strings.Contains(contentStr, "callout-note") {
		t.Errorf("expected rendered callout-note in HTML, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "table-wrapper") {
		t.Errorf("expected table-wrapper in HTML, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "mermaid-wrapper") {
		t.Errorf("expected mermaid-wrapper in HTML, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "mermaid-toolbar") {
		t.Errorf("expected mermaid-toolbar in HTML, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, `href="docs/intro.html"`) {
		t.Errorf("expected rewritten markdown link href=\"docs/intro.html\", got:\n%s", contentStr)
	}
	if len(page.TOC) < 2 {
		t.Errorf("expected at least 2 TOC items, got %d", len(page.TOC))
	}
}

func TestDocsiteBuilder_Mock(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "out")

	// Create mock repo structure
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Mock Repo\nHello world"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "docs", "cli"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "docs", "intro.md"), []byte("# Intro\nWelcome"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "docs", "cli", "slice.md"), []byte("# gira slice\nExtracts context"), 0644)

	cfg := &SiteConfig{
		Title:     "Mock Gira",
		RepoRoot:  tmpDir,
		OutputDir: outDir,
	}

	b := NewBuilder(cfg)
	if err := b.Build(); err != nil {
		t.Fatalf("b.Build() failed: %v", err)
	}

	// Verify generated files
	expectedFiles := []string{
		"index.html",
		"tour.html",
		"intro.html",
		"cli/slice.html",
		"architecture/graph.html",
		"search-index.json",
		"assets/style.css",
		"assets/app.js",
	}

	for _, file := range expectedFiles {
		p := filepath.Join(outDir, file)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected generated file %s, but stat returned error: %v", file, err)
		}
	}

	// Verify Landing Page content
	indexContent, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}
	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "hero-title") {
		t.Errorf("index.html missing hero-title")
	}
	if !strings.Contains(indexStr, "terminal-card") {
		t.Errorf("index.html missing terminal-card")
	}
	if !strings.Contains(indexStr, "pillar-card") {
		t.Errorf("index.html missing pillar-card")
	}

	// Verify Tour Page content
	tourContent, err := os.ReadFile(filepath.Join(outDir, "tour.html"))
	if err != nil {
		t.Fatalf("failed to read tour.html: %v", err)
	}
	tourStr := string(tourContent)
	if !strings.Contains(tourStr, "Step 1 of 6") {
		t.Errorf("tour.html missing Step 1 of 6")
	}
	if !strings.Contains(tourStr, "Step 6 of 6") {
		t.Errorf("tour.html missing Step 6 of 6")
	}
}
