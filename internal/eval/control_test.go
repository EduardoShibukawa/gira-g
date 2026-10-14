package eval_test

import (
	"encoding/json"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
)

// SC2 (F2): control-total + seeded max(20,5%) sampling, blind re-derivation
// matches 100% or reports an unplugged diff; never plug/drop/force-match.
// Frozen digests are compared before reconciliation.
func TestSCL2ControlSample(t *testing.T) {
	// max(20,5%) rule.
	for total, want := range map[int]int{0: 0, 10: 10, 20: 20, 100: 20, 400: 20, 1000: 50, 2000: 100} {
		if got := eval.SampleSize(total); got != want {
			t.Errorf("SampleSize(%d) = %d, want %d", total, got, want)
		}
	}

	// Seeded determinism: same seed → same sample; indices unique + in range.
	a := eval.SeededSample(1000, 42)
	b := eval.SeededSample(1000, 42)
	if len(a) != 50 || len(b) != 50 {
		t.Fatalf("sample size for 1000 must be 50, got %d/%d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("same seed must give same sample")
		}
	}
	seen := map[int]bool{}
	for _, idx := range a {
		if idx < 0 || idx >= 1000 || seen[idx] {
			t.Fatalf("sample index out of range or duplicated: %d", idx)
		}
		seen[idx] = true
	}
	c := eval.SeededSample(1000, 43)
	same := len(a) == len(c)
	if same {
		diff := false
		for i := range a {
			if a[i] != c[i] {
				diff = true
				break
			}
		}
		same = !diff
	}
	if same {
		t.Fatalf("different seeds should (overwhelmingly likely) differ")
	}

	// Control-total emitted per run carries all five fields.
	rep := eval.ControlReport(1000, 48, 2, 42)
	raw, _ := json.Marshal(rep)
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"total", "matched", "mismatched", "sampled_n", "seed"} {
		if _, ok := decoded[k]; !ok {
			t.Fatalf("control report must carry %q, got %s", k, raw)
		}
	}

	// Frozen digests compared before reconciliation: mismatch fails closed.
	if err := eval.CompareFrozenDigests("abc", "abc"); err != nil {
		t.Fatalf("equal digests must compare clean: %v", err)
	}
	if err := eval.CompareFrozenDigests("abc", "def"); err == nil {
		t.Fatalf("differing frozen digests must fail before reconciliation")
	}

	// Blind re-derivation on agreement: 100% match, no diff.
	first := map[int]string{0: "a", 1: "b", 2: "c"}
	second := map[int]string{0: "a", 1: "b", 2: "c"}
	res, err := eval.BlindReDerive([]int{0, 1, 2}, first, second, "fp1", "fp1")
	if err != nil || res.Mismatched != 0 || res.Matched != 3 || res.Diff != "" {
		t.Fatalf("agreement must be 100%% match, got %+v err=%v", res, err)
	}

	// Genuine divergence: reports an unplugged diff, never a forced match.
	second[1] = "DIFFERENT"
	res, err = eval.BlindReDerive([]int{0, 1, 2}, first, second, "fp1", "fp1")
	if err != nil {
		t.Fatalf("divergence is reported, not an error: %v", err)
	}
	if res.Matched != 2 || res.Mismatched != 1 || res.Diff == "" {
		t.Fatalf("divergence must report matched=2 mismatched=1 + diff, got %+v", res)
	}

	// Never plug: extra keys outside the sample fail.
	if _, err := eval.BlindReDerive([]int{0, 1}, first, map[int]string{0: "a", 1: "b", 99: "plugged"}, "fp1", "fp1"); err == nil {
		t.Fatalf("plugged extra key must fail")
	}
	// Never drop: missing sample index fails.
	if _, err := eval.BlindReDerive([]int{0, 1, 2}, first, map[int]string{0: "a", 1: "b"}, "fp1", "fp1"); err == nil {
		t.Fatalf("dropped sample index must fail")
	}
	// Never force-match: caller claiming match over differing values is rejected.
	if err := eval.AssertNoForceMatch(first, map[int]string{0: "a", 1: "DIFFERENT"}, []int{0, 1}, true); err == nil {
		t.Fatalf("forced match claim over differing values must fail")
	}
	// Frozen drift fails before any comparison.
	if _, err := eval.BlindReDerive([]int{0}, first, first, "fp1", "fp2"); err == nil {
		t.Fatalf("frozen digest drift must fail closed")
	}
}
