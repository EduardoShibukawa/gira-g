package mcp_test

import (
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func TestDebugEmitDirect(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)

	report, err := engine.ExplainEvidence("service:frontend", "")
	if err != nil {
		t.Fatal(err)
	}

	// Test Emit with budget
	out, err := report.Emit(slices.FormatJSON, 500)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Output with budget=500:")
	fmt.Println(string(out))

	// Test Emit without budget
	out2, err := report.Emit(slices.FormatJSON, 0)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("\nOutput with budget=0:")
	fmt.Println(string(out2))
}
