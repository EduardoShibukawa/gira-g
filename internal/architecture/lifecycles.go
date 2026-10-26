package architecture

// DefaultLifecycles returns the canonical execution flows across architectural layers.
func DefaultLifecycles() []LifecycleFlow {
	return []LifecycleFlow{
		{
			Intent:      "ast_edit",
			Title:       "Lifecycle of a Surgical AST Edit",
			Description: "Deterministic retrieval, brace-balanced boundary detection, and atomic surgical replacement of a code symbol without full-file reformatting or external language servers.",
			MermaidSeq: `sequenceDiagram
  autonumber
  actor User as Human / AI Agent
  participant CLI as CLI (cmd/gira:runWriteSymbol)
  participant Manip as AST Manipulator (internal/manipulator)
  participant Parser as Language Parser (Go/TS/Python)
  participant Brace as SafeBraceCounter
  participant Disk as File System (atomic write)

  User->>CLI: gira write <file> <symbol> --body "..."
  CLI->>Manip: WriteSymbol(filePath, symbolQuery, newContent)
  Manip->>Parser: FindSymbols(filePath)
  Parser->>Parser: Tokenize source & identify symbol ranges
  Manip->>Brace: SafeBraceCounter.Count(line, inString, inComment)
  Brace-->>Manip: Exact start/end line bounds with zero string/comment bleed
  Manip->>Manip: Splice newContent at exact indentation
  Manip->>Disk: Atomic write to target file
  Manip-->>CLI: Success (symbol updated)
  CLI-->>User: successfully updated symbol <query> in <file>`,
			Steps: []LifecycleStep{
				{
					Order:       1,
					Component:   "cmd/gira",
					Domain:      string(DomainEntrypoints),
					Action:      "Parses CLI arguments (--body or piped stdin) and verifies target file existence.",
					Artifact:    "Invocation Parameters",
					CodeLocator: "file:///cmd/gira/symbol_ops.go#L48-L91",
				},
				{
					Order:       2,
					Component:   "internal/manipulator",
					Domain:      string(DomainManipulator),
					Action:      "Invokes FindSymbols to detect all class, interface, function, and method ranges.",
					Artifact:    "[]SymbolRange",
					CodeLocator: "file:///internal/manipulator/manipulator.go#L36-L54",
				},
				{
					Order:       3,
					Component:   "internal/manipulator (SafeBraceCounter)",
					Domain:      string(DomainManipulator),
					Action:      "Tokenizes code boundaries with string literals and comment escapes to identify true block termination.",
					Artifact:    "Matched SymbolRange",
					CodeLocator: "file:///internal/manipulator/manipulator.go#L430-L500",
				},
				{
					Order:       4,
					Component:   "internal/manipulator",
					Domain:      string(DomainManipulator),
					Action:      "Slices exact lines, adjusts indentation to preserve container style, and performs atomic string join.",
					Artifact:    "Updated File Content",
					CodeLocator: "file:///internal/manipulator/manipulator.go#L85-L140",
				},
				{
					Order:       5,
					Component:   "os.WriteFile",
					Domain:      string(DomainStorage),
					Action:      "Writes updated buffer back to disk with original permissions (0644).",
					Artifact:    "Written File on Disk",
					CodeLocator: "file:///internal/manipulator/manipulator.go#L135-L139",
				},
			},
		},
		{
			Intent:      "impact",
			Title:       "Lifecycle of an Impact Calculation",
			Description: "Deterministic traversal of mediated contracts, upstream dependencies, and downstream consumers with bounded token envelopes.",
			MermaidSeq: `sequenceDiagram
  autonumber
  actor User as Human / AI Agent
  participant CLI as CLI / MCP (cmd/gira:runImpact)
  participant Store as Store / Resolver (internal/store)
  participant Engine as Slices Engine (internal/slices)
  participant Domain as MacroMap & Evidence (internal/domain)
  participant Tokens as Token Budgeting (internal/tokens)

  User->>CLI: gira impact <target_id> --depth 2 --budget 500
  CLI->>Store: LoadMacroMap(indexPath)
  Store-->>CLI: MacroMap (with Merkle verification)
  CLI->>Engine: NewEngine(mm).GetImpactBriefing(targetID, depth)
  Engine->>Domain: Traverse providesByContract & consumesByContract
  Engine->>Domain: Collect UnsupportedFiles & KnownUnknowns
  Engine-->>CLI: ImpactBriefing struct
  CLI->>Tokens: Fit(budget, dropOrder, renderFn)
  Tokens-->>CLI: Bounded JSON or Markdown envelope
  CLI-->>User: Formatted Impact Briefing`,
			Steps: []LifecycleStep{
				{
					Order:       1,
					Component:   "cmd/gira / internal/mcp",
					Domain:      string(DomainEntrypoints),
					Action:      "Receives target_id, depth, and budget parameters via CLI flag or MCP JSON-RPC call.",
					Artifact:    "Query Request",
					CodeLocator: "file:///cmd/gira/main.go#L973-L1041",
				},
				{
					Order:       2,
					Component:   "internal/store",
					Domain:      string(DomainStorage),
					Action:      "Loads macro-map.json, validates schema version, and inspects repository HEAD freshness.",
					Artifact:    "domain.MacroMap",
					CodeLocator: "file:///internal/store/store.go#L30-L75",
				},
				{
					Order:       3,
					Component:   "internal/slices (Engine)",
					Domain:      string(DomainCore),
					Action:      "Executes breadth-first traversal across provides and consumes mediation edges.",
					Artifact:    "Impact Subgraph",
					CodeLocator: "file:///internal/slices/engine.go#L231-L320",
				},
				{
					Order:       4,
					Component:   "internal/domain",
					Domain:      string(DomainCore),
					Action:      "Appends unhandled files and unknown language constructs to preserve known unknowns.",
					Artifact:    "[]UnsupportedBoundary",
					CodeLocator: "file:///internal/domain/models.go#L112-L120",
				},
				{
					Order:       5,
					Component:   "internal/tokens",
					Domain:      string(DomainCore),
					Action:      "Applies deterministic token-fitting drops if requested budget is exceeded.",
					Artifact:    "Budget Envelope (JSON or Markdown)",
					CodeLocator: "file:///internal/slices/budget.go#L80-L150",
				},
			},
		},
		{
			Intent:      "index",
			Title:       "Lifecycle of Multi-Layer Architecture Extraction & Indexing",
			Description: "Full repository discovery: manifest parsing, Layer 1 repository metadata, Layer 2 infrastructure declarations, and Layer 3 source code AST extraction with Merkle hash integrity.",
			MermaidSeq: `sequenceDiagram
  autonumber
  actor User as CI / Developer
  participant CLI as CLI (cmd/gira:runIndex)
  participant Manifest as Manifest Parser (internal/manifest)
  participant Cache as Git Cache Manager (internal/gitcache)
  participant Pipeline as Extractor Pipeline (internal/extractors)
  participant Indexer as Indexer Orchestrator (internal/indexer)
  participant Store as Deterministic Store (internal/store)

  User->>CLI: gira index gira.yaml --output .gira/macro-map.json
  CLI->>Manifest: Parse(gira.yaml)
  CLI->>Cache: NewCacheManager(cacheDir)
  CLI->>Indexer: Build(ctx, manifest, cacheManager, opts)
  Indexer->>Pipeline: DefaultPipeline().Extractors()
  Pipeline->>Pipeline: Run Layer 1 (Repo Data)
  Pipeline->>Pipeline: Run Layer 2 (Compose, Docker, Helm, Mise, OpenAPI)
  Pipeline->>Pipeline: Run Layer 3 (Go AST, TS AST, Python AST, Protobuf)
  Pipeline-->>Indexer: Raw Nodes, Relationships, Unsupported Boundaries
  Indexer->>Indexer: Resolve cross-file intra-service references
  Indexer-->>CLI: Canonical MacroMap
  CLI->>Store: BuildSourceIdentityFromManifest(...)
  CLI->>Store: WriteMacroMapSafe(mm, outputPath)
  Store-->>CLI: Atomic persist with lock
  CLI-->>User: Successfully indexed macro-map.json`,
			Steps: []LifecycleStep{
				{
					Order:       1,
					Component:   "cmd/gira:runIndex",
					Domain:      string(DomainEntrypoints),
					Action:      "Parses gira.yaml path and configuration options.",
					Artifact:    "Run Options",
					CodeLocator: "file:///cmd/gira/main.go#L393-L440",
				},
				{
					Order:       2,
					Component:   "internal/manifest",
					Domain:      string(DomainCore),
					Action:      "Parses manifest YAML, validating repository URLs, pinned commit SHAs, and path scopes.",
					Artifact:    "manifest.ProjectManifest",
					CodeLocator: "file:///internal/manifest/manifest.go#L30-L80",
				},
				{
					Order:       3,
					Component:   "internal/gitcache",
					Domain:      string(DomainStorage),
					Action:      "Clones or updates local bare mirrors and pins worktrees to specified commit SHAs.",
					Artifact:    "Local Worktrees",
					CodeLocator: "file:///internal/gitcache/gitcache.go#L40-L110",
				},
				{
					Order:       4,
					Component:   "internal/extractors",
					Domain:      string(DomainExtractors),
					Action:      "Executes Layer 1 (Data), Layer 2 (Compose/Docker/Mise), and Layer 3 (Language ASTs: Go, TS, Py, Proto).",
					Artifact:    "Extracted Nodes & Evidence",
					CodeLocator: "file:///internal/extractors/extractor.go#L45-L95",
				},
				{
					Order:       5,
					Component:   "internal/indexer",
					Domain:      string(DomainCore),
					Action:      "Synthesizes mediated edges, resolves contract providers and consumers, and tags intra-service metadata.",
					Artifact:    "domain.MacroMap",
					CodeLocator: "file:///internal/indexer/indexer.go#L50-L160",
				},
				{
					Order:       6,
					Component:   "internal/store",
					Domain:      string(DomainStorage),
					Action:      "Generates cryptographic SHA256 Merkle source identity and writes sorted, indented JSON under a cross-platform lock.",
					Artifact:    "macro-map.json",
					CodeLocator: "file:///internal/store/store.go#L80-L140",
				},
			},
		},
		{
			Intent:      "gate",
			Title:       "Lifecycle of Governance Gate & Drift Verification",
			Description: "Automated verification of architectural invariants, contract break detection, untracked construct reporting, and scorecard generation.",
			MermaidSeq: `sequenceDiagram
  autonumber
  actor User as CI / Developer
  participant CLI as CLI (cmd/gira:runGate)
  participant Gate as Gate Engine (internal/gate)
  participant Eval as Eval & Scorecard (internal/eval)
  participant Diff as Arch Diff (internal/archdiff)

  User->>CLI: gira gate --status
  CLI->>Gate: Evaluate(policy, macroMap, baselineMap)
  Gate->>Diff: ComputeDiff(baselineMap, macroMap)
  Diff-->>Gate: Added/Removed Nodes, Edges, Contracts
  Gate->>Eval: CalculateScorecard(metrics)
  Eval-->>Gate: Scorecard Record (Precision, FastPass, Drift)
  Gate-->>CLI: GateStatus (PASS / FAIL / WARN)
  CLI-->>User: Exit 0 (passed) or Nonzero (drift/violations)`,
			Steps: []LifecycleStep{
				{
					Order:       1,
					Component:   "cmd/gira:runGate",
					Domain:      string(DomainEntrypoints),
					Action:      "Initializes gate flags (--status, --reverify, --strict, --fail-on-ci).",
					Artifact:    "Gate Config",
					CodeLocator: "file:///cmd/gira/gate.go#L20-L70",
				},
				{
					Order:       2,
					Component:   "internal/gate",
					Domain:      string(DomainGovernance),
					Action:      "Loads policy rules and evaluates current MacroMap against baseline snapshot.",
					Artifact:    "Evaluation Context",
					CodeLocator: "file:///internal/gate/gate.go#L30-L90",
				},
				{
					Order:       3,
					Component:   "internal/archdiff",
					Domain:      string(DomainGovernance),
					Action:      "Calculates structural differences: node renames, removed contracts, and newly unhandled boundaries.",
					Artifact:    "archdiff.DiffResult",
					CodeLocator: "file:///internal/archdiff/archdiff.go#L25-L95",
				},
				{
					Order:       4,
					Component:   "internal/eval",
					Domain:      string(DomainGovernance),
					Action:      "Computes objective scorecards and checks against regression floor guards.",
					Artifact:    "eval.Scorecard",
					CodeLocator: "file:///internal/eval/scorecard.go#L40-L115",
				},
				{
					Order:       5,
					Component:   "cmd/gira",
					Domain:      string(DomainEntrypoints),
					Action:      "Emits human-readable or JSON findings report; fails closed if strict gate criteria fail.",
					Artifact:    "Gate Report",
					CodeLocator: "file:///cmd/gira/gate.go#L75-L120",
				},
			},
		},
	}
}
