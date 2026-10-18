package gitcache_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// helper to init a local git repo with a commit
func createTestLocalRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()

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

	// Create root file
	rootFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(rootFile, []byte("# Hello World\n"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	// Create nested directory and file
	subDir := filepath.Join(repoDir, "pkg", "service")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	subFile := filepath.Join(subDir, "main.go")
	if err := os.WriteFile(subFile, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write subFile: %v", err)
	}

	runGit("add", ".")
	runGit("commit", "-m", "Initial commit")

	commitSHA := runGit("rev-parse", "HEAD")
	if len(commitSHA) != 40 {
		t.Fatalf("expected 40-char SHA, got %s", commitSHA)
	}

	return repoDir, commitSHA
}

func TestNormalizeURLAndRepoHash(t *testing.T) {
	url1 := "https://github.com/open-telemetry/opentelemetry-demo.git"
	url2 := "https://github.com/open-telemetry/opentelemetry-demo.git/"

	hash1 := gitcache.RepoHash(url1)
	hash2 := gitcache.RepoHash(url2)

	if hash1 != hash2 {
		t.Errorf("RepoHash should normalize trailing slashes: %s != %s", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("RepoHash should be 64-char sha256 hex digest, got %d", len(hash1))
	}

	fileURL1 := "file:///tmp/repo/demo/"
	fileURL2 := "file:///tmp/repo/demo"
	if gitcache.RepoHash(fileURL1) != gitcache.RepoHash(fileURL2) {
		t.Errorf("RepoHash should normalize file URLs: %s != %s", gitcache.RepoHash(fileURL1), gitcache.RepoHash(fileURL2))
	}
}

func TestCacheManager_New(t *testing.T) {
	cm, err := gitcache.NewCacheManager("/tmp/custom-cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cm.BaseDir != "/tmp/custom-cache" {
		t.Errorf("expected BaseDir /tmp/custom-cache, got %s", cm.BaseDir)
	}

	// Test default resolution
	t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
	cmXDG, err := gitcache.NewCacheManager("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedXDG := filepath.Join("/tmp/xdg-cache", "gira", "repos")
	if cmXDG.BaseDir != expectedXDG {
		t.Errorf("expected XDG BaseDir %s, got %s", expectedXDG, cmXDG.BaseDir)
	}
}

func TestCacheManager_EnsureRepo_OfflineAndOnline(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		t.Fatalf("failed to init CacheManager: %v", err)
	}

	localRepoDir, commitSHA := createTestLocalRepo(t)
	localURL := "file://" + localRepoDir

	repoDecl := manifest.RepositoryDeclaration{
		ID:        "test-repo",
		URL:       localURL,
		SHA:       commitSHA,
		PathScope: ".",
	}

	// 1. Offline mode before clone: should fail with ErrRepoNotFound
	err = cm.EnsureRepo(ctx, repoDecl, true)
	if err == nil {
		t.Fatalf("expected error in offline mode when repo is not cached, got nil")
	}
	var errNotFound gitcache.ErrRepoNotFound
	if !errors.As(err, &errNotFound) {
		t.Errorf("expected ErrRepoNotFound, got %T: %v", err, err)
	}

	// 2. Online mode: should clone bare repository successfully
	err = cm.EnsureRepo(ctx, repoDecl, false)
	if err != nil {
		t.Fatalf("failed to clone repo: %v", err)
	}

	// Verify repo directory exists
	repoBareDir := cm.RepoDir(repoDecl)
	if fi, err := os.Stat(repoBareDir); err != nil || !fi.IsDir() {
		t.Fatalf("expected bare repo directory at %s", repoBareDir)
	}

	// 3. Offline mode now that repo is cloned: should succeed
	err = cm.EnsureRepo(ctx, repoDecl, true)
	if err != nil {
		t.Fatalf("expected offline mode to succeed when commit is cached: %v", err)
	}

	// 4. Missing commit in offline mode: should fail with ErrCommitMissing
	missingDecl := repoDecl
	missingDecl.SHA = "0000000000000000000000000000000000000000"
	err = cm.EnsureRepo(ctx, missingDecl, true)
	if err == nil {
		t.Fatalf("expected ErrCommitMissing, got nil")
	}
	var errMissing gitcache.ErrCommitMissing
	if !errors.As(err, &errMissing) {
		t.Errorf("expected ErrCommitMissing, got %T: %v", err, err)
	}

	// 5. Unreachable remote handling
	unreachableDecl := manifest.RepositoryDeclaration{
		ID:  "unreachable-repo",
		URL: "file:///non/existent/path/for/sure",
		SHA: commitSHA,
	}
	err = cm.EnsureRepo(ctx, unreachableDecl, false)
	if err == nil {
		t.Fatalf("expected UnreachableError for non-existent remote, got nil")
	}
	var errUnreachable gitcache.UnreachableError
	if !errors.As(err, &errUnreachable) {
		t.Errorf("expected UnreachableError, got %T: %v", err, err)
	}
}

func TestCacheManager_ListTree_And_ReadFile(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		t.Fatalf("failed to init CacheManager: %v", err)
	}

	localRepoDir, commitSHA := createTestLocalRepo(t)
	localURL := "file://" + localRepoDir

	repoDecl := manifest.RepositoryDeclaration{
		ID:        "test-repo",
		URL:       localURL,
		SHA:       commitSHA,
		PathScope: ".",
	}

	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("failed to ensure repo: %v", err)
	}

	// List entire tree
	entries, err := cm.ListTree(ctx, repoDecl)
	if err != nil {
		t.Fatalf("failed to list tree: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 files in tree, got %d", len(entries))
	}

	paths := make(map[string]bool)
	for _, entry := range entries {
		paths[entry.Path] = true
	}
	if !paths["README.md"] || !paths["pkg/service/main.go"] {
		t.Errorf("unexpected files listed: %v", paths)
	}

	// List scoped tree
	scopedDecl := repoDecl
	scopedDecl.PathScope = "pkg/service"
	scopedEntries, err := cm.ListTree(ctx, scopedDecl)
	if err != nil {
		t.Fatalf("failed to list scoped tree: %v", err)
	}
	if len(scopedEntries) != 1 {
		t.Fatalf("expected 1 file in scoped tree, got %d", len(scopedEntries))
	}
	if scopedEntries[0].Path != "pkg/service/main.go" {
		t.Errorf("expected pkg/service/main.go, got %s", scopedEntries[0].Path)
	}

	// Read file contents
	content, err := cm.ReadFile(ctx, repoDecl, "README.md")
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	if string(content) != "# Hello World\n" {
		t.Errorf("expected content '# Hello World\\n', got %q", string(content))
	}

	// Read nested file
	mainContent, err := cm.ReadFile(ctx, repoDecl, "pkg/service/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if !strings.Contains(string(mainContent), "package main") {
		t.Errorf("unexpected content in main.go: %s", string(mainContent))
	}

	// Read non-existent file
	_, err = cm.ReadFile(ctx, repoDecl, "non-existent.txt")
	if err == nil {
		t.Fatalf("expected error reading non-existent file, got nil")
	}
}

func TestCacheManager_ExportTree(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		t.Fatalf("failed to init CacheManager: %v", err)
	}

	localRepoDir, commitSHA := createTestLocalRepo(t)
	localURL := "file://" + localRepoDir

	repoDecl := manifest.RepositoryDeclaration{
		ID:        "test-repo",
		URL:       localURL,
		SHA:       commitSHA,
		PathScope: ".",
	}

	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("failed to ensure repo: %v", err)
	}

	exportDir := filepath.Join(t.TempDir(), "export")
	if err := cm.ExportTree(ctx, repoDecl, exportDir); err != nil {
		t.Fatalf("failed to export tree: %v", err)
	}

	// Verify exported files exist on disk
	readmePath := filepath.Join(exportDir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Errorf("exported README.md not found: %v", err)
	}

	mainPath := filepath.Join(exportDir, "pkg", "service", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		t.Errorf("exported pkg/service/main.go not found: %v", err)
	}
}


func TestCacheManager_ListTreePreservesUnusualPathnames(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		t.Fatalf("failed to init CacheManager: %v", err)
	}

	localRepoDir, _ := createTestLocalRepo(t)
	oddNames := []string{"tab\tname.txt", "line\nbreak.txt", "back\\slash.txt"}
	for _, name := range oddNames {
		if err := os.WriteFile(filepath.Join(localRepoDir, name), []byte("unusual path\n"), 0644); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = localRepoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v, output: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	runGit("add", "--all")
	runGit("commit", "-m", "Add unusual pathnames")
	commitSHA := runGit("rev-parse", "HEAD")
	repoDecl := manifest.RepositoryDeclaration{
		ID: "odd-paths",
		URL: "file://" + localRepoDir,
		SHA: commitSHA,
		PathScope: ".",
	}
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("ensure repo: %v", err)
	}
	entries, err := cm.ListTree(ctx, repoDecl)
	if err != nil {
		t.Fatalf("ListTree: %v", err)
	}
	got := make(map[string]bool, len(entries))
	for _, entry := range entries {
		got[entry.Path] = true
	}
	for _, name := range oddNames {
		if !got[name] {
			t.Errorf("ListTree omitted or corrupted pathname %q; got %q", name, mapKeys(got))
		}
	}
}

func mapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}


func TestCacheManager_ExportTreeReapsTarWhenArchiveFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script process fixture is Unix-only")
	}
	binDir := t.TempDir()
	fakeGit := filepath.Join(binDir, "git")
	fakeTar := filepath.Join(binDir, "tar")
	if err := os.WriteFile(fakeGit, []byte("#!/bin/sh\nexit 23\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fakeTar, []byte("#!/bin/sh\nwhile IFS= read -r line; do :; done\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repoDecl := manifest.RepositoryDeclaration{ID: "archive-failure", SHA: "deadbeef"}
	err = cm.ExportTree(context.Background(), repoDecl, filepath.Join(t.TempDir(), "target"))
	if err == nil || !strings.Contains(err.Error(), "git archive failed") {
		t.Fatalf("ExportTree error = %v, want git archive failure", err)
	}
}
