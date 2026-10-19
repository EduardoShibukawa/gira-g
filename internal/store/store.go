package store

import (
	"bytes"
	"encoding/json"
	"errors"
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

// ErrCorruptIndex indicates that an existing index file could not be decoded or failed structural validation.
var ErrCorruptIndex = errors.New("corrupt or invalid macro-map index")

// PublishOptions controls pre-publication validation and recovery behavior.
type PublishOptions struct {
	// DeletedPaths lists repository paths known to have been deleted in this update.
	DeletedPaths map[string]bool
	// AllowSymbolLoss explicitly authorizes unexplained removal of real symbols.
	AllowSymbolLoss bool
	// RecoverCorrupt explicitly authorizes overwriting a corrupted or invalid existing index file.
	RecoverCorrupt bool
	// ManifestPath is an optional path to gira.yaml for M lock acquisition.
	ManifestPath string
	// LockTimeout overrides default lock acquisition timeout.
	LockTimeout time.Duration
	// InjectedLockError simulates platform locking failure returning ErrLockUnavailable.
	InjectedLockError error
	// BeforeRenameError is an optional hook (used by tests) invoked after the temporary
	// file is synced and closed, immediately before os.Rename promotes it.
	BeforeRenameError func() error
	// SyncDir is an optional hook (used by tests) to override parent directory fsync
	// after os.Rename completes. When nil, syncDirectory is used.
	SyncDir func(dir string) error
}

// Store handles persistence and retrieval of canonical MacroMap indexes.
type Store struct {
	BaseDir string
}

func NewStore(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

// Save validates and atomically writes a MacroMap to <BaseDir>/<targetSystem>/macro-map.json,
// enforcing both corrupt-index protection and the symbol-loss gate by default.
// Callers that need to authorize symbol loss or corrupt-index recovery must use Publish with explicit PublishOptions.
func (s *Store) Save(targetSystem string, mm *domain.MacroMap) error {
	_, err := s.Publish(targetSystem, mm, PublishOptions{})
	return err
}

// Publish validates and atomically publishes a MacroMap to <BaseDir>/<targetSystem>/macro-map.json,
// enforcing corrupt-index protection and the symbol-loss gate against any previous index.
func (s *Store) Publish(targetSystem string, mm *domain.MacroMap, opts PublishOptions) (*indexer.SymbolLossReport, error) {
	if strings.TrimSpace(targetSystem) == "" {
		return nil, fmt.Errorf("target system name cannot be empty")
	}
	dest := filepath.Join(s.BaseDir, targetSystem, "macro-map.json")
	return s.PublishFile(dest, mm, opts)
}

// SaveFile validates and atomically writes a MacroMap to dest, enforcing both
// corrupt-index protection and the symbol-loss gate by default.
// Callers that need to authorize symbol loss or corrupt-index recovery must use PublishFile with explicit PublishOptions.
func (s *Store) SaveFile(dest string, mm *domain.MacroMap) error {
	_, err := s.PublishFile(dest, mm, PublishOptions{})
	return err
}

// PublishFile validates mm, inspects any existing file at dest (failing closed if it is
// corrupted unless opts.RecoverCorrupt is true, and running CheckSymbolLossGate unless
// opts.AllowSymbolLoss is true), and atomically writes the new index via a temporary file
// in the same directory, tmpFile.Sync(), os.Rename, and parent directory Sync().
//
// Durability boundaries:
//   - Any failure before or during os.Rename leaves the previous file at dest byte-for-byte
//     intact and removes the temporary file.
//   - Once os.Rename succeeds, the new index is live in the filesystem namespace; if the
//     subsequent parent directory Sync() fails, PublishFile returns an error indicating that
//     post-rename directory metadata durability could not be guaranteed.
func (s *Store) PublishFile(dest string, mm *domain.MacroMap, opts PublishOptions) (*indexer.SymbolLossReport, error) {
	if strings.TrimSpace(dest) == "" {
		return nil, fmt.Errorf("destination path cannot be empty")
	}
	if mm == nil {
		return nil, fmt.Errorf("macro map cannot be nil")
	}
	if err := mm.Validate(); err != nil {
		return nil, fmt.Errorf("refusing to publish invalid macro map: %w", err)
	}

	data, err := mm.CanonicalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize macro map: %w", err)
	}

	var report *indexer.SymbolLossReport
	if fi, statErr := os.Stat(dest); statErr == nil {
		if fi.IsDir() {
			return nil, fmt.Errorf("destination path %s is a directory", dest)
		}
		prev, loadErr := s.LoadFile(dest)
		if loadErr != nil {
			if !opts.RecoverCorrupt {
				return nil, fmt.Errorf("existing index at %s cannot be validated and will not be overwritten automatically (pass --recover-corrupt to overwrite): %w", dest, loadErr)
			}
		} else {
			rep, gateErr := indexer.CheckSymbolLossGate(prev, mm, opts.DeletedPaths, opts.AllowSymbolLoss)
			report = rep
			if gateErr != nil {
				return report, fmt.Errorf("refusing to overwrite %s: %w", dest, gateErr)
			}
		}
	} else if !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("failed to stat destination %s: %w", dest, statErr)
	}

	targetDir := filepath.Dir(dest)
	if targetDir != "" && targetDir != "." {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return report, fmt.Errorf("failed to create target store directory: %w", err)
		}
	} else {
		targetDir = "."
	}

	tmpFile, err := os.CreateTemp(targetDir, ".macro-map-*.tmp")
	if err != nil {
		return report, fmt.Errorf("failed to create temporary index file in %s: %w", targetDir, err)
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return report, fmt.Errorf("failed writing temporary index file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return report, fmt.Errorf("failed syncing temporary index file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return report, fmt.Errorf("failed closing temporary index file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return report, fmt.Errorf("failed setting permissions on temporary index file: %w", err)
	}

	if opts.BeforeRenameError != nil {
		if err := opts.BeforeRenameError(); err != nil {
			return report, fmt.Errorf("aborted before atomic rename: %w", err)
		}
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		return report, fmt.Errorf("failed to atomically promote index file to %s: %w", dest, err)
	}
	cleanupTmp = false

	syncFn := syncDirectory
	if opts.SyncDir != nil {
		syncFn = opts.SyncDir
	}
	if err := syncFn(targetDir); err != nil {
		return report, fmt.Errorf("failed syncing parent directory %s after rename: %w", targetDir, err)
	}

	return report, nil
}

// syncDirectory opens dir and invokes Sync() to flush the directory entry update
// after os.Rename. On Windows, directory handle fsync is unsupported by the OS
// and is skipped.
func syncDirectory(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	syncErr := d.Sync()
	closeErr := d.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

// Load reads and deserializes <BaseDir>/<targetSystem>/macro-map.json.
func (s *Store) Load(targetSystem string) (*domain.MacroMap, error) {
	if strings.TrimSpace(targetSystem) == "" {
		return nil, fmt.Errorf("target system name cannot be empty")
	}
	dest := filepath.Join(s.BaseDir, targetSystem, "macro-map.json")
	return s.LoadFile(dest)
}

// LoadFile reads, decodes, and validates a MacroMap from an arbitrary JSON file path.
func (s *Store) LoadFile(filePath string) (*domain.MacroMap, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read macro map file %s: %w", filePath, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("%w at %s: file is empty", ErrCorruptIndex, filePath)
	}

	var mm domain.MacroMap
	if err := json.Unmarshal(data, &mm); err != nil {
		return nil, fmt.Errorf("%w at %s: failed to decode JSON: %w", ErrCorruptIndex, filePath, err)
	}
	if err := mm.Validate(); err != nil {
		return nil, fmt.Errorf("%w at %s: %w", ErrCorruptIndex, filePath, err)
	}

	return &mm, nil
}

// List returns all target system / lineage names under BaseDir containing either
// a legacy macro-map.json file, a .head pointer, or a versions/ directory.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list systems in base dir %s: %w", s.BaseDir, err)
	}

	seen := make(map[string]bool)
	var systems []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			legacyCandidate := filepath.Join(s.BaseDir, name, "macro-map.json")
			headCandidate := filepath.Join(s.BaseDir, name, ".head")
			versionsCandidate := filepath.Join(s.BaseDir, name, "versions")
			if fi, err := os.Stat(legacyCandidate); err == nil && !fi.IsDir() {
				if !seen[name] {
					seen[name] = true
					systems = append(systems, name)
				}
			} else if fi, err := os.Stat(headCandidate); err == nil && !fi.IsDir() {
				if !seen[name] {
					seen[name] = true
					systems = append(systems, name)
				}
			} else if fi, err := os.Stat(versionsCandidate); err == nil && fi.IsDir() {
				if !seen[name] {
					seen[name] = true
					systems = append(systems, name)
				}
			}
		}
	}

	sort.Strings(systems)
	return systems, nil
}

