package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// runEval is the additive F1+F2 entry point: `gira eval --artifact-verified`
// (default on; `--no-artifact-verified` is debug-only) plus
// `--control-total --sample max20-5pct --seed <u64>`. It never changes an
// existing command. The run always ends with the ARTIFACT-VERIFIED closing
// line unless explicitly disabled.
func runEval(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var noArtifact bool
	var check, expect, cwd string
	var controlTotal bool
	var sample string
	var seed uint64
	var tokenReport string
	fs.BoolVar(&noArtifact, "no-artifact-verified", false, "Debug-only: suppress the ARTIFACT-VERIFIED closing line")
	fs.StringVar(&check, "check", "gira eval", "CHECK: command label recorded in the digest")
	fs.StringVar(&expect, "expect", "ok", "EXPECT: success marker recorded in the digest")
	fs.StringVar(&cwd, "cwd", "", "CWD recorded in the digest (default: current directory)")
	fs.BoolVar(&controlTotal, "control-total", false, "Emit control:{total,matched,mismatched,sampled_n,seed} line (F2)")
	fs.StringVar(&sample, "sample", "max20-5pct", "Sampling rule (only max20-5pct supported)")
	fs.Uint64Var(&seed, "seed", 0, "Seed for the seeded sampler (F2)")
	fs.StringVar(&tokenReport, "token-report", "", "Token accounting mode: measured (F4, measurement only)")
	if err := fs.Parse(args); err != nil {
		return err
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
		"CHECK": []byte(check),
		"EXPECT": []byte(expect),
	}
	if data, err := os.ReadFile("gira.yaml"); err == nil {
		inputs["gira.yaml"] = data
	}
	fp := eval.FingerprintInputs(inputs)
	digest := eval.ArtifactDigest("CHECK: "+check, "EXPECT: "+expect, cwd, 0, fp)
	run := eval.ArtifactRun{
		Check:        "CHECK: " + check,
		Expect:       "EXPECT: " + expect,
		CWD:          cwd,
		ExitCode:     0,
		Inputs:       inputs,
		ClosingLine:  eval.FormatClosingLine(digest),
		AllowedScope: []string{"out/"},
	}
	body := fmt.Sprintf("eval ok: check=%q expect=%q cwd=%q\n", check, expect, cwd)
	if controlTotal {
		if sample != "max20-5pct" {
			return fmt.Errorf("unsupported --sample %q: only max20-5pct is supported", sample)
		}
		// Demo control denominator: the frozen input count. Real eval runs
		// feed the task count; the format and rule are what F2 pins.
		total := len(inputs)
		rep := eval.ControlReport(total, total, 0, seed)
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
	_, err := fmt.Fprint(stdout, eval.EmitWithArtifact(body, run))
	return err
}

// runVerify is the additive F1 companion: `gira verify --check ... --expect ...`.
// It recomputes and prints the closing line without running an evaluation.
func runVerify(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var check, expect, cwd, fingerprint string
	var exitCode int
	fs.StringVar(&check, "check", "", "CHECK: command label")
	fs.StringVar(&expect, "expect", "", "EXPECT: success marker")
	fs.StringVar(&cwd, "cwd", ".", "CWD recorded in the digest")
	fs.StringVar(&fingerprint, "fingerprint", "", "Frozen input fingerprint hex (empty = fingerprint of check/expect)")
	fs.IntVar(&exitCode, "exit", 0, "Exit code recorded in the digest")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fingerprint == "" {
		fingerprint = eval.FingerprintInputs(map[string][]byte{
			"CHECK": []byte(check), "EXPECT": []byte(expect),
		})
	}
	digest := eval.ArtifactDigest(check, expect, cwd, exitCode, fingerprint)
	_, err := fmt.Fprintln(stdout, eval.FormatClosingLine(digest))
	return err
}
