package extractors

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// TypeScriptASTExtractor extracts Library, Contract nodes, methods, and functions from TypeScript (.ts, .tsx) files.
type TypeScriptASTExtractor struct{}

// NewTypeScriptASTExtractor creates a new Layer 3 TypeScript extractor.
func NewTypeScriptASTExtractor() *TypeScriptASTExtractor {
	return &TypeScriptASTExtractor{}
}

func (e *TypeScriptASTExtractor) Name() string {
	return "Layer3TypeScriptAST"
}

func (e *TypeScriptASTExtractor) Layer() Layer {
	return LayerLanguageAST
}

// CanExtract matches TypeScript source files, excluding declaration files (.d.ts) and node_modules.
func (e *TypeScriptASTExtractor) CanExtract(path string) bool {
	norm := filepath.ToSlash(path)
	if strings.Contains(norm, "node_modules/") || strings.Contains(norm, "/.git/") {
		return false
	}
	if strings.HasSuffix(norm, ".d.ts") || strings.HasSuffix(norm, ".d.tsx") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(norm))
	return ext == ".ts" || ext == ".tsx" || ext == ".mts" || ext == ".cts"
}

var (
	tsClassRegex     = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:abstract\s+)?class\s+([A-Za-z0-9_$]+)(?:<[^>]+>)?(?:\s+extends\s+([A-Za-z0-9_$.< >]+))?(?:\s+implements\s+([^{]+))?`)
	tsInterfaceRegex = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?interface\s+([A-Za-z0-9_$]+)(?:<[^>]+>)?(?:\s+extends\s+([^{]+))?`)
	tsTypeRegex      = regexp.MustCompile(`^(?:export\s+)?type\s+([A-Za-z0-9_$]+)(?:<[^>]+>)?\s*=`)
	tsFuncRegex      = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:async\s+)?function(?:\s+([A-Za-z0-9_$]+))\s*(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*:\s*([^;{]+))?`)
	tsConstFuncRegex = regexp.MustCompile(`^(?:export\s+)?(?:const|let|var)\s+([A-Za-z0-9_$]+)\s*(?::\s*[^=]+)?\s*=\s*(?:async\s*)?\(([^)]*)\)(?:\s*:\s*([^=>]+))?\s*=>`)
	tsImportRegex    = regexp.MustCompile(`^import\s+(?:type\s+)?(?:\{[^}]*\}|[A-Za-z0-9_*$]+(?:\s*,\s*\{[^}]*\})?)\s+from\s+['"]([^'"]+)['"]`)
	tsMethodRegex    = regexp.MustCompile(`^(?:\s*)(?:public\s+|private\s+|protected\s+|static\s+|async\s+)*([A-Za-z0-9_$]+)\s*(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*:\s*([^;{]+))?`)
)

type tsMethodInfo struct {
	Name       string `json:"name"`
	Parameters string `json:"parameters"`
	ReturnType string `json:"return_type,omitempty"`
	StartLine  int    `json:"start_line"`
}

type tsContractInfo struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"` // class, interface, type
	Bases     []string       `json:"bases,omitempty"`
	StartLine int            `json:"start_line"`
	EndLine   int            `json:"end_line"`
	Methods   []tsMethodInfo `json:"methods"`
}

// Extract parses a TypeScript source file and produces domain Nodes and Relationships.
func (e *TypeScriptASTExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	lines := strings.Split(string(file.Content), "\n")
	discoveredAt := time.Now().UTC()

	var nodes []domain.Node
	var rels []domain.Relationship

	var contracts []tsContractInfo
	var currentContract *tsContractInfo
	var openBraces int
	var contractBraceLevel int
	var topLevelFuncs []tsMethodInfo
	var fileImports []string

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1
		trimmed := strings.TrimSpace(line)

		// Rastreia contagem de chaves para saber término de blocos
		deltaOpen := strings.Count(line, "{") - strings.Count(line, "}")

		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			openBraces += deltaOpen
			if currentContract != nil && openBraces <= contractBraceLevel && deltaOpen < 0 {
				currentContract.EndLine = lineNum
				contracts = append(contracts, *currentContract)
				currentContract = nil
			}
			continue
		}

		// Imports
		if m := tsImportRegex.FindStringSubmatch(trimmed); m != nil {
			fileImports = append(fileImports, m[1])
			openBraces += deltaOpen
			continue
		}

		// Classes
		if m := tsClassRegex.FindStringSubmatch(trimmed); m != nil && currentContract == nil {
			className := m[1]
			var bases []string
			if len(m) > 2 && strings.TrimSpace(m[2]) != "" {
				bases = append(bases, strings.TrimSpace(m[2]))
			}
			if len(m) > 3 && strings.TrimSpace(m[3]) != "" {
				for _, imp := range strings.Split(m[3], ",") {
					clean := strings.TrimSpace(imp)
					if clean != "" {
						bases = append(bases, clean)
					}
				}
			}

			contractBraceLevel = openBraces
			openBraces += deltaOpen
			currentContract = &tsContractInfo{
				Name:      className,
				Kind:      "class",
				Bases:     bases,
				StartLine: lineNum,
				EndLine:   len(lines),
				Methods:   []tsMethodInfo{},
			}
			continue
		}

		// Interfaces
		if m := tsInterfaceRegex.FindStringSubmatch(trimmed); m != nil && currentContract == nil {
			ifaceName := m[1]
			var bases []string
			if len(m) > 2 && strings.TrimSpace(m[2]) != "" {
				for _, b := range strings.Split(m[2], ",") {
					clean := strings.TrimSpace(b)
					if clean != "" {
						bases = append(bases, clean)
					}
				}
			}

			contractBraceLevel = openBraces
			openBraces += deltaOpen
			currentContract = &tsContractInfo{
				Name:      ifaceName,
				Kind:      "interface",
				Bases:     bases,
				StartLine: lineNum,
				EndLine:   len(lines),
				Methods:   []tsMethodInfo{},
			}
			continue
		}

		// Types
		if m := tsTypeRegex.FindStringSubmatch(trimmed); m != nil && currentContract == nil {
			typeName := m[1]
			contracts = append(contracts, tsContractInfo{
				Name:      typeName,
				Kind:      "type",
				StartLine: lineNum,
				EndLine:   lineNum,
			})
			openBraces += deltaOpen
			continue
		}

		// Top-level functions
		if currentContract == nil {
			if m := tsFuncRegex.FindStringSubmatch(trimmed); m != nil && m[1] != "" {
				retType := ""
				if len(m) > 3 {
					retType = strings.TrimSpace(m[3])
				}
				topLevelFuncs = append(topLevelFuncs, tsMethodInfo{
					Name:       m[1],
					Parameters: strings.TrimSpace(m[2]),
					ReturnType: retType,
					StartLine:  lineNum,
				})
			} else if m := tsConstFuncRegex.FindStringSubmatch(trimmed); m != nil && m[1] != "" {
				retType := ""
				if len(m) > 3 {
					retType = strings.TrimSpace(m[3])
				}
				topLevelFuncs = append(topLevelFuncs, tsMethodInfo{
					Name:       m[1],
					Parameters: strings.TrimSpace(m[2]),
					ReturnType: retType,
					StartLine:  lineNum,
				})
			}
		} else {
			// Métodos dentro de classe / interface
			if m := tsMethodRegex.FindStringSubmatch(trimmed); m != nil {
				name := m[1]
				if name != "if" && name != "for" && name != "while" && name != "switch" && name != "catch" {
					retType := ""
					if len(m) > 3 {
						retType = strings.TrimSpace(m[3])
					}
					currentContract.Methods = append(currentContract.Methods, tsMethodInfo{
						Name:       name,
						Parameters: strings.TrimSpace(m[2]),
						ReturnType: retType,
						StartLine:  lineNum,
					})
				}
			}
		}

		openBraces += deltaOpen
		if currentContract != nil && openBraces <= contractBraceLevel && deltaOpen < 0 {
			currentContract.EndLine = lineNum
			contracts = append(contracts, *currentContract)
			currentContract = nil
		}
	}

	if currentContract != nil {
		contracts = append(contracts, *currentContract)
	}

	// Cria nós de contratos (classes, interfaces, types)
	for _, c := range contracts {
		nodeID := fmt.Sprintf("contract:%s:%s", filepath.ToSlash(file.Path), c.Name)
		methodsList := make([]map[string]any, len(c.Methods))
		for i, m := range c.Methods {
			methodsList[i] = map[string]any{
				"name":        m.Name,
				"parameters":  m.Parameters,
				"return_type": m.ReturnType,
				"line":        m.StartLine,
			}
		}

		loc := domain.Locator{
			Repo:      file.Repo,
			CommitSHA: file.CommitSHA,
			Path:      file.Path,
			StartLine: c.StartLine,
			EndLine:   c.EndLine,
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeContract,
			Name:        c.Name,
			Description: fmt.Sprintf("TypeScript %s %s with %d methods", c.Kind, c.Name, len(c.Methods)),
			Metadata: map[string]any{
				"language":      "typescript",
				"ts_kind":       c.Kind,
				"bases":         c.Bases,
				"methods":       methodsList,
				"declared_file": file.Path,
			},
			Evidence: []domain.EvidenceRecord{
				{
					Locator:      loc,
					Origin:       domain.OriginExtracted,
					RuleID:       "typescript-ast",
					Status:       domain.StatusCurrent,
					DiscoveredAt: discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)
	}

	// Se houver funções no nível superior ou se não houver classes, cria nó Library para o módulo
	if len(topLevelFuncs) > 0 || len(contracts) == 0 {
		modName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
		nodeID := fmt.Sprintf("lib:%s:%s", filepath.ToSlash(file.Path), modName)
		loc := domain.Locator{
			Repo:      file.Repo,
			CommitSHA: file.CommitSHA,
			Path:      file.Path,
			StartLine: 1,
			EndLine:   len(lines),
		}

		funcsList := make([]map[string]any, len(topLevelFuncs))
		for i, f := range topLevelFuncs {
			funcsList[i] = map[string]any{
				"name":        f.Name,
				"parameters":  f.Parameters,
				"return_type": f.ReturnType,
				"line":        f.StartLine,
			}
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeLibrary,
			Name:        modName,
			Description: fmt.Sprintf("TypeScript module %s with %d functions", modName, len(topLevelFuncs)),
			Metadata: map[string]any{
				"language":  "typescript",
				"functions": funcsList,
				"imports":   fileImports,
			},
			Evidence: []domain.EvidenceRecord{
				{
					Locator:      loc,
					Origin:       domain.OriginExtracted,
					RuleID:       "typescript-module-ast",
					Status:       domain.StatusCurrent,
					DiscoveredAt: discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)
	}

	return nodes, rels, nil, nil
}
