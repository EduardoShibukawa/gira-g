package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

type impactSet struct {
	Up, Down, Contracts []string
}

func impactSetOf(t *testing.T, out []byte) impactSet {
	t.Helper()
	var b slices.ImpactBriefing
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	ids := func(ns []domain.Node) []string {
		out := []string{}
		for _, n := range ns {
			out = append(out, n.ID)
		}
		return out
	}
	return impactSet{ids(b.UpstreamServices), ids(b.DownstreamServices), ids(b.AffectedContracts)}
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.invalid", "GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_AUTHOR_DATE=2026-01-02T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-02T00:00:00Z")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s", args, out)
	}
	return strings.TrimSpace(string(out))
}

// SC7: a change to the shared proto's PaymentService definition yields, through
// --from-diff and get_pr_impact, the Impact Set of the direct Contract target.
func TestSC7ImpactFromDiffMatchesContractTarget(t *testing.T) {
	_, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	_, index := hermeticIndex(t)

	repo := filepath.Join(t.TempDir(), "repo")
	base, err := corpus.Materialize(context.Background(), c.Dir, repo)
	if err != nil {
		t.Fatal(err)
	}
	proto := filepath.Join(repo, "pb", "demo.proto")
	data, err := os.ReadFile(proto)
	if err != nil {
		t.Fatal(err)
	}
	charge := "    rpc Charge(ChargeRequest) returns (ChargeResponse) {}\n"
	if strings.Count(string(data), charge) != 1 {
		t.Fatal("fixture drifted: PaymentService.Charge not found once")
	}
	edited := strings.Replace(string(data), charge, charge+"    rpc Refund(ChargeRequest) returns (ChargeResponse) {}\n", 1)
	if err := os.WriteFile(proto, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "commit", "-q", "--no-gpg-sign", "-am", "add PaymentService.Refund")
	head := gitIn(t, repo, "rev-parse", "HEAD")
	rng := base + ".." + head

	direct, err := runCLI(t, "impact", "contract:PaymentService", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	want := impactSetOf(t, direct)
	if len(want.Up)+len(want.Down) == 0 {
		t.Fatalf("direct briefing is empty, the fixture proves nothing: %+v", want)
	}

	cli, err := runCLI(t, "impact", "--from-diff", rng, "--repo", repo, "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if got := impactSetOf(t, cli); !reflect.DeepEqual(got, want) {
		t.Fatalf("--from-diff impact set\n got %+v\nwant %+v", got, want)
	}
	var b slices.ImpactBriefing
	_ = json.Unmarshal(cli, &b)
	if b.Diff == nil || !reflect.DeepEqual(b.Diff.Targets, []string{"contract:PaymentService"}) || len(b.Diff.UnmatchedPaths) != 0 {
		t.Fatalf("diff targets: %+v", b.Diff)
	}

	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))
	text, isErr := mcpCall(t, server, "get_pr_impact", map[string]any{"range": rng, "repo_path": repo})
	if isErr {
		t.Fatal(text)
	}
	if got := impactSetOf(t, []byte(text)); !reflect.DeepEqual(got, want) {
		t.Fatalf("get_pr_impact impact set\n got %+v\nwant %+v", got, want)
	}
	if text != string(cli) {
		t.Fatal("get_pr_impact and --from-diff must emit the same JSON")
	}

	diffText := gitIn(t, repo, "diff", "-U0", rng)
	text, isErr = mcpCall(t, server, "get_pr_impact", map[string]any{"range": rng, "diff": diffText + "\n"})
	if isErr || text != string(cli) {
		t.Fatalf("get_pr_impact with a supplied diff differs: %s", text)
	}

	budgeted, err := runCLI(t, "impact", "--from-diff", rng, "--repo", repo, "--index", index, "--budget", "1500")
	if err != nil {
		t.Fatal(err)
	}
	checkBudget(t, "from-diff budget", budgeted, 1500, slices.FormatJSON, "diff:"+rng)
}

func TestFromDiffReportsUnmatchedPaths(t *testing.T) {
	_, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	_, index := hermeticIndex(t)
	repo := filepath.Join(t.TempDir(), "repo")
	base, err := corpus.Materialize(context.Background(), c.Dir, repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "UNKNOWN.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-q", "--no-gpg-sign", "-m", "unknown file")
	out, err := runCLI(t, "impact", "--from-diff", base+"..HEAD", "--repo", repo, "--index", index, "--format", "md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Unknown to the index: `UNKNOWN.txt`") {
		t.Fatalf("unmatched path not reported:\n%s", out)
	}
}

// Review F3: a change inside a file the index cites, but outside every cited
// range, falls back to the Components citing the file; it is never reported as
// unknown to the index.
func TestFromDiffFileLevelFallback(t *testing.T) {
	_, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	_, index := hermeticIndex(t)
	repo := filepath.Join(t.TempDir(), "repo")
	base, err := corpus.Materialize(context.Background(), c.Dir, repo)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(repo, "src", "checkout", "main.go")
	data, _ := os.ReadFile(p)
	lines := strings.Split(string(data), "\n")
	lines[399] += " // touched"
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "commit", "-q", "--no-gpg-sign", "-am", "edit checkout logic")
	out, err := runCLI(t, "impact", "--from-diff", base+"..HEAD", "--repo", repo, "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	var b slices.ImpactBriefing
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Diff.UnmatchedPaths) != 0 || !reflect.DeepEqual(b.Diff.FileLevelPaths, []string{"src/checkout/main.go"}) {
		t.Fatalf("diff mapping: %+v", b.Diff)
	}
	found := false
	for _, id := range b.Diff.Targets {
		found = found || id == "service:checkout"
	}
	if !found || len(b.UpstreamServices) == 0 {
		t.Fatalf("file-level fallback must target service:checkout: %+v", b.Diff)
	}
}
