// Gate contract: success-only marker + digest, stale-on-drift fail-closed,
// --status (non-executing) / --reverify split, ABANDON as qualified handoff
// that never reports ALL MET (spec gira-evidence-gate F5, SC5).
//
// COPY = idea only, reimplemented in Go from the unlazy gate contract.
// CI mode is warn-only until E3 (owner decision 2026-10-14); local runs fail
// closed. No existing behavior changes.
package gate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Markers and statuses. Only MarkerSuccess ever lints; ABANDON never maps to ALL MET.
const (
	MarkerSuccess = "GATE-OK"
	StatusAllMet  = "ALL MET"
	StatusAbandon = "ABANDON"
)

// State is the persisted gate ledger entry for one scope.
type State struct {
	Marker string `json:"marker"`
	Digest string `json:"digest"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// StatusForAbandon returns the handoff status for an abandoned scope.
// It is never ALL MET by construction.
func StatusForAbandon() string { return StatusAbandon }

// WriteState persists one gate entry (additive ledger file).
func WriteState(path string, s State) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Status reads the ledger without executing anything (--status path).
func Status(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// Reverify recomputes against the expected digest (--reverify path).
// Drift fails closed in every mode.
func Reverify(path, expectedDigest string) error {
	s, err := Status(path)
	if err != nil {
		return err
	}
	if s.Marker != MarkerSuccess {
		return fmt.Errorf("reverify: success-only marker missing (got %q)", s.Marker)
	}
	if s.Digest != expectedDigest {
		return fmt.Errorf("reverify: stale digest (have %q, want %q)", s.Digest, expectedDigest)
	}
	if s.Status == StatusAllMet || s.Status == StatusAbandon {
		return nil
	}
	return fmt.Errorf("reverify: unknown status %q", s.Status)
}

// LintOptions selects strictness and the CI warn-only path.
type LintOptions struct {
	Strict bool
	CI     bool
}

// LintResult is the gate-lint outcome.
type LintResult struct {
	Pass     bool
	Status   string
	Warnings []string
	Error    string
}

// LintFile lints one ledger file against the expected digest.
// Local strict mode fails closed; CI strict mode warns (owner decision).
func LintFile(path, expectedDigest string, opts LintOptions) LintResult {
	s, err := Status(path)
	if err != nil {
		return LintResult{Error: err.Error()}
	}
	fail := func(msg string) LintResult {
		if opts.CI {
			return LintResult{Pass: true, Status: s.Status, Warnings: []string{msg}}
		}
		return LintResult{Error: msg}
	}
	if s.Marker != MarkerSuccess {
		return fail(fmt.Sprintf("success-only marker missing (got %q)", s.Marker))
	}
	if s.Digest != expectedDigest {
		return fail(fmt.Sprintf("stale digest: have %q, want %q", s.Digest, expectedDigest))
	}
	if s.Status == StatusAbandon {
		return LintResult{Pass: true, Status: StatusAbandon, Warnings: []string{"qualified handoff ABANDON: " + s.Reason}}
	}
	if s.Status != StatusAllMet {
		return fail(fmt.Sprintf("unknown status %q", s.Status))
	}
	if !opts.Strict {
		return LintResult{Pass: true, Status: s.Status}
	}
	return LintResult{Pass: true, Status: StatusAllMet}
}
