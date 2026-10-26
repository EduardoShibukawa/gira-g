package ask

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestProcessQuery_Purpose(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "test-system",
		Nodes: []domain.Node{
			{ID: "service:backend", Name: "Backend Service", Kind: "service"},
		},
	}

	resp := ProcessQuery("Qual o objetivo deste projeto?", mm, true)
	if resp.Intent != "purpose" {
		t.Fatalf("expected intent 'purpose', got %q", resp.Intent)
	}
	if !strings.Contains(resp.Answer, "test-system") {
		t.Fatalf("expected answer to contain 'test-system', got: %s", resp.Answer)
	}
	if resp.Diagram == "" {
		t.Fatalf("expected mermaid diagram to be generated")
	}

	caveman := resp.RenderCaveman()
	if !strings.Contains(caveman, "QUERY:") || !strings.Contains(caveman, "INTENT: purpose") || !strings.Contains(caveman, "FACT:") {
		t.Fatalf("RenderCaveman missing required fields:\n%s", caveman)
	}
}

func TestProcessQuery_Quickstart(t *testing.T) {
	resp := ProcessQuery("Como eu utilizo o sistema?", nil, false)
	if resp.Intent != "quickstart" {
		t.Fatalf("expected intent 'quickstart', got %q", resp.Intent)
	}
	if !strings.Contains(resp.Answer, "gira init") {
		t.Fatalf("expected answer to contain 'gira init', got: %s", resp.Answer)
	}
}

func TestProcessQuery_Impact(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "store-system",
		Nodes: []domain.Node{
			{ID: "service:orders", Name: "Orders Service", Kind: "service"},
			{ID: "service:billing", Name: "Billing Service", Kind: "service"},
		},
		Relationships: []domain.Relationship{
			{SourceID: "service:orders", TargetID: "service:billing", Verb: "consumes"},
		},
	}

	resp := ProcessQuery("O que quebra se alterar billing?", mm, true)
	if resp.Intent != "impact" {
		t.Fatalf("expected intent 'impact', got %q", resp.Intent)
	}
	if len(resp.MatchedNodes) == 0 || resp.MatchedNodes[0] != "service:billing" {
		t.Fatalf("expected matched node 'service:billing', got %v", resp.MatchedNodes)
	}
	if !strings.Contains(resp.Answer, "service:orders") {
		t.Fatalf("expected impact answer to list orders caller, got: %s", resp.Answer)
	}
	if resp.Diagram == "" {
		t.Fatalf("expected impact mermaid diagram")
	}
}
