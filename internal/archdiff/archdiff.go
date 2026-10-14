// Package archdiff compares two indexes of a Target System (spec A.2,
// gira-arch-diff): nodes and Contracts added, removed, renamed or changed, edges
// added or removed, evidence that moved, and unsupported paths that appeared or
// went away. It works on the indexes alone. Renames are detected from structure,
// never from a declared table, so that a diff of two real releases needs no input
// besides the two indexes.
package archdiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// Rename pairs a node of the first index with its counterpart in the second.
// Origin is always "inferred": a rename is a structural match, not an extracted
// fact (spec rule I1).
type Rename struct {
	From   string          `json:"from"`
	To     string          `json:"to"`
	Kind   domain.NodeType `json:"kind"`
	Score  float64         `json:"score"`
	Origin string          `json:"origin"`
}

// Edge is a relationship by its endpoints and verb.
// Locator cites the edge's first evidence in its own index, or is "unknown".
type Edge struct {
	Source  string              `json:"source"`
	Verb    domain.RelationVerb `json:"verb"`
	Target  string              `json:"target"`
	Locator string              `json:"locator,omitempty"`
}

func (e Edge) String() string { return fmt.Sprintf("%s %s %s", e.Source, e.Verb, e.Target) }

func firstLocator(evs []domain.EvidenceRecord) string {
	if len(evs) == 0 {
		return "unknown"
	}
	return evs[0].Locator.String()
}

// Diff is the result of comparing index A (before) with index B (after). Removed
// items use A's IDs and added items use B's IDs.
type Diff struct {
	From                string   `json:"from"`
	To                  string   `json:"to"`
	Renamed             []Rename `json:"renamed"`
	AddedNodes          []string `json:"added_nodes"`
	RemovedNodes        []string `json:"removed_nodes"`
	AddedEdges          []Edge   `json:"added_edges"`
	RemovedEdges        []Edge   `json:"removed_edges"`
	ChangedContracts    []string `json:"changed_contracts"`
	NewlyUnsupported    []string `json:"newly_unsupported"`
	NoLongerUnsupported []string `json:"no_longer_unsupported"`
	// EvidenceMoved lists B node IDs whose evidence locators (path and lines,
	// commit ignored) differ. It is informational: moved evidence is not an
	// architectural change.
	EvidenceMoved []string `json:"evidence_moved"`
	// TokenRenames are the identifier spellings derived from Renamed, used to map
	// paths and Contract metadata between the indexes.
	TokenRenames map[string]string `json:"token_renames,omitempty"`
	// Locators cites, for every node id the diff reports (renamed, added, removed,
	// changed), the node's first evidence in its own index, or "unknown".
	Locators map[string]string `json:"locators"`
}

// Structural reports whether the diff has any architectural change: renames,
// added or removed nodes and edges, changed Contracts or unsupported paths.
func (d *Diff) Structural() bool {
	return len(d.Renamed)+len(d.AddedNodes)+len(d.RemovedNodes)+len(d.AddedEdges)+len(d.RemovedEdges)+
		len(d.ChangedContracts)+len(d.NewlyUnsupported)+len(d.NoLongerUnsupported) > 0
}

// Rename pairs are scored as
//
//	jaccard(neighbours) + 2·jaccard(stable evidence) + 0.5·jaccard(loose evidence)
//
// Neighbours are (verb, direction, neighbour) with already-matched neighbours
// mapped into B's names. Stable evidence is path and line range in files whose
// path exists in both indexes; loose evidence is base name and line range in the
// other files, a weak signal because their lines may shift. A pair is accepted
// when each side is the other's unique best candidate, with a score of at least
// minScore; matching repeats until no pair is added.
const minScore = 0.5

// Compare diffs index a against index b.
func Compare(a, b *domain.MacroMap) *Diff {
	an, bn := byID(a), byID(b)
	rename := map[string]string{} // A id -> B id, renamed nodes only

	var removed, added []string
	for id := range an {
		if _, ok := bn[id]; !ok {
			removed = append(removed, id)
		}
	}
	for id := range bn {
		if _, ok := an[id]; !ok {
			added = append(added, id)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)

	adjA, adjB := adjacency(a), adjacency(b)
	stable := stablePaths(a, b)
	toB := func(id string) string {
		if r, ok := rename[id]; ok {
			return r
		}
		return id
	}
	scores := map[[2]string]float64{}
	for {
		takenB := map[string]bool{}
		for _, to := range rename {
			takenB[to] = true
		}
		type cand struct {
			r, d  string
			score float64
		}
		var cands []cand
		for _, r := range removed {
			if _, done := rename[r]; done {
				continue
			}
			nr := mapped(adjA[r], toB)
			for _, d := range added {
				if takenB[d] || an[r].Kind != bn[d].Kind {
					continue
				}
				fr, fd := footprint(an[r], stable), footprint(bn[d], stable)
				s := jaccard(nr, adjB[d]) + 2*jaccard(fr.stable, fd.stable) + 0.5*jaccard(fr.loose, fd.loose)
				if s >= minScore {
					cands = append(cands, cand{r, d, s})
				}
			}
		}
		// Accept best mutual pairs only: r's best is d and d's best is r, both unique.
		bestR, bestD := map[string]cand{}, map[string]cand{}
		tieR, tieD := map[string]bool{}, map[string]bool{}
		for _, c := range cands {
			if b, ok := bestR[c.r]; !ok || c.score > b.score {
				bestR[c.r], tieR[c.r] = c, false
			} else if c.score == b.score {
				tieR[c.r] = true
			}
			if b, ok := bestD[c.d]; !ok || c.score > b.score {
				bestD[c.d], tieD[c.d] = c, false
			} else if c.score == b.score {
				tieD[c.d] = true
			}
		}
		progress := false
		for r, c := range bestR {
			if tieR[r] || tieD[c.d] || bestD[c.d].r != r {
				continue
			}
			rename[r] = c.d
			scores[[2]string{r, c.d}] = c.score
			progress = true
		}
		if !progress {
			break
		}
	}

	d := &Diff{From: a.TargetSystem, To: b.TargetSystem}
	d.Renamed = []Rename{}
	for r, to := range rename {
		d.Renamed = append(d.Renamed, Rename{From: r, To: to, Kind: an[r].Kind, Score: round(scores[[2]string{r, to}]), Origin: "inferred"})
	}
	sort.Slice(d.Renamed, func(i, j int) bool { return d.Renamed[i].From < d.Renamed[j].From })
	d.RemovedNodes = []string{}
	for _, r := range removed {
		if _, ok := rename[r]; !ok {
			d.RemovedNodes = append(d.RemovedNodes, r)
		}
	}
	matchedB := map[string]bool{}
	for _, to := range rename {
		matchedB[to] = true
	}
	d.AddedNodes = []string{}
	for _, id := range added {
		if !matchedB[id] {
			d.AddedNodes = append(d.AddedNodes, id)
		}
	}

	tokens := tokenRenames(d.Renamed)
	if len(tokens) > 0 {
		d.TokenRenames = tokens
	}
	tr := replacer(tokens)

	// Edges, compared in B's namespace; removed edges are reported with A's IDs.
	inB := map[Edge]bool{}
	for _, r := range b.Relationships {
		inB[Edge{r.SourceID, r.Verb, r.TargetID, ""}] = true
	}
	inA := map[Edge]bool{}
	d.RemovedEdges = []Edge{}
	for _, r := range a.Relationships {
		e := Edge{toB(r.SourceID), r.Verb, toB(r.TargetID), ""}
		inA[e] = true
		if !inB[e] {
			d.RemovedEdges = append(d.RemovedEdges, Edge{r.SourceID, r.Verb, r.TargetID, firstLocator(r.Evidence)})
		}
	}
	d.AddedEdges = []Edge{}
	for _, r := range b.Relationships {
		if !inA[Edge{r.SourceID, r.Verb, r.TargetID, ""}] {
			d.AddedEdges = append(d.AddedEdges, Edge{r.SourceID, r.Verb, r.TargetID, firstLocator(r.Evidence)})
		}
	}
	sortEdges(d.RemovedEdges)
	sortEdges(d.AddedEdges)
	d.RemovedEdges = dedupEdges(d.RemovedEdges)
	d.AddedEdges = dedupEdges(d.AddedEdges)

	// Contracts present on both sides whose metadata differs once renames apply.
	d.ChangedContracts = []string{}
	d.EvidenceMoved = []string{}
	for id, n := range an {
		m, ok := bn[toB(id)]
		if !ok {
			continue
		}
		if n.Kind == domain.NodeContract && tr.Replace(canonical(n.Metadata)) != canonical(m.Metadata) {
			d.ChangedContracts = append(d.ChangedContracts, m.ID)
		}
		if !sameEvidence(n.Evidence, m.Evidence, tr) {
			d.EvidenceMoved = append(d.EvidenceMoved, m.ID)
		}
	}
	sort.Strings(d.ChangedContracts)
	sort.Strings(d.EvidenceMoved)

	ua, ub := map[string]bool{}, map[string]bool{}
	for _, u := range a.Unsupported {
		ua[tr.Replace(u.Path)] = true
	}
	for _, u := range b.Unsupported {
		ub[u.Path] = true
	}
	d.NewlyUnsupported = []string{}
	for p := range ub {
		if !ua[p] {
			d.NewlyUnsupported = append(d.NewlyUnsupported, p)
		}
	}
	d.NoLongerUnsupported = []string{}
	for _, u := range a.Unsupported {
		if !ub[tr.Replace(u.Path)] {
			d.NoLongerUnsupported = append(d.NoLongerUnsupported, u.Path)
		}
	}
	sort.Strings(d.NewlyUnsupported)
	sort.Strings(d.NoLongerUnsupported)
	d.NoLongerUnsupported = dedup(d.NoLongerUnsupported)

	d.Locators = map[string]string{}
	for _, r := range d.Renamed {
		d.Locators[r.From] = firstLocator(an[r.From].Evidence)
		d.Locators[r.To] = firstLocator(bn[r.To].Evidence)
	}
	for _, id := range d.RemovedNodes {
		d.Locators[id] = firstLocator(an[id].Evidence)
	}
	for _, id := range append(append([]string{}, d.AddedNodes...), d.ChangedContracts...) {
		d.Locators[id] = firstLocator(bn[id].Evidence)
	}
	return d
}

func byID(m *domain.MacroMap) map[string]domain.Node {
	out := make(map[string]domain.Node, len(m.Nodes))
	for _, n := range m.Nodes {
		out[n.ID] = n
	}
	return out
}

type nb struct {
	key, id string // key = verb plus direction; id = the neighbour
}

func adjacency(m *domain.MacroMap) map[string][]nb {
	out := map[string][]nb{}
	for _, r := range m.Relationships {
		out[r.SourceID] = append(out[r.SourceID], nb{string(r.Verb) + ">", r.TargetID})
		out[r.TargetID] = append(out[r.TargetID], nb{string(r.Verb) + "<", r.SourceID})
	}
	return out
}

func mapped(ns []nb, f func(string) string) []nb {
	out := make([]nb, len(ns))
	for i, n := range ns {
		out[i] = nb{n.key, f(n.id)}
	}
	return out
}

func jaccard[T comparable](a, b []T) float64 {
	sa, sb := map[T]bool{}, map[T]bool{}
	for _, x := range a {
		sa[x] = true
	}
	for _, x := range b {
		sb[x] = true
	}
	if len(sa)+len(sb) == 0 {
		return 0
	}
	inter := 0
	for x := range sa {
		if sb[x] {
			inter++
		}
	}
	return float64(inter) / float64(len(sa)+len(sb)-inter)
}

type prints struct{ stable, loose []string }

// footprint is a node's evidence without its name (see the scoring rule above).
func footprint(n domain.Node, stable map[string]bool) prints {
	var p prints
	for _, ev := range n.Evidence {
		l := ev.Locator
		if stable[l.Path] {
			p.stable = append(p.stable, fmt.Sprintf("%s#%d-%d", l.Path, l.StartLine, l.EndLine))
		} else {
			p.loose = append(p.loose, fmt.Sprintf("%s#%d-%d", path.Base(l.Path), l.StartLine, l.EndLine))
		}
	}
	return p
}

// stablePaths are the evidence paths cited in both indexes.
func stablePaths(a, b *domain.MacroMap) map[string]bool {
	paths := func(m *domain.MacroMap) map[string]bool {
		out := map[string]bool{}
		for _, n := range m.Nodes {
			for _, ev := range n.Evidence {
				out[ev.Locator.Path] = true
			}
		}
		for _, r := range m.Relationships {
			for _, ev := range r.Evidence {
				out[ev.Locator.Path] = true
			}
		}
		return out
	}
	pa, pb := paths(a), paths(b)
	out := map[string]bool{}
	for p := range pa {
		if pb[p] {
			out[p] = true
		}
	}
	return out
}

func sameEvidence(a, b []domain.EvidenceRecord, tr *strings.Replacer) bool {
	key := func(ev domain.EvidenceRecord, f func(string) string) string {
		return fmt.Sprintf("%s#%d-%d", f(ev.Locator.Path), ev.Locator.StartLine, ev.Locator.EndLine)
	}
	ka, kb := make([]string, len(a)), make([]string, len(b))
	for i, ev := range a {
		ka[i] = key(ev, tr.Replace)
	}
	for i, ev := range b {
		kb[i] = key(ev, func(s string) string { return s })
	}
	sort.Strings(ka)
	sort.Strings(kb)
	return strings.Join(ka, "\n") == strings.Join(kb, "\n")
}

var tokenSplit = regexp.MustCompile(`[A-Z]+[a-z0-9]*|[a-z0-9]+`)

// tokenRenames derives identifier spellings from renamed IDs: the IDs are split
// into words, the common leading and trailing words are dropped, and what is left
// on each side is one rename, in the spelling used by the ID (kebab-case words are
// rejoined with hyphens, CamelCase words without a separator).
func tokenRenames(rs []Rename) map[string]string {
	out := map[string]string{}
	for _, r := range rs {
		pair := segments(r.From, r.To)
		if pair[0] == "" || pair[1] == "" || pair[0] == pair[1] {
			continue
		}
		// Two renames disagreeing on a spelling keep the smaller target, so the
		// map does not depend on iteration order.
		if prev, ok := out[pair[0]]; !ok || pair[1] < prev {
			out[pair[0]] = pair[1]
		}
	}
	return out
}

// segments returns the differing middle of two IDs, in their own spelling.
func segments(a, b string) [2]string {
	ta, tb := tokenSplit.FindAllStringIndex(a, -1), tokenSplit.FindAllStringIndex(b, -1)
	word := func(s string, ix []int) string { return s[ix[0]:ix[1]] }
	i := 0
	for i < len(ta) && i < len(tb) && word(a, ta[i]) == word(b, tb[i]) {
		i++
	}
	j := 0
	for j < len(ta)-i && j < len(tb)-i && word(a, ta[len(ta)-1-j]) == word(b, tb[len(tb)-1-j]) {
		j++
	}
	if i == len(ta)-j || i == len(tb)-j {
		return [2]string{}
	}
	return [2]string{a[ta[i][0]:ta[len(ta)-1-j][1]], b[tb[i][0]:tb[len(tb)-1-j][1]]}
}

func replacer(m map[string]string) *strings.Replacer {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	pairs := make([]string, 0, 2*len(keys))
	for _, k := range keys {
		pairs = append(pairs, k, m[k])
	}
	return strings.NewReplacer(pairs...)
}

func canonical(v any) string {
	data, _ := json.Marshal(v) // map keys are sorted by encoding/json
	return string(data)
}

func round(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }

func sortEdges(es []Edge) {
	sort.Slice(es, func(i, j int) bool { return es[i].String() < es[j].String() })
}

func dedupEdges(es []Edge) []Edge {
	out := es[:0]
	for i, e := range es {
		if i == 0 || e != es[i-1] {
			out = append(out, e)
		}
	}
	return out
}

func dedup(xs []string) []string {
	out := xs[:0]
	for i, x := range xs {
		if i == 0 || x != xs[i-1] {
			out = append(out, x)
		}
	}
	return out
}

// CanonicalJSON renders the diff as indented, byte-deterministic JSON.
func (d *Diff) CanonicalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMarkdown renders the diff as a human changelog.
func (d *Diff) RenderMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Architecture diff: %s → %s\n", d.From, d.To))
	if !d.Structural() {
		sb.WriteString("\nNo architectural changes.\n")
	}
	section := func(title string, items []string) {
		if len(items) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("\n## %s (%d)\n", title, len(items)))
		for _, it := range items {
			sb.WriteString("- " + it + "\n")
		}
	}
	var rn []string
	for _, r := range d.Renamed {
		rn = append(rn, fmt.Sprintf("`%s` → `%s` (%s, inferred from structure, score %.2f) [%s → %s]", r.From, r.To, r.Kind, r.Score, d.Locators[r.From], d.Locators[r.To]))
	}
	section("Renamed", rn)
	cite := func(ids []string) []string {
		out := quote(ids)
		for i, id := range ids {
			out[i] += " [" + d.Locators[id] + "]"
		}
		return out
	}
	section("Added nodes", cite(d.AddedNodes))
	section("Removed nodes", cite(d.RemovedNodes))
	edges := func(es []Edge) []string {
		var out []string
		for _, e := range es {
			out = append(out, fmt.Sprintf("`%s` %s `%s` [%s]", e.Source, e.Verb, e.Target, e.Locator))
		}
		return out
	}
	section("Added edges", edges(d.AddedEdges))
	section("Removed edges", edges(d.RemovedEdges))
	section("Changed contracts", cite(d.ChangedContracts))
	section("Newly unsupported", quote(d.NewlyUnsupported))
	section("No longer unsupported", quote(d.NoLongerUnsupported))
	if len(d.EvidenceMoved) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Evidence moved (%d, not an architectural change)\n", len(d.EvidenceMoved)))
		for _, id := range d.EvidenceMoved {
			sb.WriteString("- `" + id + "`\n")
		}
	}
	return sb.String()
}

func quote(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = "`" + x + "`"
	}
	return out
}
