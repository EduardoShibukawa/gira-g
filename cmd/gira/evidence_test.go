package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// F1/F2: eval stdout alone must re-derive the closing digest (autoresearch:
// the metric comes from the run log, never from assertion).
func TestCLI_EvalStdoutRederives(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"eval", "--check", "demo", "--expect", "ok", "--taxonomy", "measured"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected body + fingerprint + closing lines, got %q", stdout.String())
	}
	closing := lines[len(lines)-1]
	var fp string
	for _, l := range lines {
		if strings.HasPrefix(l, "fingerprint: ") {
			fp = strings.TrimPrefix(l, "fingerprint: ")
		}
	}
	if fp == "" {
		t.Fatalf("eval must print the frozen fingerprint, got %q", stdout.String())
	}
	cwd, _ := os.Getwd()
	if !eval.VerifyClosingLine(closing, "demo", "ok", cwd, 0, fp, "measured") {
		t.Fatalf("stdout must re-derive the closing line (fp=%q)", fp)
	}

	// The spec'd explicit flag must exist (default on).
	stdout.Reset()
	if err := Run([]string{"eval", "--artifact-verified", "--check", "x"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("--artifact-verified must be accepted: %v", err)
	}
	if !strings.Contains(stdout.String(), "ARTIFACT-VERIFIED sha256:") {
		t.Fatalf("explicit --artifact-verified must emit the closing line")
	}
}

// F2: --control-total runs a genuine seeded sample + blind round.
func TestCLI_EvalControlHonest(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"eval", "--control-total", "--seed", "42"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("eval --control-total failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "control:{") {
		t.Fatalf("expected control line, got %q", stdout.String())
	}
}

// F3/F7: fix-check reads the real path against --expect-content.
func TestCLI_ReviewFixCheckReadsPath(t *testing.T) {
	dir := t.TempDir()
	repro := filepath.Join(dir, "repro.txt")
	if err := os.WriteFile(repro, []byte("bug"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg := eval.Register{Findings: []eval.Finding{{
		ID: "F-1", Taxonomy: eval.TaxonomyMeasured,
		SeverityCoordinator: "major", SeverityReviewer: "major",
		Summary: "s", ReproPath: repro, Status: "open",
	}}}
	regPath := filepath.Join(dir, "findings.md")
	if err := reg.Save(regPath); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := Run([]string{"review", "--findings-register", regPath, "--fix-check", "--finding", "F-1", "--expect-content", "fixed"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("fix-check must fail while the path still shows the bug")
	}
	if err := os.WriteFile(repro, []byte("fixed"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := Run([]string{"review", "--findings-register", regPath, "--fix-check", "--finding", "F-1", "--expect-content", "fixed"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("fix-check must pass after the real path is fixed: %v", err)
	}
}
