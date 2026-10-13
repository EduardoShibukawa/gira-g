package extractors

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestMiseExtractor_MetadataAndCanExtract(t *testing.T) {
	ext := NewMiseExtractor()

	if ext.Name() != "Layer2Mise" {
		t.Fatalf("expected name 'Layer2Mise', got '%s'", ext.Name())
	}

	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("expected layer %d, got %d", LayerToolMeaning, ext.Layer())
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{"mise.toml", true},
		{".mise.toml", true},
		{"sub/dir/mise.toml", true},
		{"mise/config.toml", true},
		{".mise/config.toml", true},
		{"project/mise/config.toml", true},
		{"other.toml", false},
		{"config.toml", false},
		{"mise.yaml", false},
	}

	for _, tc := range testCases {
		res := ext.CanExtract(tc.path)
		if res != tc.expected {
			t.Errorf("CanExtract(%q) = %v, expected %v", tc.path, res, tc.expected)
		}
	}
}

func TestMiseExtractor_ExtractToolsAndTasks(t *testing.T) {
	ext := NewMiseExtractor()

	miseContent := `[tools]
go = "1.24.0"
bun = "1.2.4"
golangci-lint = "1.64.5"
cosign = "3.1.3"

[env]
NODE_ENV = "production"
GO111MODULE = "on"

[tasks.check]
description = "Execute complete quality checks (Go tests, linting)"
run = "go test -race ./..."

[tasks.lint]
description = "Run static analysis with golangci-lint"
run = "golangci-lint run"

[tasks.build]
description = "Compile the standalone binary"
run = "go build -o bin/app"
depends = ["check"]
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "sha456",
		Path:      "mise.toml",
		Content:   []byte(miseContent),
	}

	nodes, rels, bounds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	if len(bounds) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(bounds))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// 1. Verify Library tool nodes
	goTool, ok := nodeMap["library:go"]
	if !ok {
		t.Fatalf("missing library:go node")
	}
	if goTool.Kind != domain.NodeLibrary {
		t.Errorf("expected go tool kind Library, got %s", goTool.Kind)
	}
	if goTool.Metadata["version"] != "1.24.0" {
		t.Errorf("expected go version '1.24.0', got %v", goTool.Metadata["version"])
	}
	if len(goTool.Evidence) == 0 {
		t.Fatalf("missing evidence for go tool")
	}
	if goTool.Evidence[0].DeploymentProfile != "mise" {
		t.Errorf("expected deployment_profile 'mise', got %s", goTool.Evidence[0].DeploymentProfile)
	}
	if goTool.Evidence[0].Locator.StartLine <= 0 {
		t.Errorf("invalid locator start line: %d", goTool.Evidence[0].Locator.StartLine)
	}

	bunTool, ok := nodeMap["library:bun"]
	if !ok {
		t.Fatalf("missing library:bun node")
	}
	if bunTool.Metadata["version"] != "1.2.4" {
		t.Errorf("expected bun version '1.2.4', got %v", bunTool.Metadata["version"])
	}

	// 2. Verify Infrastructure tool node (cosign)
	cosignTool, ok := nodeMap["library:cosign"]
	if !ok {
		t.Fatalf("missing library:cosign node")
	}
	if cosignTool.Kind != domain.NodeInfrastructure {
		t.Errorf("expected cosign tool kind Infrastructure, got %s", cosignTool.Kind)
	}

	// 3. Verify Task nodes (Service)
	checkTask, ok := nodeMap["service:task:check"]
	if !ok {
		t.Fatalf("missing service:task:check node")
	}
	if checkTask.Kind != domain.NodeService {
		t.Errorf("expected check task kind Service, got %s", checkTask.Kind)
	}
	if checkTask.Name != "check" {
		t.Errorf("expected check task name 'check', got %s", checkTask.Name)
	}
	if checkTask.Description != "Execute complete quality checks (Go tests, linting)" {
		t.Errorf("expected correct description, got %s", checkTask.Description)
	}
	if checkTask.Metadata["run"] != "go test -race ./..." {
		t.Errorf("expected correct run command, got %v", checkTask.Metadata["run"])
	}

	lintTask, ok := nodeMap["service:task:lint"]
	if !ok {
		t.Fatalf("missing service:task:lint node")
	}
	if lintTask.Metadata["run"] != "golangci-lint run" {
		t.Errorf("expected correct run command, got %v", lintTask.Metadata["run"])
	}

	buildTask, ok := nodeMap["service:task:build"]
	if !ok {
		t.Fatalf("missing service:task:build node")
	}
	deps, ok := buildTask.Metadata["depends"].([]string)
	if !ok || len(deps) != 1 || deps[0] != "check" {
		t.Errorf("expected build task depends ['check'], got %v", buildTask.Metadata["depends"])
	}

	// 4. Verify Relationships
	var foundTaskGoRel bool
	var foundTaskLintRel bool
	var foundBuildCheckRel bool

	for _, r := range rels {
		if r.SourceID == "service:task:check" && r.TargetID == "library:go" && r.Verb == domain.VerbConsumes {
			foundTaskGoRel = true
			if len(r.Evidence) == 0 || r.Evidence[0].DeploymentProfile != "mise" {
				t.Errorf("invalid evidence in task->go rel: %+v", r.Evidence)
			}
		}
		if r.SourceID == "service:task:lint" && r.TargetID == "library:golangci-lint" && r.Verb == domain.VerbConsumes {
			foundTaskLintRel = true
		}
		if r.SourceID == "service:task:build" && r.TargetID == "service:task:check" && r.Verb == domain.VerbConsumes {
			foundBuildCheckRel = true
		}
	}

	if !foundTaskGoRel {
		t.Errorf("relationship check task -> go tool not found")
	}
	if !foundTaskLintRel {
		t.Errorf("relationship lint task -> golangci-lint tool not found")
	}
	if !foundBuildCheckRel {
		t.Errorf("relationship build task -> check task not found")
	}
}

func TestMiseExtractor_EdgeCases(t *testing.T) {
	ext := NewMiseExtractor()

	// Empty file
	fileEmpty := TargetFile{
		Repo:    "duh-homelab",
		Path:    "mise.toml",
		Content: []byte(""),
	}
	nodes, rels, bounds, err := ext.Extract(context.Background(), fileEmpty)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("expected empty result on empty file, got %v, nodes: %v", err, nodes)
	}

	// Only comments and blank lines
	fileComments := TargetFile{
		Repo:    "duh-homelab",
		Path:    "mise.toml",
		Content: []byte("# Mise toolchain\n\n# Another comment\n"),
	}
	nodes, rels, bounds, err = ext.Extract(context.Background(), fileComments)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("expected empty result on comment-only file, got %v, nodes: %v", err, nodes)
	}
}
