package slices

import (
	"fmt"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// cavemanMap: service:a focal, provides contract:c1 (with RPC definitions kept
// at the contracts tier), consumed by service:b, imports library:l1.
func cavemanMap() *domain.MacroMap {
	mm := domain.NewMacroMap("t")
	_ = mm.AddNode(domain.Node{ID: "service:a", Kind: domain.NodeService, Name: "a", Evidence: ev(1)})
	_ = mm.AddNode(domain.Node{ID: "contract:c1", Kind: domain.NodeContract, Name: "C1", Metadata: map[string]any{
		"syntax":  "proto3",
		"methods": []string{"Checkout"},
		"rpc_definitions": []map[string]string{
			{"name": "Checkout", "request": "CheckoutRequest", "response": "CheckoutResponse"},
		},
	}, Evidence: ev(1)})
	_ = mm.AddNode(domain.Node{ID: "service:b", Kind: domain.NodeService, Name: "b", Evidence: ev(1)})
	_ = mm.AddNode(domain.Node{ID: "library:l1", Kind: domain.NodeLibrary, Name: "l1", Evidence: ev(1)})
	_ = mm.AddRelationship(domain.Relationship{ID: "rel:a-provides-c1", SourceID: "service:a", TargetID: "contract:c1", Verb: domain.VerbProvides, Evidence: ev(1)})
	_ = mm.AddRelationship(domain.Relationship{ID: "rel:b-consumes-a", SourceID: "service:b", TargetID: "service:a", Verb: domain.VerbConsumes, Evidence: ev(1)})
	_ = mm.AddRelationship(domain.Relationship{ID: "rel:a-imports-l1", SourceID: "service:a", TargetID: "library:l1", Verb: domain.VerbImports, Evidence: ev(1)})
	return mm
}

// SC1: the caveman grammar is exactly one fact per line, deterministic between
// runs over the same index and budget.
func TestSCAgentCavemanGrammar(t *testing.T) {
	eng := NewEngine(cavemanMap())
	cs, err := eng.GetContextSlice("service:a", TierContracts)
	if err != nil {
		t.Fatal(err)
	}
	out, err := cs.Emit(FormatCaveman, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"Component: Service[service:a] @ src/file-00.go:L1",
		"Contracts:",
		"  - Contract[contract:c1] @ src/file-00.go:L1",
		"    - Checkout(CheckoutRequest) -> CheckoutResponse",
		"Dependencies:",
		"  - imports library:l1 @ src/file-00.go:L1",
		"  - provides contract:c1 @ src/file-00.go:L1",
		"Impact:",
		"  - consumed by Service[service:b] @ src/file-00.go:L1",
	}, "\n")
	got := strings.TrimRight(string(out), "\n")
	if got != want {
		t.Fatalf("grammar drift:\ngot:\n%s\nwant:\n%s", got, want)
	}

	// Determinism: same index, same budget, byte-identical.
	a, _ := cs.Emit(FormatCaveman, 400)
	b, _ := cs.Emit(FormatCaveman, 400)
	if string(a) != string(b) {
		t.Fatal("caveman output is not deterministic")
	}

	// Impact briefing grammar.
	ib, err := eng.GetImpactBriefing("service:a", 1)
	if err != nil {
		t.Fatal(err)
	}
	iout, err := ib.Emit(FormatCaveman, 0)
	if err != nil {
		t.Fatal(err)
	}
	is := string(iout)
	if !strings.HasPrefix(is, "Target: Service[service:a]\nDepth: 1\n") {
		t.Fatalf("briefing grammar:\n%s", is)
	}
	for _, line := range strings.Split(strings.TrimRight(is, "\n"), "\n") {
		if line == "" {
			t.Fatalf("blank briefing line: %q", line)
		}
		body := strings.TrimPrefix(line, "  ")
		if body != strings.TrimSpace(body) || (body != line && !strings.HasPrefix(body, "- ")) {
			t.Fatalf("mal-indented briefing line: %q", line)
		}
		ok := false
		for _, p := range []string{"Target: ", "Depth: ", "Upstream:", "Downstream:", "Libraries:", "Contracts:", "Boundaries:", "Budget:", "- "} {
			if strings.HasPrefix(body, p) {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("briefing line outside the grammar: %q", line)
		}
	}

	// Locate (boundary result) grammar.
	mm := cavemanMap()
	br := &BoundaryResult{Matches: []BoundaryMatch{
		{Kind: "node", ID: "service:a", Node: &mm.Nodes[0]},
	}}
	bout, err := br.Emit(FormatCaveman, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(bout)); got != "Match: Service[service:a] @ src/file-00.go:L1" {
		t.Fatalf("locate grammar: %q", got)
	}
}

// SC2: --budget on caveman truncates by whole facts (drop order), never inside
// a line, and keeps the honest budget marker; a medium component fits 400.
func TestSCAgentCavemanBudget(t *testing.T) {
	mm := domain.NewMacroMap("t")
	_ = mm.AddNode(domain.Node{ID: "service:checkout", Kind: domain.NodeService, Name: "checkout", Evidence: ev(1)})
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("service:neighbor-%02d", i)
		_ = mm.AddNode(domain.Node{ID: id, Kind: domain.NodeService, Name: id, Evidence: ev(3)})
		_ = mm.AddRelationship(domain.Relationship{ID: "rel-" + id, SourceID: "service:checkout", TargetID: id, Verb: domain.VerbImports, Evidence: ev(2)})
	}
	cs, err := NewEngine(mm).GetContextSlice("service:checkout", TierSummary)
	if err != nil {
		t.Fatal(err)
	}

	full, err := cs.Emit(FormatCaveman, 0)
	if err != nil {
		t.Fatal(err)
	}
	if tokens.Estimate(full) > 400 {
		t.Fatalf("medium component exceeds 400 without budget: %d", tokens.Estimate(full))
	}

	fitted, err := cs.Emit(FormatCaveman, tokens.Estimate(full)-50)
	if err != nil {
		t.Fatal(err)
	}
	if est := tokens.Estimate(fitted); est > tokens.Estimate(full)-50 {
		t.Fatalf("budget not honoured: %d", est)
	}
	lines := strings.Split(strings.TrimRight(string(fitted), "\n"), "\n")
	if !strings.HasPrefix(lines[0], "Component: Service[service:checkout]") {
		t.Fatal("focal fact must never be dropped")
	}
	marker := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "Budget: limit=") {
			marker = i
		}
	}
	if marker < 0 {
		t.Fatal("budget marker missing")
	}
	last := lines[marker]
	if !strings.Contains(last, "truncated=true") {
		t.Fatalf("budget marker must report truncation honestly: %q", last)
	}
	for _, l := range lines[marker+1:] {
		if !strings.HasPrefix(l, "Dropped: ") && !strings.HasPrefix(l, "Escalate: ") {
			t.Fatalf("only marker extras may follow the budget line: %q", l)
		}
	}
	// No fact is cut in the middle: every line is a whole grammar line.
	for _, l := range lines {
		if l == "" {
			t.Fatal("blank caveman line")
		}
		body := strings.TrimPrefix(l, "  ")
		if body != strings.TrimSpace(body) || (l != body && !strings.HasPrefix(body, "- ")) {
			t.Fatalf("mal-indented line: %q", l)
		}
	}
	// Honest envelope: Estimated equals the estimate of the emitted bytes.
	var markerEst int
	if _, err := fmt.Sscanf(last, "Budget: limit=%d estimated=%d", &markerEst, &markerEst); err != nil {
		t.Fatalf("marker: %q", last)
	}
	if markerEst != tokens.Estimate(fitted) {
		t.Fatalf("marker estimated=%d, bytes estimate=%d", markerEst, tokens.Estimate(fitted))
	}
}
