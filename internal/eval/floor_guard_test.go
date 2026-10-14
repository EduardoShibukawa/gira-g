package eval_test

import (
	"context"
	"os"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// Sealed hermetic floors (E3 Q4): the numbers E1/E2 sealed at. The floor guard
// is separate from the ADR 0007 thresholds and opt-in (E3_FLOOR_GUARD=1), so the
// default suite never hard-codes a headline. It is a floor of floors: recall and
// precision may improve, never regress below the sealed values.
const (
	floorPublishedRecall    = 82.4
	floorPublishedPrecision = 100.0
	floorPerturbedRecall    = 81.2
	floorPerturbedPrecision = 100.0
)

// TestFloorGuardHermetic asserts the sealed hermetic R/P floors on both tours.
// Opt-in: set E3_FLOOR_GUARD=1 (the CI floor-guard job does).
func TestFloorGuardHermetic(t *testing.T) {
	if os.Getenv("E3_FLOOR_GUARD") != "1" {
		t.Skip("opt-in floor guard: set E3_FLOOR_GUARD=1")
	}
	ctx := context.Background()

	gt := loadPublished(t)
	published := eval.Suite{
		Corpus:      publishedCorpus(gt),
		GroundTruth: gt,
		Thresholds:  eval.DefaultThresholds(),
		WorkDir:     t.TempDir(),
	}
	pubReport, err := published.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertFloor(t, "published",
		pubReport.AggregateMetrics.Recall*100, floorPublishedRecall,
		pubReport.AggregateMetrics.Precision*100, floorPublishedPrecision)

	pertReport, err := perturbedSuite(t).Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertFloor(t, "perturbed",
		pertReport.AggregateMetrics.Recall*100, floorPerturbedRecall,
		pertReport.AggregateMetrics.Precision*100, floorPerturbedPrecision)
}

func assertFloor(t *testing.T, variant string, recall, floorRecall, precision, floorPrecision float64) {
	t.Helper()
	const eps = 0.05
	if recall+eps < floorRecall {
		t.Errorf("%s recall %.1f%% regressed below the sealed floor %.1f%%", variant, recall, floorRecall)
	}
	if precision+eps < floorPrecision {
		t.Errorf("%s precision %.1f%% regressed below the sealed floor %.1f%%", variant, precision, floorPrecision)
	}
	t.Logf("%s floor: recall %.1f%% (>= %.1f%%), precision %.1f%% (>= %.1f%%)",
		variant, recall, floorRecall, precision, floorPrecision)
}
