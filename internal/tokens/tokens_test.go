package tokens_test

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

func TestEstimateRoundsUp(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 4: 1, 5: 2, 4000: 1000}
	for n, want := range cases {
		if got := tokens.EstimateLen(n); got != want {
			t.Errorf("EstimateLen(%d) = %d, want %d", n, got, want)
		}
	}
	if tokens.Estimate([]byte("abcdefgh")) != 2 {
		t.Error("Estimate must agree with EstimateLen")
	}
}
