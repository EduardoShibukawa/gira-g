package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SC5 (CLI surface): a corrupt versioned cache under the store root fails with
// a message that names the broken file and shows the --index escape — the
// discovery error an agent can act on (G3).
func TestSCIndexCorruptErrorCLI(t *testing.T) {
	const ver = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	xdg := t.TempDir()
	storeRoot := filepath.Join(xdg, "gira", "indexes")
	writeTree(t, storeRoot, map[string]string{
		"clitest/.head": ver,
		"clitest/versions/" + ver + "/meta.json": "{not json",
	})
	proj := t.TempDir()
	writeTree(t, proj, map[string]string{
		"gira.yaml": "version: 1\ntarget_system:\n  name: clitest\n  description: t\nrepositories:\n  - id: repo-core\n    url: \"https://example.invalid/repo-core.git\"\n    ref: v1.0.0\n    sha: \"1111111111111111111111111111111111111111\"\n    path_scope: \".\"\n",
	})
	t.Setenv("XDG_CACHE_HOME", xdg)
	t.Chdir(proj)

	var errOut strings.Builder
	err := Run([]string{"slice", "service:x"}, strings.NewReader(""), &strings.Builder{}, &errOut)
	if err == nil {
		t.Fatal("corrupt cache must fail discovery")
	}
	msg := err.Error()
	if !strings.Contains(msg, "index_corrupt") {
		t.Fatalf("structured code lost: %v", msg)
	}
	brokenDir := filepath.Join(storeRoot, "clitest", "versions", ver)
	if !strings.Contains(msg, brokenDir) || !strings.Contains(msg, ".json") {
		t.Fatalf("error must name an absolute broken cache file path, got: %v", msg)
	}
	if !strings.Contains(msg, "--index <path>") {
		t.Fatalf("error must suggest --index, got: %v", msg)
	}
}

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
