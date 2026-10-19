package extractors

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func TestDataFormatExtractor(t *testing.T) {
	ext := NewDataFormatExtractor()
	if ext.Name() != "Layer1DataFormat" {
		t.Fatalf("unexpected name: %s", ext.Name())
	}
	if ext.Layer() != LayerDataFormat {
		t.Fatalf("unexpected layer: %v", ext.Layer())
	}

	if !ext.CanExtract("config.json") || !ext.CanExtract("data.yaml") || !ext.CanExtract("app.yml") {
		t.Fatalf("CanExtract failed for valid data formats")
	}
	if ext.CanExtract("code.go") || ext.CanExtract("service.py") {
		t.Fatalf("CanExtract returned true for non-data formats")
	}

	// Valid JSON
	validJSON := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "test.json",
		Content:   []byte(`{"key": "value", "list": [1, 2, 3]}`),
	}
	nodes, rels, bounds, err := ext.Extract(context.Background(), validJSON)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("unexpected result on valid JSON: %v, nodes: %v", err, nodes)
	}

	// Invalid JSON
	invalidJSON := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "bad.json",
		Content:   []byte(`{malformed json`),
	}
	_, _, _, err = ext.Extract(context.Background(), invalidJSON)
	if err == nil || !strings.Contains(err.Error(), "malformed JSON") {
		t.Fatalf("expected malformed JSON error, got %v", err)
	}

	// Valid YAML
	validYAML := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "test.yaml",
		Content:   []byte("key: value\nitems:\n  - a\n  - b\n"),
	}
	nodes, rels, bounds, err = ext.Extract(context.Background(), validYAML)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("unexpected result on valid YAML: %v, nodes: %v", err, nodes)
	}

	// Invalid YAML
	invalidYAML := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "bad.yaml",
		Content:   []byte("key: [unterminated"),
	}
	_, _, _, err = ext.Extract(context.Background(), invalidYAML)
	if err == nil || !strings.Contains(err.Error(), "malformed YAML") {
		t.Fatalf("expected malformed YAML error, got %v", err)
	}
}

func TestPipeline_ProcessFiles(t *testing.T) {
	pipe := DefaultPipeline()

	composeContent := `
version: '3.8'
services:
  frontend:
    image: otel/demo-frontend:v1.0
    depends_on:
      - cartservice
  cartservice:
    image: otel/demo-cart:v1.0
`

	chartContent := `
apiVersion: v2
name: opentelemetry-demo
version: 0.1.0
dependencies:
  - name: kafka
    version: 1.0.0
`

	jsonContent := `{"version": "1.0.0"}`

	files := []TargetFile{
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "docker-compose.yaml",
			Content:   []byte(composeContent),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "Chart.yaml",
			Content:   []byte(chartContent),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "package.json",
			Content:   []byte(jsonContent),
		},
		// Unsupported non-Go files:
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "src/payment/index.js",
			Content:   []byte("console.log('payment');"),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "src/shipping/ShippingService.java",
			Content:   []byte("public class ShippingService {}"),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha123",
			Path:      "src/quote/main.rs",
			Content:   []byte("fn main() {}"),
		},
	}

	mm, err := pipe.ProcessFiles(context.Background(), "otel-demo", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	if mm.TargetSystem != "otel-demo" {
		t.Fatalf("unexpected target system: %s", mm.TargetSystem)
	}

	// Verify nodes
	// docker-compose generated: frontend, cartservice
	// Chart.yaml generated: opentelemetry-demo, kafka
	if len(mm.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(mm.Nodes))
	}

	// Verify relationships:
	// frontend -> cartservice is depends_on-only (start ordering, J2) and
	// must not reach the map; opentelemetry-demo -> kafka (1) remains.
	if len(mm.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(mm.Relationships))
	}
	for _, rel := range mm.Relationships {
		if rel.SourceID == "service:frontend" && rel.TargetID == "service:cartservice" {
			t.Fatalf("ordering-only frontend->cartservice must be pruned (J2)")
		}
	}

	// Verify UnsupportedBoundaries
	// .js, .java, .rs should all be recorded as unsupported
	if len(mm.Unsupported) != 3 {
		t.Fatalf("expected 3 unsupported boundaries, got %d", len(mm.Unsupported))
	}

	unsupportedMap := make(map[string]domain.UnsupportedBoundary)
	for _, u := range mm.Unsupported {
		unsupportedMap[u.Path] = u
	}

	jsUb, ok := unsupportedMap["src/payment/index.js"]
	if !ok {
		t.Fatalf("missing unsupported record for index.js")
	}
	if jsUb.Reason != "no extractor registered for file extension" {
		t.Fatalf("unexpected reason: %s", jsUb.Reason)
	}
	if jsUb.Locator.Repo != "otel-demo" || jsUb.Locator.CommitSHA != "sha123" {
		t.Fatalf("unexpected locator: %+v", jsUb.Locator)
	}

	if _, ok := unsupportedMap["src/shipping/ShippingService.java"]; !ok {
		t.Fatalf("missing unsupported record for ShippingService.java")
	}
	if _, ok := unsupportedMap["src/quote/main.rs"]; !ok {
		t.Fatalf("missing unsupported record for main.rs")
	}

	// Verify CanonicalJSON deterministic serialization
	json1, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	json2, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if string(json1) != string(json2) {
		t.Fatalf("canonical JSON is not deterministic")
	}
}

func TestPipeline_NodeDeduplicationAndEvidenceAggregation(t *testing.T) {
	pipe := DefaultPipeline()

	// Same service declared in compose and in values.yaml
	composeContent := `
services:
  cartservice:
    image: otel/demo-cart:v1.0
`

	valuesContent := `
components:
  cartservice:
    image:
      repository: otel/demo-cart
`

	files := []TargetFile{
		{
			Repo:      "otel-demo",
			CommitSHA: "sha1",
			Path:      "docker-compose.yaml",
			Content:   []byte(composeContent),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha1",
			Path:      "values.yaml",
			Content:   []byte(valuesContent),
		},
	}

	mm, err := pipe.ProcessFiles(context.Background(), "otel-demo", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	if len(mm.Nodes) != 1 {
		t.Fatalf("expected 1 deduplicated node, got %d", len(mm.Nodes))
	}

	cart := mm.Nodes[0]
	if cart.Name != "cartservice" {
		t.Fatalf("unexpected node name: %s", cart.Name)
	}

	// Evidence should be merged from both files
	if len(cart.Evidence) != 2 {
		t.Fatalf("expected 2 evidence records, got %d", len(cart.Evidence))
	}
}

func TestPipeline_ApplyOverrides_Relationships(t *testing.T) {
	pipe := DefaultPipeline()
	mm := domain.NewMacroMap("test-sys")

	_ = mm.AddNode(domain.Node{
		ID:   "service:frontend",
		Kind: domain.NodeService,
		Name: "frontend",
	})
	_ = mm.AddNode(domain.Node{
		ID:   "service:backend",
		Kind: domain.NodeService,
		Name: "backend",
	})
	_ = mm.AddNode(domain.Node{
		ID:   "contract:existing-cart",
		Kind: domain.NodeContract,
		Name: "existing-cart",
	})
	_ = mm.AddRelationship(domain.Relationship{
		ID:       "rel:frontend-accesses-backend",
		SourceID: "service:frontend",
		TargetID: "service:backend",
		Verb:     domain.VerbAccesses,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{Path: "docker-compose.yaml"},
				Origin:  domain.OriginExtracted,
			},
		},
	})

	overrides := &manifest.OverridesDeclaration{
		Relationships: []manifest.RelationshipOverride{
			// 1. Missing target node -> must NOT fabricate a dummy node; must record UnresolvedReference
			{
				SourceID:  "service:frontend",
				TargetID:  "contract:custom-cart",
				Verb:      "consumes",
				StartLine: 12,
				EndLine:   14,
			},
			// 2. Existing target node -> adds relationship with declared:override evidence
			{
				SourceID:  "service:frontend",
				TargetID:  "contract:existing-cart",
				Verb:      "consumes",
				StartLine: 15,
				EndLine:   17,
			},
			// 3. Override existing relationship verb -> updates verb and records ConflictRecord
			{
				SourceID:  "service:frontend",
				TargetID:  "service:backend",
				Verb:      "provides",
				StartLine: 18,
				EndLine:   20,
			},
		},
	}

	err := pipe.ApplyOverrides(mm, overrides, "custom-gira.yaml")
	if err != nil {
		t.Fatalf("ApplyOverrides failed: %v", err)
	}

	// Verify contract:custom-cart node was NOT fabricated as a dummy node
	for i := range mm.Nodes {
		if mm.Nodes[i].ID == "contract:custom-cart" {
			t.Fatalf("expected contract:custom-cart NOT to be fabricated as a phantom node")
		}
	}

	// Verify UnresolvedReference recorded for contract:custom-cart with exact line span
	if len(mm.UnresolvedRefs) != 1 {
		t.Fatalf("expected 1 unresolved reference, got %d: %+v", len(mm.UnresolvedRefs), mm.UnresolvedRefs)
	}
	uref := mm.UnresolvedRefs[0]
	if uref.SourceID != "service:frontend" || uref.TargetID != "contract:custom-cart" || uref.MissingID != "contract:custom-cart" {
		t.Errorf("unexpected unresolved reference: %+v", uref)
	}
	if len(uref.Evidence) == 0 || uref.Evidence[0].Locator.Path != "custom-gira.yaml" || uref.Evidence[0].Locator.StartLine != 12 || uref.Evidence[0].Locator.EndLine != 14 {
		t.Errorf("expected custom-gira.yaml:12-14 in unresolved ref evidence, got %+v", uref.Evidence)
	}

	// Verify valid new relationship to existing contract:existing-cart
	var customRel *domain.Relationship
	for i := range mm.Relationships {
		if mm.Relationships[i].TargetID == "contract:existing-cart" {
			customRel = &mm.Relationships[i]
			break
		}
	}
	if customRel == nil {
		t.Fatalf("expected relationship to contract:existing-cart")
	}
	if customRel.Verb != domain.VerbConsumes {
		t.Errorf("expected consumes verb, got %s", customRel.Verb)
	}
	if len(customRel.Evidence) == 0 || customRel.Evidence[0].Origin != OriginDeclaredOverride {
		t.Errorf("expected declared:override origin in relationship evidence")
	}
	if customRel.Evidence[0].Locator.Path != "custom-gira.yaml" || customRel.Evidence[0].Locator.StartLine != 15 {
		t.Errorf("expected custom-gira.yaml:15 locator, got %+v", customRel.Evidence[0].Locator)
	}

	// Verify overridden relationship and ConflictRecord
	var backendRel *domain.Relationship
	for i := range mm.Relationships {
		if mm.Relationships[i].TargetID == "service:backend" {
			backendRel = &mm.Relationships[i]
			break
		}
	}
	if backendRel == nil {
		t.Fatalf("expected relationship to service:backend")
	}
	if backendRel.Verb != domain.VerbProvides {
		t.Errorf("expected updated verb 'provides', got %s", backendRel.Verb)
	}
	if len(backendRel.Evidence) != 2 {
		t.Fatalf("expected 2 evidence records, got %d", len(backendRel.Evidence))
	}
	lastEvidence := backendRel.Evidence[1]
	if lastEvidence.Origin != OriginDeclaredOverride || lastEvidence.Locator.Path != "custom-gira.yaml" {
		t.Errorf("unexpected last evidence record: %+v", lastEvidence)
	}
	if len(mm.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict record for overridden relationship verb, got %d", len(mm.Conflicts))
	}
	if mm.Conflicts[0].Kind != "relationship_verb" || mm.Conflicts[0].PreviousVal != "accesses" || mm.Conflicts[0].DeclaredVal != "provides" {
		t.Errorf("unexpected conflict record: %+v", mm.Conflicts[0])
	}
}

func TestPipeline_ApplyOverrides_Metadata(t *testing.T) {
	mm := domain.NewMacroMap("test-sys")
	_ = mm.AddNode(domain.Node{
		ID:   "service:frontend",
		Kind: domain.NodeService,
		Name: "frontend",
		Metadata: map[string]any{
			"team": "legacy-team",
			"env":  "production",
		},
	})

	overrides := &manifest.OverridesDeclaration{
		Metadata: map[string]map[string]any{
			"service:frontend": {
				"team": "web-core",
				"tier": "tier-1",
			},
			"service:missing": {
				"team": "ghost-team",
			},
		},
		MetadataLines: map[string]int{
			"service:frontend": 20,
			"service:missing":  24,
		},
	}

	err := ApplyOverrides(mm, overrides, "gira.yaml")
	if err != nil {
		t.Fatalf("ApplyOverrides failed: %v", err)
	}

	if len(mm.Nodes) != 1 {
		t.Fatalf("expected 1 node (service:missing must not be created), got %d", len(mm.Nodes))
	}

	node := mm.Nodes[0]
	if node.Metadata["team"] != "web-core" {
		t.Errorf("expected team updated to web-core, got %v", node.Metadata["team"])
	}
	if node.Metadata["env"] != "production" {
		t.Errorf("expected env production preserved, got %v", node.Metadata["env"])
	}
	if node.Metadata["tier"] != "tier-1" {
		t.Errorf("expected tier tier-1 added, got %v", node.Metadata["tier"])
	}

	// Verify evidence added with declared:override and default manifest path gira.yaml + line 20
	if len(node.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record on node, got %d", len(node.Evidence))
	}
	ev := node.Evidence[0]
	if ev.Origin != OriginDeclaredOverride || ev.Locator.Path != "gira.yaml" || ev.Locator.StartLine != 20 {
		t.Errorf("unexpected evidence record on node: %+v", ev)
	}

	// Verify ConflictRecord for overwritten metadata key "team"
	if len(mm.Conflicts) != 1 {
		t.Fatalf("expected 1 metadata conflict record, got %d", len(mm.Conflicts))
	}
	if mm.Conflicts[0].Kind != "metadata_value" || mm.Conflicts[0].Field != "team" || mm.Conflicts[0].PreviousVal != "legacy-team" || mm.Conflicts[0].DeclaredVal != "web-core" {
		t.Errorf("unexpected metadata conflict record: %+v", mm.Conflicts[0])
	}

	// Verify UnresolvedReference for service:missing with line 24
	if len(mm.UnresolvedRefs) != 1 {
		t.Fatalf("expected 1 unresolved ref for service:missing, got %d", len(mm.UnresolvedRefs))
	}
	if mm.UnresolvedRefs[0].MissingID != "service:missing" || mm.UnresolvedRefs[0].Evidence[0].Locator.StartLine != 24 {
		t.Errorf("unexpected unresolved ref for service:missing: %+v", mm.UnresolvedRefs[0])
	}
}

func TestPipeline_ApplyOverrides_NilAndEdgeCases(t *testing.T) {
	if err := ApplyOverrides(nil, nil, ""); err == nil {
		t.Errorf("expected error for nil MacroMap, got nil")
	}

	mm := domain.NewMacroMap("test-sys")
	if err := ApplyOverrides(mm, nil, ""); err != nil {
		t.Errorf("expected nil error for nil overrides, got %v", err)
	}

	// Active overrides with empty manifest locator must return ErrManifestPathRequired
	activeOverrides := &manifest.OverridesDeclaration{
		Relationships: []manifest.RelationshipOverride{
			{SourceID: "s1", TargetID: "s2", Verb: "consumes"},
		},
	}
	if err := ApplyOverrides(mm, activeOverrides, ""); !errors.Is(err, manifest.ErrManifestPathRequired) {
		t.Errorf("expected ErrManifestPathRequired, got %v", err)
	}
}

func TestPipeline_Wave2Extractors(t *testing.T) {
	pipe := DefaultPipeline()

	// Verify all 12 extractors registered (Data, Dockerfile, Compose, EnvFile, Helm, GoMod, GHA, OpenAPI, Mise, Markdown, Proto, GoAST)
	extractors := pipe.Extractors()
	if len(extractors) != 12 {
		t.Fatalf("expected 12 extractors in DefaultPipeline, got %d", len(extractors))
	}

	ghaYaml := `name: CI
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`

	openapiYaml := `openapi: 3.0.0
info:
  title: Users Service
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
`

	miseToml := `[tools]
go = "1.24.0"

[tasks.test]
run = "go test ./..."
`

	files := []TargetFile{
		{Repo: "repo", CommitSHA: "c1", Path: ".github/workflows/ci.yml", Content: []byte(ghaYaml)},
		{Repo: "repo", CommitSHA: "c1", Path: "api/users-openapi.yaml", Content: []byte(openapiYaml)},
		{Repo: "repo", CommitSHA: "c1", Path: "mise.toml", Content: []byte(miseToml)},
	}

	mm, err := pipe.ProcessFiles(context.Background(), "test-system", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range mm.Nodes {
		nodeMap[n.ID] = n
	}

	// GitHub Actions nodes
	if _, ok := nodeMap["service:job:build"]; !ok {
		t.Errorf("missing service:job:build node")
	}
	if _, ok := nodeMap["library:actions/checkout"]; !ok {
		t.Errorf("missing library:actions/checkout node")
	}

	// OpenAPI nodes
	if _, ok := nodeMap["service:users-service"]; !ok {
		t.Errorf("missing service:users-service node")
	}
	if _, ok := nodeMap["contract:/users:get"]; !ok {
		t.Errorf("missing contract:/users:get node")
	}

	// Mise nodes
	if _, ok := nodeMap["library:go"]; !ok {
		t.Errorf("missing library:go node")
	}
	if _, ok := nodeMap["service:task:test"]; !ok {
		t.Errorf("missing service:task:test node")
	}

	// Relationships
	relMap := make(map[string]bool)
	for _, r := range mm.Relationships {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = true
	}

	if !relMap["service:job:build->consumes->library:actions/checkout"] {
		t.Errorf("missing GHA consumes relationship")
	}
	if !relMap["service:users-service->provides->contract:/users:get"] {
		t.Errorf("missing OpenAPI provides relationship")
	}
	if !relMap["service:task:test->consumes->library:go"] {
		t.Errorf("missing Mise consumes relationship")
	}
}

func TestPipeline_MarkdownIntegration(t *testing.T) {
	pipe := DefaultPipeline()

	taskMD := `---
schema: task/1
tier: standard
status: active
---

# Feature Task

Blocked by: [Task Dependency](tasks/dep-task.md)
implements: 0001-architecture
implements: missing-policy

## Verification

| ID | Check | Result |
|---|---|---|
| AC1 | Model expansion | pass |
`

	adrMD := `---
schema: dec/1
---

# ADR 0001: Architecture Decision

governs: payment-service, auth-service
`

	skillMD := `---
name: code-reviewer
---

# Code Reviewer Skill
`

	files := []TargetFile{
		{Repo: "repo", CommitSHA: "c1", Path: "tasks/feature-task.md", Content: []byte(taskMD)},
		{Repo: "repo", CommitSHA: "c1", Path: "docs/decisions/0001-architecture.md", Content: []byte(adrMD)},
		{Repo: "repo", CommitSHA: "c1", Path: "skills/code-reviewer/SKILL.md", Content: []byte(skillMD)},
	}

	mm, err := pipe.ProcessFiles(context.Background(), "test-system", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range mm.Nodes {
		nodeMap[n.ID] = n
	}

	// Only the 3 real extracted nodes should exist (no dummy nodes for dep-task, missing-policy, check:ac1, payment-service, auth-service)
	if len(mm.Nodes) != 3 {
		t.Fatalf("expected exactly 3 real extracted nodes, got %d: %+v", len(mm.Nodes), mm.Nodes)
	}

	taskNode, hasTask := nodeMap["task:feature-task"]
	if !hasTask {
		t.Fatalf("missing task:feature-task node")
	}
	if taskNode.Kind != domain.NodeTask {
		t.Errorf("expected NodeTask, got %v", taskNode.Kind)
	}
	if taskNode.Name != "Feature Task" {
		t.Errorf("expected 'Feature Task', got %s", taskNode.Name)
	}

	adrNode, hasAdr := nodeMap["policy:0001-architecture"]
	if !hasAdr {
		t.Fatalf("missing policy:0001-architecture node")
	}
	if adrNode.Kind != domain.NodePolicy {
		t.Errorf("expected NodePolicy, got %v", adrNode.Kind)
	}

	skillNode, hasSkill := nodeMap["skill:code-reviewer"]
	if !hasSkill {
		t.Fatalf("missing skill:code-reviewer node")
	}
	if skillNode.Kind != domain.NodeSkill {
		t.Errorf("expected NodeSkill, got %v", skillNode.Kind)
	}

	// Verify resolved relationship between task:feature-task and policy:0001-architecture
	relMap := make(map[string]bool)
	for _, r := range mm.Relationships {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = true
	}
	if !relMap["task:feature-task->implements->policy:0001-architecture"] {
		t.Errorf("missing resolved implements relationship from feature-task to 0001-architecture")
	}

	// Verify references to absent nodes are recorded in UnresolvedRefs with provenance
	unrefMap := make(map[string]domain.UnresolvedReference)
	for _, u := range mm.UnresolvedRefs {
		unrefMap[u.SourceID+"->"+string(u.Verb)+"->"+u.TargetID] = u
	}
	wantUnresolved := []string{
		"task:dep-task->governs->task:feature-task",
		"task:feature-task->implements->policy:missing-policy",
		"task:feature-task->verifies->contract:check:ac1",
		"policy:0001-architecture->governs->service:payment-service",
		"policy:0001-architecture->governs->service:auth-service",
	}
	for _, key := range wantUnresolved {
		u, ok := unrefMap[key]
		if !ok {
			t.Errorf("missing expected UnresolvedRef %q (got %+v)", key, mm.UnresolvedRefs)
			continue
		}
		if len(u.Evidence) == 0 || u.Evidence[0].Locator.StartLine <= 0 {
			t.Errorf("expected line provenance on UnresolvedRef %q, got %+v", key, u.Evidence)
		}
	}
}

func TestPipeline_DockerfileIntegration(t *testing.T) {
	pipe := DefaultPipeline()

	dockerfileMultistage := `# Multi-stage backend build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/backend ./cmd/server

FROM alpine:3.19 AS runner
WORKDIR /app
COPY --from=builder /bin/backend /app/backend
EXPOSE 8080 9090
ENTRYPOINT ["/app/backend"]
`

	dockerfileDb := `FROM postgres:16-alpine
EXPOSE 5432
`

	files := []TargetFile{
		{
			Repo:      "myorg/myrepo",
			CommitSHA: "commit123",
			Path:      "services/backend/Dockerfile",
			Content:   []byte(dockerfileMultistage),
		},
		{
			Repo:      "myorg/myrepo",
			CommitSHA: "commit123",
			Path:      "infra/postgres/db.dockerfile",
			Content:   []byte(dockerfileDb),
		},
	}

	mm, err := pipe.ProcessFiles(context.Background(), "myrepo-system", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Verify no unsupported boundaries for Dockerfiles
	if len(mm.Unsupported) != 0 {
		t.Errorf("expected 0 unsupported boundaries, got %d: %+v", len(mm.Unsupported), mm.Unsupported)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range mm.Nodes {
		nodeMap[n.ID] = n
	}

	// Each Dockerfile is one Service named by the directory rule; stages and base
	// images are its metadata, not nodes (spec gira-impact-precision, cause 1).
	if len(mm.Nodes) != 2 || len(mm.Relationships) != 0 {
		t.Fatalf("want 2 Service nodes and no edges, got %+v / %+v", mm.Nodes, mm.Relationships)
	}
	backend, ok := nodeMap["service:backend"]
	if !ok || backend.Kind != domain.NodeService {
		t.Fatalf("missing service:backend, nodes: %+v", mm.Nodes)
	}
	ports, ok := backend.Metadata["ports"].([]string)
	if !ok || len(ports) != 2 || ports[0] != "8080" || ports[1] != "9090" {
		t.Errorf("expected ports [8080 9090], got %v", backend.Metadata["ports"])
	}
	if backend.Metadata["entrypoint"] != `["/app/backend"]` {
		t.Errorf("expected entrypoint [\"/app/backend\"], got %v", backend.Metadata["entrypoint"])
	}
	ev := backend.Evidence[0]
	if ev.DeploymentProfile != "dockerfile" || ev.Locator.Path != "services/backend/Dockerfile" || ev.Locator.StartLine != 9 || ev.Locator.EndLine != 9 {
		t.Errorf("backend cites the final FROM at line 9, got %+v", ev)
	}
	stages, _ := backend.Metadata["stages"].([]map[string]any)
	if len(stages) != 2 || stages[0]["name"] != "builder" || stages[1]["name"] != "runner" {
		t.Errorf("stages = %v", stages)
	}
	images, _ := backend.Metadata["base_images"].([]map[string]any)
	if len(images) != 2 || images[0]["image"] != "golang:1.24-alpine" || images[1]["image"] != "alpine:3.19" {
		t.Errorf("base_images = %v", images)
	}

	db, ok := nodeMap["service:db"]
	if !ok || db.Kind != domain.NodeService {
		t.Fatalf("missing service:db, got: %+v", mm.Nodes)
	}
	if images, _ := db.Metadata["base_images"].([]map[string]any); len(images) != 1 || images[0]["image"] != "postgres:16-alpine" {
		t.Errorf("db base_images = %v", db.Metadata["base_images"])
	}
}
