package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func TestParse_ValidManifest(t *testing.T) {
	yamlData := `
version: 1
target_system:
  name: opentelemetry-demo
  description: "OpenTelemetry Demo Reference Corpus (v3.1.0)"
repositories:
  - id: demo
    url: "https://github.com/open-telemetry/opentelemetry-demo.git"
    ref: "3.1.0"
    sha: "dedc0178918e260823323b8d95005a8cb924b007"
    path_scope: "."
  - id: helm-charts
    url: "git@github.com:open-telemetry/opentelemetry-helm-charts.git"
    ref: "opentelemetry-demo-0.42.3"
    sha: "a9f2dd0131b933f828049124b113957882b05518"
    path_scope: "charts/opentelemetry-demo"
`

	m, err := manifest.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error parsing valid manifest: %v", err)
	}

	if m.TargetSystem.Name != "opentelemetry-demo" {
		t.Errorf("expected target_system.name to be opentelemetry-demo, got %s", m.TargetSystem.Name)
	}

	if len(m.Repositories) != 2 {
		t.Fatalf("expected 2 repositories, got %d", len(m.Repositories))
	}

	if m.Repositories[0].PathScope != "." {
		t.Errorf("expected path_scope '.' got %q", m.Repositories[0].PathScope)
	}

	if m.Repositories[1].PathScope != "charts/opentelemetry-demo" {
		t.Errorf("expected path_scope 'charts/opentelemetry-demo' got %q", m.Repositories[1].PathScope)
	}
}

func TestParse_DefaultPathScope(t *testing.T) {
	yamlData := `
version: 1
target_system:
  name: test-system
repositories:
  - id: repo1
    url: "file:///tmp/repo1"
    sha: "1234567890abcdef1234567890abcdef12345678"
`

	m, err := manifest.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Repositories[0].PathScope != "." {
		t.Errorf("expected default path_scope '.', got %q", m.Repositories[0].PathScope)
	}
}

func TestParse_InvalidCases(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "empty target_system name",
			yaml: `
version: 1
target_system:
  name: ""
repositories:
  - id: repo1
    url: "https://example.com/repo.git"
    sha: "1234567890abcdef1234567890abcdef12345678"
`,
			wantErr: "target_system.name cannot be empty",
		},
		{
			name: "no repositories",
			yaml: `
version: 1
target_system:
  name: test
repositories: []
`,
			wantErr: "at least one repository must be declared",
		},
		{
			name: "empty repo id",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: "   "
    url: "https://example.com/repo.git"
    sha: "1234567890abcdef1234567890abcdef12345678"
`,
			wantErr: "has empty id",
		},
		{
			name: "duplicate repo id",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "https://example.com/repo1.git"
    sha: "1234567890abcdef1234567890abcdef12345678"
  - id: repo1
    url: "https://example.com/repo2.git"
    sha: "abcdef1234567890abcdef1234567890abcdef12"
`,
			wantErr: "duplicate repository id",
		},
		{
			name: "invalid sha length",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "https://example.com/repo.git"
    sha: "12345678"
`,
			wantErr: "must be exactly 40 hex characters",
		},
		{
			name: "invalid sha non-hex",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "https://example.com/repo.git"
    sha: "1234567890abcdef1234567890abcdef1234567z"
`,
			wantErr: "must be exactly 40 hex characters",
		},
		{
			name: "unsafe path scope traversal",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "https://example.com/repo.git"
    sha: "1234567890abcdef1234567890abcdef12345678"
    path_scope: "../outside"
`,
			wantErr: "unsafe path_scope",
		},
		{
			name: "unsafe path scope absolute",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "https://example.com/repo.git"
    sha: "1234567890abcdef1234567890abcdef12345678"
    path_scope: "/etc/passwd"
`,
			wantErr: "unsafe path_scope",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := manifest.Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestParse_WithOverrides(t *testing.T) {
	yamlData := `
version: 1
target_system:
  name: opentelemetry-demo
  description: "OpenTelemetry Demo with overrides"
repositories:
  - id: demo
    url: "https://github.com/open-telemetry/opentelemetry-demo.git"
    sha: "dedc0178918e260823323b8d95005a8cb924b007"
overrides:
  relationships:
    - source_id: "service:frontend"
      target_id: "contract:custom"
      verb: "consumes"
  metadata:
    "service:frontend":
      team: "web-core"
`

	m, err := manifest.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error parsing manifest with overrides: %v", err)
	}

	if m.Overrides == nil {
		t.Fatalf("expected Overrides to be non-nil")
	}

	if len(m.Overrides.Relationships) != 1 {
		t.Fatalf("expected 1 relationship override, got %d", len(m.Overrides.Relationships))
	}

	rel := m.Overrides.Relationships[0]
	if rel.SourceID != "service:frontend" || rel.TargetID != "contract:custom" || rel.Verb != "consumes" {
		t.Errorf("unexpected relationship override values: %+v", rel)
	}

	if m.Overrides.Metadata == nil || m.Overrides.Metadata["service:frontend"]["team"] != "web-core" {
		t.Errorf("unexpected metadata overrides: %+v", m.Overrides.Metadata)
	}

	// Also test overrides inside target_system
	nestedYaml := `
version: 1
target_system:
  name: opentelemetry-demo
  overrides:
    relationships:
      - source_id: "service:a"
        target_id: "service:b"
        verb: "accesses"
repositories:
  - id: demo
    url: "file:///tmp/repo"
    sha: "1111222233334444555566667777888899990000"
`
	m2, err := manifest.Parse([]byte(nestedYaml))
	if err != nil {
		t.Fatalf("unexpected error with nested overrides: %v", err)
	}
	if m2.Overrides == nil || len(m2.Overrides.Relationships) != 1 {
		t.Fatalf("expected normalized overrides from target_system")
	}
	if m2.Overrides.Relationships[0].Verb != "accesses" {
		t.Errorf("expected verb accesses, got %s", m2.Overrides.Relationships[0].Verb)
	}
}

func TestParse_InvalidOverrides(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "empty source_id",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "file:///tmp/r"
    sha: "1111222233334444555566667777888899990000"
overrides:
  relationships:
    - source_id: ""
      target_id: "contract:c"
      verb: "consumes"
`,
			wantErr: "has empty source_id",
		},
		{
			name: "empty target_id",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "file:///tmp/r"
    sha: "1111222233334444555566667777888899990000"
overrides:
  relationships:
    - source_id: "service:a"
      target_id: ""
      verb: "consumes"
`,
			wantErr: "has empty target_id",
		},
		{
			name: "empty verb",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "file:///tmp/r"
    sha: "1111222233334444555566667777888899990000"
overrides:
  relationships:
    - source_id: "service:a"
      target_id: "service:b"
      verb: ""
`,
			wantErr: "has empty verb",
		},
		{
			name: "invalid verb",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "file:///tmp/r"
    sha: "1111222233334444555566667777888899990000"
overrides:
  relationships:
    - source_id: "service:a"
      target_id: "service:b"
      verb: "kills"
`,
			wantErr: "has invalid verb \"kills\"",
		},
		{
			name: "empty metadata node id",
			yaml: `
version: 1
target_system:
  name: test
repositories:
  - id: repo1
    url: "file:///tmp/r"
    sha: "1111222233334444555566667777888899990000"
overrides:
  metadata:
    "":
      team: "web"
`,
			wantErr: "empty node id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := manifest.Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "gira.yaml")

	content := `
version: 1
target_system:
  name: file-test
repositories:
  - id: r1
    url: "file:///tmp/repo"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	m, err := manifest.ParseFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.TargetSystem.Name != "file-test" {
		t.Errorf("got %s, want file-test", m.TargetSystem.Name)
	}

	_, err = manifest.ParseFile(filepath.Join(tmpDir, "non-existent.yaml"))
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}
}

func contains(s, substr string) bool {
	return filepath.Clean(s) != "" && (len(s) >= len(substr)) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && (s != "" && (os.PathSeparator != 0) && (len(s) >= len(substr))) && (search(s, substr))))
}

func search(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
