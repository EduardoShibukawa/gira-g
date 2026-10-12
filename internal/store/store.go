package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"labs/gira-g/internal/domain"
)

// Store handles persistence and retrieval of canonical MacroMap indexes.
type Store struct {
	BaseDir string
}

func NewStore(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

// Save serializes and writes a MacroMap to <BaseDir>/<targetSystem>/macro-map.json.
func (s *Store) Save(targetSystem string, mm *domain.MacroMap) error {
	if strings.TrimSpace(targetSystem) == "" {
		return fmt.Errorf("target system name cannot be empty")
	}
	if mm == nil {
		return fmt.Errorf("macro map cannot be nil")
	}

	data, err := mm.CanonicalJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize macro map: %w", err)
	}

	targetDir := filepath.Join(s.BaseDir, targetSystem)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target store directory: %w", err)
	}

	dest := filepath.Join(targetDir, "macro-map.json")
	return os.WriteFile(dest, data, 0644)
}

// Load reads and deserializes <BaseDir>/<targetSystem>/macro-map.json.
func (s *Store) Load(targetSystem string) (*domain.MacroMap, error) {
	if strings.TrimSpace(targetSystem) == "" {
		return nil, fmt.Errorf("target system name cannot be empty")
	}
	dest := filepath.Join(s.BaseDir, targetSystem, "macro-map.json")
	return s.LoadFile(dest)
}

// LoadFile reads and deserializes a MacroMap from an arbitrary JSON file path.
func (s *Store) LoadFile(filePath string) (*domain.MacroMap, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read macro map file %s: %w", filePath, err)
	}

	var mm domain.MacroMap
	if err := json.Unmarshal(data, &mm); err != nil {
		return nil, fmt.Errorf("failed to decode macro map JSON from %s: %w", filePath, err)
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
