package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// hermeticIndex indexes the vendored corpus through the real pipeline and writes
// the MacroMap where the CLI can load it.
func hermeticIndex(t *testing.T) (*eval.GroundTruth, string) {
	t.Helper()
	gt, c, err := eval.Hermetic(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	idx, err := c.Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := idx.MacroMap.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "macro-map.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return gt, path
}

func runCLI(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err := Run(args, strings.NewReader(""), &out, &errOut)
	return out.Bytes(), err
}

func mcpCall(t *testing.T, s *mcp.Server, tool string, args map[string]any) (string, bool) {
	t.Helper()
	req, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": tool, "arguments": args}})
	var out bytes.Buffer
	if err := s.Serve(bytes.NewReader(append(req, '\n')), &out); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		Result mcp.CallToolResult `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp.Result.Content[0].Text, resp.Result.IsError
}

// checkBudget asserts the hard ceiling on one emitted output and returns its budget.
func checkBudget(t *testing.T, label string, out []byte, limit int, format slices.Format, focal string) {
	t.Helper()
	est := tokens.Estimate(out)
	if est > limit {
		t.Errorf("%s: estimated %d tokens > budget %d", label, est, limit)
	}
	if format == slices.FormatMarkdown {
		if !strings.Contains(string(out), fmt.Sprintf("- limit: %d, estimated: %d,", limit, est)) {
			t.Errorf("%s: markdown budget marker missing or wrong", label)
		}
		if !strings.Contains(string(out), "`"+focal+"`") && !strings.Contains(string(out), focal) {
			t.Errorf("%s: focal %s dropped", label, focal)
		}
		return
	}
	var doc struct {
		TargetID  string              `json:"target_id"`
		FocalNode struct{ ID string } `json:"focal_node"`
		Budget    *slices.Budget      `json:"budget"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	if doc.Budget == nil {
		t.Fatalf("%s: no budget object", label)
	}
	if doc.Budget.Limit != limit || doc.Budget.Estimated != est {
		t.Errorf("%s: budget %+v, want limit %d estimated %d", label, *doc.Budget, limit, est)
	}
	if doc.Budget.Truncated != (doc.Budget.Dropped != slices.Dropped{}) {
		t.Errorf("%s: truncated=%t but dropped=%+v", label, doc.Budget.Truncated, doc.Budget.Dropped)
	}
	if doc.Budget.Truncated && doc.Budget.EscalateHint == "" {
		t.Errorf("%s: truncated without an escalate hint", label)
	}
	if doc.TargetID != focal && doc.FocalNode.ID != focal {
		t.Errorf("%s: focal %s dropped", label, focal)
	}
}

// SC6: for every task and every tier, CLI (json and md) and MCP outputs are within
// budget, carry the budget object, are byte-identical on repeat and keep the focal node.
func TestSC6HardBudgetsAcrossSuite(t *testing.T) {
	gt, index := hermeticIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))

	for _, task := range gt.Tasks {
		if _, err := slices.NewEngine(mm).GetImpactBriefing(task.TargetID, task.Depth); err != nil {
			// Spec gira-impact-precision SC3: every task's target is now extracted.
			t.Errorf("%s: target %s not in the hermetic index (%v)", task.ID, task.TargetID, err)
			continue
		}
		limit := slices.BudgetBriefing
		for _, f := range []slices.Format{slices.FormatJSON, slices.FormatMarkdown} {
			args := []string{"impact", task.TargetID, "--index", index, "--depth", fmt.Sprint(task.Depth), "--budget", fmt.Sprint(limit), "--format", string(f)}
			a, err := runCLI(t, args...)
			if err != nil {
				t.Fatalf("%s impact %s: %v", task.ID, f, err)
			}
			b, _ := runCLI(t, args...)
			if !bytes.Equal(a, b) {
				t.Errorf("%s impact %s: repeated call differs", task.ID, f)
			}
			checkBudget(t, task.ID+" cli impact "+string(f), a, limit, f, task.TargetID)
		}
		text, isErr := mcpCall(t, server, "get_impact_briefing", map[string]any{"target_id": task.TargetID, "depth": task.Depth, "budget": limit})
		if isErr {
			t.Fatalf("%s mcp impact: %s", task.ID, text)
		}
		checkBudget(t, task.ID+" mcp impact", []byte(text), limit, slices.FormatJSON, task.TargetID)

		for _, tier := range []slices.SliceTier{slices.TierSummary, slices.TierContracts, slices.TierDeep} {
			limit := slices.DefaultBudget(tier)
			for _, f := range []slices.Format{slices.FormatJSON, slices.FormatMarkdown} {
				args := []string{"slice", task.TargetID, "--index", index, "--tier", string(tier), "--budget", fmt.Sprint(limit), "--format", string(f)}
				a, err := runCLI(t, args...)
				if err != nil {
					t.Fatalf("%s slice %s %s: %v", task.ID, tier, f, err)
				}
				b, _ := runCLI(t, args...)
				if !bytes.Equal(a, b) {
					t.Errorf("%s slice %s %s: repeated call differs", task.ID, tier, f)
				}
				checkBudget(t, fmt.Sprintf("%s cli slice %s %s", task.ID, tier, f), a, limit, f, task.TargetID)
			}
			text, isErr := mcpCall(t, server, "get_context_slice", map[string]any{"node_id": task.TargetID, "tier": string(tier), "budget": limit})
			if isErr {
				t.Fatalf("%s mcp slice %s: %s", task.ID, tier, text)
			}
			checkBudget(t, fmt.Sprintf("%s mcp slice %s", task.ID, tier), []byte(text), limit, slices.FormatJSON, task.TargetID)
		}
	}
}

// SC6: a budget smaller than the full output truncates in the documented order,
// reports what it dropped, and still never drops the focal node.
func TestSC6TruncationIsReportedAndDeterministic(t *testing.T) {
	_, index := hermeticIndex(t)
	full, err := runCLI(t, "impact", "service:checkout", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	limit := tokens.Estimate(full) / 2
	for _, f := range []slices.Format{slices.FormatJSON, slices.FormatMarkdown} {
		a, err := runCLI(t, "impact", "service:checkout", "--index", index, "--budget", fmt.Sprint(limit), "--format", string(f))
		if err != nil {
			t.Fatal(err)
		}
		b, _ := runCLI(t, "impact", "service:checkout", "--index", index, "--budget", fmt.Sprint(limit), "--format", string(f))
		if !bytes.Equal(a, b) {
			t.Fatalf("%s: truncated output is not byte-identical", f)
		}
		checkBudget(t, "checkout half budget "+string(f), a, limit, f, "service:checkout")
		if f == slices.FormatJSON {
			var doc struct{ Budget slices.Budget }
			_ = json.Unmarshal(a, &doc)
			if !doc.Budget.Truncated || doc.Budget.Dropped.Nodes == 0 {
				t.Fatalf("half budget must drop nodes: %+v", doc.Budget)
			}
		}
	}

	_, err = runCLI(t, "slice", "service:checkout", "--index", index, "--tier", "deep", "--budget", "5")
	if err == nil || !strings.Contains(err.Error(), "budget_exceeded") {
		t.Fatalf("a budget below the focal node must fail with budget_exceeded, got %v", err)
	}
}

// Additive: without --budget the output is exactly what it was before budgets.
func TestNoBudgetKeepsOutputUnchanged(t *testing.T) {
	_, index := hermeticIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := slices.NewEngine(mm).GetContextSliceWithView("service:checkout", slices.TierDeep, slices.ViewFull)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := cs.CanonicalJSON()
	got, err := runCLI(t, "slice", "service:checkout", "--index", index, "--tier", "deep")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || bytes.Contains(got, []byte(`"budget"`)) {
		t.Fatal("unbudgeted output changed")
	}
}

// Review F1: under a tight budget the focal Contract node stays in
// affected_contracts (or the call fails with budget_exceeded); it is never dropped.
func TestBudgetNeverDropsFocalContract(t *testing.T) {
	_, index := hermeticIndex(t)
	for _, limit := range []int{250, 300, 400, 600} {
		out, err := runCLI(t, "impact", "contract:PaymentService", "--index", index, "--budget", fmt.Sprint(limit))
		if err != nil {
			if !strings.Contains(err.Error(), "budget_exceeded") {
				t.Fatalf("budget %d: %v", limit, err)
			}
			continue
		}
		var b slices.ImpactBriefing
		if err := json.Unmarshal(out, &b); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range b.AffectedContracts {
			found = found || n.ID == "contract:PaymentService"
		}
		if !found {
			t.Fatalf("budget %d dropped the focal contract: %+v", limit, b.Budget)
		}
	}
}

// Review F7: a target the extractors never emit produces no output at all (node
// not found), so there is nothing a budget could let through. The topic used to
// be contract:kafka:orders, which the extractors now emit (spec
// gira-impact-precision SC3), so the test names a topic that does not exist.
func TestBudgetOnMissingTargetEmitsNothing(t *testing.T) {
	_, index := hermeticIndex(t)
	out, err := runCLI(t, "impact", "contract:kafka:no-such-topic", "--index", index, "--budget", "1500")
	if err == nil || len(out) != 0 || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("missing target must fail with no output: %v %q", err, out)
	}
}
