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
	Overrides    *OverridesDeclaration   `yaml:"overrides,omitempty"`
}

// TargetSystemMeta describes the target system identity and context.
type TargetSystemMeta struct {
	Name        string                `yaml:"name"`
	Description string                `yaml:"description,omitempty"`
	Overrides   *OverridesDeclaration `yaml:"overrides,omitempty"`
}

// RelationshipOverride declares an explicit relationship correction or addition.
type RelationshipOverride struct {
	SourceID  string `yaml:"source_id"`
	TargetID  string `yaml:"target_id"`
	Verb      string `yaml:"verb"`
	StartLine int    `yaml:"-"`
	EndLine   int    `yaml:"-"`
}

// OverridesDeclaration defines declarative human corrections for relationships and metadata.
type OverridesDeclaration struct {
	Relationships []RelationshipOverride    `yaml:"relationships,omitempty"`
	Metadata      map[string]map[string]any `yaml:"metadata,omitempty"`
	MetadataLines map[string]int            `yaml:"-"`
}

// UnmarshalYAML decodes OverridesDeclaration while recording the exact YAML source lines
// for each relationship item and metadata entry.
func (o *OverridesDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type rawOverrides struct {
		Relationships []RelationshipOverride    `yaml:"relationships,omitempty"`
		Metadata      map[string]map[string]any `yaml:"metadata,omitempty"`
	}
	var raw rawOverrides
	if err := value.Decode(&raw); err != nil {
		return err
	}
	o.Relationships = raw.Relationships
	o.Metadata = raw.Metadata

	if value.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]
		switch keyNode.Value {
		case "relationships":
			if valNode.Kind == yaml.SequenceNode {
				for idx, item := range valNode.Content {
					if idx < len(o.Relationships) {
						start, end := yamlNodeLineSpan(item)
						o.Relationships[idx].StartLine = start
						o.Relationships[idx].EndLine = end
					}
				}
			}
		case "metadata":
			if valNode.Kind == yaml.MappingNode {
				if o.MetadataLines == nil {
					o.MetadataLines = make(map[string]int)
				}
				for j := 0; j+1 < len(valNode.Content); j += 2 {
					nodeKey := strings.TrimSpace(valNode.Content[j].Value)
					if nodeKey != "" && valNode.Content[j].Line > 0 {
						o.MetadataLines[nodeKey] = valNode.Content[j].Line
					}
				}
			}
		}
	}
	return nil
}

func yamlNodeLineSpan(n *yaml.Node) (int, int) {
	if n == nil {
		return 0, 0
	}
	start := n.Line
	end := n.Line
	var walk func(cur *yaml.Node)
	walk = func(cur *yaml.Node) {
		if cur == nil {
			return
		}
		if cur.Line > 0 {
			if start == 0 || cur.Line < start {
				start = cur.Line
			}
			if cur.Line > end {
				end = cur.Line
			}
		}
		for _, child := range cur.Content {
			walk(child)
		}
	}
	walk(n)
	return start, end
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
			// The separator-prefixed check is platform-specific by construction:
			// on Linux it duplicates the "/" check below, on Windows it catches
			// rooted scopes such as "\etc\passwd", which filepath.Clean produces
			// from "/etc/passwd" and filepath.IsAbs does not report as absolute.
			if cleanScope == ".." || strings.HasPrefix(cleanScope, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanScope, "/") || strings.HasPrefix(cleanScope, string(filepath.Separator)) || filepath.IsAbs(cleanScope) {
				return fmt.Errorf("repository %q has unsafe path_scope %q", repo.ID, repo.PathScope)
			}
			repo.PathScope = filepath.ToSlash(cleanScope)
		}
	}

	// Normalize overrides if declared under target_system
	if m.Overrides == nil && m.TargetSystem.Overrides != nil {
		m.Overrides = m.TargetSystem.Overrides
	} else if m.Overrides != nil && m.TargetSystem.Overrides == nil {
		m.TargetSystem.Overrides = m.Overrides
	}

	if m.Overrides != nil {
		if err := m.Overrides.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate performs validation on declarative overrides.
func (o *OverridesDeclaration) Validate() error {
	for i := range o.Relationships {
		rel := &o.Relationships[i]
		rel.SourceID = strings.TrimSpace(rel.SourceID)
		if rel.SourceID == "" {
			return fmt.Errorf("override relationship at index %d has empty source_id", i)
		}
		rel.TargetID = strings.TrimSpace(rel.TargetID)
		if rel.TargetID == "" {
			return fmt.Errorf("override relationship at index %d has empty target_id", i)
		}
		rel.Verb = strings.TrimSpace(rel.Verb)
		if rel.Verb == "" {
			return fmt.Errorf("override relationship at index %d has empty verb", i)
		}
		switch rel.Verb {
		case "provides", "consumes", "imports", "accesses":
		default:
			return fmt.Errorf("override relationship at index %d has invalid verb %q", i, rel.Verb)
		}
	}

	for nodeID, meta := range o.Metadata {
		if strings.TrimSpace(nodeID) == "" {
			return fmt.Errorf("override metadata has empty node id")
		}
		if meta == nil {
			return fmt.Errorf("override metadata for node %q cannot be nil", nodeID)
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
	if u == "." || strings.HasPrefix(u, "/") || strings.HasPrefix(u, "./") || strings.HasPrefix(u, "../") || filepath.IsAbs(u) {
		return true
	}

	// Host-relative SSH formats like git@github.com:org/repo.git or user@host:path
	if strings.Contains(u, "@") && strings.Contains(u, ":") {
		return true
	}

	return false
}
