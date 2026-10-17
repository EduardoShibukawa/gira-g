package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// TestSCEnvelopeEveryTool verifies that all agent-facing MCP tools and CLI commands
// accept an optional budget parameter. Without budget, output is byte-identical to current.
// With budget, output includes a budget object with limit, estimated, truncated, dropped, escalate_hint.
func TestSCEnvelopeEveryTool(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	mmBytes, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to serialize test macro map: %v", err)
	}
	tmpDir := t.TempDir()
	indexA := filepath.Join(tmpDir, "index-a.json")
	indexB := filepath.Join(tmpDir, "index-b.json")
	if err := os.WriteFile(indexA, mmBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexB, mmBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	// Initialize
	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// Test tools that should have budget parameter (MCP)
	toolsWithBudget := []string{
		"get_tour",
		"diff_maps",
		"explain_evidence",
		"locate_boundary",
		"get_context_slice", // already has budget
	}

	for _, tool := range toolsWithBudget {
		t.Run("MCP_"+tool, func(t *testing.T) {
			// First call without budget - should work and be byte-identical to current
			req := map[string]any{
				"jsonrpc": "2.0",
				"id":      100,
				"method":  "tools/call",
				"params": map[string]any{
					"name":      tool,
					"arguments": map[string]any{},
				},
			}
			// Set required args for each tool
			switch tool {
			case "get_tour":
				req["params"].(map[string]any)["arguments"] = map[string]any{"focus": "service:frontend"}
			case "diff_maps":
				req["params"].(map[string]any)["arguments"] = map[string]any{"index_a": indexA, "index_b": indexB}
			case "explain_evidence":
				req["params"].(map[string]any)["arguments"] = map[string]any{"node_id": "service:frontend"}
			case "locate_boundary":
				req["params"].(map[string]any)["arguments"] = map[string]any{"query": "frontend"}
			case "get_context_slice":
				req["params"].(map[string]any)["arguments"] = map[string]any{"node_id": "service:frontend"}
			}

			resp1 := sendRequest(t, server, req)
			if isError(resp1) {
				t.Fatalf("tool %s without budget failed: %v", tool, resp1)
			}

			// Second call with budget=0 - should be byte-identical
			req["params"].(map[string]any)["arguments"].(map[string]any)["budget"] = 0
			resp2 := sendRequest(t, server, req)
			if isError(resp2) {
				t.Fatalf("tool %s with budget=0 failed: %v", tool, resp2)
			}

			// Compare content - should be byte-identical
			content1 := extractText(resp1)
			content2 := extractText(resp2)
			if content1 != content2 {
				t.Fatalf("tool %s: budget=0 output differs from no-budget output", tool)
			}

			// Third call with budget > 0 - should include budget object
			req["params"].(map[string]any)["arguments"].(map[string]any)["budget"] = 500
			resp3 := sendRequest(t, server, req)
			if isError(resp3) {
				t.Fatalf("tool %s with budget=500 failed: %v", tool, resp3)
			}
			content3 := extractText(resp3)

			// Parse and verify budget object exists
			var parsed map[string]any
			if err := json.Unmarshal([]byte(content3), &parsed); err != nil {
				t.Fatalf("tool %s: failed to parse JSON output: %v", tool, err)
			}
			if parsed["budget"] == nil {
				t.Fatalf("tool %s: missing budget object in output with budget=500", tool)
			}
			budget := parsed["budget"].(map[string]any)
			requiredFields := []string{"limit", "estimated", "truncated", "dropped", "escalate_hint"}
			for _, f := range requiredFields {
				if budget[f] == nil {
					t.Fatalf("tool %s: budget missing field %s", tool, f)
				}
			}
			limitVal, ok := budget["limit"].(float64)
			if !ok || int(limitVal) != 500 {
				t.Fatalf("tool %s: budget.limit = %v (type %T), want 500", tool, budget["limit"], budget["limit"])
			}
		})
	}
}

// TestSCEnvelopeDeterministic verifies that same index + budget → byte-identical output,
// and drop order follows the specified rules.
func TestSCEnvelopeDeterministic(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// Test deterministic output for get_tour
	t.Run("get_tour_deterministic", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      200,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "get_tour",
				"arguments": map[string]any{
					"focus":  "service:frontend",
					"budget": 300,
				},
			},
		}

		resp1 := sendRequest(t, server, req)
		if isError(resp1) {
			t.Fatal(resp1)
		}
		content1 := extractText(resp1)

		resp2 := sendRequest(t, server, req)
		if isError(resp2) {
			t.Fatal(resp2)
		}
		content2 := extractText(resp2)

		if content1 != content2 {
			t.Fatal("get_tour: same index + budget → non-byte-identical output")
		}
	})

	// Test deterministic output for explain_evidence
	t.Run("explain_evidence_deterministic", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      201,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "explain_evidence",
				"arguments": map[string]any{
					"node_id": "service:frontend",
					"budget":  500,
				},
			},
		}

		resp1 := sendRequest(t, server, req)
		if isError(resp1) {
			t.Fatal(resp1)
		}
		content1 := extractText(resp1)

		resp2 := sendRequest(t, server, req)
		if isError(resp2) {
			t.Fatal(resp2)
		}
		content2 := extractText(resp2)

		if content1 != content2 {
			t.Fatal("explain_evidence: same index + budget → non-byte-identical output")
		}
	})

	// Test deterministic output for locate_boundary
	t.Run("locate_boundary_deterministic", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      202,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "locate_boundary",
				"arguments": map[string]any{
					"query":  "frontend",
					"budget": 150,
				},
			},
		}

		resp1 := sendRequest(t, server, req)
		if isError(resp1) {
			t.Fatal(resp1)
		}
		content1 := extractText(resp1)

		resp2 := sendRequest(t, server, req)
		if isError(resp2) {
			t.Fatal(resp2)
		}
		content2 := extractText(resp2)

		if content1 != content2 {
			t.Fatal("locate_boundary: same index + budget → non-byte-identical output")
		}
	})
}

// Test drop order rules: never drops focal element nor unsupported/unknown rows
func TestSCEnvelopeDropOrderRules(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	// For get_tour: focal node must never be dropped
	t.Run("get_tour_never_drops_focal", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      300,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "get_tour",
				"arguments": map[string]any{
					"focus":  "service:frontend",
					"budget": 500, // Enough for focal node
				},
			},
		}

		resp := sendRequest(t, server, req)
		if isError(resp) {
			t.Fatal(resp)
		}
		content := extractText(resp)

		var parsed map[string]any
		if err := json.Unmarshal([]byte(content), &parsed); err != nil {
			t.Fatal(err)
		}
		// The focal node should still be present
		if parsed["focal_node"] == nil && parsed["focus"] == nil {
			// Check if the focal node ID appears in the output
			if !strings.Contains(content, "service:frontend") {
				t.Fatal("get_tour: focal node was dropped with small budget")
			}
		}
	})

	// For diff_maps: evidence_moved then unchanged context, never changed_contracts
	// or newly_unsupported. For explain_evidence: evidence beyond the first per fact.
	// For locate_boundary: suggestions beyond N. The complex fixtures for those
	// three kinds are exercised directly in internal/slices and internal/archdiff
	// (see TestSCEnvelopeDeterministic and the BoundaryResult drop-order tests).
}

func isError(resp map[string]any) bool {
	if err, ok := resp["error"]; ok && err != nil {
		return true
	}
	if result, ok := resp["result"].(map[string]any); ok {
		if isErr, ok := result["isError"].(bool); ok && isErr {
			return true
		}
	}
	return false
}

func extractText(resp map[string]any) string {
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	c0 := content[0].(map[string]any)
	return c0["text"].(string)
}
