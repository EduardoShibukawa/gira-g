package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

func validTestSchema2Map(targetSystem string) *domain.MacroMap {
	si := &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Extractors: []domain.ExtractorIdentity{
			{ID: "go", Version: "1.0.0"},
		},
		Sources: []domain.SourceRepoIdentity{
			{RepoID: "repo-1", CommitSHA: "2222222222222222222222222222222222222222", SHA: "2222222222222222222222222222222222222222", URLSHA256: "hash-repo-1"},
		},
	}
	mm := domain.NewMacroMapV2(targetSystem, si)
	_ = mm.AddNode(domain.Node{
		ID:   "service:api",
		Kind: domain.NodeService,
		Name: "api",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-1",
					CommitSHA: "2222222222222222222222222222222222222222",
					Path:      "main.go",
					StartLine: 1,
					EndLine:   10,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC),
			},
		},
	})
	return mm
}

// AC3 & AC7: CLI commands with explicit --index handle legacy v1 with warning,
// and distinguish index_corrupt, index_schema_unsupported, and index_missing.
func TestCLI_Resolver_ExplicitIndexErrorsAndWarning(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Legacy v1 index via explicit --index: succeeds in export and emits warning
	legacyFile := filepath.Join(tempDir, "legacy.json")
	legacyJSON := `{"schema_version": "1.0.0", "target_system": "legacy-sys", "nodes": [{"id": "service:leg", "kind": "Service", "name": "leg"}]}`
	if err := os.WriteFile(legacyFile, []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run([]string{"export", "--index", legacyFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected export with explicit legacy index to succeed: %v", err)
	}
	if !strings.Contains(stdout.String(), "service:leg") {
		t.Errorf("expected legacy node in export output")
	}

	// 2. Corrupt index via explicit --index: returns index_corrupt
	corruptFile := filepath.Join(tempDir, "corrupt.json")
	if err := os.WriteFile(corruptFile, []byte(`{ "schema_version": 2, `), 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"export", "--index", corruptFile}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for corrupt index, got nil")
	}
	if !strings.Contains(err.Error(), "index_corrupt") {
		t.Errorf("expected error to cite index_corrupt, got: %v", err)
	}

	// 3. Unsupported schema via explicit --index: returns index_schema_unsupported
	unsupportedFile := filepath.Join(tempDir, "unsupported.json")
	if err := os.WriteFile(unsupportedFile, []byte(`{"schema_version": 3, "target_system": "sys"}`), 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"export", "--index", unsupportedFile}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for unsupported schema, got nil")
	}
	if !strings.Contains(err.Error(), "index_schema_unsupported") {
		t.Errorf("expected error to cite index_schema_unsupported, got: %v", err)
	}

	// 4. Missing file via explicit --index: returns index_missing
	missingFile := filepath.Join(tempDir, "nonexistent.json")
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"export", "--index", missingFile}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for missing index, got nil")
	}
	if !strings.Contains(err.Error(), "index_missing") {
		t.Errorf("expected error to cite index_missing, got: %v", err)
	}
}

// AC4: Com gira.yaml, legado nunca é carregado implicitamente.
func TestCLI_Resolver_OrphanLegacyNeverLoadedWithManifest(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	storeRoot := filepath.Join(tempHome, ".cache", "gira", "indexes")
	legacyDir := filepath.Join(storeRoot, "test-sys")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write orphan legacy macro-map.json without .head
	legacyJSON := `{"schema_version": "1.0.0", "target_system": "test-sys", "nodes": [{"id": "service:orphan", "kind": "Service", "name": "orphan"}]}`
	if err := os.WriteFile(filepath.Join(legacyDir, "macro-map.json"), []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	projDir := t.TempDir()
	manifestYAML := `version: 1
target_system:
  name: test-sys
repositories:
  - id: repo
    url: "https://example.com/repo.git"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(filepath.Join(projDir, "gira.yaml"), []byte(manifestYAML), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(projDir)

	var stdout, stderr bytes.Buffer
	err := Run([]string{"export"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected export to fail with index_missing when only orphan legacy index exists, got nil")
	}
	if !strings.Contains(err.Error(), "index_missing") || !strings.Contains(err.Error(), "gira index") {
		t.Errorf("expected error mentioning index_missing and instructing 'gira index', got: %v", err)
	}
}

// AC5: Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch;
// descoberta ambígua retorna index_ambiguous.
func TestCLI_Resolver_ToolMismatchAndAmbiguousDiscovery(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	storeRoot := filepath.Join(tempHome, ".cache", "gira", "indexes")
	st := store.NewStore(storeRoot)

	projDir := t.TempDir()
	manifestYAML := `version: 1
target_system:
  name: tool-sys
repositories:
  - id: repo
    url: "https://example.com/repo.git"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(filepath.Join(projDir, "gira.yaml"), []byte(manifestYAML), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(projDir)

	// Publish version with explicit foreign tool_id
	mm := validTestSchema2Map("tool-sys")
	meta := &store.VersionMeta{
		Format:        "gira.index.version/v1",
		SchemaVersion: 2,
		VersionID:     "v-foreign",
		ToolID:        "foreign-tool-id-9999",
		TargetSystem:  "tool-sys",
	}
	_, err := st.PublishVersion(nil, "tool-sys", "v-foreign", mm, meta, store.PublishVersionOptions{
		NoLock: true,
	})
	if err != nil {
		t.Fatalf("PublishVersion failed: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = Run([]string{"export"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected export to fail with index_tool_mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "index_tool_mismatch") {
		t.Errorf("expected error citing index_tool_mismatch, got: %v", err)
	}
}
