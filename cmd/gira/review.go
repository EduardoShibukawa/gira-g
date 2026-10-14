package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// runReview is the additive F3 entry point:
// `gira review --findings-register <path> [--fix-check --finding <id> --expect-content <s>]`.
// The fix-check rerun reads the recorded repro path and compares its bytes
// with --expect-content: the path is always touched, self-attestation is
// impossible (no flag can bypass the read).
func runReview(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var registerPath, finding, expectContent string
	var fixCheck bool
	fs.StringVar(&registerPath, "findings-register", "reviews/findings.md", "Path to the findings register (pointer only, outside slices)")
	fs.BoolVar(&fixCheck, "fix-check", false, "Re-run the finding's real path independently")
	fs.StringVar(&finding, "finding", "", "Finding ID for --fix-check")
	fs.StringVar(&expectContent, "expect-content", "", "Expected repro-path bytes for --fix-check (the path is always read)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !fixCheck {
		reg, err := eval.LoadRegister(registerPath)
		if err != nil {
			return fmt.Errorf("loading register %s: %w", registerPath, err)
		}
		fmt.Fprintf(stdout, "register %s: %d findings\n", registerPath, len(reg.Findings))
		for _, f := range reg.Findings {
			fmt.Fprintf(stdout, "- %s [%s→%s] %s/%s status=%s\n",
				f.ID, f.Taxonomy, f.Taxonomy.GiraTier(), f.SeverityCoordinator, f.SeverityReviewer, f.Status)
		}
		return nil
	}
	if finding == "" {
		return fmt.Errorf("--fix-check requires --finding <id>")
	}
	rerun := func(reproPath string) (bool, error) {
		data, err := os.ReadFile(reproPath)
		if err != nil {
			return false, err
		}
		return string(data) == expectContent, nil
	}
	if err := eval.FixCheck(finding, registerPath, rerun); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "fix-check %s: pass (independent rerun)\n", finding)
	return nil
}
