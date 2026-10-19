// Package indexer turns a Target System declaration into a MacroMap. It is the
// single indexing path shared by `gira index` and the internal evaluator, so that
// evaluation always exercises the real repo-loading and extraction pipeline.
package indexer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Options controls a single indexing run.
type Options struct {
	// Offline forbids network egress; repositories must already be cached.
	Offline bool
	// ManifestPath is recorded as the locator of declarative override evidence.
	ManifestPath string
	// Pipeline overrides the default extractor pipeline (used by mutation tests).
	Pipeline *extractors.Pipeline
	// AllowShrink permits intentional architectural node or contract removal without file deletion.
	AllowShrink bool
}

// Build ensures every declared repository is cached at its pinned commit, runs the
// extraction pipeline over each repository's path scope, applies manifest overrides,
// and stamps evidence timestamps deterministically.
func Build(ctx context.Context, m *manifest.Manifest, cm *gitcache.CacheManager, opts Options) (*domain.MacroMap, error) {
	if m == nil {
		return nil, fmt.Errorf("manifest cannot be nil")
	}
	if cm == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}

	var overrideLocator string
	if m.HasDeclarativeOverrides() {
		var err error
		overrideLocator, err = manifest.ResolveManifestLocator(opts.ManifestPath)
		if err != nil {
			return nil, err
		}
	}

	commitTimes := make(map[string]time.Time, len(m.Repositories))
	var latest time.Time

	var targetFiles []extractors.TargetFile
	for _, repo := range m.Repositories {
		if err := cm.EnsureRepo(ctx, repo, opts.Offline); err != nil {
			return nil, fmt.Errorf("failed to ensure repository %q: %w", repo.ID, err)
		}

		ct, err := commitTime(ctx, cm.RepoDir(repo), repo.SHA)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit time for repository %q: %w", repo.ID, err)
		}
		commitTimes[repo.ID] = ct
		if ct.After(latest) {
			latest = ct
		}

		entries, err := cm.ListTree(ctx, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to list tree for repository %q: %w", repo.ID, err)
		}

		for _, entry := range entries {
			content, err := cm.ReadFile(ctx, repo, entry.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %q from repository %q: %w", entry.Path, repo.ID, err)
			}

			targetFiles = append(targetFiles, extractors.TargetFile{
				Repo:      repo.ID,
				CommitSHA: repo.SHA,
				Path:      entry.Path,
				Content:   content,
			})
		}
	}

	pipe := opts.Pipeline
	if pipe == nil {
		pipe = extractors.DefaultPipeline()
	}
	mm, err := pipe.ProcessFiles(ctx, m.TargetSystem.Name, targetFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to extract files: %w", err)
	}

	if m.Overrides != nil {
		if err := pipe.ApplyOverrides(mm, m.Overrides, overrideLocator); err != nil {
			return nil, fmt.Errorf("failed to apply overrides: %w", err)
		}
	}

	StampEvidence(mm, commitTimes, latest)
	return mm, nil
}

// StampEvidence replaces wall-clock discovery timestamps with the commit time of the
// repository each evidence record cites, so the same commits always index to the
// same bytes. Evidence without a repository (declarative overrides) gets the latest
// commit time across the Target System.
func StampEvidence(mm *domain.MacroMap, commitTimes map[string]time.Time, fallback time.Time) {
	stamp := func(evs []domain.EvidenceRecord) {
		for i := range evs {
			if t, ok := commitTimes[evs[i].Locator.Repo]; ok {
				evs[i].DiscoveredAt = t
			} else {
				evs[i].DiscoveredAt = fallback
			}
		}
	}
	for i := range mm.Nodes {
		stamp(mm.Nodes[i].Evidence)
	}
	for i := range mm.Relationships {
		stamp(mm.Relationships[i].Evidence)
	}
	for i := range mm.UnresolvedRefs {
		stamp(mm.UnresolvedRefs[i].Evidence)
	}
	for i := range mm.Conflicts {
		stamp(mm.Conflicts[i].Evidence)
	}
}

func commitTime(ctx context.Context, repoDir, sha string) (time.Time, error) {
	cmd := exec.CommandContext(ctx, "git", "--git-dir="+repoDir, "show", "-s", "--format=%cI", sha)
	cmd.Env = gitcache.GitEnvForRead()
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("git show %s: %w", sha, err)
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
