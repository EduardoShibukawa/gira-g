package extractors

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// GoASTExtractor extracts Service, Library, and Contract relationships from Go code.
type GoASTExtractor struct{}

// NewGoASTExtractor creates a new Layer 3 Go AST extractor.
func NewGoASTExtractor() *GoASTExtractor {
	return &GoASTExtractor{}
}

func (e *GoASTExtractor) Name() string {
	return "Layer3GoAST"
}

func (e *GoASTExtractor) Layer() Layer {
	return LayerLanguageAST
}

// CanExtract matches Go source files.
func (e *GoASTExtractor) CanExtract(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".go"
}

var (
	grpcServerRegisterRegex = regexp.MustCompile(`^Register([A-Za-z0-9_]+)Server$`)
	grpcServerEmbedRegex    = regexp.MustCompile(`^(?:Unimplemented)?([A-Za-z0-9_]+)Server$`)
	grpcClientNewRegex      = regexp.MustCompile(`^New([A-Za-z0-9_]+)Client$`)
	grpcClientTypeRegex     = regexp.MustCompile(`^([A-Za-z0-9_]+)Client$`)
)

// stdlibHealthContracts is the fixed literal denylist for J1 (SC1): stdlib
// gRPC health registrations are infrastructure, never business contracts.
// Literal-only on purpose: no pattern that could swallow a business name.
func stdlibHealthContract(name string) bool {
	return name == "Health"
}

func deriveServiceName(filePath string, pkgName string) string {
	clean := filepath.ToSlash(filepath.Clean(filePath))
	parts := strings.Split(clean, "/")
	for i, part := range parts {
		lower := strings.ToLower(part)
		if (lower == "src" || lower == "services" || lower == "service" || lower == "apps" || lower == "app" || lower == "cmd") && i+1 < len(parts) {
			next := parts[i+1]
			if !strings.HasSuffix(next, ".go") {
				return next
			}
		}
	}
	if len(parts) > 1 && !strings.HasSuffix(parts[0], ".go") {
		return parts[0]
	}
	return pkgName
}

func sanitizeRelID(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ":", "-")
	return strings.ToLower(s)
}

// Extract parses Go source code, extracting Service, Library imports, and gRPC mediation relationships.
func (e *GoASTExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	// Skip test files from architectural graph extraction
	if strings.HasSuffix(file.Path, "_test.go") {
		return nil, nil, nil, nil
	}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file.Path, file.Content, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse Go source in %s: %w", file.Path, err)
	}

	discoveredAt := time.Now().UTC()
	pkgName := astFile.Name.Name
	serviceName := deriveServiceName(file.Path, pkgName)
	serviceID := fmt.Sprintf("service:%s", serviceName)

	nodeMap := make(map[string]domain.Node)
	relMap := make(map[string]domain.Relationship)

	pkgLine := fset.Position(astFile.Package).Line

	// 1. Create or register the Service node
	serviceNode := domain.Node{
		ID:   serviceID,
		Kind: domain.NodeService,
		Name: serviceName,
		Metadata: map[string]any{
			"package": pkgName,
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
					StartLine: pkgLine,
					EndLine:   pkgLine,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      discoveredAt,
			},
		},
	}
	nodeMap[serviceID] = serviceNode

	// 2. Extract Library nodes and imports relationships
	for _, imp := range astFile.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		// Consider external packages (containing a domain dot) as architectural libraries
		if strings.Contains(importPath, ".") {
			startLine := fset.Position(imp.Pos()).Line
			endLine := fset.Position(imp.End()).Line

			libID := fmt.Sprintf("library:%s", importPath)
			libEvidence := domain.EvidenceRecord{
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
					StartLine: startLine,
					EndLine:   endLine,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: "go",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      discoveredAt,
			}

			if existing, found := nodeMap[libID]; found {
				existing.Evidence = append(existing.Evidence, libEvidence)
				nodeMap[libID] = existing
			} else {
				nodeMap[libID] = domain.Node{
					ID:   libID,
					Kind: domain.NodeLibrary,
					Name: importPath,
					Metadata: map[string]any{
						"import_path": importPath,
					},
					Evidence: []domain.EvidenceRecord{libEvidence},
				}
			}

			relID := fmt.Sprintf("rel:%s-imports-%s", serviceName, sanitizeRelID(importPath))
			if existingRel, found := relMap[relID]; found {
				existingRel.Evidence = append(existingRel.Evidence, libEvidence)
				relMap[relID] = existingRel
			} else {
				relMap[relID] = domain.Relationship{
					ID:       relID,
					SourceID: serviceID,
					TargetID: libID,
					Verb:     domain.VerbImports,
					Evidence: []domain.EvidenceRecord{libEvidence},
				}
			}
		}
	}

	// 3. Inspect AST for gRPC server providers and client consumers
	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		select {
		case <-ctx.Done():
			return false
		default:
		}

		// A. Function calls: Register<Name>Server, or New<Name>Client (the Contract node only)
		if call, ok := n.(*ast.CallExpr); ok {
			var funcName string
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				funcName = fn.Sel.Name
			case *ast.Ident:
				funcName = fn.Name
			}

			startLine := fset.Position(call.Pos()).Line
			endLine := fset.Position(call.End()).Line

			// Server registration: Register<Contract>Server
			if m := grpcServerRegisterRegex.FindStringSubmatch(funcName); len(m) > 1 {
				contractName := m[1]
				if stdlibHealthContract(contractName) {
					return true
				}
				contractID := fmt.Sprintf("contract:%s", contractName)

				e.ensureContractNode(nodeMap, contractID, contractName, file, startLine, endLine, discoveredAt)

				relID := fmt.Sprintf("rel:%s-provides-%s", serviceName, sanitizeRelID(contractName))
				e.recordRelationship(relMap, relID, serviceID, contractID, domain.VerbProvides, file, startLine, endLine, discoveredAt)
			}

			// Client invocation: New<Contract>Client
			if m := grpcClientNewRegex.FindStringSubmatch(funcName); len(m) > 1 {
				contractName := m[1]
				if stdlibHealthContract(contractName) {
					return true
				}
				contractID := fmt.Sprintf("contract:%s", contractName)

				// Building a client is not consuming: the consumes edge comes from a
				// method call on it, joined across the package (Hints, ADR 0013).
				e.ensureContractNode(nodeMap, contractID, contractName, file, startLine, endLine, discoveredAt)
			}
		}

		// B. Struct embedding: Unimplemented<Contract>Server or <Contract>Server
		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, isStruct := ts.Type.(*ast.StructType); isStruct && st.Fields != nil {
				for _, f := range st.Fields.List {
					if len(f.Names) == 0 { // embedded struct
						var typeName string
						switch t := f.Type.(type) {
						case *ast.SelectorExpr:
							typeName = t.Sel.Name
						case *ast.Ident:
							typeName = t.Name
						}

						if m := grpcServerEmbedRegex.FindStringSubmatch(typeName); len(m) > 1 {
							contractName := m[1]
							if contractName == "" || contractName == "grpc" || stdlibHealthContract(contractName) {
								continue
							}
							startLine := fset.Position(f.Pos()).Line
							endLine := fset.Position(f.End()).Line
							contractID := fmt.Sprintf("contract:%s", contractName)

							e.ensureContractNode(nodeMap, contractID, contractName, file, startLine, endLine, discoveredAt)

							relID := fmt.Sprintf("rel:%s-provides-%s", serviceName, sanitizeRelID(contractName))
							e.recordRelationship(relMap, relID, serviceID, contractID, domain.VerbProvides, file, startLine, endLine, discoveredAt)
						}
					}
				}
			}
		}

		// C. Struct field: client field like `pb.<Contract>Client` (the Contract node only)
		if f, ok := n.(*ast.Field); ok && len(f.Names) > 0 {
			if sel, isSel := f.Type.(*ast.SelectorExpr); isSel {
				typeName := sel.Sel.Name
				if m := grpcClientTypeRegex.FindStringSubmatch(typeName); len(m) > 1 {
					contractName := m[1]
					if contractName != "" && contractName != "grpc" && !strings.HasPrefix(contractName, "Client") && !stdlibHealthContract(contractName) {
						startLine := fset.Position(f.Pos()).Line
						endLine := fset.Position(f.End()).Line
						contractID := fmt.Sprintf("contract:%s", contractName)

						// Declaring a client field is not consuming either (see above).
						e.ensureContractNode(nodeMap, contractID, contractName, file, startLine, endLine, discoveredAt)
					}
				}
			}
		}

		return true
	})

	nodes := make([]domain.Node, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}

	relationships := make([]domain.Relationship, 0, len(relMap))
	for _, r := range relMap {
		relationships = append(relationships, r)
	}

	return nodes, relationships, nil, nil
}

func (e *GoASTExtractor) ensureContractNode(nodeMap map[string]domain.Node, contractID string, contractName string, file TargetFile, startLine, endLine int, discoveredAt time.Time) {
	ev := domain.EvidenceRecord{
		Locator: domain.Locator{
			Repo:      file.Repo,
			CommitSHA: file.CommitSHA,
			Path:      file.Path,
			StartLine: startLine,
			EndLine:   endLine,
		},
		Origin:            domain.OriginExtracted,
		DeploymentProfile: "grpc",
		Status:            domain.StatusCurrent,
		DiscoveredAt:      discoveredAt,
	}

	if existing, found := nodeMap[contractID]; found {
		existing.Evidence = append(existing.Evidence, ev)
		nodeMap[contractID] = existing
	} else {
		nodeMap[contractID] = domain.Node{
			ID:       contractID,
			Kind:     domain.NodeContract,
			Name:     contractName,
			Evidence: []domain.EvidenceRecord{ev},
		}
	}
}

func (e *GoASTExtractor) recordRelationship(relMap map[string]domain.Relationship, relID, sourceID, targetID string, verb domain.RelationVerb, file TargetFile, startLine, endLine int, discoveredAt time.Time) {
	ev := domain.EvidenceRecord{
		Locator: domain.Locator{
			Repo:      file.Repo,
			CommitSHA: file.CommitSHA,
			Path:      file.Path,
			StartLine: startLine,
			EndLine:   endLine,
		},
		Origin:            domain.OriginExtracted,
		DeploymentProfile: "grpc",
		Status:            domain.StatusCurrent,
		DiscoveredAt:      discoveredAt,
	}

	if existing, found := relMap[relID]; found {
		existing.Evidence = append(existing.Evidence, ev)
		relMap[relID] = existing
	} else {
		relMap[relID] = domain.Relationship{
			ID:       relID,
			SourceID: sourceID,
			TargetID: targetID,
			Verb:     verb,
			Evidence: []domain.EvidenceRecord{ev},
		}
	}
}
