package layer1

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/bench/internal/benchtest"
)

var repoRoot = filepath.Join("..", "..")

func loadGT(t *testing.T) *GroundTruth {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, "fixtures", "evaluation", "reference-corpus", "ground-truth.v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gt GroundTruth
	if err := json.Unmarshal(data, &gt); err != nil {
		t.Fatal(err)
	}
	return &gt
}

// hermeticCorpus returns the indexed hermetic corpus shared by bench tests.
func hermeticCorpus(t *testing.T) (repo, sha, gira, index string) {
	c := benchtest.Hermetic(t)
	return c.Repo, c.SHA, c.Gira, c.Index
}

// SC14: in mock-model mode the runner executes all 12 tasks for arms A and B
// offline, at least 5 times each, and every telemetry record is valid against the
// schema with every field of spec B present.
func TestSC14MockModeRunsAllTasksForArmsAAndB(t *testing.T) {
	gt := loadGT(t)
	if len(gt.Tasks) != 12 {
		t.Fatalf("want the 12 ADR 0007 tasks, got %d", len(gt.Tasks))
	}
	repo, sha, gira, index := hermeticCorpus(t)
	ctx := context.Background()
	mcp, err := StartMCP(ctx, gira, "serve", "mcp", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	defer mcp.Close()

	base := BaseTools{Root: repo}
	arms := []Arm{{ID: "A", Base: base}, {ID: "B", Base: base, MCP: mcp}}
	r := Runner{Model: MockModel{}, GT: gt, CorpusSHA: sha, TranscriptDir: filepath.Join(t.TempDir(), "out", "transcripts")}

	const reps = 5
	cells := map[string][]float64{}
	records := 0
	for _, arm := range arms {
		for _, task := range gt.Tasks {
			for rep := 1; rep <= reps; rep++ {
				tel := r.RunTask(ctx, arm, task, rep)
				data, err := json.Marshal(tel)
				if err != nil {
					t.Fatal(err)
				}
				if err := ValidateTelemetry(data); err != nil {
					t.Fatalf("%s/%s/%d: %v\n%s", arm.ID, task.ID, rep, err, data)
				}
				if !tel.Answered || tel.Error != "" {
					t.Fatalf("%s/%s: not answered: %s", arm.ID, task.ID, tel.Error)
				}
				wantTool := map[string]string{"A": "grep", "B": "get_impact_briefing"}[arm.ID]
				if tel.ToolCallsByTool[wantTool] != 1 || tel.Turns != 2 || tel.TokensInput == 0 || tel.TokensOutput == 0 {
					t.Fatalf("%s/%s: unexpected loop telemetry %+v", arm.ID, task.ID, tel)
				}
				cells[arm.ID+"/"+task.ID] = append(cells[arm.ID+"/"+task.ID], tel.F1)
				records++
			}
		}
	}
	if records != 2*12*reps {
		t.Fatalf("want %d records, got %d", 2*12*reps, records)
	}
	var meanA, meanB float64
	for k, xs := range cells {
		s := Summarize(xs)
		if s.N != reps || s.IQR != 0 {
			t.Fatalf("%s: mock runs must be identical across repetitions: %+v", k, s)
		}
		if strings.HasPrefix(k, "A/") {
			meanA += s.Median / 12
		} else {
			meanB += s.Median / 12
		}
	}
	t.Logf("mock policy, median F1 averaged over tasks: arm A %.2f, arm B %.2f (tests the loop, not a model)", meanA, meanB)
	if entries, _ := os.ReadDir(filepath.Join(r.TranscriptDir, "B", "mock-policy/1")); len(entries) != 12*reps {
		t.Fatalf("want one transcript per run in the output dir, got %d", len(entries))
	}
}

// SC14: the scorer matches a hand-computed fixture.
func TestScorerMatchesHandComputedFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "scorer-fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fx struct {
		Files    map[string]string `json:"files"`
		Ignore   []string          `json:"ignore"`
		Universe []string          `json:"universe"`
		Task     Task              `json:"task"`
		Answer   Answer            `json:"answer"`
		Want     Score             `json:"want"`
	}
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for p, c := range fx.Files {
		_ = os.WriteFile(filepath.Join(root, p), []byte(c), 0o644)
	}
	set := func(xs []string) map[string]bool {
		m := map[string]bool{}
		for _, x := range xs {
			m[x] = true
		}
		return m
	}
	got := ScoreAnswer(fx.Answer, fx.Task, set(fx.Ignore), set(fx.Universe), BaseTools{Root: root}.CheckLocator)
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if !near(got.Precision, fx.Want.Precision) || !near(got.Recall, fx.Want.Recall) || !near(got.F1, fx.Want.F1) ||
		!near(got.LocatorValidity, fx.Want.LocatorValidity) || got.LocatorsTotal != fx.Want.LocatorsTotal ||
		!reflect.DeepEqual(got.HallucinatedNodes, fx.Want.HallucinatedNodes) {
		t.Fatalf("score %+v, want %+v", got, fx.Want)
	}
}

func TestValidateTelemetryRejectsIncompleteRecords(t *testing.T) {
	good, _ := json.Marshal(Telemetry{Schema: TelemetrySchemaID, ToolCallsByTool: map[string]int{}, Hallucinated: []string{}})
	if err := ValidateTelemetry(good); err != nil {
		t.Fatalf("complete record rejected: %v", err)
	}
	var m map[string]any
	_ = json.Unmarshal(good, &m)
	delete(m, "cache_read_tokens")
	bad, _ := json.Marshal(m)
	if err := ValidateTelemetry(bad); err == nil {
		t.Fatal("a record without cache_read_tokens must be rejected")
	}
	m["cache_read_tokens"] = 1.5
	bad, _ = json.Marshal(m)
	if err := ValidateTelemetry(bad); err == nil {
		t.Fatal("a non-integer token count must be rejected")
	}
}

func TestSummarize(t *testing.T) {
	s := Summarize([]float64{5, 1, 3, 2, 4})
	if s.Median != 3 || s.IQR != 2 || s.N != 5 {
		t.Fatalf("got %+v", s)
	}
}

// Paid endpoints are refused without an approval reference; local ones work.
func TestOpenAIModelGuardAndParse(t *testing.T) {
	_, err := OpenAIModel{Endpoint: "https://api.example.com", Model: "frontier"}.Complete(context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "approval") {
		t.Fatalf("a non-local endpoint must be refused without approval: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["model"] != "llama3.1:8b" || r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "bad request", 400)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"","tool_calls":[{"id":"c1","type":"function","function":{"name":"grep","arguments":"{\"pattern\":\"cart\"}"}}]}}],"usage":{"prompt_tokens":120,"completion_tokens":9,"prompt_tokens_details":{"cached_tokens":64}}}`))
	}))
	defer srv.Close()
	r, err := OpenAIModel{Endpoint: srv.URL, Model: "llama3.1:8b"}.Complete(context.Background(), []Message{{Role: "user", Content: "x"}}, BaseTools{}.Specs())
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ToolCalls) != 1 || r.ToolCalls[0].Arguments["pattern"] != "cart" || r.Usage != (Usage{Input: 120, Output: 9, CacheRead: 64}) {
		t.Fatalf("parsed reply %+v", r)
	}
}

func TestBaseToolsStayInsideTheCorpus(t *testing.T) {
	if _, err := (BaseTools{Root: t.TempDir()}).Run("read", map[string]any{"path": "../../etc/passwd"}); err == nil {
		t.Fatal("read outside the corpus must be refused")
	}
}
