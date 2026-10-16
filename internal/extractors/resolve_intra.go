package extractors

import (
	"fmt"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// goModulePath is the module path of a go.mod file ("" when absent): the
// first `module <path>` line, in file order.
func goModulePath(content []byte) string {
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 && fields[0] == "module" {
			return strings.Trim(fields[1], `"`)
		}
	}
	return ""
}

// goModulesByRepo indexes go.mod module paths by repo, sorted and deduped,
// so import scope is a pure function of (repo, path).
func goModulesByRepo(files []TargetFile) map[string][]string {
	out := map[string][]string{}
	seen := map[string]bool{}
	for _, f := range files {
		if !strings.HasSuffix(f.Path, "go.mod") {
			continue
		}
		m := goModulePath(f.Content)
		if m == "" || seen[f.Repo+"\x00"+m] {
			continue
		}
		seen[f.Repo+"\x00"+m] = true
		out[f.Repo] = append(out[f.Repo], m)
	}
	for repo := range out {
		sorted := append([]string(nil), out[repo]...)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
		out[repo] = sorted
	}
	return out
}

// importScope classifies an import path for repo: intra-module under any of
// the repo's go.mod modules, stdlib when the path has no domain dot,
// external otherwise.
func importScope(modules map[string][]string, repo, path string) domain.GoImportScope {
	for _, m := range modules[repo] {
		if path == m || strings.HasPrefix(path, m+"/") {
			return domain.GoImportIntraModule
		}
	}
	if strings.Contains(path, ".") {
		return domain.GoImportExternal
	}
	return domain.GoImportStdlib
}

// importDir maps an intra-module import path back to its repo-relative
// directory via the module prefix it matched ("" when no module matches).
func importDir(modules map[string][]string, repo, path string) string {
	for _, m := range modules[repo] {
		if path == m {
			return "."
		}
		if rest, ok := strings.CutPrefix(path, m+"/"); ok {
			return rest
		}
	}
	return ""
}

// canonicalService follows the Dockerfile→compose attribution renames to
// the service ID whose node the detail must attach to.
func (r *resolution) canonicalService(s string) string {
	seen := map[string]bool{s: true}
	for {
		to, ok := r.renamed[s]
		if !ok || seen[to] {
			return s
		}
		seen[to] = true
		s = to
	}
}

// intraName is the bare service name for stable IDs, matching the
// extractors' relationship IDs ("rel:checkout-consumes-email").
func intraName(serviceID string) string {
	return strings.TrimPrefix(serviceID, "service:")
}

// intraBuilder accumulates one service's F2 records from sorted hints.
type intraBuilder struct {
	service string
	modules []string
	pkgs    map[string]*domain.GoPackageSymbol // dir → package (first wins)
	types   []*domain.GoTypeSymbol
	funcs   []*domain.GoFuncSymbol
	funcDir map[string]string // func symbol ID → dir
	imports []*domain.GoImportRecord
	calls   []Hint // call hints, in sorted order
}

// localName is the caller/callee lookup name of a func symbol.
func localName(f *domain.GoFuncSymbol) string {
	if f.Receiver != "" {
		return f.Receiver + "." + f.Name
	}
	return f.Name
}

type symKey struct {
	repo, dir, name string
}

// resolveIntraService assembles the per-service intra-service detail (F2)
// from the Go symbol hints and stores it as service node metadata. It mints
// no nodes and no edges: only a new metadata key on services that Go files
// produced. Anything it cannot resolve stays an explicit unknown.
func resolveIntraService(r *resolution) {
	builders := map[string]*intraBuilder{} // canonical service → builder
	funcIdx := map[symKey][]*domain.GoFuncSymbol{}
	typeSet := map[symKey]bool{}

	svcOf := func(h Hint) *intraBuilder {
		svc := r.canonicalService(h.Service)
		b, ok := builders[svc]
		if !ok {
			b = &intraBuilder{
				service: svc,
				modules: r.modules[h.Locator.Repo],
				pkgs:    map[string]*domain.GoPackageSymbol{},
				funcDir: map[string]string{},
			}
			builders[svc] = b
		}
		return b
	}

	//Pass 1: symbols and imports. Hints arrive sorted by
	//(repo, path, line), so first-wins dedupe is deterministic.
	seenSym := map[string]bool{}
	for _, h := range r.hints {
		switch h.Kind {
		case HintGoPackage, HintGoType, HintGoFunc, HintGoImport, HintGoCall:
		default:
			continue
		}
		b := svcOf(h)
		dir := h.Context
		switch h.Kind {
		case HintGoPackage:
			if _, dup := b.pkgs[dir]; !dup {
				b.pkgs[dir] = &domain.GoPackageSymbol{
					ID:      fmt.Sprintf("gosym:%s:package:%s", intraName(b.service), dir),
					Name:    h.Key,
					Dir:     dir,
					Locator: h.Locator,
					Origin:  domain.OriginDeclared,
				}
			}
		case HintGoType:
			id := fmt.Sprintf("gosym:%s:type:%s.%s", intraName(b.service), dir, h.Key)
			if seenSym[id] {
				continue
			}
			seenSym[id] = true
			typeSet[symKey{h.Locator.Repo, dir, h.Key}] = true
			b.types = append(b.types, &domain.GoTypeSymbol{
				ID:      id,
				Name:    h.Key,
				Kind:    domain.GoTypeKind(h.Value),
				Locator: h.Locator,
				Origin:  domain.OriginExtracted,
			})
		case HintGoFunc:
			id := fmt.Sprintf("gosym:%s:func:%s.%s", intraName(b.service), dir, h.Key)
			if h.Value != "" {
				id = fmt.Sprintf("gosym:%s:func:%s.%s.%s", intraName(b.service), dir, h.Value, h.Key)
			}
			if seenSym[id] {
				continue
			}
			seenSym[id] = true
			f := &domain.GoFuncSymbol{
				ID:       id,
				Name:     h.Key,
				Receiver: h.Value,
				Exported: h.Mode == "exported",
				Locator:  h.Locator,
				Origin:   domain.OriginExtracted,
			}
			b.funcs = append(b.funcs, f)
			b.funcDir[f.ID] = dir
			funcIdx[symKey{h.Locator.Repo, dir, h.Key}] = append(funcIdx[symKey{h.Locator.Repo, dir, h.Key}], f)
		case HintGoImport:
			id := fmt.Sprintf("goimport:%s:%s:%s#L%d", intraName(b.service), h.Locator.Repo, h.Locator.Path, h.Locator.StartLine)
			if seenSym[id] {
				continue
			}
			seenSym[id] = true
			b.imports = append(b.imports, &domain.GoImportRecord{
				ID:      id,
				Path:    h.Key,
				Alias:   h.Value,
				Scope:   importScope(r.modules, h.Locator.Repo, h.Key),
				Locator: h.Locator,
				Origin:  domain.OriginExtracted,
			})
		case HintGoCall:
			b.calls = append(b.calls, h)
		}
	}

	//Pass 2: resolve calls, derive entrypoints, attach.
	for _, b := range builders {
		detail := &domain.IntraServiceDetail{
			Schema:    domain.IntraServiceSchema,
			ServiceID: b.service,
		}
		if len(b.modules) > 0 {
			detail.GoModule = b.modules[0]
		}
		for _, p := range b.pkgs {
			detail.Packages = append(detail.Packages, *p)
		}
		for _, t := range b.types {
			detail.Types = append(detail.Types, *t)
		}
		for _, f := range b.funcs {
			detail.Funcs = append(detail.Funcs, *f)
		}
		for _, im := range b.imports {
			detail.Imports = append(detail.Imports, *im)
		}
		detail.Calls = b.resolveCalls(r, funcIdx, typeSet)
		detail.Entrypoints = b.entrypoints()
		detail.Sort()
		node, ok := r.nodes[b.service]
		if !ok {
			continue
		}
		if node.Metadata == nil {
			node.Metadata = map[string]any{}
		}
		node.Metadata[domain.IntraServiceMetadataKey] = *detail
	}
}

// resolveCalls classifies every call site (one record per site, never
// collapsed) with a deterministic rule:
//
//   - a callee naming a type in the target directory is a conversion, not a
//     call, and is dropped;
//   - an intra-module target resolves iff exactly one func of that name is
//     in the corpus directory (never guessed when ambiguous or absent);
//   - an out-of-module target becomes an explicit external stub;
//   - anything else is an explicit unresolved unknown.
func (b *intraBuilder) resolveCalls(r *resolution, funcIdx map[symKey][]*domain.GoFuncSymbol, typeSet map[symKey]bool) []domain.GoCallRecord {
	out := []domain.GoCallRecord{}
	seq := map[string]int{} // (repo, path, line) → per-site sequence
	for _, h := range b.calls {
		repo, dir := h.Locator.Repo, h.Context
		name := h.Value
		seqKey := repo + "\x00" + h.Locator.Path + "\x00" + fmt.Sprint(h.Locator.StartLine)
		seq[seqKey]++
		rec := domain.GoCallRecord{
			ID:     fmt.Sprintf("gocall:%s:%s:%s#L%d-%d", intraName(b.service), repo, h.Locator.Path, h.Locator.StartLine, seq[seqKey]),
			Caller: b.callerID(dir, h.Key),
			Callee: name,
			//Resolution and target are set below; the zero value is
			//never emitted (every branch assigns explicitly).
			Locator: h.Locator,
			Origin:  domain.OriginExtracted,
		}
		switch h.Mode {
		case "pkg":
			resolution, target, ext, drop := b.resolvePkgCall(r, funcIdx, typeSet, repo, h.Ref, name)
			if drop {
				continue // conversion pkg.Type(x), not a call
			}
			rec.Resolution, rec.TargetID, rec.External = resolution, target, ext
		case "bare-dotimport":
			//A bare call in a dot-importing file may come from
			//another package: resolving it would be a guess.
			rec.Resolution = domain.GoCallUnresolved
		default: // "local", "bare": same-directory candidates only.
			if typeSet[symKey{repo, dir, name}] {
				continue // conversion, not a call
			}
			cands := funcIdx[symKey{repo, dir, name}]
			if len(cands) == 1 {
				rec.Resolution, rec.TargetID = domain.GoCallResolved, cands[0].ID
			} else {
				rec.Resolution = domain.GoCallUnresolved
			}
		}
		out = append(out, rec)
	}
	return out
}

// resolvePkgCall classifies a qualified pkg.Name call: intra-module targets
// resolve by (directory, name) with the same exactly-one rule; out-of-module
// targets become external stubs. drop is true for pkg.Type(x) conversions,
// which are not calls.
func (b *intraBuilder) resolvePkgCall(r *resolution, funcIdx map[symKey][]*domain.GoFuncSymbol, typeSet map[symKey]bool, repo, importPath, name string) (domain.GoCallResolution, string, *domain.GoExternalStub, bool) {
	if importScope(r.modules, repo, importPath) != domain.GoImportIntraModule {
		return domain.GoCallExternal, "", &domain.GoExternalStub{ImportPath: importPath, Name: name}, false
	}
	dir := importDir(r.modules, repo, importPath)
	k := symKey{repo, dir, name}
	if typeSet[k] {
		return "", "", nil, true
	}
	cands := funcIdx[k]
	if len(cands) == 1 {
		return domain.GoCallResolved, cands[0].ID, nil, false
	}
	return domain.GoCallUnresolved, "", nil, false
}

// callerID maps a call hint's enclosing func back to its symbol ID ("" at
// file scope, or when the enclosing func left no symbol, which cannot happen
// for valid input but stays honest rather than guessing).
func (b *intraBuilder) callerID(dir, caller string) string {
	if caller == "" {
		return ""
	}
	for _, f := range b.funcs {
		if b.funcDir[f.ID] == dir && localName(f) == caller {
			return f.ID
		}
	}
	return ""
}

// entrypoints derives the service entrypoints by rule: a func main in a
// package main. The func declaration is extracted; its entrypoint role is
// inferred, hence the origin.
func (b *intraBuilder) entrypoints() []domain.GoEntrypoint {
	var out []domain.GoEntrypoint
	for _, f := range b.funcs {
		if f.Name != "main" || f.Receiver != "" {
			continue
		}
		dir := b.funcDir[f.ID]
		p, ok := b.pkgs[dir]
		if !ok || p.Name != "main" {
			continue
		}
		out = append(out, domain.GoEntrypoint{
			ID:      fmt.Sprintf("goentry:%s:%s", intraName(b.service), dir),
			Func:    "main",
			Package: "main",
			Dir:     dir,
			Locator: f.Locator,
			Origin:  domain.OriginInferred,
		})
	}
	return out
}
