package extractors

import (
	"fmt"
	"sort"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func corrobEv(path string, line int) domain.EvidenceRecord {
	return domain.EvidenceRecord{
		Locator: domain.Locator{Repo: "r", CommitSHA: "c", Path: path, StartLine: line},
		Origin:  domain.OriginExtracted,
	}
}

func corrobRel(src string, verb domain.RelationVerb, dst string, ev ...domain.EvidenceRecord) *domain.Relationship {
	return &domain.Relationship{ID: relID(src, verb, dst), SourceID: src, TargetID: dst, Verb: verb, Evidence: ev}
}

func corrobDependsOn(svc, target string, line int) Hint {
	return Hint{
		Kind:    HintDependsOn,
		Locator: domain.Locator{Repo: "r", CommitSHA: "c", Path: "compose.yaml", StartLine: line},
		Service: svc,
		Value:   target,
	}
}

func citedLines(rel *domain.Relationship) []string {
	var out []string
	for _, e := range rel.Evidence {
		out = append(out, fmt.Sprintf("%s#L%d", e.Locator.Path, e.Locator.StartLine))
	}
	sort.Strings(out)
	return out
}

// F5 (spec J2): corroboration of a depends_on-only edge also accepts a
// contract reference — the consumer consumes a Contract the target
// provides — and an accesses edge between the same pair. A kept edge cites
// both the depends_on line and the corroborating line. An edge with no
// such corroboration (contract provided by a third service, or an accesses
// edge resting on the same depends_on line) is still pruned.
func TestDependsOnCorroboratedByContractOrAccesses(t *testing.T) {
	rels := []*domain.Relationship{
		// Contract reference: a consumes contract:Pay, b provides it.
		corrobRel("service:a", domain.VerbConsumes, "service:b", corrobEv("compose.yaml", 4)),
		corrobRel("service:a", domain.VerbConsumes, "contract:Pay", corrobEv("src/a/main.go", 10)),
		corrobRel("service:b", domain.VerbProvides, "contract:Pay", corrobEv("pb/pay.proto", 3)),
		// accesses corroboration between the same pair.
		corrobRel("service:c", domain.VerbConsumes, "service:d", corrobEv("compose.yaml", 8)),
		corrobRel("service:c", domain.VerbAccesses, "service:d", corrobEv("compose.yaml", 12)),
		// Uncorroborated: e consumes contract:Ship, but g (not f) provides it;
		// e accesses f only through the same depends_on line.
		corrobRel("service:e", domain.VerbConsumes, "service:f", corrobEv("compose.yaml", 15)),
		corrobRel("service:e", domain.VerbConsumes, "contract:Ship", corrobEv("src/e/main.go", 7)),
		corrobRel("service:g", domain.VerbProvides, "contract:Ship", corrobEv("pb/ship.proto", 2)),
		corrobRel("service:e", domain.VerbAccesses, "service:f", corrobEv("compose.yaml", 15)),
	}
	r := &resolution{nodes: map[string]*domain.Node{}, rels: map[string]*domain.Relationship{}}
	for _, rel := range rels {
		r.rels[rel.ID] = rel
	}
	r.hints = []Hint{
		corrobDependsOn("service:a", "b", 4),
		corrobDependsOn("service:c", "d", 8),
		corrobDependsOn("service:e", "f", 15),
	}
	pruneOrderingOnlyEdges(r)

	byPair := map[string]*domain.Relationship{}
	for _, rel := range r.rels {
		if rel.Verb == domain.VerbConsumes {
			byPair[rel.SourceID+">"+rel.TargetID] = rel
		}
	}
	want := map[string]string{
		"service:a>service:b": "[compose.yaml#L4 src/a/main.go#L10]",
		"service:c>service:d": "[compose.yaml#L12 compose.yaml#L8]",
	}
	for pair, w := range want {
		rel, ok := byPair[pair]
		if !ok {
			t.Errorf("corroborated %s was pruned", pair)
			continue
		}
		if got := fmt.Sprint(citedLines(rel)); got != w {
			t.Errorf("%s cites %s, want both lines %s", pair, got, w)
		}
	}
	if rel, ok := byPair["service:e>service:f"]; ok {
		t.Errorf("uncorroborated ordering-only e->f survived citing %v", citedLines(rel))
	}
	// The corroborating edges themselves are untouched.
	for _, id := range []string{
		relID("service:a", domain.VerbConsumes, "contract:Pay"),
		relID("service:c", domain.VerbAccesses, "service:d"),
		relID("service:e", domain.VerbAccesses, "service:f"),
	} {
		if _, ok := r.rels[id]; !ok {
			t.Errorf("%s must not be pruned", id)
		}
	}
}
