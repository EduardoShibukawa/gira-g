package eval_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

func TestPublicScorecardReproducible(t *testing.T) {
	ctx := context.Background()

	gt := loadPublished(t)
	pubCorpus := publishedCorpus(gt)
	pubSuite := eval.Suite{
		Corpus:      pubCorpus,
		GroundTruth: gt,
		Thresholds:  eval.DefaultThresholds(),
		WorkDir:     t.TempDir(),
	}
	pubReport, err := pubSuite.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pubIndexed, err := pubCorpus.Index(ctx, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}

	pertSuite := perturbedSuite(t)
	pertReport, err := pertSuite.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pertIndexed, err := pertSuite.Corpus.Index(ctx, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}

	sc := eval.BuildScorecard(ctx, pubReport, pertReport, pubIndexed.MacroMap, pertIndexed.MacroMap)
	if sc.Schema != eval.ScorecardSchema {
		t.Fatalf("schema = %q, want %q", sc.Schema, eval.ScorecardSchema)
	}

	for _, name := range []string{"published", "perturbed"} {
		v, ok := sc.Variants[name]
		if !ok {
			t.Fatalf("missing variant %q in scorecard", name)
		}
		if v.OutputSHA256 == "" || v.NodeCount == 0 || v.RelationshipCount == 0 {
			t.Fatalf("%s variant has empty structural counts: %+v", name, v)
		}
		if v.Metrics.DanglingEdges != 0 {
			t.Fatalf("%s DanglingEdges = %d, want 0", name, v.Metrics.DanglingEdges)
		}
		if v.Metrics.ProvenanceResolvablePct != 100.0 {
			t.Fatalf("%s ProvenanceResolvablePct = %.1f, want 100.0", name, v.Metrics.ProvenanceResolvablePct)
		}
		if v.Metrics.DeterminismStability != 100.0 {
			t.Fatalf("%s DeterminismStability = %.1f, want 100.0", name, v.Metrics.DeterminismStability)
		}
		if v.Metrics.UnsupportedSafety != 100.0 {
			t.Fatalf("%s UnsupportedSafety = %.1f, want 100.0", name, v.Metrics.UnsupportedSafety)
		}
		if v.Metrics.ImpactPrecision != 100.0 {
			t.Fatalf("%s ImpactPrecision = %.1f, want 100.0", name, v.Metrics.ImpactPrecision)
		}
		if len(v.RelationMatrix) == 0 {
			t.Fatalf("%s RelationMatrix is empty", name)
		}
		if len(v.TaskResults) != publishedTasks {
			t.Fatalf("%s TaskResults len = %d, want %d", name, len(v.TaskResults), publishedTasks)
		}
	}

	if sc.Variants["published"].Metrics.ImpactRecall < 82.4 {
		t.Fatalf("published ImpactRecall = %.1f, want >= 82.4", sc.Variants["published"].Metrics.ImpactRecall)
	}
	if sc.Variants["perturbed"].Metrics.ImpactRecall < 81.2 {
		t.Fatalf("perturbed ImpactRecall = %.1f, want >= 81.2", sc.Variants["perturbed"].Metrics.ImpactRecall)
	}
	if !sc.PerturbationRobustness.NodeParity {
		t.Fatalf("expected node parity between published (%d) and perturbed (%d)",
			sc.PerturbationRobustness.PublishedNodes, sc.PerturbationRobustness.PerturbedNodes)
	}

	gotBytes, err := sc.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}

	scorecardPath := filepath.Join(repoRoot, "docs", "scorecard.json")
	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := sc.WriteFile(scorecardPath); err != nil {
			t.Fatal(err)
		}
		return
	}

	wantBytes, err := os.ReadFile(scorecardPath)
	if err != nil {
		t.Fatalf("reading %s (run with UPDATE_SNAPSHOTS=1 to generate): %v", scorecardPath, err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("docs/scorecard.json is out of date with BuildScorecard output (run UPDATE_SNAPSHOTS=1 go test ./internal/eval -run TestPublicScorecardReproducible)")
	}
}
