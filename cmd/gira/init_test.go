package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// SCSingleInit: `gira init` gera gira.yaml pinando o HEAD e indexa; segunda
// chamada sem --force reaproveita o manifest; --force repina após novo commit.
func TestSCSingleInit(t *testing.T) {
	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/acme\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	cacheDir := t.TempDir()
	outDir := t.TempDir()
	outputPath := filepath.Join(outDir, "macro-map.json")
	manifestPath := filepath.Join(repoDir, "gira.yaml")

	// 1. init gera manifest + indexa.
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"init", repoDir, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira init failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Initialized") {
		t.Errorf("expected Initialized in stdout, got %q", stdout.String())
	}

	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected gira.yaml to be created: %v", err)
	}
	m, err := manifest.Parse(raw)
	if err != nil {
		t.Fatalf("generated gira.yaml does not parse: %v", err)
	}
	if len(m.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(m.Repositories))
	}
	if m.Repositories[0].SHA != sha {
		t.Errorf("expected pinned SHA %s, got %s", sha, m.Repositories[0].SHA)
	}

	s := store.NewStore("")
	mm, err := s.LoadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to load indexed macro map: %v", err)
	}
	if mm.TargetSystem != filepath.Base(repoDir) {
		t.Errorf("expected target system %q, got %q", filepath.Base(repoDir), mm.TargetSystem)
	}
	if len(mm.Nodes) == 0 {
		t.Errorf("expected at least one node in single-service index")
	}

	// 2. Segundo init sem --force reaproveita o manifest e reindexa.
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read gira.yaml: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"init", repoDir, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("second gira init failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Using existing") {
		t.Errorf("expected Using existing in stdout, got %q", stdout.String())
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to re-read gira.yaml: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("init without --force must not rewrite an existing manifest")
	}

	// 3. Novo commit + --force repina o SHA.
	if err := os.WriteFile(filepath.Join(repoDir, "extra.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write extra.go: %v", err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "second"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v, out: %s", args, err, string(out))
		}
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD failed: %v", err)
	}
	newSHA := strings.TrimSpace(string(out))

	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"init", repoDir, "--force", "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira init --force failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Repinned") {
		t.Errorf("expected Repinned in stdout, got %q", stdout.String())
	}
	repinned, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read repinned gira.yaml: %v", err)
	}
	rm, err := manifest.Parse(repinned)
	if err != nil {
		t.Fatalf("repinned gira.yaml does not parse: %v", err)
	}
	if rm.Repositories[0].SHA != newSHA {
		t.Errorf("expected repinned SHA %s, got %s", newSHA, rm.Repositories[0].SHA)
	}

	// 4. Fora de git repo: erro honesto.
	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"init", t.TempDir(), "--output", filepath.Join(t.TempDir(), "m.json")}, nil, &stdout, &stderr); err == nil {
		t.Errorf("expected error for non-git directory, got nil")
	}
}
