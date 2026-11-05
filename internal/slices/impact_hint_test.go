package slices

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestImpactUnknownTargetHint(t *testing.T) {
	e := NewEngine(&domain.MacroMap{})
	_, err := e.GetImpactBriefing("internal/x.go", 1)
	if err == nil || !strings.Contains(err.Error(), "expected a node ID") {
		t.Fatalf("want node-ID format hint, got %v", err)
	}
}
