package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_DocsCommand(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "docs-site")
	var stdout, stderr bytes.Buffer

	err := Run([]string{"docs", "--output", outDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira docs command failed: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Documentation site successfully generated") {
		t.Errorf("expected success banner, got:\n%s", out)
	}

	// Verify key generated artifacts
	requiredFiles := []string{
		"index.html",
		"intro.html",
		"architecture/index.html",
		"architecture/graph.html",
		"search-index.json",
		"assets/style.css",
		"assets/app.js",
	}

	for _, f := range requiredFiles {
		p := filepath.Join(outDir, f)
		stat, err := os.Stat(p)
		if err != nil {
			t.Errorf("expected file %s to exist, err: %v", f, err)
			continue
		}
		if stat.Size() == 0 {
			t.Errorf("file %s is unexpectedly empty", f)
		}
	}

	// Verify index.html contains proper navbar and full-width landing content
	indexContent, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}
	if !strings.Contains(string(indexContent), "site-header") {
		t.Errorf("expected site-header in index.html")
	}
	if !strings.Contains(string(indexContent), "landing-hero") {
		t.Errorf("expected landing-hero in full-width index.html")
	}

	// Verify docs pages contain sidebar navigation
	introContent, err := os.ReadFile(filepath.Join(outDir, "intro.html"))
	if err != nil {
		t.Fatalf("failed to read intro.html: %v", err)
	}
	if !strings.Contains(string(introContent), "site-sidebar") {
		t.Errorf("expected site-sidebar in intro.html")
	}
}
