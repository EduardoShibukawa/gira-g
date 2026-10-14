package extractors

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestEnvFileHintsReadValuesAsComposeDoes(t *testing.T) {
	content := "# comment\n\nA=1\nexport B=two\nC=\"quoted # kept\"\nD=plain    # trailing comment\nE=${A}:x\nnot a var\n"
	hints, err := NewEnvFileExtractor().Hints(context.Background(), TargetFile{Repo: "r", Path: ".env", Content: []byte(content)})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	lines := map[string]int{}
	for _, h := range hints {
		got[h.Key], lines[h.Key] = h.Value, h.Locator.StartLine
	}
	want := map[string]string{"A": "1", "B": "two", "C": "quoted # kept", "D": "plain", "E": "${A}:x"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	if lines["A"] != 3 || lines["E"] != 7 {
		t.Errorf("hints must carry their line: %v", lines)
	}
}

func TestInterpolateResolvesOrRefuses(t *testing.T) {
	env := newEnvScopes([]Hint{
		{Kind: HintEnvValue, Locator: domain.Locator{Repo: "r", Path: ".env"}, Key: "HOST", Value: "frontend"},
		{Kind: HintEnvValue, Locator: domain.Locator{Repo: "r", Path: ".env"}, Key: "PORT", Value: "8080"},
		{Kind: HintEnvValue, Locator: domain.Locator{Repo: "r", Path: ".env"}, Key: "ADDR", Value: "${HOST}:${PORT}"},
		{Kind: HintEnvValue, Locator: domain.Locator{Repo: "r", Path: ".env"}, Key: "EMPTY", Value: ""},
		{Kind: HintEnvValue, Locator: domain.Locator{Repo: "r", Path: "sub/.env"}, Key: "HOST", Value: "other"},
	})
	for _, tc := range []struct {
		dir, in, want string
		ok            bool
	}{
		{".", "${ADDR}", "frontend:8080", true},
		{".", "http://$HOST/x", "http://frontend/x", true},
		{".", "${MISSING:-def}", "def", true},
		{".", "${EMPTY:-def}", "def", true},
		{".", "${EMPTY-def}", "", true},
		{".", "${MISSING}", "", false},
		{".", "a${HOST}${MISSING}", "", false},
		{"sub", "${HOST}", "other", true},
		{"elsewhere", "${HOST}", "", false},
	} {
		got, ok := env.interpolate("r", tc.dir, tc.in)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("interpolate(%q in %s) = %q, %v; want %q, %v", tc.in, tc.dir, got, ok, tc.want, tc.ok)
		}
	}
}

// A Dockerfile is attributed to the compose service that builds it; one that no
// compose service builds, or whose dockerfile variable does not resolve, keeps
// the directory rule.
func TestPipelineAttributesDockerfilesToComposeServices(t *testing.T) {
	compose := `services:
  cart:
    build:
      context: ./
      dockerfile: ${CART_DOCKERFILE}
  ghost:
    build:
      context: ./
      dockerfile: ${UNDEFINED_DOCKERFILE}
  web:
    build: ./src/web
`
	files := []TargetFile{
		{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(compose)},
		{Repo: "r", CommitSHA: "c", Path: ".env", Content: []byte("CART_DOCKERFILE=./src/cart/src/Dockerfile\n")},
		{Repo: "r", CommitSHA: "c", Path: "src/cart/src/Dockerfile", Content: []byte("FROM alpine\n")},
		{Repo: "r", CommitSHA: "c", Path: "src/web/Dockerfile", Content: []byte("FROM alpine\n")},
		{Repo: "r", CommitSHA: "c", Path: "src/phantom/Dockerfile", Content: []byte("FROM alpine\n")},
		{Repo: "r", CommitSHA: "c", Path: "tools/lint/Dockerfile", Content: []byte("FROM alpine\n")},
	}
	mm, err := NewPipeline(NewComposeExtractor(), NewEnvFileExtractor(), NewDockerfileExtractor()).ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}
	nodes := map[string]domain.Node{}
	for _, n := range mm.Nodes {
		nodes[n.ID] = n
	}
	cites := func(id, p string) bool {
		for _, ev := range nodes[id].Evidence {
			if ev.Locator.Path == p {
				return true
			}
		}
		return false
	}
	if _, ok := nodes["service:src"]; ok || !cites("service:cart", "src/cart/src/Dockerfile") || !cites("service:cart", "compose.yaml") {
		t.Errorf("cart's Dockerfile must merge into service:cart: %v", nodes)
	}
	if !cites("service:web", "src/web/Dockerfile") {
		t.Errorf("build: <context> uses <context>/Dockerfile: %v", nodes["service:web"])
	}
	if cites("service:ghost", "src/phantom/Dockerfile") || !cites("service:phantom", "src/phantom/Dockerfile") {
		t.Errorf("an undefined dockerfile variable attributes nothing: %v", nodes)
	}
	if !cites("service:lint", "tools/lint/Dockerfile") {
		t.Errorf("a Dockerfile no compose service builds keeps the directory rule: %v", nodes)
	}
	if len(mm.Unsupported) != 0 {
		t.Errorf(".env must be claimed by an extractor: %v", mm.Unsupported)
	}
}

// An unresolvable dockerfile variable never renames: the Dockerfile keeps its
// directory name even when a compose service could plausibly own it.
func TestUnresolvedDockerfileVariableIsNotGuessed(t *testing.T) {
	files := []TargetFile{
		{Repo: "r", Path: "compose.yaml", Content: []byte("services:\n  cart:\n    build:\n      dockerfile: ${CART_DOCKERFILE}\n")},
		{Repo: "r", Path: "src/cart/src/Dockerfile", Content: []byte("FROM alpine\n")},
	}
	mm, err := NewPipeline(NewComposeExtractor(), NewDockerfileExtractor()).ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, n := range mm.Nodes {
		ids[n.ID] = true
	}
	if !ids["service:src"] || !ids["service:cart"] {
		t.Fatalf("without .env nothing is attributed: %v", ids)
	}
}
