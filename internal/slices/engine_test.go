package slices_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// buildSyntheticOTelDemoMap constructs a representative MacroMap fixture modeled after OpenTelemetry Demo.
func buildSyntheticOTelDemoMap() *domain.MacroMap {
	m := domain.NewMacroMap("opentelemetry-demo")
	t0 := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)

	// Nodes
	_ = m.AddNode(domain.Node{
		ID:   "service:frontend",
		Kind: domain.NodeService,
		Name: "frontend",
		Metadata: map[string]any{
			"package": "main",
			"runtime": "go",
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/frontend/main.go",
					StartLine: 1,
					EndLine:   50,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 10,
					EndLine:   25,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "compose",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "service:cartservice",
		Kind: domain.NodeService,
		Name: "cartservice",
		Metadata: map[string]any{
			"package": "cart",
			"runtime": "go",
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/cartservice/main.go",
					StartLine: 1,
					EndLine:   40,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 30,
					EndLine:   45,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "compose",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "service:checkoutservice",
		Kind: domain.NodeService,
		Name: "checkoutservice",
		Metadata: map[string]any{
			"package": "checkout",
			"runtime": "go",
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/checkoutservice/main.go",
					StartLine: 1,
					EndLine:   70,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "service:paymentservice",
		Kind: domain.NodeService,
		Name: "paymentservice",
		Metadata: map[string]any{
			"package": "payment",
			"runtime": "go",
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/paymentservice/main.go",
					StartLine: 1,
					EndLine:   35,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "contract:CartService",
		Kind: domain.NodeContract,
		Name: "CartService",
		Metadata: map[string]any{
			"syntax":  "proto3",
			"package": "hipstershop",
			"methods": []string{"AddItem", "GetCart", "EmptyCart"},
			"rpc_definitions": []map[string]string{
				{"name": "AddItem", "request": "AddItemRequest", "response": "Empty"},
				{"name": "GetCart", "request": "GetCartRequest", "response": "Cart"},
				{"name": "EmptyCart", "request": "EmptyCartRequest", "response": "Empty"},
			},
			"messages": []string{"AddItemRequest", "GetCartRequest", "EmptyCartRequest", "Cart"},
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "pb/demo.proto",
					StartLine: 20,
					EndLine:   45,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "protobuf",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "contract:CheckoutService",
		Kind: domain.NodeContract,
		Name: "CheckoutService",
		Metadata: map[string]any{
			"syntax":  "proto3",
			"package": "hipstershop",
			"methods": []string{"PlaceOrder"},
			"rpc_definitions": []map[string]string{
				{"name": "PlaceOrder", "request": "PlaceOrderRequest", "response": "OrderResult"},
			},
			"messages": []string{"PlaceOrderRequest", "OrderResult"},
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "pb/demo.proto",
					StartLine: 50,
					EndLine:   65,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "protobuf",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "contract:PaymentService",
		Kind: domain.NodeContract,
		Name: "PaymentService",
		Metadata: map[string]any{
			"syntax":  "proto3",
			"package": "hipstershop",
			"methods": []string{"Charge"},
			"rpc_definitions": []map[string]string{
				{"name": "Charge", "request": "ChargeRequest", "response": "ChargeResponse"},
			},
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "pb/demo.proto",
					StartLine: 70,
					EndLine:   85,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "protobuf",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "infrastructure:redis",
		Kind: domain.NodeInfrastructure,
		Name: "redis",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 80,
					EndLine:   90,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "compose",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "infrastructure:postgres",
		Kind: domain.NodeInfrastructure,
		Name: "postgres",
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 95,
					EndLine:   110,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "compose",
				Status:            domain.StatusStale, // Stale evidence for testing warnings
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddNode(domain.Node{
		ID:   "library:google.golang.org/grpc",
		Kind: domain.NodeLibrary,
		Name: "google.golang.org/grpc",
		Metadata: map[string]any{
			"import_path": "google.golang.org/grpc",
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/frontend/go.mod",
					StartLine: 12,
					EndLine:   12,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	// Relationships
	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:cartservice-provides-cartservice",
		SourceID: "service:cartservice",
		TargetID: "contract:CartService",
		Verb:     domain.VerbProvides,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/cartservice/main.go",
					StartLine: 25,
					EndLine:   25,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:frontend-consumes-cartservice",
		SourceID: "service:frontend",
		TargetID: "contract:CartService",
		Verb:     domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/frontend/main.go",
					StartLine: 35,
					EndLine:   35,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:checkoutservice-consumes-cartservice",
		SourceID: "service:checkoutservice",
		TargetID: "contract:CartService",
		Verb:     domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/checkoutservice/main.go",
					StartLine: 40,
					EndLine:   40,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:checkoutservice-provides-checkoutservice",
		SourceID: "service:checkoutservice",
		TargetID: "contract:CheckoutService",
		Verb:     domain.VerbProvides,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/checkoutservice/main.go",
					StartLine: 20,
					EndLine:   20,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:frontend-consumes-checkoutservice",
		SourceID: "service:frontend",
		TargetID: "contract:CheckoutService",
		Verb:     domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/frontend/main.go",
					StartLine: 45,
					EndLine:   45,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:paymentservice-provides-paymentservice",
		SourceID: "service:paymentservice",
		TargetID: "contract:PaymentService",
		Verb:     domain.VerbProvides,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/paymentservice/main.go",
					StartLine: 18,
					EndLine:   18,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:checkoutservice-consumes-paymentservice",
		SourceID: "service:checkoutservice",
		TargetID: "contract:PaymentService",
		Verb:     domain.VerbConsumes,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/checkoutservice/main.go",
					StartLine: 55,
					EndLine:   55,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:cartservice-accesses-redis",
		SourceID: "service:cartservice",
		TargetID: "infrastructure:redis",
		Verb:     domain.VerbAccesses,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 35,
					EndLine:   35,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "compose",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:checkoutservice-accesses-postgres",
		SourceID: "service:checkoutservice",
		TargetID: "infrastructure:postgres",
		Verb:     domain.VerbAccesses,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "docker-compose.yml",
					StartLine: 65,
					EndLine:   65,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "compose",
				Status:            domain.StatusStale, // Stale relationship evidence
				DiscoveredAt:      t0,
			},
		},
	})

	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:frontend-imports-grpc",
		SourceID: "service:frontend",
		TargetID: "library:google.golang.org/grpc",
		Verb:     domain.VerbImports,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      "otel-demo",
					CommitSHA: "e0c2e04",
					Path:      "src/frontend/main.go",
					StartLine: 8,
					EndLine:   8,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      t0,
			},
		},
	})

	// Unsupported boundaries
	m.Unsupported = append(m.Unsupported, domain.UnsupportedBoundary{
		Path:   "src/recommendationservice/recommendation_server.py",
		Reason: "unsupported language: python",
		Locator: domain.Locator{
			Repo:      "otel-demo",
			CommitSHA: "e0c2e04",
			Path:      "src/recommendationservice/recommendation_server.py",
			StartLine: 1,
			EndLine:   100,
		},
	})

	m.Unsupported = append(m.Unsupported, domain.UnsupportedBoundary{
		Path:   "src/emailservice/email_server.py",
		Reason: "no extractor registered for python",
		Locator: domain.Locator{
			Repo:      "otel-demo",
			CommitSHA: "e0c2e04",
			Path:      "src/emailservice/email_server.py",
			StartLine: 1,
			EndLine:   80,
		},
	})

	m.Unsupported = append(m.Unsupported, domain.UnsupportedBoundary{
		Path:   "src/cartservice/unsupported_helper.sh",
		Reason: "no extractor registered for shell script",
		Locator: domain.Locator{
			Repo:      "otel-demo",
			CommitSHA: "e0c2e04",
			Path:      "src/cartservice/unsupported_helper.sh",
			StartLine: 1,
			EndLine:   15,
		},
	})

	return m
}

func TestImpactBriefing_FromContractPerspective(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	briefing, err := engine.GetImpactBriefing("contract:CartService", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if briefing.TargetID != "contract:CartService" {
		t.Errorf("expected target ID contract:CartService, got %s", briefing.TargetID)
	}
	if briefing.TargetKind != domain.NodeContract {
		t.Errorf("expected target kind Contract, got %s", briefing.TargetKind)
	}

	// Upstream producer should be cartservice
	if len(briefing.UpstreamServices) != 1 || briefing.UpstreamServices[0].ID != "service:cartservice" {
		t.Fatalf("expected 1 upstream service (cartservice), got %v", briefing.UpstreamServices)
	}

	// Downstream consumers should be frontend and checkoutservice
	if len(briefing.DownstreamServices) != 2 {
		t.Fatalf("expected 2 downstream services, got %d", len(briefing.DownstreamServices))
	}
	expectedDownstream := map[string]bool{
		"service:checkoutservice": true,
		"service:frontend":        true,
	}
	for _, svc := range briefing.DownstreamServices {
		if !expectedDownstream[svc.ID] {
			t.Errorf("unexpected downstream service: %s", svc.ID)
		}
	}

	// Affected contracts must include contract:CartService
	if len(briefing.AffectedContracts) != 1 || briefing.AffectedContracts[0].ID != "contract:CartService" {
		t.Errorf("expected affected contract contract:CartService, got %v", briefing.AffectedContracts)
	}

	// Mediation edges must include provides and consumes edges
	if len(briefing.MediationEdges) != 3 {
		t.Errorf("expected 3 mediation edges, got %d", len(briefing.MediationEdges))
	}
}

func TestImpactBriefing_FromServicePerspective_Depth1(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	briefing, err := engine.GetImpactBriefing("service:cartservice", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if briefing.TargetID != "service:cartservice" {
		t.Errorf("expected target ID service:cartservice, got %s", briefing.TargetID)
	}

	// Downstream consumers of CartService contract
	if len(briefing.DownstreamServices) != 2 {
		t.Fatalf("expected 2 downstream services, got %d", len(briefing.DownstreamServices))
	}

	// Upstream dependencies: redis infrastructure
	if len(briefing.UpstreamServices) != 1 || briefing.UpstreamServices[0].ID != "infrastructure:redis" {
		t.Errorf("expected upstream dependency infrastructure:redis, got %v", briefing.UpstreamServices)
	}

	// Affected contracts: contract:CartService
	if len(briefing.AffectedContracts) != 1 || briefing.AffectedContracts[0].ID != "contract:CartService" {
		t.Errorf("expected affected contract contract:CartService, got %v", briefing.AffectedContracts)
	}

	// Correlated unsupported boundary: cartservice helper script
	var foundCartUB bool
	for _, ub := range briefing.UnsupportedBoundaries {
		if strings.Contains(ub.Path, "cartservice") {
			foundCartUB = true
			break
		}
	}
	if !foundCartUB {
		t.Errorf("expected unsupported boundary for cartservice, got %v", briefing.UnsupportedBoundaries)
	}
}

func TestImpactBriefing_FromServicePerspective_Depth2(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// paymentservice is consumed by checkoutservice, which provides CheckoutService consumed by frontend
	briefingDepth1, err := engine.GetImpactBriefing("service:paymentservice", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(briefingDepth1.DownstreamServices) != 1 || briefingDepth1.DownstreamServices[0].ID != "service:checkoutservice" {
		t.Errorf("depth 1 should only find checkoutservice, got %v", briefingDepth1.DownstreamServices)
	}

	briefingDepth2, err := engine.GetImpactBriefing("service:paymentservice", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At depth 2, checkoutservice provides CheckoutService which frontend consumes!
	// So downstream services must include checkoutservice and frontend
	downstreamIDs := make(map[string]bool)
	for _, s := range briefingDepth2.DownstreamServices {
		downstreamIDs[s.ID] = true
	}
	if !downstreamIDs["service:checkoutservice"] || !downstreamIDs["service:frontend"] {
		t.Errorf("depth 2 should include checkoutservice and frontend, got %v", downstreamIDs)
	}

	// Affected contracts at depth 2: PaymentService and CheckoutService
	contractIDs := make(map[string]bool)
	for _, c := range briefingDepth2.AffectedContracts {
		contractIDs[c.ID] = true
	}
	if !contractIDs["contract:PaymentService"] || !contractIDs["contract:CheckoutService"] {
		t.Errorf("expected both PaymentService and CheckoutService in affected contracts, got %v", contractIDs)
	}
}

func TestImpactBriefing_InvalidTarget(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	_, err := engine.GetImpactBriefing("service:nonexistent", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent target, got nil")
	}
}

func TestContextSlice_TiersProgression(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// 1. Summary Tier
	summarySlice, err := engine.GetContextSlice("service:cartservice", slices.TierSummary)
	if err != nil {
		t.Fatalf("unexpected error in summary slice: %v", err)
	}
	if summarySlice.Tier != slices.TierSummary {
		t.Errorf("expected TierSummary, got %s", summarySlice.Tier)
	}

	// Verify neighbors in summary tier: contract:CartService, infrastructure:redis
	if len(summarySlice.Neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(summarySlice.Neighbors))
	}

	// In summary tier: contract neighbor should NOT have RPC methods or message definitions
	for _, n := range summarySlice.Neighbors {
		if n.Kind == domain.NodeContract {
			if _, hasMethods := n.Metadata["methods"]; hasMethods {
				t.Errorf("summary tier should strip methods from contract metadata, found: %v", n.Metadata["methods"])
			}
			if _, hasRPCDefs := n.Metadata["rpc_definitions"]; hasRPCDefs {
				t.Errorf("summary tier should strip rpc_definitions from contract metadata, found: %v", n.Metadata["rpc_definitions"])
			}
			if _, hasMessages := n.Metadata["messages"]; hasMessages {
				t.Errorf("summary tier should strip messages from contract metadata, found: %v", n.Metadata["messages"])
			}
		}
		// Evidence should have at most 1 primary locator in summary tier
		if len(n.Evidence) > 1 {
			t.Errorf("summary tier should limit evidence to at most 1, got %d", len(n.Evidence))
		}
	}
	// Focal node evidence should also be limited to 1 primary record
	if len(summarySlice.FocalNode.Evidence) > 1 {
		t.Errorf("summary tier focal node evidence should be <= 1, got %d", len(summarySlice.FocalNode.Evidence))
	}
	// Unsupported boundaries should not be included in summary tier
	if len(summarySlice.UnsupportedBoundaries) != 0 {
		t.Errorf("summary tier should not populate unsupported boundaries, got %d", len(summarySlice.UnsupportedBoundaries))
	}

	// 2. Contracts Tier
	contractsSlice, err := engine.GetContextSlice("service:cartservice", slices.TierContracts)
	if err != nil {
		t.Fatalf("unexpected error in contracts slice: %v", err)
	}
	if contractsSlice.Tier != slices.TierContracts {
		t.Errorf("expected TierContracts, got %s", contractsSlice.Tier)
	}

	// In contracts tier: contract neighbor MUST have RPC methods and syntax
	var contractFound bool
	for _, n := range contractsSlice.Neighbors {
		if n.Kind == domain.NodeContract {
			contractFound = true
			if methods, ok := n.Metadata["methods"].([]string); !ok || len(methods) != 3 {
				t.Errorf("contracts tier must expand contract methods, got: %v", n.Metadata["methods"])
			}
			if rDefs, ok := n.Metadata["rpc_definitions"].([]map[string]string); !ok || len(rDefs) != 3 {
				t.Errorf("contracts tier must expand rpc_definitions, got: %v", n.Metadata["rpc_definitions"])
			}
		}
	}
	if !contractFound {
		t.Fatal("expected contract neighbor in contracts tier")
	}

	// 3. Deep Tier
	deepSlice, err := engine.GetContextSlice("service:cartservice", slices.TierDeep)
	if err != nil {
		t.Fatalf("unexpected error in deep slice: %v", err)
	}
	if deepSlice.Tier != slices.TierDeep {
		t.Errorf("expected TierDeep, got %s", deepSlice.Tier)
	}

	// In deep tier: focal node must have all evidence records (2 in our fixture: go + compose)
	if len(deepSlice.FocalNode.Evidence) != 2 {
		t.Errorf("deep tier must retain all focal node evidence records, got %d", len(deepSlice.FocalNode.Evidence))
	}

	// Deep tier must include correlated unsupported boundaries (e.g. cartservice shell helper)
	if len(deepSlice.UnsupportedBoundaries) == 0 {
		t.Errorf("deep tier must include relevant unsupported boundaries, got 0")
	}

	// Verify token budget scaling: Summary < Contracts < Deep
	sumJSON, _ := summarySlice.CanonicalJSON()
	conJSON, _ := contractsSlice.CanonicalJSON()
	deepJSON, _ := deepSlice.CanonicalJSON()

	if len(sumJSON) >= len(conJSON) {
		t.Errorf("expected summary JSON payload (%d) < contracts JSON payload (%d)", len(sumJSON), len(conJSON))
	}
	if len(conJSON) >= len(deepJSON) {
		t.Errorf("expected contracts JSON payload (%d) < deep JSON payload (%d)", len(conJSON), len(deepJSON))
	}
}

func TestContextSlice_InvalidNode(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	_, err := engine.GetContextSlice("service:ghost", slices.TierSummary)
	if err == nil {
		t.Fatal("expected error for nonexistent nodeID, got nil")
	}
}

func TestExplainEvidence(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// Test 1: Node evidence only
	reportNode, err := engine.ExplainEvidence("service:frontend", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reportNode.TargetID != "service:frontend" {
		t.Errorf("expected TargetID service:frontend, got %s", reportNode.TargetID)
	}
	if len(reportNode.Evidence) != 2 {
		t.Errorf("expected 2 evidence records for frontend, got %d", len(reportNode.Evidence))
	}
	if len(reportNode.DeploymentProfiles) != 2 {
		t.Errorf("expected 2 profiles (compose, go), got %v", reportNode.DeploymentProfiles)
	}
	if len(reportNode.Origins) != 2 {
		t.Errorf("expected 2 origins (declared, extracted), got %v", reportNode.Origins)
	}
	if len(reportNode.Statuses) != 1 || reportNode.Statuses[0] != domain.StatusCurrent {
		t.Errorf("expected status current, got %v", reportNode.Statuses)
	}

	// Test 2: Edge evidence only
	reportEdge, err := engine.ExplainEvidence("", "rel:cartservice-provides-cartservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reportEdge.EdgeID != "rel:cartservice-provides-cartservice" {
		t.Errorf("expected edgeID, got %s", reportEdge.EdgeID)
	}
	if len(reportEdge.Evidence) != 1 {
		t.Errorf("expected 1 evidence record for edge, got %d", len(reportEdge.Evidence))
	}

	// Test 3: Both node and edge
	reportBoth, err := engine.ExplainEvidence("service:cartservice", "rel:cartservice-provides-cartservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reportBoth.NodeID != "service:cartservice" || reportBoth.EdgeID != "rel:cartservice-provides-cartservice" {
		t.Errorf("expected both IDs populated, got node=%s, edge=%s", reportBoth.NodeID, reportBoth.EdgeID)
	}
	if len(reportBoth.Evidence) != 3 { // 2 on node + 1 on edge
		t.Errorf("expected 3 total evidence records, got %d", len(reportBoth.Evidence))
	}

	// Test 4: Stale evidence detection
	reportStale, err := engine.ExplainEvidence("infrastructure:postgres", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var hasStale bool
	for _, s := range reportStale.Statuses {
		if s == domain.StatusStale {
			hasStale = true
		}
	}
	if !hasStale {
		t.Errorf("expected stale status in postgres evidence, got %v", reportStale.Statuses)
	}

	// Test 5: Empty parameters error
	_, err = engine.ExplainEvidence("", "")
	if err == nil {
		t.Fatal("expected error when both nodeID and edgeID are empty")
	}

	// Test 6: Nonexistent node error
	_, err = engine.ExplainEvidence("node:invalid", "")
	if err == nil {
		t.Fatal("expected error for nonexistent nodeID")
	}

	// Test 7: Nonexistent edge error
	_, err = engine.ExplainEvidence("", "rel:invalid")
	if err == nil {
		t.Fatal("expected error for nonexistent edgeID")
	}
}

func TestLocateBoundary(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// 1. Search by node ID
	matches, err := engine.LocateBoundary("cartservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) < 2 {
		t.Errorf("expected at least 2 matches for cartservice, got %d", len(matches))
	}

	// 2. Search by contract RPC method name
	matchesMethod, err := engine.LocateBoundary("AddItem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var foundMethodMatch bool
	for _, m := range matchesMethod {
		if m.Kind == "contract_method" && strings.Contains(m.Details, "AddItem") {
			foundMethodMatch = true
			if m.ID != "contract:CartService" {
				t.Errorf("expected method match on contract:CartService, got %s", m.ID)
			}
		}
	}
	if !foundMethodMatch {
		t.Errorf("expected match on RPC method AddItem, got %v", matchesMethod)
	}

	// 3. Search by unsupported boundary path and reason
	matchesPython, err := engine.LocateBoundary("python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matchesPython) != 2 {
		t.Errorf("expected 2 matches for python unsupported boundaries, got %d", len(matchesPython))
	}
	for _, m := range matchesPython {
		if m.Kind != "unsupported" {
			t.Errorf("expected kind unsupported, got %s", m.Kind)
		}
	}

	// 4. Search by node kind
	matchesKind, err := engine.LocateBoundary("infrastructure")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matchesKind) != 2 {
		t.Errorf("expected 2 infrastructure nodes, got %d", len(matchesKind))
	}

	// 5. Empty query error
	_, err = engine.LocateBoundary("   ")
	if err == nil {
		t.Fatal("expected error on empty query")
	}
}

func TestEngine_GetContextSliceWithView(t *testing.T) {
	m := domain.NewMacroMap("test-views")

	// Service node
	_ = m.AddNode(domain.Node{
		ID:   "service:api",
		Kind: domain.NodeService,
		Name: "api",
	})

	// Infrastructure node (neighbor 1)
	_ = m.AddNode(domain.Node{
		ID:   "infra:redis",
		Kind: domain.NodeInfrastructure,
		Name: "redis",
	})
	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:api-consumes-redis",
		SourceID: "service:api",
		TargetID: "infra:redis",
		Verb:     domain.VerbConsumes,
	})

	// Policy node (neighbor 2)
	_ = m.AddNode(domain.Node{
		ID:   "policy:adr-001",
		Kind: domain.NodePolicy,
		Name: "ADR 001",
	})
	_ = m.AddRelationship(domain.Relationship{
		ID:       "rel:adr-governs-api",
		SourceID: "policy:adr-001",
		TargetID: "service:api",
		Verb:     domain.VerbGoverns,
	})

	engine := slices.NewEngine(m)

	// 1. ViewFull: should see both infra:redis and policy:adr-001
	sliceFull, err := engine.GetContextSliceWithView("service:api", slices.TierSummary, slices.ViewFull)
	if err != nil {
		t.Fatalf("GetContextSliceWithView full failed: %v", err)
	}
	if len(sliceFull.Neighbors) != 2 {
		t.Errorf("expected 2 neighbors in ViewFull, got %d", len(sliceFull.Neighbors))
	}
	if len(sliceFull.Relationships) != 2 {
		t.Errorf("expected 2 relationships in ViewFull, got %d", len(sliceFull.Relationships))
	}

	// 2. ViewArchitecture: should ONLY see infra:redis
	sliceArch, err := engine.GetContextSliceWithView("service:api", slices.TierSummary, slices.ViewArchitecture)
	if err != nil {
		t.Fatalf("GetContextSliceWithView architecture failed: %v", err)
	}
	if len(sliceArch.Neighbors) != 1 || sliceArch.Neighbors[0].ID != "infra:redis" {
		t.Errorf("expected 1 neighbor (infra:redis) in ViewArchitecture, got %+v", sliceArch.Neighbors)
	}
	if len(sliceArch.Relationships) != 1 || sliceArch.Relationships[0].ID != "rel:api-consumes-redis" {
		t.Errorf("expected 1 relationship in ViewArchitecture, got %+v", sliceArch.Relationships)
	}

	// 3. ViewGovernance: should ONLY see policy:adr-001
	sliceGov, err := engine.GetContextSliceWithView("service:api", slices.TierSummary, slices.ViewGovernance)
	if err != nil {
		t.Fatalf("GetContextSliceWithView governance failed: %v", err)
	}
	if len(sliceGov.Neighbors) != 1 || sliceGov.Neighbors[0].ID != "policy:adr-001" {
		t.Errorf("expected 1 neighbor (policy:adr-001) in ViewGovernance, got %+v", sliceGov.Neighbors)
	}
	if len(sliceGov.Relationships) != 1 || sliceGov.Relationships[0].ID != "rel:adr-governs-api" {
		t.Errorf("expected 1 relationship in ViewGovernance, got %+v", sliceGov.Relationships)
	}
}

