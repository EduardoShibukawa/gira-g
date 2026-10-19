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

func TestCacheManager_CloneAndFetchRespectUserGitConfig(t *testing.T) {
	ctx := context.Background()
	localRepoDir, firstSHA := createTestLocalRepo(t)
	remoteURL := "https://example.invalid/gira-g-user-config.git"
	localURL := "file://" + filepath.ToSlash(localRepoDir)
	configPath := filepath.Join(t.TempDir(), "gitconfig")
	config := "[url \"" + localURL + "\"]\n\tinsteadOf = " + remoteURL + "\n"
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	// A host Git URL rewrite must remain active for clone/fetch. Local probes
	// use gitCommandEnv separately; CacheManager transport commands inherit this
	// user config by design (ADR 0009).
	t.Setenv("GIT_CONFIG_GLOBAL", configPath)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	decl := manifest.RepositoryDeclaration{ID: "user-config", URL: remoteURL, SHA: firstSHA, PathScope: "."}
	if err := cm.EnsureRepo(ctx, decl, false); err != nil {
		t.Fatalf("clone through configured URL rewrite failed: %v", err)
	}
	bareDir := cm.RepoDir(decl)
	if !cm.HasCommit(ctx, bareDir, firstSHA) {
		t.Fatalf("clone did not cache pinned commit %s", firstSHA)
	}

	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = localRepoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	if err := os.WriteFile(filepath.Join(localRepoDir, "second.txt"), []byte("second commit\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "second.txt")
	runGit("commit", "-m", "second commit")
	secondSHA := runGit("rev-parse", "HEAD")
	decl.SHA = secondSHA
	if err := cm.EnsureRepo(ctx, decl, false); err != nil {
		t.Fatalf("fetch through configured URL rewrite failed: %v", err)
	}
	if !cm.HasCommit(ctx, bareDir, secondSHA) {
		t.Fatalf("fetch did not cache commit %s", secondSHA)
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
	if runtime.GOOS == "windows" {
		t.Skip("Windows filesystems forbid control characters and backslashes in filenames")
	}
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


func TestCacheManager_ExportTreeDoesNotPublishPartialArchive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("partial tar-stream fixture uses a Unix shell wrapper")
	}
	realGitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cacheDir := t.TempDir()
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	localRepoDir, commitSHA := createTestLocalRepo(t)
	repoDecl := manifest.RepositoryDeclaration{
		ID: "partial-archive",
		URL: "file://" + localRepoDir,
		SHA: commitSHA,
		PathScope: ".",
	}
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("ensure repo: %v", err)
	}

	targetParent := t.TempDir()
	targetDir := filepath.Join(targetParent, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	keepPath := filepath.Join(targetDir, "keep.txt")
	if err := os.WriteFile(keepPath, []byte("previous tree\n"), 0644); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	// Emit the first 1024 bytes of a real archive, then fail. On common Unix
	// tar implementations this is enough to materialize at least one entry,
	// reproducing the partial-output shape seen from Git for Windows.
	quotedGit := "'" + strings.ReplaceAll(realGitPath, "'", "'\\''") + "'"
	fakeGit := filepath.Join(binDir, "git")
	script := "#!/bin/sh\n" + quotedGit + " \"$@\" | head -c 1024\nexit 23\n"
	if err := os.WriteFile(fakeGit, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err = cm.ExportTree(ctx, repoDecl, targetDir)
	if err == nil || !strings.Contains(err.Error(), "git archive failed") {
		t.Fatalf("ExportTree error = %v, want archive failure", err)
	}
	if got, err := os.ReadFile(keepPath); err != nil || string(got) != "previous tree\n" {
		t.Fatalf("previous target was changed after partial archive: got=%q err=%v", got, err)
	}
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "keep.txt" {
		t.Fatalf("partial archive leaked into target: entries=%v", mapDirEntries(entries))
	}
	parentEntries, err := os.ReadDir(targetParent)
	if err != nil {
		t.Fatal(err)
	}
	if len(parentEntries) != 1 || parentEntries[0].Name() != "target" {
		t.Fatalf("staging directory leaked after archive failure: entries=%v", mapDirEntries(parentEntries))
	}
}

func TestCacheManager_ExportTreeHandlesTabPathWithoutPartialPublish(t *testing.T) {
	ctx := context.Background()
	localRepoDir, baseSHA := createTestLocalRepo(t)
	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = localRepoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	const contents = "tab pathname fixture\n"
	blobCmd := exec.Command("git", "hash-object", "-w", "--stdin")
	blobCmd.Dir = localRepoDir
	blobCmd.Stdin = strings.NewReader(contents)
	blobOut, err := blobCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git hash-object failed: %v: %s", err, blobOut)
	}
	blobSHA := strings.TrimSpace(string(blobOut))
	treeCmd := exec.Command("git", "mktree", "-z")
	treeCmd.Dir = localRepoDir
	treeCmd.Stdin = strings.NewReader("100644 blob " + blobSHA + "\ttab\tname.txt\x00")
	treeOut, err := treeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git mktree -z failed: %v: %s", err, treeOut)
	}
	treeSHA := strings.TrimSpace(string(treeOut))
	commitSHA := runGit("commit-tree", treeSHA, "-p", baseSHA, "-m", "tab pathname fixture")
	// Make the generated commit reachable to a bare clone without ever asking
	// the local filesystem to create the unmaterializable pathname.
	runGit("update-ref", "refs/heads/tab-path-test", commitSHA)

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repoDecl := manifest.RepositoryDeclaration{
		ID: "tab-path",
		URL: "file://" + localRepoDir,
		SHA: commitSHA,
		PathScope: ".",
	}
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("ensure repo with synthetic tab-path tree: %v", err)
	}

	targetParent := t.TempDir()
	targetDir := filepath.Join(targetParent, "export")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	keepPath := filepath.Join(targetDir, "keep.txt")
	if err := os.WriteFile(keepPath, []byte("previous tree\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err = cm.ExportTree(ctx, repoDecl, targetDir)
	if err != nil {
		if runtime.GOOS != "windows" {
			t.Fatalf("Unix should materialize a tab pathname: %v", err)
		}
		if !strings.Contains(err.Error(), "git archive failed") && !strings.Contains(err.Error(), "tar extract failed") {
			t.Fatalf("unmaterializable path returned no archive/extraction diagnostic: %v", err)
		}
		if got, readErr := os.ReadFile(keepPath); readErr != nil || string(got) != "previous tree\n" {
			t.Fatalf("previous target changed on failed Windows extraction: got=%q err=%v", got, readErr)
		}
		entries, readErr := os.ReadDir(targetDir)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if len(entries) != 1 || entries[0].Name() != "keep.txt" {
			t.Fatalf("partial Windows extraction leaked into target: entries=%v", mapDirEntries(entries))
		}
		return
	}

	got, err := os.ReadFile(filepath.Join(targetDir, "tab\tname.txt"))
	if err != nil || string(got) != contents {
		t.Fatalf("tab pathname was not exported intact: got=%q err=%v", got, err)
	}
	if _, err := os.Stat(keepPath); !os.IsNotExist(err) {
		t.Fatalf("successful export did not replace the previous tree: stat err=%v", err)
	}
}

func mapDirEntries(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name())
	}
	return out
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
