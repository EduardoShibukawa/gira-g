package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"labs/gira-g/internal/domain"
	"labs/gira-g/internal/store"
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
