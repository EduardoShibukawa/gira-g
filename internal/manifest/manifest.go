package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	sha40Regex = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
)

// Manifest represents the top-level declaration in gira.yaml.
type Manifest struct {
	Version      int                     `yaml:"version"`
	TargetSystem TargetSystemMeta        `yaml:"target_system"`
	Repositories []RepositoryDeclaration `yaml:"repositories"`
}

// TargetSystemMeta describes the target system identity and context.
type TargetSystemMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// RepositoryDeclaration defines a Git repository belonging to the target system.
type RepositoryDeclaration struct {
	ID        string `yaml:"id"`
	URL       string `yaml:"url"`
	Ref       string `yaml:"ref,omitempty"`
	SHA       string `yaml:"sha"`
	PathScope string `yaml:"path_scope,omitempty"`
}

// Parse unmarshals and strictly validates manifest YAML data.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	return &m, nil
}

// ParseFile reads a file from disk and parses it as a Manifest.
func ParseFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	return Parse(data)
}

// Validate performs strict validation on the manifest structure and values.
func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.TargetSystem.Name) == "" {
		return fmt.Errorf("target_system.name cannot be empty")
	}

	if len(m.Repositories) == 0 {
		return fmt.Errorf("at least one repository must be declared")
	}

	seenIDs := make(map[string]bool, len(m.Repositories))

	for i := range m.Repositories {
		repo := &m.Repositories[i]

		repo.ID = strings.TrimSpace(repo.ID)
		if repo.ID == "" {
			return fmt.Errorf("repository at index %d has empty id", i)
		}

		if seenIDs[repo.ID] {
			return fmt.Errorf("duplicate repository id %q", repo.ID)
		}
		seenIDs[repo.ID] = true

		repo.URL = strings.TrimSpace(repo.URL)
		if repo.URL == "" {
			return fmt.Errorf("repository %q has empty url", repo.ID)
		}
		if !isValidRepoURL(repo.URL) {
			return fmt.Errorf("repository %q has invalid url %q", repo.ID, repo.URL)
		}

		repo.SHA = strings.TrimSpace(repo.SHA)
		if !sha40Regex.MatchString(repo.SHA) {
			return fmt.Errorf("repository %q has invalid commit sha %q (must be exactly 40 hex characters)", repo.ID, repo.SHA)
		}

		scope := strings.TrimSpace(repo.PathScope)
		if scope == "" || scope == "." {
			repo.PathScope = "."
		} else {
			cleanScope := filepath.Clean(scope)
			if cleanScope == ".." || strings.HasPrefix(cleanScope, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanScope, "/") || filepath.IsAbs(cleanScope) {
				return fmt.Errorf("repository %q has unsafe path_scope %q", repo.ID, repo.PathScope)
			}
			repo.PathScope = filepath.ToSlash(cleanScope)
		}
	}

	return nil
}

// isValidRepoURL checks if the URL matches acceptable git repository URL schemes or relative/local paths.
func isValidRepoURL(u string) bool {
	if strings.HasPrefix(u, "https://") ||
		strings.HasPrefix(u, "http://") ||
		strings.HasPrefix(u, "ssh://") ||
		strings.HasPrefix(u, "git@") ||
		strings.HasPrefix(u, "file://") {
		return true
	}

	// Relative or absolute filesystem path
	if strings.HasPrefix(u, "/") || strings.HasPrefix(u, "./") || strings.HasPrefix(u, "../") || filepath.IsAbs(u) {
		return true
	}

	// Host-relative SSH formats like git@github.com:org/repo.git or user@host:path
	if strings.Contains(u, "@") && strings.Contains(u, ":") {
		return true
	}

	return false
}
