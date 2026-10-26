package architecture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/manipulator"
)

// QueryDoc inspects a query string and returns semantic documentation for a module or symbol.
func QueryDoc(repoRoot string, query string) (*ModuleDoc, *SymbolDoc, error) {
	pkgs, err := ScanRepository(repoRoot)
	if err != nil {
		return nil, nil, err
	}

	cleanQuery := strings.TrimSpace(query)
	cleanQuery = strings.TrimPrefix(cleanQuery, "./")

	// 1. Check if query matches a package/module
	for _, p := range pkgs {
		if strings.EqualFold(p.Path, cleanQuery) ||
			strings.EqualFold(filepath.Base(p.Path), cleanQuery) ||
			strings.EqualFold("internal/"+cleanQuery, p.Path) {

			var coreAbs []ClusteredNode
			var typeNames []string
			var funcNames []string

			for _, t := range p.Types {
				typeNames = append(typeNames, fmt.Sprintf("%s (%s)", t.Name, t.Kind))
				coreAbs = append(coreAbs, ClusteredNode{
					ID:          p.Path + ":" + t.Name,
					Name:        t.Name,
					Kind:        domain.NodeContract,
					Domain:      p.Domain,
					Role:        RoleCoreAbstraction,
					Description: t.Doc,
					FileLocator: t.Locator,
					Methods:     funcNamesFromMethods(t.Methods),
				})
			}

			for _, f := range p.Functions {
				if f.Exported {
					funcNames = append(funcNames, f.Signature)
				}
			}

			doc := &ModuleDoc{
				Package:          p.Name,
				Path:             p.Path,
				Domain:           p.Domain,
				Description:      p.Doc,
				CoreAbstractions: coreAbs,
				Types:            typeNames,
				Functions:        funcNames,
				Imports:          p.Imports,
			}
			return doc, nil, nil
		}
	}

	// 2. Check if query matches an exported symbol
	for _, p := range pkgs {
		for _, t := range p.Types {
			if strings.EqualFold(t.Name, cleanQuery) {
				return nil, &SymbolDoc{
					Symbol:     t.Name,
					Kind:       t.Kind,
					Signature:  t.Signature,
					DocComment: t.Doc,
					Package:    p.Path,
					Domain:     p.Domain,
					Locator:    t.Locator,
				}, nil
			}

			for _, m := range t.Methods {
				if strings.EqualFold(m.Name, cleanQuery) || strings.EqualFold(t.Name+"."+m.Name, cleanQuery) {
					return nil, &SymbolDoc{
						Symbol:     m.Name,
						Container:  t.Name,
						Kind:       "method",
						Signature:  m.Signature,
						DocComment: m.Doc,
						Package:    p.Path,
						Domain:     p.Domain,
						Locator:    m.Locator,
					}, nil
				}
			}
		}

		for _, f := range p.Functions {
			if strings.EqualFold(f.Name, cleanQuery) {
				return nil, &SymbolDoc{
					Symbol:     f.Name,
					Kind:       "function",
					Signature:  f.Signature,
					DocComment: f.Doc,
					Package:    p.Path,
					Domain:     p.Domain,
					Locator:    f.Locator,
				}, nil
			}
		}
	}

	// 3. Fallback: try manipulator symbol lookup on any Go file
	for _, p := range pkgs {
		for _, f := range p.Files {
			absPath := filepath.Join(repoRoot, f)
			snippet, symRange, err := manipulator.ReadSymbol(absPath, cleanQuery)
			if err == nil && symRange != nil {
				return nil, &SymbolDoc{
					Symbol:    symRange.FullName,
					Container: symRange.Container,
					Kind:      string(symRange.Kind),
					Signature: symRange.Signature,
					Package:   p.Path,
					Domain:    p.Domain,
					Locator: domain.Locator{
						Repo:      "local",
						CommitSHA: "head",
						Path:      f,
						StartLine: symRange.StartLine,
						EndLine:   symRange.EndLine,
					},
					Code: snippet,
				}, nil
			}
		}
	}

	return nil, nil, nil
}

func funcNamesFromMethods(methods []FuncInfo) []string {
	var out []string
	for _, m := range methods {
		out = append(out, m.Name)
	}
	return out
}

// RenderModuleDocMarkdown formats a ModuleDoc as human-friendly markdown.
func RenderModuleDocMarkdown(repoRoot string, mod *ModuleDoc) string {
	var sb strings.Builder
	pkgLink := formatFileLink(repoRoot, mod.Path)
	sb.WriteString(fmt.Sprintf("# Module Document: [`%s`](%s)\n\n", mod.Path, pkgLink))
	sb.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", mod.Domain))
	sb.WriteString(fmt.Sprintf("- **Package**: `%s`\n", mod.Package))
	if mod.Description != "" {
		sb.WriteString(fmt.Sprintf("- **Description**: %s\n", mod.Description))
	}
	sb.WriteString("\n")

	if len(mod.Imports) > 0 {
		sb.WriteString("## Internal Dependencies\n\n")
		for _, imp := range mod.Imports {
			sb.WriteString(fmt.Sprintf("- [`%s`](%s)\n", imp, formatFileLink(repoRoot, imp)))
		}
		sb.WriteString("\n")
	}

	if len(mod.Types) > 0 {
		sb.WriteString(fmt.Sprintf("## Exported Types & Contracts (%d)\n\n", len(mod.Types)))
		for _, t := range mod.Types {
			sb.WriteString(fmt.Sprintf("- `%s`\n", t))
		}
		sb.WriteString("\n")
	}

	if len(mod.Functions) > 0 {
		sb.WriteString(fmt.Sprintf("## Exported Functions (%d)\n\n", len(mod.Functions)))
		sb.WriteString("```go\n")
		for _, f := range mod.Functions {
			sb.WriteString(fmt.Sprintf("%s\n", f))
		}
		sb.WriteString("```\n")
	}

	return sb.String()
}

// RenderSymbolDocMarkdown formats a SymbolDoc as human-friendly markdown.
func RenderSymbolDocMarkdown(repoRoot string, sym *SymbolDoc) string {
	var sb strings.Builder
	cit := formatCitationLink(repoRoot, sym.Locator)
	sb.WriteString(fmt.Sprintf("# Symbol Document: `%s`\n\n", sym.Symbol))
	sb.WriteString(fmt.Sprintf("- **Kind**: `%s`\n", sym.Kind))
	if sym.Container != "" {
		sb.WriteString(fmt.Sprintf("- **Container**: `%s`\n", sym.Container))
	}
	sb.WriteString(fmt.Sprintf("- **Package**: [`%s`](%s)\n", sym.Package, formatFileLink(repoRoot, sym.Package)))
	sb.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", sym.Domain))
	sb.WriteString(fmt.Sprintf("- **Locator**: [%s](%s)\n\n", sym.Locator.String(), cit))

	if sym.DocComment != "" {
		sb.WriteString(fmt.Sprintf("### Documentation\n\n%s\n\n", sym.DocComment))
	}

	if sym.Signature != "" {
		sb.WriteString("### Declaration Signature\n\n```go\n")
		sb.WriteString(sym.Signature + "\n")
		sb.WriteString("```\n\n")
	}

	if sym.Code != "" {
		sb.WriteString("### Implementation Snippet\n\n```go\n")
		sb.WriteString(sym.Code + "\n")
		sb.WriteString("```\n")
	}

	return sb.String()
}

// QueryFlows finds the architectural flow corresponding to an intent or entrypoint.
func QueryFlows(entryOrIntent string) *LifecycleFlow {
	clean := strings.ToLower(strings.TrimSpace(entryOrIntent))
	lifecycles := DefaultLifecycles()

	for _, lc := range lifecycles {
		if strings.EqualFold(lc.Intent, clean) || strings.Contains(strings.ToLower(lc.Title), clean) {
			return &lc
		}
	}

	// Keyword intent matching
	switch {
	case strings.Contains(clean, "ast") || strings.Contains(clean, "read") || strings.Contains(clean, "write") || strings.Contains(clean, "add") || strings.Contains(clean, "manipulat"):
		return &lifecycles[0]
	case strings.Contains(clean, "impact") || strings.Contains(clean, "slice") || strings.Contains(clean, "blast"):
		return &lifecycles[1]
	case strings.Contains(clean, "index") || strings.Contains(clean, "init") || strings.Contains(clean, "extract"):
		return &lifecycles[2]
	case strings.Contains(clean, "gate") || strings.Contains(clean, "eval") || strings.Contains(clean, "lint") || strings.Contains(clean, "scorecard"):
		return &lifecycles[3]
	}

	return nil
}

// RenderFlowMarkdown formats a LifecycleFlow as human-friendly markdown.
func RenderFlowMarkdown(repoRoot string, flow *LifecycleFlow) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Architecture Flow: %s (`intent: %s`)\n\n", flow.Title, flow.Intent))
	sb.WriteString(fmt.Sprintf("%s\n\n", flow.Description))

	if flow.MermaidSeq != "" {
		sb.WriteString("## Sequence Diagram\n\n```mermaid\n")
		sb.WriteString(flow.MermaidSeq)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("## Step-by-Step Path\n\n")
	sb.WriteString("| Order | Domain | Component | Action | Code Citation |\n")
	sb.WriteString("| :---: | :--- | :--- | :--- | :--- |\n")
	for _, s := range flow.Steps {
		cit := s.CodeLocator
		if !strings.HasPrefix(cit, "file://") && s.CodeLocator != "" {
			cit = "file://" + filepath.ToSlash(filepath.Join(repoRoot, s.CodeLocator))
		}
		sb.WriteString(fmt.Sprintf("| %d | `%s` | `%s` | %s | [%s](%s) |\n",
			s.Order, s.Domain, s.Component, s.Action, s.Component, cit))
	}

	return sb.String()
}

// MarshalCompactJSON serializes data compactly for AI agents.
func MarshalCompactJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
