package mcp_test

import (
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func TestDebugCeiling(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

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
	fmt.Printf("Response: %+v\n", resp)
}
