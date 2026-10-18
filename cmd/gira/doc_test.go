package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/tokens"
)

// The doc fixture mirrors the F2 intra-service corpus (internal/slices
// SCDocIntraService): a Go service with an entrypoint, symbols, contracts,
// imports and calls, a runnable manifest task, and an unsupported file the
// extractors refuse to guess about.
const scDocGoMod = `module example.com/acme

go 1.24

require github.com/external/dep v1.4.0
`

const scDocMainGo = `package main

import (
	"fmt"

	"github.com/external/dep"
	pb "example.com/acme/genproto/orders"
	"example.com/acme/internal/slices"
)

type ordersServer struct {
	pb.UnimplementedOrdersServer
}

type billingWiring struct {
	client pb.BillingClient
}

func main() {
	pb.RegisterOrdersServer(nil, &ordersServer{})
	fmt.Println(slices.Hello())
	fmt.Println(slices.Hello())
	dep.Process()
	w := &billingWiring{}
	w.client.Submit()
}
`

const scDocSlicesGo = `package slices

import "strings"

var warm = Hello()

type Server struct {
	Addr string
}

type Store interface {
	Get(key string) string
}

func Hello() string { return strings.ToUpper("hi") }

func hello() string { return "hi" }

func Count(items []string) int { return len(items) }

func (s Server) Start() string { return Hello() + s.Addr }
`

const scDocHelperPy = `# helper outside any extractor
# talks about the Billing settlement
print("helper")
`

const scDocMiseToml = `[tools]
go = "1.24.0"

[tasks.check]
description = "Run the Go test suite"
run = "go test -race ./..."
`

// docFixtureIndex indexes the doc fixture through the real pipeline and
// writes the MacroMap where the CLI can load it.
func docFixtureIndex(t *testing.T) string {
	t.Helper()
	files := []extractors.TargetFile{
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "go.mod", Content: []byte(scDocGoMod)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "cmd/acme/main.go", Content: []byte(scDocMainGo)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "internal/slices/slices.go", Content: []byte(scDocSlicesGo)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "cmd/acme/helper.py", Content: []byte(scDocHelperPy)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "mise.toml", Content: []byte(scDocMiseToml)},
	}
	mm, err := extractors.DefaultPipeline().ProcessFiles(context.Background(), "acme-system", files)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}
	data, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	path := filepath.Join(t.TempDir(), "macro-map.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write test index: %v", err)
	}
	return path
}

// docFactTokens collects every fact ID and evidence locator the JSON envelope
// carries: the parity contract is that the Markdown shows every one of them.
func docFactTokens(doc *slices.Doc) []string {
	var out []string
	add := func(s string) {
		if s != "" {
			out = append(out, s)
		}
	}
	add(doc.ServiceID)
	add(doc.Intra.ServiceLocator.String())
	for _, e := range doc.Intra.Entrypoints {
		add(e.ID)
		add(e.Locator.String())
	}
	for _, p := range doc.Intra.Packages {
		add(p.ID)
		add(p.Locator.String())
	}
	for _, s := range doc.Intra.Types {
		add(s.ID)
		add(s.Locator.String())
	}
	for _, f := range doc.Intra.Funcs {
		add(f.ID)
		add(f.Locator.String())
	}
	for _, c := range append(append([]slices.IntraContractRef(nil), doc.Intra.ProvidedContracts...), doc.Intra.ConsumedContracts...) {
		add(c.ContractID)
		add(c.ContractName)
		add(c.EdgeID)
		for _, l := range c.Locators {
			add(l.String())
		}
	}
	for _, im := range doc.Intra.IntraModuleImports {
		add(im.ID)
		add(im.Path)
		add(im.Locator.String())
	}
	for _, im := range doc.Intra.ExternalImports {
		add(im.ID)
		add(im.Path)
		add(im.Locator.String())
	}
	for _, im := range doc.Intra.StdlibImports {
		add(im.ID)
		add(im.Path)
		add(im.Locator.String())
	}
	for _, c := range doc.Intra.Calls {
		add(c.ID)
		add(c.Callee)
		add(c.TargetID)
		if c.External != nil {
			add(c.External.ImportPath)
			add(c.External.Name)
		}
		add(c.Locator.String())
	}
	for _, u := range doc.Intra.KnownUnknowns {
		add(u.Path)
		add(u.Locator.String())
		for _, j := range u.UnknownJoins {
			add(j)
		}
	}
	for _, rt := range doc.RunTasks {
		add(rt.ID)
		add(rt.Run)
		if rt.Locator.Path != "" {
			add(rt.Locator.String())
		}
	}
	for _, dir := range doc.ChangeDirs {
		add(dir)
	}
	if doc.Tour != nil {
		for _, s := range doc.Tour.Steps {
			for _, id := range s.NodeIDs {
				add(id)
			}
			add(s.WhyLocator)
			for _, l := range s.Read {
				add(l)
			}
		}
	}
	if doc.Flows != nil {
		for _, fl := range doc.Flows.Flows {
			for _, h := range fl.Steps {
				add(h.NodeID)
				add(h.Via)
				add(h.Locator)
			}
		}
	}
	return out
}

// SCDocParity (SC3): `doc --format md` and `--format json` cover the same
// facts — every node/edge/evidence ID the JSON envelope carries appears in
// the Markdown — and the plain Markdown path is golden-tested byte for byte.
func TestSCDocParity(t *testing.T) {
	index := docFixtureIndex(t)

	jsonOut, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index)
	if err != nil {
		t.Fatalf("doc json failed: %v", err)
	}
	var doc slices.Doc
	if err := json.Unmarshal(jsonOut, &doc); err != nil {
		t.Fatalf("doc json does not parse: %v\n%s", err, jsonOut)
	}
	if doc.ServiceID != "service:acme" {
		t.Fatalf("doc service_id = %q, want service:acme", doc.ServiceID)
	}
	if len(doc.Intra.Entrypoints) == 0 || len(doc.Intra.Funcs) == 0 {
		t.Fatalf("doc is missing the intra-service facts: %+v", doc.Intra)
	}
	if len(doc.Intra.ProvidedContracts) == 0 || len(doc.Intra.ConsumedContracts) == 0 {
		t.Fatalf("doc is missing the contract facts: %+v", doc.Intra)
	}
	if len(doc.Intra.KnownUnknowns) == 0 {
		t.Fatalf("doc is missing the Known Unknowns")
	}
	if len(doc.RunTasks) == 0 {
		t.Fatalf("doc is missing the runnable manifest tasks")
	}

	mdOut, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatalf("doc md failed: %v", err)
	}
	md := string(mdOut)
	for _, section := range []string{
		"# Service Doc:", "## Entrypoints", "## Symbols", "## Contracts Provided",
		"## Contracts Consumed", "## Dependencies: Imports", "## Calls",
		"## Run & Test", "## Where to Change", "## Guided Tour",
		"## Flows from", "## Known Unknowns",
	} {
		if !strings.Contains(md, section) {
			t.Errorf("md is missing section %q:\n%s", section, md)
		}
	}
	for _, tok := range docFactTokens(&doc) {
		if !strings.Contains(md, tok) {
			t.Errorf("md is missing the JSON fact %q", tok)
		}
	}

	htmlOut, err := runCLI(t, "doc", "service:acme", "--format", "html", "--index", index)
	if err != nil {
		t.Fatalf("doc html failed: %v", err)
	}
	htmlStr := string(htmlOut)
	if !strings.Contains(htmlStr, "<!DOCTYPE html>") {
		t.Errorf("html output missing <!DOCTYPE html>")
	}
	for _, bad := range []string{"<script src=", "<link ", `src="http`, `href="http`, "@import", "url(http"} {
		if strings.Contains(htmlStr, bad) {
			t.Errorf("doc html is not self-contained: contains %q", bad)
		}
	}
	for _, tok := range docFactTokens(&doc) {
		if !strings.Contains(htmlStr, tok) {
			t.Errorf("html is missing the JSON fact %q", tok)
		}
	}
	htmlAgain, err := runCLI(t, "doc", "service:acme", "--format", "html", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(htmlOut, htmlAgain) {
		t.Errorf("html output is not byte-identical on repeat")
	}

	// Repeat runs are byte-identical in both formats (the envelope carries
	// no wall-clock timestamps).
	jsonAgain, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(jsonOut, jsonAgain) {
		t.Errorf("json output is not byte-identical on repeat")
	}
	mdAgain, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mdOut, mdAgain) {
		t.Errorf("md output is not byte-identical on repeat")
	}

	// Plain (non-TTY / NO_COLOR) md is byte-identical across arms, carries
	// no ANSI escapes, and matches the golden.
	t.Setenv("NO_COLOR", "1")
	noColor, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	os.Unsetenv("NO_COLOR")
	plain, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(plain, []byte("\x1b[")) {
		t.Fatalf("plain output carries ANSI escapes:\n%q", plain)
	}
	if !bytes.Equal(plain, noColor) {
		t.Fatalf("NO_COLOR changed the plain output\n%q\n%q", plain, noColor)
	}
	golden := filepath.Join("testdata", "human", "doc-acme.md")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, plain, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update-golden)", err)
	}
	if !bytes.Equal(plain, want) {
		t.Fatalf("plain doc output drifted from %s", golden)
	}
}

// Without --format the doc is Markdown, byte-identical to --format md.
func TestSCDocDefaultFormat(t *testing.T) {
	index := docFixtureIndex(t)
	def, err := runCLI(t, "doc", "service:acme", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	md, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(def, md) {
		t.Fatalf("default format differs from --format md")
	}
}

// The indexed-commit banner goes to stderr for human output only, like the
// sibling commands; a checkout whose HEAD the index does not cite is stale.
func TestSCDocBannerAndStale(t *testing.T) {
	index := docFixtureIndex(t)

	var out, errOut bytes.Buffer
	if err := Run([]string{"doc", "service:acme", "--format", "md", "--index", index}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "indexed commit:") {
		t.Fatalf("missing the indexed-commit banner: %q", errOut.String())
	}

	var htmlOut, htmlErr bytes.Buffer
	if err := Run([]string{"doc", "service:acme", "--format", "html", "--index", index}, strings.NewReader(""), &htmlOut, &htmlErr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(htmlErr.String(), "indexed commit:") {
		t.Fatalf("html output missing the indexed-commit banner: %q", htmlErr.String())
	}

	var jsonOut, jsonErr bytes.Buffer
	if err := Run([]string{"doc", "service:acme", "--format", "json", "--index", index}, strings.NewReader(""), &jsonOut, &jsonErr); err != nil {
		t.Fatal(err)
	}
	if jsonErr.String() != "" {
		t.Fatalf("json output must keep stderr clean, got %q", jsonErr.String())
	}

	dir := t.TempDir()
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "unrelated")
	var staleOut, staleErr bytes.Buffer
	if err := Run([]string{"doc", "service:acme", "--format", "md", "--index", index, "--checkout", dir}, strings.NewReader(""), &staleOut, &staleErr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(staleErr.String(), "stale:") {
		t.Fatalf("expected a stale banner: %q", staleErr.String())
	}
}

// Unknown services error honestly; 
// unsupported formats error honestly.
func TestSCDocUsageErrors(t *testing.T) {
	index := docFixtureIndex(t)

	if _, err := runCLI(t, "doc", "service:nope", "--format", "json", "--index", index); err == nil {
		t.Errorf("expected error for unknown service, got nil")
	}
	if _, err := runCLI(t, "doc", "service:acme", "--format", "xml", "--index", index); err == nil || !strings.Contains(err.Error(), "unsupported doc format") {
		t.Errorf("expected an unsupported format error, got %v", err)
	}
	if _, err := runCLI(t, "doc", "--format", "json", "--index", index); err == nil {
		t.Errorf("expected error for missing service_id, got nil")
	}
}

// SCDocMCPParity (F3 MCP): `get_doc` returns byte-identical bytes to
// `doc --format json` — same engine, same facts/IDs — and carries the honest
// budget envelope only when asked. Unknown services and a missing service_id
// error honestly; a ceiling below the never-dropped parts refuses with
// budget_exceeded instead of cutting mid-JSON.
func TestSCDocMCPParity(t *testing.T) {
	index := docFixtureIndex(t)
	mm, err := loadIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer("gira", "test", slices.NewEngine(mm))

	cliOut, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index)
	if err != nil {
		t.Fatalf("doc json failed: %v", err)
	}
	text, isErr := mcpCall(t, server, "get_doc", map[string]any{"service_id": "service:acme"})
	if isErr || text != string(cliOut) {
		t.Fatalf("get_doc differs from the CLI doc json (isErr=%v)", isErr)
	}

	// Same facts: every fact ID the envelope carries parses, names the
	// focal service, and shows up in the Markdown twin too.
	var doc slices.Doc
	if err := json.Unmarshal([]byte(text), &doc); err != nil {
		t.Fatalf("get_doc output does not parse: %v", err)
	}
	if doc.ServiceID != "service:acme" {
		t.Fatalf("get_doc service_id = %q, want service:acme", doc.ServiceID)
	}
	if len(doc.Intra.Entrypoints) == 0 || len(doc.Intra.Funcs) == 0 {
		t.Fatalf("get_doc is missing the intra-service facts")
	}
	if doc.Budget != nil {
		t.Fatalf("unbudgeted get_doc must carry no budget envelope, got %+v", doc.Budget)
	}
	mdOut, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range docFactTokens(&doc) {
		if !strings.Contains(string(mdOut), tok) {
			t.Errorf("md twin is missing the get_doc fact %q", tok)
		}
	}

	// Budgeted call carries the honest envelope: limit echoed, estimated
	// matching the emitted bytes under the bytes/4 convention.
	limit := tokens.Estimate(cliOut) * 2
	budgeted, isErr := mcpCall(t, server, "get_doc", map[string]any{"service_id": "service:acme", "budget": limit})
	if isErr {
		t.Fatalf("budgeted get_doc failed")
	}
	var bdoc struct {
		ServiceID string         `json:"service_id"`
		Budget    *slices.Budget `json:"budget"`
	}
	if err := json.Unmarshal([]byte(budgeted), &bdoc); err != nil {
		t.Fatal(err)
	}
	if bdoc.Budget == nil {
		t.Fatalf("budgeted get_doc carries no budget envelope")
	}
	if bdoc.Budget.Limit != limit || bdoc.Budget.Estimated != tokens.Estimate([]byte(budgeted)) {
		t.Errorf("budget %+v, want limit %d estimated %d", *bdoc.Budget, limit, tokens.Estimate([]byte(budgeted)))
	}

	// Honest errors: unknown service, missing service_id, tight ceiling.
	if _, isErr := mcpCall(t, server, "get_doc", map[string]any{"service_id": "service:nope"}); !isErr {
		t.Errorf("expected error for unknown service, got success")
	}
	if _, isErr := mcpCall(t, server, "get_doc", map[string]any{}); !isErr {
		t.Errorf("expected error for missing service_id, got success")
	}
	tight, isErr := mcpCall(t, server, "get_doc", map[string]any{"service_id": "service:acme", "budget": 5})
	if !isErr || !strings.Contains(tight, "budget_exceeded") {
		t.Errorf("expected budget_exceeded for a tight ceiling, got isErr=%v %q", isErr, tight)
	}
}

// A --budget ceiling wraps the same facts in the budget envelope without
// dropping the focal service.
func TestSCDocBudgetEnvelope(t *testing.T) {
	index := docFixtureIndex(t)

	full, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(full, []byte(`"budget"`)) {
		t.Fatalf("unbudgeted output must carry no budget envelope")
	}
	limit := tokens.Estimate(full) * 2
	out, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index, "--budget", strconv.Itoa(limit))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ServiceID string         `json:"service_id"`
		Budget    *slices.Budget `json:"budget"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ServiceID != "service:acme" {
		t.Fatalf("budgeted doc dropped the focal service: %s", out)
	}
	if doc.Budget == nil {
		t.Fatalf("budgeted doc carries no budget envelope")
	}
	if doc.Budget.Limit != limit || doc.Budget.Estimated != tokens.Estimate(out) {
		t.Errorf("budget %+v, want limit %d estimated %d", *doc.Budget, limit, tokens.Estimate(out))
	}

	mdOut, err := runCLI(t, "doc", "service:acme", "--format", "md", "--index", index, "--budget", strconv.Itoa(limit))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mdOut), "## Budget") {
		t.Fatalf("budgeted md is missing the budget marker:\n%s", mdOut)
	}

	// A ceiling below the full output truncates calls first, reports what
	// it dropped, and stays byte-identical on repeat.
	tight := tokens.Estimate(full) - 100
	tightOut, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index, "--budget", strconv.Itoa(tight))
	if err != nil {
		t.Fatalf("tight budget failed: %v", err)
	}
	var tightDoc struct {
		ServiceID string         `json:"service_id"`
		Budget    *slices.Budget `json:"budget"`
	}
	if err := json.Unmarshal(tightOut, &tightDoc); err != nil {
		t.Fatal(err)
	}
	if tightDoc.ServiceID != "service:acme" {
		t.Fatalf("truncated doc dropped the focal service: %s", tightOut)
	}
	if tightDoc.Budget == nil || !tightDoc.Budget.Truncated || tightDoc.Budget.Dropped.Relationships == 0 {
		t.Fatalf("tight budget must truncate calls and say so: %+v", tightDoc.Budget)
	}
	if tightDoc.Budget.EscalateHint == "" {
		t.Fatalf("truncated doc carries no escalate hint: %+v", tightDoc.Budget)
	}
	tightAgain, err := runCLI(t, "doc", "service:acme", "--format", "json", "--index", index, "--budget", strconv.Itoa(tight))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tightOut, tightAgain) {
		t.Fatalf("truncated output is not byte-identical on repeat")
	}
}
