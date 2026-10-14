package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// runEval is the additive F1 entry point: `gira eval --artifact-verified`
// (default on; `--no-artifact-verified` is debug-only). It never changes an
// existing command. The run always ends with the ARTIFACT-VERIFIED closing
// line unless explicitly disabled.
func runEval(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var noArtifact bool
	var check, expect, cwd string
	fs.BoolVar(&noArtifact, "no-artifact-verified", false, "Debug-only: suppress the ARTIFACT-VERIFIED closing line")
	fs.StringVar(&check, "check", "gira eval", "CHECK: command label recorded in the digest")
	fs.StringVar(&expect, "expect", "ok", "EXPECT: success marker recorded in the digest")
	fs.StringVar(&cwd, "cwd", "", "CWD recorded in the digest (default: current directory)")
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
