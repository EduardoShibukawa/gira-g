package mcp_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSCGateLintStrictness verifies the gate-lint CI wiring is warn-only:
// the workflow runs gate-lint with continue-on-error so drift warns, not fails.
func TestSCGateLintStrictness(t *testing.T) {
	ciPath := filepath.Join("..", "..", ".github", "workflows", "ci.yml")
	content, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("failed to read CI workflow: %v", err)
	}
	ci := string(content)
	if !strings.Contains(ci, "gate-lint") {
		t.Fatal("CI workflow missing gate-lint job")
	}
	if !strings.Contains(ci, "continue-on-error: true") {
		t.Fatal("gate-lint must be warn-only (continue-on-error: true)")
	}
}

// TestCICIConfiguration verifies the CI workflow wires the validator, gate-lint
// and the opt-in floor guard on ubuntu-latest. The quarantined comparative
// module is intentionally absent (ADR 0012; enforced by the static check).
func TestCICIConfiguration(t *testing.T) {
	ciPath := filepath.Join("..", "..", ".github", "workflows", "ci.yml")
	content, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("failed to read CI workflow: %v", err)
	}
	ciContent := string(content)

	// Validator step.
	if !strings.Contains(ciContent, "validate.js") {
		t.Fatal("CI workflow missing validate.js step")
	}

	// Gate-lint step.
	if !strings.Contains(ciContent, "gate-lint") {
		t.Fatal("CI workflow missing gate-lint step")
	}

	// The floor guard is opt-in and env-gated.
	if !strings.Contains(ciContent, "E3_FLOOR_GUARD") {
		t.Fatal("CI workflow missing the E3_FLOOR_GUARD opt-in floor guard")
	}

	// Validator, gate-lint and floor guard run on ubuntu-latest only (Q3).
	if !strings.Contains(ciContent, "ubuntu-latest") {
		t.Fatal("CI workflow missing ubuntu-latest runner")
	}
}

// TestHygieneGitignore verifies .gitignore includes results.tsv and run.log
func TestHygieneGitignore(t *testing.T) {
	gitignorePath := filepath.Join("..", "..", ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	gitignore := string(content)

	if !strings.Contains(gitignore, "results.tsv") {
		t.Fatal(".gitignore missing results.tsv")
	}
	if !strings.Contains(gitignore, "run.log") {
		t.Fatal(".gitignore missing run.log")
	}
}

// TestHygieneREADMELink verifies README.md:22 link is fixed
func TestHygieneREADMELink(t *testing.T) {
	readmePath := filepath.Join("..", "..", "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	readme := string(content)

	lines := strings.Split(readme, "\n")
	if len(lines) < 22 {
		t.Fatal("README.md has fewer than 22 lines")
	}
	line22 := lines[21] // 0-indexed
	if strings.Contains(line22, ".scratch/system-context-harness/map.md") {
		t.Fatal("README.md:22 still has broken link to .scratch/system-context-harness/map.md")
	}
}

const section61TaskRecord = `---
schema: task/1
tier: standard
status: active
---

# Task X

## Routing and approvals

- tier: standard

## Objective

x

## Acceptance criteria

- AC1 = something

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | build | pass | HEAD | ok |
`

const section61SpecRecord = `---
schema: spec/1
status: draft
version: 1
---

# Spec S

## Problem
p
## Expected behavior
e
## Scope
s
## Acceptance criteria
- SC1 x
## Interfaces and dependencies
i
## Decisions
d
## Risks and constraints
r
## Verification
v
## Rollback
rb
## Open questions
oq
## Revisions
- v1
`

// TestValidatorSection61NonLinearHistory is the regression guard for the §6.1
// false positive: in non-linear history the walk compared a record against
// state from a parallel branch and reported it "deleted" although it is present
// at HEAD. A record counts as deleted only when one of its direct parents has it.
// The positive control (a genuine deletion) must still be caught, so the rule
// cannot be fixed by disabling it.
func TestValidatorSection61NonLinearHistory(t *testing.T) {
	runtime := validatorRuntime(t)
	dir := t.TempDir()

	gitRun(t, dir, nil, "init", "-q", "-b", "main")
	gitRun(t, dir, nil, "config", "user.email", "t@t")
	gitRun(t, dir, nil, "config", "user.name", "t")
	gitRun(t, dir, dateEnv("2026-01-01T10:00:00"), "commit", "-q", "--allow-empty", "-m", "base")

	// branch a (earlier): adds a committed task record
	gitRun(t, dir, nil, "checkout", "-q", "-b", "a")
	writeFile(t, dir, "tasks/2026-01-01-x.md", section61TaskRecord)
	gitRun(t, dir, dateEnv("2026-01-01T11:00:00"), "add", "-A")
	gitRun(t, dir, dateEnv("2026-01-01T11:00:00"), "commit", "-q", "-m", "a: add task x")

	// branch b (later): touches specs/, never has tasks/x.md
	gitRun(t, dir, nil, "checkout", "-q", "main")
	gitRun(t, dir, nil, "checkout", "-q", "-b", "b")
	writeFile(t, dir, "specs/2026-01-01-s.md", section61SpecRecord)
	gitRun(t, dir, dateEnv("2026-01-01T12:00:00"), "add", "-A")
	gitRun(t, dir, dateEnv("2026-01-01T12:00:00"), "commit", "-q", "-m", "b: add spec s")

	// merge both back into main (non-linear history)
	gitRun(t, dir, nil, "checkout", "-q", "main")
	gitRun(t, dir, dateEnv("2026-01-01T13:00:00"), "merge", "-q", "--no-ff", "a", "-m", "merge a")
	gitRun(t, dir, dateEnv("2026-01-01T14:00:00"), "merge", "-q", "--no-ff", "b", "-m", "merge b")
	writeFile(t, dir, "manifest.json", `{"files":{}}`+"\n")

	out := runValidator(t, runtime, dir)
	if !strings.Contains(out, "artifact-records") {
		t.Fatalf("validator did not walk records; output:\n%s", out)
	}
	if strings.Contains(out, "was deleted") {
		t.Fatalf("false §6.1 deletion on non-linear history; output:\n%s", out)
	}

	// positive control: a genuine deletion is still caught
	gitRun(t, dir, nil, "checkout", "-q", "-b", "del")
	gitRun(t, dir, nil, "rm", "-q", "tasks/2026-01-01-x.md")
	gitRun(t, dir, dateEnv("2026-01-01T15:00:00"), "commit", "-q", "-m", "delete task x")
	out = runValidator(t, runtime, dir)
	if !strings.Contains(out, "was deleted") || !strings.Contains(out, "§6.1") {
		t.Fatalf("genuine deletion must still be caught; output:\n%s", out)
	}
}

// validatorRuntime returns a JS runtime able to run the bundled validator, or
// skips the test when neither node nor bun is installed.
func validatorRuntime(t *testing.T) string {
	t.Helper()
	for _, r := range []string{"node", "bun"} {
		if p, err := exec.LookPath(r); err == nil {
			return p
		}
	}
	t.Skip("neither node nor bun is installed; skipping validator history-walk regression")
	return ""
}

func runValidator(t *testing.T, runtime, root string) string {
	t.Helper()
	js, err := filepath.Abs(filepath.Join("..", "..", "tools", "validate", "validate.js"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(runtime, js, "--root", root)
	out, _ := cmd.CombinedOutput() // exit 1 on findings is expected
	return string(out)
}

func gitRun(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func dateEnv(d string) []string {
	return []string{"GIT_AUTHOR_DATE=" + d, "GIT_COMMITTER_DATE=" + d}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
