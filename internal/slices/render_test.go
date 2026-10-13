package slices_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

func approxTokens(s string) int {
	// A standard conservative heuristic for token counting: words + punctuation / ~3-4 chars
	fields := strings.Fields(s)
	return int(float64(len(fields))*1.3) + len(s)/20
}

func TestCanonicalJSON_DeterminismAndFormat(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// 1. ContextSlice CanonicalJSON
	cs, err := engine.GetContextSlice("service:cartservice", slices.TierDeep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bytes1, err := cs.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	bytes2, err := cs.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed second pass: %v", err)
	}

	if !bytes.Equal(bytes1, bytes2) {
		t.Fatal("ContextSlice CanonicalJSON is not byte-deterministic")
	}

	// Verify 2-space indentation
	if !strings.Contains(string(bytes1), "  \"focal_node\":") {
		t.Errorf("expected 2-space indentation in ContextSlice JSON: %s", string(bytes1[:100]))
	}

	// 2. ImpactBriefing CanonicalJSON
	ib, err := engine.GetImpactBriefing("contract:CartService", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ibBytes1, err := ib.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed on ImpactBriefing: %v", err)
	}
	ibBytes2, err := ib.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed second pass on ImpactBriefing: %v", err)
	}

	if !bytes.Equal(ibBytes1, ibBytes2) {
		t.Fatal("ImpactBriefing CanonicalJSON is not byte-deterministic")
	}
	if !strings.Contains(string(ibBytes1), "  \"target_id\":") {
		t.Errorf("expected 2-space indentation in ImpactBriefing JSON: %s", string(ibBytes1[:100]))
	}

	// 3. EvidenceReport CanonicalJSON
	er, err := engine.ExplainEvidence("service:cartservice", "rel:cartservice-provides-cartservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	erBytes1, err := er.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed on EvidenceReport: %v", err)
	}
	erBytes2, err := er.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed second pass on EvidenceReport: %v", err)
	}

	if !bytes.Equal(erBytes1, erBytes2) {
		t.Fatal("EvidenceReport CanonicalJSON is not byte-deterministic")
	}
	if !strings.Contains(string(erBytes1), "  \"target_id\":") {
		t.Errorf("expected 2-space indentation in EvidenceReport JSON: %s", string(erBytes1[:100]))
	}
}

func TestRenderMarkdown_ContextSliceNotationAndBudgets(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// Context slice on service:frontend (has provides, consumes, imports)
	summarySlice, err := engine.GetContextSlice("service:frontend", slices.TierSummary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	contractsSlice, err := engine.GetContextSlice("service:frontend", slices.TierContracts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deepSlice, err := engine.GetContextSlice("service:frontend", slices.TierDeep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaryMD := summarySlice.RenderMarkdown()
	contractsMD := contractsSlice.RenderMarkdown()
	deepMD := deepSlice.RenderMarkdown()

	// 1. Verify ASCII notation in Markdown
	if !strings.Contains(summaryMD, "<-consumes-") {
		t.Errorf("expected <-consumes- notation in markdown, got:\n%s", summaryMD)
	}
	if !strings.Contains(summaryMD, "-imports->") {
		t.Errorf("expected -imports-> notation in markdown, got:\n%s", summaryMD)
	}

	// Also check cartservice slice for -provides-> and -accesses->
	cartSlice, err := engine.GetContextSlice("service:cartservice", slices.TierSummary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cartMD := cartSlice.RenderMarkdown()
	if !strings.Contains(cartMD, "-provides->") {
		t.Errorf("expected -provides-> notation in markdown, got:\n%s", cartMD)
	}
	if !strings.Contains(cartMD, "-accesses->") {
		t.Errorf("expected -accesses-> notation in markdown, got:\n%s", cartMD)
	}

	// 2. Canonical locator citation format: [repo@commit:path#Lstart-Lend]
	expectedLocator := "[otel-demo@e0c2e04:src/frontend/main.go#L1-L50]"
	if !strings.Contains(summaryMD, expectedLocator) {
		t.Errorf("expected locator citation %s in markdown, got:\n%s", expectedLocator, summaryMD)
	}

	// 3. Progressive token budget and payload scaling
	summaryTokens := approxTokens(summaryMD)
	contractsTokens := approxTokens(contractsMD)
	deepTokens := approxTokens(deepMD)

	if summaryTokens >= contractsTokens {
		t.Errorf("expected summaryTokens (%d) < contractsTokens (%d)", summaryTokens, contractsTokens)
	}
	if contractsTokens >= deepTokens {
		t.Errorf("expected contractsTokens (%d) < deepTokens (%d)", contractsTokens, deepTokens)
	}

	// 4. Contract definitions: summary must NOT have RPC methods, contracts and deep MUST have them
	if strings.Contains(summaryMD, "AddItem(AddItemRequest)") {
		t.Errorf("summary tier should not render RPC method signatures")
	}
	if !strings.Contains(contractsMD, "AddItem(AddItemRequest)") {
		t.Errorf("contracts tier must render RPC method signatures")
	}
	if !strings.Contains(deepMD, "AddItem(AddItemRequest)") {
		t.Errorf("deep tier must render RPC method signatures")
	}

	// 5. Deep tier must have discrete evidence section
	if !strings.Contains(deepMD, "## Discrete Evidence Records") {
		t.Errorf("deep tier must render discrete evidence section")
	}
}

func TestRenderMarkdown_StaleAndUnsupportedNotices(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	// Checkoutservice is connected to postgres which has stale evidence
	checkoutSlice, err := engine.GetContextSlice("service:checkoutservice", slices.TierDeep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkoutMD := checkoutSlice.RenderMarkdown()

	if !strings.Contains(checkoutMD, "⚠️ NOTICE: Stale evidence for") {
		t.Errorf("expected stale notice in markdown for postgres neighbor, got:\n%s", checkoutMD)
	}

	// Impact briefing on cartservice should include unsupported notice
	briefing, err := engine.GetImpactBriefing("service:cartservice", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	briefingMD := briefing.RenderMarkdown()

	if !strings.Contains(briefingMD, "⚠️ UNSUPPORTED:") {
		t.Errorf("expected unsupported notice in impact briefing, got:\n%s", briefingMD)
	}
}

func TestRenderMarkdown_EvidenceReport(t *testing.T) {
	mm := buildSyntheticOTelDemoMap()
	engine := slices.NewEngine(mm)

	report, err := engine.ExplainEvidence("infrastructure:postgres", "rel:checkoutservice-accesses-postgres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := report.RenderMarkdown()

	if !strings.Contains(md, "# Evidence Report:") {
		t.Errorf("expected Evidence Report header in markdown")
	}
	if !strings.Contains(md, "Profiles: compose") {
		t.Errorf("expected Profiles in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "⚠️ NOTICE: Stale locator citation") {
		t.Errorf("expected stale locator citation notice in markdown, got:\n%s", md)
	}
}

func TestRenderBoundaryMatches(t *testing.T) {
	matches := []slices.BoundaryMatch{
		{
			Kind:      string(domain.NodeService),
			ID:        "service:cartservice",
			Name:      "cartservice",
			MatchedOn: "name",
		},
		{
			Kind:      "contract_method",
			ID:        "contract:CartService",
			Details:   "AddItem(AddItemRequest) -> Empty",
			MatchedOn: "method",
		},
		{
			Kind:      "unsupported",
			ID:        "src/recommendationservice/recommendation_server.py",
			Details:   "unsupported language: python",
			MatchedOn: "path",
		},
	}

	md := slices.RenderBoundaryMatches(matches)

	if !strings.Contains(md, "service:cartservice") {
		t.Errorf("expected service in boundary match markdown")
	}
	if !strings.Contains(md, "RPC Method `AddItem(AddItemRequest) -> Empty`") {
		t.Errorf("expected RPC method detail in boundary match markdown")
	}
	if !strings.Contains(md, "⚠️ UNSUPPORTED:") {
		t.Errorf("expected unsupported banner in boundary match markdown")
	}
}
