package domain

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SC3: the index reports the commits it was built from and can tell whether a
// local HEAD is one of them.
func TestSCStaleness(t *testing.T) {
	mm := &MacroMap{
		Nodes: []Node{{Evidence: []EvidenceRecord{
			{Locator: Locator{Repo: "r", CommitSHA: "aaa", Path: "x"}},
		}}},
		Relationships: []Relationship{{Evidence: []EvidenceRecord{
			{Locator: Locator{Repo: "r2", CommitSHA: "bbb", Path: "y"}},
		}}},
	}
	got := mm.IndexedCommits()
	if len(got) != 2 || got[0] != "r2@bbb" || got[1] != "r@aaa" {
		t.Fatalf("IndexedCommits = %v", got)
	}
	if !mm.HasCommit("aaa") || !mm.HasCommit("bbb") {
		t.Fatal("HasCommit must find an indexed commit")
	}
	if mm.HasCommit("zzz") || mm.HasCommit("") {
		t.Fatal("HasCommit must reject unknown/empty shas")
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func initTestRepo(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	for p, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "initial")
	sha := runGit(t, dir, "rev-parse", "HEAD")
	return dir, sha
}

func TestInspectRepoFreshnessStates(t *testing.T) {
	dir, sha1 := initTestRepo(t, map[string]string{
		"src/a.go": "package a\n",
		"src/b.go": "package b\n",
	})

	mm := &MacroMap{
		Nodes: []Node{
			{
				ID:   "service:a",
				Kind: NodeService,
				Name: "a",
				Evidence: []EvidenceRecord{
					{Locator: Locator{Repo: "demo", CommitSHA: sha1, Path: "src/a.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
				},
			},
			{
				ID:   "service:b",
				Kind: NodeService,
				Name: "b",
				Evidence: []EvidenceRecord{
					{Locator: Locator{Repo: "demo", CommitSHA: sha1, Path: "src/b.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
				},
			},
		},
	}

	// 1. Current
	rep := InspectRepoFreshness(mm, "demo", dir)
	if rep.State != FreshnessCurrent || rep.Dirty || rep.AheadCount != 0 || rep.BehindCount != 0 {
		t.Fatalf("expected current state, got %+v", rep)
	}

	// 2. Dirty worktree
	if err := os.WriteFile(filepath.Join(dir, "src", "a.go"), []byte("package a // modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep = InspectRepoFreshness(mm, "demo", dir)
	if rep.State != FreshnessDirty || !rep.Dirty {
		t.Fatalf("expected dirty state, got %+v", rep)
	}

	// Per-evidence refresh on dirty worktree: src/a.go -> stale, src/b.go -> current
	mm.RefreshEvidenceFreshness(map[string]string{"demo": dir})
	if mm.Nodes[0].Evidence[0].Status != StatusStale {
		t.Fatalf("expected dirty file src/a.go evidence to be stale, got %s", mm.Nodes[0].Evidence[0].Status)
	}
	if mm.Nodes[1].Evidence[0].Status != StatusCurrent {
		t.Fatalf("expected untouched file src/b.go evidence to stay current, got %s", mm.Nodes[1].Evidence[0].Status)
	}

	// 3. Ahead (commit the change to src/a.go)
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "second")
	sha2 := runGit(t, dir, "rev-parse", "HEAD")

	rep = InspectRepoFreshness(mm, "demo", dir)
	if rep.State != FreshnessAhead || rep.AheadCount != 1 || rep.BehindCount != 0 || rep.Dirty {
		t.Fatalf("expected ahead (+1/-0) state, got %+v", rep)
	}

	// Per-evidence refresh when ahead: src/a.go changed between sha1 and sha2 -> stale; src/b.go unchanged -> current
	mm.RefreshEvidenceFreshness(map[string]string{"demo": dir})
	if mm.Nodes[0].Evidence[0].Status != StatusStale {
		t.Fatalf("expected modified file src/a.go to be stale at ahead HEAD, got %s", mm.Nodes[0].Evidence[0].Status)
	}
	if mm.Nodes[1].Evidence[0].Status != StatusCurrent {
		t.Fatalf("expected unmodified file src/b.go to remain current at ahead HEAD, got %s", mm.Nodes[1].Evidence[0].Status)
	}

	// 4. Behind (index cites sha2, checkout resets to sha1)
	mm2 := &MacroMap{
		Nodes: []Node{{
			ID: "service:a",
			Evidence: []EvidenceRecord{
				{Locator: Locator{Repo: "demo", CommitSHA: sha2, Path: "src/a.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
			},
		}},
	}
	runGit(t, dir, "checkout", sha1)
	rep = InspectRepoFreshness(mm2, "demo", dir)
	if rep.State != FreshnessBehind || rep.BehindCount != 1 || rep.AheadCount != 0 {
		t.Fatalf("expected behind (+0/-1) state, got %+v", rep)
	}

	// 5. Diverged (create a new commit from sha1 on a branch while index cites sha2)
	runGit(t, dir, "checkout", "-b", "diverged-branch", sha1)
	if err := os.WriteFile(filepath.Join(dir, "src", "b.go"), []byte("package b // diverged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "diverged commit")
	rep = InspectRepoFreshness(mm2, "demo", dir)
	if rep.State != FreshnessDiverged || rep.AheadCount != 1 || rep.BehindCount != 1 {
		t.Fatalf("expected diverged (+1/-1) state, got %+v", rep)
	}

	// 6. Missing commit (index cites a commit SHA not in the repository)
	mmMissing := &MacroMap{
		Nodes: []Node{{
			ID: "service:a",
			Evidence: []EvidenceRecord{
				{Locator: Locator{Repo: "demo", CommitSHA: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", Path: "src/a.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
			},
		}},
	}
	rep = InspectRepoFreshness(mmMissing, "demo", dir)
	if rep.State != FreshnessMissingCommit {
		t.Fatalf("expected missing_commit state, got %+v", rep)
	}
	mmMissing.RefreshEvidenceFreshness(map[string]string{"demo": dir})
	if mmMissing.Nodes[0].Evidence[0].Status != StatusUnverified {
		t.Fatalf("expected missing_commit evidence status to become unverified, got %s", mmMissing.Nodes[0].Evidence[0].Status)
	}

	// 7. Unverified when checkout directory is missing or unknown
	rep = InspectRepoFreshness(mm, "demo", filepath.Join(t.TempDir(), "does-not-exist"))
	if rep.State != FreshnessUnverified {
		t.Fatalf("expected unverified state for missing checkout dir, got %+v", rep)
	}
	mm.RefreshEvidenceFreshness(map[string]string{})
	if mm.Nodes[0].Evidence[0].Status != StatusUnverified || mm.Nodes[1].Evidence[0].Status != StatusUnverified {
		t.Fatalf("expected unverified evidence when checkout map is empty")
	}
}

func TestRefreshEvidenceFreshnessRemovedFile(t *testing.T) {
	dir, sha := initTestRepo(t, map[string]string{
		"src/keep.go":   "package main\n",
		"src/remove.go": "package main\n",
	})

	mm := &MacroMap{
		Nodes: []Node{
			{
				ID: "service:keep",
				Evidence: []EvidenceRecord{
					{Locator: Locator{Repo: "demo", CommitSHA: sha, Path: "src/keep.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
				},
			},
		},
		Relationships: []Relationship{
			{
				ID:       "rel:remove",
				SourceID: "service:keep",
				TargetID: "service:keep",
				Verb:     VerbConsumes,
				Evidence: []EvidenceRecord{
					{Locator: Locator{Repo: "demo", CommitSHA: sha, Path: "src/remove.go", StartLine: 1}, Origin: OriginExtracted, Status: StatusCurrent},
				},
			},
		},
	}

	// Remove src/remove.go from disk after indexing
	if err := os.Remove(filepath.Join(dir, "src", "remove.go")); err != nil {
		t.Fatal(err)
	}

	reports := mm.RefreshEvidenceFreshness(map[string]string{"demo": dir})
	if reports["demo"].State != FreshnessDirty {
		t.Fatalf("expected dirty repo state after file removal, got %+v", reports["demo"])
	}
	if mm.Nodes[0].Evidence[0].Status != StatusCurrent {
		t.Fatalf("expected kept file evidence to remain current, got %s", mm.Nodes[0].Evidence[0].Status)
	}
	if mm.Relationships[0].Evidence[0].Status != StatusStale {
		t.Fatalf("expected removed file evidence to transition to stale, got %s", mm.Relationships[0].Evidence[0].Status)
	}
}

