package slices

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEmitDebug(t *testing.T) {
	// Create a simple EvidenceReport-like struct with Budget
	type TestReport struct {
		Data   string  `json:"data"`
		Budget *Budget `json:"budget,omitempty"`
	}

	render := func(r *TestReport) ([]byte, error) {
		return json.Marshal(r)
	}

	// Test fit with a build function that uses Budget
	out, err := fit(500, []string{}, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		r := &TestReport{Data: "test"}
		return func(b *Budget) ([]byte, error) {
			r.Budget = b
			return render(r)
		}, Dropped{}
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Output: %s\n", string(out))
}
