package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/archdiff"
	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// perturbedIndex indexes the Perturbed Variant of the hermetic corpus.
func perturbedIndex(t *testing.T) (*eval.Recipe, string) {
	t.Helper()
	root := filepath.Join("..", "..")
	_, c, err := eval.Hermetic(root)
	if err != nil {
		t.Fatal(err)
	}
	recipe, err := eval.LoadRecipe(filepath.Join(root, "fixtures", "evaluation", "reference-corpus", "perturbation.json"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "perturbed")
	if err := corpus.Perturb(c.Dir, dir, recipe.Perturbation); err != nil {
		t.Fatal(err)
	}
	c.Dir = dir
	idx, err := c.Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := idx.MacroMap.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "perturbed.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return recipe, path
}

// SC9: diffing the published and perturbed indexes reports exactly the renames and
// altered edges the perturbation table declares, with no other change.
func TestSC9ArchDiffMatchesPerturbationTable(t *testing.T) {
	_, published := hermeticIndex(t)
	recipe, perturbed := perturbedIndex(t)

	out, err := runCLI(t, "diff", published, perturbed)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := runCLI(t, "diff", published, perturbed)
	if !bytes.Equal(out, again) {
		t.Fatal("diff is not byte-identical across runs")
	}
	var d archdiff.Diff
	if err := json.Unmarshal(out, &d); err != nil {
		t.Fatal(err)
	}

	// Declared renames: every published node whose id the rename table changes.
	pub, err := loadIndex(published)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{}
	for _, n := range pub.Nodes {
		if to := recipe.Apply(n.ID); to != n.ID {
			want[n.ID] = to
		}
	}
	got := map[string]string{}
	for _, r := range d.Renamed {
		got[r.From] = r.To
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("renames differ from the table\n got %v\nwant %v", got, want)
	}

	var wantEdges, gotEdges []string
	for _, e := range recipe.IndexChanges.RemovedEdges {
		wantEdges = append(wantEdges, e.Source+" "+e.Verb+" "+e.Target)
	}
	for _, e := range d.RemovedEdges {
		gotEdges = append(gotEdges, e.String())
	}
	sort.Strings(wantEdges)
	if !reflect.DeepEqual(gotEdges, wantEdges) {
		t.Fatalf("removed edges\n got %v\nwant %v", gotEdges, wantEdges)
	}

	extra := map[string]int{
		"added_nodes": len(d.AddedNodes), "removed_nodes": len(d.RemovedNodes), "added_edges": len(d.AddedEdges),
		"changed_contracts": len(d.ChangedContracts), "newly_unsupported": len(d.NewlyUnsupported), "no_longer_unsupported": len(d.NoLongerUnsupported),
	}
	for k, n := range extra {
		if n != 0 {
			t.Errorf("unexpected %s: %d", k, n)
		}
	}
	t.Logf("%d renames, %d removed edges, %d nodes with moved evidence (informational)", len(d.Renamed), len(d.RemovedEdges), len(d.EvidenceMoved))

	mm, _ := loadIndex(published)
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))
	text, isErr := mcpCall(t, server, "diff_maps", map[string]any{"index_a": published, "index_b": perturbed})
	if isErr || text != string(out) {
		t.Fatalf("diff_maps differs from gira diff: %s", text)
	}

	md, err := runCLI(t, "diff", published, perturbed, "--format", "md")
	if err != nil || !bytes.Contains(md, []byte("`service:checkout` consumes `service:currency`")) {
		t.Fatalf("markdown changelog missing the removed edge: %v\n%s", err, md)
	}
}

func TestArchDiffOfIdenticalIndexesIsEmpty(t *testing.T) {
	_, published := hermeticIndex(t)
	out, err := runCLI(t, "diff", published, published, "--format", "md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("No architectural changes.")) {
		t.Fatalf("identical indexes must diff empty:\n%s", out)
	}
}
