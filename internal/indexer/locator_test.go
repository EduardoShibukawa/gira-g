package indexer_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func initGitRepo(t *testing.T) string {
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
	runGit("commit", "--allow-empty", "-m", "init")

	return dir
}

func TestIndexer_LocatorNormalization_WithinGitRoot(t *testing.T) {
	ctx := context.Background()
	repoGitDir := initGitRepo(t)

	// Materialize corpus inside git repo
	srcDir := t.TempDir()
	protoContent := "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n"
	if err := os.MkdirAll(filepath.Join(srcDir, "proto"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "proto", "order.proto"), []byte(protoContent), 0644); err != nil {
		t.Fatal(err)
	}
	sha, err := corpus.Materialize(ctx, srcDir, filepath.Join(repoGitDir, "source"))
	if err != nil {
		t.Fatal(err)
	}

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Manifest in subfolder of git repo
	subDir := filepath.Join(repoGitDir, "configs", "manifests")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifestFile := filepath.Join(subDir, "gira.yaml")
	if err := os.WriteFile(manifestFile, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: "test-system"},
		Repositories: []manifest.RepositoryDeclaration{
			{ID: "repo1", URL: filepath.Join(repoGitDir, "source"), SHA: sha, PathScope: "."},
		},
		Overrides: &manifest.OverridesDeclaration{
			Relationships: []manifest.RelationshipOverride{
				{SourceID: "service:OrderService", TargetID: "contract:custom-orders", Verb: "consumes", StartLine: 10, EndLine: 12},
			},
		},
	}

	// Build indexer with ManifestPath pointing to absolute path
	mm, err := indexer.Build(ctx, m, cm, indexer.Options{
		ManifestPath: manifestFile,
	})
	if err != nil {
		t.Fatalf("indexer.Build failed: %v", err)
	}

	var overrideEv *domain.EvidenceRecord
	for _, rel := range mm.Relationships {
		if rel.TargetID == "contract:custom-orders" {
			for i := range rel.Evidence {
				if rel.Evidence[i].Origin == "declared:override" {
					overrideEv = &rel.Evidence[i]
					break
				}
			}
		}
	}
	if overrideEv == nil {
		for _, unres := range mm.UnresolvedRefs {
			for i := range unres.Evidence {
				if unres.Evidence[i].Origin == "declared:override" {
					overrideEv = &unres.Evidence[i]
					break
				}
			}
		}
	}
	if overrideEv == nil {
		t.Fatalf("expected evidence with origin declared:override")
	}

	wantLocator := "configs/manifests/gira.yaml"
	if overrideEv.Locator.Path != wantLocator {
		t.Errorf("Locator.Path = %q, want %q", overrideEv.Locator.Path, wantLocator)
	}
}

func TestIndexer_LocatorNormalization_OutsideGitRoot_Fails(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	ctx := context.Background()
	outsideDir := t.TempDir()
	manifestFile := filepath.Join(outsideDir, "gira.yaml")
	if err := os.WriteFile(manifestFile, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: "test-system"},
		Repositories: []manifest.RepositoryDeclaration{
			{ID: "repo1", URL: t.TempDir(), SHA: strings.Repeat("a", 40), PathScope: "."},
		},
		Overrides: &manifest.OverridesDeclaration{
			Relationships: []manifest.RelationshipOverride{
				{SourceID: "service:A", TargetID: "service:B", Verb: "consumes"},
			},
		},
	}

	_, err = indexer.Build(ctx, m, cm, indexer.Options{
		ManifestPath: manifestFile,
	})
	if !errors.Is(err, manifest.ErrManifestOutsideGitRoot) {
		t.Errorf("expected ErrManifestOutsideGitRoot, got %v", err)
	}
	if !strings.Contains(err.Error(), "manifest_outside_git_root") {
		t.Errorf("expected error containing 'manifest_outside_git_root', got %v", err)
	}
}

func TestIndexer_LocatorNormalization_EmptyPath_Fails(t *testing.T) {
	ctx := context.Background()
	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: "test-system"},
		Repositories: []manifest.RepositoryDeclaration{
			{ID: "repo1", URL: t.TempDir(), SHA: strings.Repeat("a", 40), PathScope: "."},
		},
		Overrides: &manifest.OverridesDeclaration{
			Relationships: []manifest.RelationshipOverride{
				{SourceID: "service:A", TargetID: "service:B", Verb: "consumes"},
			},
		},
	}

	_, err = indexer.Build(ctx, m, cm, indexer.Options{
		ManifestPath: "",
	})
	if !errors.Is(err, manifest.ErrManifestPathRequired) {
		t.Errorf("expected ErrManifestPathRequired, got %v", err)
	}
	if !strings.Contains(err.Error(), "manifest_path_required") {
		t.Errorf("expected error containing 'manifest_path_required', got %v", err)
	}
}

func TestIndexer_NoOverrides_OutsideGitRoot_Succeeds(t *testing.T) {
	origHook := manifest.GitRootDiscoveryHook
	manifest.GitRootDiscoveryHook = nil
	t.Cleanup(func() { manifest.GitRootDiscoveryHook = origHook })

	ctx := context.Background()
	srcDir := t.TempDir()
	protoContent := "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n"
	if err := os.MkdirAll(filepath.Join(srcDir, "proto"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "proto", "order.proto"), []byte(protoContent), 0644); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(t.TempDir(), "repo")
	sha, err := corpus.Materialize(ctx, srcDir, repoDir)
	if err != nil {
		t.Fatal(err)
	}

	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Outside git directory, but NO overrides
	outsideDir := t.TempDir()
	manifestFile := filepath.Join(outsideDir, "gira.yaml")
	if err := os.WriteFile(manifestFile, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: "test-system"},
		Repositories: []manifest.RepositoryDeclaration{
			{ID: "repo1", URL: repoDir, SHA: sha, PathScope: "."},
		},
	}

	mm, err := indexer.Build(ctx, m, cm, indexer.Options{
		ManifestPath: manifestFile,
	})
	if err != nil {
		t.Fatalf("indexer.Build without overrides should succeed outside git root, got: %v", err)
	}
	if mm.TargetSystem != "test-system" {
		t.Errorf("got target system %q, want 'test-system'", mm.TargetSystem)
	}
}
