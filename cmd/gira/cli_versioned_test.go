package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// AC1: index --output P não altera store/head/manifesto.
func TestAC1_IndexDetachedSnapshotDoesNotTouchStoreOrManifest(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	storeRoot := filepath.Join(tmpHome, ".cache", "gira", "indexes")

	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/ac1\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	cacheDir := t.TempDir()
	outDir := t.TempDir()
	outputPath := filepath.Join(outDir, "detached-snapshot.json")

	manifestDir := t.TempDir()
	manifestPath := filepath.Join(manifestDir, "gira.yaml")
	manifestContent := fmt.Sprintf(`version: 1
target_system:
  name: ac1-sys
repositories:
  - id: local
    url: "file://%s"
    sha: "%s"
    path_scope: .
`, filepath.ToSlash(repoDir), sha)
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", manifestPath, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira index --output failed: %v, stderr: %s", err, stderr.String())
	}

	// 1. Snapshot file P exists and is valid schema 2
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected snapshot at %s: %v", outputPath, err)
	}
	s := store.NewStore("")
	mm, err := s.LoadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to load detached snapshot: %v", err)
	}
	if !store.IsSchema2(mm) {
		t.Errorf("expected Schema 2 for detached snapshot, got %v", mm.SchemaVersion)
	}
	if mm.SourceIdentity == nil {
		t.Errorf("expected embedded SourceIdentity in detached snapshot")
	}

	// 2. Store root was NOT modified: no versions directory, no .head
	headPath := filepath.Join(storeRoot, "ac1-sys", ".head")
	if _, err := os.Stat(headPath); !os.IsNotExist(err) {
		t.Errorf("expected .head NOT to exist in store root, got stat err: %v", err)
	}
	versionsDir := filepath.Join(storeRoot, "ac1-sys", "versions")
	if _, err := os.Stat(versionsDir); !os.IsNotExist(err) {
		t.Errorf("expected versions dir NOT to exist in store root, got stat err: %v", err)
	}

	// 3. Manifest was NOT modified
	rawAfter, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(rawAfter) != manifestContent {
		t.Errorf("index --output must not modify the manifest")
	}

	// 4. Output notice indicates P is a detached snapshot
	if !strings.Contains(stderr.String(), "detached snapshot") {
		t.Errorf("expected notice about detached snapshot in stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Indexed") {
		t.Errorf("expected Indexed in stdout, got: %s", stdout.String())
	}
}

// AC2: init --output P publica canônico e P com bytes idênticos; manifesto é gravado por último.
func TestAC2_InitOutputPublishesCanonicalAndIdenticalSnapshotManifestLast(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	storeRoot := filepath.Join(tmpHome, ".cache", "gira", "indexes")

	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/ac2\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	cacheDir := t.TempDir()
	outDir := t.TempDir()
	outputPath := filepath.Join(outDir, "snapshot.json")
	manifestPath := filepath.Join(repoDir, "gira.yaml")

	var stdout, stderr bytes.Buffer
	err := Run([]string{"init", repoDir, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira init --output failed: %v, stderr: %s", err, stderr.String())
	}

	// 1. Canonical version published in store
	targetSystem := filepath.Base(repoDir)
	s := store.NewStore(storeRoot)
	headVer, err := s.ReadHead(targetSystem)
	if err != nil {
		t.Fatalf("expected active .head in store: %v", err)
	}
	if headVer == "" {
		t.Fatalf("expected non-empty head version")
	}

	canonicalPath := filepath.Join(storeRoot, targetSystem, "versions", headVer, "macro-map.json")
	canonicalBytes, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("expected canonical macro-map.json at %s: %v", canonicalPath, err)
	}

	// 2. Snapshot P has byte-for-byte identical content
	snapshotBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected snapshot at %s: %v", outputPath, err)
	}
	if !bytes.Equal(canonicalBytes, snapshotBytes) {
		t.Errorf("canonical bytes and snapshot bytes differ!\ncanonical SHA: %x\nsnapshot SHA: %x",
			sha256.Sum256(canonicalBytes), sha256.Sum256(snapshotBytes))
	}

	// 3. Manifest was written and pins HEAD
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected gira.yaml at %s: %v", manifestPath, err)
	}
	m, err := manifest.Parse(rawManifest)
	if err != nil {
		t.Fatalf("failed parsing gira.yaml: %v", err)
	}
	if len(m.Repositories) != 1 || m.Repositories[0].SHA != sha {
		t.Errorf("manifest repository SHA mismatch: got %v, want %s", m.Repositories, sha)
	}

	// 4. Resolver finds version without --index
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origCwd) })
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	resolved, err := loadIndex("")
	if err != nil {
		t.Fatalf("loadIndex without --index failed: %v", err)
	}
	if resolved.TargetSystem != targetSystem {
		t.Errorf("expected target system %q, got %q", targetSystem, resolved.TargetSystem)
	}
}

// AC3: Falha de build/gate/sync não cria manifesto novo nem altera o existente.
func TestAC3_InitFailureDoesNotCreateOrModifyManifest(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/ac3\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})
	manifestPath := filepath.Join(repoDir, "gira.yaml")

	// Case 3a: Build failure on fresh init -> gira.yaml NOT created, snapshot NOT created
	snapPath := filepath.Join(t.TempDir(), "snap.json")
	initHookBeforeBuild = func() error {
		return fmt.Errorf("injected build failure")
	}
	t.Cleanup(func() { initHookBeforeBuild = nil })

	var stdout, stderr bytes.Buffer
	err := Run([]string{"init", repoDir, "--output", snapPath}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected init to fail on injected build error")
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Errorf("gira.yaml must NOT be created when build fails: %v", err)
	}
	if _, err := os.Stat(snapPath); !os.IsNotExist(err) {
		t.Errorf("snapshot must NOT be created when build fails: %v", err)
	}
	initHookBeforeBuild = nil

	// Successful init to establish existing manifest
	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"init", repoDir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("initial init failed: %v", err)
	}
	manifestBefore, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	// Case 3b: Failure before commit on existing manifest -> gira.yaml NOT modified
	initHookBeforeCommit = func() error {
		return fmt.Errorf("injected gate/commit failure")
	}
	t.Cleanup(func() { initHookBeforeCommit = nil })

	// Add new commit so --force would repin if successful
	if err := os.WriteFile(filepath.Join(repoDir, "extra.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", repoDir, "add", ".")
	cmd.Env = manifest.GitCommandEnv()
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "second commit")
	cmd.Env = manifest.GitCommandEnv()
	_ = cmd.Run()

	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"init", repoDir, "--force"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected init --force to fail on injected commit error")
	}

	manifestAfter, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(manifestBefore, manifestAfter) {
		t.Errorf("existing gira.yaml must NOT be modified when failure occurs:\nbefore: %s\nafter: %s",
			string(manifestBefore), string(manifestAfter))
	}
	_ = sha
}

func createShortRepoWithFiles(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	repoDir, err := os.MkdirTemp(os.TempDir(), "g-r-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(repoDir) })

	runGit := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v, out: %s", args, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}

	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")

	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", relPath, err)
		}
	}

	runGit("add", ".")
	runGit("commit", "-m", "Initial commit")
	commitSHA := runGit("rev-parse", "HEAD")
	return repoDir, commitSHA
}

// AC4: Alteração concorrente do manifesto durante init gera conflito sem publicar estado misto.
func TestAC4_InitConcurrentManifestModificationGeneratesConflict(t *testing.T) {
	tmpHome, err := os.MkdirTemp(os.TempDir(), "g-h-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpHome) })
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	repoDir, _ := createShortRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/ac4\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})
	manifestPath := filepath.Join(repoDir, "gira.yaml")

	// Case 4a: Manifest created externally during init of new repo
	initHookBeforeBuild = func() error {
		// External process writes gira.yaml while init is running
		return os.WriteFile(manifestPath, []byte("version: 1\ntarget_system:\n  name: injected\n"), 0644)
	}
	t.Cleanup(func() { initHookBeforeBuild = nil })

	var stdout, stderr bytes.Buffer
	err := Run([]string{"init", repoDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected init to fail on concurrent manifest creation")
	}
	if !strings.Contains(err.Error(), "manifest conflict") {
		t.Errorf("expected error to cite 'manifest conflict', got: %v", err)
	}
	// Verify external content was preserved
	raw, err := os.ReadFile(manifestPath)
	if err != nil || !strings.Contains(string(raw), "name: injected") {
		t.Errorf("expected external manifest content to be preserved, got: %s", string(raw))
	}

	// Clean up for Case 4b
	initHookBeforeBuild = nil
	_ = os.Remove(manifestPath)

	// Run successful init
	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"init", repoDir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Case 4b: Existing manifest modified externally during init --force
	initHookBeforePublish = func() error {
		return os.WriteFile(manifestPath, []byte("version: 1\ntarget_system:\n  name: altered-externally\n"), 0644)
	}
	t.Cleanup(func() { initHookBeforePublish = nil })

	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"init", repoDir, "--force"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected init to fail on concurrent manifest alteration")
	}
	if !strings.Contains(err.Error(), "manifest conflict") {
		t.Errorf("expected error to cite 'manifest conflict', got: %v", err)
	}
	raw, err = os.ReadFile(manifestPath)
	if err != nil || !strings.Contains(string(raw), "name: altered-externally") {
		t.Errorf("expected altered content to be preserved, got: %s", string(raw))
	}
}

// AC5: eval sem manifesto real retorna manifest_path_required quando overrides exigem locator.
func TestAC5_EvalWithoutManifestPathReturnsManifestPathRequired(t *testing.T) {
	corpusDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(corpusDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c := eval.Corpus{
		Dir:          corpusDir,
		RepoID:       "repo-1",
		TargetSystem: "sys",
		Manifest: &manifest.Manifest{
			Version:      1,
			TargetSystem: manifest.TargetSystemMeta{Name: "sys"},
			Overrides: &manifest.OverridesDeclaration{
				Relationships: []manifest.RelationshipOverride{
					{SourceID: "s1", TargetID: "s2", Verb: "consumes"},
				},
			},
		},
		ManifestPath: "",
	}
	_, err := c.Index(context.Background(), t.TempDir(), nil)
	if err == nil {
		t.Fatalf("expected error when overrides present but ManifestPath empty")
	}
	if !errors.Is(err, manifest.ErrManifestPathRequired) {
		t.Errorf("expected ErrManifestPathRequired, got %v", err)
	}
	if !strings.Contains(err.Error(), "manifest_path_required") {
		t.Errorf("expected error to contain 'manifest_path_required', got: %v", err)
	}
}

// AC6: skills install/check --index avisa para legado e normaliza caminho relativo para absoluto.
func TestAC6_SkillsInstallAndCheckLegacyWarningAndAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	projDir := filepath.Join(tempDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create legacy index at a subpath
	indicesDir := filepath.Join(tempDir, "indexes")
	if err := os.MkdirAll(indicesDir, 0755); err != nil {
		t.Fatal(err)
	}
	legacyFile := filepath.Join(indicesDir, "legacy.json")
	legacyContent := `{"schema_version": "1.0.0", "target_system": "legacy-sys", "nodes": [{"id": "service:s", "kind": "Service", "name": "s"}]}`
	if err := os.WriteFile(legacyFile, []byte(legacyContent), 0644); err != nil {
		t.Fatal(err)
	}

	origCwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origCwd) })
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	// 1. skills install with relative path to legacy index
	relLegacy := filepath.Join("indexes", "legacy.json")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"skills", "install", "--host", "claude", "--dir", projDir, "--index", relLegacy}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("skills install failed: %v", err)
	}

	// Must emit legacy warning
	expectedWarning := "warning: legacy index format (schema 1.0.0) is deprecated and read-only; re-index with 'gira index' to upgrade to schema 2"
	if !strings.Contains(stderr.String(), expectedWarning) {
		t.Errorf("expected legacy warning in stderr, got: %s", stderr.String())
	}

	// Must normalize relative path to absolute in .mcp.json
	mcpPath := filepath.Join(projDir, ".mcp.json")
	mcpData, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("failed reading .mcp.json: %v", err)
	}
	absLegacy, _ := filepath.Abs(relLegacy)
	var mcpCfg struct {
		McpServers map[string]struct {
			Args []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(mcpData, &mcpCfg); err != nil {
		t.Fatalf("failed unmarshaling .mcp.json: %v", err)
	}
	found := false
	for _, arg := range mcpCfg.McpServers["gira"].Args {
		if filepath.Clean(arg) == filepath.Clean(absLegacy) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected absolute path %q in .mcp.json args, got: %s", absLegacy, string(mcpData))
	}

	// 2. skills check with relative path to legacy index
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"skills", "check", "--host", "claude", "--dir", projDir, "--index", relLegacy}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("skills check failed: %v", err)
	}
	if !strings.Contains(stderr.String(), expectedWarning) {
		t.Errorf("expected legacy warning in skills check stderr, got: %s", stderr.String())
	}

	// 3. skills install with schema 2 index: no warning emitted
	v2File := filepath.Join(indicesDir, "v2.json")
	si := &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Extractors:     []domain.ExtractorIdentity{{ID: "go", Version: "1.0.0"}},
		Sources:        []domain.SourceRepoIdentity{{RepoID: "repo-1", URLSHA256: "hash-repo-1"}},
	}
	mmV2 := domain.NewMacroMapV2("v2-sys", si)
	v2Bytes, _ := mmV2.CanonicalJSON()
	if err := os.WriteFile(v2File, v2Bytes, 0644); err != nil {
		t.Fatal(err)
	}

	relV2 := filepath.Join("indexes", "v2.json")
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"skills", "install", "--host", "cursor", "--dir", projDir, "--index", relV2}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("skills install with v2 failed: %v", err)
	}
	if strings.Contains(stderr.String(), "legacy index format") {
		t.Errorf("unexpected legacy warning for schema 2 index: %s", stderr.String())
	}
	cursorMcpPath := filepath.Join(projDir, ".cursor", "mcp.json")
	cursorMcpData, err := os.ReadFile(cursorMcpPath)
	if err != nil {
		t.Fatalf("failed reading cursor mcp.json: %v", err)
	}
	absV2, _ := filepath.Abs(relV2)
	if !strings.Contains(string(cursorMcpData), absV2) {
		t.Errorf("expected absolute path %q in cursor mcp.json, got: %s", absV2, string(cursorMcpData))
	}
}

// AC7: Mensagens refletem publicação parcial/visible_unconfirmed corretamente.
func TestAC7_InitPartialPublicationAndDurability(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	storeRoot := filepath.Join(tmpHome, ".cache", "gira", "indexes")

	repoDir, _ := createGitRepoWithFiles(t, map[string]string{
		"go.mod":  "module example.com/ac7\n\ngo 1.24\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})
	manifestPath := filepath.Join(repoDir, "gira.yaml")

	// 1. Partial publication: canonical succeeds, but snapshot copy fails because destination is a directory
	badSnapshotPath := t.TempDir() // a directory cannot be overwritten by a file
	var stdout, stderr bytes.Buffer
	err := Run([]string{"init", repoDir, "--output", badSnapshotPath}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected partial publication error when snapshot copy fails")
	}
	if !strings.Contains(err.Error(), "partial publication") {
		t.Errorf("expected error to cite 'partial publication', got: %v", err)
	}

	// Canonical version WAS published
	targetSystem := filepath.Base(repoDir)
	s := store.NewStore(storeRoot)
	headVer, err := s.ReadHead(targetSystem)
	if err != nil || headVer == "" {
		t.Errorf("expected canonical version to be published despite snapshot copy failure: %v", err)
	}

	// gira.yaml was NOT created because partial publication is an error
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Errorf("gira.yaml must NOT be created when publication is partial")
	}

	// 2. Durability: visible_unconfirmed on Windows or directory sync error
	if runtime.GOOS == "windows" {
		stdout.Reset()
		stderr.Reset()
		goodSnapshot := filepath.Join(t.TempDir(), "good.json")
		err := Run([]string{"init", repoDir, "--output", goodSnapshot}, nil, &stdout, &stderr)
		if err != nil {
			t.Fatalf("expected init on Windows to succeed with exit 0: %v", err)
		}
		if !strings.Contains(stderr.String(), "visible_unconfirmed") {
			t.Errorf("expected visible_unconfirmed warning on Windows stderr, got: %s", stderr.String())
		}
	}
}
