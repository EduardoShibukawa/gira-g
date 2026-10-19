package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// AC7: Erros de load não são colapsados para index_missing em MCP/CLI.
func TestMCP_DiffMaps_LoadErrorsNotCollapsed(t *testing.T) {
	tempDir := t.TempDir()

	validMap := domain.NewMacroMapV2("demo", &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		Extractors:     []domain.ExtractorIdentity{{ID: "go", Version: "1.0.0"}},
		Sources:        []domain.SourceRepoIdentity{{RepoID: "repo", URLSHA256: "hash"}},
	})
	_ = validMap.AddNode(domain.Node{ID: "service:demo", Kind: domain.NodeService, Name: "demo"})
	validBytes, err := validMap.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}

	validPath := filepath.Join(tempDir, "valid.json")
	if err := os.WriteFile(validPath, validBytes, 0644); err != nil {
		t.Fatal(err)
	}

	corruptPath := filepath.Join(tempDir, "corrupt.json")
	if err := os.WriteFile(corruptPath, []byte(`{ "schema_version": 2, `), 0644); err != nil {
		t.Fatal(err)
	}

	unsupportedPath := filepath.Join(tempDir, "unsupported.json")
	if err := os.WriteFile(unsupportedPath, []byte(`{"schema_version": 3, "target_system": "foo"}`), 0644); err != nil {
		t.Fatal(err)
	}

	missingPath := filepath.Join(tempDir, "nonexistent.json")

	engine := slices.NewEngine(validMap)
	server := mcp.NewServer("gira", "0.1.0", engine)

	callDiffMaps := func(indexA, indexB string) (string, string) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "tools/call",
			"params": map[string]any{
				"name": "diff_maps",
				"arguments": map[string]any{
					"index_a": indexA,
					"index_b": indexB,
				},
			},
		}
		resp := sendRequest(t, server, req)
		if !isError(resp) {
			t.Fatalf("expected error for diff_maps with invalid input")
		}
		result := resp["result"].(map[string]any)
		content := result["content"].([]any)
		c0 := content[0].(map[string]any)
		text := c0["text"].(string)

		var structuredErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(text), &structuredErr); err != nil {
			t.Fatalf("error is not structured JSON: %s", text)
		}
		return structuredErr.Code, structuredErr.Message
	}

	t.Run("index_a_corrupt_returns_index_corrupt", func(t *testing.T) {
		code, msg := callDiffMaps(corruptPath, validPath)
		if code != "index_corrupt" {
			t.Errorf("expected code 'index_corrupt', got %q (msg: %s)", code, msg)
		}
	})

	t.Run("index_b_corrupt_returns_index_corrupt", func(t *testing.T) {
		code, msg := callDiffMaps(validPath, corruptPath)
		if code != "index_corrupt" {
			t.Errorf("expected code 'index_corrupt', got %q (msg: %s)", code, msg)
		}
	})

	t.Run("index_a_unsupported_returns_index_schema_unsupported", func(t *testing.T) {
		code, msg := callDiffMaps(unsupportedPath, validPath)
		if code != "index_schema_unsupported" {
			t.Errorf("expected code 'index_schema_unsupported', got %q (msg: %s)", code, msg)
		}
	})

	t.Run("index_b_unsupported_returns_index_schema_unsupported", func(t *testing.T) {
		code, msg := callDiffMaps(validPath, unsupportedPath)
		if code != "index_schema_unsupported" {
			t.Errorf("expected code 'index_schema_unsupported', got %q (msg: %s)", code, msg)
		}
	})

	t.Run("index_a_missing_returns_index_missing", func(t *testing.T) {
		code, msg := callDiffMaps(missingPath, validPath)
		if code != "index_missing" {
			t.Errorf("expected code 'index_missing', got %q (msg: %s)", code, msg)
		}
	})

	t.Run("index_b_missing_returns_index_missing", func(t *testing.T) {
		code, msg := callDiffMaps(validPath, missingPath)
		if code != "index_missing" {
			t.Errorf("expected code 'index_missing', got %q (msg: %s)", code, msg)
		}
	})
}
