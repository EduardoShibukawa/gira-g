package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/docsite"
)

func runDocs(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("docs", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outputDir string
	var indexPath string
	var repoRoot string
	var title string
	var servePort string

	fs.StringVar(&outputDir, "output", "build", "Target directory to export static documentation site")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json for graph generation")
	fs.StringVar(&repoRoot, "root", "", "Repository root directory (defaults to git top-level)")
	fs.StringVar(&title, "title", "Gira-G Architecture Harness", "Documentation portal title")
	fs.StringVar(&servePort, "serve", "", "Optional local HTTP server address/port to preview (e.g. 8080 or :8080)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	root := repoRoot
	if root == "" {
		root = gitToplevel(".")
		if root == "" {
			root = "."
		}
	}

	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		absOutput = outputDir
	}

	cfg := &docsite.SiteConfig{
		Title:     title,
		Tagline:   "Deterministic Code Intelligence & Living Architecture",
		Version:   "v0.3.0",
		GitHubURL: "https://github.com/EduardoShibukawa/gira-g",
		RepoRoot:  root,
		OutputDir: absOutput,
		IndexPath: indexPath,
	}

	builder := docsite.NewBuilder(cfg)
	if err := builder.Build(); err != nil {
		return fmt.Errorf("failed to build documentation site: %w", err)
	}

	fmt.Fprintf(stdout, "Documentation site successfully generated at: %s\n", absOutput)
	fmt.Fprintf(stdout, " - Index: file://%s/index.html\n", filepath.ToSlash(absOutput))
	fmt.Fprintf(stdout, " - Architecture: file://%s/architecture/index.html\n", filepath.ToSlash(absOutput))
	fmt.Fprintf(stdout, " - Visual Explorer: file://%s/architecture/graph.html\n", filepath.ToSlash(absOutput))
	fmt.Fprintf(stdout, " - Total pages: %d\n", len(builder.Pages)+1)

	if servePort != "" {
		if !strings.Contains(servePort, ":") {
			servePort = ":" + servePort
		}
		fmt.Fprintf(stdout, "Serving documentation on http://localhost%s (press Ctrl+C to stop)...\n", servePort)
		fsHandler := http.FileServer(http.Dir(absOutput))
		return http.ListenAndServe(servePort, fsHandler)
	}

	return nil
}
