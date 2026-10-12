package mcp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"labs/gira-g/internal/domain"
	"labs/gira-g/internal/mcp"
	"labs/gira-g/internal/slices"
)

func buildTestMacroMap() *domain.MacroMap {
	mm := domain.NewMacroMap("test-sys")

	_ = mm.AddNode(domain.Node{
		ID:   "service:frontend",
		Kind: domain.NodeService,
		Name: "frontend",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-fe",
					CommitSHA: "1111111111111111111111111111111111111111",
					Path:      "main.go",
					StartLine: 1,
					EndLine:   10,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "dev",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      time.Now(),
			},
		},
	})

	_ = mm.AddNode(domain.Node{
		ID:   "contract:auth",
		Kind: domain.NodeContract,
		Name: "AuthService",
		Metadata: map[string]any{
			"methods": []string{"Login", "Logout"},
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-auth",
					CommitSHA: "2222222222222222222222222222222222222222",
					Path:      "proto/auth.proto",
					StartLine: 5,
					EndLine:   15,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "prod",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      time.Now(),
			},
		},
	})

	_ = mm.AddNode(domain.Node{
		ID:   "service:auth-backend",
		Kind: domain.NodeService,
		Name: "auth-backend",
	})

	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:fe-consumes-auth",
		SourceID: "service:frontend",
		TargetID: "contract:auth",
		Verb:     domain.VerbConsumes,
	})

	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:backend-provides-auth",
		SourceID: "service:auth-backend",
		TargetID: "contract:auth",
		Verb:     domain.VerbProvides,
	})

	mm.Unsupported = append(mm.Unsupported, domain.UnsupportedBoundary{
		Path:   "scripts/deploy.py",
		Reason: "unsupported language python",
		Locator: domain.Locator{
			Repo:      "repo-fe",
			CommitSHA: "1111111111111111111111111111111111111111",
			Path:      "scripts/deploy.py",
		},
	})

	return mm
}

func sendRequest(t *testing.T, s *mcp.Server, req map[string]any) map[string]any {
	reqData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var inBuf bytes.Buffer
	inBuf.Write(reqData)
	inBuf.WriteByte('\n')

	var outBuf bytes.Buffer
	if err := s.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response %q: %v", outBuf.String(), err)
	}

	return resp
}

func TestServer_InitializeAndPing(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	// initialize
	initResp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	})

	result, ok := initResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result in response, got %v", initResp)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", result["protocolVersion"])
	}

	// ping
	pingResp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "ping",
	})
	if pingResp["result"] == nil {
		t.Errorf("expected empty result object, got %v", pingResp)
	}

	// notifications/initialized (should produce no response)
	var inBuf bytes.Buffer
	inBuf.WriteString(`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n")
	var outBuf bytes.Buffer
	if err := server.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected no response for notification, got %q", outBuf.String())
	}
}

func TestServer_ToolsList(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	resp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/list",
	})

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map: %v", resp)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools list: %v", result)
	}

	if len(tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tAny := range tools {
		tDef := tAny.(map[string]any)
		name := tDef["name"].(string)
		toolNames[name] = true
	}

	expectedTools := []string{
		"get_impact_briefing",
		"get_context_slice",
		"explain_evidence",
		"locate_boundary",
	}

	for _, exp := range expectedTools {
		if !toolNames[exp] {
			t.Errorf("expected tool %s to be listed", exp)
		}
	}
}

func TestServer_ToolCalls(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	// 1. get_impact_briefing
	resp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      100,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_impact_briefing",
			"arguments": map[string]any{
				"target_id": "service:frontend",
				"depth":     1,
			},
		},
	})
	res := resp["result"].(map[string]any)
	content := res["content"].([]any)
	c0 := content[0].(map[string]any)
	text := c0["text"].(string)
	if !strings.Contains(text, "service:frontend") {
		t.Errorf("expected impact briefing text to contain target_id, got %s", text)
	}

	// 2. get_context_slice
	resp = sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      101,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_context_slice",
			"arguments": map[string]any{
				"node_id": "service:frontend",
				"tier":    "deep",
			},
		},
	})
	res = resp["result"].(map[string]any)
	content = res["content"].([]any)
	c0 = content[0].(map[string]any)
	text = c0["text"].(string)
	if !strings.Contains(text, "service:frontend") {
		t.Errorf("expected context slice to contain focal node, got %s", text)
	}

	// 3. explain_evidence
	resp = sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      102,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "explain_evidence",
			"arguments": map[string]any{
				"node_id": "service:frontend",
			},
		},
	})
	res = resp["result"].(map[string]any)
	content = res["content"].([]any)
	c0 = content[0].(map[string]any)
	text = c0["text"].(string)
	if !strings.Contains(text, "1111111111111111111111111111111111111111") {
		t.Errorf("expected explain evidence to contain commit SHA, got %s", text)
	}

	// 4. locate_boundary
	resp = sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      103,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "locate_boundary",
			"arguments": map[string]any{
				"query": "Login",
			},
		},
	})
	res = resp["result"].(map[string]any)
	content = res["content"].([]any)
	c0 = content[0].(map[string]any)
	text = c0["text"].(string)
	if !strings.Contains(text, "Login") {
		t.Errorf("expected locate boundary to find method Login, got %s", text)
	}
}

func TestServer_ToolCallErrors(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	// Unknown tool
	resp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      200,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "nonexistent_tool",
			"arguments": map[string]any{},
		},
	})
	res := resp["result"].(map[string]any)
	if res["isError"] != true {
		t.Errorf("expected isError to be true for unknown tool")
	}

	// Missing required arg target_id
	resp = sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      201,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "get_impact_briefing",
			"arguments": map[string]any{},
		},
	})
	res = resp["result"].(map[string]any)
	if res["isError"] != true {
		t.Errorf("expected isError to be true for missing target_id")
	}

	// Unknown method
	resp = sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      202,
		"method":  "some/unknown/method",
	})
	if resp["error"] == nil {
		t.Errorf("expected JSON-RPC error for unknown method")
	}
}

func TestServer_ContentLengthFraming(t *testing.T) {
	mm := buildTestMacroMap()
	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", "0.1.0", engine)

	msg := `{"jsonrpc":"2.0","id":42,"method":"ping"}`
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(msg), msg)

	var inBuf bytes.Buffer
	inBuf.WriteString(header)

	var outBuf bytes.Buffer
	if err := server.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode resp: %v", err)
	}
	if resp["id"] == nil || resp["id"].(float64) != 42 {
		t.Errorf("expected id 42, got %v", resp["id"])
	}
}
