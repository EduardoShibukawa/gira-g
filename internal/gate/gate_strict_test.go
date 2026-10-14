package gate_test

import (
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/gate"
)

// SC5 follow-up (review F10): CI is warn-only on every path including a
// missing ledger, and --strict must actually change the verdict.
func TestSCGateLintStrictness(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.json")

	// Missing ledger: warn-only in CI, fail-closed locally.
	res := gate.LintFile(missing, "abc", gate.LintOptions{Strict: true, CI: true})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("missing ledger in CI must warn-only, got %+v", res)
	}
	res = gate.LintFile(missing, "abc", gate.LintOptions{Strict: true, CI: false})
	if res.Pass {
		t.Fatalf("missing ledger locally must fail closed, got %+v", res)
	}

	// Drift: non-strict warns and passes; strict local fails.
	sealed := gate.State{Marker: gate.MarkerSuccess, Digest: "abc123", Status: gate.StatusAllMet}
	p := filepath.Join(dir, "gate.json")
	if err := gate.WriteState(p, sealed); err != nil {
		t.Fatal(err)
	}
	res = gate.LintFile(p, "STALE", gate.LintOptions{Strict: false, CI: false})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("non-strict drift must warn-and-pass, got %+v", res)
	}
	res = gate.LintFile(p, "STALE", gate.LintOptions{Strict: true, CI: false})
	if res.Pass || res.Error == "" {
		t.Fatalf("strict local drift must fail closed, got %+v", res)
	}
}
