package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var shaExpr = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// SC3: deterministic manifest with valid SHA-256 digests; check passes after
// install and fails on missing/drifted regions without modifying disk.
func TestSCSkillsManifestAndCheck(t *testing.T) {
	m1, err := BuildManifest("test")
	if err != nil {
		t.Fatal(err)
	}
	m2, err := BuildManifest("test")
	if err != nil {
		t.Fatal(err)
	}
	b1, _ := json.Marshal(m1)
	b2, _ := json.Marshal(m2)
	if string(b1) != string(b2) {
		t.Fatal("manifest is not deterministic")
	}
	if len(m1.Skills) != 6 {
		t.Fatalf("manifest has %d skills, want 6", len(m1.Skills))
	}
	names := make([]string, 0, len(m1.Skills))
	for _, s := range m1.Skills {
		names = append(names, s.Name)
		if !shaExpr.MatchString(s.SHA256) {
			t.Errorf("%s: invalid digest %q", s.Name, s.SHA256)
		}
		if s.Description == "" || len(s.Operations) == 0 || s.OutputContract == "" || s.Budget == nil {
			t.Errorf("%s: manifest entry missing metadata", s.Name)
		}
		raw, err := CanonicalBytes(s.Name)
		if err != nil {
			t.Fatal(err)
		}
		want := "sha256:" + shaHex(raw)
		if s.SHA256 != want {
			t.Errorf("%s: digest %s does not hash canonical SKILL.md bytes", s.Name, s.SHA256)
		}
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("manifest skills not sorted: %v", names)
	}
	if !shaExpr.MatchString(m1.Digest) {
		t.Fatalf("invalid catalog digest %q", m1.Digest)
	}

	for _, host := range Hosts {
		t.Run("check-clean-"+host, func(t *testing.T) {
			root := t.TempDir()
			opts := Options{Root: root}
			if host == "claude" {
				opts.Hook = true
			}
			if _, err := Install(host, opts); err != nil {
				t.Fatal(err)
			}
			rep, err := Check(host, opts)
			if err != nil {
				t.Fatal(err)
			}
			if !rep.InSync || len(rep.Missing) != 0 || len(rep.Drifted) != 0 {
				t.Fatalf("check after install not clean: %+v", rep)
			}
		})
	}

	t.Run("check-missing-readonly", func(t *testing.T) {
		root := t.TempDir()
		before := dirSnapshot(t, root)
		rep, err := Check("claude", Options{Root: root})
		if err != nil {
			t.Fatal(err)
		}
		if rep.InSync || len(rep.Missing) == 0 {
			t.Fatalf("missing files not reported: %+v", rep)
		}
		if !equalSnap(before, dirSnapshot(t, root)) {
			t.Fatal("check modified disk on missing files")
		}
		if len(dirSnapshot(t, root)) != 0 {
			t.Fatal("check created files on disk")
		}
	})

	t.Run("check-drift-readonly", func(t *testing.T) {
		root := t.TempDir()
		if _, err := Install("cursor", Options{Root: root}); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(root, ".cursor", "rules", "gira.mdc")
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(strings.Replace(string(data), "## Skill", "## Tampered", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		tampered, _ := os.ReadFile(p)
		rep, err := Check("cursor", Options{Root: root})
		if err != nil {
			t.Fatal(err)
		}
		if rep.InSync || len(rep.Drifted) == 0 {
			t.Fatalf("drift not reported: %+v", rep)
		}
		after, _ := os.ReadFile(p)
		if string(after) != string(tampered) {
			t.Fatal("check modified drifted file on disk")
		}
	})
}

func dirSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			data, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(root, p)
			out[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	})
	return out
}
