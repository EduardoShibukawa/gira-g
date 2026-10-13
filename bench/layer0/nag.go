// Package layer0 is the structural graph evaluator of the comparative benchmark
// (spec B, Layer 0). One adapter per tool maps the tool's own output into a
// Normalized Architecture Graph (NAG); the evaluator, not the tool, lifts file and
// symbol nodes to Components with one path→Component table shared by every tool.
// Adapters are written from each tool's documented schema; no tool code is used.
package layer0

import (
	"bytes"
	"encoding/json"
	"errors"
	"path"
	"sort"
	"strings"
)

// Verbs of a NAG edge. VerbUnknown is used when the tool has no matching verb.
const (
	VerbProvides   = "provides"
	VerbConsumes   = "consumes"
	VerbImports    = "imports"
	VerbAccesses   = "accesses"
	VerbPublishes  = "publishes"
	VerbSubscribes = "subscribes"
	VerbUnknown    = "unknown"
)

// Evidence is a cited range of a corpus file.
type Evidence struct {
	Path  string `json:"path"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Edge is a directed edge between two Components.
type Edge struct {
	Source   string     `json:"source"`
	Target   string     `json:"target"`
	Verb     string     `json:"verb"`
	Evidence []Evidence `json:"evidence"`
}

// NAG is the Normalized Architecture Graph of one tool run.
type NAG struct {
	Tool       string   `json:"tool"`
	Version    string   `json:"version,omitempty"`
	Commit     string   `json:"commit,omitempty"`
	Components []string `json:"components"`
	Contracts  []string `json:"contracts"`
	Edges      []Edge   `json:"edges"`
	// Projected is true when the tool has no Component concept and the NAG was
	// lifted from file or symbol nodes (reports must say so).
	Projected bool `json:"projected"`
}

// ComponentPath assigns files under PathPrefix to Component (ground truth).
type ComponentPath struct {
	PathPrefix string `json:"path_prefix"`
	Component  string `json:"component"`
}

// Lifter maps corpus paths to Components by longest matching prefix.
type Lifter struct{ table []ComponentPath }

// NewLifter builds a Lifter from the ground-truth table.
func NewLifter(table []ComponentPath) Lifter {
	t := append([]ComponentPath(nil), table...)
	sort.Slice(t, func(i, j int) bool { return len(t[i].PathPrefix) > len(t[j].PathPrefix) })
	return Lifter{t}
}

// Component returns the Component of a corpus path, or "".
func (l Lifter) Component(p string) string {
	p = strings.TrimPrefix(path.Clean(strings.ReplaceAll(p, "\\", "/")), "./")
	for _, row := range l.table {
		if strings.HasPrefix(p, row.PathPrefix) {
			return row.Component
		}
	}
	return ""
}

// builder accumulates a NAG, merging duplicate edges and their evidence.
type builder struct {
	nag        NAG
	components map[string]bool
	contracts  map[string]bool
	edges      map[[3]string]map[Evidence]bool
}

func newBuilder(tool, version, commit string, projected bool) *builder {
	return &builder{
		nag:        NAG{Tool: tool, Version: version, Commit: commit, Projected: projected},
		components: map[string]bool{},
		contracts:  map[string]bool{},
		edges:      map[[3]string]map[Evidence]bool{},
	}
}

func (b *builder) component(id string) {
	if id != "" {
		b.components[id] = true
	}
}

func (b *builder) contract(id string) { b.contracts[id] = true }

// edge adds a Component edge; self-edges (inside one Component) are not edges of
// the architecture and are skipped.
func (b *builder) edge(src, dst, verb string, ev ...Evidence) {
	if src == "" || dst == "" || src == dst {
		return
	}
	b.component(src)
	b.component(dst)
	k := [3]string{src, dst, verb}
	if b.edges[k] == nil {
		b.edges[k] = map[Evidence]bool{}
	}
	for _, e := range ev {
		if e.Path != "" {
			b.edges[k][e] = true
		}
	}
}

func (b *builder) done() *NAG {
	n := b.nag
	n.Components = keys(b.components)
	n.Contracts = keys(b.contracts)
	n.Edges = []Edge{}
	for k, evs := range b.edges {
		e := Edge{Source: k[0], Target: k[1], Verb: k[2], Evidence: []Evidence{}}
		for ev := range evs {
			e.Evidence = append(e.Evidence, ev)
		}
		sort.Slice(e.Evidence, func(i, j int) bool {
			a, c := e.Evidence[i], e.Evidence[j]
			if a.Path != c.Path {
				return a.Path < c.Path
			}
			if a.Start != c.Start {
				return a.Start < c.Start
			}
			return a.End < c.End
		})
		n.Edges = append(n.Edges, e)
	}
	sort.Slice(n.Edges, func(i, j int) bool {
		a, c := n.Edges[i], n.Edges[j]
		if a.Source != c.Source {
			return a.Source < c.Source
		}
		if a.Target != c.Target {
			return a.Target < c.Target
		}
		return a.Verb < c.Verb
	})
	return &n
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CanonicalJSON renders a NAG as indented, byte-deterministic JSON.
func (n *NAG) CanonicalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(n)
	return buf.Bytes(), err
}

// LoadComponentTable reads the component_paths table of a ground-truth fixture.
func LoadComponentTable(data []byte) (Lifter, error) {
	var gt struct {
		ComponentPaths []ComponentPath `json:"component_paths"`
	}
	if err := json.Unmarshal(data, &gt); err != nil {
		return Lifter{}, err
	}
	if len(gt.ComponentPaths) == 0 {
		return Lifter{}, errNoTable
	}
	return NewLifter(gt.ComponentPaths), nil
}

var errNoTable = errors.New("ground truth has no component_paths table")
