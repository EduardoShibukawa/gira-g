package manipulator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypeScriptManipulation(t *testing.T) {
	tsCode := `import { Card } from "./types";

export interface Board {
  id: string;
  name: string;
}

export class KanbanEngine {
  private boards: Map<string, Board> = new Map();

  createBoard(name: string): Board {
    const id = "b1";
    const board = { id, name };
    this.boards.set(id, board);
    return board;
  }

  addColumn(boardId: string, name: string): void {
    console.log("adding column");
  }
}

export function helperFunction(val: string): boolean {
  return val.length > 0;
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "engine.ts")
	if err := os.WriteFile(filePath, []byte(tsCode), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// 1. FindSymbols
	symbols, err := FindSymbols(filePath)
	if err != nil {
		t.Fatalf("FindSymbols failed: %v", err)
	}

	foundNames := make(map[string]bool)
	for _, s := range symbols {
		foundNames[s.FullName] = true
	}

	expected := []string{"Board", "KanbanEngine", "KanbanEngine.createBoard", "KanbanEngine.addColumn", "helperFunction"}
	for _, exp := range expected {
		if !foundNames[exp] {
			t.Errorf("expected symbol %q not found. Found: %+v", exp, foundNames)
		}
	}

	// 2. ReadSymbol
	snippet, sym, err := ReadSymbol(filePath, "KanbanEngine.createBoard")
	if err != nil {
		t.Fatalf("ReadSymbol failed: %v", err)
	}
	if !strings.Contains(snippet, "createBoard(name: string): Board {") {
		t.Errorf("unexpected snippet content: %s", snippet)
	}
	if sym.StartLine != 11 || sym.EndLine != 16 {
		t.Errorf("expected L11-L16, got L%d-L%d", sym.StartLine, sym.EndLine)
	}

	// 3. WriteSymbol
	newMethod := `  createBoard(name: string): Board {
    const id = "fixed_id";
    const board = { id, name, columns: [] };
    this.boards.set(id, board);
    return board;
  }`
	if err := WriteSymbol(filePath, "KanbanEngine.createBoard", newMethod); err != nil {
		t.Fatalf("WriteSymbol failed: %v", err)
	}

	// Verify written file
	newSnippet, _, err := ReadSymbol(filePath, "KanbanEngine.createBoard")
	if err != nil {
		t.Fatalf("ReadSymbol after write failed: %v", err)
	}
	if !strings.Contains(newSnippet, "columns: []") {
		t.Errorf("new method body not written properly: %s", newSnippet)
	}

	// 4. AddMember
	newMember := `getBoard(id: string): Board | null {
  return this.boards.get(id) || null;
}`
	if err := AddMember(filePath, "KanbanEngine", newMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	addedSnippet, _, err := ReadSymbol(filePath, "KanbanEngine.getBoard")
	if err != nil {
		t.Fatalf("ReadSymbol for added member failed: %v", err)
	}
	if !strings.Contains(addedSnippet, "return this.boards.get(id) || null;") {
		t.Errorf("added member not found or corrupt: %s", addedSnippet)
	}
}

func TestPythonManipulation(t *testing.T) {
	pyCode := `class Calculator:
    def __init__(self):
        self.val = 0

    def add(self, a, b):
        return a + b

def standalone_func():
    return 42
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "calc.py")
	if err := os.WriteFile(filePath, []byte(pyCode), 0644); err != nil {
		t.Fatalf("failed to write temp py file: %v", err)
	}

	symbols, err := FindSymbols(filePath)
	if err != nil {
		t.Fatalf("FindSymbols py failed: %v", err)
	}

	foundNames := make(map[string]bool)
	for _, s := range symbols {
		foundNames[s.FullName] = true
	}

	if !foundNames["Calculator.add"] {
		t.Errorf("Calculator.add not found. Symbols: %+v", foundNames)
	}
	if !foundNames["standalone_func"] {
		t.Errorf("standalone_func not found. Symbols: %+v", foundNames)
	}

	// Read
	snippet, _, err := ReadSymbol(filePath, "Calculator.add")
	if err != nil {
		t.Fatalf("ReadSymbol py failed: %v", err)
	}
	if !strings.Contains(snippet, "def add(self, a, b):") {
		t.Errorf("unexpected py snippet: %s", snippet)
	}

	// Write
	newPyMethod := `    def add(self, a, b):
        result = a + b
        return result`
	if err := WriteSymbol(filePath, "Calculator.add", newPyMethod); err != nil {
		t.Fatalf("WriteSymbol py failed: %v", err)
	}

	updated, _, err := ReadSymbol(filePath, "Calculator.add")
	if err != nil {
		t.Fatalf("ReadSymbol py after write failed: %v", err)
	}
	if !strings.Contains(updated, "result = a + b") {
		t.Errorf("py method not updated: %s", updated)
	}
}

func TestSafeBraceCounter(t *testing.T) {
	code := `
		// { line comment with open brace
		/* { multiline
		   comment } */
		const str = "}";
		const char = '{';
		const template = ` + "`safe { brace } with ${ 1 + 2 }`" + `;
		const re = /[{}]/;
		function validFunc() {
			if (true) {
				return 42;
			}
		}
	`
	openB, closeB := CountSafeBraces(code)
	if openB != 2 || closeB != 2 {
		t.Fatalf("expected exactly 2 code open braces and 2 code close braces, got %d open, %d close", openB, closeB)
	}

	// Verify syntax should pass even with unbalanced braces inside comments/strings
	err := verifySyntax(code, ".ts")
	if err != nil {
		t.Fatalf("verifySyntax failed on valid code with comments/strings braces: %v", err)
	}
}

func TestTypeScriptBracesInCommentsAndStrings(t *testing.T) {
	tsCode := `export class ComplexService {
  firstMethod(): string {
    // { unclosed brace in comment
    const jsonStr = "{\"k\": 1}";
    const regex = /[{}]/;
    return "first";
  }

  secondMethod(): number {
    /* multiline
       { comment }
    */
    return 2;
  }
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "complex.ts")
	if err := os.WriteFile(filePath, []byte(tsCode), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	symbols, err := FindSymbols(filePath)
	if err != nil {
		t.Fatalf("FindSymbols failed: %v", err)
	}

	var first, second *SymbolRange
	for i := range symbols {
		if symbols[i].FullName == "ComplexService.firstMethod" {
			first = &symbols[i]
		}
		if symbols[i].FullName == "ComplexService.secondMethod" {
			second = &symbols[i]
		}
	}

	if first == nil {
		t.Fatal("ComplexService.firstMethod was not parsed as a symbol")
	}
	if second == nil {
		t.Fatal("ComplexService.secondMethod was not parsed as a symbol")
	}

	// firstMethod starts at line 2 and ends at line 7
	if first.StartLine != 2 || first.EndLine != 7 {
		t.Errorf("expected firstMethod L2-L7, got L%d-L%d", first.StartLine, first.EndLine)
	}

	// secondMethod starts at line 9 and ends at line 14
	if second.StartLine != 9 || second.EndLine != 14 {
		t.Errorf("expected secondMethod L9-L14, got L%d-L%d", second.StartLine, second.EndLine)
	}
}

