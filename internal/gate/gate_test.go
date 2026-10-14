package gate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/gate"
)

// SC5 (F5): gate-lint strict + status/reverify split + ABANDON handoff.
// Test-first: stale digest must fail closed; ABANDON must never yield ALL MET.
func TestSCGateLint(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "gate.json")

	// Sealed run passes strict lint locally.
	sealed := gate.State{Marker: gate.MarkerSuccess, Digest: "abc123", Status: gate.StatusAllMet}
	if err := gate.WriteState(statePath, sealed); err != nil {
		t.Fatal(err)
	}
	res := gate.LintFile(statePath, "abc123", gate.LintOptions{Strict: true, CI: false})
	if !res.Pass || res.Status != gate.StatusAllMet {
		t.Fatalf("sealed run must pass strict lint, got %+v", res)
	}

	// Stale digest / drift fails closed locally.
	res = gate.LintFile(statePath, "STALE", gate.LintOptions{Strict: true, CI: false})
	if res.Pass || res.Error == "" {
		t.Fatalf("stale digest must fail closed locally, got %+v", res)
	}

	// CI mode is warn-only until E3: same drift warns instead of failing.
	res = gate.LintFile(statePath, "STALE", gate.LintOptions{Strict: true, CI: true})
	if !res.Pass || len(res.Warnings) == 0 {
		t.Fatalf("CI strict must be warn-only, got %+v", res)
	}

	// Success-only marker: wrong marker fails even with the right digest.
	bad := gate.State{Marker: "DONE", Digest: "abc123", Status: gate.StatusAllMet}
	if err := gate.WriteState(statePath, bad); err != nil {
		t.Fatal(err)
	}
	res = gate.LintFile(statePath, "abc123", gate.LintOptions{Strict: true, CI: false})
	if res.Pass {
		t.Fatalf("non-success marker must fail lint, got %+v", res)
	}

	// ABANDON is a qualified handoff that never reports ALL MET.
	if gate.StatusForAbandon() == gate.StatusAllMet {
		t.Fatalf("ABANDON must never yield ALL MET")
	}
	ab := gate.State{Marker: gate.MarkerSuccess, Digest: "abc123", Status: gate.StatusAbandon, Reason: "blocked: need owner call"}
	if err := gate.WriteState(statePath, ab); err != nil {
		t.Fatal(err)
	}
	res = gate.LintFile(statePath, "abc123", gate.LintOptions{Strict: true, CI: false})
	if !res.Pass || res.Status == gate.StatusAllMet {
		t.Fatalf("ABANDON lint must pass as handoff but never ALL MET, got %+v", res)
	}

	// --status is non-executing: it reads without running and without writing.
	before, _ := os.Stat(statePath)
	st, err := gate.Status(statePath)
	if err != nil || st.Status != gate.StatusAbandon {
		t.Fatalf("status must read without executing, got %+v err=%v", st, err)
	}
	after, _ := os.Stat(statePath)
	if before.ModTime() != after.ModTime() {
		t.Fatalf("--status must not modify the state file")
	}

	// --reverify re-runs: drift detected on reverify fails.
	if err := gate.Reverify(statePath, "abc123"); err != nil {
		t.Fatalf("reverify on fresh digest must pass: %v", err)
	}
	if err := gate.Reverify(statePath, "drifted"); err == nil {
		t.Fatalf("reverify on drifted digest must fail")
	}
}
