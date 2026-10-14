// Package quarantine is the static isolation check of the comparative benchmark
// (ADR 0012 rule 3; spec SC12). It proves, from the files alone, that:
//   - nothing in the Go module, mise.toml, CI workflows, git hooks or agent
//     configuration references a benchmark asset or a benchmarked tool;
//   - the benchmark output directory is ignored by git;
//   - every container query phase declares networking disabled, and containers
//     mount only the corpus copy and their output directory;
//   - every container recipe pins its base image by digest and its source by commit.
package quarantine

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// forbidden matches a reference to a benchmark asset or to a benchmarked tool.
// "bench" counts as a path word wherever it stands alone ("cd bench", "./bench",
// "bench/x", "working-directory: bench"), but not inside "benchmark".
var forbidden = regexp.MustCompile(`(?i)(^|[^a-z0-9_-])\.?/?bench([^a-z0-9_-]|$)|gira-bench|graphify|understand[-_ ]?anything|egonex-ai`)

// Guarded reports whether a repository-relative path is one the benchmark must
// never be referenced from.
func Guarded(rel string) bool {
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "bench/") {
		return false
	}
	base := filepath.Base(rel)
	switch {
	case rel == "go.mod", rel == "go.sum", rel == "go.work", rel == "go.work.sum", strings.HasSuffix(rel, ".go"),
		strings.HasPrefix(rel, "internal/"), strings.HasPrefix(rel, "cmd/"):
		return true // the Go module, including files it embeds
	case rel == "mise.toml", rel == ".mise.toml", strings.HasPrefix(rel, ".mise/"):
		return true
	case strings.HasPrefix(rel, ".github/workflows/"), rel == ".gitlab-ci.yml":
		return true // CI
	case base == "lefthook.yml", base == "lefthook.yaml", strings.HasPrefix(rel, ".lefthook/"), strings.HasPrefix(rel, ".githooks/"), strings.HasPrefix(rel, ".husky/"),
		strings.HasPrefix(rel, ".git/hooks/") && !strings.HasSuffix(rel, ".sample"):
		return true // git hooks
	case base == "AGENTS.md", base == "CLAUDE.md", base == "GEMINI.md", rel == ".mcp.json",
		strings.HasPrefix(rel, ".claude/"), strings.HasPrefix(rel, ".agents/"), strings.HasPrefix(rel, ".cursor/"), strings.HasPrefix(rel, ".gemini/"):
		return true // agent configuration
	}
	return false
}

// References scans every guarded file under root and returns one violation per
// forbidden reference, as "path:line: match".
func References(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "bench" || rel == ".git/objects" || rel == ".git/refs" || rel == ".git/logs" {
				return filepath.SkipDir
			}
			return nil
		}
		if !Guarded(rel) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if m := forbidden.FindString(line); m != "" {
				out = append(out, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(m)))
			}
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

// OutputIgnored reports whether git ignores files in the benchmark output dir.
func OutputIgnored(repoRoot, outputDir string) error {
	probe := filepath.ToSlash(filepath.Join(outputDir, "probe.json"))
	cmd := exec.Command("git", "-C", repoRoot, "check-ignore", "-q", probe)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s is not ignored by git: %w", probe, err)
	}
	return nil
}

// Phases is the subset of phases.json the check reads.
type Phases struct {
	OutputDir string `json:"output_dir"`
	Arms      []struct {
		ID     string `json:"id"`
		Tool   string `json:"tool"`
		Recipe string `json:"recipe"`
		Build  *Phase `json:"build"`
		Query  *Phase `json:"query"`
	} `json:"arms"`
}

// Phase is one phase of one arm.
type Phase struct {
	Container bool    `json:"container"`
	Network   string  `json:"network"`
	Mounts    []Mount `json:"mounts"`
}

// Mount is a container bind mount, source relative to the benchmark directory.
type Mount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
}

// PhaseRules checks phases.json: every query phase disables networking, and
// containers mount only the read-only corpus copy and their own output directory.
func PhaseRules(benchDir string) ([]string, *Phases, error) {
	data, err := os.ReadFile(filepath.Join(benchDir, "phases.json"))
	if err != nil {
		return nil, nil, err
	}
	var ph Phases
	if err := json.Unmarshal(data, &ph); err != nil {
		return nil, nil, err
	}
	var out []string
	for _, arm := range ph.Arms {
		if arm.Query != nil && arm.Query.Network != "none" {
			out = append(out, fmt.Sprintf("arm %s: query phase network is %q, want \"none\"", arm.ID, arm.Query.Network))
		}
		for name, phase := range map[string]*Phase{"build": arm.Build, "query": arm.Query} {
			if phase == nil || !phase.Container {
				continue
			}
			for _, m := range phase.Mounts {
				switch {
				case m.Source == "corpus" && m.ReadOnly:
				case m.Source == filepath.ToSlash(filepath.Join(ph.OutputDir, arm.ID)):
				default:
					out = append(out, fmt.Sprintf("arm %s %s: mount %s is neither the read-only corpus nor %s/%s", arm.ID, name, m.Source, ph.OutputDir, arm.ID))
				}
			}
		}
	}
	sort.Strings(out)
	return out, &ph, nil
}

var (
	fromRE   = regexp.MustCompile(`(?m)^FROM\s+(\S+)`)
	syntaxRE = regexp.MustCompile(`(?m)^# syntax=(\S+)`)
	addGitRE = regexp.MustCompile(`(?m)^ADD\s+.*?(https://\S+\.git)(#\S+)?\s`)
	digestRE = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)
	commitRE = regexp.MustCompile(`^#[0-9a-f]{40}$`)
)

// RecipePins checks a Dockerfile: base images and the syntax frontend by digest,
// git sources by full commit.
func RecipePins(dockerfile string) ([]string, error) {
	data, err := os.ReadFile(dockerfile)
	if err != nil {
		return nil, err
	}
	text := string(data)
	var out []string
	for _, m := range append(fromRE.FindAllStringSubmatch(text, -1), syntaxRE.FindAllStringSubmatch(text, -1)...) {
		if !digestRE.MatchString(m[1]) {
			out = append(out, fmt.Sprintf("%s: image %s is not pinned by digest", dockerfile, m[1]))
		}
	}
	adds := addGitRE.FindAllStringSubmatch(text, -1)
	if len(adds) == 0 {
		out = append(out, fmt.Sprintf("%s: no pinned git source", dockerfile))
	}
	for _, m := range adds {
		if !commitRE.MatchString(m[2]) {
			out = append(out, fmt.Sprintf("%s: source %s is not pinned to a full commit", dockerfile, m[1]))
		}
	}
	return out, nil
}
