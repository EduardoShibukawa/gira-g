package eval_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func TestMetricsCalculationAndThresholds(t *testing.T) {
	briefing := &slices.ImpactBriefing{
		TargetID:   "service:checkoutservice",
		TargetKind: domain.NodeService,
		Depth:      1,
		DownstreamServices: []domain.Node{
			{ID: "service:frontend"},
		},
		UpstreamServices: []domain.Node{
			{ID: "service:paymentservice"},
			{ID: "service:emailservice"},
			{ID: "infra:kafka"},
		},
		AffectedContracts: []domain.Node{
			{ID: "contract:demo.proto:CheckoutService"},
			{ID: "contract:demo.proto:PaymentService"},
			{ID: "contract:demo.proto:EmailService"},
			{ID: "contract:kafka:orders"},
		},
	}

	gt := eval.GroundTruthImpact{
		TargetID:                   "service:checkoutservice",
		ExpectedUpstreamServices:   []string{"service:paymentservice", "service:emailservice", "infra:kafka"},
		ExpectedDownstreamServices: []string{"service:frontend"},
		ExpectedAffectedContracts: []string{"contract:demo.proto:CheckoutService", "contract:demo.proto:PaymentService", "contract:demo.proto:EmailService", "contract:kafka:orders"},
		BaselineGrepTokens:         9500,
	}

	mm := domain.NewMacroMap("test")
	metrics := eval.CalculateMetrics(briefing, gt, mm)

	if metrics.Recall != 1.0 {
		t.Fatalf("expected recall 1.0, got %f", metrics.Recall)
	}
	if metrics.Precision != 1.0 {
		t.Fatalf("expected precision 1.0, got %f", metrics.Precision)
	}
	if metrics.TokenReductionPct < 60.0 {
		t.Fatalf("expected token reduction >= 60%%, got %f%%", metrics.TokenReductionPct)
	}
	if metrics.UnsupportedSafety != 100.0 {
		t.Fatalf("expected unsupported safety 100%%, got %f%%", metrics.UnsupportedSafety)
	}
}

func TestCanonical12BenchmarkTasksPublished(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "fixtures", "evaluation", "reference-corpus", "ground-truth.published.json")
	fixture, err := eval.LoadCorpusFixture(fixturePath)
	if err != nil {
		t.Fatalf("failed loading published fixture: %v", err)
	}

	tasks := eval.BuildCanonicalTasks(false)
	if len(tasks) != 12 {
		t.Fatalf("expected 12 canonical tasks, got %d", len(tasks))
	}

	evaluator := eval.NewEvaluator(&fixture.MacroMap, eval.DefaultThresholds())
	report := evaluator.EvaluateSuite(context.Background(), "published", tasks)

	if report.PassedTasks != 12 {
		for _, tr := range report.TaskResults {
			if !tr.Passed {
				t.Errorf("task %s failed: %+v (details: %s)", tr.TaskID, tr.Metrics, tr.Details)
			}
		}
		t.Fatalf("expected all 12 tasks to pass in published variant, passed: %d/12", report.PassedTasks)
	}

	if !report.GoDecision {
		t.Fatalf("expected Go decision on published variant, got No-Go: %+v", report.GateBreakdowns)
	}
}

func TestCanonical12BenchmarkTasksPerturbedGoNoGoGate(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "fixtures", "evaluation", "reference-corpus", "ground-truth.perturbed.json")
	fixture, err := eval.LoadCorpusFixture(fixturePath)
	if err != nil {
		t.Fatalf("failed loading perturbed fixture: %v", err)
	}

	tasks := eval.BuildCanonicalTasks(true)
	if len(tasks) != 12 {
		t.Fatalf("expected 12 canonical tasks, got %d", len(tasks))
	}

	thresholds := eval.DefaultThresholds()
	evaluator := eval.NewEvaluator(&fixture.MacroMap, thresholds)
	report := evaluator.EvaluateSuite(context.Background(), "perturbed", tasks)

	// Validate ADR 0007 Go/No-Go Acceptance Criteria strictly
	if report.AggregateMetrics.Recall*100.0 < thresholds.MinRecallPct {
		t.Fatalf("Impact Set Recall violated: %f%% < %f%%", report.AggregateMetrics.Recall*100.0, thresholds.MinRecallPct)
	}
	if report.AggregateMetrics.Precision*100.0 < thresholds.MinPrecisionPct {
		t.Fatalf("Impact Set Precision violated: %f%% < %f%%", report.AggregateMetrics.Precision*100.0, thresholds.MinPrecisionPct)
	}
	if report.AggregateMetrics.TokenReductionPct < thresholds.MinTokenReductionPct {
		t.Fatalf("Token Reduction violated: %f%% < %f%%", report.AggregateMetrics.TokenReductionPct, thresholds.MinTokenReductionPct)
	}
	if report.AggregateMetrics.UnsupportedSafety < thresholds.MinUnsupportedSafetyPct {
		t.Fatalf("Unsupported Safety violated: %f%% < %f%%", report.AggregateMetrics.UnsupportedSafety, thresholds.MinUnsupportedSafetyPct)
	}
	if report.AggregateMetrics.DeterminismPct < thresholds.MinDeterminismPct {
		t.Fatalf("Determinism violated: %f%% < %f%%", report.AggregateMetrics.DeterminismPct, thresholds.MinDeterminismPct)
	}

	if !report.GoDecision {
		t.Fatalf("Expected strictly Go decision for Perturbed Variant, got No-Go: %+v", report.GateBreakdowns)
	}

	for gate, passed := range report.GateBreakdowns {
		if !passed {
			t.Errorf("Gate %s failed", gate)
		}
	}
}

func TestDeterminismAndSafetyEnforcement(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "fixtures", "evaluation", "reference-corpus", "ground-truth.perturbed.json")
	fixture, err := eval.LoadCorpusFixture(fixturePath)
	if err != nil {
		t.Fatalf("failed loading perturbed fixture: %v", err)
	}

	evaluator := eval.NewEvaluator(&fixture.MacroMap, eval.DefaultThresholds())
	tasks := eval.BuildCanonicalTasks(true)

	// Test determinism over multiple repeated runs
	res1 := evaluator.EvaluateTask(context.Background(), tasks[0])
	res2 := evaluator.EvaluateTask(context.Background(), tasks[0])

	j1, _ := res1.Briefing.CanonicalJSON()
	j2, _ := res2.Briefing.CanonicalJSON()
	if string(j1) != string(j2) {
		t.Fatalf("non-deterministic output detected across runs: %s vs %s", j1, j2)
	}
}
