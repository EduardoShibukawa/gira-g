package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

func createTestMacroMap(targetSystem string) *domain.MacroMap {
	mm := domain.NewMacroMap(targetSystem)
	_ = mm.AddNode(domain.Node{
		ID:   "service:auth",
		Kind: domain.NodeService,
		Name: "auth-service",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-auth",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "cmd/server/main.go",
					StartLine: 1,
					EndLine:   20,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})
	_ = mm.AddNode(domain.Node{
		ID:   "contract:auth-grpc",
		Kind: domain.NodeContract,
		Name: "AuthService",
	})
	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:auth-provides-auth-grpc",
		SourceID: "service:auth",
		TargetID: "contract:auth-grpc",
		Verb:     domain.VerbProvides,
	})
	return mm
}

func TestStore_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)

	orig := createTestMacroMap("my-system")
	if err := s.Save("my-system", orig); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := s.Load("my-system")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.TargetSystem != orig.TargetSystem {
		t.Errorf("expected target system %q, got %q", orig.TargetSystem, loaded.TargetSystem)
	}
	if len(loaded.Nodes) != len(orig.Nodes) {
		t.Errorf("expected %d nodes, got %d", len(orig.Nodes), len(loaded.Nodes))
	}
	if len(loaded.Relationships) != len(orig.Relationships) {
		t.Errorf("expected %d relationships, got %d", len(orig.Relationships), len(loaded.Relationships))
	}
}

func TestStore_LoadFile(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)

	orig := createTestMacroMap("custom-system")
	customPath := filepath.Join(tempDir, "custom", "test-map.json")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	data, err := orig.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if err := os.WriteFile(customPath, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := s.LoadFile(customPath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if loaded.TargetSystem != "custom-system" {
		t.Errorf("expected TargetSystem 'custom-system', got %q", loaded.TargetSystem)
	}
}

func TestStore_List(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)

	systems, err := s.List()
	if err != nil {
		t.Fatalf("List on empty dir failed: %v", err)
	}
	if len(systems) != 0 {
		t.Errorf("expected 0 systems, got %d", len(systems))
	}

	mm1 := createTestMacroMap("sys-b")
	mm2 := createTestMacroMap("sys-a")
	if err := s.Save("sys-b", mm1); err != nil {
		t.Fatalf("Save sys-b failed: %v", err)
	}
	if err := s.Save("sys-a", mm2); err != nil {
		t.Fatalf("Save sys-a failed: %v", err)
	}

	// Create an empty dir without macro-map.json
	_ = os.MkdirAll(filepath.Join(tempDir, "empty-dir"), 0755)

	systems, err = s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(systems) != 2 {
		t.Fatalf("expected 2 systems, got %d: %v", len(systems), systems)
	}
	if systems[0] != "sys-a" || systems[1] != "sys-b" {
		t.Errorf("expected sorted systems [sys-a sys-b], got %v", systems)
	}
}

func TestStore_Errors(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)

	// Empty targetSystem on save
	if err := s.Save("", nil); err == nil {
		t.Error("expected error for empty targetSystem, got nil")
	}

	// Nil MacroMap on save
	if err := s.Save("sys", nil); err == nil {
		t.Error("expected error for nil MacroMap, got nil")
	}

	// Nonexistent system on load
	if _, err := s.Load("nonexistent"); err == nil {
		t.Error("expected error loading nonexistent system, got nil")
	}

	// Invalid JSON on LoadFile
	badFile := filepath.Join(tempDir, "bad.json")
	if err := os.WriteFile(badFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write bad file: %v", err)
	}
	if _, err := s.LoadFile(badFile); err == nil {
		t.Error("expected error loading invalid JSON, got nil")
	}

	// Nonexistent file on LoadFile
	if _, err := s.LoadFile(filepath.Join(tempDir, "missing.json")); err == nil {
		t.Error("expected error loading missing file, got nil")
	}
}

func TestStore_PublishAtomicPreservesPreviousAndCleansTmp(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)
	targetPath := filepath.Join(tempDir, "sys", "macro-map.json")

	initial := createTestMacroMap("sys")
	if _, err := s.PublishFile(targetPath, initial, store.PublishOptions{}); err != nil {
		t.Fatalf("initial PublishFile failed: %v", err)
	}
	beforeBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Simulate failure right before rename
	updated := createTestMacroMap("sys")
	_ = updated.AddNode(domain.Node{
		ID:   "service:billing",
		Kind: domain.NodeService,
		Name: "billing",
	})
	_, err = s.PublishFile(targetPath, updated, store.PublishOptions{
		BeforeRenameError: func() error { return os.ErrPermission },
	})
	if err == nil {
		t.Fatalf("expected BeforeRenameError to fail PublishFile")
	}

	// Previous index must be untouched
	afterBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile after failed publish failed: %v", err)
	}
	if string(beforeBytes) != string(afterBytes) {
		t.Fatalf("expected previous index to remain byte-identical after failed publish")
	}

	// Temporary files must be cleaned up
	entries, err := os.ReadDir(filepath.Dir(targetPath))
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "macro-map.json" {
			t.Errorf("unexpected leftover temporary file after failed publish: %s", e.Name())
		}
	}

	// Simulate parent directory Sync() failure after os.Rename
	var syncedDir string
	_, err = s.PublishFile(targetPath, updated, store.PublishOptions{
		SyncDir: func(dir string) error {
			syncedDir = dir
			return os.ErrInvalid
		},
	})
	if err == nil {
		t.Fatalf("expected SyncDir error to fail PublishFile")
	}
	if syncedDir != filepath.Dir(targetPath) {
		t.Errorf("expected SyncDir called with %q, got %q", filepath.Dir(targetPath), syncedDir)
	}
}

func TestStore_CorruptIndexPreservedUntilRecoverCorrupt(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)
	targetPath := filepath.Join(tempDir, "sys", "macro-map.json")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		t.Fatal(err)
	}

	corruptContent := []byte("{corrupt-json-do-not-overwrite")
	if err := os.WriteFile(targetPath, corruptContent, 0644); err != nil {
		t.Fatal(err)
	}

	validMap := createTestMacroMap("sys")

	// 1. Default PublishFile, Save, and SaveFile must refuse to overwrite corrupt index and preserve its bytes
	if _, err := s.PublishFile(targetPath, validMap, store.PublishOptions{}); err == nil {
		t.Fatalf("expected error when existing index is corrupt and RecoverCorrupt=false")
	}
	if err := s.Save("sys", validMap); err == nil {
		t.Fatalf("expected Save to refuse overwriting corrupt index")
	}
	if err := s.SaveFile(targetPath, validMap); err == nil {
		t.Fatalf("expected SaveFile to refuse overwriting corrupt index")
	}
	gotBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotBytes) != string(corruptContent) {
		t.Fatalf("expected corrupt file bytes to be preserved, got %q", string(gotBytes))
	}

	// 2. Deliberate recovery with RecoverCorrupt=true succeeds
	_, err = s.PublishFile(targetPath, validMap, store.PublishOptions{RecoverCorrupt: true})
	if err != nil {
		t.Fatalf("expected PublishFile with RecoverCorrupt=true to succeed, got %v", err)
	}
	loaded, err := s.LoadFile(targetPath)
	if err != nil {
		t.Fatalf("expected recovered index to load cleanly, got %v", err)
	}
	if loaded.TargetSystem != "sys" {
		t.Errorf("expected TargetSystem 'sys', got %q", loaded.TargetSystem)
	}
}

func TestStore_PublishEnforcesSymbolLossGateAndPreservesPrevious(t *testing.T) {
	tempDir := t.TempDir()
	s := store.NewStore(tempDir)
	targetPath := filepath.Join(tempDir, "sys", "macro-map.json")

	initial := createTestMacroMap("sys")
	if err := s.Save("sys", initial); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}
	beforeBytes, _ := os.ReadFile(targetPath)

	// Next map drops service:auth (which has real extracted evidence) and contract:auth-grpc.
	// Save, SaveFile, and PublishFile (with default PublishOptions) must ALL block and preserve previous index.
	shrunk := domain.NewMacroMap("sys")
	if err := s.Save("sys", shrunk); err == nil {
		t.Fatalf("expected Save to enforce symbol-loss gate and fail when real extracted node disappears")
	}
	if err := s.SaveFile(targetPath, shrunk); err == nil {
		t.Fatalf("expected SaveFile to enforce symbol-loss gate and fail when real extracted node disappears")
	}
	if _, err := s.PublishFile(targetPath, shrunk, store.PublishOptions{}); err == nil {
		t.Fatalf("expected PublishFile to enforce symbol-loss gate and fail when real extracted node disappears")
	}

	// Previous index must be preserved byte-for-byte
	afterBytes, _ := os.ReadFile(targetPath)
	if string(beforeBytes) != string(afterBytes) {
		t.Fatalf("expected previous index to be preserved when symbol-loss gate blocks publication")
	}

	// With explicit AllowSymbolLoss: true on PublishFile, publication succeeds
	rep, err := s.PublishFile(targetPath, shrunk, store.PublishOptions{AllowSymbolLoss: true})
	if err != nil {
		t.Fatalf("expected PublishFile with AllowSymbolLoss=true to succeed, got %v", err)
	}
	if !rep.AllowedByShrink {
		t.Errorf("expected AllowedByShrink=true")
	}
}

