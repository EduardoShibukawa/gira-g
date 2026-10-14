package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EduardoShibukawa/gira-g/internal/gate"
)

// runGateLint is the additive F5 entry point: `gira gate-lint --strict`.
// CI mode (GIRA_CI=1) is warn-only until E3; local runs fail closed.
func runGateLint(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("gate-lint", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var strict bool
	var state, digest string
	fs.BoolVar(&strict, "strict", false, "Strict lint (fail-closed locally, warn-only in CI)")
	fs.StringVar(&state, "state", ".gira/gate.json", "Gate ledger file")
	fs.StringVar(&digest, "digest", "", "Expected digest hex")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ci := os.Getenv("GIRA_CI") != "" || os.Getenv("CI") != ""
	res := gate.LintFile(state, digest, gate.LintOptions{Strict: strict, CI: ci})
	for _, w := range res.Warnings {
		fmt.Fprintf(stderr, "warning: %s\n", w)
	}
	if !res.Pass {
		return fmt.Errorf("gate-lint: %s", res.Error)
	}
	fmt.Fprintf(stdout, "gate-lint: %s\n", res.Status)
	return nil
}

// runGate implements `gira gate --status | --reverify` (F5 split).
func runGate(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var showStatus, reverify bool
	var state, digest string
	fs.BoolVar(&showStatus, "status", false, "Non-executing status read")
	fs.BoolVar(&reverify, "reverify", false, "Re-run verification against the digest")
	fs.StringVar(&state, "state", ".gira/gate.json", "Gate ledger file")
	fs.StringVar(&digest, "digest", "", "Expected digest hex for --reverify")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if showStatus == reverify {
		return fmt.Errorf("gira gate requires exactly one of --status or --reverify")
	}
	if showStatus {
		s, err := gate.Status(state)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "gate status: %s (marker=%s digest=%s)\n", s.Status, s.Marker, s.Digest)
		if s.Reason != "" {
			fmt.Fprintf(stdout, "reason: %s\n", s.Reason)
		}
		return nil
	}
	if err := gate.Reverify(state, digest); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "reverify: pass")
	return nil
}
