package manifest_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func initTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = manifest.GitCommandEnv()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")
	runGit("commit", "--allow-empty", "-m", "initial commit")

	return dir
}

func TestResolveManifestLocator_EmptyPath(t *testing.T) {
	for _, p := range []string{"", "   ", "\t", "\n"} {
		_, err := manifest.ResolveManifestLocator(p)
		if !errors.Is(err, manifest.ErrManifestPathRequired) {
			t.Errorf("for path %q, got err %v, want %v", p, err, manifest.ErrManifestPathRequired)
		}
		if !strings.Contains(err.Error(), "manifest_path_required") {
			t.Errorf("expected error string to contain 'manifest_path_required', got %q", err.Error())
		}
	}
}

func TestResolveManifestLocator_NonExistentFile(t *testing.T) {
	_, err := manifest.ResolveManifestLocator("/path/to/definitely/nonexistent/gira.yaml")
	if err == nil {
		t.Fatalf("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read manifest file") {
		t.Errorf("expected 'failed to read manifest file', got: %v", err)
	}
}

func TestResolveManifestLocator_Convergence(t *testing.T) {
	repoDir := initTestGitRepo(t)

	// 1. Root manifest
	rootManifest := filepath.Join(repoDir, "gira.yaml")
	if err := os.WriteFile(rootManifest, []byte("version: 1\ntarget_system:\n  name: sys\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Subdirectory manifest
	subDir := filepath.Join(repoDir, "conf", "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	otherDir := filepath.Join(repoDir, "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	subManifest := filepath.Join(subDir, "gira.yaml")
	if err := os.WriteFile(subManifest, []byte("version: 1\ntarget_system:\n  name: sys\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. File symlink pointing to subManifest
	fileSymlink := filepath.Join(repoDir, "symlink-sub.yaml")
	if err := os.Symlink(subManifest, fileSymlink); err != nil {
		t.Skipf("symlinks not supported on this filesystem: %v", err)
	}

	// 4. Directory symlink pointing to subDir
	dirSymlink := filepath.Join(repoDir, "conf-symlink")
	if err := os.Symlink(subDir, dirSymlink); err != nil {
		t.Skipf("directory symlinks not supported on this filesystem: %v", err)
	}
	dirSymlinkManifest := filepath.Join(dirSymlink, "gira.yaml")

	// 5. Worktree
	wtDir := filepath.Join(t.TempDir(), "wt")
	cmd := exec.Command("git", "-C", repoDir, "worktree", "add", wtDir, "-b", "wt-branch")
	cmd.Env = manifest.GitCommandEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create worktree: %v\nOutput: %s", err, out)
	}
	wtManifest := filepath.Join(wtDir, "gira.yaml")
	if err := os.WriteFile(wtManifest, []byte("version: 1\ntarget_system:\n  name: sys\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test Convergence
	cases := []struct {
		name string
		cwd  string
		path string
		want string
	}{
		{"root absolute", repoDir, rootManifest, "gira.yaml"},
		{"root relative", repoDir, "gira.yaml", "gira.yaml"},
		{"sub absolute", repoDir, subManifest, "conf/sub/gira.yaml"},
		{"sub relative from root", repoDir, "conf/sub/gira.yaml", "conf/sub/gira.yaml"},
		{"sub relative from inside sub", subDir, "gira.yaml", "conf/sub/gira.yaml"},
		{"sub relative with ..", filepath.Join(repoDir, "other"), "../conf/sub/gira.yaml", "conf/sub/gira.yaml"},
		{"file symlink", repoDir, fileSymlink, "conf/sub/gira.yaml"},
		{"dir symlink", repoDir, dirSymlinkManifest, "conf/sub/gira.yaml"},
		{"worktree absolute", wtDir, wtManifest, "gira.yaml"},
		{"worktree relative", wtDir, "gira.yaml", "gira.yaml"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origCwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(tc.cwd); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(origCwd) })

			loc, err := manifest.ResolveManifestLocator(tc.path)
			if err != nil {
				t.Fatalf("ResolveManifestLocator(%q) failed: %v", tc.path, err)
			}
			if loc != tc.want {
				t.Errorf("ResolveManifestLocator(%q) = %q, want %q", tc.path, loc, tc.want)
			}
			// Verify forward slash normalization
			if strings.Contains(loc, "\\") {
				t.Errorf("locator contains backslashes: %q", loc)
			}
		})
	}
}

func TestResolveManifestLocator_OutsideGitRoot(t *testing.T) {
	nonGitDir := t.TempDir()
	manifestPath := filepath.Join(nonGitDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Non-git directory
	_, err := manifest.ResolveManifestLocator(manifestPath)
	if !errors.Is(err, manifest.ErrManifestOutsideGitRoot) {
		t.Errorf("expected ErrManifestOutsideGitRoot, got %v", err)
	}
	if !strings.Contains(err.Error(), "manifest_outside_git_root") {
		t.Errorf("expected error message to contain 'manifest_outside_git_root', got %q", err.Error())
	}

	// 2. Symlink inside git repo pointing to manifest outside git repo
	repoDir := initTestGitRepo(t)
	symlinkPath := filepath.Join(repoDir, "outside-link.yaml")
	if err := os.Symlink(manifestPath, symlinkPath); err == nil {
		_, err = manifest.ResolveManifestLocator(symlinkPath)
		if !errors.Is(err, manifest.ErrManifestOutsideGitRoot) {
			t.Errorf("expected ErrManifestOutsideGitRoot for symlink pointing outside git root, got %v", err)
		}
	}

	// 3. Bare repository
	bareDir := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "init", "--bare", bareDir)
	cmd.Env = manifest.GitCommandEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to init bare repo: %v\nOutput: %s", err, out)
	}
	bareManifest := filepath.Join(bareDir, "gira.yaml")
	if err := os.WriteFile(bareManifest, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = manifest.ResolveManifestLocator(bareManifest)
	if !errors.Is(err, manifest.ErrManifestOutsideGitRoot) {
		t.Errorf("expected ErrManifestOutsideGitRoot for bare repo, got %v", err)
	}

	// 4. Inside .git directory
	dotGitManifest := filepath.Join(repoDir, ".git", "gira.yaml")
	if err := os.WriteFile(dotGitManifest, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = manifest.ResolveManifestLocator(dotGitManifest)
	if !errors.Is(err, manifest.ErrManifestOutsideGitRoot) {
		t.Errorf("expected ErrManifestOutsideGitRoot for manifest inside .git, got %v", err)
	}
}

func TestResolveManifestLocator_GitRootUnavailable(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = func(p string) (string, error) {
		return "", manifest.ErrGitRootUnavailable
	}
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	repoDir := initTestGitRepo(t)
	manifestPath := filepath.Join(repoDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := manifest.ResolveManifestLocator(manifestPath)
	if !errors.Is(err, manifest.ErrGitRootUnavailable) {
		t.Fatalf("expected ErrGitRootUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), "git_root_unavailable") {
		t.Errorf("expected error message to contain 'git_root_unavailable', got %q", err.Error())
	}
}

func TestResolveManifestLocator_AdversarialGitEnv(t *testing.T) {
	repoDir := initTestGitRepo(t)
	manifestPath := filepath.Join(repoDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	otherDir := t.TempDir()
	t.Setenv("GIT_DIR", filepath.Join(otherDir, "bogus-git"))
	t.Setenv("GIT_WORK_TREE", otherDir)
	t.Setenv("GIT_COMMON_DIR", filepath.Join(otherDir, "bogus-common"))
	t.Setenv("GIT_INDEX_FILE", filepath.Join(otherDir, "bogus-index"))
	t.Setenv("GIT_OBJECT_DIRECTORY", filepath.Join(otherDir, "bogus-objects"))
	t.Setenv("GIT_ALTERNATE_OBJECT_DIRECTORIES", filepath.Join(otherDir, "bogus-alts"))
	t.Setenv("GIT_NAMESPACE", "adversarial")
	t.Setenv("GIT_PREFIX", "injected/")
	t.Setenv("GIT_CEILING_DIRECTORIES", repoDir)
	t.Setenv("GIT_CONFIG", filepath.Join(otherDir, "bogus-config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.bare")
	t.Setenv("GIT_CONFIG_VALUE_0", "true")

	loc, err := manifest.ResolveManifestLocator(manifestPath)
	if err != nil {
		t.Fatalf("expected discovery to succeed despite adversarial environment, got: %v", err)
	}
	if loc != "gira.yaml" {
		t.Errorf("got locator %q, want 'gira.yaml'", loc)
	}
}

func TestHasDeclarativeOverrides(t *testing.T) {
	var m *manifest.Manifest
	if m.HasDeclarativeOverrides() {
		t.Errorf("nil manifest should return false")
	}

	m = &manifest.Manifest{
		Version: 1,
		TargetSystem: manifest.TargetSystemMeta{
			Name: "test",
		},
	}
	if m.HasDeclarativeOverrides() {
		t.Errorf("manifest with no overrides should return false")
	}

	m.Overrides = &manifest.OverridesDeclaration{}
	if m.HasDeclarativeOverrides() {
		t.Errorf("manifest with empty overrides should return false")
	}

	m.Overrides.Relationships = []manifest.RelationshipOverride{
		{SourceID: "s1", TargetID: "s2", Verb: "consumes"},
	}
	if !m.HasDeclarativeOverrides() {
		t.Errorf("manifest with relationship overrides should return true")
	}

	m.Overrides.Relationships = nil
	m.Overrides.Metadata = map[string]map[string]any{
		"s1": {"team": "core"},
	}
	if !m.HasDeclarativeOverrides() {
		t.Errorf("manifest with metadata overrides should return true")
	}

	m.Overrides = nil
	m.TargetSystem.Overrides = &manifest.OverridesDeclaration{
		Relationships: []manifest.RelationshipOverride{
			{SourceID: "s1", TargetID: "s2", Verb: "consumes"},
		},
	}
	if !m.HasDeclarativeOverrides() {
		t.Errorf("manifest with target_system overrides should return true")
	}
}
