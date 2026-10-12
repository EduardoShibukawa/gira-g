package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"labs/gira-g/internal/domain"
	"labs/gira-g/internal/store"
)

func createTestIndex(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "macro-map.json")

	mm := domain.NewMacroMap("test-corpus")
	_ = mm.AddNode(domain.Node{
		ID:   "service:order",
		Kind: domain.NodeService,
		Name: "order-service",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-order",
					CommitSHA: "abc1112223334445556667778889990001112223",
					Path:      "main.go",
					StartLine: 1,
					EndLine:   50,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "default",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      time.Now(),
			},
		},
	})
	_ = mm.AddNode(domain.Node{
		ID:   "contract:order-grpc",
		Kind: domain.NodeContract,
		Name: "OrderService",
		Metadata: map[string]any{
			"methods": []string{"PlaceOrder", "GetOrder"},
		},
	})
	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:order-provides-grpc",
		SourceID: "service:order",
		TargetID: "contract:order-grpc",
		Verb:     domain.VerbProvides,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-order",
					CommitSHA: "abc1112223334445556667778889990001112223",
					Path:      "service.go",
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})
	mm.Unsupported = append(mm.Unsupported, domain.UnsupportedBoundary{
		Path:   "legacy/billing.js",
		Reason: "unsupported javascript",
	})

	s := store.NewStore("")
	data, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to serialize test map: %v", err)
	}
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		t.Fatalf("failed to write test index: %v", err)
	}
	_ = s

	return indexPath
}

func TestCLI_VersionAndHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// version
	if err := Run([]string{"version"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "gira ") {
		t.Errorf("expected version output, got %q", stdout.String())
	}

	// help
	stdout.Reset()
	if err := Run([]string{"--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage output, got %q", stdout.String())
	}

	// no args
	stdout.Reset()
	if err := Run([]string{}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("no args failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage output, got %q", stdout.String())
	}
}

func TestCLI_Impact(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format (default)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"impact", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("impact json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"target_id": "service:order"`) {
		t.Errorf("expected target_id in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"impact", "service:order", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("impact md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Impact Briefing: service:order") {
		t.Errorf("expected markdown header in output, got %s", stdout.String())
	}
}

func TestCLI_Slice(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format (default)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"slice", "service:order", "--tier", "contracts", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("slice json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"focal_node"`) {
		t.Errorf("expected focal_node in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"slice", "service:order", "--tier", "contracts", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("slice md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Context Slice: order-service") {
		t.Errorf("expected markdown header in output, got %s", stdout.String())
	}
}

func TestCLI_Explain(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format for node
	var stdout, stderr bytes.Buffer
	err := Run([]string{"explain", "--node", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("explain json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "abc1112223334445556667778889990001112223") {
		t.Errorf("expected locator SHA in explain output, got %s", stdout.String())
	}

	// Markdown format for edge
	stdout.Reset()
	err = Run([]string{"explain", "--edge", "rel:order-provides-grpc", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("explain edge md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Evidence Report: rel:order-provides-grpc") {
		t.Errorf("expected markdown report in output, got %s", stdout.String())
	}

	// Positional node ID fallback: gira explain service:order
	stdout.Reset()
	err = Run([]string{"explain", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("positional explain failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"node_id": "service:order"`) {
		t.Errorf("expected node_id in output, got %s", stdout.String())
	}
}

func TestCLI_Locate(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format
	var stdout, stderr bytes.Buffer
	err := Run([]string{"locate", "billing", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("locate json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "legacy/billing.js") {
		t.Errorf("expected unsupported match in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"locate", "PlaceOrder", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("locate md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "RPC Method `PlaceOrder`") {
		t.Errorf("expected method match in md output, got %s", stdout.String())
	}
}

func TestCLI_ServeMCP(t *testing.T) {
	indexPath := createTestIndex(t)

	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdin.WriteString(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")

	err := Run([]string{"serve", "mcp", "--index", indexPath}, &stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("serve mcp failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"id":1`) {
		t.Errorf("expected ping response, got %s", stdout.String())
	}
}

func TestCLI_ErrorHandling(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Unknown command
	if err := Run([]string{"unknown_cmd"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for unknown command, got nil")
	}

	// Missing arguments
	if err := Run([]string{"impact"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing impact target, got nil")
	}

	if err := Run([]string{"slice"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing slice node, got nil")
	}

	if err := Run([]string{"locate"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing locate query, got nil")
	}

	if err := Run([]string{"explain"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing explain args, got nil")
	}

	// Nonexistent index file
	if err := Run([]string{"impact", "service:order", "--index", "/nonexistent/index.json"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for nonexistent index file, got nil")
	}

	// Invalid export format
	if err := Run([]string{"export", "--format", "invalid_fmt"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for invalid export format, got nil")
	}
}

func TestCLI_Export(t *testing.T) {
	indexPath := createTestIndex(t)

	// 1. Export Markdown to stdout (default format md)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"export", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export md failed: %v", err)
	}
	outMD := stdout.String()
	if !strings.Contains(outMD, "# Target System Architecture: test-corpus") {
		t.Errorf("expected Target System Architecture in export md output, got:\n%s", outMD)
	}
	if !strings.Contains(outMD, "## Services (1)") {
		t.Errorf("expected Services (1) in export md output")
	}
	if !strings.Contains(outMD, "## Unsupported Boundaries & Known Unknowns (1)") {
		t.Errorf("expected Unsupported Boundaries section in export md output")
	}

	// 2. Export HTML to stdout
	stdout.Reset()
	err = Run([]string{"export", "--format", "html", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export html failed: %v", err)
	}
	outHTML := stdout.String()
	if !strings.Contains(outHTML, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in export html output")
	}
	if !strings.Contains(outHTML, "<svg viewBox=") {
		t.Errorf("expected SVG in export html output")
	}
	if !strings.Contains(outHTML, "badge-extracted") {
		t.Errorf("expected origin badge in export html output")
	}

	// 3. Export to file via --output
	tmpOutFile := filepath.Join(t.TempDir(), "sub", "export.html")
	stdout.Reset()
	err = Run([]string{"export", "--format", "html", "--output", tmpOutFile, "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export --output failed: %v", err)
	}
	data, err := os.ReadFile(tmpOutFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in written export file")
	}
}

func TestCLI_View(t *testing.T) {
	indexPath := createTestIndex(t)

	// 1. View with default temporary directory
	var stdout, stderr bytes.Buffer
	err := Run([]string{"view", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("view failed: %v", err)
	}
	viewOut := stdout.String()
	if !strings.Contains(viewOut, "Exported HTML view to: file://") {
		t.Errorf("expected file:// URI in view output, got: %s", viewOut)
	}

	// 2. View with specified --output directory
	customDir := filepath.Join(t.TempDir(), "custom-view")
	stdout.Reset()
	err = Run([]string{"view", "--output", customDir, "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("view with custom output dir failed: %v", err)
	}
	expectedFile := filepath.Join(customDir, "index.html")
	if fi, err := os.Stat(expectedFile); err != nil || fi.IsDir() {
		t.Fatalf("expected view index.html at %s", expectedFile)
	}
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read view index.html: %v", err)
	}
	if !strings.Contains(string(data), "Gira-G Architecture Export: test-corpus") {
		t.Errorf("expected title in view index.html")
	}
}
