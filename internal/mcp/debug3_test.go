package mcp_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func TestDebugEmitWithError(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)

	report, err := engine.ExplainEvidence("service:frontend", "")
	if err != nil {
		t.Fatal(err)
	}

	// Test Emit with budget - check error
	out, err := report.Emit(slices.FormatJSON, 500)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}

	fmt.Println("Output with budget=500:")
	fmt.Println(string(out))

	// Check if budget field exists in output
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if budget, ok := parsed["budget"]; ok {
		fmt.Println("Budget found:", budget)
	} else {
		fmt.Println("NO BUDGET FIELD IN OUTPUT")
	}
}
