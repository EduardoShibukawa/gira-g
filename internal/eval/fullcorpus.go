package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// FullCorpusEnv opts in to the full pinned corpus (spec part C, "Full, opt-in").
// Any non-empty value enables it; FullCorpusCacheEnv optionally names the git cache.
const (
	FullCorpusEnv      = "GIRA_EVAL_FULL_CORPUS"
	FullCorpusCacheEnv = "GIRA_EVAL_CACHE_DIR"
)

// FullCorpusSkipReason returns "" when the full corpus is enabled, or the visible
// reason a test prints when it skips.
func FullCorpusSkipReason(ref CorpusRef) string {
	if os.Getenv(FullCorpusEnv) != "" {
		return ""
	}
	return fmt.Sprintf("full pinned corpus is opt-in: set %s=1 to fetch %s@%s through the git cache (optional %s=<dir>)",
		FullCorpusEnv, ref.UpstreamURL, ref.UpstreamSHA, FullCorpusCacheEnv)
}

// Upstream declares the pinned upstream repository a ground truth was derived from.
func (ref CorpusRef) Upstream() manifest.RepositoryDeclaration {
	return manifest.RepositoryDeclaration{ID: ref.RepoID, URL: ref.UpstreamURL, SHA: ref.UpstreamSHA, PathScope: "."}
}

// FetchFullCorpus writes the upstream tree at the pinned SHA into dstDir, minus the
// memory-leak paths, and returns it as an evaluation Corpus.
func FetchFullCorpus(ctx context.Context, ref CorpusRef, cm *gitcache.CacheManager, dstDir string) (Corpus, int, error) {
	if ref.UpstreamURL == "" || ref.UpstreamSHA == "" {
		return Corpus{}, 0, fmt.Errorf("ground truth does not pin an upstream url and sha")
	}
	n, err := corpus.FetchUpstream(ctx, cm, ref.Upstream(), dstDir, false)
	if err != nil {
		return Corpus{}, n, err
	}
	return Corpus{Dir: dstDir, RepoID: ref.RepoID, TargetSystem: ref.Name}, n, nil
}

// FileReader reads one corpus-relative file.
type FileReader func(path string) ([]byte, error)

// DirReader reads files from a corpus tree on disk.
func DirReader(dir string) FileReader {
	return func(p string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, filepath.FromSlash(p)))
	}
}

// CommitReader reads files from a cached repository at its pinned SHA.
func CommitReader(ctx context.Context, cm *gitcache.CacheManager, repo manifest.RepositoryDeclaration) FileReader {
	return func(p string) ([]byte, error) { return cm.ReadFile(ctx, repo, p) }
}

// CheckAnchors resolves every ground-truth anchor through read: the file must
// exist and contain the anchor string. It returns one
// violation per anchor that does not resolve, sorted.
func CheckAnchors(tasks []BenchmarkTask, read FileReader) (checked int, violations []string) {
	for _, t := range tasks {
		for _, a := range t.GroundTruth.Anchors {
			checked++
			data, err := read(a.Path)
			if err != nil {
				violations = append(violations, fmt.Sprintf("%s: %s: file does not resolve", t.ID, a.Path))
				continue
			}
			if !strings.Contains(string(data), a.Contains) {
				violations = append(violations, fmt.Sprintf("%s: %s: anchor %q not found", t.ID, a.Path, a.Contains))
			}
		}
	}
	sort.Strings(violations)
	return checked, violations
}

// Hermetic loads the published ground truth under repoRoot and its vendored
// corpus, for tests outside this package that need a real index.
func Hermetic(repoRoot string) (*GroundTruth, Corpus, error) {
	gt, err := LoadGroundTruth(filepath.Join(repoRoot, "fixtures", "evaluation", "reference-corpus", "ground-truth.v2.json"))
	if err != nil {
		return nil, Corpus{}, err
	}
	c := Corpus{Dir: filepath.Join(repoRoot, filepath.FromSlash(gt.Corpus.Snapshot)), RepoID: gt.Corpus.RepoID, TargetSystem: gt.Corpus.Name}
	return gt, c, nil
}
