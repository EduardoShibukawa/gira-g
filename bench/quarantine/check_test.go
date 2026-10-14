package quarantine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	benchDir = filepath.Join("..")
	repoRoot = filepath.Join("..", "..")
)

// SC12: no benchmark asset is referenced from the Go module, mise.toml, CI, git
// hooks or agent configuration of this repository.
func TestNoReferencesFromGuardedFiles(t *testing.T) {
	v, err := References(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) > 0 {
		t.Fatalf("benchmark referenced from guarded files:\n%s", strings.Join(v, "\n"))
	}
}

// SC12: the output directory is ignored by git.
func TestOutputDirIsIgnored(t *testing.T) {
	_, ph, err := PhaseRules(benchDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := OutputIgnored(repoRoot, filepath.Join("bench", ph.OutputDir)); err != nil {
		t.Fatal(err)
	}
}

// SC12: every query phase declares networking disabled; containers mount only the
// corpus copy and their output directory; the owner-run script matches.
func TestQueryPhasesDisableNetworking(t *testing.T) {
	v, ph, err := PhaseRules(benchDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) > 0 {
		t.Fatal(strings.Join(v, "\n"))
	}
	containers := 0
	for _, arm := range ph.Arms {
		if arm.Query != nil && arm.Query.Container {
			containers++
		}
	}
	if containers != 2 {
		t.Fatalf("want the Graphify and Understand Anything query phases in containers, got %d", containers)
	}
	script, err := os.ReadFile(filepath.Join(benchDir, "scripts", "container.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"docker run --rm -i --network none", `"$image_id"`, `D) mounts=(-v "$PWD/out/D:/out:ro")`} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("container.sh query phase does not match phases.json: missing %s", want)
		}
	}
}

func TestRecipesArePinned(t *testing.T) {
	_, ph, err := PhaseRules(benchDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, arm := range ph.Arms {
		if arm.Recipe == "" {
			continue
		}
		v, err := RecipePins(filepath.Join(benchDir, arm.Recipe))
		if err != nil {
			t.Fatal(err)
		}
		if len(v) > 0 {
			t.Error(strings.Join(v, "\n"))
		}
	}
}

// The check catches each kind of leak.
func TestCheckCatchesViolations(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"mise.toml":                          "[tasks.check]\nrun = \"cd bench && go test\"\n",
		".github/workflows/ci.yml":           "steps:\n  - run: pip install graphify\n",
		".claude/settings.json":              "{\"mcpServers\": {\"ua\": {\"command\": \"understand-anything\"}}}\n",
		"internal/x/x.go":                    "package x // see gira-bench\n",
		"docs/notes.md":                      "graphify is discussed here, which is fine\n",
		"bench/phases.json":                  "{}",
		"lefthook.yml":                       "pre-commit:\n  commands: {}\n",
		"go.work":                            "go 1.24.0\nuse ./bench\n",
		".github/workflows/b.yml":            "jobs:\n  x:\n    defaults:\n      run:\n        working-directory: bench\n",
		"internal/skills/catalog/x/SKILL.md": "run bench/layer1 first\n",
		"internal/eval/doc.go":               "package eval // the 12-task benchmark suite is fine\n",
	}
	for p, c := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	v, err := References(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".claude/settings.json:1", ".github/workflows/b.yml:5", ".github/workflows/ci.yml:2", "go.work:2", "internal/skills/catalog/x/SKILL.md:1", "internal/x/x.go:1", "mise.toml:2"}
	if len(v) != len(want) {
		t.Fatalf("got %v, want %v", v, want)
	}
	for i, w := range want {
		if !strings.HasPrefix(v[i], w+":") {
			t.Errorf("violation %d = %s, want %s", i, v[i], w)
		}
	}

	bad := filepath.Join(root, "Dockerfile")
	_ = os.WriteFile(bad, []byte("FROM python:3.14-slim\nADD https://github.com/x/y.git#main /src\n"), 0o644)
	rv, err := RecipePins(bad)
	if err != nil {
		t.Fatal(err)
	}
	if len(rv) != 2 {
		t.Fatalf("floating tag and branch pin must both fail: %v", rv)
	}
}

// E3: the walk ignores .claude/worktrees/, which embeds a copy of the
// repository, so a harness worktree never trips the static check. A guarded file
// outside it still does (control).
func TestWorktreesAreIgnored(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, ".claude", "worktrees", "x", "internal", "x", "x.go")
	if err := os.MkdirAll(filepath.Dir(inside), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inside, []byte("package x // see gira-bench\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	control := filepath.Join(root, "internal", "y", "y.go")
	if err := os.MkdirAll(filepath.Dir(control), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(control, []byte("package y // see gira-bench\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := References(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range v {
		if strings.Contains(line, ".claude/worktrees/") {
			t.Fatalf(".claude/worktrees/ must be ignored, got %q", line)
		}
	}
	if len(v) != 1 {
		t.Fatalf("control violation must be reported once, got %v", v)
	}
}
