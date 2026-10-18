package main

import (
	"bytes"
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
	"gopkg.in/yaml.v3"
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

// humanBanner writes the indexed-commit line and, when checkout is given and its
// HEAD is not an indexed commit, a stale line — to stderr, so the stdout
// document (and its E3 goldens/budget markers) stays byte-identical (ADR 0008
// §4). It only fires for human (Markdown and HTML) output, so JSON stays clean.
func humanBanner(stderr io.Writer, mm *domain.MacroMap, checkout string, format slices.Format) {
	if mm == nil || (format != slices.FormatMarkdown && format != slices.FormatHTML) {
		return
	}
	if cs := mm.IndexedCommits(); len(cs) > 0 {
		fmt.Fprintf(stderr, "indexed commit: %s\n", strings.Join(cs, ", "))
	}
	if checkout != "" {
		if head := gitHead(checkout); head != "" && !mm.HasCommit(head) {
			fmt.Fprintf(stderr, "stale: local HEAD %s is not an indexed commit\n", head)
		}
	}
}

func gitHead(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	cmd.Env = gitCommandEnv()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// excerptFor reads a locator's excerpt from the checkout. When the checkout is a
// Git repository it must contain the cited commit, so an unrelated checkout
// cannot masquerade as the source (F2).
func excerptFor(repoRoot string, loc domain.Locator) (string, bool) {
	if repoRoot == "" || loc.Path == "" {
		return "", false
	}
	if loc.CommitSHA != "" && isGitRepo(repoRoot) && !checkoutHasCommit(repoRoot, loc.CommitSHA) {
		return "", false
	}
	return slices.Excerpt(repoRoot, loc, slices.SourceWindow)
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	cmd.Env = gitCommandEnv()
	return cmd.Run() == nil
}

func checkoutHasCommit(dir, sha string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "--quiet", sha+"^{commit}")
	cmd.Env = gitCommandEnv()
	return cmd.Run() == nil
}

// sourceSection renders the on-demand source excerpts for an evidence report,
// one block per distinct locator. A locator whose source cannot be read shows
// the "source unavailable" marker. Reading is at render time only (ADR 0001).
func sourceSection(r *slices.EvidenceReport, repoRoot string) string {
	var sb strings.Builder
	seen := map[string]bool{}
	blocks := 0
	for _, ev := range r.Evidence {
		loc := ev.Locator
		if loc.Path == "" {
			continue
		}
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		if blocks == 0 {
			sb.WriteString("\n## Source Excerpts\n")
		}
		blocks++
		fmt.Fprintf(&sb, "\n### %s\n\n```\n", key)
		if ex, ok := excerptFor(repoRoot, loc); ok {
			sb.WriteString(ex)
		} else {
			sb.WriteString(slices.SourceUnavailable + "\n")
		}
		sb.WriteString("```\n")
	}
	return sb.String()
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

	case "init":
		return runInit(rest, stdout, stderr)

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

	case "doc":
		return runDoc(rest, stdout, stderr)

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
  init [dir] [--force] [--output <path>] [--offline] [--cache-dir <dir>] Generate gira.yaml from a local repo (HEAD pinned) and index it
  serve mcp [--index <path>]                                Start stdio MCP protocol adapter
  impact <target_id> [--depth <int>] [--budget <tokens>] [--index <path>] Compute impact briefing for target component
  impact --from-diff <base>..<head> [--repo <dir>] [--repo-id <id>] Impact briefing of the nodes a git diff touches
  slice <node_id> [--tier <tier>] [--view <view>] [--budget <tokens>] [--index <path>] Extract bounded context slice (tier: summary|contracts|deep, view: full|architecture|governance)
  explain [--node <id>] [--edge <id>] [--index <path>]      Explain discrete evidence records
  tour [--focus <node>] [--max-steps <n>] [--index <path>]  Deterministic guided reading tour from entry points or a focus node
  doc <service-id> [--format md|json|html] [--index <path>] [--checkout <dir>] [--budget <tokens>] Single-service onboarding doc (md human, json agent; same facts)
  diff <index-a> <index-b> [--format json|md]               Architecture diff: renames, nodes, edges, contracts, unsupported paths
  skills install --host claude|cursor|agents [--dir <project>] [--index <path>] Generate GIRA skills and MCP registration for a host
  skills manifest [--format json|md] [--output <path>]  Emit the deterministic skills catalog manifest
  skills check --host claude|cursor|agents [--dir <project>] [--hook] Non-mutating drift check for installed host files
  flows <entry> [--index <path>]                            Deterministic chains from an entry through consumes edges to infrastructure
  locate <query> [--index <path>]                           Search boundaries, methods, and unhandled files
  export [--format md|html] [--output <path>] [--index <p>] Export standalone architecture document
  view [--output <dir>] [--index <path>]                    Export HTML and view in browser
  eval [--artifact-verified] [--no-artifact-verified]       Eval stub ending with ARTIFACT-VERIFIED closing line (F1)
  verify --check <c> --expect <e> [--cwd <d>]               Print ARTIFACT-VERIFIED closing line for inputs (F1)
  review --findings-register <path> [--fix-check] [--gc] L3 findings register + independent fix-check (F3); --gc compacts resolved (F5)
  gate-lint --strict [--fail-on-ci] [--state <p> --digest <h>] Strict gate lint: warn-only in CI, fail-closed locally (F5); --fail-on-ci enforces in CI
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

// findProjectManifest searches for gira.yaml starting at the current working
// directory and walking up parent directories so CLI commands executed from a
// project subdirectory still resolve the enclosing project's gira.yaml.
func findProjectManifest() string {
	dir, err := os.Getwd()
	if err != nil {
		if fi, statErr := os.Stat("gira.yaml"); statErr == nil && !fi.IsDir() {
			return "gira.yaml"
		}
		return ""
	}
	for {
		candidate := filepath.Join(dir, "gira.yaml")
		if fi, statErr := os.Stat(candidate); statErr == nil && !fi.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func loadIndex(indexPath string) (*domain.MacroMap, error) {
	s := store.NewStore("")

	// 1. Explicit --index path: must exist and be valid.
	if strings.TrimSpace(indexPath) != "" {
		fi, err := os.Stat(indexPath)
		if err != nil || fi.IsDir() {
			return nil, fmt.Errorf("index_missing: index file not found: %s", indexPath)
		}
		return s.LoadFile(indexPath)
	}

	// 2. If gira.yaml exists in the current project (or an ancestor directory),
	// resolve EXCLUSIVELY the index for its declared target_system.name. Never fall back to another system in cache.
	if manifestPath := findProjectManifest(); manifestPath != "" {
		m, err := manifest.ParseFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse project manifest %s: %w", manifestPath, err)
		}
		sysName := m.TargetSystem.Name
		projectRoot := filepath.Dir(manifestPath)
		var sysCandidates []string
		sysCandidates = append(sysCandidates, filepath.Join(projectRoot, ".gira", "indexes", sysName, "macro-map.json"))
		if cwdCandidate := filepath.Join(".gira", "indexes", sysName, "macro-map.json"); cwdCandidate != sysCandidates[0] {
			sysCandidates = append(sysCandidates, cwdCandidate)
		}
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			sysCandidates = append(sysCandidates, filepath.Join(home, ".cache", "gira", "indexes", sysName, "macro-map.json"))
		}
		for _, cand := range sysCandidates {
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				return s.LoadFile(cand)
			}
		}
		return nil, fmt.Errorf("index_missing: no index found for target system %q declared in %s (run 'gira index' first)", sysName, manifestPath)
	}

	// 3. No gira.yaml in current directory hierarchy: check explicit local index files in CWD.
	localFiles := []string{
		filepath.Join(".gira", "macro-map.json"),
		filepath.Join(".gira", "index.json"),
		"macro-map.json",
		"index.json",
	}
	for _, cand := range localFiles {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return s.LoadFile(cand)
		}
	}

	// 4. Inspect local and user cache stores. Only auto-select when exactly one system exists.
	type sysCandidate struct {
		name string
		path string
	}
	var discovered []sysCandidate
	seenSys := make(map[string]bool)

	sLocal := store.NewStore(filepath.Join(".gira", "indexes"))
	if systems, err := sLocal.List(); err == nil {
		for _, sys := range systems {
			if !seenSys[sys] {
				seenSys[sys] = true
				discovered = append(discovered, sysCandidate{
					name: sys,
					path: filepath.Join(".gira", "indexes", sys, "macro-map.json"),
				})
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		sHome := store.NewStore(filepath.Join(home, ".cache", "gira", "indexes"))
		if systems, err := sHome.List(); err == nil {
			for _, sys := range systems {
				if !seenSys[sys] {
					seenSys[sys] = true
					discovered = append(discovered, sysCandidate{
						name: sys,
						path: filepath.Join(home, ".cache", "gira", "indexes", sys, "macro-map.json"),
					})
				}
			}
		}
	}

	if len(discovered) == 1 {
		return s.LoadFile(discovered[0].path)
	}
	if len(discovered) > 1 {
		names := make([]string, 0, len(discovered))
		for _, d := range discovered {
			names = append(names, d.name)
		}
		return nil, fmt.Errorf("ambiguous index: multiple systems found in cache (%s); run from a directory with gira.yaml or pass --index <path>", strings.Join(names, ", "))
	}

	return nil, fmt.Errorf("index_missing: no index found (run 'gira index' first or pass --index <path>)")
}

func runIndex(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outputPath string
	var offline bool
	var cacheDir string
	var manifestFlag string
	var allowSymbolLoss bool
	var recoverCorrupt bool

	fs.StringVar(&outputPath, "output", "", "Path to write output macro-map.json")
	fs.BoolVar(&offline, "offline", false, "Disable network egress and use local cache exclusively")
	fs.StringVar(&cacheDir, "cache-dir", "", "Local repository cache directory")
	fs.StringVar(&manifestFlag, "manifest", "", "Path to gira.yaml manifest")
	fs.BoolVar(&allowSymbolLoss, "allow-symbol-loss", false, "Explicitly authorize removal of real symbols without file deletion")
	fs.BoolVar(&recoverCorrupt, "recover-corrupt", false, "Explicitly authorize overwriting a corrupted existing index file")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	manifestPath := "gira.yaml"
	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) > 0 && strings.TrimSpace(allPositionals[0]) != "" {
		manifestPath = allPositionals[0]
	} else if strings.TrimSpace(manifestFlag) != "" {
		manifestPath = manifestFlag
	} else if found := findProjectManifest(); found != "" {
		manifestPath = found
	}

	// 1. Read manifest
	m, err := manifest.ParseFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest %s: %w", manifestPath, err)
	}

	return indexManifest(m, manifestPath, outputPath, offline, cacheDir, store.PublishOptions{
		AllowSymbolLoss: allowSymbolLoss,
		RecoverCorrupt:  recoverCorrupt,
	}, stdout)
}

// indexManifest builds the MacroMap from an already-parsed manifest and
// persists it atomically. It is the single indexing path shared by `index`
// (manifest given) and `init` (manifest generated from a local checkout).
func indexManifest(m *manifest.Manifest, manifestPath, outputPath string, offline bool, cacheDir string, pubOpts store.PublishOptions, stdout io.Writer) error {
	// 2. Initialize CacheManager
	cm, err := gitcache.NewCacheManager(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// 3–5. Fetch repos, extract, apply overrides (shared with the evaluator)
	mm, err := indexer.Build(context.Background(), m, cm, indexer.Options{
		Offline:      offline,
		ManifestPath: manifestPath,
		AllowShrink:  pubOpts.AllowSymbolLoss,
	})
	if err != nil {
		return err
	}

	// 6. Atomically publish generated MacroMap (validating previous index and symbol-loss gate)
	if outputPath != "" {
		s := store.NewStore("")
		if _, err := s.PublishFile(outputPath, mm, pubOpts); err != nil {
			return fmt.Errorf("failed to publish output file: %w", err)
		}
		fmt.Fprintf(stdout, "Indexed %d nodes and %d relationships for target system %q (saved to %s)\n",
			len(mm.Nodes), len(mm.Relationships), m.TargetSystem.Name, outputPath)
	} else {
		storeDir := filepath.Join(".gira", "indexes")
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			storeDir = filepath.Join(home, ".cache", "gira", "indexes")
		}
		s := store.NewStore(storeDir)
		if _, err := s.Publish(m.TargetSystem.Name, mm, pubOpts); err != nil {
			return fmt.Errorf("failed to persist macro map: %w", err)
		}
		dest := filepath.Join(storeDir, m.TargetSystem.Name, "macro-map.json")
		fmt.Fprintf(stdout, "Indexed %d nodes and %d relationships for target system %q (saved to %s)\n",
			len(mm.Nodes), len(mm.Relationships), m.TargetSystem.Name, dest)
	}

	return nil
}

// runInit generates gira.yaml from a local git checkout (HEAD pinned) and
// indexes it. It is the zero-config entry for a single service alone:
// `gira init [dir]` writes the manifest only when absent (`--force` repins
// the SHA), then runs the same indexing path as `gira index`. `index` itself
// never writes manifests.
func runInit(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var force bool
	var outputPath string
	var offline bool
	var cacheDir string
	var allowSymbolLoss bool
	var recoverCorrupt bool

	fs.BoolVar(&force, "force", false, "Repin the manifest SHA to the local HEAD even when gira.yaml exists")
	fs.StringVar(&outputPath, "output", "", "Path to write output macro-map.json")
	fs.BoolVar(&offline, "offline", false, "Disable network egress and use local cache exclusively")
	fs.StringVar(&cacheDir, "cache-dir", "", "Local repository cache directory")
	fs.BoolVar(&allowSymbolLoss, "allow-symbol-loss", false, "Explicitly authorize removal of real symbols without file deletion")
	fs.BoolVar(&recoverCorrupt, "recover-corrupt", false, "Explicitly authorize overwriting a corrupted existing index file")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	dir := "."
	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) > 0 && strings.TrimSpace(allPositionals[0]) != "" {
		dir = allPositionals[0]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory %q: %w", dir, err)
	}

	toplevel := gitToplevel(absDir)
	if toplevel == "" {
		return fmt.Errorf("not a git repository: %q (run 'git init' and commit first)", absDir)
	}
	head := gitHead(toplevel)
	if head == "" {
		return fmt.Errorf("failed to read HEAD of %q (no commits yet?)", toplevel)
	}
	if gitDirty(toplevel) {
		fmt.Fprintf(stderr, "warning: worktree dirty; index reflects HEAD %s, not uncommitted changes\n", head)
	}

	manifestPath := filepath.Join(absDir, "gira.yaml")
	_, statErr := os.Stat(manifestPath)
	manifestExists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("failed to inspect manifest %s: %w", manifestPath, statErr)
	}

	stagedNewManifest := !manifestExists
	stagedRepin := manifestExists && force
	loadPath := manifestPath
	if stagedNewManifest {
		// Build the manifest beside its final destination so relative paths have
		// the same base, but do not publish it until indexing has succeeded.
		tmp, err := os.CreateTemp(absDir, ".gira-init-manifest-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to stage manifest: %w", err)
		}
		loadPath = tmp.Name()
		if err := tmp.Close(); err != nil {
			_ = os.Remove(loadPath)
			return fmt.Errorf("failed to close staged manifest: %w", err)
		}
		defer os.Remove(loadPath)
		if err := writeLocalManifest(loadPath, toplevel, head); err != nil {
			return err
		}
	}

	m, err := manifest.ParseFile(loadPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest %s: %w", loadPath, err)
	}
	if stagedRepin {
		if len(m.Repositories) != 1 {
			return fmt.Errorf("gira init --force only repins single-repository manifests (found %d repositories in %s)", len(m.Repositories), manifestPath)
		}
		m.Repositories[0].SHA = head
		if err := m.Validate(); err != nil {
			return fmt.Errorf("failed to validate repinned manifest: %w", err)
		}
	}

	// Capture the indexing output until the manifest is successfully published,
	// so a failed manifest write does not print a misleading success message.
	var indexOutput bytes.Buffer
	if err := indexManifest(m, manifestPath, outputPath, offline, cacheDir, store.PublishOptions{
		AllowSymbolLoss: allowSymbolLoss,
		RecoverCorrupt:  recoverCorrupt,
	}, &indexOutput); err != nil {
		return err
	}

	if stagedRepin {
		if err := repinLocalManifest(manifestPath, head); err != nil {
			return fmt.Errorf("index was published, but repinning manifest %s failed: %w", manifestPath, err)
		}
		fmt.Fprintf(stdout, "Repinned %s to HEAD %s\n", manifestPath, head)
	} else if stagedNewManifest {
		if err := writeLocalManifest(manifestPath, toplevel, head); err != nil {
			return fmt.Errorf("index was published, but publishing manifest %s failed: %w", manifestPath, err)
		}
		fmt.Fprintf(stdout, "Initialized %s from %s (HEAD %s)\n", manifestPath, toplevel, head)
	} else {
		fmt.Fprintf(stdout, "Using existing %s\n", manifestPath)
	}
	_, err = io.Copy(stdout, &indexOutput)
	return err
}

// gitCommandEnv removes repository-redirection variables inherited from the
// caller. A hostile or stale GIT_DIR/GIT_WORK_TREE must not redirect local
// repository probes away from the path explicitly supplied with -C.
func gitCommandEnv() []string {
	blocked := map[string]bool{
		"GIT_DIR": true,
		"GIT_WORK_TREE": true,
		"GIT_COMMON_DIR": true,
		"GIT_INDEX_FILE": true,
		"GIT_OBJECT_DIRECTORY": true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
		"GIT_NAMESPACE": true,
		"GIT_CONFIG": true,
		"GIT_CONFIG_COUNT": true,
		"GIT_CONFIG_PARAMETERS": true,
	}
	env := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok || blocked[key] || strings.HasPrefix(key, "GIT_CONFIG_KEY_") || strings.HasPrefix(key, "GIT_CONFIG_VALUE_") {
			continue
		}
		env = append(env, item)
	}
	return env
}

func gitToplevel(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	cmd.Env = gitCommandEnv()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitDirty(toplevel string) bool {
	cmd := exec.Command("git", "-C", toplevel, "status", "--porcelain")
	cmd.Env = gitCommandEnv()
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// writeLocalManifest generates a single-repository manifest pinning head.
func writeLocalManifest(manifestPath, toplevel, head string) error {
	name := filepath.Base(toplevel)
	if strings.TrimSpace(name) == "" || name == "." || name == string(filepath.Separator) {
		name = "local-system"
	}
	m := &manifest.Manifest{
		Version:      1,
		TargetSystem: manifest.TargetSystemMeta{Name: name},
		Repositories: []manifest.RepositoryDeclaration{
			{
				ID:        "local",
				URL:       "file://" + filepath.ToSlash(toplevel),
				SHA:       head,
				PathScope: ".",
			},
		},
	}
	if err := m.Validate(); err != nil {
		return fmt.Errorf("failed to build local manifest: %w", err)
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to serialize local manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest %s: %w", manifestPath, err)
	}
	return nil
}

// repinLocalManifest updates the SHA of a single-repository manifest to head,
// keeping the target name, repo identity and overrides. Multi-repository
// manifests are refused: init only manages manifests it can regenerate.
func repinLocalManifest(manifestPath, head string) error {
	m, err := manifest.ParseFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest %s: %w", manifestPath, err)
	}
	if len(m.Repositories) != 1 {
		return fmt.Errorf("gira init --force only repins single-repository manifests (found %d repositories in %s)", len(m.Repositories), manifestPath)
	}
	m.Repositories[0].SHA = head
	if err := m.Validate(); err != nil {
		return fmt.Errorf("failed to validate repinned manifest: %w", err)
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to serialize repinned manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest %s: %w", manifestPath, err)
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

	var fromDiff, repoDir, repoID, checkout string

	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget; briefing default 1500)")
	fs.IntVar(&depth, "depth", 1, "Graph traversal depth")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&fromDiff, "from-diff", "", "Start from a git revision range <base>..<head> instead of a target id")
	fs.StringVar(&repoDir, "repo", ".", "Repository the --from-diff range is read from")
	fs.StringVar(&repoID, "repo-id", "", "Only match locators of this repository id (--from-diff)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")

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
	humanBanner(stderr, mm, checkout, emitFormat)
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
	var checkout string
	var budget int

	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget; tier defaults 500/1500/3500)")

	fs.StringVar(&tier, "tier", "summary", "Context slice tier (summary|contracts|deep)")
	fs.StringVar(&view, "view", "full", "Context slice architectural view (full|architecture|governance; governance keeps only Policy/Task/Skill neighbors)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")

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
	humanBanner(stderr, mm, checkout, emitFormat)
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
	var repo, checkout string
	var source, noSource bool
	var budget int

	fs.StringVar(&nodeID, "node", "", "Node ID")
	fs.StringVar(&edgeID, "edge", "", "Edge ID")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&repo, "repo", ".", "Local checkout the locators are read from (source excerpts)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.BoolVar(&source, "source", true, "Fetch source excerpts from --repo (default true)")
	fs.BoolVar(&noSource, "no-source", false, "Do not fetch source excerpts (locators only)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")

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

	emitFormat := slices.ParseFormat(strings.ToLower(format))
	humanBanner(stderr, mm, checkout, emitFormat)

	if budget > 0 {
		data, err := report.Emit(emitFormat, budget)
		if err != nil {
			return err
		}
		return emitHuman(stdout, data, emitFormat)
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		text := report.RenderMarkdown()
		if source && !noSource {
			text += sourceSection(report, repo)
		}
		if err := emitMarkdownLine(stdout, text); err != nil {
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
	var checkout string
	var budget int

	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")

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

	emitFormat := slices.ParseFormat(strings.ToLower(format))
	humanBanner(stderr, mm, checkout, emitFormat)

	if budget > 0 {
		br := &slices.BoundaryResult{Matches: matches}
		data, err := br.Emit(emitFormat, budget)
		if err != nil {
			return err
		}
		return emitHuman(stdout, data, emitFormat)
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
	var checkout string

	fs.StringVar(&format, "format", "md", "Export format (md|html)")
	fs.StringVar(&outputPath, "output", "", "Output file path (default stdout)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	humanBanner(stderr, mm, checkout, slices.FormatMarkdown)

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
	var checkout string
	var noOpen bool

	fs.StringVar(&outputDir, "output", "", "Directory to generate HTML artifact (default temp dir)")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.BoolVar(&noOpen, "no-open", false, "Write the artifact but do not open a browser")

	if err := fs.Parse(args); err != nil {
		return err
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	humanBanner(stderr, mm, checkout, slices.FormatMarkdown)
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

	// Attempt non-blocking browser open unless --no-open or headless/CI.
	openInBrowser(absPath, noOpen)

	return nil
}

func openInBrowser(filePath string, noOpen bool) {
	// --no-open, or a CI/NO_BROWSER environment, never launches a browser.
	if noOpen {
		return
	}
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
	var focus, indexPath, format, checkout string
	var maxSteps, budget int
	fs.StringVar(&focus, "focus", "", "Start the tour at this node instead of the entry points")
	fs.IntVar(&maxSteps, "max-steps", slices.DefaultTourSteps, "Maximum number of tour steps")
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")
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
	tourFormat := slices.ParseFormat(strings.ToLower(format))
	humanBanner(stderr, mm, checkout, tourFormat)
	if budget > 0 {
		data, err := tour.Emit(tourFormat, budget)
		if err != nil {
			return err
		}
		return emitHuman(stdout, data, tourFormat)
	}
	if tourFormat == slices.FormatMarkdown {
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
	var indexPath, format, checkout string
	var budget int
	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")
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
	flowsFormat := slices.ParseFormat(strings.ToLower(format))
	humanBanner(stderr, mm, checkout, flowsFormat)
	if budget > 0 {
		data, err := flows.Emit(flowsFormat, budget)
		if err != nil {
			return err
		}
		return emitHuman(stdout, data, flowsFormat)
	}
	if flowsFormat == slices.FormatMarkdown {
		return emitMarkdownLine(stdout, flows.RenderMarkdown())
	}
	data, err := flows.CanonicalJSON()
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

// runDoc renders the F3 single-service onboarding document: Markdown or HTML
// for humans, a budget-envelope JSON for agents, all from the same engine doc
// so they cover the same facts.
func runDoc(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("doc", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var indexPath, format, checkout string
	var budget int

	fs.StringVar(&indexPath, "index", "", "Path to macro-map.json")
	fs.StringVar(&format, "format", "md", "Output format (md|json|html)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	all := append(positionals, fs.Args()...)
	if len(all) == 0 {
		return fmt.Errorf("missing service_id argument. Usage: gira doc <service-id> [--format md|json|html] [--index <path>]")
	}

	var docFormat slices.Format
	switch strings.ToLower(format) {
	case "md", "markdown", "":
		docFormat = slices.FormatMarkdown
	case "json":
		docFormat = slices.FormatJSON
	case "html":
		docFormat = slices.FormatHTML
	default:
		return fmt.Errorf("unsupported doc format %q. Supported formats: md, json, html", format)
	}

	mm, err := loadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	engine := slices.NewEngine(mm)
	doc, err := engine.GetDoc(all[0])
	if err != nil {
		return err
	}

	data, err := doc.Emit(docFormat, budget)
	if err != nil {
		return err
	}
	humanBanner(stderr, mm, checkout, docFormat)
	return emitHuman(stdout, data, docFormat)
}

func runDiff(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var format, checkout string
	var budget int
	fs.StringVar(&format, "format", "json", "Output format (json|md)")
	fs.StringVar(&checkout, "checkout", "", "Local checkout whose HEAD is compared against the indexed commits (staleness)")
	fs.IntVar(&budget, "budget", 0, "Hard ceiling on estimated output tokens (0 = no budget)")
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
	diffFormat := slices.ParseFormat(strings.ToLower(format))
	humanBanner(stderr, a, checkout, diffFormat)
	d := archdiff.Compare(a, b)
	if budget > 0 {
		data, err := d.Emit(strings.ToLower(format), budget)
		if err != nil {
			return err
		}
		return emitHuman(stdout, data, diffFormat)
	}
	if diffFormat == slices.FormatMarkdown {
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
	if len(args) > 0 && args[0] == "manifest" {
		return runSkillsManifest(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "check" {
		return runSkillsCheck(args[1:], stdout, stderr)
	}
	if len(args) == 0 || args[0] != "install" {
		return fmt.Errorf("usage: gira skills install|manifest|check --host %s [--dir <project>] [--index <path>] [--hook]", strings.Join(skills.Hosts, "|"))
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
	if _, err = fmt.Fprintln(stdout, string(data)); err != nil {
		return err
	}
	// The stdout report stays machine-readable; the hint for agents goes to stderr.
	fmt.Fprintln(stderr, skills.InstallHint)
	return nil
}

func runSkillsManifest(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("skills manifest", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var format, output string
	fs.StringVar(&format, "format", "json", "Output format: json|md")
	fs.StringVar(&output, "output", "", "Write manifest to <path> instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if format != "json" && format != "md" {
		return fmt.Errorf("invalid --format %q: want json|md", format)
	}
	m, err := skills.BuildManifest(version)
	if err != nil {
		return err
	}
	var data []byte
	if format == "md" {
		data = []byte(m.Markdown())
	} else {
		data, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
	}
	if output != "" {
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return err
		}
		return os.WriteFile(output, data, 0o644)
	}
	_, err = stdout.Write(data)
	return err
}

func runSkillsCheck(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("skills check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var host, dir, index, command string
	fs.StringVar(&host, "host", "", "Host family: "+strings.Join(skills.Hosts, ", "))
	fs.StringVar(&dir, "dir", ".", "Project directory to check")
	fs.StringVar(&index, "index", "", "Index the registered MCP server should load")
	fs.StringVar(&command, "command", "gira", "Command the host runs to start the MCP server")
	var hook bool
	fs.BoolVar(&hook, "hook", false, "claude host: also check the opt-in PreToolUse nudge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rep, err := skills.Check(host, skills.Options{Root: dir, Index: index, Command: command, Hook: hook})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, string(data)); err != nil {
		return err
	}
	if !rep.InSync {
		return fmt.Errorf("skills check: host %q out of sync (%d missing, %d drifted)", host, len(rep.Missing), len(rep.Drifted))
	}
	return nil
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
