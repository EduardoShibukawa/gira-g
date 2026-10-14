package slices

import (
	"testing"
)

func TestFitDebug(t *testing.T) {
	// Simple test of fit with empty drops
	out, err := fit(500, []string{}, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		return func(b *Budget) ([]byte, error) {
			return []byte(`{"test":true}`), nil
		}, Dropped{}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Output: %s", string(out))
}
