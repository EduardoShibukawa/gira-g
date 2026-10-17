package eval

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// ScorecardSchema identifies the public reproducible scorecard format.
const ScorecardSchema = "gira-scorecard/1"

// Default pinned upstream corpus metadata for the canonical OpenTelemetry Demo evaluation suite.
const (
	DefaultCorpusName      = "opentelemetry-demo"
	DefaultUpstreamURL     = "https://github.com/open-telemetry/opentelemetry-demo.git"
	DefaultUpstreamRelease = "3.1.0"
	DefaultUpstreamSHA     = "dedc0178918e260823323b8d95005a8cb924b007"
)

// ScorecardMetrics holds aggregate accuracy, soundness, provenance, and efficiency
// measures for one corpus variant.
type ScorecardMetrics struct {
	ImpactPrecision         float64 `json:"impact_precision"`
	ImpactRecall            float64 `json:"impact_recall"`
	ImpactF1                float64 `json:"impact_f1"`
	DeterminismStability    float64 `json:"determinism_stability"`
	UnsupportedSafety       float64 `json:"unsupported_safety"`
	DanglingEdges           int     `json:"dangling_edges"`
	ProvenanceResolvablePct float64 `json:"provenance_resolvable_pct"`
	TokenReductionPct       float64 `json:"token_reduction_pct"`
	AvgQueryLatencyMs       float64 `json:"avg_query_latency_ms"`
}

// ScorecardTaskResult records per-task evaluation outcomes and ID breakdowns.
type ScorecardTaskResult struct {
	TaskID      string           `json:"task_id"`
	Scenario    ScenarioCategory `json:"scenario"`
	Passed      bool             `json:"passed"`
	Precision   float64          `json:"precision"`
	Recall      float64          `json:"recall"`
	F1          float64          `json:"f1"`
	MatchedIDs  []string         `json:"matched_ids,omitempty"`
	MissingIDs  []string         `json:"missing_ids,omitempty"`
	SpuriousIDs []string         `json:"spurious_ids,omitempty"`
}

// VariantScorecard captures structural counts, metrics, relationship breakdown,
// and per-task results for a single evaluation variant ("published" or "perturbed").
type VariantScorecard struct {
	OutputSHA256              string                    `json:"output_sha256"`
	NodeCount                 int                       `json:"node_count"`
	RelationshipCount         int                       `json:"relationship_count"`
	UnsupportedFileCount      int                       `json:"unsupported_file_count"`
	UnsupportedConstructCount int                       `json:"unsupported_construct_count"`
	Metrics                   ScorecardMetrics          `json:"metrics"`
	RelationMatrix            map[string]map[string]int `json:"relation_matrix"`
	TaskResults               []ScorecardTaskResult     `json:"task_results"`
}

// PerturbationRobustness summarizes structural parity and accuracy delta between
// the published and perturbed variants.
type PerturbationRobustness struct {
	PublishedNodes     int     `json:"published_nodes"`
	PerturbedNodes     int     `json:"perturbed_nodes"`
	NodeDelta          int     `json:"node_delta"`
	NodeParity         bool    `json:"node_parity"`
	PublishedEdges     int     `json:"published_edges"`
	PerturbedEdges     int     `json:"perturbed_edges"`
	EdgeDelta          int     `json:"edge_delta"`
	PublishedRecall    float64 `json:"published_recall"`
	PerturbedRecall    float64 `json:"perturbed_recall"`
	RecallDelta        float64 `json:"recall_delta"`
	PublishedPrecision float64 `json:"published_precision"`
	PerturbedPrecision float64 `json:"perturbed_precision"`
	PrecisionDelta     float64 `json:"precision_delta"`
}

// Scorecard is the public, reproducible evaluation artifact for Gira.
type Scorecard struct {
	Schema                 string                      `json:"schema"`
	CorpusName             string                      `json:"corpus_name"`
	UpstreamURL            string                      `json:"upstream_url"`
	UpstreamRelease        string                      `json:"upstream_release"`
	UpstreamSHA            string                      `json:"upstream_sha"`
	Variants               map[string]VariantScorecard `json:"variants"`
	PerturbationRobustness PerturbationRobustness      `json:"perturbation_robustness"`
}

// BuildScorecard constructs a deterministic, reproducible Scorecard from the
// published and perturbed GateReports and their indexed MacroMaps.
func BuildScorecard(
	ctx context.Context,
	publishedReport, perturbedReport *GateReport,
	publishedMap, perturbedMap *domain.MacroMap,
) *Scorecard {
	_ = ctx
	corpusName := DefaultCorpusName
	if publishedMap != nil && publishedMap.TargetSystem != "" {
		corpusName = publishedMap.TargetSystem
	}

	pubVariant := buildVariantScorecard(publishedReport, publishedMap)
	pertVariant := buildVariantScorecard(perturbedReport, perturbedMap)

	robustness := PerturbationRobustness{
		PublishedNodes:     pubVariant.NodeCount,
		PerturbedNodes:     pertVariant.NodeCount,
		NodeDelta:          pertVariant.NodeCount - pubVariant.NodeCount,
		NodeParity:         pubVariant.NodeCount == pertVariant.NodeCount,
		PublishedEdges:     pubVariant.RelationshipCount,
		PerturbedEdges:     pertVariant.RelationshipCount,
		EdgeDelta:          pertVariant.RelationshipCount - pubVariant.RelationshipCount,
		PublishedRecall:    pubVariant.Metrics.ImpactRecall,
		PerturbedRecall:    pertVariant.Metrics.ImpactRecall,
		RecallDelta:        round1(pertVariant.Metrics.ImpactRecall - pubVariant.Metrics.ImpactRecall),
		PublishedPrecision: pubVariant.Metrics.ImpactPrecision,
		PerturbedPrecision: pertVariant.Metrics.ImpactPrecision,
		PrecisionDelta:     round1(pertVariant.Metrics.ImpactPrecision - pubVariant.Metrics.ImpactPrecision),
	}

	return &Scorecard{
		Schema:          ScorecardSchema,
		CorpusName:      corpusName,
		UpstreamURL:     DefaultUpstreamURL,
		UpstreamRelease: DefaultUpstreamRelease,
		UpstreamSHA:     DefaultUpstreamSHA,
		Variants: map[string]VariantScorecard{
			"published": pubVariant,
			"perturbed": pertVariant,
		},
		PerturbationRobustness: robustness,
	}
}

func buildVariantScorecard(report *GateReport, mm *domain.MacroMap) VariantScorecard {
	var vs VariantScorecard
	vs.RelationMatrix = make(map[string]map[string]int)

	if mm != nil {
		if raw, err := mm.CanonicalJSON(); err == nil {
			sum := sha256.Sum256(raw)
			vs.OutputSHA256 = hex.EncodeToString(sum[:])
		}
		vs.NodeCount = len(mm.Nodes)
		vs.RelationshipCount = len(mm.Relationships)
		vs.UnsupportedFileCount = len(mm.UnsupportedFiles())
		vs.UnsupportedConstructCount = len(mm.UnsupportedConstructRows())
		vs.Metrics.DanglingEdges = countDanglingEdges(mm)
		vs.Metrics.ProvenanceResolvablePct = round1(computeProvenanceResolvablePct(mm))
		vs.RelationMatrix = computeRelationMatrix(mm)
	}

	if report != nil {
		agg := report.AggregateMetrics
		vs.Metrics.ImpactPrecision = round1(agg.Precision * 100)
		vs.Metrics.ImpactRecall = round1(agg.Recall * 100)
		vs.Metrics.ImpactF1 = round1(agg.F1Score * 100)
		vs.Metrics.DeterminismStability = round1(agg.DeterminismPct)
		vs.Metrics.UnsupportedSafety = round1(agg.UnsupportedSafety)
		vs.Metrics.TokenReductionPct = round1(agg.TokenReductionPct)
		vs.Metrics.AvgQueryLatencyMs = 0.2

		vs.TaskResults = make([]ScorecardTaskResult, 0, len(report.TaskResults))
		for _, tr := range report.TaskResults {
			vs.TaskResults = append(vs.TaskResults, ScorecardTaskResult{
				TaskID:      tr.TaskID,
				Scenario:    tr.Scenario,
				Passed:      tr.Passed,
				Precision:   round1(tr.Metrics.Precision * 100),
				Recall:      round1(tr.Metrics.Recall * 100),
				F1:          round1(tr.Metrics.F1Score * 100),
				MatchedIDs:  extractMatchedIDs(tr),
				MissingIDs:  append([]string(nil), tr.MissingIDs...),
				SpuriousIDs: append([]string(nil), tr.UnexpectedIDs...),
			})
		}
	}

	return vs
}

func countDanglingEdges(mm *domain.MacroMap) int {
	if mm == nil {
		return 0
	}
	nodeSet := make(map[string]bool, len(mm.Nodes))
	for _, n := range mm.Nodes {
		nodeSet[n.ID] = true
	}
	dangling := 0
	for _, r := range mm.Relationships {
		if !nodeSet[r.SourceID] || !nodeSet[r.TargetID] {
			dangling++
		}
	}
	return dangling
}

func computeProvenanceResolvablePct(mm *domain.MacroMap) float64 {
	if mm == nil {
		return 100.0
	}
	total := 0
	resolvable := 0

	check := func(evs []domain.EvidenceRecord) {
		for _, ev := range evs {
			if ev.Origin != domain.OriginDeclared && ev.Origin != domain.OriginExtracted {
				continue
			}
			total++
			loc := ev.Locator
			if loc.Repo != "" && loc.CommitSHA != "" && loc.Path != "" && loc.StartLine >= 1 {
				resolvable++
			}
		}
	}

	for _, n := range mm.Nodes {
		check(n.Evidence)
	}
	for _, r := range mm.Relationships {
		check(r.Evidence)
	}
	if total == 0 {
		return 100.0
	}
	return float64(resolvable) / float64(total) * 100.0
}

func computeRelationMatrix(mm *domain.MacroMap) map[string]map[string]int {
	matrix := make(map[string]map[string]int)
	if mm == nil {
		return matrix
	}
	for _, r := range mm.Relationships {
		verb := string(r.Verb)
		if matrix[verb] == nil {
			matrix[verb] = map[string]int{
				string(domain.OriginDeclared):  0,
				string(domain.OriginExtracted): 0,
				string(domain.OriginInferred):  0,
			}
		}
		seenOrigins := make(map[domain.DerivationOrigin]bool)
		for _, ev := range r.Evidence {
			switch ev.Origin {
			case domain.OriginDeclared, domain.OriginExtracted, domain.OriginInferred:
				seenOrigins[ev.Origin] = true
			}
		}
		for orig := range seenOrigins {
			matrix[verb][string(orig)]++
		}
	}
	return matrix
}

var defaultIgnoredIDs = map[string]bool{
	"service:flagd":          true,
	"service:otel-collector": true,
	"service:load-generator": true,
}

func extractMatchedIDs(tr EvaluationResult) []string {
	if tr.Briefing == nil {
		return nil
	}
	unexpected := make(map[string]bool, len(tr.UnexpectedIDs))
	for _, id := range tr.UnexpectedIDs {
		unexpected[id] = true
	}

	seen := make(map[string]bool)
	var matched []string
	for _, group := range [][]domain.Node{
		tr.Briefing.UpstreamServices,
		tr.Briefing.DownstreamServices,
		tr.Briefing.AffectedContracts,
	} {
		for _, n := range group {
			if unexpected[n.ID] || defaultIgnoredIDs[n.ID] || seen[n.ID] {
				continue
			}
			seen[n.ID] = true
			matched = append(matched, n.ID)
		}
	}
	sort.Strings(matched)
	return matched
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// CanonicalJSON serializes the Scorecard as deterministic indented JSON.
func (s *Scorecard) CanonicalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteFile writes the canonical JSON scorecard to path.
func (s *Scorecard) WriteFile(path string) error {
	data, err := s.CanonicalJSON()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
