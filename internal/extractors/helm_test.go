package extractors

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestHelmExtractor_ChartYAML(t *testing.T) {
	chartContent := `
apiVersion: v2
name: opentelemetry-demo
description: OpenTelemetry Astronomy Shop Demo
version: 0.25.0
appVersion: 1.5.0
dependencies:
  - name: kafka
    version: 22.1.5
    repository: https://charts.bitnami.com/bitnami
    condition: kafka.enabled
  - name: valkey-cart
    version: 1.0.0
`
	ext := NewHelmExtractor()
	if ext.Name() != "Layer2Helm" {
		t.Fatalf("unexpected name: %s", ext.Name())
	}
	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("unexpected layer: %v", ext.Layer())
	}
	if !ext.CanExtract("Chart.yaml") || !ext.CanExtract("Chart.yml") || !ext.CanExtract("values.yaml") || !ext.CanExtract("values-prod.yaml") {
		t.Fatalf("CanExtract failed for helm files")
	}

	tf := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "abc1234567890",
		Path:      "charts/opentelemetry-demo/Chart.yaml",
		Content:   []byte(chartContent),
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

	// Root chart node
	chartNode, ok := nodeMap["opentelemetry-demo"]
	if !ok {
		t.Fatalf("missing opentelemetry-demo node")
	}
	if chartNode.Kind != domain.NodeService {
		t.Fatalf("expected NodeService, got %v", chartNode.Kind)
	}
	if chartNode.Evidence[0].DeploymentProfile != "helm:chart" {
		t.Fatalf("unexpected deployment profile: %s", chartNode.Evidence[0].DeploymentProfile)
	}

	// Dependencies
	kafka, ok := nodeMap["kafka"]
	if !ok {
		t.Fatalf("missing kafka node")
	}
	if kafka.Kind != domain.NodeInfrastructure {
		t.Fatalf("expected kafka to be NodeInfrastructure, got %v", kafka.Kind)
	}

	valkey, ok := nodeMap["valkey-cart"]
	if !ok {
		t.Fatalf("missing valkey-cart node")
	}
	if valkey.Kind != domain.NodeInfrastructure {
		t.Fatalf("expected valkey-cart to be NodeInfrastructure, got %v", valkey.Kind)
	}

	// Relationships
	if len(rels) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(rels))
	}
	for _, r := range rels {
		if r.Verb != domain.VerbAccesses {
			t.Fatalf("expected accesses verb for infrastructure dependency, got %s", r.Verb)
		}
	}
}

func TestHelmExtractor_ValuesYAML(t *testing.T) {
	valuesContent := `
components:
  frontend:
    image:
      repository: otel/demo-frontend
    env:
      - name: CART_SERVICE_ADDR
        value: "cartservice:7070"
      - name: VALKEY_ADDR
        value: "valkey:6379"
  cartservice:
    image:
      repository: otel/demo-cart
    env:
      - name: VALKEY_ADDR
        value: "valkey:6379"
  valkey:
    image:
      repository: valkey/valkey
`
	ext := NewHelmExtractor()
	tf := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "abc1234567890",
		Path:      "charts/opentelemetry-demo/values.yaml",
		Content:   []byte(valuesContent),
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

	valkeyNode, ok := nodeMap["valkey"]
	if !ok {
		t.Fatalf("missing valkey node")
	}
	if valkeyNode.Kind != domain.NodeInfrastructure {
		t.Fatalf("expected valkey to be NodeInfrastructure, got %v", valkeyNode.Kind)
	}

	// Relationships:
	// frontend -> cartservice (consumes)
	// frontend -> valkey (accesses)
	// cartservice -> valkey (accesses)
	if len(rels) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(rels))
	}

	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.ID] = r
	}

	if r, ok := relMap["rel:frontend-consumes-cartservice"]; !ok || r.Verb != domain.VerbConsumes {
		t.Fatalf("missing or wrong frontend->cartservice relation")
	}
	if r, ok := relMap["rel:frontend-accesses-valkey"]; !ok || r.Verb != domain.VerbAccesses {
		t.Fatalf("missing or wrong frontend->valkey relation")
	}
	if r, ok := relMap["rel:cartservice-accesses-valkey"]; !ok || r.Verb != domain.VerbAccesses {
		t.Fatalf("missing or wrong cartservice->valkey relation")
	}
}

func TestHelmExtractor_Malformed(t *testing.T) {
	ext := NewHelmExtractor()
	tf := TargetFile{
		Repo:      "test",
		CommitSHA: "123",
		Path:      "Chart.yaml",
		Content:   []byte("name: [malformed"),
	}
	_, _, _, err := ext.Extract(context.Background(), tf)
	if err == nil {
		t.Fatalf("expected error on malformed Chart.yaml")
	}
}
