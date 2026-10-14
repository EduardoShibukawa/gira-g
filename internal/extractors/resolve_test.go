package extractors

import (
	"context"
	"fmt"
	"strings"
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

const producerGo = `package kafka

import (
	"os"

	"github.com/IBM/sarama"
)

var Topic = getTopic()

func getTopic() string {
	if t := os.Getenv("KAFKA_TOPIC"); t != "" {
		return t
	}
	return "orders"
}

func Create(brokers []string) (sarama.AsyncProducer, error) {
	return sarama.NewAsyncProducer(brokers, nil)
}
`

const mainGo = `package main

import (
	"os"

	"github.com/IBM/sarama"
	"example.com/shop/src/checkout/kafka"
)

type svc struct{ broker string }

func main() {
	s := &svc{}
	s.broker = os.Getenv("KAFKA_ADDR")
	p, _ := kafka.Create([]string{s.broker})
	p.Input() <- &sarama.ProducerMessage{Topic: kafka.Topic}
}
`

func kafkaIndex(t *testing.T, files ...TargetFile) map[string]domain.Relationship {
	t.Helper()
	all := append([]TargetFile{
		{Repo: "r", Path: "src/checkout/kafka/producer.go", Content: []byte(producerGo)},
		{Repo: "r", Path: "src/checkout/main.go", Content: []byte(mainGo)},
	}, files...)
	mm, err := NewPipeline(NewComposeExtractor(), NewEnvFileExtractor(), NewGoASTExtractor()).ProcessFiles(context.Background(), "t", all)
	if err != nil {
		t.Fatal(err)
	}
	rels := map[string]domain.Relationship{}
	for _, r := range mm.Relationships {
		if strings.HasPrefix(r.TargetID, "contract:kafka:") {
			rels[r.SourceID+" "+string(r.Verb)+" "+r.TargetID] = r
		}
	}
	return rels
}

func origins(r domain.Relationship) string {
	var out []string
	for _, ev := range r.Evidence {
		out = append(out, fmt.Sprintf("%s@%s#L%d", ev.Origin, ev.Locator.Path, ev.Locator.StartLine))
	}
	return strings.Join(out, " ")
}

func TestKafkaTopicFromComposeWiring(t *testing.T) {
	compose := "services:\n  checkout:\n    environment:\n      - KAFKA_ADDR\n      - KAFKA_TOPIC\n  kafka:\n    image: kafka\n  accounting:\n    environment:\n      - KAFKA_TOPIC\n  audit:\n    environment:\n      KAFKA_TOPIC: payments\n"
	rels := kafkaIndex(t,
		TargetFile{Repo: "r", Path: "compose.yaml", Content: []byte(compose)},
		TargetFile{Repo: "r", Path: ".env", Content: []byte("KAFKA_HOST=kafka\nKAFKA_ADDR=${KAFKA_HOST}:9092\nKAFKA_TOPIC=orders\n")},
	)
	want := map[string]string{
		"service:checkout provides contract:kafka:orders":   "extracted@src/checkout/main.go#L16 declared@compose.yaml#L5",
		"service:kafka provides contract:kafka:orders":      "declared@compose.yaml#L4",
		"service:accounting consumes contract:kafka:orders": "inferred@compose.yaml#L10",
	}
	if len(rels) != len(want) {
		t.Fatalf("got %v", rels)
	}
	for k, o := range want {
		if got := origins(rels[k]); got != o {
			t.Errorf("%s: evidence %q, want %q", k, got, o)
		}
	}
}

func TestKafkaTopicFallsBackToTheCodeDefault(t *testing.T) {
	rels := kafkaIndex(t)
	if got := origins(rels["service:checkout provides contract:kafka:orders"]); got != "extracted@src/checkout/main.go#L16 extracted@src/checkout/kafka/producer.go#L15" {
		t.Fatalf("without compose the code default names the topic: %v", rels)
	}
	if len(rels) != 1 {
		t.Errorf("no wiring, so no broker and no consumers: %v", rels)
	}
}

func TestKafkaTopicThatDoesNotResolveEmitsNothing(t *testing.T) {
	src := strings.Replace(mainGo, "Topic: kafka.Topic", "Topic: pick()", 1)
	mm, err := NewPipeline(NewGoASTExtractor()).ProcessFiles(context.Background(), "t", []TargetFile{{Repo: "r", Path: "src/checkout/main.go", Content: []byte(src)}})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range mm.Nodes {
		if strings.HasPrefix(n.ID, "contract:kafka:") {
			t.Fatalf("an unresolved topic must not be named: %s", n.ID)
		}
	}
}

func TestHTTPCallsAndEnvWiring(t *testing.T) {
	checkout := `package main

import (
	"net/http"
	"os"
)

type svc struct{ emailAddr, quoteAddr, unknownAddr string }

func mustMapEnv(target *string, key string) { *target = os.Getenv(key) }

func (s *svc) run() {
	mustMapEnv(&s.emailAddr, "EMAIL_ADDR")
	s.quoteAddr = os.Getenv("QUOTE_ADDR")
	mustMapEnv(&s.unknownAddr, "NOT_WIRED_ADDR")
	http.NewRequest("POST", s.emailAddr+"/confirm", nil)
	http.Post(s.emailAddr+"/bounce", "application/json", nil)
	http.Get(s.quoteAddr)
	http.Get(s.unknownAddr + "/x")
}
`
	compose := `services:
  checkout:
    environment:
      - EMAIL_ADDR
      - QUOTE_ADDR
      - NOT_WIRED_ADDR
      - OTEL_SERVICE_NAME=email
      - SELF_ADDR=checkout:1
  email:
    environment:
      - CACHE_HOST=redis
  quote:
    image: quote
  redis:
    image: redis
`
	mm, err := NewPipeline(NewComposeExtractor(), NewEnvFileExtractor(), NewGoASTExtractor()).ProcessFiles(context.Background(), "t", []TargetFile{
		{Repo: "r", Path: "compose.yaml", Content: []byte(compose)},
		{Repo: "r", Path: ".env", Content: []byte("EMAIL_ADDR=http://email:8080\nQUOTE_ADDR=quote:9\n")},
		{Repo: "r", Path: "src/checkout/main.go", Content: []byte(checkout)},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range mm.Relationships {
		got[r.SourceID+" "+string(r.Verb)+" "+r.TargetID] = fmt.Sprintf("%v %s", r.Metadata, origins(r))
	}
	want := map[string]string{
		"service:checkout consumes service:email": "map[protocol:http routes:[/bounce /confirm]] extracted@src/checkout/main.go#L16 extracted@src/checkout/main.go#L17 declared@compose.yaml#L4",
		"service:checkout consumes service:quote": "map[protocol:http] extracted@src/checkout/main.go#L18 declared@compose.yaml#L5",
		"service:email accesses service:redis":    "map[] declared@compose.yaml#L11",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got\n%v\nwant\n%v\n(NOT_WIRED_ADDR has no value; OTEL_SERVICE_NAME is not an address; SELF_ADDR is the service itself)", got, want)
	}
}

// SC2 (J2): a depends_on-only edge is start ordering, not consumption: it
// must never reach the Macro Map. A corroborated edge (env wiring here)
// must survive with both lines cited. Test-first: fails while every
// depends_on becomes consumes.
func TestDependsOnNeedsCorroboration(t *testing.T) {
	compose := `services:
  frontend:
    image: demo/frontend:1.0
    depends_on:
      - quote
  quote:
    image: demo/quote:1.0
  backend:
    image: demo/backend:1.0
    depends_on:
      - db-api
    environment:
      - API_URL=http://db-api:8000
  db-api:
    image: demo/db-api:1.0
  app:
    image: demo/app:1.0
    depends_on:
      - cache
  cache:
    image: redis:7
`
	files := []TargetFile{
		{Repo: "r", CommitSHA: "c", Path: "compose.yaml", Content: []byte(compose)},
	}
	mm, err := NewPipeline(NewComposeExtractor()).ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}
	rels := map[string]domain.Relationship{}
	for _, rel := range mm.Relationships {
		rels[rel.SourceID+"\x00"+rel.TargetID] = rel
	}
	if _, ok := rels["service:frontend\x00service:quote"]; ok {
		t.Fatalf("ordering-only frontend->quote must not reach the map")
	}
	corroborated, ok := rels["service:backend\x00service:db-api"]
	if !ok {
		t.Fatalf("corroborated backend->db-api must survive")
	}
	if len(corroborated.Evidence) < 2 {
		t.Fatalf("corroborated edge must cite both lines, got %d evidence records", len(corroborated.Evidence))
	}
	infra, ok := rels["service:app\x00service:cache"]
	if !ok || infra.Verb != domain.VerbAccesses {
		t.Fatalf("infrastructure wiring must stay accesses")
	}
}
