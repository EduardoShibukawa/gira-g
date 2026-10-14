package eval_test

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC2 honesty follow-up (review F3–F4): plugs are rejected on both sides and
// a control line must be internally consistent (matched+mismatched ==
// sampled_n), so a bare assertion can never pass as a blind round.
func TestSCL2ControlHonesty(t *testing.T) {
	second := map[int]string{0: "a", 1: "b", 2: "c"}

	// Never plug via first either.
	if _, err := eval.BlindReDerive([]int{0, 1}, map[int]string{0: "a", 1: "b", 99: "plugged"}, second, "fp1", "fp1"); err == nil {
		t.Fatalf("plugged extra key in first must fail")
	}

	// Consistent lines validate; inconsistent ones fail.
	rep := eval.ControlReport(1000, 48, 2, 42)
	if err := rep.Validate(); err != nil {
		t.Fatalf("consistent control line must validate: %v", err)
	}
	for _, bad := range []eval.ControlTotal{
		{Total: 1000, Matched: 40, Mismatched: 2, SampledN: 50, Seed: 42}, // 40+2 != 50
		{Total: 1000, Matched: -1, Mismatched: 51, SampledN: 50, Seed: 42}, // negative
		{Total: -5, Matched: 0, Mismatched: 0, SampledN: 0, Seed: 42},      // negative total
		{Total: 1000, Matched: 48, Mismatched: 2, SampledN: 20, Seed: 42},  // wrong sampled_n
	} {
		if err := bad.Validate(); err == nil {
			t.Fatalf("inconsistent control line %+v must fail validation", bad)
		}
	}
	// SampledN must equal the rule for the total.
	rep100 := eval.ControlReport(100, 20, 0, 7)
	if rep100.SampledN != eval.SampleSize(100) {
		t.Fatalf("sampled_n must follow max(20,5%%), got %+v", rep100)
	}
}
