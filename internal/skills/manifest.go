package skills

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// ManifestSkill is one canonical skill pinned by the SHA-256 of its source bytes.
type ManifestSkill struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Operations     []string `json:"operations"`
	OutputContract string   `json:"output_contract"`
	Budget         any      `json:"budget"`
	SHA256         string   `json:"sha256"`
}

// Manifest is the deterministic catalog document emitted by
// `gira skills manifest`. Skills is sorted by name; Digest is the catalog-level
// `sha256:<hex>` over the sorted "name\ndigest\n" lines.
type Manifest struct {
	Version string          `json:"version"`
	Digest  string          `json:"digest"`
	Skills  []ManifestSkill `json:"skills"`
}

// CanonicalBytes returns the exact embedded bytes of catalog/<name>/SKILL.md.
func CanonicalBytes(name string) ([]byte, error) {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || name == "." || name == ".." {
		return nil, fmt.Errorf("invalid skill name %q", name)
	}
	return catalog.ReadFile(path.Join("catalog", name, "SKILL.md"))
}

func shaHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// BuildManifest pins every canonical skill by content hash. The result is
// deterministic: skills sorted by name, catalog digest over the sorted entries.
func BuildManifest(version string) (*Manifest, error) {
	skills, err := Catalog()
	if err != nil {
		return nil, err
	}
	m := &Manifest{Version: version, Skills: make([]ManifestSkill, 0, len(skills))}
	h := sha256.New()
	for _, s := range skills {
		raw, err := CanonicalBytes(s.Name)
		if err != nil {
			return nil, err
		}
		digest := "sha256:" + shaHex(raw)
		m.Skills = append(m.Skills, ManifestSkill{
			Name:           s.Name,
			Description:    s.Description,
			Operations:     append([]string(nil), s.Operations...),
			OutputContract: s.OutputContract,
			Budget:         s.Budget,
			SHA256:         digest,
		})
		fmt.Fprintf(h, "%s\n%s\n", s.Name, digest)
	}
	m.Digest = "sha256:" + hex.EncodeToString(h.Sum(nil))
	return m, nil
}

// Markdown renders the manifest as a deterministic Markdown document.
func (m *Manifest) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Skills manifest (%s)\n\n", m.Version)
	fmt.Fprintf(&b, "Catalog digest: `%s`\n\n", m.Digest)
	b.WriteString("| Skill | SHA-256 | Description |\n")
	b.WriteString("|---|---|---|\n")
	for _, s := range m.Skills {
		fmt.Fprintf(&b, "| %s | `%s` | %s |\n", s.Name, s.SHA256, s.Description)
	}
	return b.String()
}

// CheckReport is the non-mutating drift report for one host family.
type CheckReport struct {
	Host    string   `json:"host"`
	InSync  bool     `json:"in_sync"`
	Missing []string `json:"missing"`
	Drifted []string `json:"drifted"`
	OK      bool     `json:"ok"`
}

// Check compares the expected host files against disk without touching it.
// A path is missing when its file does not exist; it is drifted when
// re-rendering the owned regions in memory differs from disk (or the owned
// region is unreadable). It never creates directories or modifies files.
func Check(host string, opts Options) (*CheckReport, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	files, err := planFiles(host, opts)
	if err != nil {
		return nil, err
	}
	rep := &CheckReport{Host: host, Missing: []string{}, Drifted: []string{}}
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(opts.Root, rel)
		old, err := os.ReadFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				rep.Missing = append(rep.Missing, filepath.ToSlash(rel))
				continue
			}
			return nil, err
		}
		next, err := files[rel](old)
		if err != nil {
			rep.Drifted = append(rep.Drifted, filepath.ToSlash(rel))
			continue
		}
		if !bytes.Equal(old, next) {
			rep.Drifted = append(rep.Drifted, filepath.ToSlash(rel))
		}
	}
	rep.InSync = len(rep.Missing) == 0 && len(rep.Drifted) == 0
	rep.OK = rep.InSync
	return rep, nil
}
