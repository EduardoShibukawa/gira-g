package gitcache_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	if runtime.GOOS != "windows" {
		info, err := os.Stat(exportDir)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0755 {
			t.Fatalf("export root permissions = %04o, want 0755", got)
		}
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


func TestCacheManager_ExportTreeDoesNotStartExtractorWhenArchiveFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-backed process fixture is Unix-only")
	}
	binDir := t.TempDir()
	fakeGit := filepath.Join(binDir, "git")
	fakeTar := filepath.Join(binDir, "tar")
	tarMarker := filepath.Join(t.TempDir(), "tar-was-run")
	if err := os.WriteFile(fakeGit, []byte("#!/bin/sh\nexit 23\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fakeTar, []byte("#!/bin/sh\necho called > \"$TAR_CALLED\"\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TAR_CALLED", tarMarker)
	t.Setenv("PATH", binDir)

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "target")
	repoDecl := manifest.RepositoryDeclaration{ID: "archive-failure", SHA: "deadbeef"}
	err = cm.ExportTree(context.Background(), repoDecl, targetDir)
	if err == nil || !strings.Contains(err.Error(), "git archive failed") {
		t.Fatalf("ExportTree error = %v, want git archive failure", err)
	}
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("target was published despite archive failure: %v", err)
	}
	if _, err := os.Stat(tarMarker); !os.IsNotExist(err) {
		t.Fatalf("tar extractor ran after git archive failed: %v", err)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary extraction artifacts leaked after archive failure: %#v", entries)
	}
}
func TestCacheManager_ReadsPinnedObjectsIndependentOfReplaceRefsAndGlobalConfig(t *testing.T) {
	ctx := context.Background()
	// Keep this fixture directly under os.TempDir. t.TempDir includes the
	// long test name on Windows and can push Git's pack .keep path over MAX_PATH.
	localRepoDir, err := os.MkdirTemp(os.TempDir(), "gira-repo-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(localRepoDir) })
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
	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(localRepoDir, "README.md"), []byte("# Hello World\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "Initial commit")
	initialSHA := runGit("rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(localRepoDir, "README.md"), []byte("pinned README\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localRepoDir, "café.proto"), []byte("syntax = \"proto3\";\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "--all")
	runGit("commit", "-m", "second pinned version")
	pinnedSHA := runGit("rev-parse", "HEAD")

	repoDecl := manifest.RepositoryDeclaration{
		ID: "pinned-source",
		URL: "file://" + filepath.ToSlash(localRepoDir),
		SHA: pinnedSHA,
		PathScope: ".",
	}
	cacheBase, err := os.MkdirTemp(os.TempDir(), "gira-cache-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(cacheBase) })
	cm, err := gitcache.NewCacheManager(cacheBase)
	if err != nil {
		t.Fatal(err)
	}
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}
	cacheDir := cm.RepoDir(repoDecl)
	cachedGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"--git-dir=" + cacheDir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git --git-dir=%s %v failed: %v: %s", cacheDir, args, err, out)
		}
	}
	cachedGit("config", "core.quotePath", "true")
	quotedEntries, err := cm.ListTree(ctx, repoDecl)
	if err != nil {
		t.Fatalf("ListTree with repository core.quotePath=true: %v", err)
	}
	cachedGit("config", "core.quotePath", "false")
	plainEntries, err := cm.ListTree(ctx, repoDecl)
	if err != nil {
		t.Fatalf("ListTree with repository core.quotePath=false: %v", err)
	}
	if !reflect.DeepEqual(quotedEntries, plainEntries) {
		t.Fatalf("NUL-delimited pinned tree changed with core.quotePath: true=%#v false=%#v", quotedEntries, plainEntries)
	}
	replace := exec.Command("git", "--git-dir="+cacheDir, "replace", pinnedSHA, initialSHA)
	if out, err := replace.CombinedOutput(); err != nil {
		t.Fatalf("create replace ref: %v: %s", err, out)
	}

	configTrue := filepath.Join(t.TempDir(), "quote-true.gitconfig")
	configFalse := filepath.Join(t.TempDir(), "quote-false.gitconfig")
	if err := os.WriteFile(configTrue, []byte("[core]\n\tquotePath = true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFalse, []byte("[core]\n\tquotePath = false\n"), 0600); err != nil {
		t.Fatal(err)
	}

	type snapshot struct {
		entries []gitcache.FileEntry
		readme string
		proto string
		exportedReadme string
		exportedProto string
		exportOutcome string
	}
	readSnapshot := func(configPath, suffix string) snapshot {
		t.Helper()
		t.Setenv("GIT_CONFIG_GLOBAL", configPath)
		t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
		t.Setenv("GIT_NO_REPLACE_OBJECTS", "0")
		t.Setenv("GIT_PREFIX", "redirected/")
		t.Setenv("GIT_CEILING_DIRECTORIES", localRepoDir)
		t.Setenv("GIT_LITERAL_PATHSPECS", "1")
		t.Setenv("GIT_GLOB_PATHSPECS", "1")
		t.Setenv("GIT_NOGLOB_PATHSPECS", "1")
		t.Setenv("GIT_ICASE_PATHSPECS", "1")
		t.Setenv("GIT_ATTR_SOURCE", "deadbeef")

		entries, err := cm.ListTree(ctx, repoDecl)
		if err != nil {
			t.Fatalf("ListTree with config %q: %v", configPath, err)
		}
		foundProto := false
		for _, entry := range entries {
			if entry.Path == "café.proto" {
				foundProto = true
			}
		}
		if !foundProto {
			t.Fatalf("pinned tree omitted café.proto with replace ref present: %#v", entries)
		}
		readme, err := cm.ReadFile(ctx, repoDecl, "README.md")
		if err != nil {
			t.Fatalf("ReadFile README with config %q: %v", configPath, err)
		}
		proto, err := cm.ReadFile(ctx, repoDecl, "café.proto")
		if err != nil {
			t.Fatalf("ReadFile non-ASCII path with config %q: %v", configPath, err)
		}
		if string(readme) != "pinned README\n" {
			t.Fatalf("ReadFile used replacement tree: got %q", readme)
		}

		exportDir := filepath.Join(t.TempDir(), "export-"+suffix)
		if err := cm.ExportTree(ctx, repoDecl, exportDir); err != nil {
			if runtime.GOOS != "windows" {
				t.Fatalf("ExportTree with config %q: %v", configPath, err)
			}
			if !strings.Contains(err.Error(), "café.proto") && !strings.Contains(err.Error(), "tar extract failed") {
				t.Fatalf("Windows rejected export without a path-materialization diagnostic: %v", err)
			}
			if _, statErr := os.Stat(exportDir); !os.IsNotExist(statErr) {
				t.Fatalf("failed export published a partial target directory: %v", statErr)
			}
			leftovers, readErr := os.ReadDir(filepath.Dir(exportDir))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(leftovers) != 0 {
				t.Fatalf("failed export left staging artifacts: %#v", leftovers)
			}
			return snapshot{entries: entries, readme: string(readme), proto: string(proto), exportOutcome: "rejected"}
		}
		exportedReadme, err := os.ReadFile(filepath.Join(exportDir, "README.md"))
		if err != nil {
			t.Fatalf("read exported README: %v", err)
		}
		exportedProto, err := os.ReadFile(filepath.Join(exportDir, "café.proto"))
		if err != nil {
			t.Fatalf("ExportTree returned success without materializing café.proto: %v", err)
		}
		return snapshot{entries: entries, readme: string(readme), proto: string(proto), exportedReadme: string(exportedReadme), exportedProto: string(exportedProto), exportOutcome: "published"}
	}
	gotTrue := readSnapshot(configTrue, "true")
	gotFalse := readSnapshot(configFalse, "false")
	if !reflect.DeepEqual(gotTrue, gotFalse) {
		t.Fatalf("pinned reads changed with global config\nquotePath=true: %#v\nquotePath=false: %#v", gotTrue, gotFalse)
	}
	if gotTrue.proto != "syntax = \"proto3\";\n" {
		t.Fatalf("pinned file read differs from source bytes: %#v", gotTrue)
	}
	if gotTrue.exportOutcome == "published" && (gotTrue.exportedReadme != gotTrue.readme || gotTrue.exportedProto != gotTrue.proto) {
		t.Fatalf("published export differs from pinned bytes: %#v", gotTrue)
	}
	if runtime.GOOS != "windows" && gotTrue.exportOutcome != "published" {
		t.Fatalf("Unix export unexpectedly failed closed: %#v", gotTrue)
	}
}

func TestCacheManager_PreservesUserURLRewriteForCloneAndFetch(t *testing.T) {
	ctx := context.Background()
	localRepoDir, initialSHA := createTestLocalRepo(t)
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
	fileURL := "file://" + filepath.ToSlash(localRepoDir)
	if runtime.GOOS == "windows" {
		fileURL = "file:///" + filepath.ToSlash(localRepoDir)
	}
	configPath := filepath.Join(t.TempDir(), "user.gitconfig")
	config := "[url \"" + fileURL + "\"]\n\tinsteadOf = https://example.invalid/gira/repo\n"
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", configPath)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	repoDecl := manifest.RepositoryDeclaration{ID: "rewrite", URL: "https://example.invalid/gira/repo", SHA: initialSHA}
	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("clone should respect user url.insteadOf config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(localRepoDir, "new-after-clone.txt"), []byte("fetch proof\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "--all")
	runGit("commit", "-m", "new commit for fetch")
	fetchedSHA := runGit("rev-parse", "HEAD")
	repoDecl.SHA = fetchedSHA
	if err := cm.EnsureRepo(ctx, repoDecl, false); err != nil {
		t.Fatalf("fetch should respect user url.insteadOf config: %v", err)
	}
	if !cm.HasCommit(ctx, cm.RepoDir(repoDecl), fetchedSHA) {
		t.Fatalf("fetched pinned commit %s is not present in cache", fetchedSHA)
	}
}

func TestCacheManager_ExportTreeDoesNotPublishPartialArchive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-backed partial extraction fixture is Unix-only")
	}
	ctx := context.Background()
	archivePath := filepath.Join(t.TempDir(), "partial.tar")
	payload := []byte("partial output stays private")
	var archiveBytes bytes.Buffer
	tw := tar.NewWriter(&archiveBytes)
	if err := tw.WriteHeader(&tar.Header{Name: "partial.txt", Mode: 0644, Size: int64(len(payload)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivePath, archiveBytes.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	realTar, err := exec.LookPath("tar")
	if err != nil {
		t.Fatalf("find real tar executable: %v", err)
	}
	binDir := t.TempDir()
	fakeGit := filepath.Join(binDir, "git")
	fakeTar := filepath.Join(binDir, "tar")
	gitScript := "#!/bin/sh\ncat \"$GIRA_ARCHIVE_FIXTURE\"\nexit 0\n"
	tarScript := "#!/bin/sh\n\"$REAL_TAR\" \"$@\"\nexit 23\n"
	if err := os.WriteFile(fakeGit, []byte(gitScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fakeTar, []byte(tarScript), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIRA_ARCHIVE_FIXTURE", archivePath)
	t.Setenv("REAL_TAR", realTar)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cacheRoot, err := os.MkdirTemp(os.TempDir(), "gira-partial-cache-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(cacheRoot) })
	cm, err := gitcache.NewCacheManager(cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := os.MkdirTemp(os.TempDir(), "gira-partial-out-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(parent) })
	targetDir := filepath.Join(parent, "export")
	repoDecl := manifest.RepositoryDeclaration{ID: "partial-stream", SHA: "deadbeef"}

	err = cm.ExportTree(ctx, repoDecl, targetDir)
	if err == nil || !strings.Contains(err.Error(), "tar extract failed") {
		t.Fatalf("ExportTree error = %v, want tar extraction failure", err)
	}
	if _, statErr := os.Stat(targetDir); !os.IsNotExist(statErr) {
		t.Fatalf("partial extraction became visible at targetDir; stat error = %v", statErr)
	}
	leftovers, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("staging/archive artifacts leaked after failed extraction: %v", leftovers)
	}
}
func TestCacheManager_ExportTreeFailsClosedOnUnmaterializableWindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Git for Windows path-materialization behavior is the subject of this test")
	}
	repoDecl := manifest.RepositoryDeclaration{ID: "windows-tab-path"}
	cacheRoot, err := os.MkdirTemp(os.TempDir(), "gira-win-cache-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(cacheRoot) })
	cm, err := gitcache.NewCacheManager(cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	repoDir := cm.RepoDir(repoDecl)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	runGit := func(input string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"--git-dir=" + repoDir}, args...)...)
		cmd.Env = gitcache.GitEnvForRead()
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	runGit("", "init", "--bare")
	runGit("", "config", "user.name", "Test Runner")
	runGit("", "config", "user.email", "test@example.com")
	goodBlob := runGit("good file\n", "hash-object", "-w", "--stdin")
	badBlob := runGit("unmaterializable file\n", "hash-object", "-w", "--stdin")
	treeInput := "100644 blob " + goodBlob + "\ta.txt\x00100644 blob " + badBlob + "\tb\tbad.txt\x00"
	treeSHA := runGit(treeInput, "mktree", "-z")
	commitSHA := runGit("", "commit-tree", treeSHA, "-m", "tab path fixture")
	repoDecl.SHA = commitSHA

	parent, err := os.MkdirTemp(os.TempDir(), "gira-win-out-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(parent) })
	targetDir := filepath.Join(parent, "export")
	err = cm.ExportTree(context.Background(), repoDecl, targetDir)
	if err == nil {
		t.Fatal("ExportTree unexpectedly accepted a tree with a tab pathname")
	}
	if !strings.Contains(err.Error(), "archive failed") && !strings.Contains(err.Error(), "tar extract failed") && !strings.Contains(err.Error(), "archive extraction incomplete") && !strings.Contains(err.Error(), "cannot be materialized") {
		t.Fatalf("ExportTree error is not diagnostic: %v", err)
	}
	if _, statErr := os.Stat(targetDir); !os.IsNotExist(statErr) {
		t.Fatalf("failed export left a target directory / partial tree: stat error = %v", statErr)
	}
	leftovers, readErr := os.ReadDir(parent)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(leftovers) != 0 {
		t.Fatalf("failed export left staging artifacts: %#v", leftovers)
	}
}

