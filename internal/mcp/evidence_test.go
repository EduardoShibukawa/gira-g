package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
)

func callTool(t *testing.T, name string, args map[string]any) (map[string]any, bool) {
	t.Helper()
	server := mcp.NewServer("gira", "0.1.0", nil)
	argData, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	resp := sendRequest(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": json.RawMessage(argData),
		},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result, got %v", resp)
	}
	content := result["content"].([]any)[0].(map[string]any)
	var out map[string]any
	if err := json.Unmarshal([]byte(content["text"].(string)), &out); err != nil {
		// Non-JSON text (e.g. error strings are still wrapped); surface it.
		return map[string]any{"_text": content["text"].(string)}, result["isError"] == true
	}
	return out, result["isError"] == true
}

// F4/F8: caller-supplied expansions are unmeasured input: mixed reports must
// not be labeled measured (quarantine); pure measurements stay measured.
func TestMCP_TokenAccountingHonesty(t *testing.T) {
	out, isErr := callTool(t, "get_token_accounting", map[string]any{"text": "hello", "expansions": 0})
	if isErr {
		t.Fatalf("pure measurement must succeed, got %v", out)
	}
	if out["measured_bool"] != true || out["net"] != out["gross"] {
		t.Fatalf("pure measurement must be measured with net==gross, got %v", out)
	}
	out, isErr = callTool(t, "get_token_accounting", map[string]any{"text": "hello", "expansions": 5})
	if isErr {
		t.Fatalf("mixed report must succeed (quarantined, not rejected), got %v", out)
	}
	if out["measured_bool"] != false {
		t.Fatalf("report with caller-supplied expansions must not be labeled measured, got %v", out)
	}
	if out["net"] != float64(out["gross"].(float64))-5 {
		t.Fatalf("net must still be gross-expansions, got %v", out)
	}
}

// F2/F4: inconsistent control claims are rejected, not echoed.
func TestMCP_ControlTotalConsistency(t *testing.T) {
	out, isErr := callTool(t, "get_control_total", map[string]any{
		"total": 1000, "matched": 48, "mismatched": 2, "seed": 42,
	})
	if isErr {
		t.Fatalf("consistent control line must succeed, got %v", out)
	}
	if out["sampled_n"] != float64(50) {
		t.Fatalf("sampled_n must follow max(20,5%%), got %v", out)
	}
	_, isErr = callTool(t, "get_control_total", map[string]any{
		"total": 1000, "matched": 40, "mismatched": 2, "seed": 42,
	})
	if !isErr {
		t.Fatalf("inconsistent control line (40+2 != 50) must be rejected")
	}
}

// F1/F2: verify_artifact shares the canonical preimage (taxonomy included).
func TestMCP_VerifyArtifactCanonical(t *testing.T) {
	out, isErr := callTool(t, "verify_artifact", map[string]any{
		"check": "c", "expect": "e", "cwd": "/r", "exit_code": 0,
		"fingerprint": "fp", "taxonomy": "measured",
	})
	if isErr || out["digest"] == nil || out["closing_line"] == nil {
		t.Fatalf("verify_artifact must return digest+closing_line, got %v err=%v", out, isErr)
	}
	out2, _ := callTool(t, "verify_artifact", map[string]any{
		"check": "c", "expect": "e", "cwd": "/r", "exit_code": 0,
		"fingerprint": "fp", "taxonomy": "inferred",
	})
	if out["digest"] == out2["digest"] {
		t.Fatalf("taxonomy must enter the digest")
	}
}
