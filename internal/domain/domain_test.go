package domain

import (
	"bytes"
	"testing"
	"time"
)

func TestMacroMapCanonicalJSON(t *testing.T) {
	mm := NewMacroMap("opentelemetry-demo")

	fixedTime := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

	svcCheckout := Node{
		ID:   "service:checkout",
		Kind: NodeService,
		Name: "checkoutservice",
		Evidence: []EvidenceRecord{
			{
				Locator: Locator{
					Repo:      "demo",
					CommitSHA: "dedc0178918e260823323b8d95005a8cb924b007",
					Path:      "src/checkoutservice/main.go",
					StartLine: 1,
					EndLine:   50,
				},
				Origin:            OriginExtracted,
				DeploymentProfile: "docker-compose",
				Status:            StatusCurrent,
				DiscoveredAt:      fixedTime,
			},
		},
	}

	contractOrder := Node{
		ID:   "contract:order-service",
		Kind: NodeContract,
		Name: "OrderService",
		Evidence: []EvidenceRecord{
			{
				Locator: Locator{
					Repo:      "demo",
					CommitSHA: "dedc0178918e260823323b8d95005a8cb924b007",
					Path:      "pb/demo.proto",
					StartLine: 10,
					EndLine:   35,
				},
				Origin:            OriginDeclared,
				DeploymentProfile: "protobuf",
				Status:            StatusCurrent,
				DiscoveredAt:      fixedTime,
			},
		},
	}

	svcFrontend := Node{
		ID:   "service:frontend",
		Kind: NodeService,
		Name: "frontend",
		Evidence: []EvidenceRecord{
			{
				Locator: Locator{
					Repo:      "demo",
					CommitSHA: "dedc0178918e260823323b8d95005a8cb924b007",
					Path:      "src/frontend/main.go",
					StartLine: 1,
					EndLine:   40,
				},
				Origin:            OriginExtracted,
				DeploymentProfile: "docker-compose",
				Status:            StatusCurrent,
				DiscoveredAt:      fixedTime,
			},
		},
	}

	// Add in non-alphabetical order to test sorting
	if err := mm.AddNode(svcFrontend); err != nil {
		t.Fatalf("unexpected error adding node: %v", err)
	}
	if err := mm.AddNode(svcCheckout); err != nil {
		t.Fatalf("unexpected error adding node: %v", err)
	}
	if err := mm.AddNode(contractOrder); err != nil {
		t.Fatalf("unexpected error adding node: %v", err)
	}

	relProvides := Relationship{
		ID:       "rel:checkout-provides-order",
		SourceID: "service:checkout",
		TargetID: "contract:order-service",
		Verb:     VerbProvides,
		Evidence: svcCheckout.Evidence,
	}

	relConsumes := Relationship{
		ID:       "rel:frontend-consumes-order",
		SourceID: "service:frontend",
		TargetID: "contract:order-service",
		Verb:     VerbConsumes,
		Evidence: svcFrontend.Evidence,
	}

	if err := mm.AddRelationship(relConsumes); err != nil {
		t.Fatalf("unexpected error adding relationship: %v", err)
	}
	if err := mm.AddRelationship(relProvides); err != nil {
		t.Fatalf("unexpected error adding relationship: %v", err)
	}

	mm.Unsupported = append(mm.Unsupported, UnsupportedBoundary{
		Path:   "src/paymentservice/index.js",
		Reason: "JavaScript AST parser not enabled in Wave 1",
		Locator: Locator{
			Repo:      "demo",
			CommitSHA: "dedc0178918e260823323b8d95005a8cb924b007",
			Path:      "src/paymentservice/index.js",
		},
	})

	data1, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("error serializing canonical JSON: %v", err)
	}

	data2, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("error serializing canonical JSON: %v", err)
	}

	if !bytes.Equal(data1, data2) {
		t.Fatalf("serialization is not byte-deterministic")
	}

	// Verify locator string formatting
	loc := Locator{
		Repo:      "demo",
		CommitSHA: "abc1234",
		Path:      "main.go",
		StartLine: 10,
		EndLine:   20,
	}
	expectedLoc := "demo@abc1234:main.go#L10-L20"
	if loc.String() != expectedLoc {
		t.Errorf("expected locator %q, got %q", expectedLoc, loc.String())
	}
}

func TestMacroMapValidation(t *testing.T) {
	mm := NewMacroMap("test-system")

	n1 := Node{ID: "node1", Kind: NodeService, Name: "Service 1"}
	if err := mm.AddNode(n1); err != nil {
		t.Fatalf("failed to add node: %v", err)
	}

	// Duplicate node error
	if err := mm.AddNode(n1); err == nil {
		t.Fatalf("expected error adding duplicate node, got nil")
	}

	// Missing target error
	rel := Relationship{
		ID:       "rel1",
		SourceID: "node1",
		TargetID: "nonexistent",
		Verb:     VerbConsumes,
	}
	if err := mm.AddRelationship(rel); err == nil {
		t.Fatalf("expected error adding relationship with missing target, got nil")
	}
}
