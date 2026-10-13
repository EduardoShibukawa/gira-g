package eval_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
)

var (
	repoRoot       = filepath.Join("..", "..")
	groundTruthV2  = filepath.Join(repoRoot, "fixtures", "evaluation", "reference-corpus", "ground-truth.v2.json")
	perturbRecipe  = filepath.Join(repoRoot, "fixtures", "evaluation", "reference-corpus", "perturbation.json")
	publishedTasks = 12
)

func loadPublished(t *testing.T) *eval.GroundTruth {
	t.Helper()
	gt, err := eval.LoadGroundTruth(groundTruthV2)
	if err != nil {
		t.Fatal(err)
	}
	if len(gt.Tasks) != publishedTasks {
		t.Fatalf("expected %d canonical tasks, got %d", publishedTasks, len(gt.Tasks))
	}
	return gt
}

func publishedCorpus(gt *eval.GroundTruth) eval.Corpus {
	return eval.Corpus{
		Dir:          filepath.Join(repoRoot, filepath.FromSlash(gt.Corpus.Snapshot)),
		RepoID:       gt.Corpus.RepoID,
		TargetSystem: gt.Corpus.Name,
	}
}

// perturbedSuite builds the Perturbed Variant corpus and ground truth from the
// published ones, through the declarative recipe.
func perturbedSuite(t *testing.T) eval.Suite {
	t.Helper()
	gt := loadPublished(t)
	recipe, err := eval.LoadRecipe(perturbRecipe)
	if err != nil {
		t.Fatal(err)
	}
	pgt, err := gt.Perturbed(recipe)
	if err != nil {
		t.Fatal(err)
	}
	c := publishedCorpus(gt)
	dir := filepath.Join(t.TempDir(), "perturbed")
	if err := corpus.Perturb(c.Dir, dir, recipe.Perturbation); err != nil {
		t.Fatal(err)
	}
	c.Dir = dir
	return eval.Suite{Corpus: c, GroundTruth: pgt, Thresholds: eval.DefaultThresholds(), WorkDir: t.TempDir()}
}

func reportGate(t *testing.T, report *eval.GateReport) {
	t.Helper()
	t.Log(report.Summary())
	for _, r := range report.TaskResults {
		t.Logf("  %s pass=%t P=%.2f R=%.2f tokens=%d/%d missing=%v unexpected=%v %s",
			r.TaskID, r.Passed, r.Metrics.Precision, r.Metrics.Recall,
			r.Metrics.HarnessTokens, r.Metrics.BaselineTokens, r.MissingIDs, r.UnexpectedIDs, r.Details)
	}
	dir := os.Getenv("GIRA_EVAL_REPORT_DIR")
	if dir == "" {
		dir = t.TempDir()
	}
	if err := report.WriteFile(filepath.Join(dir, "gate-report."+report.Variant+".json")); err != nil {
		t.Fatal(err)
	}
	if err := report.Enforce(); err != nil {
		t.Fatal(err)
	}
}

// SC1, SC4, SC5: the published suite indexes the real corpus end to end.
func TestSuitePublishedEndToEnd(t *testing.T) {
	gt := loadPublished(t)
	suite := eval.Suite{Corpus: publishedCorpus(gt), GroundTruth: gt, Thresholds: eval.DefaultThresholds(), WorkDir: t.TempDir()}
	report, err := suite.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.AggregateMetrics.DeterminismPct != 100 {
		t.Fatalf("index is not deterministic across cold runs")
	}
	if report.AggregateMetrics.UnsupportedSafety != 100 {
		t.Fatalf("unsupported safety %.1f%%: %v", report.AggregateMetrics.UnsupportedSafety, report.UnsupportedViolations)
	}
	reportGate(t, report)
}

// SC4, SC5 on the Perturbed Variant, the definitive ADR 0007 gate.
func TestSuitePerturbedEndToEnd(t *testing.T) {
	suite := perturbedSuite(t)
	report, err := suite.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.AggregateMetrics.DeterminismPct != 100 {
		t.Fatalf("perturbed index is not deterministic across cold runs")
	}
	reportGate(t, report)
}

// SC1: the score depends on the extractors. Removing the Go AST extractor (the only
// source of consumes/provides edges to gRPC contracts) must drop contract recall.
func TestMutationRemovingExtractorDropsRecall(t *testing.T) {
	gt := loadPublished(t)
	var contractTasks []eval.BenchmarkTask
	for _, task := range gt.Tasks {
		if strings.HasPrefix(task.TargetID, "contract:") {
			contractTasks = append(contractTasks, task)
		}
	}
	gt.Tasks = contractTasks

	run := func(pipe *extractors.Pipeline) float64 {
		suite := eval.Suite{Corpus: publishedCorpus(gt), GroundTruth: gt, Thresholds: eval.DefaultThresholds(), Pipeline: pipe, WorkDir: t.TempDir()}
		report, err := suite.Run(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return report.AggregateMetrics.Recall
	}

	full := run(nil)
	var kept []extractors.Extractor
	for _, ext := range extractors.DefaultPipeline().Extractors() {
		if ext.Name() != extractors.NewGoASTExtractor().Name() {
			kept = append(kept, ext)
		}
	}
	mutated := run(extractors.NewPipeline(kept...))
	t.Logf("contract-task recall: full pipeline %.2f, without Go AST %.2f", full, mutated)
	if mutated >= full {
		t.Fatalf("removing the Go AST extractor did not lower recall (%.2f >= %.2f): the evaluator is not exercising extraction", mutated, full)
	}
}

// SC2: the baseline is measured on the corpus. More matching files cost more tokens;
// the harness side is untouched.
func TestBaselineIsMeasuredOnCorpus(t *testing.T) {
	gt := loadPublished(t)
	src := publishedCorpus(gt).Dir

	before, err := eval.NewGrepBaseline(src)
	if err != nil {
		t.Fatal(err)
	}

	grown := filepath.Join(t.TempDir(), "grown")
	if _, err := corpus.Materialize(context.Background(), src, grown); err != nil {
		t.Fatal(err)
	}
	extra := strings.Repeat("checkout calls payment; see the checkout flow.\n", 200)
	if err := os.WriteFile(filepath.Join(grown, "NOTES.txt"), []byte(extra), 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := eval.NewGrepBaseline(grown)
	if err != nil {
		t.Fatal(err)
	}

	b0, b1 := before.Tokens("service:checkout"), after.Tokens("service:checkout")
	if b1 <= b0 {
		t.Fatalf("baseline did not grow with matching corpus content: %d -> %d", b0, b1)
	}

	briefingTokens := 321
	task := gt.Tasks[0].GroundTruth
	m0, _, _ := eval.CalculateMetrics(&emptyBriefing, task, nil, briefingTokens, b0)
	m1, _, _ := eval.CalculateMetrics(&emptyBriefing, task, nil, briefingTokens, b1)
	if m0.HarnessTokens != m1.HarnessTokens || m1.TokenReductionPct <= m0.TokenReductionPct {
		t.Fatalf("harness tokens must be unchanged and reduction must grow: %+v vs %+v", m0, m1)
	}
}

// SC3: one hallucinated extraction in an unsupported C++ source fails the gate.
func TestUnsupportedSafetyDetectsHallucination(t *testing.T) {
	gt := loadPublished(t)
	idx, err := publishedCorpus(gt).Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"src/currency", "src/payment"}

	clean, violations := eval.UnsupportedSafety(idx.MacroMap, paths)
	if clean != 100 {
		t.Fatalf("clean index should score 100, got %.1f: %v", clean, violations)
	}

	idx.MacroMap.Relationships = append(idx.MacroMap.Relationships, domain.Relationship{
		ID: "rel:currency-consumes-ghost", SourceID: "service:currency", TargetID: "contract:Ghost", Verb: domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{{
			Locator: domain.Locator{Repo: gt.Corpus.RepoID, Path: "src/currency/src/server.cpp", StartLine: 10, EndLine: 12},
			Origin:  domain.OriginExtracted,
		}},
	})
	dirty, violations := eval.UnsupportedSafety(idx.MacroMap, paths)
	if dirty >= 100 || len(violations) != 1 {
		t.Fatalf("injected hallucination not detected: score %.1f, violations %v", dirty, violations)
	}

	report := eval.GateReport{
		Variant:          "published",
		Thresholds:       eval.DefaultThresholds(),
		AggregateMetrics: eval.ScoringMetrics{UnsupportedSafety: dirty, DeterminismPct: 100},
		GateBreakdowns:   map[string]bool{"determinism_stability_100": true, "unsupported_safety_100": dirty >= 100},
	}
	if err := report.Enforce(); err == nil {
		t.Fatalf("gate must fail on a hallucinated extraction")
	}
}

// SC5: thresholds fail the build only on reviewed ground truth.
func TestGateSemantics(t *testing.T) {
	report := eval.GateReport{
		Variant: "perturbed",
		GateBreakdowns: map[string]bool{
			"impact_recall_gte_90":      false,
			"impact_precision_gte_80":   true,
			"token_reduction_gte_60":    true,
			"unsupported_safety_100":    true,
			"determinism_stability_100": true,
		},
	}
	if err := report.Enforce(); err != nil {
		t.Fatalf("unreviewed ground truth must report, not fail: %v", err)
	}
	report.Reviewed = true
	if err := report.Enforce(); err == nil {
		t.Fatalf("reviewed ground truth must fail on a recall miss")
	}
}

func TestLoadGroundTruthRejectsEmbeddedMacroMap(t *testing.T) {
	p := filepath.Join(t.TempDir(), "gt.json")
	if err := os.WriteFile(p, []byte(`{"schema":"gira-eval-ground-truth/2","macro_map":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := eval.LoadGroundTruth(p); err == nil {
		t.Fatal("a ground truth carrying a macro_map must be rejected as circular")
	}
}

func TestPerturbedGroundTruthAppliesRecipe(t *testing.T) {
	suite := perturbedSuite(t)
	first := suite.GroundTruth.Tasks[0]
	if first.TargetID != "service:ordering" {
		t.Fatalf("target not renamed: %s", first.TargetID)
	}
	for _, id := range first.GroundTruth.ExpectedUpstreamServices {
		if id == "service:exchange" || id == "service:currency" {
			t.Fatalf("altered edge still expected: %s", id)
		}
	}
	found := false
	for _, id := range first.GroundTruth.ExpectedAffectedContracts {
		found = found || id == "contract:kafka:order-events"
	}
	if !found {
		t.Fatalf("topic id rename not applied: %v", first.GroundTruth.ExpectedAffectedContracts)
	}
}

var emptyBriefing = emptyImpact()
