package layer1

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// MockModel is the offline, deterministic model of mock-model mode. It follows a
// fixed policy through the real tool loop, so the runner's loop, telemetry and
// scorer can be tested without any model endpoint. Token usage is the 4-bytes-per-
// token estimate of what it reads and writes. It knows nothing about the corpus
// beyond what its tools return.
//
//	arm with get_impact_briefing (B): call it once (budget 1500), answer from it;
//	arm with base tools only (A):     grep the target's name, answer with the
//	                                  Components and proto services the hits name.
type MockModel struct{}

// ID is recorded in telemetry as the model ID.
func (MockModel) ID() string { return "mock-policy/1" }

var (
	targetRE  = regexp.MustCompile(`(?m)^Target: (\S+)$`)
	depthRE   = regexp.MustCompile(`(?m)^Depth: (\d+)$`)
	srcDirRE  = regexp.MustCompile(`^src/([^/]+)/`)
	protoSvRE = regexp.MustCompile(`^pb/[^:]+:(\d+):service (\w+Service)`)
)

// Complete returns the policy's next step.
func (m MockModel) Complete(_ context.Context, msgs []Message, tools []ToolSpec) (Reply, error) {
	var user string
	toolResults := []string{}
	for _, msg := range msgs {
		switch msg.Role {
		case "user":
			user = msg.Content
		case "tool":
			toolResults = append(toolResults, msg.Content)
		}
	}
	tm := targetRE.FindStringSubmatch(user)
	if tm == nil {
		return Reply{}, fmt.Errorf("mock: no target in the prompt")
	}
	target := tm[1]
	depth := 1
	if dm := depthRE.FindStringSubmatch(user); dm != nil {
		fmt.Sscan(dm[1], &depth)
	}
	hasGIRA := false
	for _, t := range tools {
		hasGIRA = hasGIRA || t.Name == "get_impact_briefing"
	}

	var reply Reply
	switch {
	case len(toolResults) == 0 && hasGIRA:
		reply.ToolCalls = []ToolCall{{ID: "call-1", Name: "get_impact_briefing", Arguments: map[string]any{"target_id": target, "depth": depth, "budget": 1500}}}
	case len(toolResults) == 0:
		name := target[strings.LastIndex(target, ":")+1:]
		reply.ToolCalls = []ToolCall{{ID: "call-1", Name: "grep", Arguments: map[string]any{"pattern": name}}}
	case hasGIRA:
		reply.Content = answerFromBriefing(toolResults[0])
	default:
		reply.Content = answerFromGrep(toolResults[0], target)
	}
	in := 0
	for _, msg := range msgs {
		in += len(msg.Content)
	}
	specs, _ := json.Marshal(tools)
	calls, _ := json.Marshal(reply.ToolCalls)
	reply.Usage = Usage{Input: (in + len(specs) + 3) / 4, Output: (len(reply.Content) + len(calls) + 3) / 4}
	return reply, nil
}

// briefNode is the part of a briefing node the mock reads.
type briefNode struct {
	ID       string `json:"id"`
	Evidence []struct {
		Locator struct {
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		} `json:"locator"`
	} `json:"evidence"`
}

func answerFromBriefing(text string) string {
	var b struct {
		Up        []briefNode `json:"upstream_services"`
		Down      []briefNode `json:"downstream_services"`
		Contracts []briefNode `json:"affected_contracts"`
	}
	a := Answer{Upstream: []string{}, Downstream: []string{}, Contracts: []string{}, Locators: []string{}}
	_ = json.Unmarshal([]byte(text), &b) // a tool error leaves the answer empty
	seen := map[string]bool{}
	collect := func(dst *[]string, nodes []briefNode) {
		for _, n := range nodes {
			*dst = append(*dst, n.ID)
			if len(n.Evidence) == 0 {
				continue
			}
			l := n.Evidence[0].Locator
			loc := fmt.Sprintf("%s#L%d-L%d", l.Path, l.StartLine, l.EndLine)
			if l.StartLine > 0 && !seen[loc] {
				seen[loc] = true
				a.Locators = append(a.Locators, loc)
			}
		}
	}
	collect(&a.Upstream, b.Up)
	collect(&a.Downstream, b.Down)
	collect(&a.Contracts, b.Contracts)
	out, _ := json.Marshal(a)
	return string(out)
}

func answerFromGrep(text, target string) string {
	comps, contracts := map[string]bool{}, map[string]bool{}
	var locs []string
	for _, line := range strings.Split(text, "\n") {
		if m := srcDirRE.FindStringSubmatch(line); m != nil && "service:"+m[1] != target {
			comps["service:"+m[1]] = true
		}
		if m := protoSvRE.FindStringSubmatch(line); m != nil {
			contracts["contract:"+m[2]] = true
		}
		if parts := strings.SplitN(line, ":", 3); len(parts) == 3 && len(locs) < 5 {
			locs = append(locs, fmt.Sprintf("%s#L%s-L%s", parts[0], parts[1], parts[1]))
		}
	}
	a := Answer{Upstream: sortedKeys(comps), Downstream: []string{}, Contracts: sortedKeys(contracts), Locators: locs}
	if a.Locators == nil {
		a.Locators = []string{}
	}
	out, _ := json.Marshal(a)
	return string(out)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
