package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SC1: the plain (non-TTY / NO_COLOR) human output is byte-identical across
// NO_COLOR and non-TTY, carries no ANSI escapes, and matches a golden.
func TestSCHumanPlainFallback(t *testing.T) {
	_, index := hermeticIndex(t)
	args := []string{"impact", "service:checkout", "--index", index, "--format", "md"}

	// NO_COLOR arm (buffer stdout is already non-TTY).
	t.Setenv("NO_COLOR", "1")
	noColor, err := runCLI(t, args...)
	if err != nil {
		t.Fatal(err)
	}

	// Default arm: NO_COLOR unset; the buffer is still not a terminal.
	os.Unsetenv("NO_COLOR")
	plain, err := runCLI(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(plain, []byte("\x1b[")) {
		t.Fatalf("plain output carries ANSI escapes:\n%q", plain)
	}
	if !bytes.Equal(plain, noColor) {
		t.Fatalf("NO_COLOR changed the plain output\n%q\n%q", plain, noColor)
	}

	golden := filepath.Join("testdata", "human", "impact-checkout.md")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, plain, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update-golden)", err)
	}
	if !bytes.Equal(plain, want) {
		t.Fatalf("plain human output drifted from %s", golden)
	}
}

// SC5: the human Markdown carries the Unsupported Boundaries block when the
// slice has unsupported entries (service:checkout, deep tier).
func TestSCBoundariesSection(t *testing.T) {
	_, index := hermeticIndex(t)
	out, err := runCLI(t, "slice", "service:checkout", "--tier", "deep", "--index", index, "--format", "md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Unsupported Boundaries") {
		t.Fatalf("human slice is missing the Unsupported Boundaries block:\n%s", out)
	}
}
