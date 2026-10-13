package archdiff

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestSegments(t *testing.T) {
	for _, tc := range []struct{ a, b, from, to string }{
		{"contract:ProductCatalogService", "contract:InventoryService", "ProductCatalog", "Inventory"},
		{"service:product-catalog", "service:inventory", "product-catalog", "inventory"},
		{"library:x/src/checkout/kafka", "library:x/src/ordering/kafka", "checkout", "ordering"},
		{"service:a", "service:a", "", ""},
	} {
		got := segments(tc.a, tc.b)
		if got[0] != tc.from || got[1] != tc.to {
			t.Errorf("segments(%q,%q) = %q, want [%q %q]", tc.a, tc.b, got, tc.from, tc.to)
		}
	}
}

func node(id string, kind domain.NodeType) domain.Node {
	return domain.Node{ID: id, Kind: kind, Evidence: []domain.EvidenceRecord{}}
}

// Two indistinguishable candidates are never paired: the diff reports a removal
// and two additions instead of guessing.
func TestAmbiguousRenameIsNotGuessed(t *testing.T) {
	a := domain.NewMacroMap("s")
	_ = a.AddNode(node("service:hub", domain.NodeService))
	_ = a.AddNode(node("service:old", domain.NodeService))
	_ = a.AddRelationship(domain.Relationship{ID: "r1", SourceID: "service:hub", TargetID: "service:old", Verb: domain.VerbConsumes})
	b := domain.NewMacroMap("s")
	_ = b.AddNode(node("service:hub", domain.NodeService))
	_ = b.AddNode(node("service:new1", domain.NodeService))
	_ = b.AddNode(node("service:new2", domain.NodeService))
	_ = b.AddRelationship(domain.Relationship{ID: "r1", SourceID: "service:hub", TargetID: "service:new1", Verb: domain.VerbConsumes})
	_ = b.AddRelationship(domain.Relationship{ID: "r2", SourceID: "service:hub", TargetID: "service:new2", Verb: domain.VerbConsumes})

	d := Compare(a, b)
	if len(d.Renamed) != 0 || len(d.RemovedNodes) != 1 || len(d.AddedNodes) != 2 {
		t.Fatalf("ambiguous rename must not be guessed: %+v", d)
	}
}
