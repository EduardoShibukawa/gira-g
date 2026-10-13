// Package layer2 holds the automated proxies of the human comprehension evaluator
// (spec B, Layer 2): tour assertiveness, critical-path coverage, step count and
// words per step. They run on any tool's tour once it is normalized; the moderated
// study that complements them is described in study-protocol.md.
package layer2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Claim is one checkable statement a tour step makes, with the text that must
// appear in a cited range for the claim to count as verified.
type Claim struct {
	Text   string `json:"text"`
	Anchor string `json:"anchor"`
}

// Step is a normalized tour step.
type Step struct {
	Order   int      `json:"order"`
	Title   string   `json:"title"`
	Text    string   `json:"text"`
	NodeIDs []string `json:"node_ids"`
	Read    []string `json:"read"` // locators: [repo@sha:]path#Lstart[-Lend]
	Claims  []Claim  `json:"claims"`
}

// Tour is a normalized tour from any tool.
type Tour struct {
	Tool  string `json:"tool"`
	Steps []Step `json:"steps"`
}

// Proxies are the automated Layer 2 measures of one tour.
type Proxies struct {
	Assertiveness float64  `json:"assertiveness"` // verified claims / claims
	Claims        int      `json:"claims"`
	Verified      int      `json:"verified"`
	Coverage      float64  `json:"coverage"` // flow steps visited / flow steps
	Missed        []string `json:"missed_flow_steps"`
	Steps         int      `json:"steps"`
	WordsPerStep  float64  `json:"words_per_step"`
}

// FileReader reads a corpus file at the pinned SHA.
type FileReader func(path string) ([]byte, error)

var locRE = regexp.MustCompile(`^(?:[^@]+@([0-9a-f]+):)?([^#]+)(?:#L(\d+)(?:-L(\d+))?)?$`)

// cited returns the text of a locator's range, or an error if the file does not
// exist, the commit is not the pinned one, or the range is out of bounds.
func cited(read FileReader, pinned, loc string) (string, error) {
	m := locRE.FindStringSubmatch(loc)
	if m == nil {
		return "", fmt.Errorf("malformed locator %q", loc)
	}
	if m[1] != "" && pinned != "" && m[1] != pinned {
		return "", fmt.Errorf("%s cites commit %s, pinned is %s", loc, m[1], pinned)
	}
	data, err := read(m[2])
	if err != nil {
		return "", fmt.Errorf("%s: file does not exist", loc)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if m[3] == "" {
		return string(data), nil
	}
	start, _ := strconv.Atoi(m[3])
	end := start
	if m[4] != "" {
		end, _ = strconv.Atoi(m[4])
	}
	if start < 1 || end < start || end > len(lines) {
		return "", fmt.Errorf("%s: out of bounds", loc)
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}

// Measure computes the proxies of a tour against one ground-truth flow. A claim is
// verified when some locator the step cites resolves at the pinned SHA and its
// range contains the claim's anchor (case-insensitive). A claim with no anchor or
// a step with no locator cannot be verified.
func Measure(t Tour, flow []string, read FileReader, pinnedSHA string) Proxies {
	p := Proxies{Steps: len(t.Steps), Missed: []string{}}
	visited := map[string]bool{}
	words := 0
	for _, s := range t.Steps {
		for _, id := range s.NodeIDs {
			visited[id] = true
		}
		words += len(strings.Fields(s.Title + " " + s.Text))
		var ranges []string
		for _, l := range s.Read {
			if txt, err := cited(read, pinnedSHA, l); err == nil {
				ranges = append(ranges, strings.ToLower(txt))
			}
		}
		for _, c := range s.Claims {
			p.Claims++
			if c.Anchor == "" {
				continue
			}
			for _, r := range ranges {
				if strings.Contains(r, strings.ToLower(c.Anchor)) {
					p.Verified++
					break
				}
			}
		}
	}
	if p.Claims > 0 {
		p.Assertiveness = float64(p.Verified) / float64(p.Claims)
	}
	if len(t.Steps) > 0 {
		p.WordsPerStep = float64(words) / float64(len(t.Steps))
	}
	for _, id := range flow {
		if !visited[id] {
			p.Missed = append(p.Missed, id)
		}
	}
	if len(flow) > 0 {
		p.Coverage = float64(len(flow)-len(p.Missed)) / float64(len(flow))
	}
	return p
}
