package main

import (
	"bytes"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/store"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createTestRepo(t *testing.T) (string, string) {
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

	composeContent := "version: '3.8'\nservices:\n  frontend:\n    image: otel/demo-frontend:v1.0\n"
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte(composeContent), 0644); err != nil {
		t.Fatal(err)
	}

	runGit("add", ".")
	runGit("commit", "-m", "initial commit")
	sha := runGit("rev-parse", "HEAD")

	return dir, sha
}

// AC1: Invocações relativas/absolutas/subdiretório/symlink/worktree convergem ao mesmo locator quando representam o mesmo manifesto.
func TestCLI_AC1_LocatorConvergence(t *testing.T) {
	repoDir, sha := createTestRepo(t)
	cacheDir := t.TempDir()

	// Create manifest in subfolder
	subDir := filepath.Join(repoDir, "conf", "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	subManifest := filepath.Join(subDir, "gira.yaml")
	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: convergence-system",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
		"overrides:",
		"  metadata:",
		`    "service:frontend":`,
		`      team: "core"`,
	}, "\n")
	if err := os.WriteFile(subManifest, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// File symlink
	fileSymlink := filepath.Join(repoDir, "symlink-sub.yaml")
	if err := os.Symlink(subManifest, fileSymlink); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Directory symlink
	dirSymlink := filepath.Join(repoDir, "conf-symlink")
	if err := os.Symlink(subDir, dirSymlink); err != nil {
		t.Skipf("directory symlinks not supported: %v", err)
	}
	dirSymlinkManifest := filepath.Join(dirSymlink, "gira.yaml")

	// Other directory inside repo for relative traversal with ..
	otherDir := filepath.Join(repoDir, "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		cwd  string
		path string
	}{
		{"sub absolute", repoDir, subManifest},
		{"sub relative from root", repoDir, "conf/sub/gira.yaml"},
		{"sub relative from inside sub", subDir, "gira.yaml"},
		{"sub relative with ..", otherDir, "../conf/sub/gira.yaml"},
		{"file symlink", repoDir, fileSymlink},
		{"dir symlink", repoDir, dirSymlinkManifest},
	}

	var canonicalEvidenceLocator string
	var baselineJSON []byte

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

			outPath := filepath.Join(t.TempDir(), "output.json")
			var stdout, stderr bytes.Buffer
			err = Run([]string{"index", tc.path, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("gira index failed: %v, stderr: %s", err, stderr.String())
			}

			s := store.NewStore("")
			mm, err := s.LoadFile(outPath)
			if err != nil {
				t.Fatalf("failed to load indexed macro map: %v", err)
			}

			// Find override evidence on frontend node
			var foundLocator string
			for _, n := range mm.Nodes {
				if n.ID == "service:frontend" {
					for _, ev := range n.Evidence {
						if ev.Origin == "declared:override" {
							foundLocator = ev.Locator.Path
							break
						}
					}
				}
			}
			if foundLocator == "" {
				t.Fatalf("could not find declared:override evidence on service:frontend")
			}

			wantLocator := "conf/sub/gira.yaml"
			if foundLocator != wantLocator {
				t.Errorf("expected locator %q, got %q", wantLocator, foundLocator)
			}

			rawJSON, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatal(err)
			}

			if baselineJSON == nil {
				baselineJSON = rawJSON
				canonicalEvidenceLocator = foundLocator
			} else {
				if foundLocator != canonicalEvidenceLocator {
					t.Errorf("evidence locator diverged: got %q, want %q", foundLocator, canonicalEvidenceLocator)
				}
				if !bytes.Equal(rawJSON, baselineJSON) {
					t.Errorf("output JSON diverged across invocation forms!")
				}
			}
		})
	}
}

// AC2: Manifesto com overrides fora da raiz falha com manifest_outside_git_root; falha de descoberta retorna git_root_unavailable.
func TestCLI_AC2_OutsideGitRoot_And_GitRootUnavailable_FailClosed(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	repoDir, sha := createTestRepo(t)
	cacheDir := t.TempDir()

	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: fail-closed-system",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
		"overrides:",
		"  metadata:",
		`    "service:frontend":`,
		`      team: "core"`,
	}, "\n")

	t.Run("manifest_outside_git_root", func(t *testing.T) {
		outsideDir := t.TempDir()
		outsideManifest := filepath.Join(outsideDir, "gira.yaml")
		if err := os.WriteFile(outsideManifest, []byte(manifestContent), 0644); err != nil {
			t.Fatal(err)
		}

		outPath := filepath.Join(t.TempDir(), "output.json")
		var stdout, stderr bytes.Buffer
		err := Run([]string{"index", outsideManifest, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
		if err == nil {
			t.Fatalf("expected error for manifest with overrides outside git root, got nil")
		}
		if !strings.Contains(err.Error(), "manifest_outside_git_root") {
			t.Errorf("expected error message to contain 'manifest_outside_git_root', got: %v", err)
		}
	})

	t.Run("manifest_inside_bare_repo", func(t *testing.T) {
		bareDir := filepath.Join(t.TempDir(), "bare.git")
		cmd := exec.Command("git", "init", "--bare", bareDir)
		cmd.Env = manifest.GitCommandEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init --bare failed: %v\nOutput: %s", err, out)
		}
		bareManifest := filepath.Join(bareDir, "gira.yaml")
		if err := os.WriteFile(bareManifest, []byte(manifestContent), 0644); err != nil {
			t.Fatal(err)
		}

		outPath := filepath.Join(t.TempDir(), "output.json")
		var stdout, stderr bytes.Buffer
		err := Run([]string{"index", bareManifest, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
		if err == nil {
			t.Fatalf("expected error for manifest inside bare repo, got nil")
		}
		if !strings.Contains(err.Error(), "manifest_outside_git_root") {
			t.Errorf("expected error to contain 'manifest_outside_git_root', got: %v", err)
		}
	})

	t.Run("git_root_unavailable", func(t *testing.T) {
		insideManifest := filepath.Join(repoDir, "gira.yaml")
		if err := os.WriteFile(insideManifest, []byte(manifestContent), 0644); err != nil {
			t.Fatal(err)
		}

		origHook := manifest.GitRootDiscoveryHook
		manifest.GitRootDiscoveryHook = func(p string) (string, error) {
			return "", manifest.ErrGitRootUnavailable
		}
		t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

		outPath := filepath.Join(t.TempDir(), "output.json")
		var stdout, stderr bytes.Buffer
		err := Run([]string{"index", insideManifest, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
		if err == nil {
			t.Fatalf("expected error under git root unavailable, got nil")
		}
		if !strings.Contains(err.Error(), "git_root_unavailable") {
			t.Errorf("expected error message to contain 'git_root_unavailable', got: %v", err)
		}
	})
}

// AC3: Sem overrides, ausência de raiz Git não bloqueia indexação apenas por si.
func TestCLI_AC3_NoOverrides_OutsideGitRoot_Succeeds(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	repoDir, sha := createTestRepo(t)
	cacheDir := t.TempDir()

	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: no-overrides-system",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
	}, "\n")

	outsideDir := t.TempDir()
	outsideManifest := filepath.Join(outsideDir, "gira.yaml")
	if err := os.WriteFile(outsideManifest, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "output.json")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", outsideManifest, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira index without overrides outside git root failed: %v, stderr: %s", err, stderr.String())
	}

	s := store.NewStore("")
	mm, err := s.LoadFile(outPath)
	if err != nil {
		t.Fatalf("failed to load indexed macro map: %v", err)
	}
	if mm.TargetSystem != "no-overrides-system" {
		t.Errorf("got target system %q, want 'no-overrides-system'", mm.TargetSystem)
	}
}

// AC4: Env Git adversarial não sequestra a raiz nem init.
func TestCLI_AC4_AdversarialGitEnv_DoesNotHijack(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	repoDir, sha := createTestRepo(t)
	cacheDir := t.TempDir()

	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: adversarial-system",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
		"overrides:",
		"  metadata:",
		`    "service:frontend":`,
		`      team: "security"`,
	}, "\n")

	manifestPath := filepath.Join(repoDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	hostileDir := t.TempDir()
	t.Setenv("GIT_DIR", filepath.Join(hostileDir, "fake-git"))
	t.Setenv("GIT_WORK_TREE", hostileDir)
	t.Setenv("GIT_COMMON_DIR", filepath.Join(hostileDir, "fake-common"))
	t.Setenv("GIT_INDEX_FILE", filepath.Join(hostileDir, "fake-index"))
	t.Setenv("GIT_OBJECT_DIRECTORY", filepath.Join(hostileDir, "fake-objects"))
	t.Setenv("GIT_ALTERNATE_OBJECT_DIRECTORIES", filepath.Join(hostileDir, "fake-alts"))
	t.Setenv("GIT_NAMESPACE", "hijacked")
	t.Setenv("GIT_PREFIX", "hijacked/")
	t.Setenv("GIT_CEILING_DIRECTORIES", repoDir)
	t.Setenv("GIT_CONFIG", filepath.Join(hostileDir, "fake-config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.bare")
	t.Setenv("GIT_CONFIG_VALUE_0", "true")

	// 1. gira index should succeed and use legitimate git root
	outPath := filepath.Join(t.TempDir(), "output.json")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", manifestPath, "--output", outPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira index failed under adversarial git env: %v, stderr: %s", err, stderr.String())
	}

	// 2. gira init should also work in a real repo and not be hijacked
	initRepoDir, _ := createTestRepo(t)
	var initStdout, initStderr bytes.Buffer
	err = Run([]string{"init", initRepoDir, "--cache-dir", cacheDir}, nil, &initStdout, &initStderr)
	if err != nil {
		t.Fatalf("gira init failed under adversarial git env: %v, stderr: %s", err, initStderr.String())
	}
	initManifest := filepath.Join(initRepoDir, "gira.yaml")
	if _, err := os.Stat(initManifest); err != nil {
		t.Fatalf("expected gira.yaml created at %s, got: %v", initManifest, err)
	}
}

// AC5: Em todos os erros, store/head/snapshot/manifesto ficam ausentes ou byte a byte inalterados.
func TestCLI_AC5_FailClosed_ArtifactsIntact(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	repoDir, sha := createTestRepo(t)
	cacheDir := t.TempDir()

	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: fail-closed-evidence",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
		"overrides:",
		"  metadata:",
		`    "service:frontend":`,
		`      team: "core"`,
	}, "\n")

	outsideDir := t.TempDir()
	outsideManifest := filepath.Join(outsideDir, "gira.yaml")
	if err := os.WriteFile(outsideManifest, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Non-existent output snapshot must remain non-existent on error
	nonExistentSnapshot := filepath.Join(t.TempDir(), "nonexistent-snapshot.json")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", outsideManifest, "--output", nonExistentSnapshot, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected index to fail")
	}
	if _, err := os.Stat(nonExistentSnapshot); !os.IsNotExist(err) {
		t.Errorf("expected snapshot to not exist after error, got err: %v", err)
	}

	// 2. Pre-existing output snapshot must remain byte-for-byte unchanged
	preExistingSnapshot := filepath.Join(t.TempDir(), "existing-snapshot.json")
	initialContent := []byte(`{"pre_existing": true, "bytes": "unchanged"}`)
	if err := os.WriteFile(preExistingSnapshot, initialContent, 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", outsideManifest, "--output", preExistingSnapshot, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected index to fail")
	}
	afterBytes, err := os.ReadFile(preExistingSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterBytes, initialContent) {
		t.Errorf("pre-existing snapshot was altered! got %s, want %s", string(afterBytes), string(initialContent))
	}

	// 3. Canonical store must not have any new systems or versions created
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", outsideManifest, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected index to fail")
	}
	expectedStoreDir := filepath.Join(tmpHome, ".cache", "gira", "indexes", "fail-closed-evidence")
	if _, err := os.Stat(expectedStoreDir); !os.IsNotExist(err) {
		t.Errorf("expected store directory %s to not exist, got err: %v", expectedStoreDir, err)
	}
}
