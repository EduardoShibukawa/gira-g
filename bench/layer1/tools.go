package layer1

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BaseTools are the read-only tools every arm gets over the corpus copy: grep,
// glob and read. Arm A gets only these; arms B to D add their tool's surface.
// A read-only shell is not offered: it cannot be sandboxed from inside the runner,
// so its absence is recorded in every report rather than approximated.
type BaseTools struct{ Root string }

// Specs describes the base tools to a model.
func (b BaseTools) Specs() []ToolSpec {
	str := func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
	return []ToolSpec{
		{Name: "grep", Description: "Search corpus files for a case-insensitive substring; returns path:line:text.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"pattern": str("substring")}, "required": []string{"pattern"}}},
		{Name: "glob", Description: "List corpus files matching a glob such as src/*/main.go.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"pattern": str("glob")}, "required": []string{"pattern"}}},
		{Name: "read", Description: "Read one corpus file with line numbers.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"path": str("corpus-relative path")}, "required": []string{"path"}}},
	}
}

// Has reports whether name is a base tool.
func (b BaseTools) Has(name string) bool { return name == "grep" || name == "glob" || name == "read" }

// Run executes a base tool.
func (b BaseTools) Run(name string, args map[string]any) (string, error) {
	arg := func(k string) string { s, _ := args[k].(string); return s }
	switch name {
	case "grep":
		pat := strings.ToLower(arg("pattern"))
		if pat == "" {
			return "", fmt.Errorf("grep: empty pattern")
		}
		var out strings.Builder
		err := b.walk(func(rel string, data []byte) {
			for i, l := range strings.Split(string(data), "\n") {
				if strings.Contains(strings.ToLower(l), pat) {
					fmt.Fprintf(&out, "%s:%d:%s\n", rel, i+1, l)
				}
			}
		})
		return out.String(), err
	case "glob":
		var hits []string
		err := b.walk(func(rel string, _ []byte) {
			if ok, _ := filepath.Match(arg("pattern"), rel); ok {
				hits = append(hits, rel)
			}
		})
		return strings.Join(hits, "\n"), err
	case "read":
		p, err := b.inside(arg("path"))
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		var out strings.Builder
		for i, l := range strings.Split(string(data), "\n") {
			fmt.Fprintf(&out, "%d\t%s\n", i+1, l)
		}
		return out.String(), nil
	}
	return "", fmt.Errorf("unknown tool %q", name)
}

// inside resolves a corpus-relative path and refuses anything outside the corpus.
func (b BaseTools) inside(rel string) (string, error) {
	root, err := filepath.Abs(b.Root)
	if err != nil {
		return "", err
	}
	p := filepath.Join(root, filepath.FromSlash(rel))
	if p != root && !strings.HasPrefix(p, root+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the corpus", rel)
	}
	return p, nil
}

func (b BaseTools) walk(fn func(rel string, data []byte)) error {
	var files []string
	err := filepath.WalkDir(b.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, p := range files {
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(b.Root, p)
		fn(filepath.ToSlash(rel), data)
	}
	return nil
}

// CheckLocator is the Layer 0 locator rule over the corpus copy: the file exists
// and the range is in bounds.
func (b BaseTools) CheckLocator(rel string, start, end int) error {
	p, err := b.inside(rel)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("%s does not exist", rel)
	}
	lines := strings.Count(strings.TrimSuffix(string(data), "\n"), "\n") + 1
	if start < 1 || end < start || end > lines {
		return fmt.Errorf("%s#L%d-L%d out of bounds (%d lines)", rel, start, end, lines)
	}
	return nil
}
