package slices

import (
	"fmt"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// bigIntraDetail builds a synthetic go_intra_service far above any tier budget:
// the shape the Go extractor produces for a large service (dogfood
// service:internal carries ~1.6 MB of this metadata).
func bigIntraDetail(id string, n int) domain.IntraServiceDetail {
	d := domain.IntraServiceDetail{
		Schema:    "gira.intra-service/v1",
		ServiceID: id,
		GoModule:  "example.com/internal",
	}
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("Func%04d", i)
		loc := domain.Locator{Repo: "r", CommitSHA: "c", Path: fmt.Sprintf("internal/pkg%03d/f.go", i), StartLine: 5, EndLine: 9}
		d.Packages = append(d.Packages, domain.GoPackageSymbol{ID: id + "#pkg" + name, Name: "pkg" + name, Dir: "internal", Locator: loc, Origin: domain.OriginExtracted})
		d.Funcs = append(d.Funcs, domain.GoFuncSymbol{ID: id + "#" + name, Name: name, Exported: true, Locator: loc, Origin: domain.OriginExtracted})
		d.Calls = append(d.Calls, domain.GoCallRecord{ID: id + "#call" + name, Caller: name, Callee: "Other", Locator: loc, Origin: domain.OriginExtracted})
	}
	return d
}

func intraMap(n int) *domain.MacroMap {
	mm := domain.NewMacroMap("t")
	_ = mm.AddNode(domain.Node{ID: "service:internal", Kind: domain.NodeService, Name: "internal", Metadata: map[string]any{
		domain.IntraServiceMetadataKey: bigIntraDetail("service:internal", n),
	}, Evidence: ev(1)})
	_ = mm.AddNode(domain.Node{ID: "contract:Orders", Kind: domain.NodeContract, Name: "Orders", Evidence: ev(1)})
	_ = mm.AddRelationship(domain.Relationship{ID: "rel:internal-provides-orders", SourceID: "service:internal", TargetID: "contract:Orders", Verb: domain.VerbProvides, Evidence: ev(1)})
	return mm
}

// SC4: a Go node with a large go_intra_service must fit the summary tier under
// --budget 400 in JSON and caveman. The intra detail is dropped-able metadata
// served in full only by GetIntraServiceDoc (deep), never carried in the
// never-dropped focal block of a derived view. Regressed by G1: the 0.1.0
// binary failed slice service:internal --tier summary --budget 3000.
func TestSCAgentSummaryIntraBounded(t *testing.T) {
	eng := NewEngine(intraMap(400))

	for _, format := range []Format{FormatJSON, FormatCaveman, FormatMarkdown} {
		for _, tier := range []SliceTier{TierSummary, TierContracts, TierDeep} {
			cs, err := eng.GetContextSlice("service:internal", tier)
			if err != nil {
				t.Fatal(err)
			}
			out, err := cs.Emit(format, 400)
			if err != nil {
				t.Fatalf("%s/%s: %v", format, tier, err)
			}
			if strings.Contains(string(out), domain.IntraServiceMetadataKey) {
				t.Fatalf("%s/%s leaked %s", format, tier, domain.IntraServiceMetadataKey)
			}
		}
	}

	// The strip is the load-bearing invariant of this contract: guard it at the
	// unit level so a future cloneNode that keeps the key fails here first.
	detail := bigIntraDetail("service:internal", 3)
	node := domain.Node{ID: "service:x", Kind: domain.NodeService, Name: "x", Metadata: map[string]any{domain.IntraServiceMetadataKey: detail}, Evidence: ev(1)}
	cloned := cloneNode(node)
	if _, ok := cloned.Metadata[domain.IntraServiceMetadataKey]; ok {
		t.Fatal("cloneNode must strip go_intra_service from derived views")
	}
	if _, ok := node.Metadata[domain.IntraServiceMetadataKey]; !ok {
		t.Fatal("cloneNode must not mutate the source node (GetIntraServiceDoc reads it)")
	}
}

// SC4: the deep path keeps the full detail available through the documented
// accessor, so dropping it from slice envelopes loses no fact.
func TestSCAgentIntraDocKeepsFullDetail(t *testing.T) {
	eng := NewEngine(intraMap(400))
	doc, err := eng.GetIntraServiceDoc("service:internal")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Funcs) != 400 || len(doc.Calls) != 400 || len(doc.Packages) != 400 {
		t.Fatalf("GetIntraServiceDoc must serve the complete detail, got funcs=%d calls=%d packages=%d", len(doc.Funcs), len(doc.Calls), len(doc.Packages))
	}
}
