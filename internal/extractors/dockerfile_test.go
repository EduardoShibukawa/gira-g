package extractors

import (
	"context"
	"testing"

	"labs/gira-g/internal/domain"
)

func TestDockerfileExtractor_CanExtract(t *testing.T) {
	ext := NewDockerfileExtractor()

	positive := []string{
		"Dockerfile",
		"dockerfile",
		"DOCKERFILE",
		"Dockerfile.dev",
		"Dockerfile.prod",
		"Dockerfile.local",
		"app.dockerfile",
		"backend.dockerfile",
		"deploy/Dockerfile",
		"src/service/Dockerfile.prod",
		"frontend/app.dockerfile",
	}

	for _, p := range positive {
		if !ext.CanExtract(p) {
			t.Errorf("expected CanExtract(%q) == true, got false", p)
		}
	}

	negative := []string{
		"file.txt",
		"docker-compose.yml",
		"package.json",
		"Dockerfile_backup.txt",
		"not-a-dockerfile.go",
		"dockerfile_helper.py",
	}

	for _, p := range negative {
		if ext.CanExtract(p) {
			t.Errorf("expected CanExtract(%q) == false, got true", p)
		}
	}
}

func TestDockerfileExtractor_Simple(t *testing.T) {
	ext := NewDockerfileExtractor()

	content := `# syntax=docker/dockerfile:1
FROM node:18-alpine

WORKDIR /app

EXPOSE 3000 8080

CMD ["npm", "start"]
`

	file := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "abc1234",
		Path:      "services/auth/Dockerfile",
		Content:   []byte(content),
	}

	nodes, rels, unsupporteds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(unsupporteds) != 0 {
		t.Errorf("expected 0 unsupporteds, got %d", len(unsupporteds))
	}

	// Should have Service node and Library base image node
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	svcNode, ok := nodeMap["service:auth"]
	if !ok {
		t.Fatalf("expected service:auth node, got nodes: %+v", nodes)
	}

	if svcNode.Kind != domain.NodeService {
		t.Errorf("expected kind Service, got %v", svcNode.Kind)
	}

	// Check metadata: ports, cmd
	ports, ok := svcNode.Metadata["ports"].([]string)
	if !ok || len(ports) != 2 || ports[0] != "3000" || ports[1] != "8080" {
		t.Errorf("expected ports [3000 8080], got %v", svcNode.Metadata["ports"])
	}
	if svcNode.Metadata["cmd"] != `["npm", "start"]` {
		t.Errorf("expected cmd [npm start], got %v", svcNode.Metadata["cmd"])
	}

	// Base image node
	libNode, ok := nodeMap["library:node:18-alpine"]
	if !ok {
		t.Fatalf("expected library:node:18-alpine node, got nodes: %+v", nodes)
	}
	if libNode.Kind != domain.NodeLibrary {
		t.Errorf("expected kind Library for node:18-alpine, got %v", libNode.Kind)
	}

	// Locators
	if len(svcNode.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record for service, got %d", len(svcNode.Evidence))
	}
	ev := svcNode.Evidence[0]
	if ev.DeploymentProfile != "dockerfile" {
		t.Errorf("expected deployment_profile dockerfile, got %q", ev.DeploymentProfile)
	}
	if ev.Origin != domain.OriginExtracted {
		t.Errorf("expected origin extracted, got %v", ev.Origin)
	}
	if ev.Status != domain.StatusCurrent {
		t.Errorf("expected status current, got %v", ev.Status)
	}
	if ev.Locator.StartLine != 2 || ev.Locator.EndLine != 2 {
		t.Errorf("expected FROM at line 2, got %d-%d", ev.Locator.StartLine, ev.Locator.EndLine)
	}

	// Relationships
	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}
	rel := rels[0]
	if rel.SourceID != "service:auth" || rel.TargetID != "library:node:18-alpine" || rel.Verb != domain.VerbConsumes {
		t.Errorf("unexpected relationship: %+v", rel)
	}
}

func TestDockerfileExtractor_Multistage(t *testing.T) {
	ext := NewDockerfileExtractor()

	content := `FROM --platform=linux/amd64 golang:1.24 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/server .

FROM alpine:3.19 AS runner
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
`

	file := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "sha123",
		Path:      "Dockerfile",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// Nodes expected:
	// - service:stage:builder (intermediate stage)
	// - library:golang:1.24
	// - service:runner (or service:app if final stage, here runner since named AS runner)
	// - library:alpine:3.19
	if _, ok := nodeMap["service:stage:builder"]; !ok {
		t.Errorf("missing service:stage:builder, got nodes: %+v", nodes)
	}
	if _, ok := nodeMap["library:golang:1.24"]; !ok {
		t.Errorf("missing library:golang:1.24")
	}
	if _, ok := nodeMap["service:runner"]; !ok {
		t.Errorf("missing service:runner")
	}
	if _, ok := nodeMap["library:alpine:3.19"]; !ok {
		t.Errorf("missing library:alpine:3.19")
	}

	// Check runner metadata
	runner := nodeMap["service:runner"]
	if runner.Metadata["entrypoint"] != `["/app/server"]` {
		t.Errorf("expected entrypoint [\"/app/server\"], got %v", runner.Metadata["entrypoint"])
	}

	// Relationships expected:
	// 1. builder consumes library:golang:1.24
	// 2. runner consumes library:alpine:3.19
	// 3. runner consumes builder (via COPY --from=builder)
	relMap := make(map[string]bool)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = true
	}
	if !relMap["service:stage:builder->consumes->library:golang:1.24"] {
		t.Errorf("missing builder consumes golang relation")
	}
	if !relMap["service:runner->consumes->library:alpine:3.19"] {
		t.Errorf("missing runner consumes alpine relation")
	}

	var copyRel *domain.Relationship
	for i := range rels {
		if rels[i].SourceID == "service:runner" && rels[i].TargetID == "service:stage:builder" {
			copyRel = &rels[i]
			break
		}
	}
	if copyRel == nil {
		t.Fatalf("missing runner consumes builder relationship")
	}
	if len(copyRel.Evidence) == 0 {
		t.Fatalf("expected evidence for COPY relationship")
	}
	if copyRel.Evidence[0].Locator.StartLine != 8 || copyRel.Evidence[0].Locator.EndLine != 8 {
		t.Errorf("expected COPY locator line 8, got %d-%d", copyRel.Evidence[0].Locator.StartLine, copyRel.Evidence[0].Locator.EndLine)
	}
	if copyRel.Evidence[0].DeploymentProfile != "dockerfile" {
		t.Errorf("expected deployment_profile dockerfile, got %q", copyRel.Evidence[0].DeploymentProfile)
	}
}

func TestDockerfileExtractor_InfrastructureBaseImage(t *testing.T) {
	ext := NewDockerfileExtractor()

	content := `FROM postgres:16-alpine
ENV POSTGRES_PASSWORD=secret
EXPOSE 5432
`

	file := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "sha123",
		Path:      "db.dockerfile",
		Content:   []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// Service node should be service:db
	if _, ok := nodeMap["service:db"]; !ok {
		t.Errorf("expected service:db node, got: %+v", nodes)
	}

	// Infrastructure node for postgres
	infraNode, ok := nodeMap["infra:image:postgres:16-alpine"]
	if !ok {
		t.Fatalf("expected infra:image:postgres:16-alpine node, got: %+v", nodes)
	}
	if infraNode.Kind != domain.NodeInfrastructure {
		t.Errorf("expected kind Infrastructure, got %v", infraNode.Kind)
	}

	// Relationship
	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}
	if rels[0].SourceID != "service:db" || rels[0].TargetID != "infra:image:postgres:16-alpine" || rels[0].Verb != domain.VerbConsumes {
		t.Errorf("unexpected relationship: %+v", rels[0])
	}
}

func TestDockerfileExtractor_EdgeCases(t *testing.T) {
	ext := NewDockerfileExtractor()

	// 1. Empty content
	nodes, rels, _, err := ext.Extract(context.Background(), TargetFile{Path: "Dockerfile", Content: []byte("")})
	if err != nil {
		t.Fatalf("empty extract failed: %v", err)
	}
	if len(nodes) != 0 || len(rels) != 0 {
		t.Errorf("expected 0 nodes and rels on empty file")
	}

	// 2. Only comments and whitespace
	contentComments := `# Comment line 1
# Comment line 2

`
	nodes, rels, _, err = ext.Extract(context.Background(), TargetFile{Path: "Dockerfile", Content: []byte(contentComments)})
	if err != nil {
		t.Fatalf("comments extract failed: %v", err)
	}
	if len(nodes) != 0 || len(rels) != 0 {
		t.Errorf("expected 0 nodes on comments only")
	}

	// 3. Multi-line continuation with backslash \
	contentContinuation := `# Multi-line FROM
FROM \
  --platform=linux/arm64 \
  redis:7.2-alpine \
  AS \
  cache

EXPOSE \
  6379
`
	nodes, rels, _, err = ext.Extract(context.Background(), TargetFile{Path: "redis/Dockerfile", Content: []byte(contentContinuation)})
	if err != nil {
		t.Fatalf("continuation extract failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d: %+v", len(nodes), nodes)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	svc := nodeMap["service:redis"]
	if svc.Evidence[0].Locator.StartLine != 2 || svc.Evidence[0].Locator.EndLine != 6 {
		t.Errorf("expected line range 2-6 for multiline FROM, got %d-%d", svc.Evidence[0].Locator.StartLine, svc.Evidence[0].Locator.EndLine)
	}

	if _, ok := nodeMap["infra:image:redis:7.2-alpine"]; !ok {
		t.Errorf("missing infra redis node")
	}
}

func TestDockerfileExtractor_StageToStageReference(t *testing.T) {
	ext := NewDockerfileExtractor()

	content := `FROM node:20 AS base
WORKDIR /app

FROM base AS dev
ENV NODE_ENV=development
CMD ["yarn", "dev"]
`

	file := TargetFile{
		Repo:    "myrepo",
		Path:    "Dockerfile",
		Content: []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	relMap := make(map[string]bool)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = true
	}

	// dev stage should consume base stage
	if !relMap["service:dev->consumes->service:stage:base"] {
		t.Errorf("missing service:dev -> consumes -> service:stage:base dependency, got rels: %+v", rels)
	}

	// base stage should consume node:20
	if !relMap["service:stage:base->consumes->library:node:20"] {
		t.Errorf("missing base -> consumes -> library:node:20")
	}

	_ = nodes
}

func TestDockerfileExtractor_NumericAndExternalCopy(t *testing.T) {
	ext := NewDockerfileExtractor()

	content := `FROM golang:1.24
RUN go build -o /app/bin .

FROM alpine:3.19
COPY --from=0 /app/bin /usr/local/bin/app
COPY --from=redis:7.2 /usr/local/bin/redis-cli /usr/local/bin/redis-cli
`

	file := TargetFile{
		Repo:    "test-repo",
		Path:    "Dockerfile",
		Content: []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// External copy from redis:7.2 should create infra node
	if _, ok := nodeMap["infra:image:redis:7.2"]; !ok {
		t.Errorf("missing infra:image:redis:7.2 node from external COPY, got nodes: %+v", nodes)
	}

	relMap := make(map[string]bool)
	for _, r := range rels {
		relMap[r.SourceID+"->"+string(r.Verb)+"->"+r.TargetID] = true
	}

	// Stage 1 (app) should consume stage 0 (stage-0)
	if !relMap["service:app->consumes->service:stage:stage-0"] {
		t.Errorf("missing service:app -> consumes -> service:stage:stage-0 via COPY --from=0")
	}

	// Stage 1 (app) should consume redis infra image
	if !relMap["service:app->consumes->infra:image:redis:7.2"] {
		t.Errorf("missing service:app -> consumes -> infra:image:redis:7.2 via COPY --from=redis:7.2")
	}
}

func TestDockerfileExtractor_Deduplication(t *testing.T) {
	ext := NewDockerfileExtractor()

	// Multiple stages using the exact same base image
	content := `FROM alpine:3.19 AS stage1
RUN echo "stage1"

FROM alpine:3.19 AS stage2
RUN echo "stage2"

FROM alpine:3.19 AS final
RUN echo "final"
`

	file := TargetFile{
		Repo:    "test-repo",
		Path:    "Dockerfile",
		Content: []byte(content),
	}

	nodes, rels, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Count occurrences of library:alpine:3.19 node
	alpineCount := 0
	var alpineNode *domain.Node
	for i := range nodes {
		if nodes[i].ID == "library:alpine:3.19" {
			alpineCount++
			alpineNode = &nodes[i]
		}
	}

	if alpineCount != 1 {
		t.Fatalf("expected exactly 1 library:alpine:3.19 node due to deduplication, got %d", alpineCount)
	}

	// Should aggregate evidence from all 3 FROM directives
	if len(alpineNode.Evidence) != 3 {
		t.Errorf("expected 3 evidence records on deduplicated alpine node, got %d", len(alpineNode.Evidence))
	}

	if len(rels) != 3 {
		t.Errorf("expected 3 relationships, got %d", len(rels))
	}
}
