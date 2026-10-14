package eval_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC3 (F3): L3 taxonomy + register + independent fix-check.
// Test-first: unlabeled findings and log-reading fix-checks must fail.
func TestSCL3TaxonomyRegister(t *testing.T) {
	// Gallon S1 verbatim labels map to Gira tiers.
	for tax, wantTier := range map[eval.Taxonomy]string{
		eval.TaxonomyMeasured: "extracted",
		eval.TaxonomyChecked:  "declared",
		eval.TaxonomyInferred: "inferred",
		eval.TaxonomyAssumed:  "inferred",
	} {
		if got := tax.GiraTier(); got != wantTier {
			t.Errorf("taxonomy %q maps to %q, want %q", tax, got, wantTier)
		}
	}

	// Unlabeled finding fails.
	unlabeled := eval.Finding{ID: "F-1", SeverityCoordinator: "major", SeverityReviewer: "minor", Summary: "x"}
	if err := unlabeled.Validate(); err == nil {
		t.Fatalf("unlabeled finding must fail validation")
	}

	// Code-reading without a real-path repro is inferred at most.
	dir := t.TempDir()
	repro := filepath.Join(dir, "repro.txt")
	if err := os.WriteFile(repro, []byte("evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	measuredNoRepro := eval.Finding{
		ID: "F-2", Taxonomy: eval.TaxonomyMeasured,
		SeverityCoordinator: "major", SeverityReviewer: "major",
		Summary: "claims measurement without repro",
	}
	if err := measuredNoRepro.Validate(); err == nil {
		t.Fatalf("measured without real-path repro must fail (inferred at most)")
	}
	measured := eval.Finding{
		ID: "F-3", Taxonomy: eval.TaxonomyMeasured,
		SeverityCoordinator: "critical", SeverityReviewer: "major",
		Summary: "real-path repro", ReproPath: repro, Status: "open",
	}
	if err := measured.Validate(); err != nil {
		t.Fatalf("measured with repro + severities must pass: %v", err)
	}
	missingSeverity := eval.Finding{
		ID: "F-4", Taxonomy: eval.TaxonomyInferred, Summary: "no severity", ReproPath: repro,
	}
	if err := missingSeverity.Validate(); err == nil {
		t.Fatalf("finding without coordinator/reviewer severity must fail")
	}

	// Register round-trips ID-paired with status.
	regPath := filepath.Join(dir, "findings.md")
	reg := eval.Register{Findings: []eval.Finding{measured}}
	if err := reg.Save(regPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := eval.LoadRegister(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Findings) != 1 || loaded.Findings[0].ID != "F-3" || loaded.Findings[0].Status != "open" {
		t.Fatalf("register must round-trip ID+status, got %+v", loaded.Findings)
	}

	// Fix-check re-runs the real path independently: fails before fix, passes after.
	check := func(path string) (bool, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		return string(data) == "fixed", nil
	}
	if err := eval.FixCheck("F-3", regPath, check); err == nil {
		t.Fatalf("fix-check must fail while the repro still shows the bug")
	}
	if err := os.WriteFile(repro, []byte("fixed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := eval.FixCheck("F-3", regPath, check); err != nil {
		t.Fatalf("fix-check must pass after the real path is fixed: %v", err)
	}

	// A fix-check that reads the fixer's log instead of rerunning must fail:
	// nil rerun (no independent path) is rejected.
	if err := eval.FixCheck("F-3", regPath, nil); err == nil {
		t.Fatalf("fix-check without an independent rerun must fail")
	}
}
