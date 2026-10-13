package extractors

import (
	"context"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestGoASTExtractor_Basic(t *testing.T) {
	ext := NewGoASTExtractor()

	if ext.Name() != "Layer3GoAST" {
		t.Fatalf("expected Layer3GoAST, got %s", ext.Name())
	}
	if ext.Layer() != LayerLanguageAST {
		t.Fatalf("expected LayerLanguageAST, got %v", ext.Layer())
	}
	if !ext.CanExtract("main.go") || !ext.CanExtract("src/service/handler.go") {
		t.Fatalf("CanExtract failed for Go source files")
	}
	if ext.CanExtract("demo.proto") || ext.CanExtract("data.json") {
		t.Fatalf("CanExtract returned true for non-Go files")
	}

	// Test files should be skipped cleanly
	testFile := TargetFile{
		Repo:      "repo",
		CommitSHA: "123",
		Path:      "service_test.go",
		Content:   []byte(`package main`),
	}
	nodes, rels, bounds, err := ext.Extract(context.Background(), testFile)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("expected nil results for test file, got err=%v, nodes=%d", err, len(nodes))
	}
}

func TestGoASTExtractor_DeriveServiceName(t *testing.T) {
	cases := []struct {
		path     string
		pkg      string
		expected string
	}{
		{"src/checkoutservice/main.go", "main", "checkoutservice"},
		{"services/payment/handler.go", "payment", "payment"},
		{"cmd/gira/main.go", "main", "gira"},
		{"cartservice/money.go", "cartservice", "cartservice"},
		{"main.go", "rootpkg", "rootpkg"},
	}

	for _, c := range cases {
		res := deriveServiceName(c.path, c.pkg)
		if res != c.expected {
			t.Errorf("deriveServiceName(%s, %s) = %s, expected %s", c.path, c.pkg, res, c.expected)
		}
	}
}

func TestGoASTExtractor_ImportsAndServerProvider(t *testing.T) {
	ext := NewGoASTExtractor()

	goSource := `package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"github.com/google/uuid"
	pb "github.com/open-telemetry/opentelemetry-demo/genproto/oteldemo"
)

type paymentServer struct {
	pb.UnimplementedPaymentServiceServer
}

func main() {
	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &paymentServer{})
}
`

	file := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "commit123",
		Path:      "src/paymentservice/main.go",
		Content:   []byte(goSource),
	}

	nodes, rels, bounds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("expected 0 unsupported boundaries, got %d", len(bounds))
	}

	// Verify Service node
	var foundSvc bool
	for _, n := range nodes {
		if n.ID == "service:paymentservice" {
			foundSvc = true
			if n.Kind != domain.NodeService {
				t.Fatalf("expected NodeService, got %s", n.Kind)
			}
			if n.Name != "paymentservice" {
				t.Fatalf("expected paymentservice, got %s", n.Name)
			}
		}
	}
	if !foundSvc {
		t.Fatalf("service:paymentservice node not found")
	}

	// Verify Library nodes (grpc, uuid, pb - stdlib fmt/context excluded)
	var foundUUIDLib, foundGRPCLib bool
	for _, n := range nodes {
		if n.ID == "library:github.com/google/uuid" {
			foundUUIDLib = true
			if n.Kind != domain.NodeLibrary {
				t.Fatalf("expected NodeLibrary, got %s", n.Kind)
			}
		}
		if n.ID == "library:google.golang.org/grpc" {
			foundGRPCLib = true
		}
		if n.ID == "library:fmt" || n.ID == "library:context" {
			t.Fatalf("stdlib library should not be extracted: %s", n.ID)
		}
	}
	if !foundUUIDLib || !foundGRPCLib {
		t.Fatalf("expected uuid and grpc libraries extracted")
	}

	// Verify Contract node
	var foundContract bool
	for _, n := range nodes {
		if n.ID == "contract:PaymentService" {
			foundContract = true
			if n.Kind != domain.NodeContract {
				t.Fatalf("expected NodeContract, got %s", n.Kind)
			}
			if n.Name != "PaymentService" {
				t.Fatalf("expected PaymentService, got %s", n.Name)
			}
		}
	}
	if !foundContract {
		t.Fatalf("contract:PaymentService node not found")
	}

	// Verify Relationships
	var foundImportRel, foundProvidesRel bool
	for _, r := range rels {
		if r.SourceID == "service:paymentservice" && r.TargetID == "library:github.com/google/uuid" && r.Verb == domain.VerbImports {
			foundImportRel = true
		}
		if r.SourceID == "service:paymentservice" && r.TargetID == "contract:PaymentService" && r.Verb == domain.VerbProvides {
			foundProvidesRel = true
			if len(r.Evidence) == 0 {
				t.Fatalf("expected evidence on provides relationship")
			}
			if r.Evidence[0].Origin != domain.OriginExtracted {
				t.Fatalf("expected OriginExtracted, got %s", r.Evidence[0].Origin)
			}
		}
	}
	if !foundImportRel {
		t.Fatalf("expected import relationship for uuid")
	}
	if !foundProvidesRel {
		t.Fatalf("expected provides relationship for contract:PaymentService")
	}
}

func TestGoASTExtractor_ClientConsumer(t *testing.T) {
	ext := NewGoASTExtractor()

	goSource := `package main

import (
	"context"
	"google.golang.org/grpc"
	pb "github.com/open-telemetry/opentelemetry-demo/genproto/oteldemo"
)

type checkoutService struct {
	paymentClient pb.PaymentServiceClient
}

func initPaymentClient(conn *grpc.ClientConn) {
	client := pb.NewPaymentServiceClient(conn)
	_ = client
}
`

	file := TargetFile{
		Repo:      "otel-demo",
		CommitSHA: "commit123",
		Path:      "src/checkoutservice/client.go",
		Content:   []byte(goSource),
	}

	nodes, rels, bounds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("expected 0 unsupported boundaries, got %d", len(bounds))
	}

	// Verify contract node exists
	var foundContract bool
	for _, n := range nodes {
		if n.ID == "contract:PaymentService" {
			foundContract = true
		}
	}
	if !foundContract {
		t.Fatalf("contract:PaymentService node not found")
	}

	// Verify consumes relationship: checkoutservice -consumes-> contract:PaymentService
	var foundConsumesRel bool
	for _, r := range rels {
		if r.SourceID == "service:checkoutservice" && r.TargetID == "contract:PaymentService" && r.Verb == domain.VerbConsumes {
			foundConsumesRel = true
			if len(r.Evidence) == 0 {
				t.Fatalf("expected evidence for consumes relationship")
			}
		}
	}
	if !foundConsumesRel {
		t.Fatalf("expected consumes relationship for contract:PaymentService")
	}
}

func TestPipeline_Layer3MediationEndToEnd(t *testing.T) {
	pipe := DefaultPipeline()

	protoContent := `syntax = "proto3";
package oteldemo;

service PaymentService {
    rpc Charge (ChargeRequest) returns (ChargeResponse);
}

message ChargeRequest {
    string id = 1;
}

message ChargeResponse {
    bool ok = 1;
}
`

	paymentGoSource := `package main

import (
	"google.golang.org/grpc"
	pb "github.com/open-telemetry/opentelemetry-demo/genproto/oteldemo"
)

type paymentServer struct {
	pb.UnimplementedPaymentServiceServer
}

func register(s *grpc.Server) {
	pb.RegisterPaymentServiceServer(s, &paymentServer{})
}
`

	checkoutGoSource := `package main

import (
	"google.golang.org/grpc"
	pb "github.com/open-telemetry/opentelemetry-demo/genproto/oteldemo"
)

func callPayment(conn *grpc.ClientConn) {
	client := pb.NewPaymentServiceClient(conn)
	_ = client
}
`

	composeContent := `version: '3.8'
services:
  paymentservice:
    image: otel/demo-payment:v1.0
  checkoutservice:
    image: otel/demo-checkout:v1.0
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
			Path:      "pb/demo.proto",
			Content:   []byte(protoContent),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha1",
			Path:      "src/paymentservice/main.go",
			Content:   []byte(paymentGoSource),
		},
		{
			Repo:      "otel-demo",
			CommitSHA: "sha1",
			Path:      "src/checkoutservice/main.go",
			Content:   []byte(checkoutGoSource),
		},
	}

	macroMap, err := pipe.ProcessFiles(context.Background(), "otel-demo", files)
	if err != nil {
		t.Fatalf("failed to process files: %v", err)
	}

	// Verify nodes
	nodeMap := make(map[string]domain.Node)
	for _, n := range macroMap.Nodes {
		nodeMap[n.ID] = n
	}

	if _, ok := nodeMap["service:paymentservice"]; !ok {
		t.Fatalf("service:paymentservice not found in macroMap")
	}
	if _, ok := nodeMap["service:checkoutservice"]; !ok {
		t.Fatalf("service:checkoutservice not found in macroMap")
	}

	contractNode, ok := nodeMap["contract:PaymentService"]
	if !ok {
		t.Fatalf("contract:PaymentService not found in macroMap")
	}
	if contractNode.Kind != domain.NodeContract {
		t.Fatalf("expected NodeContract, got %s", contractNode.Kind)
	}
	// Evidence should be merged from both proto and Go code
	if len(contractNode.Evidence) < 2 {
		t.Fatalf("expected evidence merged from both proto and Go code, got %d", len(contractNode.Evidence))
	}

	// Verify mediation relationships:
	// paymentservice -provides-> contract:PaymentService
	// checkoutservice -consumes-> contract:PaymentService
	var hasProvides, hasConsumes bool
	for _, r := range macroMap.Relationships {
		if r.SourceID == "service:paymentservice" && r.TargetID == "contract:PaymentService" && r.Verb == domain.VerbProvides {
			hasProvides = true
		}
		if r.SourceID == "service:checkoutservice" && r.TargetID == "contract:PaymentService" && r.Verb == domain.VerbConsumes {
			hasConsumes = true
		}
	}

	if !hasProvides {
		t.Fatalf("missing mediation relationship: paymentservice -provides-> contract:PaymentService")
	}
	if !hasConsumes {
		t.Fatalf("missing mediation relationship: checkoutservice -consumes-> contract:PaymentService")
	}

	// Verify canonical serialization
	jsonBytes, err := macroMap.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed canonical JSON serialization: %v", err)
	}
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"contract:PaymentService"`) || !strings.Contains(jsonStr, `"provides"`) || !strings.Contains(jsonStr, `"consumes"`) {
		t.Fatalf("canonical JSON missing expected content: %s", jsonStr)
	}
}
