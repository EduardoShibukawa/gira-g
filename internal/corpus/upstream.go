package corpus

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Leaks reports whether a corpus path must be removed before any tool indexes the
// corpus (ADR 0007 memory-leak rule, spec part B): README*, AGENTS.md and docs/.
func Leaks(p string) bool {
	p = path.Clean(filepath.ToSlash(p))
	for _, seg := range strings.Split(path.Dir(p), "/") {
		if seg == "docs" {
			return true
		}
	}
	base := path.Base(p)
	return strings.HasPrefix(strings.ToUpper(base), "README") || base == "AGENTS.md"
}

// FetchUpstream ensures repo is in the git cache at its pinned SHA and writes its
// tree into dstDir, without the paths Leaks removes and without symlinks (their
// blob is a link target, not content). It returns the number of files written.
func FetchUpstream(ctx context.Context, cm *gitcache.CacheManager, repo manifest.RepositoryDeclaration, dstDir string, offline bool) (int, error) {
	if err := cm.EnsureRepo(ctx, repo, offline); err != nil {
		return 0, err
	}
	entries, err := cm.ListTree(ctx, repo)
	if err != nil {
		return 0, err
	}
	written := 0
	for _, e := range entries {
		if e.Mode == "120000" || Leaks(e.Path) {
			continue
		}
		data, err := cm.ReadFile(ctx, repo, e.Path)
		if err != nil {
			return written, err
		}
		target := filepath.Join(dstDir, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return written, fmt.Errorf("writing %s: %w", e.Path, err)
		}
		written++
	}
	return written, nil
}
