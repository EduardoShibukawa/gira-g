package extractors

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// kafkaClients are the Go Kafka client import paths whose producer messages and
// constructors the hints recognise.
var kafkaClients = map[string]bool{
	"github.com/IBM/sarama":     true,
	"github.com/Shopify/sarama": true,
}

// kafkaProducerCtors build a producer or client from a broker list (argument 0).
var kafkaProducerCtors = map[string]bool{"NewAsyncProducer": true, "NewSyncProducer": true, "NewClient": true}

var (
	envVarName    = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	importVersion = regexp.MustCompile(`^v[0-9]+$`)
)

// importName is the name a file refers to an import by: its explicit name, or
// the last path element without a version suffix or a go- / -go affix.
func importName(spec *ast.ImportSpec) (string, string) {
	p, _ := strconv.Unquote(spec.Path.Value)
	if spec.Name != nil {
		return spec.Name.Name, p
	}
	parts := strings.Split(p, "/")
	last := parts[len(parts)-1]
	if len(parts) > 1 && importVersion.MatchString(last) {
		last = parts[len(parts)-2]
	}
	last = strings.TrimSuffix(strings.TrimPrefix(last, "go-"), "-go")
	return last, p
}

// Hints returns the Go facts the resolution step joins across files (ADR 0013):
// environment bindings, package-level env variables, Kafka producer topics and
// brokers, and call arguments that read bound fields.
func (e *GoASTExtractor) Hints(ctx context.Context, file TargetFile) ([]Hint, error) {
	if strings.HasSuffix(file.Path, "_test.go") {
		return nil, nil
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file.Path, file.Content, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go source in %s: %w", file.Path, err)
	}
	g := &goHints{
		fset:    fset,
		file:    file,
		service: "service:" + deriveServiceName(file.Path, f.Name.Name),
		dir:     path.Dir(file.Path),
		imports: map[string]string{},
		funcs:   map[string]*ast.FuncDecl{},
	}
	for _, imp := range f.Imports {
		name, p := importName(imp)
		g.imports[name] = p
	}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
			g.funcs[fd.Name.Name] = fd
		}
	}
	g.packageVars(f)
	g.clientFields(f)
	g.fileSymbols(f)
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		g.fn = fd
		ast.Inspect(fd.Body, g.visit)
	}
	return g.out, nil
}

type goHints struct {
	fset    *token.FileSet
	file    TargetFile
	service string
	dir     string
	imports map[string]string // name → import path
	funcs   map[string]*ast.FuncDecl
	fn      *ast.FuncDecl // enclosing function while visiting
	out     []Hint
	// hasDotImport reports whether the file dot-imports a package: bare
	// calls then may come from another package, so the resolve step never
	// resolves them across files.
	hasDotImport bool
}

func (g *goHints) at(n ast.Node) domain.Locator {
	start, end := g.fset.Position(n.Pos()).Line, g.fset.Position(n.End()).Line
	return domain.Locator{Repo: g.file.Repo, CommitSHA: g.file.CommitSHA, Path: g.file.Path, StartLine: start, EndLine: end}
}

func (g *goHints) add(h Hint, n ast.Node) {
	h.Locator = g.at(n)
	if h.Service == "" {
		h.Service = g.service
	}
	if h.Context == "" && h.Kind != HintGoCallArg && h.Kind != HintKafkaTopic {
		h.Context = g.dir
	}
	g.out = append(g.out, h)
}

// pkgCall reports whether call is pkg.Name(...) for the import path pkgPath.
func (g *goHints) pkgCall(expr ast.Expr) (pkgPath, name string, call *ast.CallExpr) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", "", nil
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", nil
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", "", nil
	}
	return g.imports[x.Name], sel.Sel.Name, call
}

// getenv returns X for os.Getenv("X") (or os.LookupEnv).
func (g *goHints) getenv(expr ast.Expr) (string, bool) {
	p, name, call := g.pkgCall(expr)
	if p != "os" || (name != "Getenv" && name != "LookupEnv") || len(call.Args) != 1 {
		return "", false
	}
	return stringLit(call.Args[0])
}

func stringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

// bindingName is the name a field or variable is bound under: x.f → f, v → v.
func bindingName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// packageVars records package-level variables and constants whose value is an
// environment variable, a string literal, or a call to a function of this file
// that returns os.Getenv("X") with a literal fallback.
func (g *goHints) packageVars(f *ast.File) {
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || (gd.Tok != token.VAR && gd.Tok != token.CONST) {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Values) != len(vs.Names) {
				continue
			}
			for i, name := range vs.Names {
				val := vs.Values[i]
				if x, ok := g.getenv(val); ok {
					g.add(Hint{Kind: HintGoEnvVar, Key: name.Name, Value: x}, vs)
				} else if s, ok := stringLit(val); ok {
					g.add(Hint{Kind: HintGoEnvVar, Key: name.Name, Default: s}, vs)
				} else if call, ok := val.(*ast.CallExpr); ok && len(call.Args) == 0 {
					if id, ok := call.Fun.(*ast.Ident); ok && g.funcs[id.Name] != nil {
						g.envFunc(name.Name, g.funcs[id.Name])
					}
				}
			}
		}
	}
}

// envFunc records var name = fn() when fn reads os.Getenv("X") and returns a
// string literal as its fallback; the hint cites the fallback's return line.
func (g *goHints) envFunc(name string, fn *ast.FuncDecl) {
	var variable, fallback string
	var at ast.Node
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if x, ok := g.getenv(nodeExpr(n)); ok && variable == "" {
			variable = x
		}
		if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
			if s, ok := stringLit(ret.Results[0]); ok {
				fallback, at = s, ret
			}
		}
		return true
	})
	if variable != "" || fallback != "" {
		if at == nil {
			at = fn
		}
		g.add(Hint{Kind: HintGoEnvVar, Key: name, Value: variable, Default: fallback}, at)
	}
}

func nodeExpr(n ast.Node) ast.Expr {
	e, _ := n.(ast.Expr)
	return e
}

func (g *goHints) visit(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.AssignStmt:
		if len(t.Lhs) == len(t.Rhs) {
			for i := range t.Lhs {
				g.assign(t.Lhs[i], t.Rhs[i], t)
			}
		}
	case *ast.ValueSpec:
		if len(t.Names) == len(t.Values) {
			for i := range t.Names {
				g.assign(t.Names[i], t.Values[i], t)
			}
		}
	case *ast.CallExpr:
		g.call(t)
	case *ast.CompositeLit:
		g.producerMessage(t)
	}
	return true
}

// assign records what lhs = rhs binds: an environment variable, a gRPC client,
// or an alias of another field or variable.
func (g *goHints) assign(lhs, rhs ast.Expr, at ast.Node) {
	name := bindingName(lhs)
	if name == "" || name == "_" {
		return
	}
	if x, ok := g.getenv(rhs); ok {
		g.add(Hint{Kind: HintGoEnvBinding, Key: name, Value: x}, at)
		return
	}
	if p, fn, _ := g.pkgCall(rhs); p != "" {
		if m := grpcClientNewRegex.FindStringSubmatch(fn); len(m) > 1 {
			g.add(Hint{Kind: HintGRPCClient, Key: name, Value: m[1]}, at)
		}
		return
	}
	if ref := g.localRef(rhs); ref != "" && ref != name {
		g.add(Hint{Kind: HintGoAlias, Key: name, Ref: ref}, at)
	}
}

// localRef is the field or variable an expression reads (x.f → f, v → v), or ""
// for anything else, including an imported package's member.
func (g *goHints) localRef(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok && g.imports[x.Name] != "" {
			return ""
		}
		return t.Sel.Name
	case *ast.Ident:
		if g.imports[t.Name] != "" || t.Name == "nil" || t.Name == "true" || t.Name == "false" {
			return ""
		}
		return t.Name
	}
	return ""
}

// clientFields records struct fields typed as a gRPC client, pkg.<Contract>Client.
func (g *goHints) clientFields(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, field := range st.Fields.List {
			sel, ok := field.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			if x, ok := sel.X.(*ast.Ident); !ok || g.imports[x.Name] == "" {
				continue
			}
			m := grpcClientTypeRegex.FindStringSubmatch(sel.Sel.Name)
			if len(m) < 2 || m[1] == "" || m[1] == "grpc" || strings.HasPrefix(m[1], "Client") {
				continue
			}
			for _, name := range field.Names {
				g.add(Hint{Kind: HintGRPCClient, Key: name.Name, Value: m[1]}, field)
			}
		}
		return true
	})
}

// goBuiltins is Go's universe scope: calls and conversions naming these are
// language surface, never intra-service symbols, and are skipped outright.
var goBuiltins = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true, "recover": true,
	"any": true, "bool": true, "byte": true, "comparable": true,
	"complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"uintptr": true,
}

// fileSymbols records the intra-service inventory of one Go file (F2): the
// package clause, every import statement, every top-level type and func, and
// every call site with its enclosing func. Deterministic and model-free.
func (g *goHints) fileSymbols(f *ast.File) {
	g.add(Hint{Kind: HintGoPackage, Key: f.Name.Name}, f.Name)

	for _, imp := range f.Imports {
		name, p := importName(imp)
		if name == "." {
			g.hasDotImport = true
		}
		g.add(Hint{Kind: HintGoImport, Key: p, Value: name}, imp)
	}

	for _, d := range f.Decls {
		switch t := d.(type) {
		case *ast.FuncDecl:
			mode := "unexported"
			if t.Name.IsExported() {
				mode = "exported"
			}
			g.add(Hint{Kind: HintGoFunc, Key: t.Name.Name, Value: recvTypeName(t.Recv), Mode: mode}, t.Name)
			if t.Body != nil {
				caller := callerName(t)
				ast.Inspect(t.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						g.symbolCall(call, caller)
					}
					return true
				})
			}
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					g.add(Hint{Kind: HintGoType, Key: s.Name.Name, Value: string(goTypeKind(s))}, s.Name)
				case *ast.ValueSpec:
					if t.Tok != token.VAR {
						continue
					}
					for _, v := range s.Values {
						ast.Inspect(v, func(n ast.Node) bool {
							if call, ok := n.(*ast.CallExpr); ok {
								g.symbolCall(call, "")
							}
							return true
						})
					}
				}
			}
		}
	}
}

// callerName is the local reference of a func's enclosing scope:
// "Recv.Name" for a method, "Name" for a plain function.
func callerName(fd *ast.FuncDecl) string {
	if recv := recvTypeName(fd.Recv); recv != "" {
		return recv + "." + fd.Name.Name
	}
	return fd.Name.Name
}

// recvTypeName is the dereferenced receiver type name ("Server" for both
// `s Server` and `s *Server[T]`), or "" for a plain function.
func recvTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return recvExprName(recv.List[0].Type)
}

func recvExprName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return recvExprName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.IndexExpr:
		return recvExprName(t.X)
	case *ast.IndexListExpr:
		return recvExprName(t.X)
	}
	return ""
}

// goTypeKind is the discrete shape of a type declaration: alias for
// `type A = B`, else the underlying composite, else other.
func goTypeKind(ts *ast.TypeSpec) domain.GoTypeKind {
	if ts.Assign.IsValid() {
		return domain.GoTypeAlias
	}
	switch ts.Type.(type) {
	case *ast.StructType:
		return domain.GoTypeStruct
	case *ast.InterfaceType:
		return domain.GoTypeInterface
	}
	return domain.GoTypeOther
}

// symbolCall records one call site: pkg-qualified (Ref carries the import
// path), local-qualified, or bare. Builtins are language surface and
// skipped; bare calls in dot-importing files are flagged so the resolve
// step never resolves them across packages.
func (g *goHints) symbolCall(call *ast.CallExpr, caller string) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok {
			if p, isImport := g.imports[id.Name]; isImport {
				g.add(Hint{Kind: HintGoCall, Key: caller, Value: fn.Sel.Name, Ref: p, Mode: "pkg"}, call)
			} else {
				g.add(Hint{Kind: HintGoCall, Key: caller, Value: fn.Sel.Name, Mode: "local"}, call)
			}
			return
		}
		g.add(Hint{Kind: HintGoCall, Key: caller, Value: fn.Sel.Name, Mode: "local"}, call)
	case *ast.Ident:
		if goBuiltins[fn.Name] {
			return
		}
		mode := "bare"
		if g.hasDotImport {
			mode = "bare-dotimport"
		}
		g.add(Hint{Kind: HintGoCall, Key: caller, Value: fn.Name, Mode: mode}, call)
	}
}

func (g *goHints) call(call *ast.CallExpr) {
	// recv.Method(...): a method call on a field or variable.
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.IsExported() {
		if recv := g.localRef(sel.X); recv != "" {
			g.add(Hint{Kind: HintGoMethodCall, Key: recv, Value: sel.Sel.Name}, call)
		}
	}

	// f(&field, "VAR"): a helper that maps an environment variable into a field.
	if len(call.Args) == 2 {
		if u, ok := call.Args[0].(*ast.UnaryExpr); ok && u.Op == token.AND {
			if x, ok := stringLit(call.Args[1]); ok && envVarName.MatchString(x) && bindingName(u.X) != "" {
				g.add(Hint{Kind: HintGoEnvBinding, Key: bindingName(u.X), Value: x}, call)
			}
		}
	}
	p, name, _ := g.pkgCall(call)
	if p == "" {
		return
	}
	if p == "net/http" {
		if i, ok := httpURLArg[name]; ok && i < len(call.Args) {
			if ref, route := g.urlParts(call.Args[i]); ref != "" {
				g.add(Hint{Kind: HintHTTPCall, Key: ref, Value: route}, call)
			}
		}
		return
	}
	if kafkaClients[p] && kafkaProducerCtors[name] && len(call.Args) > 0 {
		g.broker(call.Args[0], call)
		return
	}
	if !strings.Contains(p, ".") {
		return // standard library
	}
	for i, arg := range call.Args {
		ast.Inspect(arg, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok && g.imports[x.Name] == "" {
					g.add(Hint{Kind: HintGoCallArg, Value: p, Key: name, Index: i, Ref: sel.Sel.Name}, call)
				}
			}
			return true
		})
	}
}

// broker records where a Kafka producer's broker list comes from.
func (g *goHints) broker(arg ast.Expr, call *ast.CallExpr) {
	if x, ok := g.getenv(arg); ok {
		g.add(Hint{Kind: HintKafkaBroker, Mode: "env", Value: x}, call)
		return
	}
	ast.Inspect(arg, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.SelectorExpr:
			if x, ok := t.X.(*ast.Ident); ok && g.imports[x.Name] == "" {
				g.add(Hint{Kind: HintKafkaBroker, Mode: "field", Key: t.Sel.Name}, call)
			}
			return false
		case *ast.Ident:
			if i := g.paramIndex(t.Name); i >= 0 {
				g.add(Hint{Kind: HintKafkaBroker, Mode: "param", Key: g.fn.Name.Name, Index: i}, call)
			} else {
				g.add(Hint{Kind: HintKafkaBroker, Mode: "field", Key: t.Name}, call)
			}
		}
		return true
	})
}

// paramIndex is the position of name among the enclosing function's parameters.
func (g *goHints) paramIndex(name string) int {
	if g.fn == nil || g.fn.Type.Params == nil {
		return -1
	}
	i := 0
	for _, field := range g.fn.Type.Params.List {
		for _, n := range field.Names {
			if n.Name == name {
				return i
			}
			i++
		}
	}
	return -1
}

// producerMessage records the Topic of a Kafka producer message literal.
func (g *goHints) producerMessage(lit *ast.CompositeLit) {
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "ProducerMessage" {
		return
	}
	if x, ok := sel.X.(*ast.Ident); !ok || !kafkaClients[g.imports[x.Name]] {
		return
	}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if k, ok := kv.Key.(*ast.Ident); !ok || k.Name != "Topic" {
			continue
		}
		h := Hint{Kind: HintKafkaTopic, Context: g.dir}
		switch v := kv.Value.(type) {
		case *ast.BasicLit:
			s, _ := stringLit(v)
			h.Mode, h.Value = "literal", s
		case *ast.CallExpr:
			x, ok := g.getenv(v)
			if !ok {
				continue
			}
			h.Mode, h.Value = "env", x
		case *ast.SelectorExpr:
			if x, ok := v.X.(*ast.Ident); ok && g.imports[x.Name] != "" {
				h.Mode, h.Key, h.Value = "var", v.Sel.Name, g.imports[x.Name]
			} else {
				h.Mode, h.Key = "field", v.Sel.Name
			}
		case *ast.Ident:
			h.Mode, h.Key = "var", v.Name
		default:
			continue
		}
		g.add(h, kv)
	}
}

// httpURLArg is the URL argument's position in the net/http request functions.
var httpURLArg = map[string]int{"NewRequest": 1, "NewRequestWithContext": 2, "Get": 0, "Head": 0, "Post": 0, "PostForm": 0}

// urlParts splits a request URL expression addr + "/route" into the field or
// variable holding the address and the literal route ("" for addr alone).
func (g *goHints) urlParts(expr ast.Expr) (ref, route string) {
	if b, ok := expr.(*ast.BinaryExpr); ok && b.Op == token.ADD {
		if s, ok := stringLit(b.Y); ok && strings.HasPrefix(s, "/") {
			return g.localRef(b.X), s
		}
		return "", ""
	}
	return g.localRef(expr), ""
}
