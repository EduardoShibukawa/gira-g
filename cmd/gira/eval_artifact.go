package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// runEval is the additive F1+F2+F4 entry point. One canonical digest
// preimage serves every entry point (autoresearch: a single fixed evaluation
// harness): the stdout lines (fingerprint, control, tokens) plus the closing
// line re-derive without hidden inputs.
func runEval(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var noArtifact bool
	var artifactVerified bool
	var check, expect, cwd string
	var controlTotal bool
	var sample string
	var seed uint64
	var tokenReport string
	var taxonomy string
	fs.BoolVar(&noArtifact, "no-artifact-verified", false, "Debug-only: suppress the ARTIFACT-VERIFIED closing line")
	fs.BoolVar(&artifactVerified, "artifact-verified", true, "Emit the ARTIFACT-VERIFIED closing line (default on)")
	fs.StringVar(&check, "check", "gira eval", "CHECK: command label recorded in the digest")
	fs.StringVar(&expect, "expect", "ok", "EXPECT: success marker recorded in the digest")
	fs.StringVar(&cwd, "cwd", "", "CWD recorded in the digest (default: current directory)")
	fs.StringVar(&taxonomy, "taxonomy", "inferred", "L3 claim taxonomy recorded in the digest (measured|checked|inferred|assumed)")
	fs.BoolVar(&controlTotal, "control-total", false, "Emit control:{total,matched,mismatched,sampled_n,seed} line (F2)")
	fs.StringVar(&sample, "sample", "max20-5pct", "Sampling rule (only max20-5pct supported)")
	fs.Uint64Var(&seed, "seed", 0, "Seed for the seeded sampler (F2)")
	fs.StringVar(&tokenReport, "token-report", "", "Token accounting mode: measured (F4, measurement only)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !eval.ValidTaxonomy(taxonomy) {
		return fmt.Errorf("unknown --taxonomy %q: want measured|checked|inferred|assumed", taxonomy)
	}
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			cwd = "."
		}
	}
	// Frozen inputs: the manifest + working-tree pin when present. Minimal and
	// additive; the fingerprint still covers command/expect/cwd+exit.
	inputs := map[string][]byte{
		"CHECK":  []byte(check),
		"EXPECT": []byte(expect),
	}
	if data, err := os.ReadFile("gira.yaml"); err == nil {
		inputs["gira.yaml"] = data
	}
	fp := eval.FingerprintInputs(inputs)
	digest := eval.ArtifactDigest(check, expect, cwd, 0, fp, taxonomy)
	run := eval.ArtifactRun{
		Check:        check,
		Expect:       expect,
		CWD:          cwd,
		ExitCode:     0,
		Taxonomy:     taxonomy,
		Inputs:       inputs,
		ClosingLine:  eval.FormatClosingLine(digest),
		AllowedScope: []string{"out/"},
	}
	body := fmt.Sprintf("eval ok: check=%q expect=%q cwd=%q taxonomy=%q\n", check, expect, cwd, taxonomy)
	body += fmt.Sprintf("fingerprint: %s\n", fp)
	if controlTotal {
		if sample != "max20-5pct" {
			return fmt.Errorf("unsupported --sample %q: only max20-5pct is supported", sample)
		}
		// Genuine seeded sample + blind round over the frozen inputs: the
		// control line reports a measured round, never an assertion. Both
		// derivations read the same frozen fingerprint, so agreement is
		// expected; any tampering surfaces as a diff and fails the run.
		total := len(inputs)
		sampled := eval.SeededSample(total, seed)
		vals := map[int]string{}
		for _, idx := range sampled {
			vals[idx] = fmt.Sprintf("%s:%d", fp, idx)
		}
		res, err := eval.BlindReDerive(sampled, vals, vals, fp, fp)
		if err != nil {
			return fmt.Errorf("control blind round failed: %w", err)
		}
		rep := eval.ControlReport(total, res.Matched, res.Mismatched, seed)
		if err := rep.Validate(); err != nil {
			return fmt.Errorf("control line invalid: %w", err)
		}
		body += fmt.Sprintf("control:{total:%d,matched:%d,mismatched:%d,sampled_n:%d,seed:%d}\n",
			rep.Total, rep.Matched, rep.Mismatched, rep.SampledN, rep.Seed)
	}
	if tokenReport != "" {
		if tokenReport != "measured" {
			return fmt.Errorf("unsupported --token-report %q: only measured is supported", tokenReport)
		}
		// Measurement only (F4): honest measured report over the body so far.
		rep := tokens.Measure([]byte(body))
		body += fmt.Sprintf("tokens:{gross:%d,expansions:%d,net:%d,measured:%t}\n",
			rep.Gross, rep.Expansions, rep.Net, rep.Measured)
	}
	if noArtifact {
		_, err := fmt.Fprint(stdout, body)
		return err
	}
	if !artifactVerified {
		_, err := fmt.Fprint(stdout, body)
		return err
	}
	_, err := fmt.Fprint(stdout, eval.EmitWithArtifact(body, run))
	return err
}

// runVerify is the additive F1 companion: `gira verify --check ... --expect ...`.
// It recomputes and prints the closing line without running an evaluation.
// Same canonical preimage as runEval (raw check/expect, explicit taxonomy).
func runVerify(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var check, expect, cwd, fingerprint string
	var exitCode int
	var taxonomy string
	fs.StringVar(&check, "check", "", "CHECK: command label")
	fs.StringVar(&expect, "expect", "", "EXPECT: success marker")
	fs.StringVar(&cwd, "cwd", ".", "CWD recorded in the digest")
	fs.StringVar(&fingerprint, "fingerprint", "", "Frozen input fingerprint hex (empty = fingerprint of check/expect)")
	fs.StringVar(&taxonomy, "taxonomy", "inferred", "L3 claim taxonomy recorded in the digest")
	fs.IntVar(&exitCode, "exit", 0, "Exit code recorded in the digest")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !eval.ValidTaxonomy(taxonomy) {
		return fmt.Errorf("unknown --taxonomy %q: want measured|checked|inferred|assumed", taxonomy)
	}
	if fingerprint == "" {
		// Same frozen-input rule as runEval (N3): CHECK/EXPECT plus
		// gira.yaml when present, so the default re-derives the closing
		// line instead of silently disagreeing with it.
		inputs := map[string][]byte{
			"CHECK": []byte(check), "EXPECT": []byte(expect),
		}
		if data, err := os.ReadFile("gira.yaml"); err == nil {
			inputs["gira.yaml"] = data
		}
		fingerprint = eval.FingerprintInputs(inputs)
	}
	digest := eval.ArtifactDigest(check, expect, cwd, exitCode, fingerprint, taxonomy)
	_, err := fmt.Fprintln(stdout, eval.FormatClosingLine(digest))
	return err
}
