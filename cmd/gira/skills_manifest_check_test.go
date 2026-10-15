package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SC3 (CLI side): manifest determinism + check pass/fail via `gira skills`.
func TestSCSkillsManifestAndCheck(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := Run([]string{"skills", "manifest", "--format", "json"}, nil, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Version string `json:"version"`
		Digest  string `json:"digest"`
		Skills  []struct {
			Name   string `json:"name"`
			SHA256 string `json:"sha256"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("manifest not JSON: %v\n%s", err, out.String())
	}
	if len(doc.Skills) != 6 {
		t.Fatalf("manifest has %d skills, want 6", len(doc.Skills))
	}
	for _, s := range doc.Skills {
		if len(s.SHA256) != len("sha256:")+64 || !strings.HasPrefix(s.SHA256, "sha256:") {
			t.Errorf("%s: invalid digest %q", s.Name, s.SHA256)
		}
	}
	var out2 bytes.Buffer
	if err := Run([]string{"skills", "manifest", "--format", "json"}, nil, &out2, &errOut); err != nil {
		t.Fatal(err)
	}
	if out.String() != out2.String() {
		t.Fatal("CLI manifest is not deterministic")
	}
	var md bytes.Buffer
	if err := Run([]string{"skills", "manifest", "--format", "md"}, nil, &md, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md.String(), "gira-arch-diff") {
		t.Fatalf("markdown manifest missing skills:\n%s", md.String())
	}

	for _, host := range []string{"claude", "cursor", "agents"} {
		t.Run("cli-check-clean-"+host, func(t *testing.T) {
			dir := t.TempDir()
			args := []string{"skills", "install", "--host", host, "--dir", dir}
			if host == "claude" {
				args = append(args, "--hook")
			}
			if err := Run(args, nil, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			check := []string{"skills", "check", "--host", host, "--dir", dir}
			if host == "claude" {
				check = append(check, "--hook")
			}
			var stdout, stderr bytes.Buffer
			if err := Run(check, nil, &stdout, &stderr); err != nil {
				t.Fatalf("check after install failed: %v\n%s", err, stdout.String())
			}
			var rep struct {
				Host   string `json:"host"`
				InSync bool   `json:"in_sync"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &rep); err != nil {
				t.Fatalf("check output not JSON: %v\n%s", err, stdout.String())
			}
			if !rep.InSync {
				t.Fatalf("check not in sync after install:\n%s", stdout.String())
			}
		})
	}

	t.Run("cli-check-missing", func(t *testing.T) {
		dir := t.TempDir()
		var stdout, stderr bytes.Buffer
		err := Run([]string{"skills", "check", "--host", "cursor", "--dir", dir}, nil, &stdout, &stderr)
		if err == nil {
			t.Fatalf("check on missing files exited 0:\n%s", stdout.String())
		}
		var rep struct {
			InSync  bool     `json:"in_sync"`
			Missing []string `json:"missing"`
		}
		if jerr := json.Unmarshal(stdout.Bytes(), &rep); jerr != nil {
			t.Fatalf("check output not JSON: %v\n%s", jerr, stdout.String())
		}
		if rep.InSync || len(rep.Missing) == 0 {
			t.Fatalf("missing not reported:\n%s", stdout.String())
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Fatal("check created files on disk")
		}
	})

	t.Run("cli-check-drift", func(t *testing.T) {
		dir := t.TempDir()
		if err := Run([]string{"skills", "install", "--host", "cursor", "--dir", dir}, nil, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, ".cursor", "rules", "gira.mdc")
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(strings.Replace(string(data), "## Skill", "## Tampered", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		tampered, _ := os.ReadFile(p)
		var stdout, stderr bytes.Buffer
		err = Run([]string{"skills", "check", "--host", "cursor", "--dir", dir}, nil, &stdout, &stderr)
		if err == nil {
			t.Fatalf("check on drifted file exited 0:\n%s", stdout.String())
		}
		after, _ := os.ReadFile(p)
		if string(after) != string(tampered) {
			t.Fatal("check modified drifted file on disk")
		}
	})
}
