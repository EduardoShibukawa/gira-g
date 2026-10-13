package extractors

import (
	"context"
	"testing"

	"labs/gira-g/internal/domain"
)

func TestGoModExtractor_Basic(t *testing.T) {
	ext := NewGoModExtractor()

	if ext.Name() != "Layer2GoMod" {
		t.Fatalf("expected Name Layer2GoMod, got %s", ext.Name())
	}
	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("expected Layer LayerToolMeaning, got %v", ext.Layer())
	}

	if !ext.CanExtract("go.mod") {
		t.Fatalf("expected CanExtract(go.mod) == true")
	}
	if !ext.CanExtract("path/to/go.mod") {
		t.Fatalf("expected CanExtract(path/to/go.mod) == true")
	}
	if ext.CanExtract("go.sum") || ext.CanExtract("main.go") || ext.CanExtract("go.mod.backup") {
		t.Fatalf("expected CanExtract to return false for non-go.mod files")
	}
}

func TestGoModExtractor_Extraction(t *testing.T) {
	ext := NewGoModExtractor()

	goModContent := `module github.com/example/sample-app

go 1.22.5

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sync v0.8.0
)
`

	file := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "abc1234",
		Path:      "go.mod",
		Content:   []byte(goModContent),
	}

	nodes, rels, unsupporteds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(rels) != 0 {
		t.Fatalf("expected 0 relationships, got %d", len(rels))
	}
	if len(unsupporteds) != 0 {
		t.Fatalf("expected 0 unsupported boundaries, got %d", len(unsupporteds))
	}

	// Direct dependencies expected: logrus, uuid, x/sync
	// test-repo should NOT include testify (since it's // indirect)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 direct library nodes, got %d", len(nodes))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// 1. Check logrus (single-line require)
	logrus, ok := nodeMap["library:github.com/sirupsen/logrus"]
	if !ok {
		t.Fatalf("missing node for logrus")
	}
	if logrus.Kind != domain.NodeLibrary {
		t.Errorf("expected Kind NodeLibrary, got %v", logrus.Kind)
	}
	if logrus.Name != "github.com/sirupsen/logrus" {
		t.Errorf("expected Name github.com/sirupsen/logrus, got %s", logrus.Name)
	}
	if logrus.Metadata["version"] != "v1.9.3" {
		t.Errorf("expected version v1.9.3, got %v", logrus.Metadata["version"])
	}
	if logrus.Metadata["go_version"] != "1.22.5" {
		t.Errorf("expected go_version 1.22.5, got %v", logrus.Metadata["go_version"])
	}
	if logrus.Metadata["module"] != "github.com/example/sample-app" {
		t.Errorf("expected module github.com/example/sample-app, got %v", logrus.Metadata["module"])
	}
	if len(logrus.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record for logrus, got %d", len(logrus.Evidence))
	}
	ev := logrus.Evidence[0]
	if ev.Locator.StartLine != 5 || ev.Locator.EndLine != 5 {
		t.Errorf("expected line 5 for logrus, got start=%d, end=%d", ev.Locator.StartLine, ev.Locator.EndLine)
	}
	if ev.Origin != domain.OriginExtracted {
		t.Errorf("expected OriginExtracted, got %s", ev.Origin)
	}
	if ev.DeploymentProfile != "gomod" {
		t.Errorf("expected DeploymentProfile gomod, got %s", ev.DeploymentProfile)
	}
	if ev.Status != domain.StatusCurrent {
		t.Errorf("expected StatusCurrent, got %s", ev.Status)
	}

	// 2. Check uuid (block require)
	uuid, ok := nodeMap["library:github.com/google/uuid"]
	if !ok {
		t.Fatalf("missing node for uuid")
	}
	if uuid.Metadata["version"] != "v1.6.0" {
		t.Errorf("expected version v1.6.0, got %v", uuid.Metadata["version"])
	}
	if len(uuid.Evidence) != 1 || uuid.Evidence[0].Locator.StartLine != 8 {
		t.Errorf("expected line 8 for uuid, got %v", uuid.Evidence)
	}

	// 3. Check x/sync (block require)
	syncNode, ok := nodeMap["library:golang.org/x/sync"]
	if !ok {
		t.Fatalf("missing node for x/sync")
	}
	if syncNode.Metadata["version"] != "v0.8.0" {
		t.Errorf("expected version v0.8.0, got %v", syncNode.Metadata["version"])
	}
	if len(syncNode.Evidence) != 1 || syncNode.Evidence[0].Locator.StartLine != 10 {
		t.Errorf("expected line 10 for sync, got %v", syncNode.Evidence)
	}

	// 4. Ensure indirect testify is not present
	if _, found := nodeMap["library:github.com/stretchr/testify"]; found {
		t.Fatalf("indirect dependency testify should have been excluded")
	}
}

func TestGoModExtractor_IndirectSingleLine(t *testing.T) {
	ext := NewGoModExtractor()

	goModContent := `module sample

go 1.24

require gopkg.in/yaml.v3 v3.0.1 // indirect
`
	file := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "abc",
		Path:      "go.mod",
		Content:   []byte(goModContent),
	}

	nodes, _, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes for indirect single line require, got %d", len(nodes))
	}
}
