package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
)

// DurabilityStatus indicates whether filesystem metadata synchronization was confirmed (D6).
type DurabilityStatus string

const (
	DurabilityConfirmed          DurabilityStatus = "confirmed"
	DurabilityVisibleUnconfirmed DurabilityStatus = "visible_unconfirmed"
)

// VersionMeta defines the strict publication metadata stored in meta.json.
type VersionMeta struct {
	Format        string    `json:"format"`
	SchemaVersion int       `json:"schema_version,omitempty"`
	VersionID     string    `json:"version_id"`
	SourceKey     string    `json:"source_key"`
	ToolID        string    `json:"tool_id"`
	TargetSystem  string    `json:"target_system"`
	MapSHA256     string    `json:"map_sha256"`
	PublishedAt   time.Time `json:"published_at"`
	LineageID     string    `json:"lineage_id,omitempty"`
}

// Validate verifies the internal consistency and cryptographic invariants of VersionMeta.
func (vm *VersionMeta) Validate() error {
	if vm.Format != "gira.index.version/v1" {
		return fmt.Errorf("invalid metadata format %q: must be gira.index.version/v1", vm.Format)
	}
	if strings.TrimSpace(vm.VersionID) == "" {
		return fmt.Errorf("metadata version_id cannot be empty")
	}
	if strings.TrimSpace(vm.TargetSystem) == "" {
		return fmt.Errorf("metadata target_system cannot be empty")
	}
	if strings.TrimSpace(vm.MapSHA256) == "" {
		return fmt.Errorf("metadata map_sha256 cannot be empty")
	}
	if vm.SourceKey != "" && vm.ToolID != "" {
		expectedID := ComputeVersionID(vm.SourceKey, vm.ToolID)
		if vm.VersionID != expectedID {
			return fmt.Errorf("metadata version_id %q does not match hash of source_key and tool_id (%q)", vm.VersionID, expectedID)
		}
	}
	return nil
}

// PublishResult contains the outcome and durability status of a publication.
type PublishResult struct {
	VersionID        string
	LineageID        string
	Status           DurabilityStatus
	Warning          string
	SymbolLossReport *indexer.SymbolLossReport
}

// PublishVersionOptions configures atomic canonical version publication behavior.
type PublishVersionOptions struct {
	// NoLock disables lock acquisition (for testing or caller-managed locks).
	NoLock bool
	// ManifestPath is the optional path to gira.yaml for M lock acquisition.
	ManifestPath string
	// DeletedPaths lists repository paths known to have been deleted in this update.
	DeletedPaths map[string]bool
	// AllowSymbolLoss explicitly authorizes unexplained removal of real symbols.
	AllowSymbolLoss bool
	// RecoverCorrupt explicitly authorizes recovering from corrupted active head or previous version.
	RecoverCorrupt bool
	// LegacyIndexPath provides an explicit legacy index file path for D9 gate fallback.
	LegacyIndexPath string
	// SourceKey optionally provides the source_key when auto-constructing meta.json.
	SourceKey string
	// ToolID optionally provides the tool_id when auto-constructing meta.json.
	ToolID string
	// LockTimeout overrides the uniform 60s timeout for lock acquisition.
	LockTimeout time.Duration
	// InjectedLockError simulates platform locking failure returning ErrLockUnavailable.
	InjectedLockError error
	// BeforeRenameError is a testing hook executed after staging files are written and synced,
	// immediately before promoting versions/.staging-<id> to versions/<id>.
	BeforeRenameError func() error
	// BeforeHeadRenameError is a testing hook executed after version promotion succeeds,
	// immediately before promoting .head.tmp to .head.
	BeforeHeadRenameError func() error
	// SyncDir is an optional hook to override parent directory fsync.
	SyncDir func(dir string) error
	// HeldLocks provides pre-acquired locks. If non-nil, PublishVersion skips acquiring locks.
	HeldLocks *HeldLocks
}

// PublishSnapshotOptions configures detached snapshot publication.
type PublishSnapshotOptions struct {
	// NoLock disables lock acquisition (for testing or caller-managed locks).
	NoLock bool
	// ManifestPath is the optional path to gira.yaml for M lock acquisition.
	ManifestPath string
	// DeletedPaths lists repository paths known to have been deleted in this update.
	DeletedPaths map[string]bool
	// AllowSymbolLoss explicitly authorizes unexplained removal of real symbols.
	AllowSymbolLoss bool
	// RecoverCorrupt explicitly authorizes overwriting a corrupted existing snapshot.
	RecoverCorrupt bool
	// LockTimeout overrides the uniform 60s timeout for lock acquisition.
	LockTimeout time.Duration
	// InjectedLockError simulates platform locking failure returning ErrLockUnavailable.
	InjectedLockError error
	// BeforeRenameError is a testing hook executed immediately before promoting tmp file.
	BeforeRenameError func() error
	// SyncDir is an optional hook to override parent directory fsync.
	SyncDir func(dir string) error
	// HeldLocks provides pre-acquired locks.
	HeldLocks *HeldLocks
}

// PublishVersion atomically publishes a canonical immutable version into the store under
// <store_root>/<lineage_id>/versions/<version_id>/ and promotes <store_root>/<lineage_id>/.head.
//
// Concurrency & Locks:
//   Acquires locks in strict global order M -> S before evaluating gates or writing to disk.
// Gate:
//   Evaluates the symbol-loss gate against the active head (or legacy index if first publication, D9).
//   If the gate is rejected, publication aborts immediately without modifying the store.
// Idempotency:
//   If versions/<version_id> already exists with identical bytes, it is reused and .head is updated.
//   If versions/<version_id> already exists with differing bytes, it fails closed with
//   ErrNonDeterministicBuild (nondeterministic_build) without overwriting.
// Durability (D6):
//   - On Unix: directory fsync failure returns an error and visible_unconfirmed status.
//   - On Windows: directory fsync is omitted by OS limitation; returns visible_unconfirmed status
//     with exit code 0 and an explicit warning without claiming power-loss durability.
// Retention (D7):
//   Does not perform automatic GC; existing versions remain intact.
func (s *Store) PublishVersion(
	ctx context.Context,
	lineageID string,
	versionID string,
	mm *domain.MacroMap,
	meta *VersionMeta,
	opts PublishVersionOptions,
) (*PublishResult, error) {
	lineageID = strings.TrimSpace(lineageID)
	if lineageID == "" || strings.ContainsAny(lineageID, "/\\") || lineageID == ".." {
		return nil, fmt.Errorf("invalid lineage_id %q: must not be empty or contain path separators", lineageID)
	}
	if mm == nil {
		return nil, fmt.Errorf("macro map cannot be nil")
	}
	if !IsSchema2(mm) {
		return nil, fmt.Errorf("%w: refusing to publish non-schema 2 index into versioned store: %v", ErrUnsupportedSchema, mm.SchemaVersion)
	}

	if mm.SourceIdentity == nil {
		manifestSHA := opts.SourceKey
		if manifestSHA == "" {
			manifestSHA = "0000000000000000000000000000000000000000000000000000000000000000"
		}
		mm.SourceIdentity = &domain.SourceIdentity{
			SchemaVersion:  2,
			ManifestSHA256: manifestSHA,
			Extractors:     []domain.ExtractorIdentity{{ID: "store", Version: "1.0.0"}},
			Sources:        []domain.SourceRepoIdentity{{RepoID: mm.TargetSystem, URLSHA256: "default-sha"}},
		}
	}

	canonicalBytes, err := mm.CanonicalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize macro map: %w", err)
	}

	mapSum := sha256.Sum256(canonicalBytes)
	mapHex := hex.EncodeToString(mapSum[:])

	if meta == nil {
		meta = &VersionMeta{
			Format:        "gira.index.version/v1",
			SchemaVersion: 2,
			VersionID:     versionID,
			SourceKey:     opts.SourceKey,
			ToolID:        opts.ToolID,
			TargetSystem:  mm.TargetSystem,
			MapSHA256:     mapHex,
			PublishedAt:   time.Now().UTC(),
			LineageID:     lineageID,
		}
	} else {
		if meta.Format == "" {
			meta.Format = "gira.index.version/v1"
		}
		if meta.SchemaVersion == 0 {
			meta.SchemaVersion = 2
		}
		if meta.TargetSystem == "" {
			meta.TargetSystem = mm.TargetSystem
		}
		if meta.LineageID == "" {
			meta.LineageID = lineageID
		}
		if meta.PublishedAt.IsZero() {
			meta.PublishedAt = time.Now().UTC()
		}
		if meta.MapSHA256 == "" {
			meta.MapSHA256 = mapHex
		} else if meta.MapSHA256 != mapHex {
			return nil, fmt.Errorf("metadata map_sha256 %q does not match canonical map hash %q", meta.MapSHA256, mapHex)
		}
	}

	if versionID == "" && meta.VersionID != "" {
		versionID = meta.VersionID
	}
	if versionID == "" && opts.SourceKey != "" && opts.ToolID != "" {
		versionID = ComputeVersionID(opts.SourceKey, opts.ToolID)
		meta.VersionID = versionID
	}
	versionID = strings.TrimSpace(versionID)
	if versionID == "" || strings.ContainsAny(versionID, "/\\") || versionID == ".." {
		return nil, fmt.Errorf("invalid version_id %q: must not be empty or contain path separators", versionID)
	}
	meta.VersionID = versionID

	if err := meta.Validate(); err != nil {
		return nil, fmt.Errorf("invalid version metadata: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	lineageDir := filepath.Join(s.BaseDir, lineageID)

	// Step 1: Acquire locks M -> S
	var held *HeldLocks
	if opts.HeldLocks != nil {
		held = opts.HeldLocks
	} else if !opts.NoLock {
		targets := LockTargets{
			ManifestPath: opts.ManifestPath,
			StoreDir:     lineageDir,
		}
		var lockOpts []LockOption
		if opts.LockTimeout > 0 {
			lockOpts = append(lockOpts, WithTimeout(opts.LockTimeout))
		}
		if opts.InjectedLockError != nil {
			lockOpts = append(lockOpts, WithInjectedError(opts.InjectedLockError))
		}
		var err error
		held, err = AcquireLocks(ctx, targets, lockOpts...)
		if err != nil {
			return nil, err
		}
		defer held.Release()
	}

	// Step 2: Evaluate symbol loss gate against active head (or legacy index if initial publication)
	prev, gateBaseMsg, err := s.findBaselineForGate(lineageDir, mm.TargetSystem, opts)
	if err != nil {
		return nil, err
	}
	var report *indexer.SymbolLossReport
	if prev != nil {
		rep, gateErr := indexer.CheckSymbolLossGate(prev, mm, opts.DeletedPaths, opts.AllowSymbolLoss)
		report = rep
		if gateErr != nil {
			return &PublishResult{
				VersionID:        versionID,
				LineageID:        lineageID,
				SymbolLossReport: report,
			}, fmt.Errorf("symbol loss gate rejected publication (%s): %w", gateBaseMsg, gateErr)
		}
	}

	versionsDir := filepath.Join(lineageDir, "versions")
	versionDir := filepath.Join(versionsDir, versionID)
	destMacroMap := filepath.Join(versionDir, "macro-map.json")

	syncFn := syncDirectory
	if opts.SyncDir != nil {
		syncFn = opts.SyncDir
	}

	// Idempotency check: Does version directory already exist?
	if fi, statErr := os.Stat(destMacroMap); statErr == nil && !fi.IsDir() {
		// Existing version directory already has macro-map.json: verify deterministic build
		if err := VerifyDeterministicBuild(destMacroMap, canonicalBytes); err != nil {
			// ErrNonDeterministicBuild: fail closed without touching the existing files
			return nil, err
		}
		// Identical canonical bytes: build is deterministic and idempotent.
		// Update .head to point to versionID if it is not already pointing there.
		currentHead, headErr := s.ReadHead(lineageID)
		if headErr != nil || currentHead != versionID {
			if err := s.atomicUpdateHead(lineageDir, versionID, opts.BeforeHeadRenameError, syncFn); err != nil {
				return nil, err
			}
		}
		return s.resolveDurabilityResult(lineageDir, versionsDir, versionID, lineageID, report, syncFn)
	}

	// Step 3: Create staging version directory and write macro-map.json and meta.json
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating versions dir %s: %w", versionsDir, err)
	}
	stagingDir := filepath.Join(versionsDir, fmt.Sprintf(".staging-%s-%d", versionID, time.Now().UnixNano()))
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating staging dir %s: %w", stagingDir, err)
	}
	cleanupStaging := true
	defer func() {
		if cleanupStaging {
			_ = os.RemoveAll(stagingDir)
		}
	}()

	// Write macro-map.json
	fMap, err := os.OpenFile(filepath.Join(stagingDir, "macro-map.json"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed creating staging macro-map.json: %w", err)
	}
	if _, err := fMap.Write(canonicalBytes); err != nil {
		_ = fMap.Close()
		return nil, fmt.Errorf("failed writing staging macro-map.json: %w", err)
	}
	if err := fMap.Sync(); err != nil {
		_ = fMap.Close()
		return nil, fmt.Errorf("failed syncing staging macro-map.json: %w", err)
	}
	if err := fMap.Close(); err != nil {
		return nil, fmt.Errorf("failed closing staging macro-map.json: %w", err)
	}

	// Write meta.json
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed serializing meta.json: %w", err)
	}
	metaBytes = append(metaBytes, '\n')
	fMeta, err := os.OpenFile(filepath.Join(stagingDir, "meta.json"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed creating staging meta.json: %w", err)
	}
	if _, err := fMeta.Write(metaBytes); err != nil {
		_ = fMeta.Close()
		return nil, fmt.Errorf("failed writing staging meta.json: %w", err)
	}
	if err := fMeta.Sync(); err != nil {
		_ = fMeta.Close()
		return nil, fmt.Errorf("failed syncing staging meta.json: %w", err)
	}
	if err := fMeta.Close(); err != nil {
		return nil, fmt.Errorf("failed closing staging meta.json: %w", err)
	}

	// Sync staging directory
	_ = syncFn(stagingDir)

	// Step 4: BeforeRename test hook & promote staging directory to versions/<version_id>
	if opts.BeforeRenameError != nil {
		if err := opts.BeforeRenameError(); err != nil {
			return nil, fmt.Errorf("aborted before promoting version directory: %w", err)
		}
	}
	if err := os.Rename(stagingDir, versionDir); err != nil {
		return nil, fmt.Errorf("failed promoting version directory to %s: %w", versionDir, err)
	}
	cleanupStaging = false

	// Step 5: Update .head atomically to point to version_id
	if err := s.atomicUpdateHead(lineageDir, versionID, opts.BeforeHeadRenameError, syncFn); err != nil {
		return &PublishResult{
			VersionID:        versionID,
			LineageID:        lineageID,
			Status:           DurabilityVisibleUnconfirmed,
			SymbolLossReport: report,
		}, err
	}

	// Step 6: fsync parent directories & evaluate durability
	return s.resolveDurabilityResult(lineageDir, versionsDir, versionID, lineageID, report, syncFn)
}

// resolveDurabilityResult computes the durability status and performs parent directory syncs.
func (s *Store) resolveDurabilityResult(
	lineageDir, versionsDir, versionID, lineageID string,
	report *indexer.SymbolLossReport,
	syncFn func(dir string) error,
) (*PublishResult, error) {
	result := &PublishResult{
		VersionID:        versionID,
		LineageID:        lineageID,
		SymbolLossReport: report,
	}

	if runtime.GOOS == "windows" {
		// D6 / Gate 5: Windows directory handle fsync is unsupported by OS.
		// Emits explicit visible_unconfirmed warning without claiming power-loss durability.
		result.Status = DurabilityVisibleUnconfirmed
		result.Warning = "visible_unconfirmed: directory fsync is unsupported on Windows; publication durability cannot be guaranteed against sudden power loss"
		return result, nil
	}

	if err := syncFn(lineageDir); err != nil {
		result.Status = DurabilityVisibleUnconfirmed
		return result, fmt.Errorf("visible_unconfirmed: failed syncing lineage directory %s: %w", lineageDir, err)
	}
	if err := syncFn(versionsDir); err != nil {
		result.Status = DurabilityVisibleUnconfirmed
		return result, fmt.Errorf("visible_unconfirmed: failed syncing versions directory %s: %w", versionsDir, err)
	}

	result.Status = DurabilityConfirmed
	return result, nil
}

// atomicUpdateHead atomically writes and promotes .head to point to versionID.
func (s *Store) atomicUpdateHead(
	lineageDir, versionID string,
	beforeHeadRenameError func() error,
	syncFn func(dir string) error,
) error {
	if beforeHeadRenameError != nil {
		if err := beforeHeadRenameError(); err != nil {
			return fmt.Errorf("aborted before updating .head: %w", err)
		}
	}

	headTmp := filepath.Join(lineageDir, fmt.Sprintf(".head.tmp-%d", time.Now().UnixNano()))
	fHead, err := os.OpenFile(headTmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed creating temporary .head at %s: %w", headTmp, err)
	}
	if _, err := fHead.WriteString(versionID + "\n"); err != nil {
		_ = fHead.Close()
		_ = os.Remove(headTmp)
		return fmt.Errorf("failed writing temporary .head: %w", err)
	}
	if err := fHead.Sync(); err != nil {
		_ = fHead.Close()
		_ = os.Remove(headTmp)
		return fmt.Errorf("failed syncing temporary .head: %w", err)
	}
	if err := fHead.Close(); err != nil {
		_ = os.Remove(headTmp)
		return fmt.Errorf("failed closing temporary .head: %w", err)
	}

	headDest := filepath.Join(lineageDir, ".head")
	if err := os.Rename(headTmp, headDest); err != nil {
		_ = os.Remove(headTmp)
		return fmt.Errorf("failed atomically promoting .head to %s: %w", headDest, err)
	}

	_ = syncFn(lineageDir)
	return nil
}

// findBaselineForGate locates the baseline MacroMap for symbol loss evaluation:
// 1. If .head exists, loads the active version.
// 2. If .head does not exist, checks legacy paths (D9).
func (s *Store) findBaselineForGate(
	lineageDir string,
	targetSystem string,
	opts PublishVersionOptions,
) (*domain.MacroMap, string, error) {
	headPath := filepath.Join(lineageDir, ".head")
	if _, err := os.Stat(headPath); err == nil {
		lineageID := filepath.Base(lineageDir)
		activeID, readErr := s.ReadHead(lineageID)
		if readErr != nil {
			if !opts.RecoverCorrupt {
				return nil, "", fmt.Errorf("%w: failed reading .head at %s: %v", ErrCorruptIndex, headPath, readErr)
			}
			return nil, "corrupted head bypassed via --recover-corrupt", nil
		}
		prevMM, _, loadErr := s.LoadVersion(lineageID, activeID)
		if loadErr != nil {
			if !opts.RecoverCorrupt {
				return nil, "", fmt.Errorf("%w: active head version %s at %s cannot be loaded: %v", ErrCorruptIndex, activeID, headPath, loadErr)
			}
			return nil, "corrupted active version bypassed via --recover-corrupt", nil
		}
		return prevMM, fmt.Sprintf("active head %s", activeID), nil
	}

	// First canonical publication in this lineage: check legacy locations (D9)
	var candidates []string
	if opts.LegacyIndexPath != "" {
		candidates = append(candidates, opts.LegacyIndexPath)
	}
	if targetSystem != "" {
		candidates = append(candidates, filepath.Join(s.BaseDir, targetSystem, "macro-map.json"))
		if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
			candidates = append(candidates, filepath.Join(xdg, "gira", "indexes", targetSystem, "macro-map.json"))
		}
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			candidates = append(candidates, filepath.Join(home, ".cache", "gira", "indexes", targetSystem, "macro-map.json"))
		}
	}

	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			prevMM, loadErr := s.LoadFile(cand)
			if loadErr != nil {
				if !opts.RecoverCorrupt {
					return nil, "", fmt.Errorf("existing legacy index at %s cannot be validated and will not be overwritten automatically (pass --recover-corrupt to overwrite): %w", cand, loadErr)
				}
				return nil, fmt.Sprintf("corrupted legacy index at %s bypassed via --recover-corrupt", cand), nil
			}
			return prevMM, fmt.Sprintf("legacy index at %s", cand), nil
		}
	}

	return nil, "no previous index (initial publication)", nil
}

// PublishSnapshot writes a detached snapshot directly to destPath without modifying the store or heads.
//
// Concurrency & Locks:
//   Acquires snapshot lock P on destPath.lock (or M -> P if opts.ManifestPath != "").
// Safety:
//   Fails closed if destPath contains a corrupted index unless opts.RecoverCorrupt is true.
//   Evaluates CheckSymbolLossGate against previous valid snapshot unless opts.AllowSymbolLoss is true.
// Durability (D6):
//   - On Unix: directory fsync failure returns visible_unconfirmed and error.
//   - On Windows: directory fsync is omitted; returns visible_unconfirmed with warning and nil error.
func (s *Store) PublishSnapshot(
	ctx context.Context,
	destPath string,
	mm *domain.MacroMap,
	opts PublishSnapshotOptions,
) (*PublishResult, error) {
	if strings.TrimSpace(destPath) == "" {
		return nil, fmt.Errorf("destination path cannot be empty")
	}
	if mm == nil {
		return nil, fmt.Errorf("macro map cannot be nil")
	}
	if IsSchema2(mm) && mm.SourceIdentity == nil {
		mm.SourceIdentity = &domain.SourceIdentity{
			SchemaVersion:  2,
			ManifestSHA256: "0000000000000000000000000000000000000000000000000000000000000000",
			Extractors:     []domain.ExtractorIdentity{{ID: "snapshot", Version: "1.0.0"}},
			Sources:        []domain.SourceRepoIdentity{{RepoID: mm.TargetSystem, URLSHA256: "snapshot-sha"}},
		}
	}

	canonicalBytes, err := mm.CanonicalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize macro map: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Acquire locks: P (or M -> P)
	var held *HeldLocks
	if opts.HeldLocks != nil {
		held = opts.HeldLocks
	} else if !opts.NoLock {
		targets := LockTargets{
			ManifestPath: opts.ManifestPath,
			SnapshotPath: destPath,
		}
		var lockOpts []LockOption
		if opts.LockTimeout > 0 {
			lockOpts = append(lockOpts, WithTimeout(opts.LockTimeout))
		}
		if opts.InjectedLockError != nil {
			lockOpts = append(lockOpts, WithInjectedError(opts.InjectedLockError))
		}
		var err error
		held, err = AcquireLocks(ctx, targets, lockOpts...)
		if err != nil {
			return nil, err
		}
		defer held.Release()
	}

	var report *indexer.SymbolLossReport
	if fi, statErr := os.Stat(destPath); statErr == nil {
		if fi.IsDir() {
			return nil, fmt.Errorf("destination path %s is a directory", destPath)
		}
		prev, loadErr := s.LoadFile(destPath)
		if loadErr != nil {
			if !opts.RecoverCorrupt {
				return nil, fmt.Errorf("existing snapshot at %s is corrupted and will not be overwritten automatically (pass --recover-corrupt to overwrite): %w", destPath, loadErr)
			}
		} else {
			if IsSchema2(prev) && !IsSchema2(mm) {
				return nil, fmt.Errorf("%w: refusing to downgrade snapshot schema from 2 to %v", ErrSchemaDowngrade, mm.SchemaVersion)
			}
			rep, gateErr := indexer.CheckSymbolLossGate(prev, mm, opts.DeletedPaths, opts.AllowSymbolLoss)
			report = rep
			if gateErr != nil {
				return &PublishResult{SymbolLossReport: report}, fmt.Errorf("refusing to overwrite snapshot %s: %w", destPath, gateErr)
			}
		}
	} else if !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("failed to stat destination %s: %w", destPath, statErr)
	}

	targetDir := filepath.Dir(destPath)
	if targetDir != "" && targetDir != "." {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed to create destination directory: %w", err)
		}
	} else {
		targetDir = "."
	}

	tmpFile, err := os.CreateTemp(targetDir, ".macro-map-*.tmp")
	if err != nil {
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed to create temporary snapshot file in %s: %w", targetDir, err)
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(canonicalBytes); err != nil {
		_ = tmpFile.Close()
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed writing temporary snapshot file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed syncing temporary snapshot file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed closing temporary snapshot file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed setting permissions on temporary snapshot file: %w", err)
	}

	if opts.BeforeRenameError != nil {
		if err := opts.BeforeRenameError(); err != nil {
			return &PublishResult{SymbolLossReport: report}, fmt.Errorf("aborted before promoting snapshot: %w", err)
		}
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return &PublishResult{SymbolLossReport: report}, fmt.Errorf("failed to promote snapshot file to %s: %w", destPath, err)
	}
	cleanupTmp = false

	syncFn := syncDirectory
	if opts.SyncDir != nil {
		syncFn = opts.SyncDir
	}

	res := &PublishResult{
		SymbolLossReport: report,
	}

	if runtime.GOOS == "windows" {
		res.Status = DurabilityVisibleUnconfirmed
		res.Warning = "visible_unconfirmed: directory fsync is unsupported on Windows; publication durability cannot be guaranteed against sudden power loss"
		return res, nil
	}

	if err := syncFn(targetDir); err != nil {
		res.Status = DurabilityVisibleUnconfirmed
		return res, fmt.Errorf("visible_unconfirmed: failed syncing parent directory %s after rename: %w", targetDir, err)
	}

	res.Status = DurabilityConfirmed
	return res, nil
}

// ReadHead reads and returns the active version_id for lineageID from its .head file.
func (s *Store) ReadHead(lineageID string) (string, error) {
	if strings.TrimSpace(lineageID) == "" {
		return "", fmt.Errorf("lineage_id cannot be empty")
	}
	headPath := filepath.Join(s.BaseDir, lineageID, ".head")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	head := strings.TrimSpace(string(data))
	if head == "" {
		return "", fmt.Errorf("%w: .head is empty at %s", ErrCorruptIndex, headPath)
	}
	return head, nil
}

// LoadVersion loads, validates, and strictly unmarshals the versioned index at
// <store_root>/<lineage_id>/versions/<version_id>/.
func (s *Store) LoadVersion(lineageID, versionID string) (*domain.MacroMap, *VersionMeta, error) {
	if strings.TrimSpace(lineageID) == "" {
		return nil, nil, fmt.Errorf("lineage_id cannot be empty")
	}
	if strings.TrimSpace(versionID) == "" {
		return nil, nil, fmt.Errorf("version_id cannot be empty")
	}

	versionDir := filepath.Join(s.BaseDir, lineageID, "versions", versionID)
	metaPath := filepath.Join(versionDir, "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: missing or unreadable meta.json at %s: %v", ErrCorruptIndex, metaPath, err)
	}

	var meta VersionMeta
	dec := json.NewDecoder(bytes.NewReader(metaData))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&meta); err != nil {
		return nil, nil, fmt.Errorf("%w: failed strictly decoding meta.json at %s: %v", ErrCorruptIndex, metaPath, err)
	}
	if err := meta.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: invalid meta.json at %s: %v", ErrCorruptIndex, metaPath, err)
	}
	if meta.VersionID != versionID {
		return nil, nil, fmt.Errorf("%w: meta version_id %q does not match directory version_id %q", ErrCorruptIndex, meta.VersionID, versionID)
	}

	mapPath := filepath.Join(versionDir, "macro-map.json")
	mapData, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: missing or unreadable macro-map.json at %s: %v", ErrCorruptIndex, mapPath, err)
	}

	mapSum := sha256.Sum256(mapData)
	mapHex := hex.EncodeToString(mapSum[:])
	if mapHex != meta.MapSHA256 {
		return nil, nil, fmt.Errorf("%w: macro-map.json sha256 %s does not match meta %s", ErrCorruptIndex, mapHex, meta.MapSHA256)
	}

	mm, err := s.LoadFileStrict(mapPath)
	if err != nil {
		return nil, nil, err
	}
	if meta.TargetSystem != "" && mm.TargetSystem != meta.TargetSystem {
		return nil, nil, fmt.Errorf("%w: macro map target_system %q does not match meta %q", ErrCorruptIndex, mm.TargetSystem, meta.TargetSystem)
	}

	return mm, &meta, nil
}

// LoadActiveVersion reads the .head for lineageID and loads that version.
func (s *Store) LoadActiveVersion(lineageID string) (*domain.MacroMap, *VersionMeta, error) {
	head, err := s.ReadHead(lineageID)
	if err != nil {
		return nil, nil, err
	}
	return s.LoadVersion(lineageID, head)
}

// ListLineages returns all lineage identifiers in the store.
func (s *Store) ListLineages() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list lineages in base dir %s: %w", s.BaseDir, err)
	}

	var lineages []string
	for _, entry := range entries {
		if entry.IsDir() {
			headPath := filepath.Join(s.BaseDir, entry.Name(), ".head")
			versionsDir := filepath.Join(s.BaseDir, entry.Name(), "versions")
			if fi, err := os.Stat(headPath); err == nil && !fi.IsDir() {
				lineages = append(lineages, entry.Name())
			} else if fi, err := os.Stat(versionsDir); err == nil && fi.IsDir() {
				lineages = append(lineages, entry.Name())
			}
		}
	}
	sort.Strings(lineages)
	return lineages, nil
}

// ListVersions returns all version_ids published under lineageID, sorted.
func (s *Store) ListVersions(lineageID string) ([]string, error) {
	if strings.TrimSpace(lineageID) == "" {
		return nil, fmt.Errorf("lineage_id cannot be empty")
	}
	versionsDir := filepath.Join(s.BaseDir, lineageID, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read versions directory %s: %w", versionsDir, err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			versions = append(versions, entry.Name())
		}
	}
	sort.Strings(versions)
	return versions, nil
}
