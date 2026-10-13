package eval_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
)

// GT review, AC2 (hermetic): every anchor resolves in the vendored snapshot, which
// is an unmodified subset of the pinned upstream commit.
func TestAnchorsResolveInSnapshot(t *testing.T) {
	gt := loadPublished(t)
	checked, violations := eval.CheckAnchors(gt.Tasks, eval.DirReader(publishedCorpus(gt).Dir))
	if checked == 0 {
		t.Fatal("ground truth declares no anchors")
	}
	if len(violations) > 0 {
		t.Fatalf("%d of %d anchors do not resolve: %v", len(violations), checked, violations)
	}
}

func TestCheckAnchorsRejectsDriftedAnchor(t *testing.T) {
	gt := loadPublished(t)
	tasks := append([]eval.BenchmarkTask(nil), gt.Tasks[0])
	tasks[0].GroundTruth.Anchors = []eval.Anchor{
		{Path: "src/checkout/main.go", Contains: "this string is not in the file"},
		{Path: "src/checkout/missing.go", Contains: "package main"},
	}
	_, violations := eval.CheckAnchors(tasks, eval.DirReader(publishedCorpus(gt).Dir))
	if len(violations) != 2 {
		t.Fatalf("want 2 violations, got %v", violations)
	}
}

// GT review, AC1: the full pinned corpus runs only when opted in, and otherwise
// skips with a visible reason. Opted in, it checks the anchors and the vendored
// snapshot against the upstream commit, then runs the published suite over the
// full tree and writes its report.
func TestFullCorpusOptIn(t *testing.T) {
	gt := loadPublished(t)
	if reason := eval.FullCorpusSkipReason(gt.Corpus); reason != "" {
		t.Skip(reason)
	}
	ctx := context.Background()
	cm, err := gitcache.NewCacheManager(os.Getenv(eval.FullCorpusCacheEnv))
	if err != nil {
		t.Fatal(err)
	}
	full, files, err := eval.FetchFullCorpus(ctx, gt.Corpus, cm, filepath.Join(t.TempDir(), "full"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("fetched %d files of %s@%s", files, gt.Corpus.UpstreamURL, gt.Corpus.UpstreamSHA)

	upstream := eval.CommitReader(ctx, cm, gt.Corpus.Upstream())
	checked, violations := eval.CheckAnchors(gt.Tasks, upstream)
	if len(violations) > 0 {
		t.Fatalf("%d of %d anchors do not resolve at %s: %v", len(violations), checked, gt.Corpus.UpstreamSHA, violations)
	}

	snapshot := publishedCorpus(gt).Dir
	err = filepath.WalkDir(snapshot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(snapshot, p)
		if rel == "NOTICE" {
			return nil
		}
		local, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		remote, err := upstream(filepath.ToSlash(rel))
		if err != nil {
			t.Errorf("vendored %s is not in the upstream commit: %v", rel, err)
			return nil
		}
		if string(local) != string(remote) {
			t.Errorf("vendored %s differs from the upstream commit", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	suite := eval.Suite{Corpus: full, GroundTruth: gt, Thresholds: eval.DefaultThresholds(), WorkDir: t.TempDir()}
	report, err := suite.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	report.Variant = "published-full"
	reportGate(t, report)
}
