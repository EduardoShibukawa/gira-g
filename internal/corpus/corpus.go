// Package corpus materializes evaluation corpora as deterministic git repositories
// and derives Perturbed Variants from them. It exists so the evaluator indexes a
// real repository through the same path as `gira index`, never a pre-built map.
package corpus

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// fixedGitEnv pins every input of a commit hash: identity, dates and config.
var fixedGitEnv = []string{
	"GIT_AUTHOR_NAME=gira-eval",
	"GIT_AUTHOR_EMAIL=eval@gira.invalid",
	"GIT_COMMITTER_NAME=gira-eval",
	"GIT_COMMITTER_EMAIL=eval@gira.invalid",
	"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z",
	"GIT_COMMITTER_DATE=2026-01-01T00:00:00Z",
	"GIT_CONFIG_GLOBAL=/dev/null",
	"GIT_CONFIG_SYSTEM=/dev/null",
}

// Materialize copies srcDir into a fresh git repository at repoDir and commits it
// with fixed identity and dates, so the same tree always yields the same SHA.
// It returns the commit SHA.
func Materialize(ctx context.Context, srcDir, repoDir string) (string, error) {
	if err := copyTree(srcDir, repoDir); err != nil {
		return "", err
	}
	steps := [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "-A"},
		{"commit", "-q", "--no-gpg-sign", "-m", "corpus snapshot"},
	}
	for _, args := range steps {
		if _, err := git(ctx, repoDir, args...); err != nil {
			return "", err
		}
	}
	sha, err := git(ctx, repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(sha), nil
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fixedGitEnv...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %v: %s: %w", args, strings.TrimSpace(stderr.String()), err)
	}
	return stdout.String(), nil
}

func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// Rename maps one identifier spelling to another across paths and contents.
type Rename struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Edit replaces an exact snippet in one file, altering a relationship on purpose.
type Edit struct {
	Path    string `json:"path"`
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// Perturbation is the full, declarative recipe for a Perturbed Variant.
type Perturbation struct {
	Renames []Rename `json:"renames"`
	Edits   []Edit   `json:"edits"`
}

// Apply returns s with every rename applied, longest spelling first so that a
// rename of "CartService" never fires inside "ProductCartService" before its own.
func (p Perturbation) Apply(s string) string {
	pairs := make([]string, 0, 2*len(p.Renames))
	for _, r := range p.sortedRenames() {
		pairs = append(pairs, r.From, r.To)
	}
	return strings.NewReplacer(pairs...).Replace(s)
}

func (p Perturbation) sortedRenames() []Rename {
	rs := append([]Rename(nil), p.Renames...)
	sort.SliceStable(rs, func(i, j int) bool {
		if len(rs[i].From) != len(rs[j].From) {
			return len(rs[i].From) > len(rs[j].From)
		}
		return rs[i].From < rs[j].From
	})
	return rs
}

// Perturb writes the Perturbed Variant of srcDir into dstDir. Edits are applied to
// the published text first (their paths and snippets name the published corpus),
// then renames are applied to every path and every file's content. Every edit must
// match exactly once, so a drifted recipe fails loudly instead of silently.
func Perturb(srcDir, dstDir string, p Perturbation) error {
	edits := make(map[string][]Edit)
	for _, e := range p.Edits {
		edits[filepath.ToSlash(e.Path)] = append(edits[filepath.ToSlash(e.Path)], e)
	}
	applied := make(map[string]int)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		for _, e := range edits[rel] {
			if n := strings.Count(text, e.Find); n != 1 {
				return fmt.Errorf("perturbation edit on %s matches %d times, want 1: %q", rel, n, e.Find)
			}
			text = strings.Replace(text, e.Find, e.Replace, 1)
			applied[rel]++
		}
		target := filepath.Join(dstDir, filepath.FromSlash(p.Apply(rel)))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, []byte(p.Apply(text)), 0o644)
	})
	if err != nil {
		return err
	}
	for path, es := range edits {
		if applied[path] != len(es) {
			return fmt.Errorf("perturbation edit targets missing file %s", path)
		}
	}
	return nil
}
