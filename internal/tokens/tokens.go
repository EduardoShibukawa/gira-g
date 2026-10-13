// Package tokens holds gira's deterministic token estimator. It is deliberately not a
// model tokenizer: ADR 0001 forbids a model dependency, and a tokenizer would differ
// per host. Budgets (ADR 0006) and the evaluator both use this one function, so a
// budget check and a token-reduction score can never disagree about a count.
package tokens

// BytesPerToken is the documented ratio of the estimator: one token per four bytes
// of UTF-8 output, rounded up. Real per-model counts are recorded by the Layer 1
// benchmark runner from each model's API, never estimated here.
const BytesPerToken = 4

// Estimate returns the estimated token count of b.
func Estimate(b []byte) int {
	return EstimateLen(len(b))
}

// EstimateLen returns the estimated token count of n bytes.
func EstimateLen(n int) int {
	if n <= 0 {
		return 0
	}
	return (n + BytesPerToken - 1) / BytesPerToken
}
