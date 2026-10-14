// L2 control-total + seeded sampling (spec gira-evidence-gate F2, SC2).
//
// COPY = idea only, reimplemented in Go: control-total doctrine, frozen A/B
// digests compared before reconciliation, seeded max(20,5%) agreement sampling,
// never-force-match / never-plug / never-drop invariants. No model calls;
// the sampler is deterministic (math/rand seeded) over indices only, so it
// cannot leak ground truth.
package eval

import (
	"fmt"
	"math/rand"
	"sort"
)

// ControlTotal is emitted per run: the control denominator, the blind
// re-derivation outcome on the seeded sample, and the seed.
type ControlTotal struct {
	Total      int    `json:"total"`
	Matched    int    `json:"matched"`
	Mismatched int    `json:"mismatched"`
	SampledN   int    `json:"sampled_n"`
	Seed       uint64 `json:"seed"`
}

// SampleSize applies the Gallon default, owner decision 2026-10-14: strict
// max(20,5%). Totals at or below 20 sample everything; above, the sample is
// the larger of 20 and 5% (rounded up), capped at the total.
func SampleSize(total int) int {
	if total <= 20 {
		if total < 0 {
			return 0
		}
		return total
	}
	n := (total*5 + 99) / 100 // ceil(5%)
	if n < 20 {
		n = 20
	}
	if n > total {
		n = total
	}
	return n
}

// SeededSample returns SampleSize(total) distinct indices in [0,total),
// deterministically shuffled by seed. It sees only the total, never values.
func SeededSample(total int, seed uint64) []int {
	n := SampleSize(total)
	if n <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(int64(seed)))
	perm := r.Perm(total)
	out := append([]int(nil), perm[:n]...)
	sort.Ints(out)
	return out
}

// ControlReport builds the per-run control line for a finished blind round.
func ControlReport(total, matched, mismatched int, seed uint64) ControlTotal {
	return ControlTotal{
		Total:      total,
		Matched:    matched,
		Mismatched: mismatched,
		SampledN:   SampleSize(total),
		Seed:       seed,
	}
}

// CompareFrozenDigests enforces frozen A/B comparison before reconciliation:
// differing input digests fail closed instead of comparing values.
func CompareFrozenDigests(a, b string) error {
	if a != b {
		return fmt.Errorf("frozen digests differ before reconciliation: %q vs %q", a, b)
	}
	return nil
}

// BlindResult is the outcome of one blind re-derivation round on the sample.
type BlindResult struct {
	Matched    int    `json:"matched"`
	Mismatched int    `json:"mismatched"`
	Diff       string `json:"diff,omitempty"`
}

// BlindReDerive compares two independent derivations on exactly the sample
// indices. Agreement must be 100% or an unplugged diff is reported: the
// function never plugs extra keys, never drops a sample index, and never
// force-matches differing values. Frozen digests are compared first.
func BlindReDerive(sample []int, first, second map[int]string, frozenA, frozenB string) (BlindResult, error) {
	if err := CompareFrozenDigests(frozenA, frozenB); err != nil {
		return BlindResult{}, err
	}
	for k := range second {
		found := false
		for _, s := range sample {
			if s == k {
				found = true
				break
			}
		}
		if !found {
			return BlindResult{}, fmt.Errorf("plug rejected: key %d outside the seeded sample", k)
		}
	}
	var matched, mismatched int
	var diffs []string
	for _, s := range sample {
		a, okA := first[s]
		b, okB := second[s]
		if !okA || !okB {
			return BlindResult{}, fmt.Errorf("drop rejected: sample index %d missing from a derivation", s)
		}
		if a == b {
			matched++
		} else {
			mismatched++
			diffs = append(diffs, fmt.Sprintf("index %d: %q vs %q", s, a, b))
		}
	}
	res := BlindResult{Matched: matched, Mismatched: mismatched}
	if mismatched > 0 {
		res.Diff = fmt.Sprintf("blind sample %d/%d matched; diffs: %v", matched, len(sample), diffs)
	}
	return res, nil
}

// AssertNoForceMatch rejects a caller-side claim of full agreement when the
// values on the sample actually differ. claimedMatch=true with any divergence
// is the force-match violation.
func AssertNoForceMatch(first, second map[int]string, sample []int, claimedMatch bool) error {
	if !claimedMatch {
		return nil
	}
	for _, s := range sample {
		if first[s] != second[s] {
			return fmt.Errorf("force-match rejected at index %d", s)
		}
	}
	return nil
}
