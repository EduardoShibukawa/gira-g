package extractors

import (
	"context"
	"strings"
	"testing"

	"labs/gira-g/internal/domain"
)

func TestMarkdownExtractor_MetadataAndCanExtract(t *testing.T) {
	ext := NewMarkdownExtractor()
	if ext.Name() != "Layer2Markdown" {
		t.Fatalf("unexpected extractor name: %s", ext.Name())
	}
	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("unexpected layer: %v", ext.Layer())
	}

	validCases := []string{
		"docs/decisions/0001-use-go.md",
		"docs/adr/0002-extractors.md",
		"tasks/2026-10-12-task.md",
		"specs/system-spec.md",
		"AGENTS.md",
		"CLAUDE.md",
		"GEMINI.md",
		"CONTEXT.md",
		"SKILL.md",
		"sub/dir/tasks/my-task.markdown",
	}

	for _, path := range validCases {
		if !ext.CanExtract(path) {
			t.Errorf("expected CanExtract(%q) to be true", path)
		}
	}

	invalidCases := []string{
		"main.go",
		"docker-compose.yaml",
		"random/notes.txt",
		"docs/general/overview.html",
	}

	for _, path := range invalidCases {
		if ext.CanExtract(path) {
			t.Errorf("expected CanExtract(%q) to be false", path)
		}
	}
}

func TestMarkdownExtractor_ExtractTask(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `---
schema: task/1
tier: standard
status: active
---

# Feature Delivery Task

Blocked by: [Prerequisite](tasks/prereq-task.md)
implements: compliance-policy

## Verification

| ID | Check | Result |
|---|---|---|
| AC1 | Tests pass | pass |
| AC2 | Race check | pass |
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      "tasks/2026-10-12-feature.md",
		Content:   []byte(content),
	}

	nodes, rels, unsupporteds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(unsupporteds) != 0 {
		t.Fatalf("expected 0 unsupporteds, got %d", len(unsupporteds))
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 task node, got %d", len(nodes))
	}

	task := nodes[0]
	if task.ID != "task:2026-10-12-feature" {
		t.Errorf("expected node ID 'task:2026-10-12-feature', got %s", task.ID)
	}
	if task.Kind != domain.NodeTask {
		t.Errorf("expected NodeTask, got %v", task.Kind)
	}
	if task.Name != "Feature Delivery Task" {
		t.Errorf("expected 'Feature Delivery Task', got %s", task.Name)
	}

	if len(task.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record, got %d", len(task.Evidence))
	}
	ev := task.Evidence[0]
	if ev.DeploymentProfile != "task" {
		t.Errorf("expected deployment profile 'task', got %s", ev.DeploymentProfile)
	}
	if ev.Locator.StartLine != 1 || ev.Locator.EndLine < 15 {
		t.Errorf("unexpected locator range: %d-%d", ev.Locator.StartLine, ev.Locator.EndLine)
	}

	// Verify relationships: 1 governs (prereq -> task), 1 implements (task -> policy), 2 verifies (task -> checks)
	if len(rels) != 4 {
		t.Fatalf("expected 4 relationships, got %d", len(rels))
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = r
	}

	if r, ok := relMap["task:prereq-task->governs->task:2026-10-12-feature"]; !ok {
		t.Errorf("missing prereq-task governs feature task relation")
	} else {
		if r.Evidence[0].Locator.StartLine <= 0 {
			t.Errorf("expected positive line locator")
		}
	}

	if _, ok := relMap["task:2026-10-12-feature->implements->policy:compliance-policy"]; !ok {
		t.Errorf("missing implements compliance-policy relation")
	}

	if _, ok := relMap["task:2026-10-12-feature->verifies->contract:check:ac1"]; !ok {
		t.Errorf("missing verifies contract:check:ac1 relation")
	}

	if _, ok := relMap["task:2026-10-12-feature->verifies->contract:check:ac2"]; !ok {
		t.Errorf("missing verifies contract:check:ac2 relation")
	}
}

func TestMarkdownExtractor_ExtractSkill(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `---
name: go-architect
description: Expert Go software architect and designer
---

# Go Architect Skill

Provides architecture guidelines for Go services.
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "sha1",
		Path:      "skills/go-architect/SKILL.md",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 skill node, got %d", len(nodes))
	}

	skill := nodes[0]
	if skill.ID != "skill:go-architect" {
		t.Errorf("expected ID 'skill:go-architect', got %s", skill.ID)
	}
	if skill.Kind != domain.NodeSkill {
		t.Errorf("expected NodeSkill, got %v", skill.Kind)
	}
	if skill.Name != "go-architect" {
		t.Errorf("expected 'go-architect', got %s", skill.Name)
	}
	if skill.Evidence[0].DeploymentProfile != "skill" {
		t.Errorf("expected deployment profile 'skill', got %s", skill.Evidence[0].DeploymentProfile)
	}
	if len(rels) != 0 {
		t.Errorf("expected 0 relationships, got %d", len(rels))
	}
}

func TestMarkdownExtractor_ExtractPolicyAndADR(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `---
schema: dec/1
---

# ADR 0005: Microservice Boundaries

Governs: order-service, inventory-service, payment-service

Details on service partitioning.
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "sha2",
		Path:      "docs/decisions/0005-microservices.md",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 policy node, got %d", len(nodes))
	}

	policy := nodes[0]
	if policy.ID != "policy:0005-microservices" {
		t.Errorf("expected ID 'policy:0005-microservices', got %s", policy.ID)
	}
	if policy.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy, got %v", policy.Kind)
	}
	if !strings.Contains(policy.Name, "ADR 0005") {
		t.Errorf("expected name with 'ADR 0005', got %s", policy.Name)
	}
	if policy.Evidence[0].DeploymentProfile != "markdown" {
		t.Errorf("expected deployment profile 'markdown', got %s", policy.Evidence[0].DeploymentProfile)
	}

	if len(rels) != 3 {
		t.Fatalf("expected 3 governs relationships, got %d", len(rels))
	}

	for _, r := range rels {
		if r.Verb != domain.VerbGoverns {
			t.Errorf("expected VerbGoverns, got %s", r.Verb)
		}
		if r.SourceID != "policy:0005-microservices" {
			t.Errorf("expected source policy:0005-microservices, got %s", r.SourceID)
		}
		if r.Evidence[0].Locator.StartLine <= 0 {
			t.Errorf("expected valid line locator")
		}
	}
}

func TestMarkdownExtractor_EdgeCases(t *testing.T) {
	ext := NewMarkdownExtractor()

	// Empty content
	fileEmpty := TargetFile{
		Repo:      "repo",
		CommitSHA: "c1",
		Path:      "docs/adr/0001.md",
		Content:   []byte(""),
	}
	nodes, rels, _, err := ext.Extract(context.Background(), fileEmpty)
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node even for empty file with valid ADR path")
	}
	if len(rels) != 0 {
		t.Fatalf("expected 0 relationships")
	}

	// File with no frontmatter, just headers
	filePlain := TargetFile{
		Repo:      "repo",
		CommitSHA: "c1",
		Path:      "tasks/plain-task.md",
		Content:   []byte("# Plain Task\n\nNo frontmatter here.\n"),
	}
	nodes, _, _, err = ext.Extract(context.Background(), filePlain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Kind != domain.NodeTask || nodes[0].Name != "Plain Task" {
		t.Fatalf("expected task node with Plain Task, got %+v", nodes)
	}
}
