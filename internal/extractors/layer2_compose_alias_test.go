package extractors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// F4 (J2 follow-up): depends_on reached through a YAML alias (`*x`) or a
// merge key (`<<: *x`), in list or map form, is still ordering-only and must
// be pruned; a corroborated aliased edge must survive (no recall loss).
func TestDependsOnYAMLAliasAndMergeKey(t *testing.T) {
	compose := `x-deps-list: &deps-list
  - quote
x-deps-map: &deps-map
  quote:
    condition: service_started
x-base-list: &base-list
  depends_on:
    - quote
x-base-map: &base-map
  depends_on:
    quote:
      condition: service_started
x-wired: &wired
  depends_on:
    - db-api
services:
  alias-list:
    image: demo/alias-list:1.0
    depends_on: *deps-list
  alias-map:
    image: demo/alias-map:1.0
    depends_on: *deps-map
  merge-list:
    <<: *base-list
    image: demo/merge-list:1.0
  merge-map:
    <<: [*base-map]
    image: demo/merge-map:1.0
  backend:
    <<: *wired
    image: demo/backend:1.0
    environment:
      - API_URL=http://db-api:8000
  quote:
    image: demo/quote:1.0
  db-api:
    image: demo/db-api:1.0
`
	dir := t.TempDir()
	p := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(p, []byte(compose), 0o644); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	files := []TargetFile{{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: content}}

	// Hints: one depends_on hint per aliased / merged dependency.
	hints, err := NewComposeExtractor().Hints(context.Background(), files[0])
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, h := range hints {
		if h.Kind == HintDependsOn {
			got[h.Service+"->"+h.Value] = true
		}
	}
	for _, want := range []string{
		"service:alias-list->quote",
		"service:alias-map->quote",
		"service:merge-list->quote",
		"service:merge-map->quote",
		"service:backend->db-api",
	} {
		if !got[want] {
			t.Errorf("missing depends_on hint %s", want)
		}
	}

	mm, err := NewPipeline(NewComposeExtractor()).ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}
	rels := map[string]domain.Relationship{}
	for _, rel := range mm.Relationships {
		rels[rel.SourceID+"\x00"+rel.TargetID] = rel
	}
	for _, src := range []string{"alias-list", "alias-map", "merge-list", "merge-map"} {
		if _, ok := rels["service:"+src+"\x00service:quote"]; ok {
			t.Errorf("ordering-only %s->quote (via YAML alias/merge) must not reach the map", src)
		}
	}
	corroborated, ok := rels["service:backend\x00service:db-api"]
	if !ok {
		t.Fatalf("corroborated backend->db-api (depends_on via merge key) must survive")
	}
	if len(corroborated.Evidence) < 2 {
		t.Fatalf("corroborated edge must cite both lines, got %d evidence records", len(corroborated.Evidence))
	}
}
