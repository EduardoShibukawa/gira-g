package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/ask"
)

func runAsk(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("ask", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var indexPath string
	var showDiagram bool
	var asJSON bool
	var format string

	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json index")
	fs.BoolVar(&showDiagram, "diagram", false, "Emit Mermaid architecture diagram for the query")
	fs.BoolVar(&asJSON, "json", false, "Output answer as structured JSON")
	fs.StringVar(&format, "format", "", "Output format: human, md, caveman, json")

	positionals, flagArgs := splitArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPos := append(positionals, fs.Args()...)
	query := strings.TrimSpace(strings.Join(allPos, " "))
	if query == "" {
		fmt.Fprintf(stdout, `💡 Gira-G Plain Language Query (gira ask)

Usage:
  gira ask "<pergunta ou dúvida em linguagem natural>" [options]

Options:
  --diagram       Include a Mermaid architecture flow diagram in the answer
  --format <fmt>  Output format: human, md, caveman (token-efficient for agents), json
  --json          Output answer as structured JSON (shorthand for --format json)
  --index <p>     Path to Macro Map index

Exemplos de Perguntas:
  gira ask "Qual é o objetivo principal deste projeto?"
  gira ask "Como eu inicializo e utilizo o sistema?"
  gira ask "Mostre o fluxo de arquitetura e componentes" --diagram
  gira ask "O que quebra se eu alterar billing?" --format caveman
`)
		return nil
	}

	mm, _ := loadIndex(indexPath, io.Discard)

	resp := ask.ProcessQuery(query, mm, showDiagram)

	fmtChoice := strings.ToLower(strings.TrimSpace(format))
	if asJSON || fmtChoice == "json" {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	}

	if fmtChoice == "caveman" || fmtChoice == "llm" {
		fmt.Fprintln(stdout, resp.RenderCaveman())
		return nil
	}

	fmt.Fprintln(stdout, resp.Answer)
	if resp.Diagram != "" {
		fmt.Fprintln(stdout, "\n```mermaid")
		fmt.Fprintln(stdout, resp.Diagram)
		fmt.Fprintln(stdout, "```")
	}

	return nil
}
