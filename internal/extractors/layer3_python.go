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

// PythonASTExtractor extracts Library, Contract nodes and methods from Python (.py) source files.
type PythonASTExtractor struct{}

// NewPythonASTExtractor creates a new Layer 3 Python extractor.
func NewPythonASTExtractor() *PythonASTExtractor {
	return &PythonASTExtractor{}
}

func (e *PythonASTExtractor) Name() string {
	return "Layer3PythonAST"
}

func (e *PythonASTExtractor) Layer() Layer {
	return LayerLanguageAST
}

// CanExtract matches Python source files.
func (e *PythonASTExtractor) CanExtract(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".py" {
		return false
	}
	// Ignora caches e scripts avulsos
	norm := filepath.ToSlash(path)
	if strings.Contains(norm, "__pycache__") || strings.Contains(norm, ".pytest_cache") ||
		strings.HasPrefix(norm, "scripts/") || strings.Contains(norm, "/scripts/") {
		return false
	}
	return true
}

var (
	pyClassRegex  = regexp.MustCompile(`^class\s+([A-Za-z0-9_]+)(?:\(([^)]*)\))?:`)
	pyFuncRegex   = regexp.MustCompile(`^(?:\s*)def\s+([A-Za-z0-9_]+)\s*\(([^)]*)\)(?:\s*->\s*([^:]+))?:`)
	pyImportRegex = regexp.MustCompile(`^(?:from\s+([A-Za-z0-9_.]+)\s+import\s+([^#\n]+)|import\s+([A-Za-z0-9_., ]+))`)
)

type pyMethodInfo struct {
	Name       string `json:"name"`
	Parameters string `json:"parameters"`
	ReturnType string `json:"return_type,omitempty"`
	StartLine  int    `json:"start_line"`
}

type pyClassInfo struct {
	Name      string         `json:"name"`
	Bases     []string       `json:"bases,omitempty"`
	StartLine int            `json:"start_line"`
	EndLine   int            `json:"end_line"`
	Docstring string         `json:"docstring,omitempty"`
	Methods   []pyMethodInfo `json:"methods"`
}

// Extract parses a Python source file and produces domain Nodes and Relationships.
func (e *PythonASTExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	lines := strings.Split(string(file.Content), "\n")
	discoveredAt := time.Now().UTC()

	var nodes []domain.Node
	var rels []domain.Relationship

	// Mapeia classes e métodos no arquivo
	var classes []pyClassInfo
	var currentClass *pyClassInfo
	var fileImports []string

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Imports
		if m := pyImportRegex.FindStringSubmatch(trimmed); m != nil {
			if m[1] != "" {
				fileImports = append(fileImports, m[1])
			} else if m[3] != "" {
				for _, part := range strings.Split(m[3], ",") {
					pkg := strings.TrimSpace(part)
					if pkg != "" {
						fileImports = append(fileImports, pkg)
					}
				}
			}
			continue
		}

		// Classes no nível superior
		if m := pyClassRegex.FindStringSubmatch(line); m != nil {
			if currentClass != nil {
				currentClass.EndLine = lineNum - 1
				classes = append(classes, *currentClass)
			}
			className := m[1]
			var bases []string
			if len(m) > 2 && strings.TrimSpace(m[2]) != "" {
				for _, b := range strings.Split(m[2], ",") {
					b = strings.TrimSpace(b)
					if b != "" {
						bases = append(bases, b)
					}
				}
			}
			currentClass = &pyClassInfo{
				Name:      className,
				Bases:     bases,
				StartLine: lineNum,
				EndLine:   len(lines),
				Methods:   []pyMethodInfo{},
			}
			continue
		}

		// Métodos / Funções
		if m := pyFuncRegex.FindStringSubmatch(line); m != nil {
			funcName := m[1]
			params := strings.TrimSpace(m[2])
			retType := ""
			if len(m) > 3 {
				retType = strings.TrimSpace(m[3])
			}

			method := pyMethodInfo{
				Name:       funcName,
				Parameters: params,
				ReturnType: retType,
				StartLine:  lineNum,
			}

			if currentClass != nil && (strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t")) {
				currentClass.Methods = append(currentClass.Methods, method)
			}
		}
	}

	if currentClass != nil {
		classes = append(classes, *currentClass)
	}

	// Cria nós para as classes encontradas
	for _, cls := range classes {
		nodeID := fmt.Sprintf("contract:%s:%s", filepath.ToSlash(file.Path), cls.Name)
		methodsList := make([]map[string]any, len(cls.Methods))
		for i, m := range cls.Methods {
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
			StartLine: cls.StartLine,
			EndLine:   cls.EndLine,
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeContract,
			Name:        cls.Name,
			Description: fmt.Sprintf("Python class %s with %d methods", cls.Name, len(cls.Methods)),
			Metadata: map[string]any{
				"language": "python",
				"bases":    cls.Bases,
				"methods":  methodsList,
			},
			Evidence: []domain.EvidenceRecord{
				{
					Locator:      loc,
					Origin:       domain.OriginExtracted,
					RuleID:       "python-class-ast",
					Status:       domain.StatusCurrent,
					DiscoveredAt: discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)
	}

	// Se não houver classes, mas for um módulo python, cria nó Library para o módulo
	if len(nodes) == 0 && len(file.Content) > 0 {
		modName := strings.TrimSuffix(filepath.Base(file.Path), ".py")
		nodeID := fmt.Sprintf("lib:%s:%s", filepath.ToSlash(file.Path), modName)
		loc := domain.Locator{
			Repo:      file.Repo,
			CommitSHA: file.CommitSHA,
			Path:      file.Path,
			StartLine: 1,
			EndLine:   len(lines),
		}
		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeLibrary,
			Name:        modName,
			Description: fmt.Sprintf("Python module %s", modName),
			Metadata: map[string]any{
				"language": "python",
				"imports":  fileImports,
			},
			Evidence: []domain.EvidenceRecord{
				{
					Locator:      loc,
					Origin:       domain.OriginExtracted,
					RuleID:       "python-module-ast",
					Status:       domain.StatusCurrent,
					DiscoveredAt: discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)
	}

	return nodes, rels, nil, nil
}
