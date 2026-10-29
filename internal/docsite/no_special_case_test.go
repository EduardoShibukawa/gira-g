package docsite

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Frozen tests for PLAN-0009 TASK-01: the docsite has no per-project special
// case. Tour, workflows, pillars and title come from the target repo's
// `docsite:` section in gira.yaml (read in manifest.go, passed via SiteConfig).

const docsiteYAML = `version: 1
target_system:
  name: demo-harness
  description: "Demo harness"
repositories:
  - id: demo
    url: "https://github.com/example/demo.git"
    ref: "1.0.0"
    sha: "dedc0178918e260823323b8d95005a8cb924b007"
    path_scope: "."
docsite:
  title: "Demo Harness"
  tour:
    - title: "Contract Proving"
      domain: "Verification"
      description: "Every change is a PLAN.md contract proven RED first."
      invariants:
        - "25-Line Bound: PLAN.md contracts stay concise."
      key_files:
        - "tasks/T33/PLAN.md"
      sample_cmd: "demo-tool verify tasks/T33/PLAN.md"
      snippet: "demo-tool verify tasks/T33/PLAN.md"
    - title: "Ledger Audit"
      domain: "Audit"
      description: "Receipts seal into an append-only ledger."
      invariants:
        - "Immutable Hash Chain: blocks cannot be rewritten."
      key_files:
        - "ledger.db"
      sample_cmd: "demo-tool ledger verify"
      snippet: "demo-tool ledger verify"
  workflows:
    - title: "Contract Verification"
      description: "Deterministic RED-then-EXEC proofing."
      command: "demo-tool verify tasks/T33/PLAN.md"
    - title: "Tamper-Evident Ledger"
      description: "Audit hash-chained execution receipts."
      command: "demo-tool ledger verify"
  pillars:
    - icon: "Scales"
      title: "Contract Prover"
      description: "Formal PLAN.md task specifications with sealed receipts."
    - icon: "Lock"
      title: "Cryptographic Ledger"
      description: "Append-only hash-chained receipts."
`

func parseDocsiteConfig(t *testing.T) *manifest.DocsiteConfig {
	t.Helper()
	m, err := manifest.Parse([]byte(docsiteYAML))
	if err != nil {
		t.Fatalf("manifest.Parse rejected a docsite section: %v", err)
	}
	if m.Docsite == nil {
		t.Fatal("manifest.Manifest.Docsite is nil; the `docsite:` section was not read")
	}
	return m.Docsite
}

func TestDocsiteSectionParsedFromManifest(t *testing.T) {
	m, err := manifest.Parse([]byte(docsiteYAML))
	if err != nil {
		t.Fatalf("manifest.Parse failed: %v", err)
	}
	if m.Docsite == nil {
		t.Fatal("manifest.Manifest.Docsite is nil")
	}
	if m.Docsite.Title != "Demo Harness" {
		t.Errorf("Docsite.Title = %q, want %q", m.Docsite.Title, "Demo Harness")
	}
	if len(m.Docsite.Tour) != 2 {
		t.Fatalf("len(Docsite.Tour) = %d, want 2", len(m.Docsite.Tour))
	}
	step := m.Docsite.Tour[0]
	if step.Title != "Contract Proving" || step.Domain != "Verification" || step.SampleCmd != "demo-tool verify tasks/T33/PLAN.md" {
		t.Errorf("Tour[0] = %+v, want title/domain/sample_cmd from YAML", step)
	}
	if len(step.Invariants) != 1 || len(step.KeyFiles) != 1 {
		t.Errorf("Tour[0] invariants/key_files = %v / %v, want one each", step.Invariants, step.KeyFiles)
	}
	if len(m.Docsite.Workflows) != 2 {
		t.Fatalf("len(Docsite.Workflows) = %d, want 2", len(m.Docsite.Workflows))
	}
	if w := m.Docsite.Workflows[1]; w.Title != "Tamper-Evident Ledger" || w.Command != "demo-tool ledger verify" {
		t.Errorf("Workflows[1] = %+v, want title/command from YAML", w)
	}
	if len(m.Docsite.Pillars) != 2 {
		t.Fatalf("len(Docsite.Pillars) = %d, want 2", len(m.Docsite.Pillars))
	}
	if p := m.Docsite.Pillars[0]; p.Icon != "Scales" || p.Title != "Contract Prover" || p.Description == "" {
		t.Errorf("Pillars[0] = %+v, want icon/title/description from YAML", p)
	}
}

func TestManifestWithoutDocsiteSection(t *testing.T) {
	m, err := manifest.Parse([]byte(`version: 1
target_system:
  name: plain-repo
  description: "No docsite section"
repositories:
  - id: plain
    url: "https://github.com/example/plain.git"
    ref: "1.0.0"
    sha: "dedc0178918e260823323b8d95005a8cb924b007"
    path_scope: "."
`))
	if err != nil {
		t.Fatalf("manifest.Parse failed on a manifest without docsite: %v", err)
	}
	if m.Docsite != nil {
		t.Errorf("Docsite = %+v, want nil when the section is absent", m.Docsite)
	}
}

func TestFormatProjectTitleHasNoProjectNameCases(t *testing.T) {
	if got := FormatProjectTitle("p-inga"); got != "P Inga" {
		t.Errorf("FormatProjectTitle(%q) = %q, want generic %q", "p-inga", got, "P Inga")
	}
	if got := FormatProjectTitle("pinga"); got != "Pinga" {
		t.Errorf("FormatProjectTitle(%q) = %q, want generic %q", "pinga", got, "Pinga")
	}
}

func TestGeneratedPagesHaveNoPingaSpecialCaseWithoutConfig(t *testing.T) {
	for _, name := range []string{"p-inga", "pinga"} {
		cfg := &SiteConfig{TargetName: name}
		tour := GenerateTourPageHTML(cfg, nil)
		landing := GenerateLandingPageHTML(cfg, nil)
		for label, html := range map[string]string{"tour": tour, "landing": landing} {
			lower := strings.ToLower(html)
			if strings.Contains(lower, "pinga") || strings.Contains(lower, "p-inga") {
				t.Errorf("TargetName %q: generated %s page mentions pinga; no-name special case is gone and generic output is expected", name, label)
			}
		}
		if !strings.Contains(tour, "Living Architecture") {
			t.Errorf("TargetName %q: tour page lost the generic path", name)
		}
	}
}

func TestTourPageUsesDocsiteConfig(t *testing.T) {
	doc := parseDocsiteConfig(t)
	cfg := &SiteConfig{TargetName: "demo-harness", Docsite: doc}
	tour := GenerateTourPageHTML(cfg, nil)
	for _, want := range []string{
		"Contract Proving",
		"Ledger Audit",
		"demo-tool verify tasks/T33/PLAN.md",
		"demo-tool ledger verify",
	} {
		if !strings.Contains(tour, want) {
			t.Errorf("tour page missing configured tour content %q", want)
		}
	}
	if strings.Contains(tour, "System Overview & Entrypoints") || strings.Contains(tour, "System Overview &amp; Entrypoints") {
		t.Error("tour page still renders the generic fallback although a docsite tour config exists")
	}
}

func TestLandingPageUsesDocsiteConfig(t *testing.T) {
	doc := parseDocsiteConfig(t)
	cfg := &SiteConfig{TargetName: "demo-harness", Docsite: doc}
	landing := GenerateLandingPageHTML(cfg, nil)
	for _, want := range []string{
		"Demo Harness",
		"Contract Verification",
		"Tamper-Evident Ledger",
		"demo-tool verify tasks/T33/PLAN.md",
		"demo-tool ledger verify",
		"Contract Prover",
		"Cryptographic Ledger",
	} {
		if !strings.Contains(landing, want) {
			t.Errorf("landing page missing configured docsite content %q", want)
		}
	}
}
