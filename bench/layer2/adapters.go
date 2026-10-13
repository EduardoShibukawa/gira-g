package layer2

import (
	"encoding/json"
	"strings"
)

// name is the bare name of a node ID ("service:cart" → "cart").
func name(id string) string { return id[strings.LastIndex(id, ":")+1:] }

// FromGIRA normalizes a `gira tour` JSON document. Each step makes one claim, the
// one its why template states, anchored on the identifier the claim is about:
//
//	focus:   "<node> is the focus"            anchor: the node's name
//	entry:   "<node> declares ingress on P"   anchor: the first declared port
//	reached: "<node> is reached because <from> <verb> it"  anchor: the node's name
func FromGIRA(data []byte) (Tour, error) {
	var g struct {
		Steps []struct {
			Order       int      `json:"order"`
			Title       string   `json:"title"`
			NodeIDs     []string `json:"node_ids"`
			Read        []string `json:"read"`
			Why         string   `json:"why"`
			WhyTemplate string   `json:"why_template"`
			WhyArgs     []string `json:"why_args"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		return Tour{}, err
	}
	t := Tour{Tool: "gira"}
	for _, s := range g.Steps {
		c := Claim{Text: s.Why}
		if len(s.WhyArgs) > 0 {
			c.Anchor = name(s.WhyArgs[0])
		}
		if s.WhyTemplate == "entry" && len(s.WhyArgs) > 1 {
			c.Anchor = strings.TrimSpace(strings.Split(strings.TrimPrefix(s.WhyArgs[1], "ports "), ",")[0])
		}
		t.Steps = append(t.Steps, Step{Order: s.Order, Title: s.Title, Text: s.Why, NodeIDs: s.NodeIDs, Read: s.Read, Claims: []Claim{c}})
	}
	return t, nil
}

// FromUA normalizes Understand Anything's documented tour: TourStep{order, title,
// description, nodeIds[], languageLesson?} inside knowledge-graph.json. Node IDs
// are lifted to Components by lift (the Layer 0 table). Each description is one
// claim. UA cites no line ranges, so a step's locators are the whole files of its
// nodes, and the only mechanical check is that the first symbol the step names
// (a non-file node) is in them. A step that names only files cannot be verified.
func FromUA(data []byte, lift func(path string) string) (Tour, error) {
	var g struct {
		Nodes []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			FilePath string `json:"filePath"`
		} `json:"nodes"`
		Tour []struct {
			Order       int      `json:"order"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			NodeIDs     []string `json:"nodeIds"`
		} `json:"tour"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		return Tour{}, err
	}
	files, symbols := map[string]string{}, map[string]string{}
	for _, n := range g.Nodes {
		files[n.ID] = n.FilePath
		if n.Type != "file" {
			symbols[n.ID] = n.Name
		}
	}
	t := Tour{Tool: "understand-anything"}
	for _, s := range g.Tour {
		claim := Claim{Text: s.Description}
		st := Step{Order: s.Order, Title: s.Title, Text: s.Description}
		seen := map[string]bool{}
		for _, id := range s.NodeIDs {
			if claim.Anchor == "" {
				claim.Anchor = symbols[id]
			}
			if f := files[id]; f != "" {
				st.Read = append(st.Read, f)
				if c := lift(f); c != "" && !seen[c] {
					seen[c] = true
					st.NodeIDs = append(st.NodeIDs, c)
				}
			}
		}
		st.Claims = []Claim{claim}
		t.Steps = append(t.Steps, st)
	}
	return t, nil
}
