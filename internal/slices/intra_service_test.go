package slices_test

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/extractors"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

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

func scDocFiles() []extractors.TargetFile {
	return []extractors.TargetFile{
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "go.mod", Content: []byte(scDocGoMod)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "cmd/acme/main.go", Content: []byte(scDocMainGo)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "internal/slices/slices.go", Content: []byte(scDocSlicesGo)},
		{Repo: "acme", CommitSHA: "c0ffee01", Path: "cmd/acme/helper.py", Content: []byte(scDocHelperPy)},
	}
}

func scDocEngine(t *testing.T) *slices.Engine {
	t.Helper()
	mm, err := extractors.DefaultPipeline().ProcessFiles(context.Background(), "acme-system", scDocFiles())
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}
	return slices.NewEngine(mm)
}

var scLocatorRe = regexp.MustCompile(`^acme@c0ffee01:[^:]+#L[0-9]+-L[0-9]+$`)

// collectLocators gathers every locator a doc cites, for the format sweep.
func collectLocators(doc *slices.IntraServiceDoc) []domain.Locator {
	var out []domain.Locator
	out = append(out, doc.ServiceLocator)
	for _, e := range doc.Entrypoints {
		out = append(out, e.Locator)
	}
	for _, p := range doc.Packages {
		out = append(out, p.Locator)
	}
	for _, s := range doc.Types {
		out = append(out, s.Locator)
	}
	for _, f := range doc.Funcs {
		out = append(out, f.Locator)
	}
	for _, c := range append(append([]slices.IntraContractRef(nil), doc.ProvidedContracts...), doc.ConsumedContracts...) {
		out = append(out, c.Locators...)
	}
	for _, im := range append(append(append([]domain.GoImportRecord(nil), doc.IntraModuleImports...), doc.ExternalImports...), doc.StdlibImports...) {
		out = append(out, im.Locator)
	}
	for _, c := range doc.Calls {
		out = append(out, c.Locator)
	}
	for _, u := range doc.KnownUnknowns {
		out = append(out, u.Locator)
	}
	return out
}

func TestSCDocIntraService(t *testing.T) {
	eng := scDocEngine(t)

	t.Run("entrypoints", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:acme")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		if len(doc.Entrypoints) != 1 {
			t.Fatalf("want exactly 1 entrypoint, got %+v", doc.Entrypoints)
		}
		ep := doc.Entrypoints[0]
		if ep.Func != "main" || ep.Package != "main" || ep.Dir != "cmd/acme" {
			t.Errorf("wrong entrypoint: %+v", ep)
		}
		if ep.Origin != domain.OriginInferred {
			t.Errorf("entrypoint origin = %q, want inferred (rule-derived)", ep.Origin)
		}
		if got := ep.Locator.String(); got != "acme@c0ffee01:cmd/acme/main.go#L19-L19" {
			t.Errorf("entrypoint locator = %q", got)
		}
	})

	t.Run("symbols", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:internal")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		if len(doc.Packages) != 1 || doc.Packages[0].Name != "slices" || doc.Packages[0].Dir != "internal/slices" {
			t.Fatalf("packages = %+v", doc.Packages)
		}
		if doc.Packages[0].Origin != domain.OriginDeclared {
			t.Errorf("package origin = %q, want declared", doc.Packages[0].Origin)
		}
		types := map[string]domain.GoTypeKind{}
		for _, ty := range doc.Types {
			types[ty.Name] = ty.Kind
			if ty.Origin != domain.OriginExtracted {
				t.Errorf("type %s origin = %q, want extracted", ty.Name, ty.Origin)
			}
		}
		if types["Server"] != domain.GoTypeStruct || types["Store"] != domain.GoTypeInterface {
			t.Errorf("types = %+v, want Server struct + Store interface", types)
		}
		funcs := map[string]domain.GoFuncSymbol{}
		for _, f := range doc.Funcs {
			funcs[f.Name] = f
		}
		hello, ok := funcs["Hello"]
		if !ok || !hello.Exported {
			t.Errorf("Hello missing or not exported: %+v", funcs)
		}
		if u, ok := funcs["hello"]; !ok || u.Exported {
			t.Errorf("unexported hello missing: %+v", funcs)
		}
		start, ok := funcs["Start"]
		if !ok || start.Receiver != "Server" {
			t.Errorf("method Start missing receiver Server: %+v", funcs)
		}
		if _, ok := funcs["Count"]; !ok {
			t.Errorf("Count missing: %+v", funcs)
		}
		// Stable IDs: path + name, identical across runs.
		if hello.ID != "gosym:internal:func:internal/slices.Hello" {
			t.Errorf("Hello ID = %q", hello.ID)
		}
		if start.ID != "gosym:internal:func:internal/slices.Server.Start" {
			t.Errorf("Start ID = %q", start.ID)
		}
	})

	t.Run("contracts", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:acme")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		if len(doc.ProvidedContracts) != 1 || doc.ProvidedContracts[0].ContractName != "Orders" {
			t.Fatalf("provided = %+v, want Orders", doc.ProvidedContracts)
		}
		providesLines := map[int]bool{}
		for _, l := range doc.ProvidedContracts[0].Locators {
			providesLines[l.StartLine] = true
		}
		if !providesLines[20] {
			t.Errorf("provides Orders locators miss the Register line L20: %+v", doc.ProvidedContracts[0].Locators)
		}
		if len(doc.ConsumedContracts) != 1 || doc.ConsumedContracts[0].ContractName != "Billing" {
			t.Fatalf("consumed = %+v, want Billing", doc.ConsumedContracts)
		}
		if len(doc.ConsumedContracts[0].Locators) == 0 || doc.ConsumedContracts[0].Locators[0].StartLine != 25 {
			t.Errorf("consumes Billing locators miss the call line L25: %+v", doc.ConsumedContracts[0].Locators)
		}
	})

	t.Run("imports", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:acme")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		intra := map[string]bool{}
		for _, im := range doc.IntraModuleImports {
			intra[im.Path] = true
			if im.Origin != domain.OriginExtracted {
				t.Errorf("import %s origin = %q, want extracted", im.Path, im.Origin)
			}
		}
		if !intra["example.com/acme/internal/slices"] || !intra["example.com/acme/genproto/orders"] {
			t.Errorf("intra-module imports = %+v", doc.IntraModuleImports)
		}
		if len(doc.ExternalImports) != 1 || doc.ExternalImports[0].Path != "github.com/external/dep" {
			t.Errorf("external imports = %+v", doc.ExternalImports)
		}
		if len(doc.StdlibImports) != 1 || doc.StdlibImports[0].Path != "fmt" {
			t.Errorf("stdlib imports = %+v", doc.StdlibImports)
		}
	})

	t.Run("calls", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:acme")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		byCallee := map[string][]domain.GoCallRecord{}
		for _, c := range doc.Calls {
			byCallee[c.Callee] = append(byCallee[c.Callee], c)
			switch c.Resolution {
			case domain.GoCallResolved:
				if c.TargetID == "" || c.External != nil {
					t.Errorf("resolved call %s must cite a symbol only", c.ID)
				}
			case domain.GoCallExternal:
				if c.External == nil || c.External.ImportPath == "" || c.TargetID != "" {
					t.Errorf("external call %s must cite a stub only", c.ID)
				}
			case domain.GoCallUnresolved:
				if c.TargetID != "" || c.External != nil {
					t.Errorf("unresolved call %s must cite nothing", c.ID)
				}
			default:
				t.Errorf("call %s resolution %q outside the discrete rubric", c.ID, c.Resolution)
			}
			if c.Origin != domain.OriginExtracted {
				t.Errorf("call %s origin = %q, want extracted", c.ID, c.Origin)
			}
		}
		// No collapsing: Hello is called twice, two records, distinct IDs.
		if len(byCallee["Hello"]) != 2 || byCallee["Hello"][0].ID == byCallee["Hello"][1].ID {
			t.Fatalf("want 2 distinct Hello call records, got %+v", byCallee["Hello"])
		}
		for _, c := range byCallee["Hello"] {
			if c.Resolution != domain.GoCallResolved || c.TargetID != "gosym:internal:func:internal/slices.Hello" {
				t.Errorf("Hello call not resolved to the slices symbol: %+v", c)
			}
		}
		if len(byCallee["Println"]) != 2 {
			t.Errorf("want 2 Println call records, got %+v", byCallee["Println"])
		}
		for _, c := range byCallee["Println"] {
			if c.Resolution != domain.GoCallExternal || c.External.ImportPath != "fmt" {
				t.Errorf("Println call not an fmt stub: %+v", c)
			}
		}
		if proc := byCallee["Process"]; len(proc) != 1 || proc[0].Resolution != domain.GoCallExternal || proc[0].External.ImportPath != "github.com/external/dep" {
			t.Errorf("dep.Process call wrong: %+v", proc)
		}
		if reg := byCallee["RegisterOrdersServer"]; len(reg) != 1 || reg[0].Resolution != domain.GoCallUnresolved {
			t.Errorf("genproto call must stay explicitly unresolved (declaration outside corpus): %+v", reg)
		}
		if sub := byCallee["Submit"]; len(sub) != 1 || sub[0].Resolution != domain.GoCallUnresolved {
			t.Errorf("Submit call must stay explicitly unresolved: %+v", sub)
		}

		internal, err := eng.GetIntraServiceDoc("service:internal")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		for _, c := range internal.Calls {
			if c.Callee == "len" {
				t.Errorf("builtin len must be skipped, got %+v", c)
			}
		}
		var fileScope bool
		for _, c := range internal.Calls {
			if c.Callee == "Hello" && c.Caller == "" && c.Resolution == domain.GoCallResolved {
				fileScope = true
			}
			if c.Callee == "Hello" && c.Caller != "" && c.Caller != "gosym:internal:func:internal/slices.Server.Start" {
				t.Errorf("Hello caller wrong: %+v", c)
			}
		}
		if !fileScope {
			t.Errorf("package-level var call (caller \"\") missing: %+v", internal.Calls)
		}
	})

	t.Run("known-unknowns", func(t *testing.T) {
		doc, err := eng.GetIntraServiceDoc("service:acme")
		if err != nil {
			t.Fatalf("GetIntraServiceDoc failed: %v", err)
		}
		var found *domain.UnsupportedBoundary
		for i, u := range doc.KnownUnknowns {
			if u.Path == "cmd/acme/helper.py" {
				found = &doc.KnownUnknowns[i]
			}
		}
		if found == nil {
			t.Fatalf("helper.py missing from Known Unknowns: %+v", doc.KnownUnknowns)
		}
		joined := false
		for _, j := range found.UnknownJoins {
			if j == "contract:Billing" {
				joined = true
			}
		}
		if !joined {
			t.Errorf("helper.py must keep its unknown join to contract:Billing: %+v", found)
		}
		// Never guessed: no symbol, import, call, contract or entrypoint is
		// minted from helper.py; only its own Known Unknown row cites it.
		for _, loc := range collectLocators(doc) {
			if loc.Path != "cmd/acme/helper.py" {
				continue
			}
			if loc.String() != found.Locator.String() {
				t.Errorf("fact minted from unsupported file: %s", loc.String())
			}
		}
	})

	t.Run("locators", func(t *testing.T) {
		for _, id := range []string{"service:acme", "service:internal"} {
			doc, err := eng.GetIntraServiceDoc(id)
			if err != nil {
				t.Fatalf("GetIntraServiceDoc(%s) failed: %v", id, err)
			}
			for _, l := range collectLocators(doc) {
				if !scLocatorRe.MatchString(l.String()) {
					t.Errorf("locator %q is not [repo@commit:path#Lstart-Lend]", l.String())
				}
			}
		}
		if _, err := eng.GetIntraServiceDoc("service:nope"); err == nil {
			t.Errorf("unknown service must error")
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		mk := func() map[string][]byte {
			mm, err := extractors.DefaultPipeline().ProcessFiles(context.Background(), "acme-system", scDocFiles())
			if err != nil {
				t.Fatalf("ProcessFiles failed: %v", err)
			}
			out := map[string][]byte{}
			for _, id := range []string{"service:acme", "service:internal"} {
				node, ok := func() (domain.Node, bool) {
					for _, n := range mm.Nodes {
						if n.ID == id {
							return n, true
						}
					}
					return domain.Node{}, false
				}()
				if !ok {
					t.Fatalf("missing node %s", id)
				}
				raw, ok := node.Metadata[domain.IntraServiceMetadataKey].(domain.IntraServiceDetail)
				if !ok {
					t.Fatalf("node %s has no intra-service detail struct", id)
				}
				if err := raw.Validate(); err != nil {
					t.Fatalf("detail fails the fragment contract: %v", err)
				}
				js, err := raw.CanonicalJSON()
				if err != nil {
					t.Fatalf("CanonicalJSON failed: %v", err)
				}
				out[id] = js
			}
			return out
		}
		a, b := mk(), mk()
		for id := range a {
			if !bytes.Equal(a[id], b[id]) {
				t.Errorf("non-deterministic detail for %s:\n%s\nvs\n%s", id, a[id], b[id])
			}
		}
	})
}
