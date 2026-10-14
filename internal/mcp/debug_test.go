package mcp_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func TestDebugExplainEvidence(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{},
	})

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "explain_evidence",
			"arguments": map[string]any{
				"node_id": "service:frontend",
				"budget":  500,
			},
		},
	}

	resp := sendRequest(t, server, req)
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	c0 := content[0].(map[string]any)
	text := c0["text"].(string)
	fmt.Println("Output:", text)

	var parsed map[string]any
	json.Unmarshal([]byte(text), &parsed)
	fmt.Println("Has budget:", parsed["budget"] != nil)
	if parsed["budget"] != nil {
		budget := parsed["budget"].(map[string]any)
		fmt.Println("Budget:", budget)
	}
}
