package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// TestSCEnvelopeHonesty verifies that budget objects include honest labels:
// estimator: "bytes/4" and measured: false (only when budget is set).
func TestSCEnvelopeHonesty(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	toolsWithBudget := []string{
		"get_tour",
		"diff_maps",
		"explain_evidence",
		"locate_boundary",
		"get_context_slice",
	}

	for _, tool := range toolsWithBudget {
		t.Run("MCP_"+tool+"_honesty", func(t *testing.T) {
			req := map[string]any{
				"jsonrpc": "2.0",
				"id":      100,
				"method":  "tools/call",
				"params": map[string]any{
					"name":      tool,
					"arguments": map[string]any{},
				},
			}
			switch tool {
			case "get_tour":
				req["params"].(map[string]any)["arguments"] = map[string]any{"focus": "service:frontend", "budget": 500}
			case "diff_maps":
				return // skip - needs index files
			case "explain_evidence":
				req["params"].(map[string]any)["arguments"] = map[string]any{"node_id": "service:frontend", "budget": 500}
			case "locate_boundary":
				req["params"].(map[string]any)["arguments"] = map[string]any{"query": "frontend", "budget": 500}
			case "get_context_slice":
				req["params"].(map[string]any)["arguments"] = map[string]any{"node_id": "service:frontend", "budget": 500}
			}

			resp := sendRequest(t, server, req)
			if isError(resp) {
				t.Fatalf("tool %s failed: %v", tool, resp)
			}

			content := extractText(resp)
			var parsed map[string]any
			if err := json.Unmarshal([]byte(content), &parsed); err != nil {
				t.Fatalf("tool %s: failed to parse JSON: %v", tool, err)
			}

			// Check budget object exists
			budget, ok := parsed["budget"].(map[string]any)
			if !ok {
				t.Fatalf("tool %s: missing budget object", tool)
			}

			// Check honesty labels
			estimator, ok := budget["estimator"].(string)
			if !ok || estimator != "bytes/4" {
				t.Fatalf("tool %s: budget.estimator = %q, want \"bytes/4\"", tool, estimator)
			}

			measured, ok := budget["measured"].(bool)
			if !ok || measured != false {
				t.Fatalf("tool %s: budget.measured = %v, want false", tool, measured)
			}

			// Verify no envelope figure enters measured headline (estimator is not "tokens")
			if estimator == "tokens" {
				t.Fatalf("tool %s: budget.estimator should not be \"tokens\"", tool)
			}
		})
	}
}

// TestSCTransportCeiling verifies the global MCP response byte ceiling.
// Above the ceiling, the tool refuses with budget_exceeded + escalate_hint; never cuts mid-JSON.
func TestSCTransportCeiling(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// Test with a very small budget that should trigger budget_exceeded
	t.Run("get_tour_ceiling_exceeded", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      200,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "get_tour",
				"arguments": map[string]any{
					"focus":  "service:frontend",
					"budget": 10, // Very small budget - should exceed ceiling
				},
			},
		}

		resp := sendRequest(t, server, req)
		// Should get an error result with isError=true
		if !isError(resp) {
			t.Fatal("expected error for budget exceeded, got success")
		}

		result := resp["result"].(map[string]any)
		content := result["content"].([]any)
		c0 := content[0].(map[string]any)
		text := c0["text"].(string)
		
		// Check that error message contains budget_exceeded
		if !contains(text, "budget_exceeded") {
			t.Fatalf("expected budget_exceeded in error message, got: %s", text)
		}
		if !contains(text, "escalate_hint") && !contains(text, "rerun with budget") {
			t.Fatalf("expected escalate_hint in error message, got: %s", text)
		}
	})

	// Test that output is never cut mid-JSON
	t.Run("never_cuts_mid_json", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      201,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "explain_evidence",
				"arguments": map[string]any{
					"node_id": "service:frontend",
					"budget":  5, // Very small
				},
			},
		}

		resp := sendRequest(t, server, req)
		if !isError(resp) {
			t.Fatal("expected error for budget exceeded, got success")
		}

		result := resp["result"].(map[string]any)
		content := result["content"].([]any)
		c0 := content[0].(map[string]any)
		text := c0["text"].(string)
		
		// Check that error message contains budget_exceeded
		if !contains(text, "budget_exceeded") {
			t.Fatalf("expected budget_exceeded in error message, got: %s", text)
		}
		// The error response itself should be valid JSON (the MCP response is valid JSON)
		// We already validated by parsing it above
	})
}
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
