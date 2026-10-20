package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

func runCLIStderr(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err = Run(args, strings.NewReader(""), &out, &errOut)
	return out.String(), errOut.String(), err
}

// SC3: --format=caveman (alias llm) is accepted by slice, impact and locate;
// an unknown format fails naming json|md|caveman instead of falling back to
// JSON silently (this is what makes the pi-harness bridge feature-detectable).
func TestSCAgentFormatValidation(t *testing.T) {
	_, index := hermeticIndex(t)
	for _, cmd := range [][]string{
		{"slice", "service:checkout", "--index", index},
		{"impact", "service:checkout", "--index", index},
		{"locate", "checkout", "--index", index},
	} {
		args := append(append([]string{}, cmd...), "--format", "caveman")
		if _, err := runCLI(t, args...); err != nil {
			t.Fatalf("%v --format caveman: %v", cmd, err)
		}
		args = append(append([]string{}, cmd...), "--format", "llm")
		if _, err := runCLI(t, args...); err != nil {
			t.Fatalf("%v --format llm: %v", cmd, err)
		}
		args = append(append([]string{}, cmd...), "--format", "xml")
		_, _, err := runCLIStderr(t, args...)
		if err == nil {
			t.Fatalf("%v --format xml accepted (silent JSON fallback)", cmd)
		}
		if !strings.Contains(err.Error(), "json|md|caveman") {
			t.Fatalf("%v --format xml error must name the accepted set: %v", cmd, err)
		}
	}

	// Legacy commands keep the old permissive behavior (D2).
	if _, err := runCLI(t, "tour", "--index", index, "--format", "xml"); err != nil && strings.Contains(err.Error(), "json|md|caveman") {
		t.Fatalf("tour must not gain strict validation here: %v", err)
	}
}

// SC1/SC2 on the CLI surface: the emitted bytes follow the grammar and honor
// the budget on all three commands.
func TestSCAgentCavemanCLI(t *testing.T) {
	_, index := hermeticIndex(t)

	out, err := runCLI(t, "slice", "service:checkout", "--index", index, "--tier", "summary", "--format", "caveman", "--budget", "400")
	if err != nil {
		t.Fatal(err)
	}
	if est := tokens.Estimate(out); est > 400 {
		t.Fatalf("slice over budget: %d", est)
	}
	if !strings.HasPrefix(string(out), "Component: ") {
		t.Fatalf("slice grammar: %q", out)
	}

	out, err = runCLI(t, "impact", "service:checkout", "--index", index, "--format", "caveman", "--budget", "400")
	if err != nil {
		t.Fatal(err)
	}
	if est := tokens.Estimate(out); est > 400 {
		t.Fatalf("impact over budget: %d", est)
	}
	if !strings.HasPrefix(string(out), "Target: ") {
		t.Fatalf("impact grammar: %q", out)
	}

	// locate without budget: still caveman lines, no JSON braces.
	out, err = runCLI(t, "locate", "checkout", "--index", index, "--format", "caveman")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "{") {
		t.Fatalf("locate emitted JSON: %q", out)
	}
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if !strings.HasPrefix(l, "Match: ") && !strings.HasPrefix(l, "Boundary: ") && !strings.HasPrefix(l, "Budget:") {
			t.Fatalf("locate grammar line: %q", l)
		}
	}
}
