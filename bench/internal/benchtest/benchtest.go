// Package benchtest builds the fixtures benchmark tests share: the hermetic corpus
// as a git repository, a gira binary built from this repository, and its index.
package benchtest

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// RepoRoot is the repository root, seen from a package directory under bench/.
var RepoRoot = filepath.Join("..", "..")

// Run runs a command in dir with a fixed git identity and dates.
func Run(t testing.TB, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=b", "GIT_AUTHOR_EMAIL=b@b.invalid", "GIT_COMMITTER_NAME=b", "GIT_COMMITTER_EMAIL=b@b.invalid",
		"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z", "GIT_CONFIG_GLOBAL=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// Corpus is the hermetic corpus indexed by gira, as arm B's build phase does it.
type Corpus struct {
	Repo, SHA, Gira, Index string
}

// Hermetic copies the vendored corpus into a git repository, builds gira and
// indexes the repository with `gira index`.
func Hermetic(t testing.TB) Corpus {
	t.Helper()
	tmp := t.TempDir()
	c := Corpus{Repo: filepath.Join(tmp, "corpus")}
	src := filepath.Join(RepoRoot, "testdata", "corpus", "otel-demo-3.1.0-mini")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(c.Repo, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(c.Repo, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	Run(t, c.Repo, "git", "init", "-q", "-b", "main")
	Run(t, c.Repo, "git", "add", "-A")
	Run(t, c.Repo, "git", "commit", "-q", "--no-gpg-sign", "-m", "corpus")
	c.SHA = Run(t, c.Repo, "git", "rev-parse", "HEAD")

	abs, _ := filepath.Abs(RepoRoot)
	c.Gira = filepath.Join(tmp, "gira")
	Run(t, abs, "go", "build", "-o", c.Gira, "./cmd/gira")

	manifest := filepath.Join(tmp, "manifest.yaml")
	m := fmt.Sprintf("version: 1\ntarget_system:\n  name: opentelemetry-demo\nrepositories:\n  - id: opentelemetry-demo\n    url: %q\n    sha: %q\n    path_scope: \".\"\n", c.Repo, c.SHA)
	if err := os.WriteFile(manifest, []byte(m), 0o644); err != nil {
		t.Fatal(err)
	}
	c.Index = filepath.Join(tmp, "macro-map.json")
	Run(t, tmp, c.Gira, "index", manifest, "--output", c.Index, "--cache-dir", filepath.Join(tmp, "cache"))
	return c
}
