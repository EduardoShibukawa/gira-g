// Package eval is gira's internal, non-circular evaluator (spec
// gira-agent-skills-and-eval, part C). A run materializes a corpus as a real git
// repository, indexes it through the same path as `gira index`, emits briefings
// through the slice engine, and only then compares them with the ground truth. No
// code path here deserializes a pre-built MacroMap.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// GroundTruthSchema is the only ground-truth format the evaluator accepts.
const GroundTruthSchema = "gira-eval-ground-truth/2"

// ScenarioCategory represents benchmark categories defined in ADR 0007.
type ScenarioCategory string

// Anchor is a string that must appear in a file at the pinned commit. Layer 0 uses
// anchors to check that a cited locator really points at the evidence.
type Anchor struct {
	Path     string `json:"path"`
	Contains string `json:"contains"`
}

// GroundTruthImpact declares the expected Impact Set of one benchmark task.
type GroundTruthImpact struct {
	ExpectedUpstreamServices   []string `json:"expected_upstream_services"`
	ExpectedDownstreamServices []string `json:"expected_downstream_services"`
	ExpectedAffectedContracts  []string `json:"expected_affected_contracts"`
	ExpectedUnsupportedPaths   []string `json:"expected_unsupported_paths,omitempty"`
	ExpectedFreshnessStatus    string   `json:"expected_freshness_status,omitempty"`
	FlagsConflict              bool     `json:"flags_conflict,omitempty"`
	Anchors                    []Anchor `json:"anchors,omitempty"`
	Notes                      string   `json:"notes,omitempty"`
}

// BenchmarkTask defines one task of the 12-task suite.
type BenchmarkTask struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Scenario    ScenarioCategory  `json:"scenario"`
	TargetID    string            `json:"target_id"`
	Depth       int               `json:"depth"`
	Tier        slices.SliceTier  `json:"tier"`
	GroundTruth GroundTruthImpact `json:"ground_truth"`
}

// CorpusRef pins the corpus a ground truth was derived from.
type CorpusRef struct {
	Name        string `json:"name"`
	Release     string `json:"release"`
	UpstreamSHA string `json:"upstream_sha"`
	UpstreamURL string `json:"upstream_url,omitempty"`
	RepoID      string `json:"repo_id"`
	Snapshot    string `json:"snapshot"`
}

// GroundTruth is the v2 fixture: expected Impact Sets only, never a MacroMap.
type GroundTruth struct {
	Schema     string          `json:"schema"`
	Variant    string          `json:"variant"`
	Reviewed   bool            `json:"reviewed"`
	ReviewNote string          `json:"review_note,omitempty"`
	Corpus     CorpusRef       `json:"corpus"`
	Ignore     []string        `json:"ignore,omitempty"`
	IgnoreNote string          `json:"ignore_note,omitempty"`
	Flows      []FlowTruth     `json:"flows,omitempty"`
	Tasks      []BenchmarkTask `json:"tasks"`
}

// FlowTruth is a critical path a human reader should be walked through, in order.
// Layer 2 and SC8 measure how much of it a tour visits.
type FlowTruth struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Entry   string   `json:"entry"`
	Steps   []string `json:"steps"`
	Anchors []Anchor `json:"anchors,omitempty"`
	Notes   string   `json:"notes,omitempty"`
}

// LoadGroundTruth reads a v2 ground-truth fixture. A fixture that still carries a
// MacroMap is rejected: that is the circular v1 format.
func LoadGroundTruth(path string) (*GroundTruth, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading ground truth %s: %w", path, err)
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("parsing ground truth %s: %w", path, err)
	}
	if _, circular := probe["macro_map"]; circular {
		return nil, fmt.Errorf("ground truth %s embeds a macro_map; the evaluator must index the corpus itself", path)
	}
	var gt GroundTruth
	if err := json.Unmarshal(data, &gt); err != nil {
		return nil, fmt.Errorf("parsing ground truth %s: %w", path, err)
	}
	if gt.Schema != GroundTruthSchema {
		return nil, fmt.Errorf("ground truth %s has schema %q, want %q", path, gt.Schema, GroundTruthSchema)
	}
	return &gt, nil
}

// Recipe is the declarative Perturbed Variant: corpus changes plus the matching
// ground-truth changes, kept in one file so they cannot drift apart.
type Recipe struct {
	corpus.Perturbation
	// IndexChanges declares the architectural changes the edits make, in published
	// ids. SC9 checks an architecture diff reports exactly these plus the renames.
	IndexChanges struct {
		RemovedEdges []struct {
			Source string `json:"source"`
			Verb   string `json:"verb"`
			Target string `json:"target"`
		} `json:"removed_edges"`
	} `json:"index_changes"`
	GroundTruth struct {
		IDRenames []corpus.Rename `json:"id_renames"`
		Remove    []struct {
			Task  string `json:"task"`
			Field string `json:"field"`
			ID    string `json:"id"`
		} `json:"remove"`
	} `json:"ground_truth"`
}

// LoadRecipe reads a perturbation recipe.
func LoadRecipe(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading perturbation %s: %w", path, err)
	}
	var r Recipe
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing perturbation %s: %w", path, err)
	}
	return &r, nil
}

// Perturbed derives the Perturbed Variant's ground truth from the published one.
func (gt *GroundTruth) Perturbed(r *Recipe) (*GroundTruth, error) {
	data, err := json.Marshal(gt)
	if err != nil {
		return nil, err
	}
	var out GroundTruth
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	out.Variant = "perturbed"

	for ti := range out.Tasks {
		t := &out.Tasks[ti]
		for _, rm := range r.GroundTruth.Remove {
			if rm.Task != t.ID {
				continue
			}
			field, err := t.GroundTruth.field(rm.Field)
			if err != nil {
				return nil, err
			}
			kept := (*field)[:0]
			removed := false
			for _, id := range *field {
				if id == rm.ID {
					removed = true
					continue
				}
				kept = append(kept, id)
			}
			if !removed {
				return nil, fmt.Errorf("perturbation removes %s from %s.%s, but it is not there", rm.ID, rm.Task, rm.Field)
			}
			*field = kept
		}
	}

	ids := corpus.Perturbation{Renames: r.GroundTruth.IDRenames}
	rename := func(s string) string { return r.Apply(ids.Apply(s)) }
	renameAll := func(xs []string) {
		for i := range xs {
			xs[i] = rename(xs[i])
		}
	}
	renameAll(out.Ignore)
	for fi := range out.Flows {
		f := &out.Flows[fi]
		f.Entry = rename(f.Entry)
		renameAll(f.Steps)
		for ai := range f.Anchors {
			f.Anchors[ai].Path = rename(f.Anchors[ai].Path)
			f.Anchors[ai].Contains = rename(f.Anchors[ai].Contains)
		}
	}
	for ti := range out.Tasks {
		t := &out.Tasks[ti]
		t.TargetID = rename(t.TargetID)
		g := &t.GroundTruth
		renameAll(g.ExpectedUpstreamServices)
		renameAll(g.ExpectedDownstreamServices)
		renameAll(g.ExpectedAffectedContracts)
		renameAll(g.ExpectedUnsupportedPaths)
		for ai := range g.Anchors {
			g.Anchors[ai].Path = rename(g.Anchors[ai].Path)
			g.Anchors[ai].Contains = rename(g.Anchors[ai].Contains)
		}
	}
	return &out, nil
}

func (g *GroundTruthImpact) field(name string) (*[]string, error) {
	switch name {
	case "expected_upstream_services":
		return &g.ExpectedUpstreamServices, nil
	case "expected_downstream_services":
		return &g.ExpectedDownstreamServices, nil
	case "expected_affected_contracts":
		return &g.ExpectedAffectedContracts, nil
	}
	return nil, fmt.Errorf("unknown ground-truth field %q", name)
}

// ScoringMetrics holds precision, recall, token and safety measures.
type ScoringMetrics struct {
	TruePositives     int     `json:"true_positives"`
	FalsePositives    int     `json:"false_positives"`
	FalseNegatives    int     `json:"false_negatives"`
	Precision         float64 `json:"precision"`
	Recall            float64 `json:"recall"`
	F1Score           float64 `json:"f1_score"`
	HarnessTokens     int     `json:"harness_tokens"`
	BaselineTokens    int     `json:"baseline_tokens"`
	TokenReductionPct float64 `json:"token_reduction_pct"`
	UnsupportedSafety float64 `json:"unsupported_safety_pct"` // 100 means no hallucinated extraction
	DeterminismPct    float64 `json:"determinism_pct"`
}

// EvaluationResult is the outcome of one task.
type EvaluationResult struct {
	TaskID          string                 `json:"task_id"`
	Scenario        ScenarioCategory       `json:"scenario"`
	Passed          bool                   `json:"passed"`
	Metrics         ScoringMetrics         `json:"metrics"`
	MissingIDs      []string               `json:"missing_ids,omitempty"`
	UnexpectedIDs   []string               `json:"unexpected_ids,omitempty"`
	ConflictFlagged *bool                  `json:"conflict_flagged,omitempty"`
	Details         string                 `json:"details,omitempty"`
	Briefing        *slices.ImpactBriefing `json:"briefing,omitempty"`
}

// ThresholdConfig represents the Go/No-Go acceptance criteria from ADR 0007.
type ThresholdConfig struct {
	MinRecallPct            float64 `json:"min_recall_pct"`
	MinPrecisionPct         float64 `json:"min_precision_pct"`
	MinTokenReductionPct    float64 `json:"min_token_reduction_pct"`
	MinUnsupportedSafetyPct float64 `json:"min_unsupported_safety_pct"`
	MinDeterminismPct       float64 `json:"min_determinism_pct"`
}

// DefaultThresholds returns the strict ADR 0007 Perturbed Variant acceptance criteria.
func DefaultThresholds() ThresholdConfig {
	return ThresholdConfig{
		MinRecallPct:            90.0,
		MinPrecisionPct:         80.0,
		MinTokenReductionPct:    60.0,
		MinUnsupportedSafetyPct: 100.0,
		MinDeterminismPct:       100.0,
	}
}

// GateReport compiles a suite run and its Go/No-Go decision.
type GateReport struct {
	Variant               string             `json:"variant"`
	CorpusSHA             string             `json:"corpus_sha"`
	Reviewed              bool               `json:"ground_truth_reviewed"`
	TotalTasks            int                `json:"total_tasks"`
	PassedTasks           int                `json:"passed_tasks"`
	AggregateMetrics      ScoringMetrics     `json:"aggregate_metrics"`
	UnsupportedViolations []string           `json:"unsupported_violations,omitempty"`
	Thresholds            ThresholdConfig    `json:"thresholds"`
	GoDecision            bool               `json:"go_decision"`
	GateBreakdowns        map[string]bool    `json:"gate_breakdowns"`
	TaskResults           []EvaluationResult `json:"task_results"`
}

// Enforce applies the gate semantics decided by the owner on 2026-10-13: determinism
// and unsupported safety always hold (gira guarantees them by construction), while
// recall, precision and token reduction fail the build only on reviewed ground truth.
func (r GateReport) Enforce() error {
	var failed []string
	for _, gate := range []string{"determinism_stability_100", "unsupported_safety_100"} {
		if !r.GateBreakdowns[gate] {
			failed = append(failed, gate)
		}
	}
	if r.Reviewed {
		for _, gate := range []string{"impact_recall_gte_90", "impact_precision_gte_80", "token_reduction_gte_60"} {
			if !r.GateBreakdowns[gate] {
				failed = append(failed, gate)
			}
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%s variant: gates failed: %s", r.Variant, strings.Join(failed, ", "))
	}
	return nil
}

// Summary is a one-line, human-readable digest of the report.
func (r GateReport) Summary() string {
	m := r.AggregateMetrics
	decision := "No-Go"
	if r.GoDecision {
		decision = "Go"
	}
	return fmt.Sprintf("%s: %s (reviewed=%t) tasks %d/%d, recall %.1f%%, precision %.1f%%, token reduction %.1f%%, unsupported safety %.1f%%, determinism %.1f%%",
		r.Variant, decision, r.Reviewed, r.PassedTasks, r.TotalTasks, m.Recall*100, m.Precision*100, m.TokenReductionPct, m.UnsupportedSafety, m.DeterminismPct)
}

// WriteFile writes the report as indented JSON.
func (r GateReport) WriteFile(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Corpus is an evaluation corpus on disk plus how it is declared to gira.
type Corpus struct {
	Dir          string // source tree, also read by the grep baseline
	RepoID       string
	TargetSystem string
}

// Indexed is one real indexing run of a corpus.
type Indexed struct {
	SHA      string
	MacroMap *domain.MacroMap
}

// Index materializes the corpus as a git repository under workDir and indexes it
// with indexer.Build, exactly as `gira index` would. Each call uses its own cache.
func (c Corpus) Index(ctx context.Context, workDir string, pipe *extractors.Pipeline) (*Indexed, error) {
	repoDir := filepath.Join(workDir, "repo")
	sha, err := corpus.Materialize(ctx, c.Dir, repoDir)
	if err != nil {
		return nil, fmt.Errorf("materializing corpus: %w", err)
	}
	cm, err := gitcache.NewCacheManager(filepath.Join(workDir, "cache"))
	if err != nil {
		return nil, err
	}
	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: c.TargetSystem},
		Repositories: []manifest.RepositoryDeclaration{{ID: c.RepoID, URL: repoDir, SHA: sha, PathScope: "."}},
	}
	mm, err := indexer.Build(ctx, m, cm, indexer.Options{Pipeline: pipe})
	if err != nil {
		return nil, err
	}
	return &Indexed{SHA: sha, MacroMap: mm}, nil
}

// Suite is one variant run: a corpus, its ground truth and the thresholds.
type Suite struct {
	Corpus      Corpus
	GroundTruth *GroundTruth
	Thresholds  ThresholdConfig
	// Pipeline overrides the extractor pipeline (mutation tests). Nil means default.
	Pipeline *extractors.Pipeline
	// WorkDir holds the materialized repositories and caches.
	WorkDir string
}

// Run indexes the corpus twice from cold caches, checks the two indexes are
// byte-identical, evaluates every task on the first index and returns the report.
func (s Suite) Run(ctx context.Context) (*GateReport, error) {
	first, err := s.Corpus.Index(ctx, filepath.Join(s.WorkDir, "run-1"), s.Pipeline)
	if err != nil {
		return nil, err
	}
	second, err := s.Corpus.Index(ctx, filepath.Join(s.WorkDir, "run-2"), s.Pipeline)
	if err != nil {
		return nil, err
	}
	j1, err := first.MacroMap.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	j2, err := second.MacroMap.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	indexDeterministic := string(j1) == string(j2) && first.SHA == second.SHA

	baseline, err := NewGrepBaseline(s.Corpus.Dir)
	if err != nil {
		return nil, err
	}

	ev := NewEvaluator(first.MacroMap, s.GroundTruth.Ignore, baseline, s.Thresholds)
	report := ev.EvaluateSuite(ctx, s.GroundTruth.Variant, s.GroundTruth.Tasks)
	report.CorpusSHA = first.SHA
	report.Reviewed = s.GroundTruth.Reviewed

	safety, violations := UnsupportedSafety(first.MacroMap, expectedUnsupportedPaths(s.GroundTruth.Tasks))
	report.AggregateMetrics.UnsupportedSafety = safety
	report.UnsupportedViolations = violations
	if !indexDeterministic {
		report.AggregateMetrics.DeterminismPct = 0
	}
	report.decide()
	return &report, nil
}

func expectedUnsupportedPaths(tasks []BenchmarkTask) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tasks {
		for _, p := range t.GroundTruth.ExpectedUnsupportedPaths {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Evaluator scores briefings from an indexed MacroMap against ground truth.
type Evaluator struct {
	engine     *slices.Engine
	macroMap   *domain.MacroMap
	ignore     map[string]bool
	baseline   *GrepBaseline
	thresholds ThresholdConfig
}

// NewEvaluator creates an Evaluator. The MacroMap must come from Corpus.Index.
func NewEvaluator(m *domain.MacroMap, ignore []string, baseline *GrepBaseline, thresholds ThresholdConfig) *Evaluator {
	ig := make(map[string]bool, len(ignore))
	for _, id := range ignore {
		ig[id] = true
	}
	return &Evaluator{
		engine:     slices.NewEngine(m),
		macroMap:   m,
		ignore:     ig,
		baseline:   baseline,
		thresholds: thresholds,
	}
}

// EvaluateTask evaluates a single task. A target the index does not know is a
// task failure with every expected ID counted as a false negative.
func (ev *Evaluator) EvaluateTask(ctx context.Context, task BenchmarkTask) EvaluationResult {
	baselineTokens := 0
	if ev.baseline != nil {
		baselineTokens = ev.baseline.Tokens(task.TargetID)
	}

	briefing, err := ev.engine.GetImpactBriefing(task.TargetID, task.Depth)
	if err != nil {
		empty := &slices.ImpactBriefing{TargetID: task.TargetID}
		metrics, missing, unexpected := CalculateMetrics(empty, task.GroundTruth, ev.ignore, 0, baselineTokens)
		return EvaluationResult{
			TaskID:        task.ID,
			Scenario:      task.Scenario,
			Metrics:       metrics,
			MissingIDs:    missing,
			UnexpectedIDs: unexpected,
			Details:       fmt.Sprintf("engine failure: %v", err),
		}
	}

	out, err := briefing.CanonicalJSON()
	if err != nil {
		return EvaluationResult{TaskID: task.ID, Scenario: task.Scenario, Details: fmt.Sprintf("render failure: %v", err)}
	}
	metrics, missing, unexpected := CalculateMetrics(briefing, task.GroundTruth, ev.ignore, tokens.Estimate(out), baselineTokens)

	res := EvaluationResult{
		TaskID:        task.ID,
		Scenario:      task.Scenario,
		Metrics:       metrics,
		MissingIDs:    missing,
		UnexpectedIDs: unexpected,
		Briefing:      briefing,
	}
	if task.GroundTruth.FlagsConflict {
		// The domain model has no conflict status yet, so no briefing can flag one.
		// Reported, not gated: ADR 0007's gate does not include trap detection.
		flagged := false
		res.ConflictFlagged = &flagged
	}
	res.Passed = metrics.Recall*100 >= ev.thresholds.MinRecallPct &&
		metrics.Precision*100 >= ev.thresholds.MinPrecisionPct &&
		metrics.TokenReductionPct >= ev.thresholds.MinTokenReductionPct
	return res
}

// EvaluateSuite evaluates every task and aggregates the metrics. Unsupported
// safety and index determinism are corpus-level and are set by Suite.Run.
func (ev *Evaluator) EvaluateSuite(ctx context.Context, variant string, tasks []BenchmarkTask) GateReport {
	results := make([]EvaluationResult, 0, len(tasks))
	agg := ScoringMetrics{UnsupportedSafety: 100, DeterminismPct: 100}
	for _, task := range tasks {
		res := ev.EvaluateTask(ctx, task)
		results = append(results, res)
		agg.TruePositives += res.Metrics.TruePositives
		agg.FalsePositives += res.Metrics.FalsePositives
		agg.FalseNegatives += res.Metrics.FalseNegatives
		agg.HarnessTokens += res.Metrics.HarnessTokens
		agg.BaselineTokens += res.Metrics.BaselineTokens
	}
	agg.Precision, agg.Recall, agg.F1Score = ratios(agg.TruePositives, agg.FalsePositives, agg.FalseNegatives)
	agg.TokenReductionPct = reduction(agg.BaselineTokens, agg.HarnessTokens)

	// Re-query determinism: the same index must give byte-identical briefings.
	for _, task := range tasks {
		b1, err1 := ev.engine.GetImpactBriefing(task.TargetID, task.Depth)
		b2, err2 := ev.engine.GetImpactBriefing(task.TargetID, task.Depth)
		if (err1 == nil) != (err2 == nil) {
			agg.DeterminismPct = 0
			break
		}
		if err1 != nil {
			continue
		}
		j1, _ := b1.CanonicalJSON()
		j2, _ := b2.CanonicalJSON()
		if string(j1) != string(j2) {
			agg.DeterminismPct = 0
			break
		}
	}

	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	report := GateReport{
		Variant:          variant,
		TotalTasks:       len(tasks),
		PassedTasks:      passed,
		AggregateMetrics: agg,
		Thresholds:       ev.thresholds,
		TaskResults:      results,
	}
	report.decide()
	return report
}

func (r *GateReport) decide() {
	m := r.AggregateMetrics
	t := r.Thresholds
	r.GateBreakdowns = map[string]bool{
		"impact_recall_gte_90":      m.Recall*100 >= t.MinRecallPct,
		"impact_precision_gte_80":   m.Precision*100 >= t.MinPrecisionPct,
		"token_reduction_gte_60":    m.TokenReductionPct >= t.MinTokenReductionPct,
		"unsupported_safety_100":    m.UnsupportedSafety >= t.MinUnsupportedSafetyPct,
		"determinism_stability_100": m.DeterminismPct >= t.MinDeterminismPct,
	}
	r.GoDecision = true
	for _, ok := range r.GateBreakdowns {
		r.GoDecision = r.GoDecision && ok
	}
}

// CalculateMetrics scores a briefing against ground truth. IDs in ignore count
// neither way. harnessTokens and baselineTokens are measured by the caller.
func CalculateMetrics(briefing *slices.ImpactBriefing, gt GroundTruthImpact, ignore map[string]bool, harnessTokens, baselineTokens int) (ScoringMetrics, []string, []string) {
	actual := map[string]bool{}
	for _, group := range [][]domain.Node{briefing.UpstreamServices, briefing.DownstreamServices, briefing.AffectedContracts} {
		for _, n := range group {
			if !ignore[n.ID] {
				actual[n.ID] = true
			}
		}
	}
	expected := map[string]bool{}
	for _, group := range [][]string{gt.ExpectedUpstreamServices, gt.ExpectedDownstreamServices, gt.ExpectedAffectedContracts} {
		for _, id := range group {
			if !ignore[id] {
				expected[id] = true
			}
		}
	}

	var missing, unexpected []string
	tp := 0
	for id := range actual {
		if expected[id] {
			tp++
		} else {
			unexpected = append(unexpected, id)
		}
	}
	for id := range expected {
		if !actual[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)

	m := ScoringMetrics{
		TruePositives:     tp,
		FalsePositives:    len(unexpected),
		FalseNegatives:    len(missing),
		HarnessTokens:     harnessTokens,
		BaselineTokens:    baselineTokens,
		TokenReductionPct: reduction(baselineTokens, harnessTokens),
		UnsupportedSafety: 100,
		DeterminismPct:    100,
	}
	m.Precision, m.Recall, m.F1Score = ratios(tp, m.FalsePositives, m.FalseNegatives)
	return m, missing, unexpected
}

func ratios(tp, fp, fn int) (precision, recall, f1 float64) {
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	return precision, recall, f1
}

func reduction(baseline, harness int) float64 {
	if baseline <= 0 {
		return 0
	}
	r := float64(baseline-harness) / float64(baseline) * 100
	if r < 0 {
		return 0
	}
	return r
}
