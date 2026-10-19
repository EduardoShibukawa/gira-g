package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// AC1, AC2: tools/index-bundle/bootstrap.sh integration test.
// Verifies that bootstrap.sh runs against gira, creates a bundle cache,
// and produces an offline-capable Schema 2 macro-map.json.
func TestConsumer_IndexBundleBootstrap(t *testing.T) {
	// Build a gira binary for testing bootstrap.sh
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "gira")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = "."
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed building test gira binary: %v\n%s", err, string(out))
	}

	// Setup git repo and manifest
	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/bundletest\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	manifestPath := filepath.Join(tmpDir, "gira.yaml")
	manifestContent := fmt.Sprintf(`version: 1
target_system:
  name: bundle-test
repositories:
  - id: bundle-repo
    url: "file://%s"
    sha: "%s"
    path_scope: .
`, filepath.ToSlash(repoDir), sha)
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	cacheDir := filepath.Join(tmpDir, "bundle-cache")
	outputPath := filepath.Join(tmpDir, "bundle-out", "macro-map.json")
	bootstrapScript := filepath.Join("..", "..", "tools", "index-bundle", "bootstrap.sh")

	// Run bootstrap.sh with GIRA_BIN set to test binary
	cmd := exec.Command(bootstrapScript,
		"--manifest", manifestPath,
		"--cache-dir", cacheDir,
		"--output", outputPath,
	)
	cmd.Env = append(os.Environ(), "GIRA_BIN="+binPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bootstrap.sh failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Verify output file exists and is valid Schema 2
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file at %s: %v", outputPath, err)
	}
	s := store.NewStore("")
	mm, err := s.LoadFile(outputPath)
	if err != nil {
		t.Fatalf("failed loading bootstrap index: %v", err)
	}
	if !store.IsSchema2(mm) {
		t.Errorf("expected Schema 2 index, got %v", mm.SchemaVersion)
	}
	if mm.SourceIdentity == nil {
		t.Errorf("expected SourceIdentity in bootstrap output")
	}

	// 2. Verify cache directory was populated
	entries, err := os.ReadDir(cacheDir)
	if err != nil || len(entries) == 0 {
		t.Errorf("expected non-empty cache directory at %s, err: %v", cacheDir, err)
	}

	// 3. Verify output text indicates success
	outStr := string(out)
	if !strings.Contains(outStr, "1/2: populating cache") || !strings.Contains(outStr, "2/2: verifying offline") {
		t.Errorf("expected two-phase progress in bootstrap output, got: %s", outStr)
	}
}

// AC1, AC2: gira-arch-diff skill workflow integration test.
// Simulates the exact workflow from internal/skills/catalog/gira-arch-diff/SKILL.md:
// 1. Index both sides: gira index ... --output base.json, gira index ... --output head.json
// 2. gira diff base.json head.json --format md|json and MCP diff_maps
func TestConsumer_GiraArchDiff(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	// Commit 1: service-a only
	repoDir, sha1 := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/archdiff\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	manifestBase := filepath.Join(tmpDir, "manifest-base.yaml")
	contentBase := fmt.Sprintf(`version: 1
target_system:
  name: archdiff-demo
repositories:
  - id: repo-a
    url: "file://%s"
    sha: "%s"
    path_scope: .
`, filepath.ToSlash(repoDir), sha1)
	if err := os.WriteFile(manifestBase, []byte(contentBase), 0644); err != nil {
		t.Fatal(err)
	}

	// Step 1a: gira index manifest-base.yaml --output base.json
	baseOutput := filepath.Join(tmpDir, "base.json")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", manifestBase, "--output", baseOutput, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("base indexing failed: %v, stderr: %s", err, stderr.String())
	}

	// Commit 2: Add contract/new file
	extraFile := filepath.Join(repoDir, "client.go")
	if err := os.WriteFile(extraFile, []byte("package main\n\nfunc CallClient() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", repoDir, "add", "client.go").Run()
	cmd := exec.Command("git", "-C", repoDir,
		"-c", "user.name=test", "-c", "user.email=test@example.com",
		"commit", "-m", "add client")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit failed: %v\n%s", err, string(out))
	}
	outSha, _ := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	sha2 := strings.TrimSpace(string(outSha))

	manifestHead := filepath.Join(tmpDir, "manifest-head.yaml")
	contentHead := fmt.Sprintf(`version: 1
target_system:
  name: archdiff-demo
repositories:
  - id: repo-a
    url: "file://%s"
    sha: "%s"
    path_scope: .
`, filepath.ToSlash(repoDir), sha2)
	if err := os.WriteFile(manifestHead, []byte(contentHead), 0644); err != nil {
		t.Fatal(err)
	}

	// Step 1b: gira index manifest-head.yaml --output head.json
	headOutput := filepath.Join(tmpDir, "head.json")
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestHead, "--output", headOutput, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("head indexing failed: %v, stderr: %s", err, stderr.String())
	}

	// Verify both snapshots are valid Schema 2
	s := store.NewStore("")
	mmBase, err := s.LoadFile(baseOutput)
	if err != nil || !store.IsSchema2(mmBase) {
		t.Fatalf("expected valid Schema 2 base index: %v", err)
	}
	mmHead, err := s.LoadFile(headOutput)
	if err != nil || !store.IsSchema2(mmHead) {
		t.Fatalf("expected valid Schema 2 head index: %v", err)
	}

	// Step 2a: gira diff base.json head.json --format json
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"diff", baseOutput, headOutput, "--format", "json"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira diff --format json failed: %v, stderr: %s", err, stderr.String())
	}
	var diffResult map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &diffResult); err != nil {
		t.Fatalf("failed unmarshaling diff JSON: %v, output: %s", err, stdout.String())
	}

	// Step 2b: gira diff base.json head.json --format md
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"diff", baseOutput, headOutput, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira diff --format md failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Architecture diff") {
		t.Errorf("expected markdown header in diff output, got: %s", stdout.String())
	}

	// Step 2c: MCP diff_maps tool call
	engine := slices.NewEngine(mmBase)
	server := mcp.NewServer("gira", version, engine)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "diff_maps",
			"arguments": map[string]any{
				"index_a": baseOutput,
				"index_b": headOutput,
			},
		},
	}
	reqBytes, _ := json.Marshal(req)
	var inBuf, outBuf bytes.Buffer
	inBuf.Write(reqBytes)
	inBuf.WriteString("\n")

	if err := server.Serve(&inBuf, &outBuf); err != nil {
		t.Fatalf("MCP server.Serve diff_maps failed: %v", err)
	}
	var mcpResp map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &mcpResp); err != nil {
		t.Fatalf("failed unmarshaling MCP diff response: %v", err)
	}
	resObj, ok := mcpResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'result' in MCP diff response, got: %s", outBuf.String())
	}
	if isErr, _ := resObj["isError"].(bool); isErr {
		t.Fatalf("MCP diff_maps returned error: %v", resObj)
	}
}
