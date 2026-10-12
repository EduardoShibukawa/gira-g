package extractors

import (
	"context"
	"strings"
	"testing"

	"labs/gira-g/internal/domain"
)

func TestProtobufExtractor_Basic(t *testing.T) {
	ext := NewProtobufExtractor()

	if ext.Name() != "Layer3Protobuf" {
		t.Fatalf("expected Layer3Protobuf, got %s", ext.Name())
	}
	if ext.Layer() != LayerLanguageAST {
		t.Fatalf("expected LayerLanguageAST, got %v", ext.Layer())
	}
	if !ext.CanExtract("demo.proto") || !ext.CanExtract("path/to/service.proto") {
		t.Fatalf("CanExtract failed for .proto file")
	}
	if ext.CanExtract("main.go") || ext.CanExtract("values.yaml") {
		t.Fatalf("CanExtract returned true for non-proto files")
	}
}

func TestProtobufExtractor_ServiceExtraction(t *testing.T) {
	ext := NewProtobufExtractor()

	protoContent := `syntax = "proto3";
package oteldemo;

/*
 Multi-line comment header
 */
service PaymentService {
    // Process charging
    rpc Charge (ChargeRequest) returns (ChargeResponse);
    rpc Refund (RefundRequest) returns (RefundResponse);
}

message ChargeRequest {
    string credit_card = 1;
    double amount = 2;
}

message ChargeResponse {
    string transaction_id = 1;
}
`

	file := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "abc1234",
		Path:      "pb/demo.proto",
		Content:   []byte(protoContent),
	}

	nodes, rels, bounds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("expected 0 unsupported boundaries, got %d", len(bounds))
	}
	if len(rels) != 0 {
		t.Fatalf("expected 0 relationships from protobuf extractor alone, got %d", len(rels))
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 Contract node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != "contract:PaymentService" {
		t.Fatalf("unexpected node ID: %s", node.ID)
	}
	if node.Kind != domain.NodeContract {
		t.Fatalf("expected NodeContract, got %s", node.Kind)
	}
	if node.Name != "PaymentService" {
		t.Fatalf("unexpected node name: %s", node.Name)
	}
	if node.Metadata["package"] != "oteldemo" {
		t.Fatalf("unexpected package metadata: %v", node.Metadata["package"])
	}
	if node.Metadata["syntax"] != "proto3" {
		t.Fatalf("unexpected syntax metadata: %v", node.Metadata["syntax"])
	}

	methods, ok := node.Metadata["methods"].([]string)
	if !ok || len(methods) != 2 || methods[0] != "Charge" || methods[1] != "Refund" {
		t.Fatalf("unexpected methods metadata: %v", node.Metadata["methods"])
	}

	if len(node.Evidence) != 1 {
		t.Fatalf("expected 1 evidence record, got %d", len(node.Evidence))
	}
	ev := node.Evidence[0]
	if ev.Locator.Path != "pb/demo.proto" {
		t.Fatalf("unexpected locator path: %s", ev.Locator.Path)
	}
	if ev.Locator.StartLine != 7 || ev.Locator.EndLine != 11 {
		t.Fatalf("unexpected service line range: L%d-L%d (expected L7-L11)", ev.Locator.StartLine, ev.Locator.EndLine)
	}
	if ev.Origin != domain.OriginDeclared {
		t.Fatalf("expected OriginDeclared, got %s", ev.Origin)
	}
	if ev.DeploymentProfile != "protobuf" {
		t.Fatalf("expected deployment profile protobuf, got %s", ev.DeploymentProfile)
	}
}

func TestProtobufExtractor_PureMessageSchema(t *testing.T) {
	ext := NewProtobufExtractor()

	protoContent := `syntax = "proto3";
package oteldemo;

message CartItem {
    string item_id = 1;
    int32 quantity = 2;
}
`

	file := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "abc1234",
		Path:      "pb/cart.proto",
		Content:   []byte(protoContent),
	}

	nodes, _, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 Contract node for schema, got %d", len(nodes))
	}
	if nodes[0].Kind != domain.NodeContract {
		t.Fatalf("expected NodeContract, got %s", nodes[0].Kind)
	}
}

func TestProtobufExtractor_MalformedUnclosedService(t *testing.T) {
	ext := NewProtobufExtractor()

	malformed := `syntax = "proto3";
service BadService {
    rpc Test(Req) returns (Resp);
`

	file := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "bad.proto",
		Content:   []byte(malformed),
	}

	_, _, _, err := ext.Extract(context.Background(), file)
	if err == nil || !strings.Contains(err.Error(), "unclosed service") {
		t.Fatalf("expected unclosed service error, got: %v", err)
	}
}
