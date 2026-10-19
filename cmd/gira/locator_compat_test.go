package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func init() {
	// Compatibility hook for legacy test fixtures (such as TestCLI_Index in main_test.go)
	// where the test fixture creates a manifest with declarative overrides inside an ephemeral
	// t.TempDir() without initializing a git repository.
	//
	// When running locator tests (in locator_test.go), tests explicitly set
	// manifest.GitRootDiscoveryHook = nil to exercise the real production Git discovery logic.
	manifest.GitRootDiscoveryHook = func(manifestPath string) (string, error) {
		root, err := manifest.DefaultGitRootDiscoverer(manifestPath)
		if err == nil {
			return root, nil
		}

		abs, absErr := filepath.Abs(manifestPath)
		if absErr == nil {
			evalPath, evalErr := filepath.EvalSymlinks(abs)
			if evalErr == nil {
				abs = evalPath
			}
			tempDir, tempErr := filepath.EvalSymlinks(os.TempDir())
			if tempErr == nil && strings.HasPrefix(abs, tempDir) {
				return filepath.Dir(abs), nil
			}
		}
		return root, err
	}
}
