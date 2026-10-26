package architecture

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// PackageInfo holds AST information for a discovered package.
type PackageInfo struct {
	Name       string
	Path       string // e.g. "internal/manipulator"
	ImportPath string // e.g. "github.com/EduardoShibukawa/gira-g/internal/manipulator"
	Domain     DomainID
	Doc        string
	Files      []string
	Imports    []string // internal packages imported
	Types      []TypeInfo
	Functions  []FuncInfo
}

// TypeInfo holds AST information for a declared type.
type TypeInfo struct {
	Name      string
	Kind      string // "struct", "interface", "type"
	Doc       string
	Methods   []FuncInfo
	Locator   domain.Locator
	Signature string
}

// FuncInfo holds AST information for a declared function or method.
type FuncInfo struct {
	Name      string
	Receiver  string
	Signature string
	Doc       string
	Locator   domain.Locator
	Exported  bool
}

// ClassifyDomain assigns a package path to an architectural domain.
func ClassifyDomain(pkgRelPath string) DomainID {
	clean := filepath.ToSlash(pkgRelPath)
	clean = strings.TrimPrefix(clean, "./")

	switch {
	case strings.HasPrefix(clean, "cmd/") || strings.Contains(clean, "internal/mcp"):
		return DomainEntrypoints
	case strings.Contains(clean, "internal/manipulator"):
		return DomainManipulator
	case strings.Contains(clean, "internal/extractors"):
		return DomainExtractors
	case strings.Contains(clean, "internal/slices") || strings.Contains(clean, "internal/domain") ||
		strings.Contains(clean, "internal/indexer") || strings.Contains(clean, "internal/manifest") ||
		strings.Contains(clean, "internal/tokens") || strings.Contains(clean, "internal/skills") ||
		strings.Contains(clean, "internal/tools") || strings.Contains(clean, "internal/human"):
		return DomainCore
	case strings.Contains(clean, "internal/gate") || strings.Contains(clean, "internal/eval") ||
		strings.Contains(clean, "internal/archdiff") || strings.Contains(clean, "internal/corpus") ||
		strings.Contains(clean, "internal/gitdiff"):
		return DomainGovernance
	case strings.Contains(clean, "internal/store") || strings.Contains(clean, "internal/gitcache"):
		return DomainStorage
	default:
		return DomainCore
	}
}

// ScanRepository scans all Go packages in root and returns parsed package metadata.
func ScanRepository(repoRoot string) ([]PackageInfo, error) {
	var packages []PackageInfo
	fset := token.NewFileSet()

	modulePrefix := "github.com/EduardoShibukawa/gira-g"

	err := filepath.WalkDir(repoRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" || name == "bench" || name == "fixtures" {
				return filepath.SkipDir
			}

			rel, _ := filepath.Rel(repoRoot, p)
			rel = filepath.ToSlash(rel)
			if rel != "." && !strings.HasPrefix(rel, "cmd") && !strings.HasPrefix(rel, "internal") {
				return nil
			}

			pkgs, parseErr := parser.ParseDir(fset, p, func(fi os.FileInfo) bool {
				return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
			}, parser.ParseComments)
			if parseErr != nil || len(pkgs) == 0 {
				return nil
			}

			for pkgName, pkgAst := range pkgs {
				info := PackageInfo{
					Name:       pkgName,
					Path:       rel,
					ImportPath: modulePrefix + "/" + rel,
					Domain:     ClassifyDomain(rel),
				}

				internalImportSet := make(map[string]bool)
				var fileNames []string

				for fName, fileAst := range pkgAst.Files {
					fileRel, _ := filepath.Rel(repoRoot, fName)
					fileRel = filepath.ToSlash(fileRel)
					fileNames = append(fileNames, fileRel)

					if info.Doc == "" && fileAst.Doc != nil {
						info.Doc = strings.TrimSpace(fileAst.Doc.Text())
					}

					// Collect internal imports
					for _, imp := range fileAst.Imports {
						val := strings.Trim(imp.Path.Value, `"`)
						if strings.HasPrefix(val, modulePrefix) {
							importedRel := strings.TrimPrefix(val, modulePrefix+"/")
							internalImportSet[importedRel] = true
						}
					}

					// Parse declarations
					for _, decl := range fileAst.Decls {
						switch d := decl.(type) {
						case *ast.GenDecl:
							if d.Tok == token.TYPE {
								for _, spec := range d.Specs {
									typeSpec, ok := spec.(*ast.TypeSpec)
									if !ok {
										continue
									}
									kind := "type"
									switch typeSpec.Type.(type) {
									case *ast.StructType:
										kind = "struct"
									case *ast.InterfaceType:
										kind = "interface"
									}

									start := fset.Position(d.Pos())
									end := fset.Position(d.End())

									doc := ""
									if d.Doc != nil {
										doc = strings.TrimSpace(d.Doc.Text())
									} else if typeSpec.Doc != nil {
										doc = strings.TrimSpace(typeSpec.Doc.Text())
									}

									info.Types = append(info.Types, TypeInfo{
										Name: typeSpec.Name.Name,
										Kind: kind,
										Doc:  doc,
										Locator: domain.Locator{
											Repo:      "local",
											CommitSHA: "head",
											Path:      fileRel,
											StartLine: start.Line,
											EndLine:   end.Line,
										},
										Signature: fmt.Sprintf("type %s %s", typeSpec.Name.Name, kind),
									})
								}
							}

						case *ast.FuncDecl:
							start := fset.Position(d.Pos())
							end := fset.Position(d.End())

							receiver := ""
							if d.Recv != nil && len(d.Recv.List) > 0 {
								recvType := d.Recv.List[0].Type
								if star, ok := recvType.(*ast.StarExpr); ok {
									if id, ok := star.X.(*ast.Ident); ok {
										receiver = id.Name
									}
								} else if id, ok := recvType.(*ast.Ident); ok {
									receiver = id.Name
								}
							}

							doc := ""
							if d.Doc != nil {
								doc = strings.TrimSpace(d.Doc.Text())
							}

							sig := formatFuncSignature(d)

							fn := FuncInfo{
								Name:      d.Name.Name,
								Receiver:  receiver,
								Signature: sig,
								Doc:       doc,
								Exported:  d.Name.IsExported(),
								Locator: domain.Locator{
									Repo:      "local",
									CommitSHA: "head",
									Path:      fileRel,
									StartLine: start.Line,
									EndLine:   end.Line,
								},
							}

							if receiver != "" {
								// Attach method to matching type if found
								found := false
								for idx := range info.Types {
									if info.Types[idx].Name == receiver {
										info.Types[idx].Methods = append(info.Types[idx].Methods, fn)
										found = true
										break
									}
								}
								if !found {
									info.Functions = append(info.Functions, fn)
								}
							} else {
								info.Functions = append(info.Functions, fn)
							}
						}
					}
				}

				for imp := range internalImportSet {
					info.Imports = append(info.Imports, imp)
				}
				sort.Strings(info.Imports)
				sort.Strings(fileNames)
				info.Files = fileNames

				packages = append(packages, info)
			}
		}
		return nil
	})

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Path < packages[j].Path
	})

	return packages, err
}

func formatFuncSignature(d *ast.FuncDecl) string {
	var sb strings.Builder
	sb.WriteString("func ")
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recvType := d.Recv.List[0].Type
		recvName := ""
		if len(d.Recv.List[0].Names) > 0 {
			recvName = d.Recv.List[0].Names[0].Name + " "
		}
		if star, ok := recvType.(*ast.StarExpr); ok {
			if id, ok := star.X.(*ast.Ident); ok {
				sb.WriteString(fmt.Sprintf("(%s*%s) ", recvName, id.Name))
			}
		} else if id, ok := recvType.(*ast.Ident); ok {
			sb.WriteString(fmt.Sprintf("(%s%s) ", recvName, id.Name))
		}
	}
	sb.WriteString(d.Name.Name)
	sb.WriteString("(")
	if d.Type.Params != nil {
		for i, p := range d.Type.Params.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			names := make([]string, len(p.Names))
			for j, n := range p.Names {
				names[j] = n.Name
			}
			paramName := strings.Join(names, ", ")
			paramType := exprToString(p.Type)
			if paramName != "" {
				sb.WriteString(paramName + " " + paramType)
			} else {
				sb.WriteString(paramType)
			}
		}
	}
	sb.WriteString(")")
	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		sb.WriteString(" ")
		if len(d.Type.Results.List) > 1 {
			sb.WriteString("(")
		}
		for i, r := range d.Type.Results.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(exprToString(r.Type))
		}
		if len(d.Type.Results.List) > 1 {
			sb.WriteString(")")
		}
	}
	return sb.String()
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + exprToString(e.Elt)
	default:
		return "any"
	}
}
