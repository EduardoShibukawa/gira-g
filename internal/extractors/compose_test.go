package extractors

import (
	"context"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestComposeExtractor_Extract(t *testing.T) {
	composeContent := `
version: '3.8'
services:
  frontend:
    image: otel/demo-frontend:v1.0
    ports:
      - "8080:8080"
    depends_on:
      - cartservice
      - redis
  cartservice:
    image: otel/demo-cart:v1.0
    depends_on:
      redis:
        condition: service_healthy
  redis:
    image: redis:6.2-alpine
`
	ext := NewComposeExtractor()
	if ext.Name() != "Layer2DockerCompose" {
		t.Fatalf("unexpected name: %s", ext.Name())
	}
	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("unexpected layer: %v", ext.Layer())
	}
	if !ext.CanExtract("docker-compose.yaml") || !ext.CanExtract("compose.yml") || !ext.CanExtract("docker-compose.override.yml") {
		t.Fatalf("CanExtract failed for valid compose names")
	}
	if ext.CanExtract("main.go") || ext.CanExtract("values.yaml") {
		t.Fatalf("CanExtract returned true for non-compose files")
	}

	tf := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "abc1234567890",
		Path:      "docker-compose.yaml",
		Content:   []byte(composeContent),
	}

	nodes, rels, boundaries, err := ext.Extract(context.Background(), tf)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(boundaries) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(boundaries))
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.Name] = n
	}

	// Verify frontend
	frontend, ok := nodeMap["frontend"]
	if !ok {
		t.Fatalf("missing frontend node")
	}
	if frontend.Kind != domain.NodeService {
		t.Fatalf("expected frontend to be NodeService, got %v", frontend.Kind)
	}
	if len(frontend.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record, got %d", len(frontend.Evidence))
	}
	if frontend.Evidence[0].Locator.StartLine <= 0 {
		t.Fatalf("expected line > 0, got %d", frontend.Evidence[0].Locator.StartLine)
	}
	if frontend.Evidence[0].DeploymentProfile != "compose:docker-compose.yaml" {
		t.Fatalf("unexpected deployment profile: %s", frontend.Evidence[0].DeploymentProfile)
	}

	// Verify redis (infrastructure)
	redis, ok := nodeMap["redis"]
	if !ok {
		t.Fatalf("missing redis node")
	}
	if redis.Kind != domain.NodeInfrastructure {
		t.Fatalf("expected redis to be NodeInfrastructure, got %v", redis.Kind)
	}

	// Verify cartservice
	cart, ok := nodeMap["cartservice"]
	if !ok {
		t.Fatalf("missing cartservice node")
	}
	if cart.Kind != domain.NodeService {
		t.Fatalf("expected cartservice to be NodeService, got %v", cart.Kind)
	}

	// Verify relationships:
	// frontend -> cartservice (consumes)
	// frontend -> redis (accesses, since redis is infrastructure)
	// cartservice -> redis (accesses)
	if len(rels) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(rels))
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.ID] = r
	}

	fcRel, ok := relMap["rel:frontend-consumes-cartservice"]
	if !ok {
		t.Fatalf("missing rel:frontend-consumes-cartservice, have %v", relMap)
	}
	if fcRel.Verb != domain.VerbConsumes {
		t.Fatalf("expected VerbConsumes, got %s", fcRel.Verb)
	}

	frRel, ok := relMap["rel:frontend-accesses-redis"]
	if !ok {
		t.Fatalf("missing rel:frontend-accesses-redis, have %v", relMap)
	}
	if frRel.Verb != domain.VerbAccesses {
		t.Fatalf("expected VerbAccesses, got %s", frRel.Verb)
	}

	crRel, ok := relMap["rel:cartservice-accesses-redis"]
	if !ok {
		t.Fatalf("missing rel:cartservice-accesses-redis, have %v", relMap)
	}
	if crRel.Verb != domain.VerbAccesses {
		t.Fatalf("expected VerbAccesses, got %s", crRel.Verb)
	}
}

func TestComposeExtractor_Malformed(t *testing.T) {
	ext := NewComposeExtractor()
	tf := TargetFile{
		Repo:      "test-repo",
		CommitSHA: "123",
		Path:      "compose.yaml",
		Content:   []byte("services: [malformed: {"),
	}
	_, _, _, err := ext.Extract(context.Background(), tf)
	if err == nil {
		t.Fatalf("expected error on malformed YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
