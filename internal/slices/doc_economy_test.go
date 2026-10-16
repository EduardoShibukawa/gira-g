package slices_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// scDocRawBytes concatenates the fixture sources: the "cat of the same
// files" baseline a raw onboarding read would pay for.
func scDocRawBytes() []byte {
	var raw []byte
	for _, f := range scDocFiles() {
		raw = append(raw, f.Content...)
	}
	return raw
}

// TestSCDocBudget (SC4/F4): every doc emits an honest budget (bytes/4,
// estimated matching the emitted bytes), stays byte-identical on repeat,
// and the economy subtest reports the doc-vs-raw ratio without chasing it.
func TestSCDocBudget(t *testing.T) {
	eng := scDocEngine(t)
	doc, err := eng.GetDoc("service:acme")
	if err != nil {
		t.Fatalf("GetDoc failed: %v", err)
	}

	t.Run("honest_envelope", func(t *testing.T) {
		full, err := doc.Emit(slices.FormatJSON, 0)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(full), `"budget"`) {
			t.Fatalf("unbudgeted doc must carry no budget envelope")
		}
		limit := tokens.Estimate(full) * 2
		out, err := doc.Emit(slices.FormatJSON, limit)
		if err != nil {
			t.Fatal(err)
		}
		var parsed struct {
			ServiceID string         `json:"service_id"`
			Budget    *slices.Budget `json:"budget"`
		}
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.ServiceID != "service:acme" {
			t.Fatalf("budgeted doc dropped the focal service")
		}
		if parsed.Budget == nil {
			t.Fatalf("budgeted doc carries no budget envelope")
		}
		b := parsed.Budget
		if b.Limit != limit {
			t.Errorf("budget limit = %d, want %d", b.Limit, limit)
		}
		if b.Estimated != tokens.Estimate(out) {
			t.Errorf("budget estimated = %d, want tokens.Estimate(output) = %d", b.Estimated, tokens.Estimate(out))
		}
		if b.Estimator != "bytes/4" {
			t.Errorf("budget estimator = %q, want bytes/4", b.Estimator)
		}
		if b.Measured {
			t.Errorf("budget measured = true, want false (an estimate is never measured)")
		}
		if b.Truncated {
			t.Errorf("generous budget must not truncate")
		}
	})

	t.Run("tight_budget_truncates_honestly", func(t *testing.T) {
		full, err := doc.Emit(slices.FormatJSON, 0)
		if err != nil {
			t.Fatal(err)
		}
		tight := tokens.Estimate(full) - 100
		out, err := doc.Emit(slices.FormatJSON, tight)
		if err != nil {
			t.Fatalf("tight budget failed: %v", err)
		}
		var parsed struct {
			ServiceID string         `json:"service_id"`
			Budget    *slices.Budget `json:"budget"`
		}
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.ServiceID != "service:acme" {
			t.Fatalf("truncated doc dropped the focal service")
		}
		if parsed.Budget == nil || !parsed.Budget.Truncated {
			t.Fatalf("tight budget must report truncation: %+v", parsed.Budget)
		}
		if parsed.Budget.Dropped == (slices.Dropped{}) {
			t.Fatalf("truncated doc must say what it dropped: %+v", parsed.Budget)
		}
		if parsed.Budget.EscalateHint == "" {
			t.Fatalf("truncated doc carries no escalate hint")
		}
		again, err := doc.Emit(slices.FormatJSON, tight)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != string(again) {
			t.Fatalf("truncated output is not byte-identical on repeat")
		}
	})

	t.Run("economy_reports_ratio", func(t *testing.T) {
		full, err := doc.Emit(slices.FormatJSON, 0)
		if err != nil {
			t.Fatal(err)
		}
		raw := scDocRawBytes()
		docTokens := tokens.Estimate(full)
		rawTokens := tokens.Estimate(raw)
		if docTokens <= 0 || rawTokens <= 0 {
			t.Fatalf("non-positive token counts: doc=%d raw=%d", docTokens, rawTokens)
		}
		t.Logf("doc-economy: doc=%d bytes (%d tokens) vs raw cat=%d bytes (%d tokens): raw/doc=%.2f",
			len(full), docTokens, len(raw), rawTokens, float64(rawTokens)/float64(docTokens))
	})
}

// BenchmarkDocEconomy reports the F4 economy: the doc token cost against
// the raw cat/grep baseline over the same fixture files. Root-module only:
// no prior-art tool and no coupling to the quarantined comparative
// benchmark module (ADR 0012) — inputs are rebuilt in memory, raw outputs
// are never committed.
func BenchmarkDocEconomy(b *testing.B) {
	mm, err := extractors.DefaultPipeline().ProcessFiles(context.Background(), "acme-system", scDocFiles())
	if err != nil {
		b.Fatal(err)
	}
	doc, err := slices.NewEngine(mm).GetDoc("service:acme")
	if err != nil {
		b.Fatal(err)
	}
	full, err := doc.Emit(slices.FormatJSON, 0)
	if err != nil {
		b.Fatal(err)
	}
	raw := scDocRawBytes()
	rawTokens, docTokens := tokens.Estimate(raw), tokens.Estimate(full)
	ratio := float64(len(raw)) / float64(len(full))
	b.Logf("doc-economy: doc=%d bytes (%d tokens) vs raw cat=%d bytes (%d tokens): raw/doc=%.2f",
		len(full), docTokens, len(raw), rawTokens, float64(rawTokens)/float64(docTokens))
	b.ReportMetric(float64(rawTokens), "raw-tokens")
	b.ReportMetric(float64(docTokens), "doc-tokens")
	b.ReportMetric(ratio, "raw/doc")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := doc.Emit(slices.FormatJSON, 0); err != nil {
			b.Fatal(err)
		}
	}
}
