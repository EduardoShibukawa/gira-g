package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC5 (F5, Q5=A + Q6=A), CLI side: `gate-lint --strict --fail-on-ci` fails
// closed under CI=true while the default CI path stays warn-only, and
// `gira review --gc [--keep-resolved <n>]` keeps all open findings while
// compacting resolved ones.
func TestSCE5Deferrals(t *testing.T) {
	t.Run("gate-lint-fail-on-ci", func(t *testing.T) {
		t.Setenv("CI", "true")
		t.Setenv("GIRA_CI", "")
		dir := t.TempDir()
		missing := filepath.Join(dir, "nope.json")

		var out, errOut bytes.Buffer
		if err := Run([]string{"gate-lint", "--strict", "--state", missing, "--digest", "abc"},
			nil, &out, &errOut); err != nil {
			t.Fatalf("default gate-lint --strict in CI must warn-only, got: %v", err)
		}

		out.Reset()
		errOut.Reset()
		if err := Run([]string{"gate-lint", "--strict", "--fail-on-ci", "--state", missing, "--digest", "abc"},
			nil, &out, &errOut); err == nil {
			t.Fatal("gate-lint --strict --fail-on-ci in CI must fail closed")
		}
	})

	t.Run("review-gc", func(t *testing.T) {
		mk := func(id, status string) eval.Finding {
			return eval.Finding{
				ID: id, Taxonomy: eval.TaxonomyInferred,
				SeverityCoordinator: "major", SeverityReviewer: "minor",
				Summary: "s " + id, Status: status,
			}
		}
		seed := eval.Register{Findings: []eval.Finding{
			mk("F-1", "open"), mk("F-2", "fixed"), mk("F-3", "open"),
			mk("F-4", "closed"), mk("F-5", "verified"), mk("F-6", "open"),
		}}
		ids := func(path string) []string {
			loaded, err := eval.LoadRegister(path)
			if err != nil {
				t.Fatalf("reload %s: %v", path, err)
			}
			var out []string
			for _, f := range loaded.Findings {
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

		// Explicit keep: all open + last resolved entry.
		dir := t.TempDir()
		regPath := filepath.Join(dir, "findings.md")
		if err := seed.Save(regPath); err != nil {
			t.Fatal(err)
		}
		var out, errOut bytes.Buffer
		if err := Run([]string{"review", "--findings-register", regPath, "--gc", "--keep-resolved", "1"},
			nil, &out, &errOut); err != nil {
			t.Fatalf("review --gc --keep-resolved 1 failed: %v", err)
		}
		if got := ids(regPath); !eq(got, "F-1", "F-3", "F-5", "F-6") {
			t.Fatalf("review --gc --keep-resolved 1 = %v, want [F-1 F-3 F-5 F-6]", got)
		}

		// Default keep (0): all open, no resolved.
		if err := seed.Save(regPath); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err := Run([]string{"review", "--findings-register", regPath, "--gc"},
			nil, &out, &errOut); err != nil {
			t.Fatalf("review --gc failed: %v", err)
		}
		if got := ids(regPath); !eq(got, "F-1", "F-3", "F-6") {
			t.Fatalf("review --gc = %v, want [F-1 F-3 F-6]", got)
		}

		// Negative keep: treated as 0 (open findings survive).
		if err := seed.Save(regPath); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err := Run([]string{"review", "--findings-register", regPath, "--gc", "--keep-resolved", "-1"},
			nil, &out, &errOut); err != nil {
			t.Fatalf("review --gc --keep-resolved -1 must be handled cleanly, got: %v", err)
		}
		if got := ids(regPath); !eq(got, "F-1", "F-3", "F-6") {
			t.Fatalf("review --gc --keep-resolved -1 = %v, want [F-1 F-3 F-6]", got)
		}

		// Without --gc the register is listed, never compacted.
		if err := seed.Save(regPath); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err := Run([]string{"review", "--findings-register", regPath},
			nil, &out, &errOut); err != nil {
			t.Fatalf("review list failed: %v", err)
		}
		if got := ids(regPath); !eq(got, "F-1", "F-2", "F-3", "F-4", "F-5", "F-6") {
			t.Fatalf("review without --gc must not compact: %v", got)
		}
	})
}
