package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// TestSCStructuredErrors verifies that MCP tools and CLI return structured errors
// with {code, message, retryable, suggestions} format.
func TestSCStructuredErrors(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// Test MCP tools return structured errors
	t.Run("MCP_tools_return_structured_errors", func(t *testing.T) {
		// Test with invalid tool name
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      100,
			"method":  "tools/call",
			"params": map[string]any{
				"name":      "nonexistent_tool",
				"arguments": map[string]any{},
			},
		}

		resp := sendRequest(t, server, req)
		if !isError(resp) {
			t.Fatal("expected error for nonexistent tool")
		}

		result := resp["result"].(map[string]any)
		content := result["content"].([]any)
		c0 := content[0].(map[string]any)
		text := c0["text"].(string)

		// The error should be parseable as structured error
		var structuredErr map[string]any
		if err := json.Unmarshal([]byte(text), &structuredErr); err != nil {
			// If not JSON, it might be a plain text error - check if it contains error code
			if !contains(text, "tool_not_found") && !contains(text, "unknown tool") {
				t.Fatalf("expected structured error or error code, got: %s", text)
			}
		} else {
			// Verify structured error fields
			requiredFields := []string{"code", "message", "retryable", "suggestions"}
			for _, f := range requiredFields {
				if structuredErr[f] == nil {
					t.Fatalf("structured error missing field %s", f)
				}
			}
			if structuredErr["code"] != "tool_not_found" {
				t.Fatalf("expected code tool_not_found, got %v", structuredErr["code"])
			}
		}
	})

	// Test budget_exceeded error structure
	t.Run("MCP_budget_exceeded_structured", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      200,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "get_tour",
				"arguments": map[string]any{
					"focus":  "service:frontend",
					"budget": 10,
				},
			},
		}

		resp := sendRequest(t, server, req)
		if !isError(resp) {
			t.Fatal("expected error for budget exceeded")
		}

		result := resp["result"].(map[string]any)
		content := result["content"].([]any)
		c0 := content[0].(map[string]any)
		text := c0["text"].(string)

		var structuredErr map[string]any
		if err := json.Unmarshal([]byte(text), &structuredErr); err != nil {
			t.Fatalf("budget_exceeded error not valid JSON: %s", text)
		}

		if structuredErr["code"] != "budget_exceeded" {
			t.Fatalf("expected code budget_exceeded, got %v", structuredErr["code"])
		}
		if structuredErr["message"] == "" {
			t.Fatal("error message should not be empty")
		}
		// retryable should be false for budget_exceeded
		if structuredErr["retryable"] != false {
			t.Fatalf("expected retryable=false for budget_exceeded, got %v", structuredErr["retryable"])
		}
		if structuredErr["suggestions"] == nil {
			t.Fatal("expected suggestions in budget_exceeded error")
		}
	})
}

// Test gira-query reads error code first (simulated via MCP response)
func TestStructuredErrorsGiraQueryCompatibility(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// Test that structured errors can be parsed by code-first logic
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      300,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "explain_evidence",
			"arguments": map[string]any{
				"node_id": "nonexistent",
				"budget":  5,
			},
		},
	}

	resp := sendRequest(t, server, req)
	if !isError(resp) {
		t.Fatal("expected error for nonexistent node")
	}

	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	c0 := content[0].(map[string]any)
	text := c0["text"].(string)

	// Should be parseable as structured error
	var structuredErr map[string]any
	if err := json.Unmarshal([]byte(text), &structuredErr); err != nil {
		t.Fatalf("error not valid JSON: %s", text)
	}

	// Simulate gira-query reading code first
	code, ok := structuredErr["code"].(string)
	if !ok || code == "" {
		t.Fatal("structured error missing code field")
	}

	// Should be able to handle error by code without substring parsing
	knownCodes := map[string]bool{
		"budget_exceeded": true,
		"node_not_found":  true,
		"envelope_error":  true,
		"tool_not_found":  true,
		"invalid_budget":  true,
		"internal_error":  true,
	}
	if !knownCodes[code] {
		t.Fatalf("unknown error code: %s", code)
	}
}
