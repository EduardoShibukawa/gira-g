package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/skills"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite cmd/gira/testdata/gira-query golden files")

// SC11: the gira-query procedure returns the same JSON for the same question with
// the MCP server registered or with only the CLI, and both match the golden file.
func TestSC11GiraQueryFallbackGolden(t *testing.T) {
	_, index := hermeticIndex(t)
	_, perturbed := perturbedIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))
	viaMCP := skills.Querier{
		Index: index,
		MCP: func(tool string, args map[string]any) (string, bool, error) {
			text, isErr := mcpCall(t, server, tool, args)
			return text, isErr, nil
		},
	}
	viaCLI := skills.Querier{
		Index: index,
		CLI: func(args []string) ([]byte, error) {
			var out, errOut bytes.Buffer
			if err := Run(args, strings.NewReader(""), &out, &errOut); err != nil {
				return nil, errors.New(err.Error() + " " + errOut.String())
			}
			return out.Bytes(), nil
		},
	}

	questions := map[string]skills.Question{
		"impact-checkout":           {Op: "impact", TargetID: "service:checkout", Budget: 1500},
		"impact-paymentservice":     {Op: "impact", TargetID: "contract:PaymentService"},
		"slice-cart-contracts":      {Op: "slice", NodeID: "service:cart", Tier: "contracts", Budget: 1500},
		"slice-checkout-deep-small": {Op: "slice", NodeID: "service:checkout", Tier: "deep", Budget: 900},
		"explain-checkout":          {Op: "explain", NodeID: "service:checkout"},
		"locate-currency":           {Op: "locate", Query: "currency"},
		"tour-frontend":             {Op: "tour", Focus: "service:frontend", MaxSteps: 5},
		"diff-published-perturbed":  {Op: "diff", IndexA: index, IndexB: perturbed},
		"error-node-not-found":      {Op: "impact", TargetID: "service:cart-svc"},
		"error-budget-exceeded":     {Op: "slice", NodeID: "service:checkout", Tier: "deep", Budget: 5},
	}
	for name, q := range questions {
		a, err := viaMCP.Ask(q)
		if err != nil {
			t.Fatalf("%s via MCP: %v", name, err)
		}
		b, err := viaCLI.Ask(q)
		if err != nil {
			t.Fatalf("%s via CLI: %v", name, err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s: MCP and CLI answers differ\nMCP: %.400s\nCLI: %.400s", name, a, b)
			continue
		}
		golden := filepath.Join("testdata", "gira-query", name+".json")
		if *updateGolden {
			if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(golden, a, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%s: %v (run with -update-golden)", name, err)
		}
		if !bytes.Equal(a, want) {
			t.Errorf("%s: answer differs from %s", name, golden)
		}
	}

	// The error vocabulary is part of the contract.
	nf, _ := viaCLI.Ask(questions["error-node-not-found"])
	if !bytes.Contains(nf, []byte(`"code": "node_not_found"`)) || !bytes.Contains(nf, []byte(`"service:cart"`)) {
		t.Fatalf("node_not_found must carry locate suggestions:\n%s", nf)
	}
	missing, _ := skills.Querier{Index: filepath.Join(t.TempDir(), "none.json"), CLI: viaCLI.CLI}.Ask(questions["impact-checkout"])
	if !bytes.Contains(missing, []byte(`"code": "index_missing"`)) {
		t.Fatalf("missing index must be index_missing:\n%s", missing)
	}
}

// The gira-query SKILL.md documents the same routing the procedure implements.
func TestGiraQuerySkillDocumentsEveryRoute(t *testing.T) {
	cat, err := skills.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	var body string
	for _, s := range cat {
		if s.Name == "gira-query" {
			body = s.Body
		}
	}
	for _, needle := range []string{"get_impact_briefing", "get_context_slice", "explain_evidence", "locate_boundary", "get_tour", "get_pr_impact", "diff_maps",
		"index_missing", "index_stale", "node_not_found", "budget_exceeded", "--format json"} {
		if !strings.Contains(body, needle) {
			t.Errorf("gira-query SKILL.md does not mention %s", needle)
		}
	}
}
