package eval

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// GrepBaseline simulates, deterministically, what an unassisted agent with grep and
// file reads must ingest to answer an impact question. It replaces the invented
// per-task constants of the v1 evaluator: the cost is measured on the actual corpus,
// so changing the corpus changes the baseline. The real agent baseline is Layer 1,
// arm A, of the comparative benchmark.
//
// Strategy (fixed, documented in the spec):
//  1. search the corpus for the target's identifier spellings;
//  2. read every matching file in full;
//  3. follow one hop: for each Component directory (src/<name>/) named in a matched
//     file, read that Component's files in full.
//
// The cost is the token estimate of the search output plus every file read once.
type GrepBaseline struct {
	files      map[string]string // corpus-relative path → content
	paths      []string          // sorted
	components map[string][]string
}

// NewGrepBaseline loads the corpus tree at dir.
func NewGrepBaseline(dir string) (*GrepBaseline, error) {
	b := &GrepBaseline{files: map[string]string{}, components: map[string][]string{}}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		b.files[rel] = string(data)
		b.paths = append(b.paths, rel)
		if parts := strings.SplitN(rel, "/", 3); len(parts) == 3 && parts[0] == "src" {
			b.components[parts[1]] = append(b.components[parts[1]], rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(b.paths)
	return b, nil
}

// Spellings returns the identifier spellings a grep agent would search for a
// gira node ID: the bare name, and hyphen-free and snake-case variants.
func Spellings(nodeID string) []string {
	name := nodeID
	if i := strings.LastIndex(name, ":"); i >= 0 {
		name = name[i+1:]
	}
	set := map[string]bool{
		strings.ToLower(name):                               true,
		strings.ToLower(strings.ReplaceAll(name, "-", "")):  true,
		strings.ToLower(strings.ReplaceAll(name, "-", "_")): true,
	}
	out := make([]string, 0, len(set))
	for s := range set {
		if s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// Tokens returns the measured baseline cost of answering for nodeID.
func (b *GrepBaseline) Tokens(nodeID string) int {
	terms := Spellings(nodeID)
	read := map[string]bool{}
	searchBytes := 0

	matches := func(content string) bool {
		lower := strings.ToLower(content)
		for _, t := range terms {
			if strings.Contains(lower, t) {
				return true
			}
		}
		return false
	}

	var matched []string
	for _, p := range b.paths {
		content := b.files[p]
		if !matches(content) {
			continue
		}
		matched = append(matched, p)
		read[p] = true
		for _, line := range strings.Split(content, "\n") {
			if matches(line) {
				searchBytes += len(p) + 1 + len(line) + 1 // "path:line\n" as grep prints it
			}
		}
	}

	names := make([]string, 0, len(b.components))
	for name := range b.components {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, p := range matched {
		lower := strings.ToLower(b.files[p])
		for _, name := range names {
			if strings.Contains(lower, strings.ToLower(name)) {
				for _, f := range b.components[name] {
					read[f] = true
				}
			}
		}
	}

	total := searchBytes
	for p := range read {
		total += len(b.files[p])
	}
	return tokens.EstimateLen(total)
}
