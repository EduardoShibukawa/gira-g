package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/docsite"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/store"
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

	manifestPath := findProjectManifest()
	var targetName string
	var targetDesc string
	if manifestPath != "" {
		if raw, err := os.ReadFile(manifestPath); err == nil {
			if m, err := manifest.Parse(raw); err == nil {
				targetName = m.TargetSystem.Name
				targetDesc = m.TargetSystem.Description
			}
		}
	}

	isDogfoodOrSelf := targetName == "gira-g" || strings.Contains(strings.ToLower(targetName), "dogfood")

	if title == "Gira-G Architecture Harness" && targetName != "" && !isDogfoodOrSelf {
		if targetDesc != "" {
			title = targetDesc
		} else {
			title = targetName
		}
	}

	tagline := targetDesc
	if tagline == "" || isDogfoodOrSelf {
		tagline = "Deterministic Code Intelligence & Living Architecture"
	}

	logoLetter := "G"
	if targetName != "" && !isDogfoodOrSelf {
		runes := []rune(targetName)
		if len(runes) > 0 {
			logoLetter = strings.ToUpper(string(runes[0]))
		}
	}

	githubURL := "https://github.com/EduardoShibukawa/gira-g"
	if remote := gitRemoteURL(root); remote != "" {
		if web := gitToWebURL(remote); web != "" {
			githubURL = web
		}
	}

	var mm *domain.MacroMap
	if indexPath != "" {
		st := store.NewStore(filepath.Dir(indexPath))
		if loaded, err := st.LoadFile(indexPath); err == nil {
			mm = loaded
		}
	} else if manifestPath != "" {
		res, err := store.Resolve(context.Background(), store.ResolveOptions{
			ManifestPath: manifestPath,
			WarnWriter:   io.Discard,
		})
		if err == nil && res != nil {
			mm = res.MacroMap
		}
	}

	cfg := &docsite.SiteConfig{
		Title:      title,
		Tagline:    tagline,
		Version:    "v0.3.0",
		GitHubURL:  githubURL,
		LogoLetter: logoLetter,
		TargetName: targetName,
		TargetDesc: targetDesc,
		RepoRoot:   root,
		OutputDir:  absOutput,
		IndexPath:  indexPath,
		MacroMap:   mm,
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

func gitToWebURL(raw string) string {
	u := strings.TrimSpace(raw)
	if strings.HasPrefix(u, "git@github.com:") {
		u = "https://github.com/" + strings.TrimPrefix(u, "git@github.com:")
	}
	u = strings.TrimSuffix(u, ".git")
	return u
}
