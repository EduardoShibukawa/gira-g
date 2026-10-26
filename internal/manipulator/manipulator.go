package manipulator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SymbolKind represents the architectural classification of a symbol.
type SymbolKind string

const (
	KindMethod    SymbolKind = "method"
	KindFunction  SymbolKind = "function"
	KindClass     SymbolKind = "class"
	KindInterface SymbolKind = "interface"
	KindType      SymbolKind = "type"
)

// SymbolRange represents a deterministic code location for an AST symbol.
type SymbolRange struct {
	Name       string     `json:"name"`
	Container  string     `json:"container,omitempty"`
	FullName   string     `json:"full_name"`
	Kind       SymbolKind `json:"kind"`
	StartLine  int        `json:"start_line"`
	EndLine    int        `json:"end_line"`
	Signature  string     `json:"signature,omitempty"`
	Indent     string     `json:"indent"`
}

// FindSymbols scans a file and returns all discoverable symbols and their exact line ranges.
func FindSymbols(filePath string) ([]SymbolRange, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		return parseTypeScriptSymbols(string(content))
	case ".py":
		return parsePythonSymbols(string(content))
	case ".go":
		return parseGoSymbols(string(content))
	default:
		return nil, fmt.Errorf("unsupported file extension %s for symbol manipulation", ext)
	}
}

// ReadSymbol reads the exact lines corresponding to symbolQuery in filePath.
func ReadSymbol(filePath, symbolQuery string) (string, *SymbolRange, error) {
	symbols, err := FindSymbols(filePath)
	if err != nil {
		return "", nil, err
	}

	sym := matchSymbol(symbols, symbolQuery)
	if sym == nil {
		return "", nil, fmt.Errorf("symbol %q not found in %s", symbolQuery, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, err
	}

	lines := strings.Split(string(content), "\n")
	if sym.StartLine < 1 || sym.EndLine > len(lines) || sym.StartLine > sym.EndLine {
		return "", nil, fmt.Errorf("invalid line range L%d-L%d for symbol %s in %s", sym.StartLine, sym.EndLine, sym.FullName, filePath)
	}

	selected := lines[sym.StartLine-1 : sym.EndLine]
	snippet := strings.Join(selected, "\n")

	header := fmt.Sprintf("// [%s:L%d-L%d | %s %s]\n", filepath.ToSlash(filePath), sym.StartLine, sym.EndLine, sym.Kind, sym.FullName)
	return header + snippet, sym, nil
}

// WriteSymbol replaces the exact definition of symbolQuery with newContent in filePath.
func WriteSymbol(filePath, symbolQuery, newContent string) error {
	symbols, err := FindSymbols(filePath)
	if err != nil {
		return err
	}

	sym := matchSymbol(symbols, symbolQuery)
	if sym == nil {
		return fmt.Errorf("symbol %q not found in %s", symbolQuery, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	if sym.StartLine < 1 || sym.EndLine > len(lines) || sym.StartLine > sym.EndLine {
		return fmt.Errorf("corrupt line range L%d-L%d for symbol %s", sym.StartLine, sym.EndLine, sym.FullName)
	}

	// Normaliza quebras de linha no newContent
	newLines := strings.Split(strings.ReplaceAll(newContent, "\r\n", "\n"), "\n")

	// Garante indentação base se o novo código não tiver
	indentedLines := formatIndentation(newLines, sym.Indent)

	// Substitui as linhas no slice
	var result []string
	result = append(result, lines[:sym.StartLine-1]...)
	result = append(result, indentedLines...)
	result = append(result, lines[sym.EndLine:]...)

	newFileContent := strings.Join(result, "\n")

	// Dry-run parse verification: garante que o novo arquivo ainda parseia sem quebra
	ext := strings.ToLower(filepath.Ext(filePath))
	if err := verifySyntax(newFileContent, ext); err != nil {
		return fmt.Errorf("dry-run verification failed: %w", err)
	}

	return os.WriteFile(filePath, []byte(newFileContent), 0644)
}

// AddMember inserts a new member into containerQuery in filePath.
func AddMember(filePath, containerQuery, newContent string) error {
	symbols, err := FindSymbols(filePath)
	if err != nil {
		return err
	}

	var container *SymbolRange
	for i := range symbols {
		s := &symbols[i]
		if (s.Kind == KindClass || s.Kind == KindInterface) && (s.Name == containerQuery || s.FullName == containerQuery) {
			container = s
			break
		}
	}
	if container == nil {
		return fmt.Errorf("container %q (class/interface) not found in %s", containerQuery, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	insertLine := container.EndLine - 1 // Logo antes do fechamento `}`
	if insertLine < container.StartLine {
		insertLine = container.EndLine
	}

	// Detecta indentação interna do contêiner
	memberIndent := container.Indent + "  "
	newLines := strings.Split(strings.ReplaceAll(newContent, "\r\n", "\n"), "\n")
	indentedLines := formatIndentation(newLines, memberIndent)

	var result []string
	result = append(result, lines[:insertLine]...)
	result = append(result, "") // Linha em branco de separação
	result = append(result, indentedLines...)
	result = append(result, lines[insertLine:]...)

	newFileContent := strings.Join(result, "\n")

	ext := strings.ToLower(filepath.Ext(filePath))
	if err := verifySyntax(newFileContent, ext); err != nil {
		return fmt.Errorf("dry-run verification failed: %w", err)
	}

	return os.WriteFile(filePath, []byte(newFileContent), 0644)
}

func matchSymbol(symbols []SymbolRange, query string) *SymbolRange {
	query = strings.TrimSpace(query)

	// 1. Busca exata por FullName (ex: KanbanEngine.addColumn)
	for i := range symbols {
		if symbols[i].FullName == query {
			return &symbols[i]
		}
	}

	// 2. Busca exata por Name (ex: addColumn)
	for i := range symbols {
		if symbols[i].Name == query {
			return &symbols[i]
		}
	}

	// 3. Busca case-insensitive
	for i := range symbols {
		if strings.EqualFold(symbols[i].FullName, query) || strings.EqualFold(symbols[i].Name, query) {
			return &symbols[i]
		}
	}

	return nil
}

func formatIndentation(lines []string, targetIndent string) []string {
	if len(lines) == 0 {
		return lines
	}

	// Se a primeira linha já tem indentação, respeita o delta
	firstLine := lines[0]
	firstIndent := getIndent(firstLine)

	var formatted []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			formatted = append(formatted, "")
			continue
		}
		currIndent := getIndent(l)
		if strings.HasPrefix(currIndent, firstIndent) {
			relIndent := currIndent[len(firstIndent):]
			formatted = append(formatted, targetIndent+relIndent+strings.TrimSpace(l))
		} else {
			formatted = append(formatted, targetIndent+strings.TrimSpace(l))
		}
	}
	return formatted
}

func getIndent(line string) string {
	idx := strings.IndexFunc(line, func(r rune) bool {
		return r != ' ' && r != '\t'
	})
	if idx == -1 {
		return line
	}
	return line[:idx]
}

func verifySyntax(content, ext string) error {
	switch ext {
	case ".ts", ".tsx", ".js", ".jsx":
		// Verifica contagem balanceada de chaves
		openBraces := strings.Count(content, "{")
		closeBraces := strings.Count(content, "}")
		if openBraces != closeBraces {
			return fmt.Errorf("unbalanced braces: %d '{' vs %d '}'", openBraces, closeBraces)
		}
	case ".py":
		// Valida se não há erros óbvios de colon em def/class
		scanner := bufio.NewScanner(strings.NewReader(content))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if (strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "class ")) && !strings.HasSuffix(line, ":") && !strings.Contains(line, "#") {
				if !strings.Contains(line, ")") {
					continue // assinatura multiline
				}
				return fmt.Errorf("line %d: missing ':' in python declaration %q", lineNum, line)
			}
		}
	}
	return nil
}

// Parser TypeScript / JavaScript
var (
	tsClassRe     = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:abstract\s+)?class\s+([A-Za-z0-9_$]+)`)
	tsInterfaceRe = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?interface\s+([A-Za-z0-9_$]+)`)
	tsTypeRe      = regexp.MustCompile(`^(?:export\s+)?type\s+([A-Za-z0-9_$]+)`)
	tsFuncRe      = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s+([A-Za-z0-9_$]+)`)
	tsMethodRe    = regexp.MustCompile(`^(?:\s*)(?:public\s+|private\s+|protected\s+|static\s+|async\s+)*([A-Za-z0-9_$]+)\s*(?:<[^>]+>)?\s*\(([^)]*)\)`)
)

func parseTypeScriptSymbols(source string) ([]SymbolRange, error) {
	lines := strings.Split(source, "\n")
	var symbols []SymbolRange

	var currentContainer *SymbolRange
	var containerBraceDepth int
	var currentMethod *SymbolRange
	var methodBraceDepth int

	totalBraces := 0

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)
		lineIndent := getIndent(line)

		// Rastreia chaves
		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")

		// Não processa comentários puros
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			totalBraces += (openCount - closeCount)
			continue
		}

		// Classes
		if m := tsClassRe.FindStringSubmatch(trimmed); m != nil && currentContainer == nil {
			containerBraceDepth = totalBraces
			totalBraces += (openCount - closeCount)
			sym := SymbolRange{
				Name:      m[1],
				FullName:  m[1],
				Kind:      KindClass,
				StartLine: lineNum,
				EndLine:   len(lines),
				Indent:    lineIndent,
			}
			currentContainer = &sym
			continue
		}

		// Interfaces
		if m := tsInterfaceRe.FindStringSubmatch(trimmed); m != nil && currentContainer == nil {
			containerBraceDepth = totalBraces
			totalBraces += (openCount - closeCount)
			sym := SymbolRange{
				Name:      m[1],
				FullName:  m[1],
				Kind:      KindInterface,
				StartLine: lineNum,
				EndLine:   len(lines),
				Indent:    lineIndent,
			}
			currentContainer = &sym
			continue
		}

		// Types
		if m := tsTypeRe.FindStringSubmatch(trimmed); m != nil && currentContainer == nil {
			symbols = append(symbols, SymbolRange{
				Name:      m[1],
				FullName:  m[1],
				Kind:      KindType,
				StartLine: lineNum,
				EndLine:   lineNum,
				Indent:    lineIndent,
			})
			totalBraces += (openCount - closeCount)
			continue
		}

		// Métodos dentro de Classe ou Interface
		if currentContainer != nil {
			if currentMethod == nil {
				if m := tsMethodRe.FindStringSubmatch(trimmed); m != nil {
					name := m[1]
					if name != "if" && name != "for" && name != "while" && name != "switch" && name != "catch" && name != "return" {
						methodBraceDepth = totalBraces
						methodSym := SymbolRange{
							Name:      name,
							Container: currentContainer.Name,
							FullName:  currentContainer.Name + "." + name,
							Kind:      KindMethod,
							StartLine: lineNum,
							EndLine:   lineNum,
							Indent:    lineIndent,
						}
						currentMethod = &methodSym
					}
				}
			}
		} else {
			// Funções top-level
			if currentMethod == nil {
				if m := tsFuncRe.FindStringSubmatch(trimmed); m != nil {
					methodBraceDepth = totalBraces
					fnSym := SymbolRange{
						Name:      m[1],
						FullName:  m[1],
						Kind:      KindFunction,
						StartLine: lineNum,
						EndLine:   lineNum,
						Indent:    lineIndent,
					}
					currentMethod = &fnSym
				}
			}
		}

		totalBraces += (openCount - closeCount)

		// Verifica fechamento de método/função
		if currentMethod != nil {
			if strings.Contains(line, ";") && !strings.Contains(line, "{") && totalBraces == methodBraceDepth {
				// Assinatura em interface terminada em ;
				currentMethod.EndLine = lineNum
				symbols = append(symbols, *currentMethod)
				currentMethod = nil
			} else if totalBraces <= methodBraceDepth && closeCount > 0 {
				currentMethod.EndLine = lineNum
				symbols = append(symbols, *currentMethod)
				currentMethod = nil
			}
		}

		// Verifica fechamento do contêiner (classe ou interface)
		if currentContainer != nil && totalBraces <= containerBraceDepth && closeCount > 0 {
			currentContainer.EndLine = lineNum
			symbols = append(symbols, *currentContainer)
			currentContainer = nil
		}
	}

	if currentMethod != nil {
		currentMethod.EndLine = len(lines)
		symbols = append(symbols, *currentMethod)
	}
	if currentContainer != nil {
		currentContainer.EndLine = len(lines)
		symbols = append(symbols, *currentContainer)
	}

	return symbols, nil
}

// Parser Python
var (
	pyClassRe = regexp.MustCompile(`^class\s+([A-Za-z0-9_]+)`)
	pyFuncRe  = regexp.MustCompile(`^(?:\s*)def\s+([A-Za-z0-9_]+)`)
)

func parsePythonSymbols(source string) ([]SymbolRange, error) {
	lines := strings.Split(source, "\n")
	var symbols []SymbolRange

	type pyBlock struct {
		sym        SymbolRange
		indentSize int
	}
	var stack []pyBlock

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		currIndent := len(getIndent(line))

		// Fecha blocos anteriores com indentação maior ou igual
		for len(stack) > 0 && currIndent <= stack[len(stack)-1].indentSize {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			top.sym.EndLine = lineNum - 1
			symbols = append(symbols, top.sym)
		}

		// Classes
		if m := pyClassRe.FindStringSubmatch(trimmed); m != nil && currIndent == 0 {
			sym := SymbolRange{
				Name:      m[1],
				FullName:  m[1],
				Kind:      KindClass,
				StartLine: lineNum,
				EndLine:   len(lines),
				Indent:    getIndent(line),
			}
			stack = append(stack, pyBlock{sym: sym, indentSize: currIndent})
			continue
		}

		// Funções / Métodos
		if m := pyFuncRe.FindStringSubmatch(trimmed); m != nil {
			name := m[1]
			container := ""
			fullName := name
			kind := KindFunction

			if len(stack) > 0 && stack[0].sym.Kind == KindClass {
				container = stack[0].sym.Name
				fullName = container + "." + name
				kind = KindMethod
			}

			sym := SymbolRange{
				Name:      name,
				Container: container,
				FullName:  fullName,
				Kind:      kind,
				StartLine: lineNum,
				EndLine:   len(lines),
				Indent:    getIndent(line),
			}
			stack = append(stack, pyBlock{sym: sym, indentSize: currIndent})
		}
	}

	// Descarrega o que sobrou na pilha
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		top.sym.EndLine = len(lines)
		symbols = append(symbols, top.sym)
	}

	return symbols, nil
}

// Parser Go
var (
	goFuncRe = regexp.MustCompile(`^func\s+(?:\((?:[^)]+)\)\s+)?([A-Za-z0-9_]+)`)
	goTypeRe = regexp.MustCompile(`^type\s+([A-Za-z0-9_]+)\s+(?:struct|interface)`)
)

func parseGoSymbols(source string) ([]SymbolRange, error) {
	lines := strings.Split(source, "\n")
	var symbols []SymbolRange

	var currentSym *SymbolRange
	var startBraceDepth int
	totalBraces := 0

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)
		lineIndent := getIndent(line)

		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			totalBraces += (openCount - closeCount)
			continue
		}

		if currentSym == nil {
			if m := goTypeRe.FindStringSubmatch(trimmed); m != nil {
				startBraceDepth = totalBraces
				kind := KindClass
				if strings.Contains(trimmed, "interface") {
					kind = KindInterface
				}
				currentSym = &SymbolRange{
					Name:      m[1],
					FullName:  m[1],
					Kind:      kind,
					StartLine: lineNum,
					EndLine:   len(lines),
					Indent:    lineIndent,
				}
			} else if m := goFuncRe.FindStringSubmatch(trimmed); m != nil {
				startBraceDepth = totalBraces
				currentSym = &SymbolRange{
					Name:      m[1],
					FullName:  m[1],
					Kind:      KindFunction,
					StartLine: lineNum,
					EndLine:   len(lines),
					Indent:    lineIndent,
				}
			}
		}

		totalBraces += (openCount - closeCount)

		if currentSym != nil && totalBraces <= startBraceDepth && closeCount > 0 {
			currentSym.EndLine = lineNum
			symbols = append(symbols, *currentSym)
			currentSym = nil
		}
	}

	if currentSym != nil {
		currentSym.EndLine = len(lines)
		symbols = append(symbols, *currentSym)
	}

	return symbols, nil
}
