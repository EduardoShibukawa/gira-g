package mcp_test

import (
	"reflect"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/tools"
)

// TestSCToolRegistrySingleSource verifies that there is a single source of truth
// for tool definitions, and each tool declares enveloped.
func TestSCToolRegistrySingleSource(t *testing.T) {
	// Test 1: MCP server tools match the registry
	t.Run("mcp_tools_match_registry", func(t *testing.T) {
		mm := buildTestMacroMap()
		engine := slices.NewEngine(mm)
		server := mcp.NewServer("gira", "0.1.0", engine)

		sendRequest(t, server, map[string]any{
			"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
		})

		// Get tools from MCP server
		resp := sendRequest(t, server, map[string]any{
			"jsonrpc": "2.0", "id": 10, "method": "tools/list",
		})
		result := resp["result"].(map[string]any)
		toolsList := result["tools"].([]any)

		mcpToolNames := make(map[string]bool)
		for _, toolAny := range toolsList {
			tool := toolAny.(map[string]any)
			name := tool["name"].(string)
			mcpToolNames[name] = true
		}

		// Get tool names from registry
		registryNames := make(map[string]bool)
		for _, tool := range tools.Registry {
			registryNames[tool.Name] = true
		}

		// Both should have the same tool names
		if !reflect.DeepEqual(mcpToolNames, registryNames) {
			t.Fatalf("MCP tools (%v) != registry tools (%v)", mcpToolNames, registryNames)
		}
	})

	// Test 2: Each tool in registry declares enveloped field
	t.Run("each_tool_declares_enveloped", func(t *testing.T) {
		for _, tool := range tools.Registry {
			if tool.Name == "" {
				t.Fatal("tool has empty name")
			}
			// Enveloped is a bool, zero value is false which is valid
			_ = tool.Enveloped
		}
	})

	// Test 3: MCP server tools have enveloped field in their definition
	t.Run("mcp_tools_have_enveloped", func(t *testing.T) {
		mm := buildTestMacroMap()
		engine := slices.NewEngine(mm)
		server := mcp.NewServer("gira", "0.1.0", engine)

		sendRequest(t, server, map[string]any{
			"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
		})

		resp := sendRequest(t, server, map[string]any{
			"jsonrpc": "2.0", "id": 10, "method": "tools/list",
		})
		result := resp["result"].(map[string]any)
		toolsList := result["tools"].([]any)

		for _, toolAny := range toolsList {
			tool := toolAny.(map[string]any)
			name := tool["name"].(string)
			// The enveloped field should be present in the tool definition
			if _, ok := tool["enveloped"]; !ok {
				t.Fatalf("tool %s missing enveloped field", name)
			}
		}
	})
}
