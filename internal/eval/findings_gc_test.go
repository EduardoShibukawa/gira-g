package eval_test

import (
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC5 (F5, Q6=A): Register.GC keeps every open finding unconditionally in
// order and compacts resolved (fixed/closed/verified) findings to the last
// keepResolved entries. Negative keepResolved is treated as 0.
func TestSCE5Deferrals(t *testing.T) {
	mk := func(id, status string) eval.Finding {
		return eval.Finding{
			ID: id, Taxonomy: eval.TaxonomyInferred,
			SeverityCoordinator: "major", SeverityReviewer: "minor",
			Summary: "s " + id, Status: status,
		}
	}
	reg := eval.Register{Findings: []eval.Finding{
		mk("F-1", "open"), mk("F-2", "fixed"), mk("F-3", "open"),
		mk("F-4", "closed"), mk("F-5", "verified"), mk("F-6", "open"),
	}}
	ids := func(r eval.Register) []string {
		var out []string
		for _, f := range r.Findings {
			out = append(out, f.ID)
		}
		return out
	}
	eq := func(got []string, want ...string) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}

	// keepResolved=1: all open in order + last resolved entry.
	if got := ids(reg.GC(1)); !eq(got, "F-1", "F-3", "F-5", "F-6") {
		t.Fatalf("GC(1) = %v, want [F-1 F-3 F-5 F-6]", got)
	}

	// keepResolved=0 (default): all open, no resolved.
	if got := ids(reg.GC(0)); !eq(got, "F-1", "F-3", "F-6") {
		t.Fatalf("GC(0) = %v, want [F-1 F-3 F-6]", got)
	}

	// keepResolved larger than the resolved count keeps everything in order.
	if got := ids(reg.GC(99)); !eq(got, "F-1", "F-2", "F-3", "F-4", "F-5", "F-6") {
		t.Fatalf("GC(99) = %v, want all in order", got)
	}

	// Negative keepResolved is treated as 0, never drops open.
	if got := ids(reg.GC(-3)); !eq(got, "F-1", "F-3", "F-6") {
		t.Fatalf("GC(-3) = %v, want [F-1 F-3 F-6]", got)
	}

	// GC never mutates the receiver.
	if len(reg.Findings) != 6 {
		t.Fatalf("GC mutated the receiver: %d findings", len(reg.Findings))
	}

	// The compacted register validates and round-trips through Save/Load.
	dir := t.TempDir()
	compacted := reg.GC(1)
	regPath := filepath.Join(dir, "findings.md")
	if err := compacted.Save(regPath); err != nil {
		t.Fatalf("compacted register must validate and save: %v", err)
	}
	loaded, err := eval.LoadRegister(regPath)
	if err != nil {
		t.Fatalf("compacted register must reload: %v", err)
	}
	if got := ids(loaded); !eq(got, "F-1", "F-3", "F-5", "F-6") {
		t.Fatalf("reloaded GC(1) = %v, want [F-1 F-3 F-5 F-6]", got)
	}
}
