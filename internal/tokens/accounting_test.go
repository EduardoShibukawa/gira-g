package tokens_test

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// SC4 (F4): honest token accounting.
// Test-first: a len/4 estimate passing as measured must fail.
func TestSCTokenHonesty(t *testing.T) {
	// net = gross − expansions, may be negative.
	if got := tokens.NewReport(100, 30, true); got.Net != 70 {
		t.Fatalf("net must be 70, got %+v", got)
	}
	if got := tokens.NewReport(10, 30, true); got.Net != -20 {
		t.Fatalf("net may be negative, got %+v", got)
	}
	if err := tokens.NewReport(100, 30, true).Validate(); err != nil {
		t.Fatalf("valid report must pass: %v", err)
	}
	bad := tokens.TokenReport{Gross: 100, Expansions: 30, Net: 999, Measured: true}
	if err := bad.Validate(); err == nil {
		t.Fatalf("wrong net must fail validation")
	}

	// Estimated-only claims are quarantined and never enter headline numbers.
	claims := []tokens.Claim{
		{Tokens: 100, Label: tokens.LabelMeasured},
		{Tokens: 9999, Label: tokens.LabelEstimated},
	}
	headline, quarantined := tokens.Quarantine(claims)
	if headline != 100 {
		t.Fatalf("headline must exclude estimated claims, got %d", headline)
	}
	if len(quarantined) != 1 || quarantined[0].Tokens != 9999 {
		t.Fatalf("estimated claim must be quarantined, got %+v", quarantined)
	}

	// No len/4: an EstimateLen-based gross marked measured must fail.
	raw := []byte("abcdefgh") // len 8 → EstimateLen 2, measured bytes differ
	estGross := tokens.EstimateLen(len(raw))
	cheat := tokens.NewReport(estGross, 0, true)
	if err := tokens.ValidateMeasured(cheat, raw); err == nil {
		t.Fatalf("len/4 estimate passing as measured must fail (gross=%d)", estGross)
	}
	// Honest measured report passes.
	honest := tokens.Measure(raw)
	if err := tokens.ValidateMeasured(honest, raw); err != nil {
		t.Fatalf("honest measured report must pass: %v", err)
	}
	if !honest.Measured {
		t.Fatalf("Measure must be labeled measured")
	}

	// Escape-tax lift runs before measuring: escaped input lifts to fewer bytes.
	escaped := []byte(`a\nb\tc\"d\\e`)
	lifted := tokens.LiftTextPart(string(escaped))
	if len(lifted) >= len(escaped) {
		t.Fatalf("lift must reduce escape tax, got %q → %q", escaped, lifted)
	}
	if got := tokens.Measure(escaped); got.Gross != len(lifted) {
		t.Fatalf("Measure must count lifted bytes (%d), got %+v", len(lifted), got)
	}

	// Single-source tool defs: every counting path uses the same definitions.
	a := tokens.CountTokensForTools([]string{tokens.ToolDefinitions[0]})
	b := tokens.CountTokensForTools([]string{tokens.ToolDefinitions[0]})
	if a != b || a <= 0 {
		t.Fatalf("single-source counting must be deterministic and positive, got %d/%d", a, b)
	}
	if len(tokens.ToolDefinitions) == 0 {
		t.Fatalf("tool definitions must be non-empty single source")
	}
}
