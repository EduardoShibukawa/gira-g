package extractors

import (
	"context"
	"strings"
	"testing"

	"labs/gira-g/internal/domain"
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
	// frontend -> cartservice (1)
	// opentelemetry-demo -> kafka (1)
	if len(mm.Relationships) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(mm.Relationships))
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
