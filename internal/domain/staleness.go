package domain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RepoFreshnessState classifies the relationship between an indexed repository
// commit and a local git checkout.
type RepoFreshnessState string

const (
	FreshnessCurrent       RepoFreshnessState = "current"
	FreshnessStale         RepoFreshnessState = "stale"
	FreshnessUnverified    RepoFreshnessState = "unverified"
	FreshnessDirty         RepoFreshnessState = "dirty"
	FreshnessBehind        RepoFreshnessState = "behind"
	FreshnessAhead         RepoFreshnessState = "ahead"
	FreshnessDiverged      RepoFreshnessState = "diverged"
	FreshnessMissingCommit RepoFreshnessState = "missing_commit"
)

// RepoFreshnessReport summarizes how a repository's local checkout compares to
// the commits cited by the MacroMap.
type RepoFreshnessReport struct {
	Repo           string             `json:"repo"`
	IndexedCommits []string           `json:"indexed_commits"`
	HeadCommit     string             `json:"head_commit,omitempty"`
	Dirty          bool               `json:"dirty"`
	State          RepoFreshnessState `json:"state"`
	AheadCount     int                `json:"ahead_count,omitempty"`
	BehindCount    int                `json:"behind_count,omitempty"`
	Details        string             `json:"details,omitempty"`
}

// IndexedCommits returns the distinct "repo@commit" pairs the locators cite,
// sorted: the commits the index was actually built from. It walks node and
// relationship evidence (ADR 0001: locators only).
func (m *MacroMap) IndexedCommits() []string {
	if m == nil {
		return nil
	}
	seen := map[string]bool{}
	add := func(evs []EvidenceRecord) {
		for _, ev := range evs {
			if ev.Locator.CommitSHA == "" {
				continue
			}
			seen[ev.Locator.Repo+"@"+ev.Locator.CommitSHA] = true
		}
	}
	for _, n := range m.Nodes {
		add(n.Evidence)
	}
	for _, r := range m.Relationships {
		add(r.Evidence)
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// HasCommit reports whether sha is one of the commits the index cites.
func (m *MacroMap) HasCommit(sha string) bool {
	if m == nil || sha == "" {
		return false
	}
	has := func(evs []EvidenceRecord) bool {
		for _, ev := range evs {
			if ev.Locator.CommitSHA == sha {
				return true
			}
		}
		return false
	}
	for _, n := range m.Nodes {
		if has(n.Evidence) {
			return true
		}
	}
	for _, r := range m.Relationships {
		if has(r.Evidence) {
			return true
		}
	}
	return false
}

// repoIndexedSHAs returns the sorted distinct commit SHAs cited for repoName.
func (m *MacroMap) repoIndexedSHAs(repoName string) []string {
	if m == nil || repoName == "" {
		return nil
	}
	prefix := repoName + "@"
	var shas []string
	for _, entry := range m.IndexedCommits() {
		if strings.HasPrefix(entry, prefix) {
			sha := strings.TrimPrefix(entry, prefix)
			if sha != "" {
				shas = append(shas, sha)
			}
		}
	}
	sort.Strings(shas)
	return shas
}

// InspectRepoFreshness inspects a local git checkout against mm.IndexedCommits()
// using git rev-parse HEAD, git status --porcelain, git cat-file -e <sha>^{commit},
// and git rev-list --left-right --count <indexed>...HEAD to classify exact states.
func InspectRepoFreshness(mm *MacroMap, repoName, checkoutDir string) RepoFreshnessReport {
	indexed := mm.repoIndexedSHAs(repoName)
	report := RepoFreshnessReport{
		Repo:           repoName,
		IndexedCommits: indexed,
		State:          FreshnessUnverified,
	}
	if mm == nil || len(indexed) == 0 {
		report.Details = "no indexed commits recorded for repository"
		return report
	}
	if strings.TrimSpace(checkoutDir) == "" {
		report.Details = "checkout directory not provided"
		return report
	}
	info, err := os.Stat(checkoutDir)
	if err != nil || !info.IsDir() {
		report.Details = "checkout directory does not exist"
		return report
	}

	headOut, err := exec.Command("git", "-C", checkoutDir, "rev-parse", "HEAD").Output()
	if err != nil {
		report.Details = fmt.Sprintf("git rev-parse HEAD failed: %v", err)
		return report
	}
	headSHA := strings.TrimSpace(string(headOut))
	report.HeadCommit = headSHA

	statusOut, err := exec.Command("git", "-C", checkoutDir, "status", "--porcelain").Output()
	if err != nil {
		report.Details = fmt.Sprintf("git status --porcelain failed: %v", err)
		return report
	}
	report.Dirty = strings.TrimSpace(string(statusOut)) != ""

	primarySHA := indexed[0]
	for _, sha := range indexed {
		if err := exec.Command("git", "-C", checkoutDir, "cat-file", "-e", sha+"^{commit}").Run(); err != nil {
			report.State = FreshnessMissingCommit
			report.Details = fmt.Sprintf("indexed commit %s not found in local git history", sha)
			return report
		}
	}

	revOut, err := exec.Command("git", "-C", checkoutDir, "rev-list", "--left-right", "--count", primarySHA+"...HEAD").Output()
	if err != nil {
		// Fallback to two-dot syntax if needed
		revOut, err = exec.Command("git", "-C", checkoutDir, "rev-list", "--left-right", "--count", primarySHA+"..HEAD").Output()
		if err != nil {
			report.Details = fmt.Sprintf("git rev-list failed: %v", err)
			return report
		}
	}
	fields := strings.Fields(strings.TrimSpace(string(revOut)))
	if len(fields) == 2 {
		behind, _ := strconv.Atoi(fields[0])
		ahead, _ := strconv.Atoi(fields[1])
		report.BehindCount = behind
		report.AheadCount = ahead
	}

	switch {
	case report.BehindCount > 0 && report.AheadCount > 0:
		report.State = FreshnessDiverged
		report.Details = fmt.Sprintf("HEAD diverged from indexed commit %s (+%d/-%d)", primarySHA, report.AheadCount, report.BehindCount)
	case report.BehindCount > 0:
		report.State = FreshnessBehind
		report.Details = fmt.Sprintf("HEAD is %d commit(s) behind indexed commit %s", report.BehindCount, primarySHA)
	case report.AheadCount > 0:
		report.State = FreshnessAhead
		report.Details = fmt.Sprintf("HEAD is %d commit(s) ahead of indexed commit %s", report.AheadCount, primarySHA)
	case report.Dirty:
		report.State = FreshnessDirty
		report.Details = "working tree has uncommitted changes at indexed commit"
	default:
		report.State = FreshnessCurrent
		report.Details = "HEAD matches indexed commit and working tree is clean"
	}
	return report
}

// RefreshEvidenceFreshness updates each EvidenceRecord.Status in mm.Nodes and
// mm.Relationships to StatusCurrent, StatusStale (when the file changed between
// ev.Locator.CommitSHA and HEAD or worktree is dirty/removed on that file), or
// StatusUnverified (when the commit or checkout is missing/unreachable).
func (m *MacroMap) RefreshEvidenceFreshness(checkouts map[string]string) map[string]RepoFreshnessReport {
	if m == nil {
		return nil
	}
	repos := map[string]bool{}
	for _, entry := range m.IndexedCommits() {
		if repo, _, ok := strings.Cut(entry, "@"); ok && repo != "" {
			repos[repo] = true
		}
	}
	for repo := range checkouts {
		if repo != "" {
			repos[repo] = true
		}
	}

	reports := make(map[string]RepoFreshnessReport, len(repos))
	for repo := range repos {
		reports[repo] = InspectRepoFreshness(m, repo, checkouts[repo])
	}

	type fileKey struct {
		repo      string
		commitSHA string
		path      string
	}
	cache := map[fileKey]FreshnessStatus{}

	evalEvidence := func(ev EvidenceRecord) FreshnessStatus {
		loc := ev.Locator
		if loc.Repo == "" || loc.CommitSHA == "" || loc.Path == "" {
			return StatusUnverified
		}
		key := fileKey{repo: loc.Repo, commitSHA: loc.CommitSHA, path: loc.Path}
		if st, ok := cache[key]; ok {
			return st
		}

		checkoutDir := checkouts[loc.Repo]
		rep, hasRep := reports[loc.Repo]
		if !hasRep || checkoutDir == "" || rep.State == FreshnessUnverified || rep.State == FreshnessMissingCommit {
			cache[key] = StatusUnverified
			return StatusUnverified
		}

		// Verify this specific commit exists in the checkout.
		if err := exec.Command("git", "-C", checkoutDir, "cat-file", "-e", loc.CommitSHA+"^{commit}").Run(); err != nil {
			cache[key] = StatusUnverified
			return StatusUnverified
		}

		relPath := filepath.ToSlash(loc.Path)

		// Check if file exists on disk in the checkout worktree.
		diskPath := filepath.Join(checkoutDir, filepath.FromSlash(relPath))
		if _, err := os.Stat(diskPath); err != nil {
			cache[key] = StatusStale
			return StatusStale
		}

		// Check if file has uncommitted changes in the worktree/index.
		porcelain, err := exec.Command("git", "-C", checkoutDir, "status", "--porcelain", "--", relPath).Output()
		if err != nil {
			cache[key] = StatusUnverified
			return StatusUnverified
		}
		if strings.TrimSpace(string(porcelain)) != "" {
			cache[key] = StatusStale
			return StatusStale
		}

		// Check if file changed between loc.CommitSHA and HEAD.
		if loc.CommitSHA != rep.HeadCommit {
			diffOut, err := exec.Command("git", "-C", checkoutDir, "diff", "--name-only", loc.CommitSHA, "HEAD", "--", relPath).Output()
			if err != nil {
				cache[key] = StatusUnverified
				return StatusUnverified
			}
			if strings.TrimSpace(string(diffOut)) != "" {
				cache[key] = StatusStale
				return StatusStale
			}
		}

		cache[key] = StatusCurrent
		return StatusCurrent
	}

	for i := range m.Nodes {
		for j := range m.Nodes[i].Evidence {
			m.Nodes[i].Evidence[j].Status = evalEvidence(m.Nodes[i].Evidence[j])
		}
	}
	for i := range m.Relationships {
		for j := range m.Relationships[i].Evidence {
			m.Relationships[i].Evidence[j].Status = evalEvidence(m.Relationships[i].Evidence[j])
		}
	}

	return reports
}

