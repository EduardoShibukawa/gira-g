package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/architecture"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func runArchitecture(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("architecture", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var format string
	var outputPath string
	var indexPath string
	var checkout string

	fs.StringVar(&format, "format", "md", "Export format (md|html|json)")
	fs.StringVar(&outputPath, "output", "", "Output file path (default stdout)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	root := gitToplevel(".")
	if root == "" {
		root = "."
	}

	pkgs, err := architecture.ScanRepository(root)
	if err != nil && len(pkgs) == 0 {
		fmt.Fprintf(stderr, "warning: repository scan encountered errors: %v\n", err)
	}

	// Attempt to load MacroMap if index path is supplied or default exists
	var arch *architecture.ClusteredArchitecture
	mm, loadErr := loadIndex(indexPath, stderr)
	if loadErr == nil && mm != nil && len(mm.Nodes) > 0 {
		humanBanner(stderr, mm, checkout, slices.FormatMarkdown)
		arch = architecture.ClusterMacroMap(mm)
	} else {
		arch = architecture.ClusterLocalRepository(pkgs)
	}

	var content string
	switch strings.ToLower(format) {
	case "md", "markdown", "":
		content = architecture.GenerateArchitectureMarkdown(root, arch, pkgs)
	case "json":
		data, mErr := architecture.MarshalCompactJSON(arch)
		if mErr != nil {
			return fmt.Errorf("failed to encode architecture JSON: %w", mErr)
		}
		content = string(data)
	case "html":
		md := architecture.GenerateArchitectureMarkdown(root, arch, pkgs)
		content = wrapHTML(arch.TargetSystem, md)
	default:
		return fmt.Errorf("unsupported architecture format %q. Supported formats: md, html, json", format)
	}

	if outputPath != "" {
		if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write architecture file: %w", err)
		}
		fmt.Fprintf(stdout, "Exported architecture documentation to: %s\n", outputPath)
		return nil
	}

	fmt.Fprint(stdout, content)
	return nil
}

func wrapHTML(title, markdownContent string) string {
	escaped := htmlEscape(markdownContent)
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — Architecture Reference</title>
<style>
:root {
  --bg: #0d1117;
  --surface: #161b22;
  --border: #30363d;
  --text: #c9d1d9;
  --heading: #f0f6fc;
  --accent: #58a6ff;
  --code-bg: #0b0f14;
}
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background-color: var(--bg);
  color: var(--text);
  line-height: 1.6;
  padding: 2rem;
  max-width: 1100px;
  margin: 0 auto;
}
h1, h2, h3, h4 { color: var(--heading); margin-top: 1.5rem; margin-bottom: 0.75rem; }
h1 { font-size: 2.2rem; border-bottom: 2px solid var(--border); padding-bottom: 0.5rem; }
h2 { font-size: 1.6rem; border-bottom: 1px solid var(--border); padding-bottom: 0.4rem; }
pre { background: var(--code-bg); padding: 1rem; border-radius: 6px; overflow-x: auto; border: 1px solid var(--border); }
code { font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; font-size: 0.9em; }
table { width: 100%%; border-collapse: collapse; margin: 1.5rem 0; }
th, td { border: 1px solid var(--border); padding: 0.6rem 0.9rem; text-align: left; }
th { background: var(--surface); color: var(--heading); }
blockquote { border-left: 4px solid var(--accent); padding-left: 1rem; color: #8b949e; margin: 1rem 0; }
a { color: var(--accent); text-decoration: none; }
a:hover { text-decoration: underline; }
</style>
</head>
<body>
<div id="content"><pre><code>%s</code></pre></div>
</body>
</html>`, title, escaped)
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
