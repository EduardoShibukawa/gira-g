package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"labs/gira-g/internal/manifest"
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
