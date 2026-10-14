package slices_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// fanoutMap is a minimal order flow: frontend calls checkout directly and the
// CartService gRPC contract; checkout produces the Kafka topic `orders`, which
// accounting and fraud-detection consume (inferred from compose env wiring).
// accounting also produces `ledger`, consumed by audit, and admin is another
// consumer of the gRPC CartService. Neither audit nor admin is on the flow.
func fanoutMap(t *testing.T) *domain.MacroMap {
	t.Helper()
	t0 := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	ev := func(path string, line int, origin domain.DerivationOrigin) []domain.EvidenceRecord {
		return []domain.EvidenceRecord{{
			Locator:           domain.Locator{Repo: "demo", CommitSHA: "abc1234", Path: path, StartLine: line, EndLine: line},
			Origin:            origin,
			DeploymentProfile: "compose:" + path,
			Status:            domain.StatusCurrent,
			DiscoveredAt:      t0,
		}}
	}
	m := domain.NewMacroMap("demo")
	svc := func(id, path string, line int) {
		if err := m.AddNode(domain.Node{ID: id, Kind: domain.NodeService, Name: strings.TrimPrefix(id, "service:"), Evidence: ev(path, line, domain.OriginExtracted)}); err != nil {
			t.Fatal(err)
		}
	}
	svc("service:frontend", "compose.yaml", 1)
	svc("service:checkout", "compose.yaml", 10)
	svc("service:cart", "compose.yaml", 20)
	svc("service:admin", "compose.yaml", 30)
	svc("service:kafka", "compose.full.yaml", 100)
	svc("service:accounting", "compose.full.yaml", 12)
	svc("service:fraud-detection", "compose.full.yaml", 47)
	svc("service:audit", "compose.full.yaml", 200)
	for _, c := range []domain.Node{
		{ID: "contract:CartService", Kind: domain.NodeContract, Name: "CartService", Evidence: ev("pb/demo.proto", 5, domain.OriginExtracted)},
		{ID: "contract:kafka:orders", Kind: domain.NodeContract, Name: "orders", Metadata: map[string]any{"protocol": "kafka", "topic": "orders"}, Evidence: ev("src/checkout/main.go", 671, domain.OriginExtracted)},
		{ID: "contract:kafka:ledger", Kind: domain.NodeContract, Name: "ledger", Metadata: map[string]any{"protocol": "kafka", "topic": "ledger"}, Evidence: ev("src/accounting/main.go", 9, domain.OriginExtracted)},
	} {
		if err := m.AddNode(c); err != nil {
			t.Fatal(err)
		}
	}
	rel := func(src string, verb domain.RelationVerb, dst, path string, line int, origin domain.DerivationOrigin) {
		id := "rel:" + strings.TrimPrefix(src, "service:") + "-" + string(verb) + "-" + dst
		if err := m.AddRelationship(domain.Relationship{ID: id, SourceID: src, TargetID: dst, Verb: verb, Evidence: ev(path, line, origin)}); err != nil {
			t.Fatal(err)
		}
	}
	rel("service:frontend", domain.VerbConsumes, "service:checkout", "compose.yaml", 5, domain.OriginDeclared)
	rel("service:frontend", domain.VerbConsumes, "contract:CartService", "src/frontend/main.go", 40, domain.OriginExtracted)
	rel("service:cart", domain.VerbProvides, "contract:CartService", "src/cart/main.go", 12, domain.OriginExtracted)
	rel("service:admin", domain.VerbConsumes, "contract:CartService", "src/admin/main.go", 7, domain.OriginExtracted)
	rel("service:checkout", domain.VerbProvides, "contract:kafka:orders", "src/checkout/main.go", 671, domain.OriginExtracted)
	rel("service:kafka", domain.VerbProvides, "contract:kafka:orders", "compose.full.yaml", 132, domain.OriginDeclared)
	rel("service:accounting", domain.VerbConsumes, "contract:kafka:orders", "compose.full.yaml", 27, domain.OriginInferred)
	rel("service:fraud-detection", domain.VerbConsumes, "contract:kafka:orders", "compose.full.yaml", 66, domain.OriginInferred)
	rel("service:accounting", domain.VerbProvides, "contract:kafka:ledger", "src/accounting/main.go", 9, domain.OriginExtracted)
	rel("service:audit", domain.VerbConsumes, "contract:kafka:ledger", "compose.full.yaml", 210, domain.OriginInferred)
	return m
}

// SC4 (J4): a tour that visits a Kafka topic contract also enqueues the topic's
// consumers, one hop only: consumers of gRPC contracts are not fanned out, and a
// consumer reached by the fan-out does not fan out again. Inferred consume edges
// stay inferred (I2), and every fan-out claim cites the consume edge (I1).
func TestSCTourFanout(t *testing.T) {
	m := fanoutMap(t)
	tour, err := slices.NewEngine(m).GetTour("service:frontend", 50)
	if err != nil {
		t.Fatal(err)
	}
	steps := map[string]slices.TourStep{}
	for _, s := range tour.Steps {
		steps[s.NodeIDs[0]] = s
	}

	for _, id := range []string{"service:checkout", "service:accounting", "service:fraud-detection"} {
		if _, ok := steps[id]; !ok {
			t.Errorf("order flow: %s is not a tour stop", id)
		}
	}
	// One hop: the fan-out reaches the topic's consumers from the stop that
	// visited the topic, and stops there.
	for _, id := range []string{"service:accounting", "service:fraud-detection"} {
		s, ok := steps[id]
		if !ok {
			continue
		}
		if s.WhyTemplate != slices.WhyConsumer {
			t.Errorf("%s: why template %q, want %q", id, s.WhyTemplate, slices.WhyConsumer)
			continue
		}
		if len(s.WhyArgs) != 5 || s.WhyArgs[0] != id || s.WhyArgs[1] != "contract:kafka:orders" || s.WhyArgs[3] != "service:checkout" {
			t.Errorf("%s: why args %v, want [%s contract:kafka:orders <tier> service:checkout <verb>]", id, s.WhyArgs, id)
		}
		// I2: the inferred consume edge is presented as inferred.
		if s.WhyArgs[2] != string(domain.OriginInferred) || !strings.Contains(s.Why, "(inferred)") {
			t.Errorf("%s: why %q does not carry the edge's inferred tier", id, s.Why)
		}
		if s.Why != slices.RenderWhy(s.WhyTemplate, s.WhyArgs) {
			t.Errorf("%s: why %q is not the template renderer's output", id, s.Why)
		}
	}
	// I1: the claim cites the consume edge. Its line sits inside the consumer's
	// own compose block, so the citation starts at the consumer's declaration;
	// fraud-detection's edge evidence is still the range's last line.
	if s := steps["service:accounting"]; s.WhyLocator != "demo@abc1234:compose.full.yaml#L12-L27" {
		t.Errorf("accounting why_locator %q, want the block from its declaration (L12) to the consume edge (L27)", s.WhyLocator)
	}
	if s := steps["service:fraud-detection"]; s.WhyLocator != "demo@abc1234:compose.full.yaml#L47-L66" {
		t.Errorf("fraud-detection why_locator %q, want L47-L66", s.WhyLocator)
	}
	// No transitive explosion: audit consumes a topic that a fan-out consumer
	// produces; admin consumes a gRPC contract the walk visits.
	for _, id := range []string{"service:audit", "service:admin"} {
		if s, ok := steps[id]; ok {
			t.Errorf("%s reached by step %d (%s): the fan-out must stop at Kafka topic consumers, one hop", id, s.Order, s.Why)
		}
	}
	// I2: the walk never relabels the macro map's evidence.
	for _, r := range m.Relationships {
		if strings.HasSuffix(r.ID, "-consumes-contract:kafka:orders") && r.Evidence[0].Origin != domain.OriginInferred {
			t.Errorf("%s: origin %q, want inferred", r.ID, r.Evidence[0].Origin)
		}
	}
}

// The block citation never spans another node's declaration: when one sits
// between the consumer's declaration and the consume edge, the claim cites the
// edge line alone.
func TestTourFanoutCitationStaysInOwnBlock(t *testing.T) {
	m := fanoutMap(t)
	if err := m.AddNode(domain.Node{ID: "service:sidecar", Kind: domain.NodeService, Name: "sidecar", Evidence: []domain.EvidenceRecord{{
		Locator: domain.Locator{Repo: "demo", CommitSHA: "abc1234", Path: "compose.full.yaml", StartLine: 50, EndLine: 50},
		Origin:  domain.OriginExtracted, Status: domain.StatusCurrent,
	}}}); err != nil {
		t.Fatal(err)
	}
	tour, err := slices.NewEngine(m).GetTour("service:checkout", 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range tour.Steps {
		if s.NodeIDs[0] == "service:fraud-detection" && s.WhyLocator != "demo@abc1234:compose.full.yaml#L66-L66" {
			t.Errorf("fraud-detection why_locator %q spans sidecar's declaration, want L66 alone", s.WhyLocator)
		}
	}
}
