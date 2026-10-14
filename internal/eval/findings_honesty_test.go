package eval_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC3 honesty follow-up (review F5, F6, F11): repro paths must exist,
// the register must survive hostile cell contents, loads must validate,
// and unknown taxonomies map to "unknown" (I1: silence over naming).
func TestSCL3RegisterHonesty(t *testing.T) {
	dir := t.TempDir()
	repro := filepath.Join(dir, "repro.txt")
	if err := os.WriteFile(repro, []byte("evidence"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A measured claim on a nonexistent path is code-reading: inferred at most.
	fake := eval.Finding{
		ID: "F-fake", Taxonomy: eval.TaxonomyMeasured,
		SeverityCoordinator: "major", SeverityReviewer: "major",
		Summary: "fake path", ReproPath: filepath.Join(dir, "nope-12345.txt"),
	}
	if err := fake.Validate(); err == nil {
		t.Fatalf("measured with a nonexistent repro path must fail")
	}

	// Hostile cell contents round-trip byte-identically.
	nasty := eval.Finding{
		ID: "F-pipe", Taxonomy: eval.TaxonomyInferred,
		SeverityCoordinator: "minor", SeverityReviewer: "minor",
		Summary: "a|b|c \\ backslash", ReproPath: repro, Status: "open",
	}
	regPath := filepath.Join(dir, "findings.md")
	if err := (eval.Register{Findings: []eval.Finding{nasty}}).Save(regPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := eval.LoadRegister(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Findings) != 1 || loaded.Findings[0].Summary != nasty.Summary {
		t.Fatalf("register must round-trip hostile cells, got %+v", loaded.Findings)
	}

	// Unknown taxonomy on load fails closed.
	badPath := filepath.Join(dir, "bad.md")
	bad := "# Findings register\n\n| ID | taxonomy | coordinator | reviewer | status | repro | summary |\n|---|---|---|---|---|---|---|\n| F-x | madeup | major | minor | open | " + repro + " | x |\n"
	if err := os.WriteFile(badPath, []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := eval.LoadRegister(badPath); err == nil {
		t.Fatalf("loading an unknown taxonomy must fail")
	}

	// Unknown taxonomies stay silent, never named.
	if got := eval.Taxonomy("weird").GiraTier(); got != "unknown" {
		t.Fatalf("unknown taxonomy must map to %q, got %q", "unknown", got)
	}
}

// N1/N2 residuals: data rows containing "---" or ending in "ID" must
// survive; truncated rows, empty IDs/statuses and duplicate IDs fail closed.
func TestSCL3RegisterResiduals(t *testing.T) {
	dir := t.TempDir()
	repro := filepath.Join(dir, "repro.txt")
	if err := os.WriteFile(repro, []byte("evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	mk := func(id, summary string) eval.Finding {
		return eval.Finding{
			ID: id, Taxonomy: eval.TaxonomyInferred,
			SeverityCoordinator: "minor", SeverityReviewer: "minor",
			Summary: summary, ReproPath: repro, Status: "open",
		}
	}
	regPath := filepath.Join(dir, "findings.md")
	in := eval.Register{Findings: []eval.Finding{
		mk("GRID", "retry --- failed again"),
		mk("VALID", "ends with ID"),
	}}
	if err := in.Save(regPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := eval.LoadRegister(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Findings) != 2 || loaded.Findings[0].Summary != "retry --- failed again" || loaded.Findings[0].ID != "GRID" {
		t.Fatalf("data rows with ---/ID must round-trip, got %+v", loaded.Findings)
	}

	// Truncated row fails closed.
	truncPath := filepath.Join(dir, "trunc.md")
	trunc := "# Findings register\n\n| ID | taxonomy | coordinator | reviewer | status | repro | summary |\n|---|---|---|---|---|---|---|\n| F-1 | inferred | minor | minor | open | " + repro + " | ok |\n| F-2 | inferred |\n"
	if err := os.WriteFile(truncPath, []byte(trunc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := eval.LoadRegister(truncPath); err == nil {
		t.Fatalf("truncated row must fail closed")
	}

	// Duplicate ID fails closed.
	dupPath := filepath.Join(dir, "dup.md")
	dup := "# Findings register\n\n| ID | taxonomy | coordinator | reviewer | status | repro | summary |\n|---|---|---|---|---|---|---|\n| F-d | inferred | minor | minor | open | " + repro + " | a |\n| F-d | inferred | minor | minor | open | " + repro + " | b |\n"
	if err := os.WriteFile(dupPath, []byte(dup), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := eval.LoadRegister(dupPath); err == nil {
		t.Fatalf("duplicate ID must fail closed")
	}
}
