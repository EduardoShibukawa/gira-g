package layer1

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Task is one benchmark task as the ground truth declares it.
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	TargetID    string `json:"target_id"`
	Depth       int    `json:"depth"`
	GroundTruth struct {
		Upstream   []string `json:"expected_upstream_services"`
		Downstream []string `json:"expected_downstream_services"`
		Contracts  []string `json:"expected_affected_contracts"`
		Anchors    []struct {
			Path     string `json:"path"`
			Contains string `json:"contains"`
		} `json:"anchors"`
	} `json:"ground_truth"`
}

// GroundTruth is the subset of the v2 fixture Layer 1 reads.
type GroundTruth struct {
	Corpus struct {
		UpstreamSHA string `json:"upstream_sha"`
	} `json:"corpus"`
	Ignore         []string `json:"ignore"`
	ComponentPaths []struct {
		Component string `json:"component"`
	} `json:"component_paths"`
	Tasks []Task `json:"tasks"`
}

// Answer is the required JSON answer schema: the Impact Set plus locators.
type Answer struct {
	Upstream   []string `json:"upstream"`
	Downstream []string `json:"downstream"`
	Contracts  []string `json:"contracts"`
	Locators   []string `json:"locators"`
}

// AnswerSchema is given to the model verbatim in every task prompt.
const AnswerSchema = `{"upstream": ["<node id>"], "downstream": ["<node id>"], "contracts": ["<contract id>"], "locators": ["<path>#L<start>-L<end>"]}`

// ParseAnswer extracts the first JSON object from a model's final message.
func ParseAnswer(text string) (Answer, error) {
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return Answer{}, fmt.Errorf("no JSON object in the answer")
	}
	var a Answer
	if err := json.Unmarshal([]byte(text[start:end+1]), &a); err != nil {
		return Answer{}, fmt.Errorf("answer does not match the schema: %w", err)
	}
	return a, nil
}

// Score is the deterministic accuracy of one answer.
type Score struct {
	Precision         float64  `json:"precision"`
	Recall            float64  `json:"recall"`
	F1                float64  `json:"f1"`
	LocatorValidity   float64  `json:"locator_validity"`
	LocatorsTotal     int      `json:"locators_total"`
	HallucinatedNodes []string `json:"hallucinated_nodes"`
}

// Universe is every node ID the corpus ground truth knows: task targets, expected
// sets, ignored ambient Components and the Component table.
func (gt *GroundTruth) Universe() map[string]bool {
	u := map[string]bool{}
	add := func(ids ...string) {
		for _, id := range ids {
			u[id] = true
		}
	}
	add(gt.Ignore...)
	for _, c := range gt.ComponentPaths {
		add(c.Component)
	}
	for _, t := range gt.Tasks {
		add(t.TargetID)
		add(t.GroundTruth.Upstream...)
		add(t.GroundTruth.Downstream...)
		add(t.GroundTruth.Contracts...)
	}
	return u
}

var locRE = regexp.MustCompile(`^(?:[^@]+@[0-9a-f]+:)?([^#]+)#L(\d+)(?:-L(\d+))?$`)

// ScoreAnswer compares an answer with a task's expected Impact Set. IDs in ignore
// count neither way. A locator is valid when layer0's rule accepts it: the file
// exists, the range is in bounds. Hallucinated nodes are answer IDs absent from
// the ground-truth universe.
func ScoreAnswer(a Answer, t Task, ignore, universe map[string]bool, check func(path string, start, end int) error) Score {
	expected, actual := map[string]bool{}, map[string]bool{}
	for _, id := range append(append(append([]string{}, t.GroundTruth.Upstream...), t.GroundTruth.Downstream...), t.GroundTruth.Contracts...) {
		if !ignore[id] {
			expected[id] = true
		}
	}
	s := Score{HallucinatedNodes: []string{}}
	for _, id := range append(append(append([]string{}, a.Upstream...), a.Downstream...), a.Contracts...) {
		if !universe[id] {
			s.HallucinatedNodes = append(s.HallucinatedNodes, id)
		}
		if !ignore[id] {
			actual[id] = true
		}
	}
	sort.Strings(s.HallucinatedNodes)
	s.HallucinatedNodes = dedup(s.HallucinatedNodes)
	tp := 0
	for id := range actual {
		if expected[id] {
			tp++
		}
	}
	if len(actual) > 0 {
		s.Precision = float64(tp) / float64(len(actual))
	}
	if len(expected) > 0 {
		s.Recall = float64(tp) / float64(len(expected))
	}
	if s.Precision+s.Recall > 0 {
		s.F1 = 2 * s.Precision * s.Recall / (s.Precision + s.Recall)
	}
	valid := 0
	for _, l := range a.Locators {
		s.LocatorsTotal++
		m := locRE.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		start, _ := strconv.Atoi(m[2])
		end := start
		if m[3] != "" {
			end, _ = strconv.Atoi(m[3])
		}
		if check(m[1], start, end) == nil {
			valid++
		}
	}
	if s.LocatorsTotal > 0 {
		s.LocatorValidity = float64(valid) / float64(s.LocatorsTotal)
	}
	return s
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
