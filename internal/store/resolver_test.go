package store_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

func validSchema2MacroMap(targetSystem string) *domain.MacroMap {
	si := &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		Extractors: []domain.ExtractorIdentity{
			{ID: "go", Version: "1.2.0"},
			{ID: "proto", Version: "1.0.0"},
		},
		Sources: []domain.SourceRepoIdentity{
			{RepoID: "repo-core", CommitSHA: "1111111111111111111111111111111111111111", SHA: "1111111111111111111111111111111111111111", URLSHA256: "hash-repo-core"},
		},
	}
	mm := domain.NewMacroMapV2(targetSystem, si)
	_ = mm.AddNode(domain.Node{
		ID:   "service:core",
		Kind: domain.NodeService,
		Name: "Core Service",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-core",
					CommitSHA: "1111111111111111111111111111111111111111",
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

// AC1: Schema 2 válido é aceito e source_identity preservado.
func TestResolver_AC1_Schema2AcceptedAndSourceIdentityPreserved(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "schema2.json")

	mm := validSchema2MacroMap("sys-ac1")
	canonicalBytes, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if err := os.WriteFile(indexPath, canonicalBytes, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	st := store.NewStore(tempDir)
	loaded, err := st.LoadFileStrict(indexPath)
	if err != nil {
		t.Fatalf("LoadFileStrict failed on valid Schema 2: %v", err)
	}

	// Verify SchemaVersion is integer 2
	if !store.IsSchema2(loaded) {
		t.Fatalf("expected IsSchema2 to be true, got SchemaVersion=%v (%T)", loaded.SchemaVersion, loaded.SchemaVersion)
	}

	// Verify source_identity is completely preserved
	if loaded.SourceIdentity == nil {
		t.Fatalf("expected source_identity to be preserved, got nil")
	}
	if loaded.SourceIdentity.SchemaVersion != 2 {
		t.Errorf("expected source_identity.schema_version 2, got %d", loaded.SourceIdentity.SchemaVersion)
	}
	if loaded.SourceIdentity.ManifestSHA256 != "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" {
		t.Errorf("unexpected ManifestSHA256: %s", loaded.SourceIdentity.ManifestSHA256)
	}
	if len(loaded.SourceIdentity.Extractors) != 2 || loaded.SourceIdentity.Extractors[0].ID != "go" {
		t.Errorf("unexpected extractors: %+v", loaded.SourceIdentity.Extractors)
	}
	if len(loaded.SourceIdentity.Sources) != 1 || loaded.SourceIdentity.Sources[0].RepoID != "repo-core" {
		t.Errorf("unexpected sources: %+v", loaded.SourceIdentity.Sources)
	}

	// Re-serialize and check bytes
	reserialized, err := loaded.CanonicalJSON()
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}
	if !bytes.Equal(canonicalBytes, reserialized) {
		t.Errorf("expected re-serialization to be byte-identical:\nGOT:\n%s\nWANT:\n%s", string(reserialized), string(canonicalBytes))
	}
}

// AC2: Schema desconhecido, tipo incorreto e campos obrigatórios ausentes retornam index_schema_unsupported;
// corrupção retorna index_corrupt.
func TestResolver_AC2_SchemaValidation_UnsupportedAndCorrupt(t *testing.T) {
	st := store.NewStore("")

	t.Run("index_schema_unsupported_cases", func(t *testing.T) {
		cases := []struct {
			name string
			json string
		}{
			{"unknown_integer_3", `{"schema_version": 3, "target_system": "sys", "source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"repo_id": "r", "url_sha256": "h"}]}}`},
			{"unknown_integer_0", `{"schema_version": 0, "target_system": "sys"}`},
			{"unknown_string_2_0_0", `{"schema_version": "2.0.0", "target_system": "sys"}`},
			{"unknown_string_9_9_9", `{"schema_version": "9.9.9", "target_system": "sys"}`},
			{"float_schema_2_5", `{"schema_version": 2.5, "target_system": "sys"}`},
			{"boolean_schema_true", `{"schema_version": true, "target_system": "sys"}`},
			{"array_schema", `{"schema_version": [2], "target_system": "sys"}`},
			{"object_schema", `{"schema_version": {"ver": 2}, "target_system": "sys"}`},
			{"missing_schema_version", `{"target_system": "sys"}`},
			{"missing_target_system", `{"schema_version": 2, "source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"repo_id": "r", "url_sha256": "h"}]}}`},
			{"schema2_missing_source_identity", `{"schema_version": 2, "target_system": "sys"}`},
			{"schema2_source_identity_wrong_ver", `{"schema_version": 2, "target_system": "sys", "source_identity": {"schema_version": 1, "manifest_sha256": "abc", "sources": [{"repo_id": "r", "url_sha256": "h"}]}}`},
			{"schema2_source_identity_missing_manifest_sha", `{"schema_version": 2, "target_system": "sys", "source_identity": {"schema_version": 2, "manifest_sha256": "", "sources": [{"repo_id": "r", "url_sha256": "h"}]}}`},
			{"schema2_source_missing_repo_id", `{"schema_version": 2, "target_system": "sys", "source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"url_sha256": "h"}]}}`},
			{"schema2_source_missing_url_sha256", `{"schema_version": 2, "target_system": "sys", "source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"repo_id": "r"}]}}`},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				tempDir := t.TempDir()
				f := filepath.Join(tempDir, "test.json")
				if err := os.WriteFile(f, []byte(tc.json), 0644); err != nil {
					t.Fatal(err)
				}
				_, err := st.LoadFileWithOptions(f, store.LoadOptions{AllowLegacy: true})
				if err == nil {
					t.Fatalf("expected error for %s, got nil", tc.name)
				}
				if !errors.Is(err, store.ErrUnsupportedSchema) {
					t.Errorf("[%s] expected error to wrap ErrUnsupportedSchema, got: %v", tc.name, err)
				}
				if !strings.Contains(err.Error(), "index_schema_unsupported") {
					t.Errorf("[%s] expected error string to contain 'index_schema_unsupported', got: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("index_corrupt_cases", func(t *testing.T) {
		cases := []struct {
			name string
			raw  []byte
		}{
			{"empty_file", []byte("")},
			{"whitespace_only", []byte("   \n\t  ")},
			{"malformed_syntax", []byte(`{ "schema_version": 2, `)},
			{"syntax_error_token", []byte(`{ "schema_version": 2 target_system: "foo" }`)},
			{"broken_graph_missing_endpoint", []byte(`{
				"schema_version": 2,
				"target_system": "sys",
				"source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"repo_id": "r", "url_sha256": "h"}]},
				"nodes": [{"id": "service:a", "kind": "Service", "name": "a"}],
				"relationships": [{"id": "rel:1", "source_id": "service:a", "target_id": "service:MISSING", "verb": "consumes"}]
			}`)},
			{"broken_graph_duplicate_nodes", []byte(`{
				"schema_version": 2,
				"target_system": "sys",
				"source_identity": {"schema_version": 2, "manifest_sha256": "abc", "sources": [{"repo_id": "r", "url_sha256": "h"}]},
				"nodes": [{"id": "service:a", "kind": "Service", "name": "a"}, {"id": "service:a", "kind": "Service", "name": "a2"}]
			}`)},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				tempDir := t.TempDir()
				f := filepath.Join(tempDir, "corrupt.json")
				if err := os.WriteFile(f, tc.raw, 0644); err != nil {
					t.Fatal(err)
				}
				_, err := st.LoadFileWithOptions(f, store.LoadOptions{AllowLegacy: true})
				if err == nil {
					t.Fatalf("expected error for %s, got nil", tc.name)
				}
				if !errors.Is(err, store.ErrCorruptIndex) {
					t.Errorf("[%s] expected error to wrap ErrCorruptIndex, got: %v", tc.name, err)
				}
				if !strings.Contains(err.Error(), "index_corrupt") {
					t.Errorf("[%s] expected error string to contain 'index_corrupt', got: %v", tc.name, err)
				}
			})
		}
	})
}

// AC3: Legado v1 só via --index explícito, read-only e com aviso.
func TestResolver_AC3_LegacySchema1OnlyExplicitWithWarning(t *testing.T) {
	tempDir := t.TempDir()
	legacyFile := filepath.Join(tempDir, "legacy.json")
	legacyJSON := `{
		"schema_version": "1.0.0",
		"target_system": "legacy-sys",
		"nodes": [{"id": "service:leg", "kind": "Service", "name": "Legacy"}],
		"relationships": []
	}`
	if err := os.WriteFile(legacyFile, []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	st := store.NewStore(tempDir)

	// Attempting to load legacy in strict mode (without explicit --index authorization) fails closed
	_, err := st.LoadFileStrict(legacyFile)
	if err == nil {
		t.Fatalf("expected error loading legacy in strict mode, got nil")
	}
	if !errors.Is(err, store.ErrUnsupportedSchema) {
		t.Errorf("expected ErrUnsupportedSchema when legacy loaded strictly, got: %v", err)
	}

	// Loading legacy with explicit --index authorization succeeds and emits deprecation warning
	var warnBuf bytes.Buffer
	mm, err := st.LoadFileWithOptions(legacyFile, store.LoadOptions{
		AllowLegacy: true,
		WarnWriter:  &warnBuf,
	})
	if err != nil {
		t.Fatalf("expected legacy to load successfully when authorized: %v", err)
	}
	if !store.IsSchema1(mm) {
		t.Errorf("expected loaded index to be Schema 1")
	}
	warnOut := warnBuf.String()
	if !strings.Contains(warnOut, "warning:") || !strings.Contains(warnOut, "1.0.0") || !strings.Contains(warnOut, "deprecated") {
		t.Errorf("expected visible deprecation warning mentioning schema 1.0.0, got: %s", warnOut)
	}
}

// AC4: Com gira.yaml, legado nunca é carregado implicitamente.
func TestResolver_AC4_NeverLoadLegacyImplicitlyWithManifest(t *testing.T) {
	tempDir := t.TempDir()
	storeRoot := filepath.Join(tempDir, "store")
	projDir := filepath.Join(tempDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create gira.yaml for "shop-system"
	manifestPath := filepath.Join(projDir, "gira.yaml")
	manifestContent := `version: 1
target_system:
  name: shop-system
repositories:
  - id: shop
    url: "https://example.com/shop.git"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Place an orphan legacy macro-map.json in the store WITHOUT .head
	legacyDir := filepath.Join(storeRoot, "shop-system")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}
	legacyJSON := `{"schema_version": "1.0.0", "target_system": "shop-system", "nodes": [{"id": "service:old", "kind": "Service", "name": "old"}]}`
	if err := os.WriteFile(filepath.Join(legacyDir, "macro-map.json"), []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Resolve with gira.yaml present: MUST NOT load legacy index! Must return index_missing.
	_, err := store.Resolve(context.Background(), store.ResolveOptions{
		ManifestPath: manifestPath,
		ProjectRoot:  projDir,
		StoreRoots:   []string{storeRoot},
	})
	if err == nil {
		t.Fatalf("expected error when resolving shop-system with only orphan legacy index, got nil")
	}
	if !errors.Is(err, store.ErrIndexMissing) {
		t.Errorf("expected ErrIndexMissing for orphan legacy with manifest, got: %v", err)
	}
	if !strings.Contains(err.Error(), "index_missing") || !strings.Contains(err.Error(), "gira index") {
		t.Errorf("expected error instructing to run 'gira index', got: %v", err)
	}

	// Now publish a valid versioned head
	st := store.NewStore(storeRoot)
	vMM := validSchema2MacroMap("shop-system")
	_, err = st.PublishVersion(context.Background(), "shop-system", "v-1", vMM, nil, store.PublishVersionOptions{
		NoLock: true,
	})
	if err != nil {
		t.Fatalf("PublishVersion failed: %v", err)
	}

	// Now resolve succeeds and loads the active version
	res, err := store.Resolve(context.Background(), store.ResolveOptions{
		ManifestPath: manifestPath,
		ProjectRoot:  projDir,
		StoreRoots:   []string{storeRoot},
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed once versioned head exists: %v", err)
	}
	if res.VersionID != "v-1" {
		t.Errorf("expected version v-1, got %s", res.VersionID)
	}
	if !store.IsSchema2(res.MacroMap) {
		t.Errorf("expected Schema 2 MacroMap")
	}
}

// AC5: Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch;
// descoberta ambígua retorna index_ambiguous.
func TestResolver_AC5_Taxonomy_Missing_Mismatch_Ambiguous(t *testing.T) {
	tempDir := t.TempDir()
	storeRoot := filepath.Join(tempDir, "store")
	projDir := filepath.Join(tempDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(projDir, "gira.yaml")
manifestContent := `version: 1
target_system:
  name: my-sys
repositories:
  - id: my-repo
    url: "https://example.com/repo.git"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Case 5a: Missing version / lineage -> index_missing
	_, err := store.Resolve(context.Background(), store.ResolveOptions{
		ManifestPath: manifestPath,
		ProjectRoot:  projDir,
		StoreRoots:   []string{storeRoot},
	})
	if err == nil || !errors.Is(err, store.ErrIndexMissing) || !strings.Contains(err.Error(), "index_missing") {
		t.Fatalf("expected index_missing for nonexistent lineage, got: %v", err)
	}

	// Case 5b: Tool mismatch -> index_tool_mismatch
	st := store.NewStore(storeRoot)
	vMM := validSchema2MacroMap("my-sys")
	meta := &store.VersionMeta{
		Format:        "gira.index.version/v1",
		SchemaVersion: 2,
		VersionID:     "v-tool-1",
		ToolID:        "tool-sha-alpha",
		TargetSystem:  "my-sys",
	}
	_, err = st.PublishVersion(context.Background(), "my-sys", "v-tool-1", vMM, meta, store.PublishVersionOptions{
		NoLock: true,
	})
	if err != nil {
		t.Fatalf("PublishVersion failed: %v", err)
	}

	// Resolving with divergent expected tool ID fails closed with index_tool_mismatch
	_, err = store.Resolve(context.Background(), store.ResolveOptions{
		ManifestPath:   manifestPath,
		ProjectRoot:    projDir,
		StoreRoots:     []string{storeRoot},
		ExpectedToolID: "tool-sha-beta",
	})
	if err == nil || !errors.Is(err, store.ErrToolMismatch) || !strings.Contains(err.Error(), "index_tool_mismatch") {
		t.Fatalf("expected index_tool_mismatch when tool divergent, got: %v", err)
	}

	// Resolving with matching tool ID succeeds
	res, err := store.Resolve(context.Background(), store.ResolveOptions{
		ManifestPath:   manifestPath,
		ProjectRoot:    projDir,
		StoreRoots:     []string{storeRoot},
		ExpectedToolID: "tool-sha-alpha",
	})
	if err != nil {
		t.Fatalf("expected resolve with matching tool ID to succeed, got: %v", err)
	}
	if res.VersionID != "v-tool-1" {
		t.Errorf("expected version v-tool-1, got %s", res.VersionID)
	}

	// Case 5c: Ambiguous discovery -> index_ambiguous
	// Publish another system into store
	mm2 := validSchema2MacroMap("second-sys")
	_, err = st.PublishVersion(context.Background(), "second-sys", "v-tool-2", mm2, nil, store.PublishVersionOptions{
		NoLock: true,
	})
	if err != nil {
		t.Fatalf("PublishVersion second-sys failed: %v", err)
	}

	// Discovery without manifest when multiple systems exist in cache
	_, err = store.Resolve(context.Background(), store.ResolveOptions{
		StoreRoots: []string{storeRoot},
	})
	if err == nil || !errors.Is(err, store.ErrIndexAmbiguous) || !strings.Contains(err.Error(), "index_ambiguous") {
		t.Fatalf("expected index_ambiguous for discovery with multiple systems, got: %v", err)
	}
	if !strings.Contains(err.Error(), "ambiguous index: multiple systems found in cache") {
		t.Errorf("expected error message to retain backwards-compatible phrasing, got: %v", err)
	}
}

// AC6: Binário/fluxo antigo não pode republicar schema 2 como schema 1 sem erro e sem alterar bytes.
func TestResolver_AC6_DowngradePrevention(t *testing.T) {
	tempDir := t.TempDir()
	snapshotPath := filepath.Join(tempDir, "snapshot.json")

	// 1. Establish valid Schema 2 snapshot on disk
	mm2 := validSchema2MacroMap("sys-ac6")
	st := store.NewStore(tempDir)
	_, err := st.PublishSnapshot(context.Background(), snapshotPath, mm2, store.PublishSnapshotOptions{
		NoLock: true,
	})
	if err != nil {
		t.Fatalf("initial PublishSnapshot failed: %v", err)
	}
	bytesBefore, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Old binary / legacy flow attempts to overwrite with Schema 1
	legacyMM := domain.NewMacroMap("sys-ac6")
	_ = legacyMM.AddNode(domain.Node{ID: "service:old", Kind: domain.NodeService, Name: "old"})

	// Attempt via PublishFile
	_, err = st.PublishFile(snapshotPath, legacyMM, store.PublishOptions{})
	if err == nil {
		t.Fatalf("expected PublishFile to refuse schema downgrade, got nil")
	}
	if !errors.Is(err, store.ErrSchemaDowngrade) && !strings.Contains(err.Error(), "downgrade") {
		t.Errorf("expected error citing schema downgrade, got: %v", err)
	}

	// Attempt via PublishSnapshot
	_, err = st.PublishSnapshot(context.Background(), snapshotPath, legacyMM, store.PublishSnapshotOptions{
		NoLock: true,
	})
	if err == nil {
		t.Fatalf("expected PublishSnapshot to refuse schema downgrade, got nil")
	}
	if !errors.Is(err, store.ErrSchemaDowngrade) && !strings.Contains(err.Error(), "downgrade") {
		t.Errorf("expected error citing schema downgrade, got: %v", err)
	}

	// 3. Confirm file on disk is 100% byte-for-byte unchanged
	bytesAfter, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytesBefore, bytesAfter) {
		t.Fatalf("file content changed after rejected downgrade attempt!")
	}
}
