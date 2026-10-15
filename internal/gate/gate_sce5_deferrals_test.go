package gate_test

import (
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/gate"
)

// SC5 (F5, Q5=A): --fail-on-ci enforces fail-closed --strict under CI for
// callers with an explicit ledger, while the default CI path stays warn-only
// (TestSCGateLintStrictness preserved).
func TestSCE5Deferrals(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.json")

	// Default strict CI remains warn-only.
	res := gate.LintFile(missing, "abc", gate.LintOptions{Strict: true, CI: true})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("default strict CI must warn-only, got %+v", res)
	}

	// Explicit ledger + FailOnCI fails closed under CI (missing ledger).
	res = gate.LintFile(missing, "abc", gate.LintOptions{Strict: true, CI: true, FailOnCI: true})
	if res.Pass || res.Error == "" {
		t.Fatalf("strict FailOnCI in CI must fail closed, got %+v", res)
	}

	// Explicit ledger + FailOnCI fails closed under CI (stale digest).
	sealed := gate.State{Marker: gate.MarkerSuccess, Digest: "abc123", Status: gate.StatusAllMet}
	p := filepath.Join(dir, "gate.json")
	if err := gate.WriteState(p, sealed); err != nil {
		t.Fatal(err)
	}
	res = gate.LintFile(p, "STALE", gate.LintOptions{Strict: true, CI: true, FailOnCI: true})
	if res.Pass || res.Error == "" {
		t.Fatalf("strict FailOnCI drift in CI must fail closed, got %+v", res)
	}

	// Same drift without the flag stays warn-only in CI.
	res = gate.LintFile(p, "STALE", gate.LintOptions{Strict: true, CI: true})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("strict CI without FailOnCI must warn-only, got %+v", res)
	}

	// Non-strict never gates: FailOnCI does not promote a reporting lint.
	res = gate.LintFile(p, "STALE", gate.LintOptions{Strict: false, CI: true, FailOnCI: true})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("non-strict FailOnCI must still warn-and-pass, got %+v", res)
	}
}
