package extractors

import (
	"context"
	"os"
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
		".scratch/system-context-harness/issues/01-first-user-and-workflow.md",
		".scratch/system-context-harness/map.md",
		".scratch/system-context-harness/brief.md",
		"CONTEXT-MAP.md",
		"context-map.md",
		"README.md",
		"readme.md",
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

func TestMarkdownExtractor_ExtractScratchIssue_Grilling(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `# Which user and first workflow set the MVP's acceptance criteria?

Type: grilling
Status: resolved
Blocked by: —
Map: [System Context Harness](../map.md)


## Question

The harness serves humans and AI agents across many scenarios.
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      ".scratch/system-context-harness/issues/01-first-user-and-workflow.md",
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
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != "task:system-context-harness:01-first-user-and-workflow" {
		t.Errorf("expected node ID 'task:system-context-harness:01-first-user-and-workflow', got %s", node.ID)
	}
	if node.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy for grilling ticket, got %v", node.Kind)
	}
	if node.Name != "Which user and first workflow set the MVP's acceptance criteria?" {
		t.Errorf("unexpected name: %s", node.Name)
	}
	if node.Metadata["type"] != "grilling" {
		t.Errorf("expected type 'grilling', got %v", node.Metadata["type"])
	}
	if node.Metadata["status"] != "resolved" {
		t.Errorf("expected status 'resolved', got %v", node.Metadata["status"])
	}
	if node.Metadata["effort"] != "system-context-harness" {
		t.Errorf("expected effort 'system-context-harness', got %v", node.Metadata["effort"])
	}
	if len(node.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record, got %d", len(node.Evidence))
	}
	if node.Evidence[0].Locator.StartLine != 1 || node.Evidence[0].Locator.EndLine < 10 {
		t.Errorf("unexpected locator range: %d-%d", node.Evidence[0].Locator.StartLine, node.Evidence[0].Locator.EndLine)
	}

	// Relationships: Map governs issue
	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship from Map, got %d", len(rels))
	}
	rel := rels[0]
	if rel.SourceID != "policy:system-context-harness:map" {
		t.Errorf("expected source 'policy:system-context-harness:map', got %s", rel.SourceID)
	}
	if rel.TargetID != "task:system-context-harness:01-first-user-and-workflow" {
		t.Errorf("expected target 'task:system-context-harness:01-first-user-and-workflow', got %s", rel.TargetID)
	}
	if rel.Verb != domain.VerbGoverns {
		t.Errorf("expected VerbGoverns, got %s", rel.Verb)
	}
	if len(rel.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record for rel, got %d", len(rel.Evidence))
	}
	if rel.Evidence[0].Locator.StartLine != 6 || rel.Evidence[0].Locator.EndLine != 6 {
		t.Errorf("expected locator line 6, got %d-%d", rel.Evidence[0].Locator.StartLine, rel.Evidence[0].Locator.EndLine)
	}
}

func TestMarkdownExtractor_ExtractScratchIssue_TaskWithBlocker(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `# Prototype Architecture Slice

Type: prototype
Status: in-progress
Blocked by: 01-first-user-and-workflow, [Prerequisite Task](tasks/prereq.md)
Map: [System Context Harness](../map.md)

Detailed implementation plan.
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      ".scratch/system-context-harness/issues/15-context-slice.md",
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
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != "task:system-context-harness:15-context-slice" {
		t.Errorf("expected node ID 'task:system-context-harness:15-context-slice', got %s", node.ID)
	}
	if node.Kind != domain.NodeTask {
		t.Errorf("expected NodeTask for prototype ticket, got %v", node.Kind)
	}
	if node.Metadata["type"] != "prototype" {
		t.Errorf("expected type 'prototype', got %v", node.Metadata["type"])
	}

	// Relationships: 2 blockers + 1 map = 3 VerbGoverns
	if len(rels) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(rels))
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = r
	}

	if _, ok := relMap["task:system-context-harness:01-first-user-and-workflow->governs->task:system-context-harness:15-context-slice"]; !ok {
		t.Errorf("missing blocker 01 governs 15 relation")
	}
	if _, ok := relMap["task:prereq->governs->task:system-context-harness:15-context-slice"]; !ok {
		t.Errorf("missing prereq governs 15 relation")
	}
	if _, ok := relMap["policy:system-context-harness:map->governs->task:system-context-harness:15-context-slice"]; !ok {
		t.Errorf("missing map governs 15 relation")
	}
}

func TestMarkdownExtractor_ExtractScratchMap(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `# Map: System Context Harness

Label: wayfinder:map

## Destination

Harness MVP.

## Notes

- **Original brief:** [brief.md](brief.md).

## Decisions so far

- [Which user and first workflow set the MVP's acceptance criteria?](issues/01-first-user-and-workflow.md)
- [Trust boundary: what may leave the machine in the MVP?](issues/02-trust-boundary.md) — ... ([ADR 0001](../../labs/gira-g/docs/adr/0001-index-stores-locators-no-model-client.md))
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      ".scratch/system-context-harness/map.md",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 map node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != "policy:system-context-harness:map" {
		t.Errorf("expected node ID 'policy:system-context-harness:map', got %s", node.ID)
	}
	if node.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy, got %v", node.Kind)
	}
	if node.Name != "Map: System Context Harness" {
		t.Errorf("unexpected name: %s", node.Name)
	}

	if len(rels) < 4 {
		t.Fatalf("expected at least 4 relationships, got %d", len(rels))
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = r
	}

	if _, ok := relMap["policy:system-context-harness:map->governs->policy:system-context-harness:brief"]; !ok {
		t.Errorf("missing map governs brief relation")
	}
	if _, ok := relMap["policy:system-context-harness:map->governs->task:system-context-harness:01-first-user-and-workflow"]; !ok {
		t.Errorf("missing map governs issue 01 relation")
	}
	if _, ok := relMap["policy:system-context-harness:map->governs->task:system-context-harness:02-trust-boundary"]; !ok {
		t.Errorf("missing map governs issue 02 relation")
	}
	if _, ok := relMap["policy:system-context-harness:map->governs->policy:0001-index-stores-locators-no-model-client"]; !ok {
		t.Errorf("missing map governs ADR 0001 relation")
	}
}

func TestMarkdownExtractor_ExtractContextMap(t *testing.T) {
	ext := NewMarkdownExtractor()

	content := `# Context Map

Maps domain contexts and initiatives within duh-homelab.

## Initiatives / Labs

### Gira-G
- **Project:** Gira-G (CLI: gira)
- **Location:** labs/gira-g/
- **Domain Context:** [CONTEXT.md](labs/gira-g/CONTEXT.md)
- **Decisions:** [docs/adr/](labs/gira-g/docs/adr/)
- **Effort Map & Issues:** [.scratch/system-context-harness/map.md](.scratch/system-context-harness/map.md)
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      "CONTEXT-MAP.md",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != "policy:context-map" {
		t.Errorf("expected node ID 'policy:context-map', got %s", node.ID)
	}
	if node.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy, got %v", node.Kind)
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = r
	}

	if _, ok := relMap["policy:context-map->governs->policy:context"]; !ok {
		t.Errorf("missing context-map governs context relation")
	}
	if _, ok := relMap["policy:context-map->governs->policy:system-context-harness:map"]; !ok {
		t.Errorf("missing context-map governs map relation")
	}
}

func TestMarkdownExtractor_ExtractRealScratchIssue(t *testing.T) {
	content, err := os.ReadFile("../../../../.scratch/system-context-harness/issues/01-first-user-and-workflow.md")
	if err != nil {
		t.Fatalf("failed to read real repo issue: %v", err)
	}

	ext := NewMarkdownExtractor()
	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "head",
		Path:      ".scratch/system-context-harness/issues/01-first-user-and-workflow.md",
		Content:   content,
	}

	nodes, rels, unsupporteds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(unsupporteds) != 0 {
		t.Fatalf("expected 0 unsupporteds, got %d", len(unsupporteds))
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	node := nodes[0]
	if node.ID != "task:system-context-harness:01-first-user-and-workflow" {
		t.Errorf("unexpected node ID: %s", node.ID)
	}
	if node.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy, got %v", node.Kind)
	}
	if len(rels) != 1 {
		t.Fatalf("expected 1 rel (from Map:), got %d", len(rels))
	}
	if rels[0].SourceID != "policy:system-context-harness:map" {
		t.Errorf("unexpected rel source: %s", rels[0].SourceID)
	}
	if rels[0].TargetID != "task:system-context-harness:01-first-user-and-workflow" {
		t.Errorf("unexpected rel target: %s", rels[0].TargetID)
	}
	if rels[0].Verb != domain.VerbGoverns {
		t.Errorf("expected VerbGoverns, got %s", rels[0].Verb)
	}
	if rels[0].Evidence[0].Locator.StartLine != 6 {
		t.Errorf("expected locator line 6, got %d", rels[0].Evidence[0].Locator.StartLine)
	}
}
