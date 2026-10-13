package extractors

import (
	"context"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestGitHubActionsExtractor_MetadataAndCanExtract(t *testing.T) {
	ext := NewGitHubActionsExtractor()

	if ext.Name() != "Layer2GitHubActions" {
		t.Fatalf("expected name 'Layer2GitHubActions', got '%s'", ext.Name())
	}

	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("expected layer %d, got %d", LayerToolMeaning, ext.Layer())
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{".github/workflows/ci.yml", true},
		{".github/workflows/build.yaml", true},
		{".github/workflows/deploy.YML", true},
		{"subproject/.github/workflows/pipeline.yaml", true},
		{".github/workflows/ci.json", false},
		{".github/workflows/README.md", false},
		{".github/ci.yml", false},
		{"workflows/ci.yml", false},
		{"ci.yaml", false},
	}

	for _, tc := range testCases {
		res := ext.CanExtract(tc.path)
		if res != tc.expected {
			t.Errorf("CanExtract(%q) = %v, expected %v", tc.path, res, tc.expected)
		}
	}
}

func TestGitHubActionsExtractor_ExtractJobsAndActions(t *testing.T) {
	ext := NewGitHubActionsExtractor()

	workflowContent := `name: CI Pipeline

on:
  push:
    branches: [main]

jobs:
  test-unit:
    name: Run Unit Tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run Tests
        run: go test -v ./...

  deploy-cloud:
    name: Deploy to Production
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
      - name: Terraform Deploy
        uses: hashicorp/setup-terraform@v3
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      ".github/workflows/ci.yml",
		Content:   []byte(workflowContent),
	}

	nodes, rels, boundaries, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	if len(boundaries) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(boundaries))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// 1. Verify test-unit job (Service)
	testJob, ok := nodeMap["service:job:test-unit"]
	if !ok {
		t.Fatalf("missing service:job:test-unit node")
	}
	if testJob.Kind != domain.NodeService {
		t.Errorf("expected test-unit kind Service, got %s", testJob.Kind)
	}
	if testJob.Name != "Run Unit Tests" {
		t.Errorf("expected test-unit name 'Run Unit Tests', got %s", testJob.Name)
	}
	if len(testJob.Evidence) == 0 {
		t.Fatalf("missing evidence for test-unit")
	}
	ev := testJob.Evidence[0]
	if ev.DeploymentProfile != "github-actions" {
		t.Errorf("expected deployment_profile 'github-actions', got %s", ev.DeploymentProfile)
	}
	if ev.Locator.StartLine <= 0 || ev.Locator.EndLine < ev.Locator.StartLine {
		t.Errorf("invalid line locators: %d - %d", ev.Locator.StartLine, ev.Locator.EndLine)
	}

	// 2. Verify deploy-cloud job (Infrastructure)
	deployJob, ok := nodeMap["service:job:deploy-cloud"]
	if !ok {
		t.Fatalf("missing service:job:deploy-cloud node")
	}
	if deployJob.Kind != domain.NodeInfrastructure {
		t.Errorf("expected deploy-cloud kind Infrastructure, got %s", deployJob.Kind)
	}

	// 3. Verify actions (Library)
	checkoutAct, ok := nodeMap["library:actions/checkout"]
	if !ok {
		t.Fatalf("missing library:actions/checkout node")
	}
	if checkoutAct.Kind != domain.NodeLibrary {
		t.Errorf("expected actions/checkout kind Library, got %s", checkoutAct.Kind)
	}
	if checkoutAct.Metadata["version"] != "v4" {
		t.Errorf("expected checkout version 'v4', got %v", checkoutAct.Metadata["version"])
	}

	setupGoAct, ok := nodeMap["library:actions/setup-go"]
	if !ok {
		t.Fatalf("missing library:actions/setup-go node")
	}
	if setupGoAct.Metadata["version"] != "v5" {
		t.Errorf("expected setup-go version 'v5', got %v", setupGoAct.Metadata["version"])
	}

	terraformAct, ok := nodeMap["library:hashicorp/setup-terraform"]
	if !ok {
		t.Fatalf("missing library:hashicorp/setup-terraform node")
	}
	if terraformAct.Metadata["version"] != "v3" {
		t.Errorf("expected terraform version 'v3', got %v", terraformAct.Metadata["version"])
	}

	// 4. Verify relationships
	if len(rels) < 3 {
		t.Fatalf("expected at least 3 relationships, got %d", len(rels))
	}

	var foundTestCheckout bool
	var foundDeployTerraform bool
	for _, r := range rels {
		if r.SourceID == "service:job:test-unit" && r.TargetID == "library:actions/checkout" && r.Verb == domain.VerbConsumes {
			foundTestCheckout = true
			if len(r.Evidence) == 0 || r.Evidence[0].DeploymentProfile != "github-actions" {
				t.Errorf("invalid evidence in test-checkout relationship: %+v", r.Evidence)
			}
			if r.Evidence[0].Locator.StartLine <= 0 {
				t.Errorf("invalid line number in relationship locator: %d", r.Evidence[0].Locator.StartLine)
			}
		}
		if r.SourceID == "service:job:deploy-cloud" && r.TargetID == "library:hashicorp/setup-terraform" && r.Verb == domain.VerbConsumes {
			foundDeployTerraform = true
		}
	}

	if !foundTestCheckout {
		t.Errorf("relationship test-unit -> actions/checkout not found")
	}
	if !foundDeployTerraform {
		t.Errorf("relationship deploy-cloud -> hashicorp/setup-terraform not found")
	}
}

func TestGitHubActionsExtractor_EdgeCases(t *testing.T) {
	ext := NewGitHubActionsExtractor()

	// Empty content
	fileEmpty := TargetFile{
		Repo:    "duh-homelab",
		Path:    ".github/workflows/empty.yml",
		Content: []byte(""),
	}
	nodes, rels, bounds, err := ext.Extract(context.Background(), fileEmpty)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("expected empty result on empty file, got err: %v, nodes: %v", err, nodes)
	}

	// Non-mapping YAML
	fileList := TargetFile{
		Repo:    "duh-homelab",
		Path:    ".github/workflows/list.yml",
		Content: []byte("- item1\n- item2"),
	}
	nodes, rels, bounds, err = ext.Extract(context.Background(), fileList)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("expected empty result on list YAML, got err: %v, nodes: %v", err, nodes)
	}

	// Malformed YAML
	fileMalformed := TargetFile{
		Repo:    "duh-homelab",
		Path:    ".github/workflows/bad.yml",
		Content: []byte("foo: [unclosed"),
	}
	_, _, _, err = ext.Extract(context.Background(), fileMalformed)
	if err == nil || !strings.Contains(err.Error(), "failed to parse GitHub Actions workflow") {
		t.Fatalf("expected parse error on malformed YAML, got %v", err)
	}
}
