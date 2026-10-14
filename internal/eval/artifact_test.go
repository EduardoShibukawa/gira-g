package eval_test

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC1 (F1): L1 artifact-verified closing line.
// Test-first: tampered digest must fail, out-of-scope writes must fail closed,
// dispatcher 5-point self-check must catch substitution.
func TestSCL1ArtifactVerified(t *testing.T) {
	inputs := map[string][]byte{
		"fixtures/ground-truth.v2.json": []byte(`{"schema":"gira-eval-ground-truth/2"}`),
		"corpus/snapshot":               []byte("sha-pinned-corpus"),
	}

	fp1 := eval.FingerprintInputs(inputs)
	fp2 := eval.FingerprintInputs(inputs)
	if fp1 == "" || fp1 != fp2 {
		t.Fatalf("fingerprint must be deterministic and non-empty, got %q vs %q", fp1, fp2)
	}

	// Tampered input changes the fingerprint.
	tampered := map[string][]byte{
		"fixtures/ground-truth.v2.json": []byte(`{"schema":"tampered"}`),
		"corpus/snapshot":               []byte("sha-pinned-corpus"),
	}
	if got := eval.FingerprintInputs(tampered); got == fp1 {
		t.Fatalf("tampered input must change fingerprint (got %q)", got)
	}

	// 1 MiB cap: content past the cap does not change the fingerprint.
	bigA := map[string][]byte{"big": append([]byte("prefix-"), make([]byte, 2<<20)...)}
	bigB := map[string][]byte{"big": append([]byte("prefix-"), make([]byte, 2<<20)...)}
	bigB["big"][len(bigB["big"])-1] ^= 0xFF // flip last byte, beyond cap
	if eval.FingerprintInputs(bigA) != eval.FingerprintInputs(bigB) {
		t.Fatalf("bytes beyond the 1 MiB frozen cap must not affect the fingerprint")
	}

	// Digest covers every field: flipping any one flips the digest.
	base := eval.ArtifactDigest("CHECK: go test ./...", "EXPECT: ok", "/repo", 0, fp1)
	for name, mutate := range map[string]func() string{
		"check":       func() string { return eval.ArtifactDigest("CHECK: other", "EXPECT: ok", "/repo", 0, fp1) },
		"expect":      func() string { return eval.ArtifactDigest("CHECK: go test ./...", "EXPECT: fail", "/repo", 0, fp1) },
		"cwd":         func() string { return eval.ArtifactDigest("CHECK: go test ./...", "EXPECT: ok", "/other", 0, fp1) },
		"exit":        func() string { return eval.ArtifactDigest("CHECK: go test ./...", "EXPECT: ok", "/repo", 1, fp1) },
		"fingerprint": func() string { return eval.ArtifactDigest("CHECK: go test ./...", "EXPECT: ok", "/repo", 0, eval.FingerprintInputs(tampered)) },
	} {
		if mutate() == base {
			t.Fatalf("digest must cover %s", name)
		}
	}

	// Closing line is the final line and round-trips; tampered digest fails.
	line := eval.FormatClosingLine(base)
	if !strings.HasPrefix(line, "ARTIFACT-VERIFIED sha256:") {
		t.Fatalf("closing line must start with ARTIFACT-VERIFIED sha256:, got %q", line)
	}
	digest, ok := eval.ParseClosingLine(strings.TrimSpace(line) + "\n")
	if !ok || digest != base {
		t.Fatalf("ParseClosingLine must round-trip, got %q,%t", digest, ok)
	}
	if eval.VerifyClosingLine("ARTIFACT-VERIFIED sha256:deadbeef", "CHECK: go test ./...", "EXPECT: ok", "/repo", 0, fp1) {
		t.Fatalf("tampered digest must fail verification")
	}
	if !eval.VerifyClosingLine(line, "CHECK: go test ./...", "EXPECT: ok", "/repo", 0, fp1) {
		t.Fatalf("valid closing line must verify")
	}

	// Writes outside the frozen scope fail closed.
	allowed := []string{"out/run-1/"}
	if err := eval.EnforceWriteScope([]string{"out/run-1/report.json"}, allowed); err != nil {
		t.Fatalf("in-scope write must pass: %v", err)
	}
	if err := eval.EnforceWriteScope([]string{"out/run-1/report.json", "/tmp/evil"}, allowed); err == nil {
		t.Fatalf("out-of-scope write must fail closed")
	}

	// Dispatcher 5-point self-check passes on a good run, fails on substitution.
	run := eval.ArtifactRun{
		Check:        "CHECK: go test ./...",
		Expect:       "EXPECT: ok",
		CWD:          "/repo",
		ExitCode:     0,
		Inputs:       inputs,
		ClosingLine:  line,
		AllowedScope: allowed,
	}
	if failures := eval.DispatcherSelfCheck(run, inputs); len(failures) != 0 {
		t.Fatalf("self-check on good run must pass, got %v", failures)
	}
	substituted := map[string][]byte{
		"fixtures/ground-truth.v2.json": []byte(`{"schema":"substituted"}`),
		"corpus/snapshot":               []byte("sha-pinned-corpus"),
	}
	if failures := eval.DispatcherSelfCheck(run, substituted); len(failures) == 0 {
		t.Fatalf("self-check must catch read-for-substitute")
	}

	// Emitted output ends with the closing line as its final line.
	out := eval.EmitWithArtifact("some output\n", run)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if lines[len(lines)-1] != line {
		t.Fatalf("closing line must be the final line, got %q", lines[len(lines)-1])
	}
}
