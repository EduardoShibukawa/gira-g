package extractors

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
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

// extractOne runs the Dockerfile extractor and requires exactly one Service node
// and no relationships: stages and base images are metadata (spec
// gira-impact-precision, cause 1).
func extractOne(t *testing.T, path, content string) domain.Node {
	t.Helper()
	nodes, rels, unsupporteds, err := NewDockerfileExtractor().Extract(context.Background(), TargetFile{Repo: "test-repo", CommitSHA: "sha123", Path: path, Content: []byte(content)})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(unsupporteds) != 0 || len(rels) != 0 || len(nodes) != 1 {
		t.Fatalf("want one Service node and nothing else, got nodes %+v rels %+v unsupported %+v", nodes, rels, unsupporteds)
	}
	if nodes[0].Kind != domain.NodeService {
		t.Fatalf("want a Service, got %v", nodes[0].Kind)
	}
	return nodes[0]
}

// entries returns name→locator for a metadata list ("base_images" by image,
// "stages" by name).
func entries(t *testing.T, n domain.Node, key, field string) map[string]string {
	t.Helper()
	out := map[string]string{}
	list, _ := n.Metadata[key].([]map[string]any)
	for _, e := range list {
		out[e[field].(string)] = e["locator"].(string)
	}
	return out
}

func TestDockerfileExtractor_Simple(t *testing.T) {
	content := `# syntax=docker/dockerfile:1
FROM node:18-alpine

WORKDIR /app

EXPOSE 3000 8080

CMD ["npm", "start"]
`
	svc := extractOne(t, "services/auth/Dockerfile", content)
	if svc.ID != "service:auth" {
		t.Fatalf("expected service:auth, got %s", svc.ID)
	}
	ports, ok := svc.Metadata["ports"].([]string)
	if !ok || len(ports) != 2 || ports[0] != "3000" || ports[1] != "8080" {
		t.Errorf("expected ports [3000 8080], got %v", svc.Metadata["ports"])
	}
	if svc.Metadata["cmd"] != `["npm", "start"]` {
		t.Errorf("expected cmd [npm start], got %v", svc.Metadata["cmd"])
	}
	if got := entries(t, svc, "base_images", "image"); got["node:18-alpine"] != "services/auth/Dockerfile#L2" || len(got) != 1 {
		t.Errorf("base_images = %v", got)
	}
	if _, ok := svc.Metadata["stages"]; ok {
		t.Errorf("a single-stage Dockerfile records no stages: %v", svc.Metadata["stages"])
	}

	if len(svc.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record for service, got %d", len(svc.Evidence))
	}
	ev := svc.Evidence[0]
	if ev.DeploymentProfile != "dockerfile" || ev.Origin != domain.OriginExtracted || ev.Status != domain.StatusCurrent {
		t.Errorf("unexpected evidence %+v", ev)
	}
	if ev.Locator.StartLine != 2 || ev.Locator.EndLine != 2 {
		t.Errorf("expected FROM at line 2, got %d-%d", ev.Locator.StartLine, ev.Locator.EndLine)
	}
}

func TestDockerfileExtractor_Multistage(t *testing.T) {
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
	// The Service is named by the directory rule, never by a stage name.
	svc := extractOne(t, "Dockerfile", content)
	if svc.ID != "service:app" {
		t.Fatalf("expected service:app, got %s", svc.ID)
	}
	if svc.Metadata["entrypoint"] != `["/app/server"]` {
		t.Errorf("expected entrypoint of the final stage, got %v", svc.Metadata["entrypoint"])
	}
	stages := entries(t, svc, "stages", "name")
	if stages["builder"] != "Dockerfile#L1" || stages["runner"] != "Dockerfile#L6" || len(stages) != 2 {
		t.Errorf("stages = %v", stages)
	}
	images := entries(t, svc, "base_images", "image")
	if len(images) != 2 || images["golang:1.24"] == "" || images["alpine:3.19"] == "" {
		t.Errorf("base_images = %v (a stage referenced by COPY --from is not an image)", images)
	}
	if svc.Evidence[0].Locator.StartLine != 6 {
		t.Errorf("the Service cites the final FROM, got line %d", svc.Evidence[0].Locator.StartLine)
	}
}

func TestDockerfileExtractor_InfrastructureBaseImage(t *testing.T) {
	content := `FROM postgres:16-alpine
ENV POSTGRES_PASSWORD=secret
EXPOSE 5432
`
	// An infrastructure base image is still metadata: the Dockerfile builds the Service.
	svc := extractOne(t, "db.dockerfile", content)
	if svc.ID != "service:db" {
		t.Fatalf("expected service:db, got %s", svc.ID)
	}
	if got := entries(t, svc, "base_images", "image"); got["postgres:16-alpine"] == "" {
		t.Errorf("base_images = %v", got)
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
	svc := extractOne(t, "redis/Dockerfile", contentContinuation)
	if svc.ID != "service:redis" {
		t.Fatalf("expected service:redis, got %s", svc.ID)
	}
	if svc.Evidence[0].Locator.StartLine != 2 || svc.Evidence[0].Locator.EndLine != 6 {
		t.Errorf("expected line range 2-6 for multiline FROM, got %d-%d", svc.Evidence[0].Locator.StartLine, svc.Evidence[0].Locator.EndLine)
	}
	if got := entries(t, svc, "base_images", "image"); got["redis:7.2-alpine"] != "redis/Dockerfile#L2-L6" {
		t.Errorf("base_images = %v", got)
	}
}

func TestDockerfileExtractor_StageToStageReference(t *testing.T) {
	content := `FROM node:20 AS base
WORKDIR /app

FROM base AS dev
ENV NODE_ENV=development
CMD ["yarn", "dev"]
`
	svc := extractOne(t, "Dockerfile", content)
	images := entries(t, svc, "base_images", "image")
	if len(images) != 1 || images["node:20"] == "" {
		t.Errorf("FROM <stage> is not a base image: %v", images)
	}
	stages, _ := svc.Metadata["stages"].([]map[string]any)
	if len(stages) != 2 || stages[1]["name"] != "dev" || stages[1]["locator"] != "Dockerfile#L4" {
		t.Errorf("stages = %v", stages)
	}
}

func TestDockerfileExtractor_NumericAndExternalCopy(t *testing.T) {
	content := `FROM golang:1.24
RUN go build -o /app/bin .

FROM alpine:3.19
COPY --from=0 /app/bin /usr/local/bin/app
COPY --from=redis:7.2 /usr/local/bin/redis-cli /usr/local/bin/redis-cli
`
	svc := extractOne(t, "Dockerfile", content)
	images := entries(t, svc, "base_images", "image")
	if images["redis:7.2"] != "Dockerfile#L6" {
		t.Errorf("COPY --from=<image> is a base image at its own line: %v", images)
	}
	if _, ok := images["0"]; ok || len(images) != 3 {
		t.Errorf("COPY --from=<index> refers to a stage, not an image: %v", images)
	}
	stages := entries(t, svc, "stages", "name")
	if stages["stage-0"] == "" || stages["stage-1"] == "" {
		t.Errorf("unnamed stages are numbered: %v", stages)
	}
}

func TestDockerfileExtractor_Deduplication(t *testing.T) {
	content := `FROM alpine:3.19 AS stage1
RUN echo "stage1"

FROM alpine:3.19 AS stage2
RUN echo "stage2"

FROM alpine:3.19 AS final
RUN echo "final"
`
	svc := extractOne(t, "Dockerfile", content)
	images := entries(t, svc, "base_images", "image")
	if len(images) != 1 || images["alpine:3.19"] != "Dockerfile#L1" {
		t.Errorf("a repeated base image is recorded once, at its first line: %v", images)
	}
	if stages := entries(t, svc, "stages", "name"); len(stages) != 3 {
		t.Errorf("stages = %v", stages)
	}
}
