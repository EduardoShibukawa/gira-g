package indexer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/corpus"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

func build(t *testing.T, src string) []byte {
	t.Helper()
	ctx := context.Background()
	repo := filepath.Join(t.TempDir(), "repo")
	sha, err := corpus.Materialize(ctx, src, repo)
	if err != nil {
		t.Fatal(err)
	}
	cm, err := gitcache.NewCacheManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := &manifest.Manifest{Version: 1, TargetSystem: manifest.TargetSystemMeta{Name: "t"},
		Repositories: []manifest.RepositoryDeclaration{{ID: "r", URL: repo, SHA: sha, PathScope: "."}}}
	mm, err := indexer.Build(ctx, m, cm, indexer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	for _, n := range mm.Nodes {
		for _, ev := range n.Evidence {
			if !ev.DiscoveredAt.Equal(want) {
				t.Fatalf("evidence on %s stamped %v, want the commit time %v", n.ID, ev.DiscoveredAt, want)
			}
		}
	}
	j, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	return j
}

// Two cold indexes of the same commit are byte-identical (spec SC4).
func TestBuildIsDeterministicAcrossColdRuns(t *testing.T) {
	src := t.TempDir()
	files := map[string]string{
		"proto/order.proto":  "syntax = \"proto3\";\npackage demo;\nservice OrderService {\n  rpc Place(Req) returns (Resp) {}\n}\n",
		"src/order/main.go":  "package main\n\nimport pb \"example.com/demo/pb\"\n\nfunc main() { pb.RegisterOrderServiceServer(nil, nil) }\n",
		"docker-compose.yml": "services:\n  order:\n    image: order\n",
	}
	for p, c := range files {
		full := filepath.Join(src, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if a, b := build(t, src), build(t, src); string(a) != string(b) {
		t.Fatal("cold indexes of the same commit differ")
	}
}
