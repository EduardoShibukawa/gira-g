package layer0

import "encoding/json"

// giraMap is the subset of GIRA's macro-map.json (schema 1.0.0) the adapter reads.
type giraMap struct {
	Nodes []struct {
		ID       string         `json:"id"`
		Kind     string         `json:"kind"`
		Metadata map[string]any `json:"metadata"`
	} `json:"nodes"`
	Relationships []struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
		Verb     string `json:"verb"`
		Evidence []struct {
			Locator struct {
				CommitSHA string `json:"commit_sha"`
				Path      string `json:"path"`
				StartLine int    `json:"start_line"`
				EndLine   int    `json:"end_line"`
			} `json:"locator"`
		} `json:"evidence"`
	} `json:"relationships"`
}

// FromGIRA maps a GIRA macro-map.json to a NAG. GIRA has Components (Service and
// Infrastructure nodes), so nothing is lifted. Its Contract edges become Component
// edges: a consumer of Contract X consumes every provider of X, citing the
// consumer's evidence. Library, Policy, Skill and Task nodes are not Components.
func FromGIRA(data []byte) (*NAG, error) {
	var m giraMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	kind := map[string]string{}
	for _, n := range m.Nodes {
		kind[n.ID] = n.Kind
	}
	isComponent := func(id string) bool { return kind[id] == "Service" || kind[id] == "Infrastructure" }

	b := newBuilder("gira", "", "", false)
	for _, n := range m.Nodes {
		switch {
		case isComponent(n.ID):
			b.component(n.ID)
		case n.Kind == "Contract":
			b.contract(n.ID)
		}
	}
	providers := map[string][]string{}
	for _, r := range m.Relationships {
		if r.Verb == "provides" && isComponent(r.SourceID) && kind[r.TargetID] == "Contract" {
			providers[r.TargetID] = append(providers[r.TargetID], r.SourceID)
		}
	}
	for _, r := range m.Relationships {
		var evs []Evidence
		for _, e := range r.Evidence {
			if b.nag.Commit == "" {
				b.nag.Commit = e.Locator.CommitSHA
			}
			evs = append(evs, Evidence{e.Locator.Path, e.Locator.StartLine, e.Locator.EndLine})
		}
		verb := r.Verb
		switch verb {
		case VerbProvides, VerbConsumes, VerbImports, VerbAccesses:
		default:
			verb = VerbUnknown
		}
		switch {
		case isComponent(r.SourceID) && isComponent(r.TargetID):
			b.edge(r.SourceID, r.TargetID, verb, evs...)
		case r.Verb == "consumes" && isComponent(r.SourceID) && kind[r.TargetID] == "Contract":
			for _, p := range providers[r.TargetID] {
				b.edge(r.SourceID, p, VerbConsumes, evs...)
			}
		}
	}
	return b.done(), nil
}
