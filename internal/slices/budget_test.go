package slices

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

func ev(n int) []domain.EvidenceRecord {
	out := make([]domain.EvidenceRecord, n)
	for i := range out {
		out[i] = domain.EvidenceRecord{Locator: domain.Locator{Repo: "r", CommitSHA: "c", Path: fmt.Sprintf("src/file-%02d.go", i), StartLine: 1, EndLine: 9}, Origin: domain.OriginExtracted, Status: domain.StatusCurrent}
	}
	return out
}

func budgetMap() *domain.MacroMap {
	mm := domain.NewMacroMap("t")
	_ = mm.AddNode(domain.Node{ID: "service:a", Kind: domain.NodeService, Name: "a", Evidence: ev(6)})
	for _, id := range []string{"service:b", "service:c", "service:d"} {
		_ = mm.AddNode(domain.Node{ID: id, Kind: domain.NodeService, Name: id, Evidence: ev(3)})
		_ = mm.AddRelationship(domain.Relationship{ID: "rel:a-" + id, SourceID: "service:a", TargetID: id, Verb: domain.VerbImports, Evidence: ev(2)})
	}
	return mm
}

func emitted(t *testing.T, cs *ContextSlice, limit int) ContextSlice {
	t.Helper()
	out, err := cs.Emit(FormatJSON, limit)
	if err != nil {
		t.Fatal(err)
	}
	var got ContextSlice
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if tokens.Estimate(out) > limit || got.Budget.Estimated != tokens.Estimate(out) {
		t.Fatalf("budget not honoured: %d bytes, budget %+v", len(out), *got.Budget)
	}
	return got
}

func TestFitDropsEvidenceBeforeNodes(t *testing.T) {
	cs, err := NewEngine(budgetMap()).GetContextSlice("service:a", TierDeep)
	if err != nil {
		t.Fatal(err)
	}
	full, _ := cs.CanonicalJSON()

	trimmed := *cs
	trimmed.FocalNode.Evidence = cs.FocalNode.Evidence[:1]
	trimmed.Neighbors = nil
	for _, n := range cs.Neighbors {
		n.Evidence = n.Evidence[:1]
		trimmed.Neighbors = append(trimmed.Neighbors, n)
	}
	trimmed.Relationships = nil
	for _, r := range cs.Relationships {
		r.Evidence = r.Evidence[:1]
		trimmed.Relationships = append(trimmed.Relationships, r)
	}
	small, _ := trimmed.CanonicalJSON()

	// Room for the evidence-trimmed slice plus the budget marker, not for the full one.
	limit := tokens.Estimate(small) + 120
	if limit >= tokens.Estimate(full) {
		t.Fatalf("fixture too small: %d vs %d", limit, tokens.Estimate(full))
	}
	got := emitted(t, cs, limit)
	if !got.Budget.Truncated || got.Budget.Dropped.Nodes != 0 || got.Budget.Dropped.Evidence != 5+3*2+3*1 {
		t.Fatalf("step 1 should drop only extra evidence: %+v", *got.Budget)
	}

	// Less room: nodes go, highest ID first, with their relationships.
	got = emitted(t, cs, tokens.Estimate(small)-60)
	if got.Budget.Dropped.Nodes == 0 || got.Budget.Dropped.Relationships != got.Budget.Dropped.Nodes {
		t.Fatalf("step 2/3 should drop nodes with their edges: %+v", *got.Budget)
	}
	if got.FocalNode.ID != "service:a" {
		t.Fatal("focal node dropped")
	}
	for _, n := range got.Neighbors {
		if n.ID == "service:d" {
			t.Fatal("service:d has the highest ID and must be dropped first")
		}
	}
}

func TestDropOrderFarthestThenHighestID(t *testing.T) {
	got := dropOrder([]string{"b", "a", "c", "z"}, map[string]int{"a": 1, "b": 2, "c": 2})
	want := []string{"z", "c", "b", "a"} // z unreachable, then distance 2 by ID desc
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

// Spec gira-impact-precision SC1: Libraries reached through imports are listed in
// upstream_libraries, never in upstream_services, and a budget drops every one of
// them before any Service.
func TestImpactListsLibrariesSeparatelyAndDropsThemFirst(t *testing.T) {
	mm := budgetMap()
	for _, id := range []string{"library:x", "library:y"} {
		_ = mm.AddNode(domain.Node{ID: id, Kind: domain.NodeLibrary, Name: id, Evidence: ev(1)})
		_ = mm.AddRelationship(domain.Relationship{ID: "rel:a-" + id, SourceID: "service:a", TargetID: id, Verb: domain.VerbImports, Evidence: ev(1)})
	}
	ib, err := NewEngine(mm).GetImpactBriefing("service:a", 1)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(ids(ib.UpstreamLibraries)) != "[library:x library:y]" {
		t.Fatalf("upstream_libraries = %v", ids(ib.UpstreamLibraries))
	}
	for _, n := range ib.UpstreamServices {
		if n.Kind == domain.NodeLibrary {
			t.Fatalf("library %s listed in upstream_services", n.ID)
		}
	}
	full, _ := ib.CanonicalJSON()
	var got ImpactBriefing
	for limit := tokens.Estimate(full) - 1; limit > 0; limit -= 20 {
		out, err := ib.Emit(FormatJSON, limit)
		if err != nil {
			break
		}
		got = ImpactBriefing{}
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatal(err)
		}
		if got.Budget.Dropped.Nodes > 0 && len(got.UpstreamServices) < len(ib.UpstreamServices) && len(got.UpstreamLibraries) > 0 {
			t.Fatalf("limit %d dropped a Service while libraries %v remain", limit, ids(got.UpstreamLibraries))
		}
	}
}

func ids(ns []domain.Node) []string {
	out := []string{}
	for _, n := range ns {
		out = append(out, n.ID)
	}
	return out
}
