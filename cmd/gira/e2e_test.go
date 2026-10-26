package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// sanitizedGitEnv drops GIT_* variables that a hook runner may export (GIT_DIR,
// GIT_WORK_TREE, GIT_INDEX_FILE, ...). Without this, a git hook that runs the
// tests can redirect the fixture repository's init/add/commit into the product
// repository. GIT_CONFIG_* and identity/dates are kept.
func sanitizedGitEnv() []string {
	drop := map[string]bool{
		"GIT_DIR":                          true,
		"GIT_WORK_TREE":                    true,
		"GIT_INDEX_FILE":                   true,
		"GIT_OBJECT_DIRECTORY":             true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
		"GIT_COMMON_DIR":                   true,
		"GIT_CEILING_DIRECTORIES":          true,
	}
	var out []string
	for _, kv := range os.Environ() {
		if i := strings.IndexByte(kv, '='); i > 0 && drop[kv[:i]] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// setupDeterministicBareRepo initializes a git repo from sourceDir and clones it as a bare repo
// into targetBareDir using deterministic timestamps and committer identities.
func setupDeterministicBareRepo(t *testing.T, sourceDir, targetBareDir, commitMsg string) string {
	t.Helper()

	workTree := t.TempDir()

	runGit := func(dir string, env []string, args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(sanitizedGitEnv(), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v in %s failed: %v\nOutput: %s", args, dir, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}

	// 1. Copy files from sourceDir to workTree
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(workTree, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		t.Fatalf("failed to copy fixtures from %s to %s: %v", sourceDir, workTree, err)
	}

	// 2. Initialize work tree git repo
	runGit(workTree, nil, "init", "-b", "main")
	runGit(workTree, nil, "config", "user.name", "Gira Dogfood Runner")
	runGit(workTree, nil, "config", "user.email", "dogfood@gira.local")
	runGit(workTree, nil, "add", ".")

	// Fix identity and dates explicitly so a leaked GIT_AUTHOR_*/GIT_COMMITTER_*
	// in the environment (for example from a git hook runner) can't change the
	// commit SHA the manifest pins.
	commitEnv := []string{
		"GIT_AUTHOR_NAME=Gira Dogfood Runner",
		"GIT_AUTHOR_EMAIL=dogfood@gira.local",
		"GIT_COMMITTER_NAME=Gira Dogfood Runner",
		"GIT_COMMITTER_EMAIL=dogfood@gira.local",
		"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2026-01-01T00:00:00Z",
	}
	runGit(workTree, commitEnv, "commit", "-m", commitMsg)

	commitSHA := runGit(workTree, nil, "rev-parse", "HEAD")

	// 3. Clone bare into targetBareDir
	if err := os.MkdirAll(filepath.Dir(targetBareDir), 0755); err != nil {
		t.Fatalf("failed to create target bare repo parent directory: %v", err)
	}
	runGit(workTree, nil, "clone", "--bare", workTree, targetBareDir)

	return commitSHA
}

// setupDogfoodBareCache prepares the bare repository caches for all repositories declared in the manifest.
func setupDogfoodBareCache(t *testing.T, m *manifest.Manifest, cacheDir string) {
	t.Helper()

	// Locate testdata fixtures directory
	testdataRoot, err := filepath.Abs(filepath.Join("..", "..", "testdata", "dogfood"))
	if err != nil {
		t.Fatalf("failed to resolve testdata root path: %v", err)
	}

	for _, repo := range m.Repositories {
		repoHash := gitcache.RepoHash(repo.URL)
		targetBareDir := filepath.Join(cacheDir, repoHash+".git")

		var fixtureSubDir string
		var commitMsg string

		switch repo.ID {
		case "dogfood-services":
			fixtureSubDir = filepath.Join(testdataRoot, "services")
			commitMsg = "feat(services): initial services architecture corpus"
		case "dogfood-charts":
			fixtureSubDir = filepath.Join(testdataRoot, "charts")
			commitMsg = "feat(charts): initial helm charts corpus"
		default:
			t.Fatalf("unknown repo ID in dogfood manifest: %s", repo.ID)
		}

		actualSHA := setupDeterministicBareRepo(t, fixtureSubDir, targetBareDir, commitMsg)
		if actualSHA != repo.SHA {
			t.Fatalf("commit SHA mismatch for repo %s: expected %s (from manifest), got %s (from fixtures)",
				repo.ID, repo.SHA, actualSHA)
		}
	}
}

// setNetworkSandboxEnvironment sets proxy and config variables to ensure zero network egress can succeed.
func setNetworkSandboxEnvironment(t *testing.T) {
	t.Helper()
	// Point network proxies to dead local endpoints
	t.Setenv("http_proxy", "http://127.0.0.1:0")
	t.Setenv("https_proxy", "http://127.0.0.1:0")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:0")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:0")
	t.Setenv("ALL_PROXY", "socks5://127.0.0.1:0")
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("CI", "true")
	t.Setenv("NO_BROWSER", "1")
}

func TestE2E_DogfoodFullLifecycleOffline(t *testing.T) {
	setNetworkSandboxEnvironment(t)

	manifestPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", "dogfood", "gira.yaml"))
	if err != nil {
		t.Fatalf("failed to resolve manifest path: %v", err)
	}

	m, err := manifest.ParseFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to parse dogfood manifest: %v", err)
	}

	cacheDir := t.TempDir()
	setupDogfoodBareCache(t, m, cacheDir)

	outDir := t.TempDir()
	macroMapPath := filepath.Join(outDir, "macro-map.json")
	htmlExportPath := filepath.Join(outDir, "architecture.html")
	mdExportPath := filepath.Join(outDir, "architecture.md")

	// -------------------------------------------------------------
	// 1. gira index --offline
	// -------------------------------------------------------------
	var stdout, stderr bytes.Buffer
	indexArgs := []string{
		"index", manifestPath,
		"--offline",
		"--cache-dir", cacheDir,
		"--output", macroMapPath,
	}

	if err := Run(indexArgs, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira index --offline failed: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Indexed") || !strings.Contains(stdout.String(), m.TargetSystem.Name) {
		t.Errorf("unexpected index stdout: %s", stdout.String())
	}

	// Verify generated macro-map.json integrity
	s := store.NewStore("")
	mm, err := s.LoadFile(macroMapPath)
	if err != nil {
		t.Fatalf("failed to load indexed macro-map.json: %v", err)
	}

	if mm.TargetSystem != "dogfood-system" {
		t.Errorf("expected TargetSystem dogfood-system, got %q", mm.TargetSystem)
	}

	// Verify required node IDs exist in the macro map
	nodeMap := make(map[string]domain.Node)
	for _, n := range mm.Nodes {
		nodeMap[n.ID] = n
	}

	expectedNodes := []struct {
		id   string
		kind domain.NodeType
	}{
		{"service:frontend", domain.NodeService},
		{"service:order", domain.NodeService},
		{"service:redis", domain.NodeInfrastructure},
		{"service:postgres", domain.NodeInfrastructure},
		{"contract:OrderService", domain.NodeContract},
		{"library:google.golang.org/grpc", domain.NodeLibrary},
	}

	for _, exp := range expectedNodes {
		node, ok := nodeMap[exp.id]
		if !ok {
			t.Errorf("expected node %q in macro map, not found", exp.id)
			continue
		}
		if node.Kind != exp.kind {
			t.Errorf("expected node %q to have kind %q, got %q", exp.id, exp.kind, node.Kind)
		}
		if len(node.Evidence) == 0 {
			t.Errorf("expected node %q to have evidence records, got 0", exp.id)
		}
	}

	// Verify overrides took effect
	frontendNode := nodeMap["service:frontend"]
	if frontendNode.Metadata["tier"] != "edge" {
		t.Errorf("expected metadata tier: edge on service:frontend, got %v", frontendNode.Metadata["tier"])
	}

	// Verify contract methods extracted from protobuf
	contractNode := nodeMap["contract:OrderService"]
	methodsRaw, ok := contractNode.Metadata["methods"].([]any)
	if !ok {
		// Try []string directly if not serialized through JSON
		if mStr, okStr := contractNode.Metadata["methods"].([]string); okStr {
			methodsRaw = make([]any, len(mStr))
			for i, v := range mStr {
				methodsRaw[i] = v
			}
		}
	}
	if len(methodsRaw) != 2 {
		t.Errorf("expected 2 RPC methods in contract:OrderService, got %d (%v)", len(methodsRaw), contractNode.Metadata["methods"])
	}

	// Verify relationships
	relMap := make(map[string]domain.Relationship)
	for _, r := range mm.Relationships {
		relMap[r.ID] = r
	}

	// Frontend consumes order: depends_on-only in the dogfood compose file
	// (no frontend call-site or env wiring), so J2 prunes it as start
	// ordering. The absence itself is the assertion.
	if relFrontendOrder, ok := relMap["rel:frontend-consumes-order"]; ok {
		t.Errorf("ordering-only rel:frontend-consumes-order must be pruned (J2), got %+v", relFrontendOrder)
	}

	// Frontend accesses redis
	relFrontendRedis := relMap["rel:frontend-accesses-redis"]
	if relFrontendRedis.Verb != domain.VerbAccesses {
		t.Errorf("expected rel:frontend-accesses-redis to have verb accesses, got %q", relFrontendRedis.Verb)
	}

	// Order accesses postgres
	relOrderPostgres := relMap["rel:order-accesses-postgres"]
	if relOrderPostgres.Verb != domain.VerbAccesses {
		t.Errorf("expected rel:order-accesses-postgres to have verb accesses, got %q", relOrderPostgres.Verb)
	}

	// Order provides contract OrderService
	relOrderProvides := relMap["rel:order-provides-orderservice"]
	if relOrderProvides.Verb != domain.VerbProvides {
		t.Errorf("expected rel:order-provides-orderservice to have verb provides, got %q", relOrderProvides.Verb)
	}

	// Unsupported boundary emitted without guessing
	if len(mm.Unsupported) == 0 {
		t.Errorf("expected unsupported boundary records for Python worker, got 0")
	} else {
		foundPy := false
		for _, ub := range mm.Unsupported {
			if strings.Contains(ub.Path, "analytics.py") {
				foundPy = true
				if ub.Reason != "no extractor registered for file extension" {
					t.Errorf("unexpected unsupported reason: %s", ub.Reason)
				}
				break
			}
		}
		if !foundPy {
			t.Errorf("expected scripts/analytics.py in unsupported boundaries")
		}
	}

	// -------------------------------------------------------------
	// 2. gira impact
	// -------------------------------------------------------------
	// 2a. JSON format
	stdout.Reset()
	stderr.Reset()
	impactArgsJSON := []string{
		"impact", "service:order",
		"--index", macroMapPath,
		"--format", "json",
	}
	if err := Run(impactArgsJSON, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira impact json failed: %v\nstderr: %s", err, stderr.String())
	}

	var briefing slices.ImpactBriefing
	if err := json.Unmarshal(stdout.Bytes(), &briefing); err != nil {
		t.Fatalf("failed to parse impact briefing JSON: %v\nOutput: %s", err, stdout.String())
	}
	if briefing.TargetID != "service:order" {
		t.Errorf("expected briefing target_id service:order, got %s", briefing.TargetID)
	}
	if len(briefing.UpstreamServices)+len(briefing.DownstreamServices)+len(briefing.AffectedContracts)+len(briefing.MediationEdges) == 0 {
		t.Errorf("expected positive impacted count for service:order, got 0")
	}

	// 2b. Markdown format
	stdout.Reset()
	stderr.Reset()
	impactArgsMD := []string{
		"impact", "service:order",
		"--index", macroMapPath,
		"--format", "md",
	}
	if err := Run(impactArgsMD, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira impact md failed: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Impact Briefing: service:order") {
		t.Errorf("expected markdown header in impact briefing, got: %s", stdout.String())
	}

	// -------------------------------------------------------------
	// 3. gira slice (summary, contracts, deep)
	// -------------------------------------------------------------
	tiers := []string{"summary", "contracts", "deep"}
	for _, tier := range tiers {
		stdout.Reset()
		stderr.Reset()
		sliceArgs := []string{
			"slice", "service:order",
			"--tier", tier,
			"--index", macroMapPath,
			"--format", "json",
		}
		if err := Run(sliceArgs, nil, &stdout, &stderr); err != nil {
			t.Fatalf("gira slice --tier %s failed: %v\nstderr: %s", tier, err, stderr.String())
		}

		var slice slices.ContextSlice
		if err := json.Unmarshal(stdout.Bytes(), &slice); err != nil {
			t.Fatalf("failed to parse slice JSON for tier %s: %v\nOutput: %s", tier, err, stdout.String())
		}
		if slice.Tier != slices.SliceTier(tier) {
			t.Errorf("expected slice tier %s, got %s", tier, slice.Tier)
		}
		if slice.FocalNode.ID != "service:order" {
			t.Errorf("expected focal node ID service:order, got %s", slice.FocalNode.ID)
		}
		if len(slice.Neighbors) == 0 && len(slice.Relationships) == 0 {
			t.Errorf("expected non-empty neighbors or relationships for tier %s", tier)
		}

		// Markdown slice rendering
		stdout.Reset()
		stderr.Reset()
		sliceArgsMD := []string{
			"slice", "service:order",
			"--tier", tier,
			"--index", macroMapPath,
			"--format", "md",
		}
		if err := Run(sliceArgsMD, nil, &stdout, &stderr); err != nil {
			t.Fatalf("gira slice md --tier %s failed: %v", tier, err)
		}
		if !strings.Contains(stdout.String(), "# Context Slice:") {
			t.Errorf("expected markdown header in slice output for tier %s", tier)
		}
	}

	// -------------------------------------------------------------
	// 4. gira export --format html
	// -------------------------------------------------------------
	stdout.Reset()
	stderr.Reset()
	exportHTMLArgs := []string{
		"export",
		"--format", "html",
		"--output", htmlExportPath,
		"--index", macroMapPath,
	}
	if err := Run(exportHTMLArgs, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira export --format html failed: %v\nstderr: %s", err, stderr.String())
	}

	htmlBytes, err := os.ReadFile(htmlExportPath)
	if err != nil {
		t.Fatalf("failed to read exported HTML file: %v", err)
	}

	if len(htmlBytes) < 500 {
		t.Errorf("exported HTML file too small (%d bytes)", len(htmlBytes))
	}

	htmlContent := string(htmlBytes)

	// Structural HTML assertions
	requiredHTMLSnippets := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"Gira-G Architecture Export: dogfood-system",
		"<svg viewBox=",
		"service:order",
		"service:frontend",
		"contract:OrderService",
		"badge-extracted",
		"Unsupported Boundaries & Known Unknowns",
		"analytics.py",
	}

	for _, snippet := range requiredHTMLSnippets {
		if !strings.Contains(htmlContent, snippet) {
			t.Errorf("exported HTML missing expected snippet %q", snippet)
		}
	}

	// -------------------------------------------------------------
	// 5. gira export --format md
	// -------------------------------------------------------------
	stdout.Reset()
	stderr.Reset()
	exportMDArgs := []string{
		"export",
		"--format", "md",
		"--output", mdExportPath,
		"--index", macroMapPath,
	}
	if err := Run(exportMDArgs, nil, &stdout, &stderr); err != nil {
		t.Fatalf("gira export --format md failed: %v\nstderr: %s", err, stderr.String())
	}

	mdBytes, err := os.ReadFile(mdExportPath)
	if err != nil {
		t.Fatalf("failed to read exported Markdown file: %v", err)
	}

	mdContent := string(mdBytes)
	requiredMDSnippets := []string{
		"# Target System Architecture: dogfood-system",
		"## Services",
		"## Infrastructure",
		"## Contracts",
		"## Unsupported Boundaries & Known Unknowns",
		"analytics.py",
	}

	for _, snippet := range requiredMDSnippets {
		if !strings.Contains(mdContent, snippet) {
			t.Errorf("exported Markdown missing expected snippet %q", snippet)
		}
	}
}

func TestE2E_OfflineStrictZeroNetworkEgress(t *testing.T) {
	setNetworkSandboxEnvironment(t)

	manifestPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", "dogfood", "gira.yaml"))
	if err != nil {
		t.Fatalf("failed to resolve manifest path: %v", err)
	}

	// 1. Missing cache directory must fail immediately in --offline mode without network calls
	emptyCacheDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err = Run([]string{"index", manifestPath, "--offline", "--cache-dir", emptyCacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error in --offline mode with empty cache, got nil")
	}

	if !strings.Contains(err.Error(), "repository not found in local cache") &&
		!strings.Contains(err.Error(), "failed to ensure repository") {
		t.Errorf("unexpected error message for missing cache: %v", err)
	}

	// 2. Cache exists but commit SHA is missing must fail immediately in --offline mode
	m, err := manifest.ParseFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	partialCacheDir := t.TempDir()
	// Create an empty bare git repo without the pinned commit
	firstRepo := m.Repositories[0]
	firstRepoDir := filepath.Join(partialCacheDir, gitcache.RepoHash(firstRepo.URL)+".git")
	if err := os.MkdirAll(firstRepoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	cmd := exec.Command("git", "init", "--bare", firstRepoDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %v, out: %s", err, string(out))
	}

	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestPath, "--offline", "--cache-dir", partialCacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error in --offline mode when commit is missing, got nil")
	}

	if !strings.Contains(err.Error(), "pinned commit") &&
		!strings.Contains(err.Error(), "not found") &&
		!strings.Contains(err.Error(), "failed to ensure repository") {
		t.Errorf("unexpected error message for missing commit: %v", err)
	}
}
