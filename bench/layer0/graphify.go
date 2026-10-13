package layer0

import (
	"encoding/json"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// graphifyGraph is Graphify's graph.json: NetworkX node-link JSON with top-level
// directed, multigraph, graph{schema_version, graphify_version}, nodes, links and
// built_at_commit. Link order is the direction (docs/prior-art/graphify.md §2).
type graphifyGraph struct {
	Graph struct {
		Version string `json:"graphify_version"`
	} `json:"graph"`
	BuiltAtCommit string `json:"built_at_commit"`
	Nodes         []struct {
		ID             string `json:"id"`
		Label          string `json:"label"`
		SourceFile     string `json:"source_file"`
		SourceLocation string `json:"source_location"`
	} `json:"nodes"`
	Links []struct {
		Source         string `json:"source"`
		Target         string `json:"target"`
		Relation       string `json:"relation"`
		SourceFile     string `json:"source_file"`
		SourceLocation string `json:"source_location"`
	} `json:"links"`
}

var lineRE = regexp.MustCompile(`^L(\d+)$`)

// line parses Graphify's single-line location ("L14").
func line(loc string) int {
	if m := lineRE.FindStringSubmatch(strings.TrimSpace(loc)); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

var protoServiceRE = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*Service$`)

// FromGraphify lifts a Graphify graph to a NAG. Each node belongs to the Component
// of its source_file; a link between nodes of different Components is a Component
// edge citing the link's source_file and line. `imports` and `imports_from` map to
// imports; every other relation has no NAG verb and maps to unknown. A node defined
// in a .proto file whose label names a service is counted as that Contract.
func FromGraphify(data []byte, l Lifter) (*NAG, error) {
	var g graphifyGraph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	b := newBuilder("graphify", g.Graph.Version, g.BuiltAtCommit, true)
	comp := map[string]string{}
	for _, n := range g.Nodes {
		comp[n.ID] = l.Component(n.SourceFile)
		b.component(comp[n.ID])
		if path.Ext(n.SourceFile) == ".proto" && protoServiceRE.MatchString(n.Label) {
			b.contract("contract:" + n.Label)
		}
	}
	for _, k := range g.Links {
		verb := VerbUnknown
		if k.Relation == "imports" || k.Relation == "imports_from" {
			verb = VerbImports
		}
		var ev []Evidence
		if ln := line(k.SourceLocation); k.SourceFile != "" && ln > 0 {
			ev = append(ev, Evidence{k.SourceFile, ln, ln})
		}
		b.edge(comp[k.Source], comp[k.Target], verb, ev...)
	}
	return b.done(), nil
}
