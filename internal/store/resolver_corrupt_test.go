package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// writeCorruptStore lays out <root>/<lineage>/{.head,versions/<v>/...} with the
// given files and returns the store root.
func writeCorruptStore(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func manifestFile(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "gira.yaml")
	body := "version: 1\ntarget_system:\n  name: " + name + "\n  description: corrupt test\nrepositories:\n  - id: repo-core\n    url: \"https://example.invalid/repo-core.git\"\n    ref: v1.0.0\n    sha: \"1111111111111111111111111111111111111111\"\n    path_scope: \".\"\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// SC5: every index_corrupt raised during discovery with a manifest names the
// exact broken cache file and suggests --index, while still wrapping
// ErrCorruptIndex so structured codes keep working.
func TestSCIndexCorruptErrorDiscoveryNamesFile(t *testing.T) {
	const ver = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	cases := []struct {
		name      string
		files     map[string]string
		wantInErr []string
	}{
		{
			name:      "empty .head",
			files:     map[string]string{"systest/.head": ""},
			wantInErr: []string{".head is empty", filepath.Join("systest", ".head"), "--index <path>"},
		},
		{
			name: "unreadable meta.json",
			files: map[string]string{
				"systest/.head":                               ver,
				"systest/versions/" + ver + "/meta.json":      "{not json",
				"systest/versions/" + ver + "/macro-map.json": "{}",
			},
			wantInErr: []string{"corrupted", "meta.json", "--index <path>"},
		},
		{
			name: "missing meta.json names the path",
			files: map[string]string{
				"systest/.head": ver,
			},
			wantInErr: []string{"meta.json", "--index <path>"},
		},
		{
			name: "corrupt macro-map under the active head names the path",
			files: map[string]string{
				"systest/.head":                               ver,
				"systest/versions/" + ver + "/meta.json":      `{"format":"gira.index.version/v1","version_id":"` + ver + `","source_key":"k","tool_id":"","target_system":"systest","map_sha256":"x","published_at":"2026-10-19T00:00:00Z"}`,
				"systest/versions/" + ver + "/macro-map.json": "###",
			},
			wantInErr: []string{"index_corrupt", "macro-map.json", "--index <path>"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeCorruptStore(t, tc.files)
			mf := manifestFile(t, "systest")
			_, err := store.Resolve(t.Context(), store.ResolveOptions{
				ManifestPath: mf,
				StoreRoots:   []string{root},
			})
			if err == nil {
				t.Fatal("expected discovery to fail on corrupt cache")
			}
			if !strings.Contains(err.Error(), "index_corrupt") {
				t.Fatalf("error lost the structured code: %v", err)
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error must contain %q, got: %v", want, err)
				}
			}
			if !strings.Contains(err.Error(), root) {
				t.Fatalf("error must name the absolute cache path %q, got: %v", root, err)
			}
		})
	}
}
