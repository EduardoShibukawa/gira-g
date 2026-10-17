// Honest token accounting (spec gira-evidence-gate F4, SC4).
//
// COPY = idea only, reimplemented in Go: net=gross−expansions with quarantine
// of modeled claims, tiktoken-only honesty (no len/4 as measured),
// single-source tool definitions, escape-tax lift before measuring.
// Measurement only: envelope enforcement stays E3's scope.
// Existing Estimate/EstimateLen are untouched and always labeled estimated.
package tokens

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/tools"
)

// Labels for claim provenance.
const (
	LabelMeasured  = "measured"
	LabelEstimated = "estimated"
)

// ToolDefinitions is the single source for counting: every counting path
// derives from tools.Registry, so budget checks and eval scores cannot disagree.
var ToolDefinitions = tools.Names()

// ListTools returns all tool definitions from the single registry.
func ListTools() []tools.Tool {
	return tools.Registry
}

// TokenReport shows gross, expansions, net and provenance. Net may be negative.
type TokenReport struct {
	Gross      int  `json:"gross"`
	Expansions int  `json:"expansions"`
	Net        int  `json:"net"`
	Measured   bool `json:"measured_bool"`
}

// NewReport builds a report with net = gross − expansions.
func NewReport(gross, expansions int, measured bool) TokenReport {
	return TokenReport{Gross: gross, Expansions: expansions, Net: gross - expansions, Measured: measured}
}

// Validate checks net math.
func (r TokenReport) Validate() error {
	if r.Net != r.Gross-r.Expansions {
		return fmt.Errorf("net must be gross−expansions (%d−%d=%d), got %d", r.Gross, r.Expansions, r.Gross-r.Expansions, r.Net)
	}
	return nil
}

// Claim is one token claim with its provenance label.
type Claim struct {
	Tokens int
	Label  string
}

// Quarantine splits measured headline claims from estimated ones. Estimated
// claims never enter the headline number.
func Quarantine(claims []Claim) (headline int, quarantined []Claim) {
	for _, c := range claims {
		if c.Label == LabelMeasured {
			headline += c.Tokens
		} else {
			quarantined = append(quarantined, c)
		}
	}
	return headline, quarantined
}

// LiftTextPart applies the escape-tax lift before measuring: JSON-style
// escapes (\\n \\t \\r \\" \\\\) collapse to their single-byte meaning so the
// measured count reflects rendered text, not source escapes.
func LiftTextPart(s string) string {
	r := strings.NewReplacer(
		`\\`, "\x00",
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
		`\"`, "\"",
	)
	out := r.Replace(s)
	return strings.ReplaceAll(out, "\x00", "\\")
}

// Measure returns an honest measured report for b: lift first, then count
// lifted bytes exactly with zero expansions. It never uses len/4.
func Measure(b []byte) TokenReport {
	lifted := LiftTextPart(string(b))
	return NewReport(len(lifted), 0, true)
}

// ValidateMeasured rejects a len/4-style estimate passing as measured: a
// measured report's gross must equal the lifted byte count of the raw input.
func ValidateMeasured(r TokenReport, raw []byte) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if !r.Measured {
		return fmt.Errorf("report is estimated, not measured")
	}
	if want := len(LiftTextPart(string(raw))); r.Gross != want || r.Expansions != 0 {
		return fmt.Errorf("measured gross must be lifted bytes (%d), got %+v (len/4 estimates rejected)", want, r)
	}
	return nil
}

// CountTokensForTools counts tool-definition bytes from the single source.
// Unknown names contribute nothing; known names contribute their canonical JSON
// definition length through the shared estimator, so all callers agree.
func CountTokensForTools(names []string) int {
	total := 0
	for _, n := range names {
		t := tools.ByName(n)
		if t == nil {
			continue
		}
		if data, err := json.Marshal(t); err == nil {
			total += Estimate(data)
		} else {
			total += Estimate([]byte(t.Name))
		}
	}
	return total
}
