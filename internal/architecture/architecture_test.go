package architecture

import (
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestDomainClassification(t *testing.T) {
	tests := []struct {
		path string
		want DomainID
	}{
		{"cmd/gira", DomainEntrypoints},
		{"internal/mcp", DomainEntrypoints},
		{"internal/manipulator", DomainManipulator},
		{"internal/extractors", DomainExtractors},
		{"internal/slices", DomainCore},
		{"internal/domain", DomainCore},
		{"internal/gate", DomainGovernance},
		{"internal/eval", DomainGovernance},
		{"internal/store", DomainStorage},
		{"internal/gitcache", DomainStorage},
	}

	for _, tt := range tests {
		got := ClassifyDomain(tt.path)
		if got != tt.want {
			t.Errorf("ClassifyDomain(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestClusterMacroMap(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "test-system",
		Nodes: []domain.Node{
			{ID: "service:api", Kind: domain.NodeService, Name: "api", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "cmd/gira/main.go"}}}},
			{ID: "service:engine", Kind: domain.NodeService, Name: "engine", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "internal/slices/engine.go"}}}},
			{ID: "contract:order", Kind: domain.NodeContract, Name: "OrderContract", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "internal/domain/models.go"}}}},
			{ID: "infra:db", Kind: domain.NodeInfrastructure, Name: "Postgres", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "internal/store/store.go"}}}},
		},
		Relationships: []domain.Relationship{
			{ID: "rel:1", SourceID: "service:api", TargetID: "service:engine", Verb: domain.VerbConsumes},
			{ID: "rel:2", SourceID: "service:engine", TargetID: "contract:order", Verb: domain.VerbProvides},
			{ID: "rel:3", SourceID: "service:engine", TargetID: "infra:db", Verb: domain.VerbAccesses},
		},
	}

	arch := ClusterMacroMap(mm)
	if arch.TotalNodes != 4 {
		t.Fatalf("TotalNodes = %d, want 4", arch.TotalNodes)
	}

	// Verify contract is marked Core Abstraction
	orderNode, ok := arch.NodesByID["contract:order"]
	if !ok {
		t.Fatal("contract:order node missing")
	}
	if orderNode.Role != RoleCoreAbstraction {
		t.Errorf("contract:order role = %q, want %q", orderNode.Role, RoleCoreAbstraction)
	}

	// Check communities exist
	if len(arch.Communities) == 0 {
		t.Fatal("communities empty")
	}
}

func TestGenerateArchitectureMarkdown(t *testing.T) {
	mm := &domain.MacroMap{
		TargetSystem: "test-system",
		Nodes: []domain.Node{
			{ID: "service:core", Kind: domain.NodeService, Name: "core", Evidence: []domain.EvidenceRecord{{Locator: domain.Locator{Path: "internal/slices/engine.go"}}}},
		},
	}
	arch := ClusterMacroMap(mm)
	md := GenerateArchitectureMarkdown("/repo", arch, nil)

	if !strings.Contains(md, "Architectural Reference Overview") {
		t.Errorf("missing title in generated markdown")
	}
	if !strings.Contains(md, "Fundamental Design Invariants") {
		t.Errorf("missing invariants section")
	}
	if !strings.Contains(md, "```mermaid") {
		t.Errorf("missing mermaid diagram")
	}
	if !strings.Contains(md, "Critical Lifecycles & Intent Flows") {
		t.Errorf("missing lifecycles section")
	}
}

func TestQueryFlows(t *testing.T) {
	flow := QueryFlows("ast_edit")
	if flow == nil {
		t.Fatal("QueryFlows(ast_edit) returned nil")
	}
	if flow.Intent != "ast_edit" {
		t.Errorf("flow.Intent = %q, want ast_edit", flow.Intent)
	}
	if len(flow.Steps) == 0 {
		t.Error("flow.Steps is empty")
	}

	impactFlow := QueryFlows("impact")
	if impactFlow == nil || impactFlow.Intent != "impact" {
		t.Errorf("QueryFlows(impact) failed")
	}
}
