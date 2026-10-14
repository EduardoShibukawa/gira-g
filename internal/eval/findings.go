// L3 taxonomy + findings register + independent fix-check
// (spec gira-evidence-gate F3, SC3).
//
// COPY = idea only, reimplemented in Go: Gallon S1 verbatim labels
// (measured/checked/inferred/assumed) with a mapping table to Gira tiers,
// coordinator/reviewer severity, an additive reviews/findings.md register
// (pointer only, outside token-budgeted slices), and a fix-check that
// re-runs the real path independently instead of reading the fixer's log.
// No model client, no paid fan-out: personas are fixed templates.
package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Taxonomy is the Gallon S1 verbatim claim label.
type Taxonomy string

const (
	TaxonomyMeasured Taxonomy = "measured"
	TaxonomyChecked  Taxonomy = "checked"
	TaxonomyInferred Taxonomy = "inferred"
	TaxonomyAssumed  Taxonomy = "assumed"
)

// GiraTier maps each Gallon label to its Gira derivation tier. Unknown
// labels stay silent ("unknown", I1), never named.
func (t Taxonomy) GiraTier() string {
	switch t {
	case TaxonomyMeasured:
		return "extracted"
	case TaxonomyChecked:
		return "declared"
	case TaxonomyInferred, TaxonomyAssumed:
		return "inferred"
	default:
		return "unknown"
	}
}

// ValidTaxonomy reports whether s is a Gallon S1 verbatim label.
func ValidTaxonomy(s string) bool {
	switch Taxonomy(s) {
	case TaxonomyMeasured, TaxonomyChecked, TaxonomyInferred, TaxonomyAssumed:
		return true
	default:
		return false
	}
}

// Finding is one L3 finding: taxonomy + dual severity + optional real-path
// repro + register status. Code-reading without ReproPath is inferred at most.
type Finding struct {
	ID                  string
	Taxonomy            Taxonomy
	SeverityCoordinator string
	SeverityReviewer    string
	Summary             string
	ReproPath           string
	Status              string
}

// Validate enforces: taxonomy present and known, both severities present,
// summary present, and measured/checked require a real-path repro that
// exists on disk. A claim on a missing path is code-reading: inferred at
// most (autoresearch: the result must come from the run, not the claim).
func (f Finding) Validate() error {
	switch f.Taxonomy {
	case TaxonomyMeasured, TaxonomyChecked, TaxonomyInferred, TaxonomyAssumed:
	default:
		return fmt.Errorf("finding %q: missing or unknown taxonomy", f.ID)
	}
	if strings.TrimSpace(f.SeverityCoordinator) == "" || strings.TrimSpace(f.SeverityReviewer) == "" {
		return fmt.Errorf("finding %q: coordinator and reviewer severity required", f.ID)
	}
	if strings.TrimSpace(f.Summary) == "" {
		return fmt.Errorf("finding %q: summary required", f.ID)
	}
	if f.Taxonomy == TaxonomyMeasured || f.Taxonomy == TaxonomyChecked {
		p := strings.TrimSpace(f.ReproPath)
		if p == "" {
			return fmt.Errorf("finding %q: %s without a real-path repro is inferred at most", f.ID, f.Taxonomy)
		}
		if fi, err := os.Stat(p); err != nil || fi.IsDir() {
			return fmt.Errorf("finding %q: %s repro path %q is not a readable file (inferred at most)", f.ID, f.Taxonomy, f.ReproPath)
		}
	}
	return nil
}

// Register is the additive findings register. It persists as a small
// markdown table (ID-paired with status) and is referenced by pointer only,
// never embedded in token-budgeted slices.
type Register struct {
	Findings []Finding `json:"findings"`
}

// Save writes the register as reviews/findings.md-style markdown.
func (r Register) Save(path string) error {
	var b strings.Builder
	b.WriteString("# Findings register\n\n")
	b.WriteString("| ID | taxonomy | coordinator | reviewer | status | repro | summary |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	esc := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "|", "\\|")
		s = strings.ReplaceAll(s, "\n", " ")
		return s
	}
	for _, f := range r.Findings {
		if err := f.Validate(); err != nil {
			return err
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			esc(f.ID), esc(string(f.Taxonomy)), esc(f.SeverityCoordinator),
			esc(f.SeverityReviewer), esc(f.Status), esc(f.ReproPath), esc(f.Summary))
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// splitRow splits a register row on unescaped | and unescapes \| and \\.
func splitRow(line string) []string {
	var cells []string
	var cur strings.Builder
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '|':
			cells = append(cells, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	cells = append(cells, cur.String())
	if escaped {
		cells[len(cells)-1] += "\\"
	}
	return cells
}

// LoadRegister reads a register written by Save and validates every row:
// unknown taxonomies, missing severities and dangling repro claims fail
// closed instead of loading silently.
func LoadRegister(path string) (Register, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Register{}, err
	}
	var out Register
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "ID |") {
			continue
		}
		parts := splitRow(line)
		// leading + trailing empty cells
		if len(parts) < 9 {
			continue
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		f := Finding{
			ID: parts[1], Taxonomy: Taxonomy(parts[2]),
			SeverityCoordinator: parts[3], SeverityReviewer: parts[4],
			Status: parts[5], ReproPath: parts[6], Summary: parts[7],
		}
		if err := f.Validate(); err != nil {
			return Register{}, fmt.Errorf("loading register %s: %w", path, err)
		}
		out.Findings = append(out.Findings, f)
	}
	return out, nil
}

// FixCheck re-runs the finding's real path independently through rerun and
// never reads the fixer's log: rerun must be non-nil and must report fixed.
// It loads the register entry by ID so the check runs against the recorded
// repro path, not caller-supplied claims.
func FixCheck(findingID, registerPath string, rerun func(reproPath string) (bool, error)) error {
	if rerun == nil {
		return fmt.Errorf("fix-check %q: independent rerun required, fixer's log is not evidence", findingID)
	}
	reg, err := LoadRegister(registerPath)
	if err != nil {
		return err
	}
	var found *Finding
	for i := range reg.Findings {
		if reg.Findings[i].ID == findingID {
			found = &reg.Findings[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("fix-check %q: not in register", findingID)
	}
	if strings.TrimSpace(found.ReproPath) == "" {
		return fmt.Errorf("fix-check %q: no real-path repro recorded", findingID)
	}
	fixed, err := rerun(found.ReproPath)
	if err != nil {
		return fmt.Errorf("fix-check %q rerun failed: %w", findingID, err)
	}
	if !fixed {
		return fmt.Errorf("fix-check %q: real path still shows the bug", findingID)
	}
	return nil
}
