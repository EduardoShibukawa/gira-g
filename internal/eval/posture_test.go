package eval_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// postureWatch names an id whose evidence lives in code no Wave 1 Extractor
// parses, and the corpus directory holding that code, per variant.
type postureWatch struct {
	id     map[string]string // variant → expected id
	needle map[string]string // variant → corpus dir of the unsupported call-site
	// reason, when set, is the text the unknown row's reason must contain to
	// name the id: for a construct whose row cites no Contract (OpAMP).
	reason string
}

// SC3 watch list (spec gira-recall-joins J3): billing, exchange, storefront,
// opamp-server and ExchangeService, under their published names too.
var postureWatches = []postureWatch{
	{map[string]string{"published": "service:payment", "perturbed": "service:billing"},
		map[string]string{"published": "src/payment", "perturbed": "src/billing"}, ""},
	{map[string]string{"published": "service:currency", "perturbed": "service:exchange"},
		map[string]string{"published": "src/currency", "perturbed": "src/exchange"}, ""},
	{map[string]string{"published": "service:frontend", "perturbed": "service:storefront"},
		map[string]string{"published": "src/frontend", "perturbed": "src/storefront"}, ""},
	{map[string]string{"published": "service:opamp-server", "perturbed": "service:opamp-server"},
		map[string]string{"published": "src/product-catalog", "perturbed": "src/inventory"}, "OpAMP client"},
	{map[string]string{"published": "contract:CurrencyService", "perturbed": "contract:ExchangeService"},
		map[string]string{"published": "src/currency", "perturbed": "src/exchange"}, ""},
}

// SC3: every expected id of the watch list is either joined (present in the
// briefing, with a locator) or explicitly unknown (an unsupported row in the
// briefing whose locator resolves to a corpus file under the id's unsupported
// directory and which cites the id). A watched id that is missing with no
// such row is a silent miss. The ground truth is read as is; it is never
// retargeted to make this pass.
//
// Citing (follow-up F6): the row's unknown_joins contains the id itself, or,
// for a Service (whose row lives in its own code), a Contract the task
// expects: the join its unsupported call-site leaves unknown. A construct
// that names no Contract cites the id through its reason instead. A row that
// is merely under the directory, citing nothing the task observes, fails.
func TestSC3UnsupportedPosture(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		if v.gt.Reviewed {
			t.Fatalf("%s: ground truth must stay reviewed=false (owner's call)", v.name)
		}
		ev := eval.NewEvaluator(v.mm, v.gt.Ignore, nil, eval.DefaultThresholds())
		seen := map[string]int{}
		for _, task := range v.gt.Tasks {
			res := ev.EvaluateTask(context.Background(), task)
			missing := map[string]bool{}
			for _, id := range res.MissingIDs {
				missing[id] = true
			}
			for _, w := range postureWatches {
				id, needle := w.id[v.name], w.needle[v.name]
				if !expects(task.GroundTruth, id) {
					continue
				}
				seen[id]++
				if res.Briefing == nil {
					t.Errorf("%s %s: no briefing, %s silently missing: %s", v.name, task.ID, id, res.Details)
					continue
				}
				if !missing[id] {
					if !joinedWithLocator(res.Briefing, id) {
						t.Errorf("%s %s: %s joined without a locator", v.name, task.ID, id)
					}
					continue
				}
				if row, ok := unknownRow(res.Briefing, v.dir, needle, citer(w, id, task.GroundTruth)); ok {
					t.Logf("%s %s: %s unknown at %s (%s) citing %v", v.name, task.ID, id, row.Locator.Path, row.Reason, row.UnknownJoins)
					continue
				}
				t.Errorf("%s %s: %s silently missing: no joined id and no unknown row under %s citing it (briefing rows: %v)",
					v.name, task.ID, id, needle, rowCitations(res.Briefing))
			}
		}
		for _, w := range postureWatches {
			if seen[w.id[v.name]] == 0 {
				t.Errorf("%s: watched %s is expected by no task; the law would pass vacuously", v.name, w.id[v.name])
			}
		}
	}
}

func expects(g eval.GroundTruthImpact, id string) bool {
	for _, set := range [][]string{g.ExpectedUpstreamServices, g.ExpectedDownstreamServices, g.ExpectedAffectedContracts} {
		for _, e := range set {
			if e == id {
				return true
			}
		}
	}
	return false
}

func joinedWithLocator(b *slices.ImpactBriefing, id string) bool {
	for _, set := range [][]domain.Node{b.UpstreamServices, b.DownstreamServices, b.AffectedContracts} {
		for _, n := range set {
			if n.ID != id {
				continue
			}
			for _, e := range n.Evidence {
				if e.Locator.Path != "" {
					return true
				}
			}
		}
	}
	return false
}

// citer returns the predicate an unknown row must meet to cite id in a task
// whose ground truth is g (see TestSC3UnsupportedPosture).
func citer(w postureWatch, id string, g eval.GroundTruthImpact) func(domain.UnsupportedBoundary) bool {
	return func(ub domain.UnsupportedBoundary) bool {
		if w.reason != "" {
			return strings.Contains(ub.Reason, w.reason)
		}
		for _, j := range ub.UnknownJoins {
			if j == id {
				return true
			}
			if !strings.HasPrefix(id, "service:") {
				continue
			}
			for _, c := range g.ExpectedAffectedContracts {
				if j == c {
					return true
				}
			}
		}
		return false
	}
}

// unknownRow finds an unsupported row whose locator is a real corpus file
// under dir and which cites the watched id.
func unknownRow(b *slices.ImpactBriefing, corpusDir, dir string, cites func(domain.UnsupportedBoundary) bool) (domain.UnsupportedBoundary, bool) {
	for _, ub := range b.UnsupportedBoundaries {
		p := path.Clean(ub.Locator.Path)
		if ub.Reason == "" || !strings.HasPrefix(p, dir+"/") || !cites(ub) {
			continue
		}
		if fi, err := os.Stat(filepath.Join(corpusDir, filepath.FromSlash(p))); err == nil && fi.Mode().IsRegular() {
			return ub, true
		}
	}
	return domain.UnsupportedBoundary{}, false
}

func rowCitations(b *slices.ImpactBriefing) []string {
	out := make([]string, 0, len(b.UnsupportedBoundaries))
	for _, ub := range b.UnsupportedBoundaries {
		out = append(out, fmt.Sprintf("%s%v", ub.Locator.Path, ub.UnknownJoins))
	}
	return out
}
