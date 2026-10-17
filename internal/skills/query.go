package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Question is one gira-query request. Op selects the operation; the other fields
// are its arguments, named as in the MCP tools.
type Question struct {
	Op        string `json:"op"` // impact, slice, explain, locate, tour, pr_impact, diff, doc, flows
	TargetID  string `json:"target_id,omitempty"`
	NodeID    string `json:"node_id,omitempty"`
	EdgeID    string `json:"edge_id,omitempty"`
	ServiceID string `json:"service_id,omitempty"`
	Entry     string `json:"entry,omitempty"`
	Tier      string `json:"tier,omitempty"`
	View      string `json:"view,omitempty"`
	Query     string `json:"query,omitempty"`
	Focus     string `json:"focus,omitempty"`
	Range     string `json:"range,omitempty"`
	RepoPath  string `json:"repo_path,omitempty"`
	IndexA    string `json:"index_a,omitempty"`
	IndexB    string `json:"index_b,omitempty"`
	Depth     int    `json:"depth,omitempty"`
	Budget    int    `json:"budget,omitempty"`
	MaxSteps  int    `json:"max_steps,omitempty"`
}

// Error codes agents branch on (spec A.1, gira-query).
const (
	CodeIndexMissing   = "index_missing"
	CodeIndexStale     = "index_stale"
	CodeNodeNotFound   = "node_not_found"
	CodeBudgetExceeded = "budget_exceeded"
	CodeInternal       = "internal"
)

// QueryError is the error document both paths return.
type QueryError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
}

// MCPCaller calls a GIRA MCP tool and returns its text content.
type MCPCaller func(tool string, args map[string]any) (text string, isError bool, err error)

// CLIRunner runs `gira <args...>` and returns its stdout, or an error carrying
// its message.
type CLIRunner func(args []string) ([]byte, error)

// Querier is the gira-query procedure. MCP is nil when no GIRA MCP tools are
// registered; the procedure then falls back to the CLI.
type Querier struct {
	MCP   MCPCaller
	CLI   CLIRunner
	Index string // index the CLI reads; the MCP server was started on the same one
	// RepoDir, when set, is checked against the index: if its HEAD is not a commit
	// the index cites, the answer is index_stale.
	RepoDir string
}

// Ask answers q and returns normalized JSON (two-space indent, trailing newline):
// the operation's document, or {"error": QueryError}.
func (qr Querier) Ask(q Question) ([]byte, error) {
	if qr.Index != "" {
		if _, err := os.Stat(qr.Index); err != nil {
			return errorDoc(QueryError{Code: CodeIndexMissing, Message: fmt.Sprintf("no index at %s: run gira index first", qr.Index)})
		}
		if qr.RepoDir != "" {
			if stale, head, err := isStale(qr.Index, qr.RepoDir); err != nil {
				return nil, err
			} else if stale {
				return errorDoc(QueryError{Code: CodeIndexStale, Message: fmt.Sprintf("the index does not cite HEAD %s of %s: re-index", head, qr.RepoDir)})
			}
		}
	}
	tool, args, cli, err := q.route(qr.Index)
	if err != nil {
		return nil, err
	}
	var out []byte
	var failure string
	if qr.MCP != nil {
		text, isErr, err := qr.MCP(tool, args)
		if err != nil {
			return nil, err
		}
		if isErr {
			failure = text
		} else {
			out = []byte(text)
		}
	} else {
		res, err := qr.CLI(cli)
		if err != nil {
			failure = err.Error()
		} else {
			out = res
		}
	}
	if failure != "" {
		return qr.classify(q, failure)
	}
	return normalize(out)
}

// route maps a question to its MCP tool call and its CLI equivalent.
func (q Question) route(index string) (string, map[string]any, []string, error) {
	args := map[string]any{}
	var cli []string
	put := func(k string, v any, flag string) {
		switch x := v.(type) {
		case string:
			if x == "" {
				return
			}
		case int:
			if x == 0 {
				return
			}
		}
		args[k] = v
		if flag != "" {
			cli = append(cli, flag, fmt.Sprint(v))
		}
	}
	var tool string
	switch q.Op {
	case "impact":
		tool, cli = "get_impact_briefing", []string{"impact", q.TargetID}
		args["target_id"] = q.TargetID
		put("depth", q.Depth, "--depth")
		put("budget", q.Budget, "--budget")
	case "slice":
		tool, cli = "get_context_slice", []string{"slice", q.NodeID}
		args["node_id"] = q.NodeID
		put("tier", q.Tier, "--tier")
		put("view", q.View, "--view")
		put("budget", q.Budget, "--budget")
	case "explain":
		tool, cli = "explain_evidence", []string{"explain"}
		put("node_id", q.NodeID, "--node")
		put("edge_id", q.EdgeID, "--edge")
		put("budget", q.Budget, "--budget")
	case "locate":
		tool, cli = "locate_boundary", []string{"locate", q.Query}
		args["query"] = q.Query
		put("budget", q.Budget, "--budget")
	case "tour":
		tool, cli = "get_tour", []string{"tour"}
		put("focus", q.Focus, "--focus")
		put("max_steps", q.MaxSteps, "--max-steps")
		put("budget", q.Budget, "--budget")
	case "doc":
		svc := q.ServiceID
		if svc == "" {
			svc = q.NodeID
		}
		if svc == "" {
			svc = q.TargetID
		}
		tool, cli = "get_doc", []string{"doc", svc}
		args["service_id"] = svc
		put("budget", q.Budget, "--budget")
	case "flows":
		entry := q.Entry
		if entry == "" {
			entry = q.NodeID
		}
		if entry == "" {
			entry = q.TargetID
		}
		tool, cli = "get_flows", []string{"flows", entry}
		args["entry"] = entry
		put("budget", q.Budget, "--budget")
	case "pr_impact":
		tool, cli = "get_pr_impact", []string{"impact", "--from-diff", q.Range}
		args["range"] = q.Range
		put("repo_path", q.RepoPath, "--repo")
		put("depth", q.Depth, "--depth")
		put("budget", q.Budget, "--budget")
	case "diff":
		tool, cli = "diff_maps", []string{"diff", q.IndexA, q.IndexB}
		args["index_a"], args["index_b"] = q.IndexA, q.IndexB
		put("budget", q.Budget, "--budget")
		cli = append(cli, "--format", "json")
		return tool, args, cli, nil
	default:
		return "", nil, nil, fmt.Errorf("unknown gira-query op %q", q.Op)
	}
	if index != "" {
		cli = append(cli, "--index", index)
	}
	cli = append(cli, "--format", "json")
	return tool, args, cli, nil
}

// classify maps a failure message from either path to the shared vocabulary.
func (qr Querier) classify(q Question, msg string) ([]byte, error) {
	var parsed struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal([]byte(msg), &parsed)
	low := strings.ToLower(msg)
	e := QueryError{Code: CodeInternal, Message: strings.TrimSpace(msg), Suggestions: []string{}}
	switch {
	case parsed.Code == CodeBudgetExceeded, strings.Contains(low, "budget_exceeded"):
		e.Code = CodeBudgetExceeded
	case parsed.Code == CodeIndexMissing, strings.Contains(low, "index file not found"), strings.Contains(low, "failed to load index"):
		e.Code = CodeIndexMissing
	case parsed.Code == CodeNodeNotFound, strings.Contains(low, "not found"):
		e.Code = CodeNodeNotFound
		id := q.TargetID + q.NodeID + q.Focus + q.ServiceID + q.Entry
		if id != "" {
			e.Suggestions = qr.suggest(id)
		}
	}
	// The message wording differs per transport; the code and suggestions don't.
	e.Message = map[string]string{
		CodeBudgetExceeded: "the parts that are never dropped exceed the budget: raise it",
		CodeIndexMissing:   "no index found: run gira index first",
		CodeNodeNotFound:   "the id is not in the index; see suggestions",
	}[e.Code]
	if e.Code == CodeInternal {
		e.Message = strings.TrimSpace(msg)
	}
	return errorDoc(e)
}

// suggest runs locate for an unknown id: on its last segment, then on each of
// that segment's words, then on ever shorter prefixes (down to 4 characters),
// and returns the matches of the first term that has any.
func (qr Querier) suggest(id string) []string {
	term := id
	if i := strings.LastIndex(term, ":"); i >= 0 {
		term = term[i+1:]
	}
	terms := []string{term}
	terms = append(terms, strings.FieldsFunc(term, func(r rune) bool { return r == '-' || r == '_' || r == '.' || r == '/' })...)
	for n := len(term) - 1; n >= 4; n-- {
		terms = append(terms, term[:n])
	}
	locator := Querier{MCP: qr.MCP, CLI: qr.CLI, Index: qr.Index}
	for _, t := range terms {
		doc, err := locator.Ask(Question{Op: "locate", Query: t})
		if err != nil {
			continue
		}
		var matches []struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(doc, &matches) != nil || len(matches) == 0 {
			continue
		}
		out := make([]string, 0, len(matches))
		seen := map[string]bool{}
		for _, m := range matches {
			if !seen[m.ID] {
				seen[m.ID] = true
				out = append(out, m.ID)
			}
		}
		return out
	}
	return []string{}
}

func errorDoc(e QueryError) ([]byte, error) {
	data, err := json.Marshal(map[string]QueryError{"error": e})
	if err != nil {
		return nil, err
	}
	return normalize(data)
}

func normalize(doc []byte) ([]byte, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, doc); err != nil {
		return nil, fmt.Errorf("not JSON: %w", err)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, compact.Bytes(), "", "  "); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

// isStale reports whether HEAD of repoDir is absent from the commits the index cites.
func isStale(index, repoDir string) (bool, string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		return false, "", fmt.Errorf("reading HEAD of %s: %w", repoDir, err)
	}
	head := strings.TrimSpace(string(out))
	data, err := os.ReadFile(index)
	if err != nil {
		return false, head, err
	}
	return !bytes.Contains(data, []byte(strconv.Quote(head))), head, nil
}
