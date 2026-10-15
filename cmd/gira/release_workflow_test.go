package main

// SC4 (release workflow, F4): .github/workflows/release.yml triggers on
// v*.*.* tags plus workflow_dispatch, keeps least-privilege permissions,
// runs preflight plus GoReleaser, and stays clear of the quarantined
// comparative module (ADR 0012 rule 3).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func readReleaseWorkflowYML(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(releaseRepoRoot(t), ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read .github/workflows/release.yml: %v", err)
	}
	var v any
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("release.yml is not valid YAML: %v", err)
	}
	return string(raw)
}

// releaseWorkflowForbidden mirrors the static isolation check of the
// quarantined comparative module (ADR 0012 rule 3). It is assembled from
// fragments so this guarded file never carries a literal the check would flag.
var releaseWorkflowForbidden = regexp.MustCompile(
	`(?i)(^|[^a-z0-9_-])\.?/?` + "ben" + "ch" + `([^a-z0-9_-]|$)` +
		`|` + "gira" + "-" + "ben" + "ch" +
		`|` + "graph" + "ify" +
		`|` + "under" + "stand" + `[-_ ]?` + "any" + "thing" +
		`|` + "ego" + "nex" + "-ai",
)

// TestSCReleaseWorkflow is SC4: the tag-triggered release workflow obeys its
// trigger, permission, preflight, GoReleaser and quarantine contract.
func TestSCReleaseWorkflow(t *testing.T) {
	yml := readReleaseWorkflowYML(t)

	// Triggers: tag push v*.*.* plus manual dispatch.
	for _, want := range []string{"v*.*.*", "workflow_dispatch", "push:", "tags:"} {
		if !strings.Contains(yml, want) {
			t.Errorf("release.yml missing trigger %q", want)
		}
	}

	// Least privilege: top-level contents: read; the release job escalates
	// only contents plus OIDC.
	for _, want := range []string{"contents: read", "contents: write", "id-token: write"} {
		if !strings.Contains(yml, want) {
			t.Errorf("release.yml missing permission %q", want)
		}
	}

	// Preflight: full-history checkout, mise toolchain, Node 22 plus the
	// standalone validator, and the race-detector suite.
	for _, want := range []string{
		"actions/checkout", "fetch-depth: 0",
		"jdx/mise-action@v5",
		"node-version: 22", "node tools/validate/validate.js",
		"go test -race ./...",
	} {
		if !strings.Contains(yml, want) {
			t.Errorf("release.yml missing preflight step %q", want)
		}
	}

	// Snapshot dry-run on workflow_dispatch, unsigned, plus the skills
	// manifest check.
	for _, want := range []string{
		"goreleaser release --snapshot --clean --skip=sign",
		"skills manifest",
	} {
		if !strings.Contains(yml, want) {
			t.Errorf("release.yml missing snapshot dry-run step %q", want)
		}
	}

	// Tag release job: gated on refs/tags/v, publishing via GoReleaser with
	// GITHUB_TOKEN and keyless cosign OIDC.
	for _, want := range []string{
		"startsWith(github.ref", "refs/tags/v",
		"goreleaser release --clean",
		"GITHUB_TOKEN",
		"cosign",
	} {
		if !strings.Contains(yml, want) {
			t.Errorf("release.yml missing tag-release step %q", want)
		}
	}

	// ADR 0012 rule 3: zero references to the quarantined module.
	for i, line := range strings.Split(yml, "\n") {
		if m := releaseWorkflowForbidden.FindString(line); m != "" {
			t.Errorf("release.yml:%d references quarantined material (%q)", i+1, strings.TrimSpace(m))
		}
	}
}
