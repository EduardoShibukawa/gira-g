package eval

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// F7: an unsupported construct inside a parsed file does not report an
// expected unsupported path; only an unsupported file does.
func TestUnsupportedSafetyIgnoresConstructRows(t *testing.T) {
	mm := domain.NewMacroMap("t")
	mm.Unsupported = []domain.UnsupportedBoundary{
		{Path: "src/catalog/main.go", Reason: "OpAMP client"},
		{Path: "src/worker/app.py", Reason: "no extractor registered for file extension"},
	}
	mm.UnsupportedConstructs = []string{"src/catalog/main.go"}

	if got, v := UnsupportedSafety(mm, []string{"src/worker"}); got != 100 || len(v) != 0 {
		t.Errorf("unsupported file must report its path: %.1f %v", got, v)
	}
	if got, v := UnsupportedSafety(mm, []string{"src/catalog"}); got != 0 || len(v) != 1 {
		t.Errorf("construct row must not report an unsupported path: %.1f %v", got, v)
	}
}
