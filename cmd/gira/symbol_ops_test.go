package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLISymbolReadWriteAdd(t *testing.T) {
	tsSource := `export class TestService {
  execute(): string {
    return "original";
  }
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "service.ts")
	if err := os.WriteFile(filePath, []byte(tsSource), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Test read
	var stdout, stderr bytes.Buffer
	err := runReadSymbol([]string{filePath, "TestService.execute"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runReadSymbol failed: %v, stderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "execute(): string") {
		t.Errorf("read output missing method signature: %s", out)
	}

	// 2. Test write
	stdout.Reset()
	stderr.Reset()
	newBody := `  execute(): string {
    return "updated";
  }`
	err = runWriteSymbol([]string{filePath, "TestService.execute", "--body", newBody}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runWriteSymbol failed: %v, stderr: %s", err, stderr.String())
	}

	// Verify write took effect
	content, _ := os.ReadFile(filePath)
	if !strings.Contains(string(content), `"updated"`) {
		t.Errorf("file was not updated with new body: %s", string(content))
	}

	// 3. Test add
	stdout.Reset()
	stderr.Reset()
	newMember := `ping(): boolean {
  return true;
}`
	err = runAddSymbol([]string{filePath, "TestService", "--body", newMember}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runAddSymbol failed: %v, stderr: %s", err, stderr.String())
	}

	content, _ = os.ReadFile(filePath)
	if !strings.Contains(string(content), "ping(): boolean") {
		t.Errorf("file missing added member: %s", string(content))
	}
}

func TestCLISymbolOmitBodyReturnsError(t *testing.T) {
	tsSource := `export class TestService {
  execute(): string {
    return "original";
  }
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "service.ts")
	if err := os.WriteFile(filePath, []byte(tsSource), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runWriteSymbol([]string{filePath, "TestService.execute"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when --body is omitted, got nil")
	}
	if !strings.Contains(err.Error(), "cannot write empty symbol body") {
		t.Errorf("unexpected error message: %v", err)
	}

	err = runAddSymbol([]string{filePath, "TestService"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when --body is omitted in add, got nil")
	}
	if !strings.Contains(err.Error(), "cannot add empty member body") {
		t.Errorf("unexpected error message: %v", err)
	}
}
