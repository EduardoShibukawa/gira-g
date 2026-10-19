package store_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

func makeTestMacroMap(targetSystem string) *domain.MacroMap {
	mm := domain.NewMacroMap(targetSystem)
	mm.SchemaVersion = 2
	_ = mm.AddNode(domain.Node{
		ID:   "service:api",
		Kind: domain.NodeService,
		Name: "api",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-api",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/api/main.go",
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
		ID:   "service:database",
		Kind: domain.NodeService,
		Name: "database",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-db",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/db/schema.sql",
					StartLine: 1,
					EndLine:   50,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})
	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:api-db",
		SourceID: "service:api",
		TargetID: "service:database",
		Verb:     domain.VerbAccesses,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-api",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/api/client.go",
					StartLine: 5,
					EndLine:   10,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})
	return mm
}

// AC1: Publicação cria versão imutável e atualiza head somente após versão completa.
func TestPublish_AC1_ImmutableVersionAndHeadUpdate(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac1"
	srcKey := "src-key-1"
	toolID := "tool-id-1"
	versionID := store.ComputeVersionID(srcKey, toolID)

	mm := makeTestMacroMap("sys-ac1")
	meta := &store.VersionMeta{
		Format:       "gira.index.version/v1",
		VersionID:    versionID,
		SourceKey:    srcKey,
		ToolID:       toolID,
		TargetSystem: mm.TargetSystem,
	}

	res, err := st.PublishVersion(context.Background(), lineageID, versionID, mm, meta, store.PublishVersionOptions{})
	if err != nil {
		t.Fatalf("PublishVersion failed: %v", err)
	}
	if res.VersionID != versionID {
		t.Errorf("expected res.VersionID %q, got %q", versionID, res.VersionID)
	}
	if res.LineageID != lineageID {
		t.Errorf("expected res.LineageID %q, got %q", lineageID, res.LineageID)
	}

	// Verify on-disk canonical layout:
	// <store_root>/<lineage_id>/versions/<version_id>/macro-map.json
	// <store_root>/<lineage_id>/versions/<version_id>/meta.json
	// <store_root>/<lineage_id>/.head
	versionDir := filepath.Join(tempDir, lineageID, "versions", versionID)
	mapPath := filepath.Join(versionDir, "macro-map.json")
	metaPath := filepath.Join(versionDir, "meta.json")
	headPath := filepath.Join(tempDir, lineageID, ".head")

	if _, err := os.Stat(mapPath); err != nil {
		t.Fatalf("expected %s to exist: %v", mapPath, err)
	}
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("expected %s to exist: %v", metaPath, err)
	}
	if _, err := os.Stat(headPath); err != nil {
		t.Fatalf("expected %s to exist: %v", headPath, err)
	}

	head, err := st.ReadHead(lineageID)
	if err != nil {
		t.Fatalf("ReadHead failed: %v", err)
	}
	if head != versionID {
		t.Fatalf("expected .head %q, got %q", versionID, head)
	}

	loadedMM, loadedMeta, err := st.LoadActiveVersion(lineageID)
	if err != nil {
		t.Fatalf("LoadActiveVersion failed: %v", err)
	}
	if loadedMM.TargetSystem != mm.TargetSystem {
		t.Errorf("loaded target system %q != %q", loadedMM.TargetSystem, mm.TargetSystem)
	}
	if loadedMeta.VersionID != versionID {
		t.Errorf("loaded meta VersionID %q != %q", loadedMeta.VersionID, versionID)
	}
	if loadedMeta.MapSHA256 == "" {
		t.Errorf("expected non-empty MapSHA256 in meta.json")
	}

	lineages, err := st.ListLineages()
	if err != nil || len(lineages) != 1 || lineages[0] != lineageID {
		t.Errorf("expected ListLineages [%s], got %v (err: %v)", lineageID, lineages, err)
	}
	versions, err := st.ListVersions(lineageID)
	if err != nil || len(versions) != 1 || versions[0] != versionID {
		t.Errorf("expected ListVersions [%s], got %v (err: %v)", versionID, versions, err)
	}
}

// AC2: Falha antes do rename deixa nenhum artefato novo visível.
func TestPublish_AC2_FailureBeforeRenameLeavesNoArtifacts(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac2"
	srcKey := "src-key-2"
	toolID := "tool-id-2"
	versionID := store.ComputeVersionID(srcKey, toolID)

	mm := makeTestMacroMap("sys-ac2")
	injectedErr := errors.New("simulated crash before version promotion")

	_, err := st.PublishVersion(context.Background(), lineageID, versionID, mm, nil, store.PublishVersionOptions{
		SourceKey: srcKey,
		ToolID:    toolID,
		BeforeRenameError: func() error {
			return injectedErr
		},
	})
	if err == nil {
		t.Fatalf("expected PublishVersion to fail with BeforeRenameError")
	}
	if !strings.Contains(err.Error(), injectedErr.Error()) {
		t.Fatalf("expected error to contain %q, got: %v", injectedErr.Error(), err)
	}

	// Verify no new version directory or head exists:
	versionDir := filepath.Join(tempDir, lineageID, "versions", versionID)
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Errorf("expected %s not to exist, got stat err: %v", versionDir, err)
	}

	headPath := filepath.Join(tempDir, lineageID, ".head")
	if _, err := os.Stat(headPath); !os.IsNotExist(err) {
		t.Errorf("expected %s not to exist, got stat err: %v", headPath, err)
	}

	// Verify no staging directory was leaked:
	versionsDir := filepath.Join(tempDir, lineageID, "versions")
	if entries, err := os.ReadDir(versionsDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".staging") {
				t.Errorf("leaked staging directory: %s", e.Name())
			}
		}
	}
}

// AC3: Falha após rename da versão deixa head intacto e retry idempotente.
func TestPublish_AC3_FailureAfterVersionRenameLeavesHeadIntactAndRetryIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac3"

	v1ID := store.ComputeVersionID("src-3-1", "tool-3-1")
	v2ID := store.ComputeVersionID("src-3-2", "tool-3-2")

	mm1 := makeTestMacroMap("sys-ac3")
	_, err := st.PublishVersion(context.Background(), lineageID, v1ID, mm1, nil, store.PublishVersionOptions{
		SourceKey: "src-3-1",
		ToolID:    "tool-3-1",
	})
	if err != nil {
		t.Fatalf("initial publish v1 failed: %v", err)
	}
	head1, err := st.ReadHead(lineageID)
	if err != nil || head1 != v1ID {
		t.Fatalf("head after v1 publish = %q (err: %v), want %q", head1, err, v1ID)
	}

	// Now publish v2, but fail right before promoting .head:
	mm2 := makeTestMacroMap("sys-ac3")
	_ = mm2.AddNode(domain.Node{
		ID:   "service:notifications",
		Kind: domain.NodeService,
		Name: "notifications",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-notif",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/notif/main.go",
					StartLine: 1,
					EndLine:   30,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})
	injectedErr := errors.New("simulated power loss before head update")

	res, err := st.PublishVersion(context.Background(), lineageID, v2ID, mm2, nil, store.PublishVersionOptions{
		SourceKey: "src-3-2",
		ToolID:    "tool-3-2",
		BeforeHeadRenameError: func() error {
			return injectedErr
		},
	})
	if err == nil {
		t.Fatalf("expected error from BeforeHeadRenameError")
	}
	if res == nil || res.Status != store.DurabilityVisibleUnconfirmed {
		t.Errorf("expected visible_unconfirmed status on failure after promotion")
	}

	// Version v2 directory was promoted and exists:
	v2Dir := filepath.Join(tempDir, lineageID, "versions", v2ID)
	if _, err := os.Stat(filepath.Join(v2Dir, "macro-map.json")); err != nil {
		t.Fatalf("expected v2 macro-map.json to exist: %v", err)
	}

	// But .head MUST remain pointing to v1!
	headAfterCrash, err := st.ReadHead(lineageID)
	if err != nil || headAfterCrash != v1ID {
		t.Fatalf("head after crash = %q (err: %v), must remain %q", headAfterCrash, err, v1ID)
	}

	// Now retry publishing v2 without the injected error:
	retryRes, err := st.PublishVersion(context.Background(), lineageID, v2ID, mm2, nil, store.PublishVersionOptions{
		SourceKey: "src-3-2",
		ToolID:    "tool-3-2",
	})
	if err != nil {
		t.Fatalf("retry PublishVersion failed: %v", err)
	}
	if retryRes.VersionID != v2ID {
		t.Errorf("retry res.VersionID = %q, want %q", retryRes.VersionID, v2ID)
	}

	// Now .head points to v2!
	headAfterRetry, err := st.ReadHead(lineageID)
	if err != nil || headAfterRetry != v2ID {
		t.Fatalf("head after retry = %q, want %q", headAfterRetry, v2ID)
	}
	activeMM, _, err := st.LoadActiveVersion(lineageID)
	if err != nil {
		t.Fatalf("LoadActiveVersion after retry failed: %v", err)
	}
	if len(activeMM.Nodes) != 3 {
		t.Errorf("expected 3 nodes in active version, got %d", len(activeMM.Nodes))
	}
}

// AC4: Republicação do mesmo version_id com bytes diferentes falha nondeterministic_build sem sobrescrever.
func TestPublish_AC4_NonDeterministicBuildFailsClosedWithoutOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac4"
	srcKey := "src-key-4"
	toolID := "tool-id-4"
	versionID := store.ComputeVersionID(srcKey, toolID)

	mm1 := makeTestMacroMap("sys-ac4")
	_, err := st.PublishVersion(context.Background(), lineageID, versionID, mm1, nil, store.PublishVersionOptions{
		SourceKey: srcKey,
		ToolID:    toolID,
	})
	if err != nil {
		t.Fatalf("initial publish failed: %v", err)
	}

	vPath := filepath.Join(tempDir, lineageID, "versions", versionID, "macro-map.json")
	initialBytes, err := os.ReadFile(vPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Attempt republishing the same versionID with divergent content (extra node):
	mm2 := makeTestMacroMap("sys-ac4")
	_ = mm2.AddNode(domain.Node{
		ID:   "service:divergent",
		Kind: domain.NodeService,
		Name: "divergent",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-div",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/div/main.go",
					StartLine: 1,
					EndLine:   10,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})

	_, err = st.PublishVersion(context.Background(), lineageID, versionID, mm2, nil, store.PublishVersionOptions{
		SourceKey: srcKey,
		ToolID:    toolID,
	})
	if err == nil {
		t.Fatalf("expected republish with divergent bytes to fail")
	}
	if !errors.Is(err, store.ErrNonDeterministicBuild) {
		t.Errorf("expected error to wrap ErrNonDeterministicBuild, got: %v", err)
	}
	if !strings.Contains(err.Error(), "nondeterministic_build") {
		t.Errorf("expected error message to contain 'nondeterministic_build', got: %v", err)
	}

	// Verify the existing file was NOT overwritten and remains byte-for-byte identical:
	currentBytes, err := os.ReadFile(vPath)
	if err != nil {
		t.Fatalf("ReadFile after failed republish failed: %v", err)
	}
	if string(initialBytes) != string(currentBytes) {
		t.Fatalf("existing published file was overwritten or modified during failed nondeterministic republish!")
	}
}

// AC5: Gate rejeitado não altera store, head nem snapshot.
func TestPublish_AC5_GateRejectionLeavesStoreAndSnapshotUntouched(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac5"
	v1ID := store.ComputeVersionID("src-5-1", "tool-5-1")
	v2ID := store.ComputeVersionID("src-5-2", "tool-5-2")

	mm1 := makeTestMacroMap("sys-ac5")
	_, err := st.PublishVersion(context.Background(), lineageID, v1ID, mm1, nil, store.PublishVersionOptions{
		SourceKey: "src-5-1",
		ToolID:    "tool-5-1",
	})
	if err != nil {
		t.Fatalf("initial publish failed: %v", err)
	}

	// Attempt to publish v2 which drops service:database without AllowSymbolLoss:
	shrunkMM := domain.NewMacroMap("sys-ac5")
	shrunkMM.SchemaVersion = 2
	_ = shrunkMM.AddNode(domain.Node{
		ID:   "service:api",
		Kind: domain.NodeService,
		Name: "api",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-api",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/api/main.go",
					StartLine: 1,
					EndLine:   20,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})

	_, err = st.PublishVersion(context.Background(), lineageID, v2ID, shrunkMM, nil, store.PublishVersionOptions{
		SourceKey:       "src-5-2",
		ToolID:          "tool-5-2",
		AllowSymbolLoss: false,
	})
	if err == nil {
		t.Fatalf("expected gate rejection error when symbols are lost")
	}
	if !strings.Contains(err.Error(), "symbol loss gate rejected") {
		t.Errorf("expected error to cite symbol loss gate rejection, got: %v", err)
	}

	// Verify v2 was NOT written:
	v2Dir := filepath.Join(tempDir, lineageID, "versions", v2ID)
	if _, err := os.Stat(v2Dir); !os.IsNotExist(err) {
		t.Errorf("v2 directory %s should not exist after gate rejection", v2Dir)
	}

	// Verify .head remains pointing to v1:
	head, err := st.ReadHead(lineageID)
	if err != nil || head != v1ID {
		t.Errorf("expected head to remain %q, got %q (err: %v)", v1ID, head, err)
	}

	// Test snapshot gate rejection:
	snapPath := filepath.Join(tempDir, "snap.json")
	_, err = st.PublishSnapshot(context.Background(), snapPath, mm1, store.PublishSnapshotOptions{NoLock: true})
	if err != nil {
		t.Fatalf("PublishSnapshot initial failed: %v", err)
	}
	snapBefore, _ := os.ReadFile(snapPath)

	_, err = st.PublishSnapshot(context.Background(), snapPath, shrunkMM, store.PublishSnapshotOptions{
		NoLock:          true,
		AllowSymbolLoss: false,
	})
	if err == nil {
		t.Fatalf("expected snapshot overwrite to be rejected by gate")
	}
	snapAfter, _ := os.ReadFile(snapPath)
	if string(snapBefore) != string(snapAfter) {
		t.Errorf("snapshot was modified despite gate rejection!")
	}
}

// AC6: Unix sync failure retorna erro e visible_unconfirmed; Windows emite aviso explícito sem falsas garantias.
func TestPublish_AC6_DurabilitySemanticsUnixAndWindows(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac6"
	srcKey := "src-6"
	toolID := "tool-6"
	versionID := store.ComputeVersionID(srcKey, toolID)

	mm := makeTestMacroMap("sys-ac6")

	// Inject SyncDir failure:
	injectedSyncErr := errors.New("simulated fsync failure on parent directory")
	res, err := st.PublishVersion(context.Background(), lineageID, versionID, mm, nil, store.PublishVersionOptions{
		SourceKey: srcKey,
		ToolID:    toolID,
		SyncDir: func(dir string) error {
			return injectedSyncErr
		},
	})

	if runtime.GOOS == "windows" {
		// On Windows: directory fsync is omitted; returns visible_unconfirmed, exit code 0 / err == nil, and warning
		if err != nil {
			t.Errorf("Windows expected err == nil, got: %v", err)
		}
		if res == nil || res.Status != store.DurabilityVisibleUnconfirmed {
			t.Errorf("Windows expected status visible_unconfirmed, got: %v", res)
		}
		if res == nil || !strings.Contains(res.Warning, "visible_unconfirmed") {
			t.Errorf("Windows expected visible_unconfirmed warning, got: %v", res)
		}
	} else {
		// On Unix: directory fsync error returns non-nil error and visible_unconfirmed
		if err == nil {
			t.Fatalf("Unix expected non-nil error on sync failure")
		}
		if res == nil || res.Status != store.DurabilityVisibleUnconfirmed {
			t.Fatalf("Unix expected status visible_unconfirmed, got: %v", res)
		}
		if !strings.Contains(err.Error(), "visible_unconfirmed") {
			t.Errorf("expected error to state visible_unconfirmed, got: %v", err)
		}

		// The version is visible on disk even though sync was unconfirmed:
		head, headErr := st.ReadHead(lineageID)
		if headErr != nil || head != versionID {
			t.Errorf("version should be visible on disk, head=%q (err: %v)", head, headErr)
		}
	}

	// Verify standard successful publication produces confirmed status on Unix:
	if runtime.GOOS != "windows" {
		v2ID := store.ComputeVersionID("src-6-clean", "tool-6-clean")
		cleanRes, cleanErr := st.PublishVersion(context.Background(), lineageID, v2ID, mm, nil, store.PublishVersionOptions{
			SourceKey: "src-6-clean",
			ToolID:    "tool-6-clean",
		})
		if cleanErr != nil {
			t.Fatalf("clean publish failed: %v", cleanErr)
		}
		if cleanRes.Status != store.DurabilityConfirmed {
			t.Errorf("expected DurabilityConfirmed, got %q", cleanRes.Status)
		}
	}
}

// AC7: Snapshots destacados são independentes do store; arquivo corrompido existente é preservado salvo --recover-corrupt.
func TestPublish_AC7_DetachedSnapshotIndependenceAndCorruptPreservation(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	snapDir := filepath.Join(tempDir, "snapshots")
	snapPath := filepath.Join(snapDir, "detached.json")

	mm := makeTestMacroMap("sys-ac7")

	// 1. Publish snapshot: independent of store (store lineages remain empty)
	res, err := st.PublishSnapshot(context.Background(), snapPath, mm, store.PublishSnapshotOptions{NoLock: true})
	if err != nil {
		t.Fatalf("PublishSnapshot failed: %v", err)
	}
	if res.Status != store.DurabilityConfirmed && runtime.GOOS != "windows" {
		t.Errorf("expected DurabilityConfirmed on Unix, got %s", res.Status)
	}

	lineages, err := st.ListLineages()
	if err != nil || len(lineages) != 0 {
		t.Errorf("expected 0 store lineages from PublishSnapshot, got %v", lineages)
	}

	// 2. Corrupt existing snapshot: must fail closed and preserve corrupt file unless RecoverCorrupt is true
	corruptPath := filepath.Join(snapDir, "corrupted.json")
	badContent := []byte("{ truncated or corrupt json data ...")
	if err := os.WriteFile(corruptPath, badContent, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err = st.PublishSnapshot(context.Background(), corruptPath, mm, store.PublishSnapshotOptions{
		NoLock:         true,
		RecoverCorrupt: false,
	})
	if err == nil {
		t.Fatalf("expected PublishSnapshot to fail on corrupt existing file")
	}
	if !strings.Contains(err.Error(), "corrupted and will not be overwritten automatically") {
		t.Errorf("expected error citing corruption protection, got: %v", err)
	}

	// Verify corrupt file was preserved:
	corruptBytes, err := os.ReadFile(corruptPath)
	if err != nil || string(corruptBytes) != string(badContent) {
		t.Fatalf("corrupted file was modified or deleted!")
	}

	// 3. Pass RecoverCorrupt = true: now overwrite succeeds
	_, err = st.PublishSnapshot(context.Background(), corruptPath, mm, store.PublishSnapshotOptions{
		NoLock:         true,
		RecoverCorrupt: true,
	})
	if err != nil {
		t.Fatalf("PublishSnapshot with RecoverCorrupt failed: %v", err)
	}
	loaded, err := st.LoadFile(corruptPath)
	if err != nil {
		t.Fatalf("LoadFile after recovered snapshot failed: %v", err)
	}
	if loaded.TargetSystem != mm.TargetSystem {
		t.Errorf("loaded snapshot targetSystem = %q, want %q", loaded.TargetSystem, mm.TargetSystem)
	}
}

// AC8: CI 3 SOs e testes de falha passam: Concorrência com travas M -> S e recuperação de falhas.
func TestPublish_AC8_ConcurrentPublishingWithLocksAndD9LegacyFallback(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	lineageID := "lin-ac8"

	// Test D9: Legacy index fallback when .head does not exist
	legacyDir := filepath.Join(tempDir, "sys-legacy")
	_ = os.MkdirAll(legacyDir, 0755)
	legacyPath := filepath.Join(legacyDir, "macro-map.json")
	legacyMM := makeTestMacroMap("sys-legacy")
	_, err := st.PublishSnapshot(context.Background(), legacyPath, legacyMM, store.PublishSnapshotOptions{NoLock: true})
	if err != nil {
		t.Fatalf("writing legacy index failed: %v", err)
	}

	// Shrunk MM that drops a real extracted symbol against legacy baseline:
	shrunkMM := domain.NewMacroMap("sys-legacy")
	shrunkMM.SchemaVersion = 2
	_ = shrunkMM.AddNode(domain.Node{
		ID:   "service:api",
		Kind: domain.NodeService,
		Name: "api",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-api",
					CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
					Path:      "src/api/main.go",
					StartLine: 1,
					EndLine:   20,
				},
				Origin:       domain.OriginExtracted,
				Status:       domain.StatusCurrent,
				DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
			},
		},
	})

	// First canonical publication in lineageID: gate catches symbol loss against legacy index!
	shrunkID := store.ComputeVersionID("src-shrunk", "tool-shrunk")
	_, err = st.PublishVersion(context.Background(), lineageID, shrunkID, shrunkMM, nil, store.PublishVersionOptions{
		SourceKey:       "src-shrunk",
		ToolID:          "tool-shrunk",
		LegacyIndexPath: legacyPath,
		AllowSymbolLoss: false,
	})
	if err == nil {
		t.Fatalf("expected gate rejection against legacy index baseline")
	}

	// Test concurrency: 4 goroutines attempting to publish distinct versions under M -> S locks
	manifestFile := filepath.Join(tempDir, "gira.yaml")
	_ = os.WriteFile(manifestFile, []byte("target_system: {name: sys-concurrent}\n"), 0644)

	errCh := make(chan error, 4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			src := fmt.Sprintf("src-conc-%d", idx)
			tool := "tool-id"
			vID := store.ComputeVersionID(src, tool)

			mm := makeTestMacroMap("sys-concurrent")
			_ = mm.AddNode(domain.Node{
				ID:   fmt.Sprintf("service:worker-%d", idx),
				Kind: domain.NodeService,
				Name: fmt.Sprintf("worker-%d", idx),
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      "repo-worker",
							CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
							Path:      fmt.Sprintf("src/worker-%d/main.go", idx),
							StartLine: 1,
							EndLine:   10,
						},
						Origin:       domain.OriginExtracted,
						Status:       domain.StatusCurrent,
						DiscoveredAt: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC),
					},
				},
			})

			_, pubErr := st.PublishVersion(context.Background(), "lin-concurrent", vID, mm, nil, store.PublishVersionOptions{
				SourceKey:       src,
				ToolID:          tool,
				ManifestPath:    manifestFile,
				AllowSymbolLoss: true,
				LockTimeout:     10 * time.Second,
			})
			errCh <- pubErr
		}(i)
	}

	for i := 0; i < 4; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent publication failed: %v", err)
		}
	}

	// All 4 versions should exist in store without corruption:
	versions, err := st.ListVersions("lin-concurrent")
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 4 {
		t.Errorf("expected 4 published versions, got %d: %v", len(versions), versions)
	}

	activeHead, err := st.ReadHead("lin-concurrent")
	if err != nil || activeHead == "" {
		t.Errorf("expected active head to be set, got: %q (err: %v)", activeHead, err)
	}
	_, _, err = st.LoadActiveVersion("lin-concurrent")
	if err != nil {
		t.Errorf("LoadActiveVersion failed: %v", err)
	}
}

func TestComputeLineageKey(t *testing.T) {
	sources := []domain.SourceRepoIdentity{
		{RepoID: "repo-b", URLSHA256: "hash-b"},
		{RepoID: "repo-a", URLSHA256: "hash-a"},
	}

	key1 := store.ComputeLineageKey("sys1", sources)
	// reverse order input should produce identical deterministic lineage key
	key2 := store.ComputeLineageKey("sys1", []domain.SourceRepoIdentity{
		{RepoID: "repo-a", URLSHA256: "hash-a"},
		{RepoID: "repo-b", URLSHA256: "hash-b"},
	})
	if key1 != key2 {
		t.Errorf("expected deterministic lineage key independent of source order: %s != %s", key1, key2)
	}

	key3 := store.ComputeLineageKey("sys2", sources)
	if key1 == key3 {
		t.Errorf("different target system must produce different lineage key")
	}
}
