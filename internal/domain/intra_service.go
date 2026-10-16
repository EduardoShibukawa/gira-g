package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// IntraServiceMetadataKey is the Service node metadata key carrying the
// deterministic intra-service detail (F2, phase 1 = Go). It is set by the
// extractors' resolve step, never by per-file extraction, so multi-file
// services assemble one record. Additive: readers that do not know the key
// ignore it; no existing key changes meaning.
const IntraServiceMetadataKey = "go_intra_service"

// IntraServiceSchema versions the IntraServiceDetail shape. Bump on any
// field change so future caches and goldens can tell records apart.
const IntraServiceSchema = "go-intra/1"

// Discrete origins for intra-service facts (no continuous confidence
// score, unlike symbol-graph prior art):
//   - declared: identity a file declares (package clause, module path).
//   - extracted: symbol, import or call site parsed deterministically
//     from a Go AST.
//   - inferred: fact the resolve step derives by rule across files
//     (entrypoints; never a guess about an unknown target).
// The Origin* constants above carry these values; Validate enforces them.

// GoSymbolKind names the granularity of an intra-service symbol.
type GoSymbolKind string

const (
	GoSymbolPackage GoSymbolKind = "package"
	GoSymbolType    GoSymbolKind = "type"
	GoSymbolFunc    GoSymbolKind = "func"
)

// GoTypeKind names the shape of a Go type declaration.
type GoTypeKind string

const (
	GoTypeStruct    GoTypeKind = "struct"
	GoTypeInterface GoTypeKind = "interface"
	GoTypeAlias     GoTypeKind = "alias"
	GoTypeOther     GoTypeKind = "other"
)

// GoImportScope classifies one import statement deterministically:
// intra-module when the path lives under a go.mod module of the same repo,
// stdlib when the path has no domain dot, external otherwise.
type GoImportScope string

const (
	GoImportIntraModule GoImportScope = "intra-module"
	GoImportExternal    GoImportScope = "external"
	GoImportStdlib      GoImportScope = "stdlib"
)

// GoCallResolution says how a call site's target was classified. Exactly one
// of TargetID / External is set for resolved / external; both are empty for
// unresolved, which is an explicit unknown, never a dangling guess.
type GoCallResolution string

const (
	GoCallResolved   GoCallResolution = "resolved"
	GoCallExternal   GoCallResolution = "external"
	GoCallUnresolved GoCallResolution = "unresolved"
)

// GoPackageSymbol is one Go package as declared by a package clause.
type GoPackageSymbol struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Dir     string           `json:"dir"`
	Locator Locator          `json:"locator"`
	Origin  DerivationOrigin `json:"origin"`
}

// GoTypeSymbol is one top-level Go type declaration.
type GoTypeSymbol struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Kind    GoTypeKind       `json:"kind"`
	Locator Locator          `json:"locator"`
	Origin  DerivationOrigin `json:"origin"`
}

// GoFuncSymbol is one top-level Go func or method declaration. Receiver is
// the dereferenced receiver type name, or "" for a plain function.
type GoFuncSymbol struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Receiver string           `json:"receiver,omitempty"`
	Exported bool             `json:"exported"`
	Locator  Locator          `json:"locator"`
	Origin   DerivationOrigin `json:"origin"`
}

// GoImportRecord is one Go import statement: path, local name and scope.
type GoImportRecord struct {
	ID      string           `json:"id"`
	Path    string           `json:"path"`
	Alias   string           `json:"alias"`
	Scope   GoImportScope    `json:"scope"`
	Locator Locator          `json:"locator"`
	Origin  DerivationOrigin `json:"origin"`
}

// GoExternalStub names a call target outside the service's modules so the
// call never dangles: every call cites either an intra-module symbol ID or
// an explicit stub (import path + symbol name).
type GoExternalStub struct {
	ImportPath string `json:"import_path"`
	Name       string `json:"name"`
}

// GoCallRecord is one Go call site. Callee is the callee name as written
// (without qualifier); Caller is the enclosing func symbol ID ("" at file
// scope). One record per site: repeated calls never collapse.
type GoCallRecord struct {
	ID         string           `json:"id"`
	Caller     string           `json:"caller,omitempty"`
	Callee     string           `json:"callee"`
	TargetID   string           `json:"target_id,omitempty"`
	External   *GoExternalStub  `json:"external,omitempty"`
	Resolution GoCallResolution `json:"resolution"`
	Locator    Locator          `json:"locator"`
	Origin     DerivationOrigin `json:"origin"`
}

// GoEntrypoint is a func main in a package main: the rule-derived,
// therefore inferred, entrypoint of a Go service.
type GoEntrypoint struct {
	ID      string           `json:"id"`
	Func    string           `json:"func"`
	Package string           `json:"package"`
	Dir     string           `json:"dir"`
	Locator Locator          `json:"locator"`
	Origin  DerivationOrigin `json:"origin"`
}

// IntraServiceDetail is the assembled per-service F2 payload: symbols
// (packages, types, funcs) plus imports and calls with file+line locators.
// GoModule is the repo's go.mod module path ("" when the batch has no
// go.mod); without it every dotted import reads as external.
type IntraServiceDetail struct {
	Schema      string            `json:"schema"`
	ServiceID   string            `json:"service_id"`
	GoModule    string            `json:"go_module,omitempty"`
	Packages    []GoPackageSymbol `json:"packages"`
	Types       []GoTypeSymbol    `json:"types"`
	Funcs       []GoFuncSymbol    `json:"funcs"`
	Imports     []GoImportRecord  `json:"imports"`
	Calls       []GoCallRecord    `json:"calls"`
	Entrypoints []GoEntrypoint    `json:"entrypoints"`
}

func validIntraLocator(l Locator) bool {
	return l.Path != "" && l.StartLine > 0 && l.EndLine >= l.StartLine
}

// Validate enforces the extraction-fragment contract: every record carries
// a stable ID, a file+line locator and a discrete origin; every call is
// explicitly resolved, external or unresolved. It reports the first
// violation.
func (d *IntraServiceDetail) Validate() error {
	if d.Schema != IntraServiceSchema {
		return fmt.Errorf("intra-service detail: schema %q, want %q", d.Schema, IntraServiceSchema)
	}
	if d.ServiceID == "" {
		return fmt.Errorf("intra-service detail: empty service_id")
	}
	seen := map[string]bool{}
	dupe := func(id, what string) error {
		if id == "" {
			return fmt.Errorf("intra-service detail: empty %s id", what)
		}
		if seen[id] {
			return fmt.Errorf("intra-service detail: duplicate id %q", id)
		}
		seen[id] = true
		return nil
	}
	for _, p := range d.Packages {
		if err := dupe(p.ID, "package"); err != nil {
			return err
		}
		if p.Name == "" || p.Dir == "" {
			return fmt.Errorf("intra-service detail: package %q missing name/dir", p.ID)
		}
		if !validIntraLocator(p.Locator) {
			return fmt.Errorf("intra-service detail: package %q without file+line locator", p.ID)
		}
		if p.Origin != OriginDeclared {
			return fmt.Errorf("intra-service detail: package %q origin %q, want declared", p.ID, p.Origin)
		}
	}
	for _, t := range d.Types {
		if err := dupe(t.ID, "type"); err != nil {
			return err
		}
		if t.Name == "" {
			return fmt.Errorf("intra-service detail: type %q missing name", t.ID)
		}
		switch t.Kind {
		case GoTypeStruct, GoTypeInterface, GoTypeAlias, GoTypeOther:
		default:
			return fmt.Errorf("intra-service detail: type %q kind %q outside rubric", t.ID, t.Kind)
		}
		if !validIntraLocator(t.Locator) {
			return fmt.Errorf("intra-service detail: type %q without file+line locator", t.ID)
		}
		if t.Origin != OriginExtracted {
			return fmt.Errorf("intra-service detail: type %q origin %q, want extracted", t.ID, t.Origin)
		}
	}
	for _, f := range d.Funcs {
		if err := dupe(f.ID, "func"); err != nil {
			return err
		}
		if f.Name == "" {
			return fmt.Errorf("intra-service detail: func %q missing name", f.ID)
		}
		if !validIntraLocator(f.Locator) {
			return fmt.Errorf("intra-service detail: func %q without file+line locator", f.ID)
		}
		if f.Origin != OriginExtracted {
			return fmt.Errorf("intra-service detail: func %q origin %q, want extracted", f.ID, f.Origin)
		}
	}
	for _, im := range d.Imports {
		if err := dupe(im.ID, "import"); err != nil {
			return err
		}
		if im.Path == "" {
			return fmt.Errorf("intra-service detail: import %q missing path", im.ID)
		}
		switch im.Scope {
		case GoImportIntraModule, GoImportExternal, GoImportStdlib:
		default:
			return fmt.Errorf("intra-service detail: import %q scope %q outside rubric", im.ID, im.Scope)
		}
		if !validIntraLocator(im.Locator) {
			return fmt.Errorf("intra-service detail: import %q without file+line locator", im.ID)
		}
		if im.Origin != OriginExtracted {
			return fmt.Errorf("intra-service detail: import %q origin %q, want extracted", im.ID, im.Origin)
		}
	}
	for _, c := range d.Calls {
		if err := dupe(c.ID, "call"); err != nil {
			return err
		}
		if c.Callee == "" {
			return fmt.Errorf("intra-service detail: call %q missing callee", c.ID)
		}
		switch c.Resolution {
		case GoCallResolved:
			if c.TargetID == "" || c.External != nil {
				return fmt.Errorf("intra-service detail: resolved call %q must cite a symbol, not a stub", c.ID)
			}
		case GoCallExternal:
			if c.External == nil || c.External.ImportPath == "" || c.External.Name == "" || c.TargetID != "" {
				return fmt.Errorf("intra-service detail: external call %q must cite a stub, not a symbol", c.ID)
			}
		case GoCallUnresolved:
			if c.TargetID != "" || c.External != nil {
				return fmt.Errorf("intra-service detail: unresolved call %q must cite neither symbol nor stub", c.ID)
			}
		default:
			return fmt.Errorf("intra-service detail: call %q resolution %q outside rubric", c.ID, c.Resolution)
		}
		if !validIntraLocator(c.Locator) {
			return fmt.Errorf("intra-service detail: call %q without file+line locator", c.ID)
		}
		if c.Origin != OriginExtracted {
			return fmt.Errorf("intra-service detail: call %q origin %q, want extracted", c.ID, c.Origin)
		}
	}
	for _, e := range d.Entrypoints {
		if err := dupe(e.ID, "entrypoint"); err != nil {
			return err
		}
		if e.Func == "" || e.Dir == "" {
			return fmt.Errorf("intra-service detail: entrypoint %q missing func/dir", e.ID)
		}
		if !validIntraLocator(e.Locator) {
			return fmt.Errorf("intra-service detail: entrypoint %q without file+line locator", e.ID)
		}
		if e.Origin != OriginInferred {
			return fmt.Errorf("intra-service detail: entrypoint %q origin %q, want inferred", e.ID, e.Origin)
		}
	}
	return nil
}

// Sort puts every list in canonical order (by ID, imports by path then
// line) so the detail serializes byte-deterministically.
func (d *IntraServiceDetail) Sort() {
	sort.Slice(d.Packages, func(i, j int) bool { return d.Packages[i].ID < d.Packages[j].ID })
	sort.Slice(d.Types, func(i, j int) bool { return d.Types[i].ID < d.Types[j].ID })
	sort.Slice(d.Funcs, func(i, j int) bool { return d.Funcs[i].ID < d.Funcs[j].ID })
	sort.Slice(d.Imports, func(i, j int) bool {
		if d.Imports[i].Path != d.Imports[j].Path {
			return d.Imports[i].Path < d.Imports[j].Path
		}
		if d.Imports[i].Locator.StartLine != d.Imports[j].Locator.StartLine {
			return d.Imports[i].Locator.StartLine < d.Imports[j].Locator.StartLine
		}
		return d.Imports[i].ID < d.Imports[j].ID
	})
	sort.Slice(d.Calls, func(i, j int) bool { return d.Calls[i].ID < d.Calls[j].ID })
	sort.Slice(d.Entrypoints, func(i, j int) bool { return d.Entrypoints[i].ID < d.Entrypoints[j].ID })
}

// CanonicalJSON serializes the detail deterministically (sorted records,
// sorted keys) for goldens and determinism checks.
func (d *IntraServiceDetail) CanonicalJSON() ([]byte, error) {
	c := *d
	c.Sort()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
