package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// SC2: `gira explain` shows an on-demand source excerpt from the local checkout,
// `--no-source` omits it, and a missing checkout reports "source unavailable".
func TestSCExplainSource(t *testing.T) {
	_, index := hermeticIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := slices.NewEngine(mm).ExplainEvidence("service:checkout", "")
	if err != nil {
		t.Fatal(err)
	}
	path := ""
	for _, ev := range rep.Evidence {
		if ev.Locator.Path != "" {
			path = ev.Locator.Path
			break
		}
	}
	if path == "" {
		t.Skip("no locator with a path for service:checkout")
	}

	repo := t.TempDir()
	full := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 1; i <= 200; i++ {
		fmt.Fprintf(&b, "SRCLINE %d\n", i)
	}
	if err := os.WriteFile(full, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "explain", "service:checkout", "--index", index, "--format", "md", "--repo", repo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "## Source Excerpts") || !strings.Contains(string(out), "SRCLINE") {
		t.Fatalf("expected a source excerpt:\n%s", out)
	}

	out2, err := runCLI(t, "explain", "service:checkout", "--index", index, "--format", "md", "--repo", repo, "--no-source")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out2), "## Source Excerpts") {
		t.Fatalf("--no-source must omit the section:\n%s", out2)
	}

	out3, err := runCLI(t, "explain", "service:checkout", "--index", index, "--format", "md", "--repo", filepath.Join(repo, "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out3), slices.SourceUnavailable) {
		t.Fatalf("missing checkout must report source unavailable:\n%s", out3)
	}

	// A Git checkout that does not contain the cited commit must not be used,
	// even when it has a same-named file (F2 review finding).
	grepo := t.TempDir()
	gitIn(t, grepo, "init", "-q")
	gfull := filepath.Join(grepo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(gfull), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gfull, []byte("OTHER CONTENT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, grepo, "add", "-A")
	gitIn(t, grepo, "commit", "-q", "-m", "unrelated")
	out4, err := runCLI(t, "explain", "service:checkout", "--index", index, "--format", "md", "--repo", grepo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out4), slices.SourceUnavailable) {
		t.Fatalf("an unrelated git checkout must not masquerade as the source:\n%s", out4)
	}
}
