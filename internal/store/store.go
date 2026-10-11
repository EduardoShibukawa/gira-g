package store

import (
	"fmt"
	"os"
	"path/filepath"

	"labs/gira-g/internal/domain"
)

// Store handles persistence and retrieval of canonical MacroMap indexes.
type Store struct {
	BaseDir string
}

func NewStore(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

func (s *Store) Save(targetSystem string, mm *domain.MacroMap) error {
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
