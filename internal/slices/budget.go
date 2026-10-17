package slices

import (
	"errors"
	"fmt"
	"sort"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// Default budgets in estimated tokens (ADR 0006 tiers; spec A.1 for the briefing).
const (
	BudgetSummary   = 500
	BudgetContracts = 1500
	BudgetDeep      = 3500
	BudgetBriefing  = 1500
)

// DefaultBudget returns the ADR 0006 budget of a tier.
func DefaultBudget(tier SliceTier) int {
	switch tier {
	case TierContracts:
		return BudgetContracts
	case TierDeep:
		return BudgetDeep
	default:
		return BudgetSummary
	}
}

// ErrBudgetExceeded is returned when even the parts that are never dropped (the
// focal node or target, unsupported boundaries and the budget marker) exceed the
// budget. The gira-query error code for it is "budget_exceeded".
var ErrBudgetExceeded = errors.New("budget_exceeded")

// Budget reports how an output was fitted to a hard token ceiling. Estimated is the
// tokens.Estimate of the exact bytes emitted, this object included.
type Budget struct {
	Limit        int     `json:"limit"`
	Estimated    int     `json:"estimated"`
	Truncated    bool    `json:"truncated"`
	Dropped      Dropped `json:"dropped"`
	EscalateHint string  `json:"escalate_hint,omitempty"`
	Estimator    string  `json:"estimator,omitempty"`
	Measured     bool    `json:"measured"`
}

// Dropped counts what truncation removed.
type Dropped struct {
	Nodes         int `json:"nodes"`
	Relationships int `json:"relationships"`
	Evidence      int `json:"evidence"`
}

// Format is an output format that budgets are measured on.
type Format string

const (
	FormatJSON     Format = "json"
	FormatMarkdown Format = "md"
	FormatHTML     Format = "html"
)

// ParseFormat maps a CLI format flag to a Format.
func ParseFormat(s string) Format {
	switch s {
	case "md", "markdown":
		return FormatMarkdown
	case "html":
		return FormatHTML
	default:
		return FormatJSON
	}
}

// Emit renders a slice in format exactly as the CLI writes it. A limit > 0 fits it
// to that hard ceiling first; limit <= 0 renders it unchanged, with no budget.
func (cs *ContextSlice) Emit(format Format, limit int) ([]byte, error) {
	render := func(c *ContextSlice) ([]byte, error) {
		if format == FormatMarkdown {
			return []byte(c.RenderMarkdown() + "\n"), nil
		}
		return c.CanonicalJSON()
	}
	if limit <= 0 {
		return render(cs)
	}
	nodes := make([]string, len(cs.Neighbors))
	for i, n := range cs.Neighbors {
		nodes[i] = n.ID
	}
	order := dropOrder(nodes, map[string]int{}) // every neighbor is one hop away
	return fit(limit, order, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		var d Dropped
		c := *cs
		c.FocalNode = trimmedNode(cs.FocalNode, trimEvidence, &d)
		c.Neighbors = nil
		for _, n := range cs.Neighbors {
			if drop[n.ID] {
				d.Nodes++
				continue
			}
			c.Neighbors = append(c.Neighbors, trimmedNode(n, trimEvidence, &d))
		}
		c.Relationships = keptRels(cs.Relationships, trimEvidence, drop, &d)
		return func(b *Budget) ([]byte, error) {
			b.Estimator = "bytes/4"
			b.Measured = false
			c.Budget = b
			return render(&c)
		}, d
	})
}

// Emit renders a briefing in format exactly as the CLI writes it, fitted to limit
// when limit > 0. Nodes are dropped farthest from the target first (hops over the
// briefing's mediation edges), then by descending ID.
func (ib *ImpactBriefing) Emit(format Format, limit int) ([]byte, error) {
	render := func(b *ImpactBriefing) ([]byte, error) {
		if format == FormatMarkdown {
			return []byte(b.RenderMarkdown() + "\n"), nil
		}
		return b.CanonicalJSON()
	}
	if limit <= 0 {
		return render(ib)
	}
	var nodes, libraries []string
	for _, n := range ib.UpstreamLibraries {
		libraries = append(libraries, n.ID)
	}
	for _, group := range [][]domain.Node{ib.UpstreamServices, ib.DownstreamServices, ib.AffectedContracts} {
		for _, n := range group {
			nodes = append(nodes, n.ID)
		}
	}
	roots := []string{ib.TargetID}
	if ib.Diff != nil {
		roots = ib.Diff.Targets
	}
	// The target (and, in diff mode, every diff target) is the focal node: it is
	// never dropped, even when it also appears in affected_contracts.
	focal := map[string]bool{}
	for _, r := range roots {
		focal[r] = true
	}
	kept := nodes[:0]
	for _, id := range nodes {
		if !focal[id] {
			kept = append(kept, id)
		}
	}
	// Libraries go first, before any Service or Contract.
	order := append(dropOrder(libraries, nil), dropOrder(kept, hops(roots, ib.MediationEdges))...)
	return fit(limit, order, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		var d Dropped
		c := *ib
		keep := func(group []domain.Node) []domain.Node {
			out := make([]domain.Node, 0, len(group))
			for _, n := range group {
				if drop[n.ID] {
					d.Nodes++
					continue
				}
				out = append(out, trimmedNode(n, trimEvidence, &d))
			}
			return out
		}
		c.UpstreamServices = keep(ib.UpstreamServices)
		if len(ib.UpstreamLibraries) > 0 {
			c.UpstreamLibraries = keep(ib.UpstreamLibraries)
		}
		c.DownstreamServices = keep(ib.DownstreamServices)
		c.AffectedContracts = keep(ib.AffectedContracts)
		c.MediationEdges = keptRels(ib.MediationEdges, trimEvidence, drop, &d)
		return func(b *Budget) ([]byte, error) {
			b.Estimator = "bytes/4"
			b.Measured = false
			c.Budget = b
			return render(&c)
		}, d
	})
}

// candidate applies truncation step 1 when trimEvidence is set and step 2 to the
// nodes in drop (step 3 follows from it). It returns the renderer of the result
// and what was removed.
type candidate func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped)

// fit applies the documented truncation order until the output fits limit:
//  1. evidence records beyond the first per fact;
//  2. nodes in dropOrder, as few as possible (binary search; output size is
//     monotone in the number dropped);
//  3. relationships to dropped nodes, together with them.
//
// The focal node or target, unsupported boundaries and the budget marker are
// never dropped. If they alone exceed limit, fit returns ErrBudgetExceeded.
func fit(limit int, order []string, build candidate) ([]byte, error) {
	render, _ := build(false, nil)
	hint := func(est int) string {
		return fmt.Sprintf("truncated to fit %d tokens; the full output needs about %d: rerun with budget >= %d", limit, est, est)
	}
	fullOut, fullEst, err := settleHint(limit, hint, render)
	if err != nil {
		return nil, err
	}
	if fullEst <= limit {
		// Even when no truncation is needed, include the budget with escalate_hint
		// so clients can detect the ceiling and plan accordingly. The estimate was
		// settled with the hint included, so Estimated matches the emitted bytes.
		return fullOut, nil
	}
	fullHint := hint(fullEst)

	try := func(k int) ([]byte, bool, error) {
		drop := make(map[string]bool, k)
		for _, id := range order[:k] {
			drop[id] = true
		}
		render, dropped := build(true, drop)
		out, est, err := settle(&Budget{Limit: limit, Truncated: dropped != Dropped{}, Dropped: dropped, EscalateHint: fullHint}, render)
		return out, est <= limit, err
	}

	out, ok, err := try(0)
	if err != nil || ok {
		return out, err
	}
	if _, ok, err = try(len(order)); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("%w: the parts that are never dropped need more than %d tokens; rerun with budget >= %d", ErrBudgetExceeded, limit, limit)
	}
	lo, hi := 1, len(order)
	for lo < hi {
		mid := (lo + hi) / 2
		if _, ok, err := try(mid); err != nil {
			return nil, err
		} else if ok {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	out, ok, err = try(lo)
	if err == nil && !ok {
		// Never trust monotonicity over the ceiling: dropping every node fits.
		out, _, err = try(len(order))
	}
	return out, err
}

// settleHint settles Estimated against the bytes that carry it, recomputing the
// escalate hint each pass because the hint text embeds the estimate itself.
// Without it a no-truncation budget would omit the hint's own bytes from
// Estimated, so a client's Estimate(output) would exceed the reported figure.
func settleHint(limit int, hint func(int) string, render func(*Budget) ([]byte, error)) ([]byte, int, error) {
	b := &Budget{Limit: limit}
	for i := 0; i < 8; i++ {
		b.EscalateHint = hint(b.Estimated)
		out, est, err := settle(b, render)
		if err != nil {
			return nil, 0, err
		}
		if hint(est) == b.EscalateHint {
			return out, est, nil
		}
		b.Estimated = est
	}
	return nil, 0, fmt.Errorf("budget estimate did not settle")
}

// settle renders until b.Estimated equals the estimate of the bytes that carry it.
func settle(b *Budget, render func(*Budget) ([]byte, error)) ([]byte, int, error) {
	for i := 0; i < 8; i++ {
		out, err := render(b)
		if err != nil {
			return nil, 0, err
		}
		got := tokens.Estimate(out)
		if got == b.Estimated {
			return out, got, nil
		}
		b.Estimated = got
	}
	return nil, 0, fmt.Errorf("budget estimate did not settle")
}

func trimmedNode(n domain.Node, trim bool, d *Dropped) domain.Node {
	if trim && len(n.Evidence) > 1 {
		d.Evidence += len(n.Evidence) - 1
		n.Evidence = n.Evidence[:1]
	}
	return n
}

func keptRels(rels []domain.Relationship, trim bool, drop map[string]bool, d *Dropped) []domain.Relationship {
	out := make([]domain.Relationship, 0, len(rels))
	for _, r := range rels {
		if drop[r.SourceID] || drop[r.TargetID] {
			d.Relationships++
			continue
		}
		if trim && len(r.Evidence) > 1 {
			d.Evidence += len(r.Evidence) - 1
			r.Evidence = r.Evidence[:1]
		}
		out = append(out, r)
	}
	return out
}

// dropOrder sorts ids so the first dropped is the farthest (missing distance means
// unreachable, farthest of all), ties broken by descending ID. The nodes kept are
// therefore always the nearest, lowest IDs.
func dropOrder(ids []string, dist map[string]int) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	d := func(id string) int {
		if v, ok := dist[id]; ok {
			return v
		}
		if len(dist) == 0 {
			return 1
		}
		return int(^uint(0) >> 1)
	}
	sort.Slice(out, func(i, j int) bool {
		if d(out[i]) != d(out[j]) {
			return d(out[i]) > d(out[j])
		}
		return out[i] > out[j]
	})
	return out
}

// hops returns undirected hop counts from the nearest root over rels.
func hops(roots []string, rels []domain.Relationship) map[string]int {
	adj := map[string][]string{}
	for _, r := range rels {
		adj[r.SourceID] = append(adj[r.SourceID], r.TargetID)
		adj[r.TargetID] = append(adj[r.TargetID], r.SourceID)
	}
	dist := map[string]int{}
	var queue []string
	for _, r := range roots {
		dist[r] = 0
		queue = append(queue, r)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range adj[cur] {
			if _, ok := dist[nb]; !ok {
				dist[nb] = dist[cur] + 1
				queue = append(queue, nb)
			}
		}
	}
	return dist
}
