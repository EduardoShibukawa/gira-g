package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	// BeforeRenameError is an optional hook (used by tests) invoked after the temporary
	// file is synced and closed, immediately before os.Rename promotes it.
	BeforeRenameError func() error
}

// Store handles persistence and retrieval of canonical MacroMap indexes.
type Store struct {
	BaseDir string
}

func NewStore(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

// Save serializes and writes a MacroMap atomically to <BaseDir>/<targetSystem>/macro-map.json.
// If an existing index at the destination is corrupted, Save refuses to overwrite it.
func (s *Store) Save(targetSystem string, mm *domain.MacroMap) error {
	if strings.TrimSpace(targetSystem) == "" {
		return fmt.Errorf("target system name cannot be empty")
	}
	dest := filepath.Join(s.BaseDir, targetSystem, "macro-map.json")
	_, err := s.PublishFile(dest, mm, PublishOptions{AllowSymbolLoss: true})
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

// SaveFile atomically writes a MacroMap to dest, refusing to overwrite a corrupted existing file.
func (s *Store) SaveFile(dest string, mm *domain.MacroMap) error {
	_, err := s.PublishFile(dest, mm, PublishOptions{AllowSymbolLoss: true})
	return err
}

// PublishFile validates mm, inspects any existing file at dest (failing closed if it is
// corrupted unless opts.RecoverCorrupt is true, and running CheckSymbolLossGate unless
// opts.AllowSymbolLoss is true), and atomically writes the new index via a temporary file
// and os.Rename so an interrupted write never corrupts or truncates the previous index.
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

	return report, nil
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

// List returns all target system names under BaseDir containing a macro-map.json file.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list systems in base dir %s: %w", s.BaseDir, err)
	}

	var systems []string
	for _, entry := range entries {
		if entry.IsDir() {
			candidate := filepath.Join(s.BaseDir, entry.Name(), "macro-map.json")
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				systems = append(systems, entry.Name())
			}
		}
	}

	sort.Strings(systems)
	return systems, nil
}

