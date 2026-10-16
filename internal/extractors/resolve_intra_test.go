package extractors

import (
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestGoModulesByRepo(t *testing.T) {
	files := []TargetFile{
		{Repo: "a", Path: "go.mod", Content: []byte("module example.com/acme\n\ngo 1.24\n")},
		{Repo: "a", Path: "sub/go.mod", Content: []byte("module example.com/other\n")},
		{Repo: "a", Path: "dup/go.mod", Content: []byte("module example.com/acme\n")},
		{Repo: "b", Path: "main.go", Content: []byte("package main\n")},
	}
	got := goModulesByRepo(files)
	if len(got["a"]) != 2 || got["a"][0] != "example.com/acme" || got["a"][1] != "example.com/other" {
		t.Fatalf("modules = %v, want sorted deduped [acme other]", got)
	}
	if _, ok := got["b"]; ok {
		t.Fatalf("repo without go.mod must be absent: %v", got)
	}
}

func TestImportScope(t *testing.T) {
	modules := map[string][]string{"r": {"example.com/acme"}}
	cases := []struct {
		path string
		want domain.GoImportScope
	}{
		{"example.com/acme", domain.GoImportIntraModule},
		{"example.com/acme/internal/slices", domain.GoImportIntraModule},
		{"example.com/acme2/internal", domain.GoImportExternal}, // prefix must end at a boundary
		{"github.com/external/dep", domain.GoImportExternal},
		{"fmt", domain.GoImportStdlib},
		{"strings", domain.GoImportStdlib},
	}
	for _, c := range cases {
		if got := importScope(modules, "r", c.path); got != c.want {
			t.Errorf("importScope(%q) = %q, want %q", c.path, got, c.want)
		}
	}
	if got := importScope(nil, "r", "github.com/x/y"); got != domain.GoImportExternal {
		t.Errorf("without go.mod every dotted import is external, got %q", got)
	}
}

func validTestDetail() domain.IntraServiceDetail {
	loc := domain.Locator{Repo: "r", CommitSHA: "c", Path: "m.go", StartLine: 1, EndLine: 1}
	return domain.IntraServiceDetail{
		Schema:    domain.IntraServiceSchema,
		ServiceID: "service:m",
		Packages:  []domain.GoPackageSymbol{{ID: "p", Name: "m", Dir: ".", Locator: loc, Origin: domain.OriginDeclared}},
		Types:     []domain.GoTypeSymbol{{ID: "t", Name: "T", Kind: domain.GoTypeStruct, Locator: loc, Origin: domain.OriginExtracted}},
		Funcs:     []domain.GoFuncSymbol{{ID: "f", Name: "F", Exported: true, Locator: loc, Origin: domain.OriginExtracted}},
		Imports:   []domain.GoImportRecord{{ID: "i", Path: "fmt", Alias: "fmt", Scope: domain.GoImportStdlib, Locator: loc, Origin: domain.OriginExtracted}},
		Calls: []domain.GoCallRecord{
			{ID: "c1", Callee: "F", TargetID: "f", Resolution: domain.GoCallResolved, Locator: loc, Origin: domain.OriginExtracted},
			{ID: "c2", Callee: "Print", External: &domain.GoExternalStub{ImportPath: "fmt", Name: "Print"}, Resolution: domain.GoCallExternal, Locator: loc, Origin: domain.OriginExtracted},
			{ID: "c3", Callee: "Maybe", Resolution: domain.GoCallUnresolved, Locator: loc, Origin: domain.OriginExtracted},
		},
		Entrypoints: []domain.GoEntrypoint{{ID: "e", Func: "main", Package: "main", Dir: ".", Locator: loc, Origin: domain.OriginInferred}},
	}
}

func TestIntraServiceValidate(t *testing.T) {
	if err := func() error {
		d := validTestDetail()
		return d.Validate()
	}(); err != nil {
		t.Fatalf("valid detail rejected: %v", err)
	}

	mutate := func(f func(*domain.IntraServiceDetail)) domain.IntraServiceDetail {
		d := validTestDetail()
		f(&d)
		return d
	}
	badOrigin := domain.DerivationOrigin("likely")
	cases := []struct {
		name string
		d    domain.IntraServiceDetail
	}{
		{"schema", mutate(func(d *domain.IntraServiceDetail) { d.Schema = "x" })},
		{"service", mutate(func(d *domain.IntraServiceDetail) { d.ServiceID = "" })},
		{"dupe", mutate(func(d *domain.IntraServiceDetail) { d.Entrypoints[0].ID = "p" })},
		{"package-origin", mutate(func(d *domain.IntraServiceDetail) { d.Packages[0].Origin = domain.OriginExtracted })},
		{"type-kind", mutate(func(d *domain.IntraServiceDetail) { d.Types[0].Kind = "class" })},
		{"func-origin", mutate(func(d *domain.IntraServiceDetail) { d.Funcs[0].Origin = badOrigin })},
		{"import-scope", mutate(func(d *domain.IntraServiceDetail) { d.Imports[0].Scope = "first-party" })},
		{"no-locator", mutate(func(d *domain.IntraServiceDetail) { d.Funcs[0].Locator.StartLine = 0 })},
		{"resolved-with-stub", mutate(func(d *domain.IntraServiceDetail) {
			d.Calls[0].External = &domain.GoExternalStub{ImportPath: "x", Name: "Y"}
		})},
		{"external-with-target", mutate(func(d *domain.IntraServiceDetail) { d.Calls[1].TargetID = "f" })},
		{"unresolved-with-target", mutate(func(d *domain.IntraServiceDetail) { d.Calls[2].TargetID = "f" })},
		{"call-resolution", mutate(func(d *domain.IntraServiceDetail) { d.Calls[2].Resolution = "probably" })},
		{"entrypoint-origin", mutate(func(d *domain.IntraServiceDetail) { d.Entrypoints[0].Origin = domain.OriginDeclared })},
	}
	for _, c := range cases {
		if err := c.d.Validate(); err == nil {
			t.Errorf("%s: invalid detail accepted", c.name)
		}
	}
}
