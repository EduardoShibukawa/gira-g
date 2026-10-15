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
// E5 (Q6=A) adds `--gc [--keep-resolved <n>]` (default 0): the register is
// compacted so every `open` finding survives in order while resolved
// (`fixed`/`closed`/`verified`) findings shrink to the last <n> entries;
// the compacted register is validated and saved atomically via Save.
func runReview(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var registerPath, finding, expectContent string
	var fixCheck bool
	var gc bool
	var keepResolved int
	fs.StringVar(&registerPath, "findings-register", "reviews/findings.md", "Path to the findings register (pointer only, outside slices)")
	fs.BoolVar(&fixCheck, "fix-check", false, "Re-run the finding's real path independently")
	fs.StringVar(&finding, "finding", "", "Finding ID for --fix-check")
	fs.StringVar(&expectContent, "expect-content", "", "Expected repro-path bytes for --fix-check (the path is always read)")
	fs.BoolVar(&gc, "gc", false, "Compact the findings register: keep all open findings, prune resolved ones")
	fs.IntVar(&keepResolved, "keep-resolved", 0, "Resolved findings to keep when --gc compacts (negative treated as 0)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if gc {
		if fixCheck {
			return fmt.Errorf("--gc cannot be combined with --fix-check")
		}
		reg, err := eval.LoadRegister(registerPath)
		if err != nil {
			return fmt.Errorf("loading register %s: %w", registerPath, err)
		}
		compacted := reg.GC(keepResolved)
		if err := compacted.Save(registerPath); err != nil {
			return fmt.Errorf("saving compacted register %s: %w", registerPath, err)
		}
		kept := len(compacted.Findings)
		open := 0
		for _, f := range compacted.Findings {
			if f.Status == "open" {
				open++
			}
		}
		fmt.Fprintf(stdout, "register %s: compacted to %d findings (%d open, %d resolved kept)\n",
			registerPath, kept, open, kept-open)
		return nil
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
