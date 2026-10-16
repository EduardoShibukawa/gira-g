package extractors

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// Self-fidelity golden: a pinned single-module tree proving single-repo
// maps stay first-class. Regenerate deliberately with UPDATE_SNAPSHOTS=1;
// the committed snapshot is the evolving record of what "correct" looks
// like at each approval.
func TestSelfFidelityGolden(t *testing.T) {
	files := []TargetFile{
		{Repo: "selftest", CommitSHA: "pinned", Path: "go.mod", Content: []byte("module example.com/acme\n\ngo 1.24\n")},
		{Repo: "selftest", CommitSHA: "pinned", Path: "cmd/acme/main.go", Content: []byte("package main\n\nimport (\n\t\"fmt\"\n\t\"example.com/acme/internal/slices\"\n)\n\nfunc main() { fmt.Println(slices.Hello()) }\n")},
		{Repo: "selftest", CommitSHA: "pinned", Path: "internal/slices/slices.go", Content: []byte("package slices\n\nimport \"strings\"\n\nfunc Hello() string { return strings.ToUpper(\"hi\") }\n")},
	}
	mm, err := DefaultPipeline().ProcessFiles(context.Background(), "selftest-system", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}
	type rel struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Verb   string `json:"verb"`
	}
	type node struct {
		ID     string         `json:"id"`
		Kind   domain.NodeType `json:"kind"`
		Detail string         `json:"origin_detail,omitempty"`
	}
	// NOTE (known noise, follow-up task): non-main Go files mint a
	// fallback Service node per top-level dir (e.g. service:internal).
	// Minting discipline is out of scope here; this golden pins the
	// additive promise (intra-module marking) only.
	out := struct {
		Note  string `json:"_note"`
		Nodes []node `json:"nodes"`
		Rels  []rel  `json:"rels"`
	}{
		Note: "service:internal is fallback minting noise; see follow-up on minting discipline",
	}
	for _, n := range mm.Nodes {
		d, _ := n.Metadata["origin_detail"].(string)
		out.Nodes = append(out.Nodes, node{ID: n.ID, Kind: n.Kind, Detail: d})
	}
	for _, r := range mm.Relationships {
		out.Rels = append(out.Rels, rel{Source: r.SourceID, Target: r.TargetID, Verb: string(r.Verb)})
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].ID < out.Nodes[j].ID })
	sort.Slice(out.Rels, func(i, j int) bool {
		if out.Rels[i].Source != out.Rels[j].Source {
			return out.Rels[i].Source < out.Rels[j].Source
		}
		return out.Rels[i].Target < out.Rels[j].Target
	})
	got, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got = append(got, '\n')
	snapPath := filepath.Join("testdata", "self_fidelity_snapshot.json")
	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := os.MkdirAll(filepath.Dir(snapPath), 0755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(snapPath, got, 0644); err != nil {
			t.Fatalf("snapshot write failed: %v", err)
		}
		return
	}
	want, err := os.ReadFile(snapPath)
	if err != nil {
		t.Fatalf("snapshot read failed (run UPDATE_SNAPSHOTS=1 to create): %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("golden mismatch:\n got: %s\nwant: %s", got, want)
	}
	// Pin the single-repo promise: own package marked, service present.
	found := false
	for _, n := range out.Nodes {
		if n.ID == "library:example.com/acme/internal/slices" && n.Detail == "intra-module" {
			found = true
		}
	}
	if !found {
		t.Fatalf("intra-module library missing from golden: %s", got)
	}
}
