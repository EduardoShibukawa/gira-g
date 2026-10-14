package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/archdiff"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/gitdiff"
	"github.com/EduardoShibukawa/gira-g/internal/human"
	"github.com/EduardoShibukawa/gira-g/internal/indexer"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
	"github.com/EduardoShibukawa/gira-g/internal/mcp"
	"github.com/EduardoShibukawa/gira-g/internal/skills"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// emitHuman writes a rendered human document. Markdown is styled only on a
// terminal with NO_COLOR unset; JSON and the plain path stay byte-identical
// (ADR 0008 §1).
func emitHuman(stdout io.Writer, data []byte, format slices.Format) error {
	text := string(data)
	if format == slices.FormatMarkdown {
		text = human.Decorate(text, human.Styled(stdout))
	}
	_, err := io.WriteString(stdout, text)
	return err
}

// emitMarkdownLine writes text plus a trailing newline, decorating only when
// stdout is a styled terminal.
func emitMarkdownLine(stdout io.Writer, text string) error {
	_, err := io.WriteString(stdout, human.Decorate(text, human.Styled(stdout))+"\n")
	return err
}

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

	case "index":
		return runIndex(rest, stdout, stderr)

	case "serve":
		return runServe(rest, stdin, stdout, stderr)

	case "impact":
		return runImpact(rest, stdout, stderr)

	case "slice":
		return runSlice(rest, stdout, stderr)

	case "explain":
		return runExplain(rest, stdout, stderr)

	case "tour":
		return runTour(rest, stdout, stderr)

	case "diff":
		return runDiff(rest, stdout, stderr)

	case "skills":
		return runSkills(rest, stdin, stdout, stderr)

	case "flows":
		return runFlows(rest, stdout, stderr)

	case "locate":
		return runLocate(rest, stdout, stderr)

	case "export":
		return runExport(rest, stdout, stderr)

	case "view":
		return runView(rest, stdout, stderr)

	case "eval":
		return runEval(rest, stdout, stderr)

	case "verify":
		return runVerify(rest, stdout, stderr)

	case "review":
		return runReview(rest, stdout, stderr)

	case "gate-lint":
		return runGateLint(rest, stdout, stderr)

	case "gate":
		return runGate(rest, stdout, stderr)

	default:
		return fmt.Errorf("unknown command %q. Run 'gira --help' for usage", cmd)
	}
}

func printUsage(out io.Writer) {
	fmt.Fprintf(out, `Gira-G Architecture Harness (%s)

Usage:
  gira <command> [options] [arguments]

Commands:
  index [manifest] [--output <path>] [--offline] [--cache-dir <dir>] Build Macro Map index from manifest
  serve mcp [--index <path>]                                Start stdio MCP protocol adapter
  impact <target_id> [--depth <int>] [--budget <tokens>] [--index <path>] Compute impact briefing for target component
  impact --from-diff <base>..<head> [--repo <dir>] [--repo-id <id>] Impact briefing of the nodes a git diff touches
  slice <node_id> [--tier <tier>] [--view <view>] [--budget <tokens>] [--index <path>] Extract bounded context slice (tier: summary|contracts|deep, view: full|architecture|governance)
  explain [--node <id>] [--edge <id>] [--index <path>]      Explain discrete evidence records
  tour [--focus <node>] [--max-steps <n>] [--index <path>]  Deterministic guided reading tour from entry points or a focus node
  diff <index-a> <index-b> [--format json|md]               Architecture diff: renames, nodes, edges, contracts, unsupported paths
  skills install --host claude|cursor|agents [--dir <project>] [--index <path>] Generate GIRA skills and MCP registration for a host
  flows <entry> [--index <path>]                            Deterministic chains from an entry through consumes edges to infrastructure
  locate <query> [--index <path>]                           Search boundaries, methods, and unhandled files
  export [--format md|html] [--output <path>] [--index <p>] Export standalone architecture document
  view [--output <dir>] [--index <path>]                    Export HTML and view in browser
  eval [--artifact-verified] [--no-artifact-verified]       Eval stub ending with ARTIFACT-VERIFIED closing line (F1)
  verify --check <c> --expect <e> [--cwd <d>]               Print ARTIFACT-VERIFIED closing line for inputs (F1)
  review --findings-register <path> [--fix-check]           L3 findings register + independent fix-check (F3)
  gate-lint --strict [--state <path> --digest <hex>]        Strict gate lint: warn-only in CI, fail-closed locally (F5)
  gate --status | --reverify [--state <p> --digest <h>]     Non-executing status vs re-verification split (F5)
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
	candidates := make([]string, 0, 8)
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
			sHome := store.NewStore(filepath.Join(home, ".cache", "gira", "indexes"))
			if systems, err := sHome.List(); err == nil {
				for _, sys := range systems {
					candidates = append(candidates, filepath.Join(home, ".cache", "gira", "indexes", sys, "macro-map.json"))
				}
			}
		}
		sLocal := store.NewStore(filepath.Join(".gira", "indexes"))
		if systems, err := sLocal.List(); err == nil {
			for _, sys := range systems {
				candidates = append(candidates, filepath.Join(".gira", "indexes", sys, "macro-map.json"))
			}
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

func runIndex(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outputPath string
	var offline bool
	var cacheDir string
	var manifestFlag string

	fs.StringVar(&outputPath, "output", "", "Path to write output macro-map.json")
	fs.BoolVar(&offline, "offline", false, "Disable network egress and use local cache exclusively")
	fs.StringVar(&cacheDir, "cache-dir", "", "Local repository cache directory")
	fs.StringVar(&manifestFlag, "manifest", "", "Path to gira.yaml manifest")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	manifestPath := "gira.yaml"
	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) > 0 && strings.TrimSpace(allPositionals[0]) != "" {
		manifestPath = allPositionals[0]
	} else if strings.TrimSpace(manifestFlag) != "" {
		manifestPath = manifestFlag
	}

	// 1. Read manifest
	m, err := manifest.ParseFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest %s: %w", manifestPath, err)
	}

	// 2. Initialize CacheManager
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// 3–5. Fetch repos, extract, apply overrides (shared with the evaluator)
	mm, err := indexer.Build(context.Background(), m, cm, indexer.Options{Offline: offline, ManifestPath: manifestPath})
	if err != nil {
		return err
	}

	// 6. Persist generated MacroMap
	if outputPath != "" {
		data, err := mm.CanonicalJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize macro map: %w", err)
		}
		if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(stdout, "Indexed %d nodes and %d relationships for target system %q (saved to %s)\n",
			len(mm.Nodes), len(mm.Relationships), m.TargetSystem.Name, outputPath)
	} else {
		storeDir := filepath.Join(".gira", "indexes")
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			storeDir = filepath.Join(home, ".cache", "gira", "indexes")
		}
		s := store.NewStore(storeDir)
		if err := s.Save(m.TargetSystem.Name, mm); err != nil {
			return fmt.Errorf("failed to persist macro map: %w", err)
		}
		dest := filepath.Join(storeDir, m.TargetSystem.Name, "macro-map.json")
		fmt.Fprintf(stdout, "Indexed %d nodes and %d relationships for target system %q (saved to %s)\n",
			len(mm.Nodes), len(mm.Relationships), m.TargetSystem.Name, dest)
	}

	return nil
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
	var budget int

	var fromDiff, repoDir, repoID string

	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget; briefing default 1500)")
	fs.IntVar(&depth, "depth", 1, "Graph traversal depth")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&fromDiff, "from-diff", "", "Start from a git revision range <base>..<head> instead of a target id")
	fs.StringVar(&repoDir, "repo", ".", "Repository the --from-diff range is read from")
	fs.StringVar(&repoID, "repo-id", "", "Only match locators of this repository id (--from-diff)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if fromDiff == "" && len(allPositionals) == 0 {
		return fmt.Errorf("missing target_id argument. Usage: gira impact <target_id> [--depth <int>] [--format json|md] or gira impact --from-diff <base>..<head>")
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	var briefing *slices.ImpactBriefing
	if fromDiff != "" {
		diff, err := gitdiff.Run(context.Background(), repoDir, fromDiff)
		if err != nil {
			return err
		}
		changes, err := gitdiff.Parse(diff)
		if err != nil {
			return err
		}
		briefing, err = engine.GetPRImpact(fromDiff, repoID, changes, depth)
		if err != nil {
			return err
		}
	} else {
		briefing, err = engine.GetImpactBriefing(allPositionals[0], depth)
		if err != nil {
			return err
		}
	}

	emitFormat := slices.ParseFormat(strings.ToLower(format))
	data, err := briefing.Emit(emitFormat, budget)
	if err != nil {
		return err
	}
	return emitHuman(stdout, data, emitFormat)
}

func runSlice(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("slice", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var tier string
	var view string
	var indexPath string
	var format string
	var budget int

	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget; tier defaults 500/1500/3500)")

	fs.StringVar(&tier, "tier", "summary", "Context slice tier (summary|contracts|deep)")
	fs.StringVar(&view, "view", "full", "Context slice architectural view (full|architecture|governance)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) == 0 {
		return fmt.Errorf("missing node_id argument. Usage: gira slice <node_id> [--tier summary|contracts|deep] [--view full|architecture|governance] [--format json|md]")
	}
	nodeID := allPositionals[0]

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	slice, err := engine.GetContextSliceWithView(nodeID, slices.SliceTier(tier), slices.GraphView(view))
	if err != nil {
		return err
	}

	emitFormat := slices.ParseFormat(strings.ToLower(format))
	data, err := slice.Emit(emitFormat, budget)
	if err != nil {
		return err
	}
	return emitHuman(stdout, data, emitFormat)
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
		if err := emitMarkdownLine(stdout, report.RenderMarkdown()); err != nil {
			return err
		}
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
		if err := emitMarkdownLine(stdout, slices.RenderBoundaryMatches(matches)); err != nil {
			return err
		}
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

func runExport(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var format string
	var outputPath string
	var indexPath string

	fs.StringVar(&format, "format", "md", "Export format (md|html)")
	fs.StringVar(&outputPath, "output", "", "Output file path (default stdout)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")

	if err := fs.Parse(args); err != nil {
		return err
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	var content string
	switch strings.ToLower(format) {
	case "md", "markdown":
		content = slices.ExportMarkdown(mm)
	case "html":
		content = slices.ExportHTML(mm)
	default:
		return fmt.Errorf("unsupported export format %q. Supported formats: md, html", format)
	}

	if outputPath != "" {
		if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write export file: %w", err)
		}
		return nil
	}

	fmt.Fprint(stdout, content)
	return nil
}

func runView(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outputDir string
	var indexPath string

	fs.StringVar(&outputDir, "output", "", "Directory to generate HTML artifact (default temp dir)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")

	if err := fs.Parse(args); err != nil {
		return err
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	htmlContent := slices.ExportHTML(mm)

	var targetDir string
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		targetDir = outputDir
	} else {
		tmpDir, err := os.MkdirTemp("", "gira-view-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		targetDir = tmpDir
	}

	artifactPath := filepath.Join(targetDir, "index.html")
	if err := os.WriteFile(artifactPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML view artifact: %w", err)
	}

	absPath, err := filepath.Abs(artifactPath)
	if err != nil {
		absPath = artifactPath
	}

	fileURI := "file://" + filepath.ToSlash(absPath)
	fmt.Fprintf(stdout, "Exported HTML view to: %s\n", fileURI)

	// Attempt non-blocking browser open if not in headless/CI environment
	openInBrowser(absPath)

	return nil
}

func openInBrowser(filePath string) {
	// If CI or NO_BROWSER environment variable is set, avoid launching external process
	if os.Getenv("CI") != "" || os.Getenv("NO_BROWSER") != "" {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Only try xdg-open if DISPLAY or WAYLAND_DISPLAY is set (interactive desktop)
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return
		}
		cmd = exec.Command("xdg-open", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath)
	default:
		return
	}

	// Non-blocking fire-and-forget; do not block or fail command if browser fails
	_ = cmd.Start()
}

func runTour(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("tour", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var focus, indexPath, format string
	var maxSteps int
	fs.StringVar(&focus, "focus", "", "Start the tour at this node instead of the entry points")
	fs.IntVar(&maxSteps, "max-steps", slices.DefaultTourSteps, "Maximum number of tour steps")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}
	tour, err := slices.NewEngine(mm).GetTour(focus, maxSteps)
	if err != nil {
		return err
	}
	if slices.ParseFormat(strings.ToLower(format)) == slices.FormatMarkdown {
		return emitMarkdownLine(stdout, tour.RenderMarkdown())
	}
	data, err := tour.CanonicalJSON()
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func runFlows(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)
	fs := flag.NewFlagSet("flows", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var indexPath, format string
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	all := append(positionals, fs.Args()...)
	if len(all) == 0 {
		return fmt.Errorf("missing entry argument. Usage: gira flows <entry> [--format json|md]")
	}
	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}
	flows, err := slices.NewEngine(mm).GetFlows(all[0])
	if err != nil {
		return err
	}
	if slices.ParseFormat(strings.ToLower(format)) == slices.FormatMarkdown {
		return emitMarkdownLine(stdout, flows.RenderMarkdown())
	}
	data, err := flows.CanonicalJSON()
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func runDiff(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var format string
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	all := append(positionals, fs.Args()...)
	if len(all) != 2 {
		return fmt.Errorf("want two index paths. Usage: gira diff <index-a> <index-b> [--format json|md]")
	}
	s := store.NewStore("")
	a, err := s.LoadFile(all[0])
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", all[0], err)
	}
	b, err := s.LoadFile(all[1])
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", all[1], err)
	}
	d := archdiff.Compare(a, b)
	if slices.ParseFormat(strings.ToLower(format)) == slices.FormatMarkdown {
		return emitMarkdownLine(stdout, d.RenderMarkdown())
	}
	data, err := d.CanonicalJSON()
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func runSkills(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "nudge" {
		return runNudge(args[1:], stdin, stdout, stderr)
	}
	if len(args) == 0 || args[0] != "install" {
		return fmt.Errorf("usage: gira skills install --host %s [--dir <project>] [--index <path>] [--hook]", strings.Join(skills.Hosts, "|"))
	}
	fs := flag.NewFlagSet("skills install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var host, dir, index, command string
	fs.StringVar(&host, "host", "", "Host family: "+strings.Join(skills.Hosts, ", "))
	fs.StringVar(&dir, "dir", ".", "Project directory to install into")
	fs.StringVar(&index, "index", "", "Index the registered MCP server should load")
	fs.StringVar(&command, "command", "gira", "Command the host runs to start the MCP server")
	var hook bool
	fs.BoolVar(&hook, "hook", false, "claude host: register the opt-in PreToolUse nudge (never blocks)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	rep, err := skills.Install(host, skills.Options{Root: dir, Index: index, Command: command, Hook: hook})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(data))
	return err
}

// runNudge is the PreToolUse hook command. It never fails the hook: problems
// produce no output and exit status 0.
func runNudge(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("skills nudge", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var indexPath string
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	if err := fs.Parse(args); err != nil {
		return nil
	}
	mm, err := loadIndex(indexPath)
	if err != nil {
		return nil
	}
	root, _ := os.Getwd()
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		root = strings.TrimSpace(string(out))
	}
	_ = skills.Nudge(stdin, stdout, mm, root)
	return nil
}
