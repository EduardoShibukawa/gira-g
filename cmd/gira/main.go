package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"labs/gira-g/internal/domain"
	"labs/gira-g/internal/mcp"
	"labs/gira-g/internal/slices"
	"labs/gira-g/internal/store"
)

var (
	version = "0.1.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run executes the command with provided args and streams.
func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "gira %s (commit: %s, date: %s)\n", version, commit, date)
		return nil

	case "help", "--help", "-h":
		printUsage(stdout)
		return nil

	case "serve":
		return runServe(rest, stdin, stdout, stderr)

	case "impact":
		return runImpact(rest, stdout, stderr)

	case "slice":
		return runSlice(rest, stdout, stderr)

	case "explain":
		return runExplain(rest, stdout, stderr)

	case "locate":
		return runLocate(rest, stdout, stderr)

	default:
		return fmt.Errorf("unknown command %q. Run 'gira --help' for usage", cmd)
	}
}

func printUsage(out io.Writer) {
	fmt.Fprintf(out, `Gira-G Architecture Harness (%s)

Usage:
  gira <command> [options] [arguments]

Commands:
  serve mcp [--index <path>]                                Start stdio MCP protocol adapter
  impact <target_id> [--depth <int>] [--index <path>]       Compute impact briefing for target component
  slice <node_id> [--tier <tier>] [--index <path>]          Extract bounded context slice (tier: summary|contracts|deep)
  explain [--node <id>] [--edge <id>] [--index <path>]      Explain discrete evidence records
  locate <query> [--index <path>]                           Search boundaries, methods, and unhandled files
  version                                                   Show harness version

Global Flags:
  --format json|md    Output format for query commands (default: json)
  --index <path>      Path to canonical macro-map.json index
`, version)
}

// splitArgs separates leading positional arguments from flag arguments.
// For example: ["service:order", "--index", "path"] -> (["service:order"], ["--index", "path"])
func splitArgs(args []string) ([]string, []string) {
	var positionals []string
	var flags []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, args[i:]...)
			break
		} else {
			positionals = append(positionals, arg)
		}
	}
	return positionals, flags
}

func loadIndex(indexPath string) (*domain.MacroMap, error) {
	candidates := make([]string, 0, 4)
	if strings.TrimSpace(indexPath) != "" {
		candidates = append(candidates, indexPath)
	} else {
		// Look in current working directory .gira/macro-map.json
		candidates = append(candidates, filepath.Join(".gira", "macro-map.json"))
		// Look in current working directory macro-map.json
		candidates = append(candidates, "macro-map.json")
		// Look in user cache directory default index
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			candidates = append(candidates, filepath.Join(home, ".cache", "gira", "indexes", "default", "macro-map.json"))
		}
	}

	s := store.NewStore("")
	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return s.LoadFile(cand)
		}
	}

	// If a specific index path was explicitly requested, return error
	if strings.TrimSpace(indexPath) != "" {
		return nil, fmt.Errorf("index file not found: %s", indexPath)
	}

	// Default fallback: return empty MacroMap
	return domain.NewMacroMap("default"), nil
}

func runServe(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("missing serve subcommand. Usage: gira serve mcp [--index <path>]")
	}

	sub := args[0]
	if sub != "mcp" {
		return fmt.Errorf("unknown serve target %q. Only 'mcp' is supported", sub)
	}

	fs := flag.NewFlagSet("serve mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var indexPath string
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	server := mcp.NewServer("gira", version, engine)
	return server.Serve(stdin, stdout)
}

func runImpact(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("impact", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var depth int
	var indexPath string
	var format string

	fs.IntVar(&depth, "depth", 1, "Graph traversal depth")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) == 0 {
		return fmt.Errorf("missing target_id argument. Usage: gira impact <target_id> [--depth <int>] [--format json|md]")
	}
	targetID := allPositionals[0]

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	briefing, err := engine.GetImpactBriefing(targetID, depth)
	if err != nil {
		return err
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		fmt.Fprintln(stdout, briefing.RenderMarkdown())
	default:
		data, err := briefing.CanonicalJSON()
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, string(data))
	}

	return nil
}

func runSlice(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("slice", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var tier string
	var indexPath string
	var format string

	fs.StringVar(&tier, "tier", "summary", "Context slice tier (summary|contracts|deep)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) == 0 {
		return fmt.Errorf("missing node_id argument. Usage: gira slice <node_id> [--tier summary|contracts|deep] [--format json|md]")
	}
	nodeID := allPositionals[0]

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	slice, err := engine.GetContextSlice(nodeID, slices.SliceTier(tier))
	if err != nil {
		return err
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		fmt.Fprintln(stdout, slice.RenderMarkdown())
	default:
		data, err := slice.CanonicalJSON()
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, string(data))
	}

	return nil
}

func runExplain(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var nodeID string
	var edgeID string
	var indexPath string
	var format string

	fs.StringVar(&nodeID, "node", "", "Node ID")
	fs.StringVar(&edgeID, "edge", "", "Edge ID")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if nodeID == "" && edgeID == "" && len(allPositionals) > 0 {
		nodeID = allPositionals[0]
	}

	if nodeID == "" && edgeID == "" {
		return fmt.Errorf("either --node or --edge must be provided")
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	report, err := engine.ExplainEvidence(nodeID, edgeID)
	if err != nil {
		return err
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		fmt.Fprintln(stdout, report.RenderMarkdown())
	default:
		data, err := report.CanonicalJSON()
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, string(data))
	}

	return nil
}

func runLocate(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("locate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var indexPath string
	var format string

	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) == 0 {
		return fmt.Errorf("missing query argument. Usage: gira locate <query> [--format json|md]")
	}
	query := allPositionals[0]

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	matches, err := engine.LocateBoundary(query)
	if err != nil {
		return err
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		fmt.Fprintln(stdout, slices.RenderBoundaryMatches(matches))
	default:
		data, err := json.MarshalIndent(matches, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		fmt.Fprint(stdout, string(data))
	}

	return nil
}
