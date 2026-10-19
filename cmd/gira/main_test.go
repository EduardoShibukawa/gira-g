package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

func createTestIndex(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "macro-map.json")

	mm := domain.NewMacroMap("test-corpus")
	_ = mm.AddNode(domain.Node{
		ID:   "service:order",
		Kind: domain.NodeService,
		Name: "order-service",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-order",
					CommitSHA: "abc1112223334445556667778889990001112223",
					Path:      "main.go",
					StartLine: 1,
					EndLine:   50,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "default",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      time.Now(),
			},
		},
	})
	_ = mm.AddNode(domain.Node{
		ID:   "contract:order-grpc",
		Kind: domain.NodeContract,
		Name: "OrderService",
		Metadata: map[string]any{
			"methods": []string{"PlaceOrder", "GetOrder"},
		},
	})
	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:order-provides-grpc",
		SourceID: "service:order",
		TargetID: "contract:order-grpc",
		Verb:     domain.VerbProvides,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "repo-order",
					CommitSHA: "abc1112223334445556667778889990001112223",
					Path:      "service.go",
				},
				Origin: domain.OriginExtracted,
				Status: domain.StatusCurrent,
			},
		},
	})
	mm.Unsupported = append(mm.Unsupported, domain.UnsupportedBoundary{
		Path:   "legacy/billing.js",
		Reason: "unsupported javascript",
	})

	s := store.NewStore("")
	data, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to serialize test map: %v", err)
	}
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		t.Fatalf("failed to write test index: %v", err)
	}
	_ = s

	return indexPath
}

func TestCLI_VersionAndHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// version
	if err := Run([]string{"version"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "gira ") {
		t.Errorf("expected version output, got %q", stdout.String())
	}

	// help
	stdout.Reset()
	if err := Run([]string{"--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage output, got %q", stdout.String())
	}

	// no args
	stdout.Reset()
	if err := Run([]string{}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("no args failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage output, got %q", stdout.String())
	}
}

func TestCLI_Impact(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format (default)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"impact", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("impact json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"target_id": "service:order"`) {
		t.Errorf("expected target_id in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"impact", "service:order", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("impact md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Impact Briefing: service:order") {
		t.Errorf("expected markdown header in output, got %s", stdout.String())
	}
}

func TestCLI_Slice(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format (default)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"slice", "service:order", "--tier", "contracts", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("slice json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"focal_node"`) {
		t.Errorf("expected focal_node in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"slice", "service:order", "--tier", "contracts", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("slice md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Context Slice: order-service") {
		t.Errorf("expected markdown header in output, got %s", stdout.String())
	}
}

func TestCLI_Explain(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format for node
	var stdout, stderr bytes.Buffer
	err := Run([]string{"explain", "--node", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("explain json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "abc1112223334445556667778889990001112223") {
		t.Errorf("expected locator SHA in explain output, got %s", stdout.String())
	}

	// Markdown format for edge
	stdout.Reset()
	err = Run([]string{"explain", "--edge", "rel:order-provides-grpc", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("explain edge md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "# Evidence Report: rel:order-provides-grpc") {
		t.Errorf("expected markdown report in output, got %s", stdout.String())
	}

	// Positional node ID fallback: gira explain service:order
	stdout.Reset()
	err = Run([]string{"explain", "service:order", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("positional explain failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"node_id": "service:order"`) {
		t.Errorf("expected node_id in output, got %s", stdout.String())
	}
}

func TestCLI_Locate(t *testing.T) {
	indexPath := createTestIndex(t)

	// JSON format
	var stdout, stderr bytes.Buffer
	err := Run([]string{"locate", "billing", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("locate json failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "legacy/billing.js") {
		t.Errorf("expected unsupported match in output, got %s", stdout.String())
	}

	// Markdown format
	stdout.Reset()
	err = Run([]string{"locate", "PlaceOrder", "--index", indexPath, "--format", "md"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("locate md failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "RPC Method `PlaceOrder`") {
		t.Errorf("expected method match in md output, got %s", stdout.String())
	}
}

func TestCLI_ServeMCP(t *testing.T) {
	indexPath := createTestIndex(t)

	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdin.WriteString(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")

	err := Run([]string{"serve", "mcp", "--index", indexPath}, &stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("serve mcp failed: %v", err)
	}
	if !strings.Contains(stdout.String(), `"id":1`) {
		t.Errorf("expected ping response, got %s", stdout.String())
	}
}

func TestCLI_ErrorHandling(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Unknown command
	if err := Run([]string{"unknown_cmd"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for unknown command, got nil")
	}

	// Missing arguments
	if err := Run([]string{"impact"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing impact target, got nil")
	}

	if err := Run([]string{"slice"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing slice node, got nil")
	}

	if err := Run([]string{"locate"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing locate query, got nil")
	}

	if err := Run([]string{"explain"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for missing explain args, got nil")
	}

	// Nonexistent index file
	if err := Run([]string{"impact", "service:order", "--index", "/nonexistent/index.json"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for nonexistent index file, got nil")
	}

	// Invalid export format
	if err := Run([]string{"export", "--format", "invalid_fmt"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected error for invalid export format, got nil")
	}
}

func TestCLI_Export(t *testing.T) {
	indexPath := createTestIndex(t)

	// 1. Export Markdown to stdout (default format md)
	var stdout, stderr bytes.Buffer
	err := Run([]string{"export", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export md failed: %v", err)
	}
	outMD := stdout.String()
	if !strings.Contains(outMD, "# Target System Architecture: test-corpus") {
		t.Errorf("expected Target System Architecture in export md output, got:\n%s", outMD)
	}
	if !strings.Contains(outMD, "## Services (1)") {
		t.Errorf("expected Services (1) in export md output")
	}
	if !strings.Contains(outMD, "## Unsupported Boundaries & Known Unknowns (1)") {
		t.Errorf("expected Unsupported Boundaries section in export md output")
	}

	// 2. Export HTML to stdout
	stdout.Reset()
	err = Run([]string{"export", "--format", "html", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export html failed: %v", err)
	}
	outHTML := stdout.String()
	if !strings.Contains(outHTML, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in export html output")
	}
	if !strings.Contains(outHTML, "<svg viewBox=") {
		t.Errorf("expected SVG in export html output")
	}
	if !strings.Contains(outHTML, "badge-extracted") {
		t.Errorf("expected origin badge in export html output")
	}

	// 3. Export to file via --output
	tmpOutFile := filepath.Join(t.TempDir(), "sub", "export.html")
	stdout.Reset()
	err = Run([]string{"export", "--format", "html", "--output", tmpOutFile, "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export --output failed: %v", err)
	}
	data, err := os.ReadFile(tmpOutFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in written export file")
	}
}

func TestCLI_View(t *testing.T) {
	indexPath := createTestIndex(t)

	// 1. View with default temporary directory
	var stdout, stderr bytes.Buffer
	err := Run([]string{"view", "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("view failed: %v", err)
	}
	viewOut := stdout.String()
	if !strings.Contains(viewOut, "Exported HTML view to: file://") {
		t.Errorf("expected file:// URI in view output, got: %s", viewOut)
	}

	// 2. View with specified --output directory
	customDir := filepath.Join(t.TempDir(), "custom-view")
	stdout.Reset()
	err = Run([]string{"view", "--output", customDir, "--index", indexPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("view with custom output dir failed: %v", err)
	}
	expectedFile := filepath.Join(customDir, "index.html")
	if fi, err := os.Stat(expectedFile); err != nil || fi.IsDir() {
		t.Fatalf("expected view index.html at %s", expectedFile)
	}
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read view index.html: %v", err)
	}
	if !strings.Contains(string(data), "Gira-G Architecture Export: test-corpus") {
		t.Errorf("expected title in view index.html")
	}
}

func createGitRepoWithFiles(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	repoDir := t.TempDir()

	runGit := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v, out: %s", args, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}

	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")

	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", relPath, err)
		}
	}

	runGit("add", ".")
	runGit("commit", "-m", "Initial commit")
	commitSHA := runGit("rev-parse", "HEAD")
	return repoDir, commitSHA
}

func TestCLI_Index(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	composeContent := `
version: '3.8'
services:
  frontend:
    image: otel/demo-frontend:v1.0
  cartservice:
    image: otel/demo-cart:v1.0
`
	repoDir, sha := createGitRepoWithFiles(t, map[string]string{
		"docker-compose.yaml": composeContent,
	})

	cacheDir := t.TempDir()
	outDir := t.TempDir()
	outputPath := filepath.Join(outDir, "custom-macro-map.json")

	manifestDir := t.TempDir()
	manifestPath := filepath.Join(manifestDir, "gira.yaml")

	manifestContent := strings.Join([]string{
		"version: 1",
		"target_system:",
		"  name: index-test-system",
		"repositories:",
		"  - id: repo-main",
		`    url: "` + filepath.ToSlash(repoDir) + `"`,
		`    sha: "` + sha + `"`,
		"    path_scope: .",
		"overrides:",
		"  relationships:",
		`    - source_id: "service:frontend"`,
		`      target_id: "service:cartservice"`,
		`      verb: "consumes"`,
		`    - source_id: "service:frontend"`,
		`      target_id: "contract:custom-checkout"`,
		`      verb: "consumes"`,
		"  metadata:",
		`    "service:frontend":`,
		`      team: "web-core"`,
	}, "\n")

	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// 1. Happy path: index with --output
	var stdout, stderr bytes.Buffer
	err := Run([]string{"index", manifestPath, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira index failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Indexed") || !strings.Contains(stdout.String(), "index-test-system") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}

	// Verify loaded macro map
	s := store.NewStore("")
	mm, err := s.LoadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to load indexed macro map: %v", err)
	}

	if mm.TargetSystem != "index-test-system" {
		t.Errorf("expected target system index-test-system, got %s", mm.TargetSystem)
	}

	var frontendNode, checkoutNode *domain.Node
	for i := range mm.Nodes {
		if mm.Nodes[i].ID == "service:frontend" {
			frontendNode = &mm.Nodes[i]
		}
		if mm.Nodes[i].ID == "contract:custom-checkout" {
			checkoutNode = &mm.Nodes[i]
		}
	}

	if frontendNode == nil {
		t.Fatalf("expected service:frontend node")
	}
	if frontendNode.Metadata["team"] != "web-core" {
		t.Errorf("expected team web-core, got %v", frontendNode.Metadata["team"])
	}

	// contract:custom-checkout does not exist in docker-compose.yaml -> must NOT be fabricated as a dummy node
	if checkoutNode != nil {
		t.Fatalf("expected contract:custom-checkout NOT to be fabricated as a phantom node")
	}

	// Verify valid override relationship to existing service:cartservice
	var overrideRel *domain.Relationship
	for i := range mm.Relationships {
		if mm.Relationships[i].TargetID == "service:cartservice" {
			overrideRel = &mm.Relationships[i]
			break
		}
	}
	if overrideRel == nil {
		t.Fatalf("expected relationship to service:cartservice")
	}
	if overrideRel.Verb != domain.VerbConsumes {
		t.Errorf("expected consumes verb, got %s", overrideRel.Verb)
	}
	if len(overrideRel.Evidence) == 0 || overrideRel.Evidence[0].Origin != "declared:override" {
		t.Errorf("expected declared:override origin")
	}
	if overrideRel.Evidence[0].Locator.StartLine != 11 || overrideRel.Evidence[0].Locator.EndLine != 13 {
		t.Errorf("expected override relationship line span 11..13, got %+v", overrideRel.Evidence[0].Locator)
	}

	// Verify unresolved reference recorded for contract:custom-checkout with YAML line span 14..16
	if len(mm.UnresolvedRefs) != 1 {
		t.Fatalf("expected 1 unresolved reference for contract:custom-checkout, got %d: %+v", len(mm.UnresolvedRefs), mm.UnresolvedRefs)
	}
	if mm.UnresolvedRefs[0].TargetID != "contract:custom-checkout" || mm.UnresolvedRefs[0].Evidence[0].Locator.StartLine != 14 || mm.UnresolvedRefs[0].Evidence[0].Locator.EndLine != 16 {
		t.Errorf("unexpected unresolved reference: %+v", mm.UnresolvedRefs[0])
	}

	// 2. Default store persistence (no --output specified)
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("gira index without --output failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Indexed") {
		t.Errorf("expected Indexed in stdout, got %s", stdout.String())
	}
	publishedPath := filepath.Join(tmpHome, ".cache", "gira", "indexes", "index-test-system", "macro-map.json")
	if _, statErr := os.Stat(publishedPath); statErr != nil {
		t.Errorf("expected default-store index at %s (temp home), got: %v", publishedPath, statErr)
	}

	// 3. Corrupt output file must NOT be overwritten unless --recover-corrupt is passed
	corruptBytes := []byte("{corrupt-existing-index")
	if err := os.WriteFile(outputPath, corruptBytes, 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestPath, "--output", outputPath, "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected gira index to fail when existing output file is corrupt")
	}
	gotCorrupt, _ := os.ReadFile(outputPath)
	if string(gotCorrupt) != string(corruptBytes) {
		t.Fatalf("expected corrupt output file to be preserved on error")
	}
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestPath, "--output", outputPath, "--cache-dir", cacheDir, "--recover-corrupt"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected gira index --recover-corrupt to succeed, got: %v", err)
	}

	// 4. Offline failure when repo not cached
	emptyCacheDir := t.TempDir()
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", manifestPath, "--offline", "--cache-dir", emptyCacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Errorf("expected error in offline mode with missing cache, got nil")
	}

	// 5. Missing manifest file error
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"index", "/nonexistent/path/gira.yaml"}, nil, &stdout, &stderr)
	if err == nil {
		t.Errorf("expected error for missing manifest, got nil")
	}
}

func TestCLI_LoadIndexResolutionAndFailClosed(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	storeRoot := filepath.Join(tempHome, ".cache", "gira", "indexes")
	st := store.NewStore(storeRoot)

	// Populate another system in cache ("other-system")
	otherMap := domain.NewMacroMap("other-system")
	_ = otherMap.AddNode(domain.Node{ID: "service:other", Kind: domain.NodeService, Name: "other"})
	if err := st.Save("other-system", otherMap); err != nil {
		t.Fatalf("failed to save other-system: %v", err)
	}

	// Case 1: Project directory has gira.yaml for "my-project", which is NOT indexed yet.
	// Even though "other-system" exists in cache, loadIndex must NOT fall back to "other-system" or empty map;
	// it must return index_missing for "my-project".
	projDir := t.TempDir()
	manifestYAML := `version: 1
target_system:
  name: my-project
repositories:
  - id: repo
    url: "https://example.com/repo.git"
    sha: "1111222233334444555566667777888899990000"
`
	if err := os.WriteFile(filepath.Join(projDir, "gira.yaml"), []byte(manifestYAML), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(projDir)

	var stdout, stderr bytes.Buffer
	err := Run([]string{"export"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected index_missing error when my-project is not indexed, got nil")
	}
	if !strings.Contains(err.Error(), "index_missing") || !strings.Contains(err.Error(), "my-project") {
		t.Errorf("expected error mentioning index_missing and my-project, got: %v", err)
	}

	// Case 1b: Running from a nested subdirectory of projDir (without its own gira.yaml)
	// must walk up to projDir/gira.yaml and still fail closed with index_missing for "my-project".
	subDir := filepath.Join(projDir, "cmd", "subpkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(subDir)
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"export"}, nil, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "index_missing") || !strings.Contains(err.Error(), "my-project") {
		t.Fatalf("expected subdirectory execution to resolve parent gira.yaml and return index_missing for my-project, got: %v", err)
	}

	// Case 2: Index "my-project" in store -> now export from both projDir and subDir resolves exclusively to "my-project"
	myMap := domain.NewMacroMap("my-project")
	_ = myMap.AddNode(domain.Node{ID: "service:my-svc", Kind: domain.NodeService, Name: "my-svc"})
	if err := st.Save("my-project", myMap); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if err := Run([]string{"export"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("expected export from subdirectory to succeed once my-project is indexed, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "service:my-svc") || strings.Contains(stdout.String(), "other") {
		t.Errorf("expected my-project nodes only from subdirectory, got: %s", stdout.String())
	}

	// Case 3: Directory without gira.yaml in its hierarchy when multiple systems ("my-project" and "other-system") exist in cache -> ambiguity error
	noManifestDir := t.TempDir()
	t.Chdir(noManifestDir)
	stdout.Reset()
	stderr.Reset()
	err = Run([]string{"export"}, nil, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "ambiguous index: multiple systems found in cache") {
		t.Errorf("expected ambiguity error when multiple systems exist in cache without gira.yaml, got: %v", err)
	}
}


func TestGitProbesIgnoreRepositoryRedirectEnvironment(t *testing.T) {
	repoDir := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		cmd.Env = gitCommandEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("ok\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "initial")

	t.Setenv("GIT_DIR", "/definitely/not/a/repository")
	t.Setenv("GIT_WORK_TREE", "/elsewhere")
	t.Setenv("GIT_COMMON_DIR", "/elsewhere/common")
	t.Setenv("GIT_INDEX_FILE", "/tmp/redirected-index")
	t.Setenv("GIT_OBJECT_DIRECTORY", "/tmp/redirected-objects")
	t.Setenv("GIT_ALTERNATE_OBJECT_DIRECTORIES", "/tmp/redirected-alternates")
	t.Setenv("GIT_NAMESPACE", "redirected")
	t.Setenv("GIT_PREFIX", "redirected-prefix/")
	t.Setenv("GIT_CEILING_DIRECTORIES", repoDir)
	t.Setenv("GIT_DISCOVERY_ACROSS_FILESYSTEM", "0")
	t.Setenv("GIT_REPLACE_REF_BASE", "refs/alternate-replace/")
	t.Setenv("GIT_GRAFT_FILE", filepath.Join(repoDir, "missing-grafts"))
	t.Setenv("GIT_SHALLOW_FILE", filepath.Join(repoDir, "missing-shallow"))
	t.Setenv("GIT_CONFIG", filepath.Join(repoDir, "missing-config"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(repoDir, "missing-global-config"))
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(repoDir, "missing-system-config"))
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.bare")
	t.Setenv("GIT_CONFIG_VALUE_0", "true")
	t.Setenv("GIT_CONFIG_PARAMETERS", "'core.bare=true'")
	// Compare filesystem identity, not path spelling: Git may return a physical
	// path on macOS or a slash/long-name path on Windows.
	wantInfo, err := os.Stat(repoDir)
	if err != nil {
		t.Fatalf("stat test repository path: %v", err)
	}
	got := gitToplevel(repoDir)
	gotInfo, err := os.Stat(filepath.FromSlash(got))
	if err != nil {
		t.Fatalf("stat gitToplevel path %q: %v", got, err)
	}
	if !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("gitToplevel() = %q, want a path identifying %q despite redirecting env", got, repoDir)
	}
	if gitDirty(repoDir) {
		t.Fatal("gitDirty() reported clean repository as dirty due to redirected environment")
	}
}


func TestInitDoesNotPublishManifestWhenIndexingFails(t *testing.T) {
	repoDir := t.TempDir()
	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "initial")

	cacheDir := filepath.Join(t.TempDir(), "empty-cache")
	var stdout, stderr bytes.Buffer
	err := Run([]string{"init", repoDir, "--offline", "--cache-dir", cacheDir}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("init unexpectedly succeeded with an empty offline cache")
	}
	if _, statErr := os.Stat(filepath.Join(repoDir, "gira.yaml")); !os.IsNotExist(statErr) {
		t.Fatalf("failed init left a manifest behind; stat error = %v", statErr)
	}
}

func TestInitForceDoesNotRepinManifestWhenIndexingFails(t *testing.T) {
	repoDir := t.TempDir()
	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	runGit("init")
	runGit("config", "user.name", "Test Runner")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "initial")

	manifestPath := filepath.Join(repoDir, "gira.yaml")
	oldSHA := strings.Repeat("0", 40)
	if err := writeLocalManifest(manifestPath, repoDir, oldSHA); err != nil {
		t.Fatalf("write initial manifest: %v", err)
	}
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = Run([]string{"init", repoDir, "--force", "--offline", "--cache-dir", filepath.Join(t.TempDir(), "empty-cache")}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("init --force unexpectedly succeeded with an empty offline cache")
	}
	after, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(after) != string(before) {
		t.Fatalf("failed init --force changed manifest bytes\nbefore: %s\nafter: %s", before, after)
	}
}
