package slices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// Tour is a deterministic reading roadmap (spec A.2, gira-tour). Steps are
// computed from the graph, never generated: the walk, the ordering and every
// `why` sentence come from fixed rules and templates.
type Tour struct {
	Focus string     `json:"focus,omitempty"`
	Entry []string   `json:"entry_points"`
	Steps []TourStep `json:"steps"`
	// Truncated is set when --max-steps cut the walk short.
	Truncated bool `json:"truncated,omitempty"`
}

// TourStep is one stop of a tour. Why is WhyTemplate rendered with WhyArgs.
type TourStep struct {
	Order       int      `json:"order"`
	Title       string   `json:"title"`
	NodeIDs     []string `json:"node_ids"`
	Read        []string `json:"read"`
	Why         string   `json:"why"`
	WhyTemplate string   `json:"why_template"`
	WhyArgs     []string `json:"why_args"`
	// WhyLocator cites the evidence of the claim in Why: the edge that reaches the
	// stop, or the focal node's own evidence. It is empty when the claim has no
	// locator, and Why then says so (rule I1).
	WhyLocator string `json:"why_locator,omitempty"`
}

// Why templates. RenderWhy is the only producer of TourStep.Why.
const (
	WhyFocus = "focus"
	WhyEntry = "entry"
	// WhyEntryDeclared is an entry point whose ports declaration has a locator.
	WhyEntryDeclared = "entry_declared"
	WhyReached       = "reached"
	whyUnknownT      = "unknown"
)

var whyTemplates = map[string]string{
	WhyFocus:         "`%s` is the focus of this tour.",
	WhyEntry:         "`%s` is an entry point: its extracted metadata declares %s (locator of the ports declaration: unknown).",
	WhyEntryDeclared: "`%s` is an entry point: it declares %s at the cited locator.",
	WhyReached:       "`%s` is reached because `%s` %s it.",
}

// RenderWhy renders a why template. Unknown templates render a marker, never text.
func RenderWhy(template string, args []string) string {
	t, ok := whyTemplates[template]
	if !ok {
		return whyUnknownT
	}
	vals := make([]any, len(args))
	for i, a := range args {
		vals[i] = a
	}
	return fmt.Sprintf(t, vals...)
}

// DefaultTourSteps caps a tour when --max-steps is not given.
const DefaultTourSteps = 20

// tourable reports whether a node can be a tour stop: runtime Components only.
// Libraries are code dependencies, and Services carrying a `stage` key are
// Dockerfile build stages, not running Components.
func tourable(n domain.Node) bool {
	switch n.Kind {
	case domain.NodeInfrastructure:
		return true
	case domain.NodeService:
		_, stage := n.Metadata["stage"]
		return !stage
	}
	return false
}

func ingress(n domain.Node) []string {
	raw, ok := n.Metadata["ports"]
	if !ok {
		return nil
	}
	var out []string
	switch v := raw.(type) {
	case []any:
		for _, p := range v {
			out = append(out, fmt.Sprint(p))
		}
	case []string:
		out = append(out, v...)
	}
	return out
}

// portsLocator cites where n declares its ports: the first Dockerfile EXPOSE
// entry, else the first compose ports entry, completed with the repo and commit
// of n's evidence for that file. It is "" when no declaration has a locator.
func portsLocator(n domain.Node) string {
	for _, key := range []string{"exposed_ports", "published_ports"} {
		var loc string
		switch entries := n.Metadata[key].(type) {
		case []map[string]any: // as extracted
			if len(entries) > 0 {
				loc, _ = entries[0]["locator"].(string)
			}
		case []any: // as read back from macro-map.json
			if len(entries) > 0 {
				if e, ok := entries[0].(map[string]any); ok {
					loc, _ = e["locator"].(string)
				}
			}
		}
		p, lines, ok := strings.Cut(loc, "#L")
		if !ok {
			continue
		}
		var start, end int
		if _, err := fmt.Sscanf(lines, "%d-L%d", &start, &end); err != nil {
			end = start
		}
		for _, ev := range n.Evidence {
			if ev.Locator.Path == p {
				return domain.Locator{Repo: ev.Locator.Repo, CommitSHA: ev.Locator.CommitSHA, Path: p, StartLine: start, EndLine: end}.String()
			}
		}
	}
	return ""
}

// walkVerbs are followed outward from a stop: what it consumes, provides and
// accesses. From a Contract the walk continues to its providers.
var walkVerbs = map[domain.RelationVerb]string{
	domain.VerbConsumes: "consumes",
	domain.VerbProvides: "provides",
	domain.VerbAccesses: "accesses",
}

// GetTour computes a tour. With focus, the walk starts there; otherwise it starts
// from every entry point (a tourable Service with declared ports). The walk is
// breadth-first; within a level, stops are ranked by fan-in (incoming consumes and
// accesses edges, descending), then by ID.
func (e *Engine) GetTour(focus string, maxSteps int) (*Tour, error) {
	if maxSteps <= 0 {
		maxSteps = DefaultTourSteps
	}
	fanIn := map[string]int{}
	for _, r := range e.macroMap.Relationships {
		if r.Verb == domain.VerbConsumes || r.Verb == domain.VerbAccesses {
			fanIn[r.TargetID]++
		}
	}
	rank := func(ids []string) {
		sort.Slice(ids, func(i, j int) bool {
			if fanIn[ids[i]] != fanIn[ids[j]] {
				return fanIn[ids[i]] > fanIn[ids[j]]
			}
			return ids[i] < ids[j]
		})
	}

	tour := &Tour{Focus: focus, Entry: []string{}, Steps: []TourStep{}}
	var level []string
	why := map[string]TourStep{}
	if focus != "" {
		n, ok := e.nodesByID[focus]
		if !ok {
			return nil, fmt.Errorf("node not found: %s", focus)
		}
		level = []string{n.ID}
		w := TourStep{WhyTemplate: WhyFocus, WhyArgs: []string{n.ID}}
		if len(n.Evidence) > 0 {
			w.WhyLocator = n.Evidence[0].Locator.String()
		}
		why[n.ID] = w
	} else {
		for _, n := range e.macroMap.Nodes {
			if n.Kind == domain.NodeService && tourable(n) && len(ingress(n)) > 0 {
				level = append(level, n.ID)
				w := TourStep{WhyTemplate: WhyEntry, WhyArgs: []string{n.ID, "ports " + strings.Join(ingress(n), ", ")}}
				if loc := portsLocator(n); loc != "" {
					w.WhyTemplate, w.WhyLocator = WhyEntryDeclared, loc
				}
				why[n.ID] = w
			}
		}
		rank(level)
		tour.Entry = append(tour.Entry, level...)
	}
	if focus != "" {
		tour.Entry = []string{focus}
	}

	visited := map[string]bool{}
	for _, id := range level {
		visited[id] = true
	}
	for len(level) > 0 {
		var next []string
		for _, id := range level {
			if len(tour.Steps) == maxSteps {
				tour.Truncated = true
				return tour, nil
			}
			step, reached := e.tourStep(id, why[id], visited)
			step.Order = len(tour.Steps) + 1
			tour.Steps = append(tour.Steps, step)
			for _, r := range reached {
				visited[r.id] = true
				why[r.id] = TourStep{WhyTemplate: WhyReached, WhyArgs: []string{r.id, id, r.verb}, WhyLocator: r.locator}
				next = append(next, r.id)
			}
		}
		rank(next)
		level = next
	}
	return tour, nil
}

type reach struct{ id, verb, locator string }

// tourStep builds the stop for id and returns the unvisited stops it reaches,
// either directly or through a Contract it consumes (Contract → its providers).
func (e *Engine) tourStep(id string, w TourStep, visited map[string]bool) (TourStep, []reach) {
	n := e.nodesByID[id]
	step := TourStep{
		Title:       fmt.Sprintf("%s (%s)", n.Name, n.Kind),
		NodeIDs:     []string{id},
		WhyTemplate: w.WhyTemplate,
		WhyArgs:     w.WhyArgs,
		WhyLocator:  w.WhyLocator,
	}
	step.Why = RenderWhy(step.WhyTemplate, step.WhyArgs)
	read := map[string]bool{}
	addRead := func(evs []domain.EvidenceRecord) {
		if len(evs) > 0 {
			read[evs[0].Locator.String()] = true
		}
	}
	addRead(n.Evidence)

	seen := map[string]bool{}
	var out []reach
	add := func(target, verb string, evs []domain.EvidenceRecord) {
		t, ok := e.nodesByID[target]
		if !ok || !tourable(t) || visited[target] || seen[target] || target == id {
			return
		}
		seen[target] = true
		addRead(evs)
		loc := ""
		if len(evs) > 0 {
			loc = evs[0].Locator.String()
		}
		out = append(out, reach{target, verb, loc})
	}
	contracts := map[string]bool{}
	for _, r := range e.edgesByNode[id] {
		verb, ok := walkVerbs[r.Verb]
		if !ok || r.SourceID != id {
			continue
		}
		t, ok := e.nodesByID[r.TargetID]
		if !ok {
			continue
		}
		if t.Kind == domain.NodeContract {
			contracts[t.ID] = true
			addRead(r.Evidence)
			for _, p := range e.providesByContract[t.ID] {
				provider := p.SourceID
				if provider == t.ID {
					provider = p.TargetID
				}
				if provider != id {
					add(provider, verb+" "+t.ID+" from", p.Evidence)
				}
			}
			continue
		}
		add(r.TargetID, verb, r.Evidence)
	}
	for c := range contracts {
		step.NodeIDs = append(step.NodeIDs, c)
	}
	sort.Strings(step.NodeIDs[1:])
	for l := range read {
		step.Read = append(step.Read, l)
	}
	sort.Strings(step.Read)
	if step.Read == nil {
		step.Read = []string{}
	}
	return step, out
}

// CanonicalJSON renders a tour as indented, byte-deterministic JSON.
func (t *Tour) CanonicalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMarkdown renders a tour as a numbered reading list.
func (t *Tour) RenderMarkdown() string {
	var sb strings.Builder
	if t.Focus != "" {
		sb.WriteString(fmt.Sprintf("# Tour: %s\n", t.Focus))
	} else {
		sb.WriteString("# Tour: entry points\n")
	}
	for _, s := range t.Steps {
		sb.WriteString(fmt.Sprintf("\n## %d. %s\n%s\n", s.Order, s.Title, s.Why))
		if s.WhyLocator != "" {
			sb.WriteString(fmt.Sprintf("- Why, evidence: %s\n", s.WhyLocator))
		}
		if len(s.NodeIDs) > 1 {
			sb.WriteString(fmt.Sprintf("- Contracts: `%s`\n", strings.Join(s.NodeIDs[1:], "`, `")))
		}
		for _, l := range s.Read {
			sb.WriteString(fmt.Sprintf("- Read: %s\n", l))
		}
	}
	if t.Truncated {
		sb.WriteString(fmt.Sprintf("\n> Tour stopped at %d steps; raise --max-steps to continue.\n", len(t.Steps)))
	}
	return sb.String()
}

// Flow is one deterministic chain from an entry through consumes edges to
// infrastructure (spec A.2, gira-explain-component). Business rules are not
// extracted; each hop only cites the evidence of the edge it follows.
type Flow struct {
	Steps []FlowHop `json:"steps"`
}

// FlowHop is one node of a flow and the edge that led to it.
type FlowHop struct {
	NodeID  string `json:"node_id"`
	Via     string `json:"via,omitempty"`
	Verb    string `json:"verb,omitempty"`
	Locator string `json:"locator,omitempty"`
}

// Flows is the result of `gira flows <entry>`.
type Flows struct {
	Entry     string `json:"entry"`
	Flows     []Flow `json:"flows"`
	Truncated bool   `json:"truncated,omitempty"`
}

const (
	maxFlowDepth = 6
	maxFlows     = 50
)

// GetFlows enumerates simple paths from entry: from a Service it follows its
// consumes and accesses edges, from a Contract its providers. A path ends at a
// stop with nowhere further to go (typically infrastructure) or at the depth
// limit. Paths are sorted and capped.
func (e *Engine) GetFlows(entry string) (*Flows, error) {
	if _, ok := e.nodesByID[entry]; !ok {
		return nil, fmt.Errorf("node not found: %s", entry)
	}
	type edge struct {
		to   string
		via  domain.Relationship
		verb string
	}
	next := func(id string) []edge {
		n := e.nodesByID[id]
		var out []edge
		if n.Kind == domain.NodeContract {
			for _, p := range e.providesByContract[id] {
				provider := p.SourceID
				if provider == id {
					provider = p.TargetID
				}
				out = append(out, edge{provider, p, "provided by"})
			}
		} else {
			for _, r := range e.edgesByNode[id] {
				if r.SourceID == id && (r.Verb == domain.VerbConsumes || r.Verb == domain.VerbAccesses) {
					out = append(out, edge{r.TargetID, r, string(r.Verb)})
				}
			}
		}
		kept := out[:0]
		for _, ed := range out {
			t, ok := e.nodesByID[ed.to]
			if ok && (t.Kind == domain.NodeContract || tourable(t)) {
				kept = append(kept, ed)
			}
		}
		sort.Slice(kept, func(i, j int) bool {
			if kept[i].to != kept[j].to {
				return kept[i].to < kept[j].to
			}
			return kept[i].via.ID < kept[j].via.ID
		})
		return kept
	}

	res := &Flows{Entry: entry, Flows: []Flow{}}
	path := []FlowHop{{NodeID: entry}}
	onPath := map[string]bool{entry: true}
	var walk func(id string)
	walk = func(id string) {
		if len(res.Flows) == maxFlows {
			res.Truncated = true
			return
		}
		var moves []edge
		if len(path) < maxFlowDepth {
			for _, ed := range next(id) {
				if !onPath[ed.to] {
					moves = append(moves, ed)
				}
			}
		}
		if len(moves) == 0 {
			if len(path) > 1 {
				res.Flows = append(res.Flows, Flow{Steps: append([]FlowHop(nil), path...)})
			}
			return
		}
		for _, ed := range moves {
			hop := FlowHop{NodeID: ed.to, Via: ed.via.ID, Verb: ed.verb}
			if len(ed.via.Evidence) > 0 {
				hop.Locator = ed.via.Evidence[0].Locator.String()
			}
			path = append(path, hop)
			onPath[ed.to] = true
			walk(ed.to)
			onPath[ed.to] = false
			path = path[:len(path)-1]
		}
	}
	walk(entry)
	return res, nil
}

// CanonicalJSON renders flows as indented, byte-deterministic JSON.
func (f *Flows) CanonicalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderMarkdown renders each flow as one chain line with its locators.
func (f *Flows) RenderMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Flows from %s (%d)\n", f.Entry, len(f.Flows)))
	for i, fl := range f.Flows {
		ids := make([]string, len(fl.Steps))
		for j, h := range fl.Steps {
			ids[j] = h.NodeID
		}
		sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, strings.Join(ids, " → ")))
		for _, h := range fl.Steps[1:] {
			if h.Locator != "" {
				sb.WriteString(fmt.Sprintf("   - %s `%s` [%s]\n", h.Verb, h.NodeID, h.Locator))
			} else {
				sb.WriteString(fmt.Sprintf("   - %s `%s` (unknown: no locator)\n", h.Verb, h.NodeID))
			}
		}
	}
	if f.Truncated {
		sb.WriteString(fmt.Sprintf("\n> Stopped at %d flows.\n", maxFlows))
	}
	return sb.String()
}
