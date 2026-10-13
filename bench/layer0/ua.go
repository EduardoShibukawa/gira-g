package layer0

import (
	"encoding/json"
	"path"
)

// uaGraph is Understand Anything's knowledge-graph.json: KnowledgeGraph{version,
// project{gitCommitHash,...}, nodes, edges, layers, tour}. Nodes are files or
// symbols with filePath and lineRange; edges are {source, target, type, direction,
// weight} and carry no location (docs/prior-art/understand-anything.md).
type uaGraph struct {
	Version string `json:"version"`
	Project struct {
		GitCommitHash string `json:"gitCommitHash"`
	} `json:"project"`
	Nodes []struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Name      string `json:"name"`
		FilePath  string `json:"filePath"`
		LineRange []int  `json:"lineRange"`
	} `json:"nodes"`
	Edges []struct {
		Source    string `json:"source"`
		Target    string `json:"target"`
		Type      string `json:"type"`
		Direction string `json:"direction"`
	} `json:"edges"`
}

// FromUA lifts an Understand Anything knowledge graph to a NAG. Nodes belong to
// the Component of their filePath. UA edges have no evidence, so their NAG edges
// carry none: the adapter never borrows a node's location for an edge. `imports`
// maps to imports and every other type to unknown. Direction "backward" reverses
// an edge; "bidirectional" yields both directions. A node in a .proto file whose
// name names a service is counted as that Contract.
func FromUA(data []byte, l Lifter) (*NAG, error) {
	var g uaGraph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	b := newBuilder("understand-anything", g.Version, g.Project.GitCommitHash, true)
	comp := map[string]string{}
	for _, n := range g.Nodes {
		comp[n.ID] = l.Component(n.FilePath)
		b.component(comp[n.ID])
		if path.Ext(n.FilePath) == ".proto" && protoServiceRE.MatchString(n.Name) {
			b.contract("contract:" + n.Name)
		}
	}
	for _, e := range g.Edges {
		verb := VerbUnknown
		if e.Type == "imports" {
			verb = VerbImports
		}
		src, dst := comp[e.Source], comp[e.Target]
		switch e.Direction {
		case "backward":
			b.edge(dst, src, verb)
		case "bidirectional":
			b.edge(src, dst, verb)
			b.edge(dst, src, verb)
		default:
			b.edge(src, dst, verb)
		}
	}
	return b.done(), nil
}
