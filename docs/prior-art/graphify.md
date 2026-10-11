# Prior art: Graphify

[Graphify](https://github.com/Graphify-Labs/graphify): Apache-2.0 (MIT before 2026-07-22), Python. A symbol/file-level code graph; no Component concept.

## How we use this tool

Studied, never depended on. Decided in [How to reuse Graphify and Understand Anything](../../.scratch/system-context-harness/issues/10-reuse-strategy.md).

- Run only by hand, in a throwaway environment outside this repo's mise toolchain.
- Never in `mise.toml`, CI, git hooks, AGENTS.md or any agent config in this repo. Its output is never committed (see `.gitignore`).
- Reimplement ideas; don't copy. Any verbatim copy (code, prompts, fixtures, queries) keeps its upstream license header and gets an entry in `THIRD_PARTY_NOTICES`.
- Each study run appends a dated section below, naming the tool version/commit and labelling each finding **Fact**, **Inference** or **Open**.

## Ideas we borrow, and who decides them

| Idea | Decided by |
|---|---|
| Discrete confidence rubric; explicit origin/tier field; external stub nodes so edges never dangle | [Minimum Macro Map schema and evidence model](../../.scratch/system-context-harness/issues/12-macro-map-schema.md) |
| Canonical sorted JSON; cache keyed on content hash + Extractor version + schema; git merge driver for a committed index | Storage and incremental refresh (map fog) |
| Repo-tag prefix + `local_id` on merge | Component identity (map fog) |
| Token-budgeted, provenance-annotated text output | [What should a context slice look like?](../../.scratch/system-context-harness/issues/15-context-slice-shape.md) |
| Always-on agent instruction block + pre-tool hook as a distribution channel | [Agent integration surface](../../.scratch/system-context-harness/issues/14-agent-integration-surface.md) |
| One-off comparison: Graphify (code-only, pinned) scored by hand on the Reference Corpus, no evaluation-pipeline baseline | [Evaluation design and go/no-go thresholds](../../.scratch/system-context-harness/issues/17-evaluation-design.md) |

What we diverge on: "EXTRACTED" must mean deterministic; one edge per node pair loses evidence; node IDs must be stable.

## Study 1 — 2026-10-09 (v0.9.82, 5b74d7d)

Originally on the throwaway branch `research/graphify` (commit 131ac94).

Ticket: `.scratch/system-context-harness/issues/04-graphify-reference.md`
Source examined: `Graphify-Labs/graphify` at commit `5b74d7d` (tag v0.9.82, 2026-10-09), shallow-cloned and run locally (`graphifyy` 0.9.82, Python 3.14, `--code-only` on hand-made fixtures). All file:line links are pinned to that commit.

Labels: **Fact** = confirmed in code, tests, docs or by running it (with a citation). **Inference** = my engineering judgement. **Open** = not resolved.

### Summary

- Graphify is a **Python library + CLI + AI-assistant "skill"**. It turns a folder into a **symbol/file-level knowledge graph** (`graph.json`, NetworkX node-link JSON). Tree-sitter covers about 40 code formats. LLM subagents cover docs, papers and images. Leiden/Louvain communities stand in for "architecture".
- It has **no Component or Service concept**. It has no routes, compose/k8s, OpenAPI/proto or HTTP/queue edges between services, and its "cross-repo" support is a merge of per-repo graphs keyed by name. On a two-service fixture it gave no edge between the services (**Fact**, run below).
- Provenance per edge is `confidence` (EXTRACTED/INFERRED/AMBIGUOUS) + `confidence_score` + `_origin` (ast/semantic) + `source_file` + `source_location` (line only). It does **not** record which extractor produced the edge, its version, a column/byte range or a commit per edge. "EXTRACTED" does **not** mean deterministic, because the LLM is also told to emit EXTRACTED.
- The persisted graph is a **simple graph: one edge per node pair**. Several relations between the same two nodes collapse by a precedence rule.
- Very active, very popular and open-core: Apache-2.0 (relicensed from MIT on 2026-07-22), daily releases, about 125k stars, about 1.5k open issues, commercial "Graphify Cloud/Enterprise" upsell.
- **For a Go/Rust harness:** treat Graphify as an **optional input through an adapter** (read `graph.json` as symbol-level evidence and roll it up to Components) and **borrow some schema ideas**. Do not fork it and do not depend on it at runtime.

---

### 1. Deterministic extraction and supported languages

- **Fact:** The pipeline is `detect → extract → build → cluster → analyze → report → export`. Stages pass plain dicts and NetworkX graphs ([ARCHITECTURE.md:5-11][arch]).
- **Fact:** Three passes: (1) tree-sitter AST for code, no LLM; (2) local faster-whisper for audio/video; (3) LLM subagents for docs, papers and images. A code-only corpus skips pass 3 ([docs/how-it-works.md][hiw]).
- **Fact:** Dispatch is a hard-coded suffix → function table, `_DISPATCH` ([extract.py:7205][dispatch]), plus shebang dispatch. Code suffixes: [detect.py:47][detect]. Covered: Python, JS/TS (+Vue/Svelte/Astro), Go, Rust, Java, Groovy/Gradle, Kotlin, Scala, C/C++/CUDA/Metal, C#, VB.NET, Razor/XAML/.sln/.csproj, Ruby, PHP, Swift, ObjC, Lua, Zig, Dart, Elixir, Erlang, Julia, R, Fortran, Pascal/Delphi, COBOL, Apex, Solidity, Verilog, OCaml, Common Lisp, Bash, PowerShell, SQL, Terraform/HCL, JSON, Markdown, Robot, DM. Several of these need optional extras ([pyproject.toml][pyproject] `_EXTRA_FOR_EXTENSION` near [extract.py:7205][dispatch]).
- **Fact:** YAML, `.txt`, `.rst` and `.html` are classed as **documents**, i.e. LLM-only ([detect.py:48][detect]). No deterministic extractor exists for docker-compose, Kubernetes, Helm, Dockerfile, OpenAPI or protobuf. Proto/OpenAPI codegen output is only *detected* so it can be skipped as generated code ([extract.py:1712-1717][gen]).
- **Fact:** Package manifests are parsed deterministically: `pyproject.toml`, `Cargo.toml`, `go.mod`, `pom.xml`, `apm.yml` → `package` nodes + `depends_on` edges ([manifest_ingest.py:22-29][manifest]). `package.json` goes through the generic JSON extractor, which emits key nodes plus `imports` to `ref_<dep>` (seen in my run). `--cargo` adds crate→crate workspace edges ([cargo_introspect.py][cargo]). `--postgres` introspects a live database ([pg_introspect.py][pg]). MCP config files are indexed too ([mcp_ingest.py][mcpingest]).
- **Fact (run):** a `depends_on` edge from `go.mod` to a third-party module is emitted by the extractor but **dropped at build**. Only import-family relations get "external stub" nodes ([build.py:94][stubrel]). As a result, `depends_on` survives only between packages that are both in the corpus.
- **Fact:** Cross-file resolution is per-language heuristic code, not compiler-grade. It lives in the ~9.5k-line `extract.py` plus `extractors/resolution.py`, `symbol_resolution.py` and others. Recent releases are mostly resolution-accuracy fixes ([CHANGELOG.md 0.9.80-0.9.82][changelog]).
- **Fact:** Determinism is a stated contribution rule ([CONTRIBUTING.md:15][contrib]). A portability regression guard checks that graphs are identical across checkout paths, warm/cold builds and line endings (CHANGELOG 0.9.81). Output JSON uses a canonical key order and sorting ([export.py:400-416][tojson]).
- **Fact (run):** **node ids depend on corpus contents.** One Go service gave `cmd_orderservice`. Two services with the same `cmd/main.go` + `type OrderService` gave `a_cmd_main_go_cmd_orderservice` / `b_cmd_main_go_cmd_orderservice`: the Go extractor keys on parent dir name ([extractors/go.py:111][goid]) and a later disambiguation remap kicks in. **Inference:** ids are not stable when unrelated files are added, so they are not a safe long-lived identity key for Components.
- **Fact:** Clustering uses Leiden (graspologic, optional extra) with `seed=42`, and falls back to NetworkX Louvain ([cluster.py:23-127][cluster]). **Inference:** community ids and labels can differ depending on which extras are installed.

### 2. Persisted graph schema

- **Fact:** Output lives in `graphify-out/`: `graph.json`, `GRAPH_REPORT.md`, `graph.html`, `.graphify_analysis.json`, `manifest.json` (per-file mtime + hashes), `cache/ast/v<ver>-s<schema>/…`, `cache/semantic/p<promptfp>/…`, and an optional `wiki/` and `memory/` (seen in my run; [cache.py:9-39][cache]).
- **Fact:** `graph.json` is NetworkX node-link JSON with top-level keys `directed, multigraph, graph{schema_version:1, graphify_version}, nodes, links, hyperedges, built_at_commit` ([export.py:176][schemaver], [export.py:451-460][tojsonmeta]; confirmed by my run).
- **Fact:** The *extraction fragment* contract (what any extractor, including an LLM, must emit) is enforced by [validate.py:4-7][validate]:
  - node: required `id, label, file_type, source_file`; `file_type ∈ {code, document, paper, image, rationale, concept}`
  - edge: required `source, target, relation, confidence, source_file`; `confidence ∈ {EXTRACTED, INFERRED, AMBIGUOUS}`
  - `relation` is a **free string** (no enum). Observed AST relations include `contains, calls, indirect_call, imports, imports_from, re_exports, references, method, inherits, implements, uses, defines, depends_on, instantiates, bound_to…`. The LLM spec adds `cites, conceptually_related_to, shares_data_with, semantically_similar_to, rationale_for`.
- **Fact (run):** extra node fields in practice: `_origin, _callable, norm_label, source_location ("L12"), community, community_name, type (package/external), ecosystem, external, metadata{namespace, unresolved_calls…}`. Extra edge fields: `_origin, context (call/import/parameter_type…), confidence_score, source_location, weight`. Hyperedges (3+ members) are stored in both `graph.hyperedges` and top-level `hyperedges` ([how-it-works.md][hiw], [cli.py merge-graphs][mergecli]).
- **Fact:** The file says `"directed": false`, but **link order is the direction**. Readers must use `paths.load_node_link_graph()` or `_src/_tgt` markers ([ARCHITECTURE.md "Edge direction"][arch], [export.py:386-399][tojson]).
- **Fact:** **One edge per unordered node pair.** `build_from_json` uses `nx.Graph`/`nx.DiGraph` ([build.py:1149][buildgraph]). Collisions are settled by a rule: specific beats generic (`references/uses/mentions`), then higher confidence rank, then the edge that has a location, then the higher score ([build.py:79-80][genrel], [build.py:1530-1600][collapse]). Opt-in MultiDiGraph mode exists only as a capability probe with "No call sites added yet" ([multigraph_compat.py:8][mdg]). The CLI has `diagnose multigraph` to report collapse risk.
- **Fact:** IDs are `[a-z0-9_]` slugs built from the repo-relative path stem + symbol name ([extraction-spec.md][spec], [extractors/base.py `_file_stem`][base]). There are no typed identifiers (no SCIP-style symbol strings).

### 3. Provenance: extracted vs inferred

- **Fact:** Each edge carries `confidence` + `confidence_score`. The rubric is EXTRACTED = 1.0; INFERRED ∈ {0.95, 0.85, 0.75, 0.65, 0.55}; AMBIGUOUS 0.1–0.3. Defaults applied on write: 1.0 / 0.55 / 0.2 ([export.py:197][confdef], [extraction-spec.md][spec]).
- **Fact:** **INFERRED is produced by deterministic code too**, not only by the LLM. Examples: a cross-file call with no import evidence gets INFERRED 0.85, and is promoted to EXTRACTED when there is an import edge ([extract.py:9178-9201][promote]); `indirect_call` (function passed as a value) gets INFERRED 0.85 ([extract.py:9143-9158][indirect]); a receiver-typed member call is INFERRED unless type-qualified ([extract.py:4150-4170][recv]). The cross-repo `same_type_as` edge is INFERRED 0.85 ([cross_repo_types.py][xtypes]).
- **Fact:** **The LLM is told to emit EXTRACTED** for "explicit" relations ([extraction-spec.md][spec]). The only marker of LLM origin is `_origin`: `ast` is stamped by `extract()` ([extract.py:9427][originstamp]); semantic items are unstamped at first and backfilled to `semantic` from their shape on reload ([build.py:45-68][origin]). When an AST node and an LLM node share a (file, label), the AST node wins ([build.py:1180-1225][astwins]).
- **Fact (run):** the evidence on an edge is `source_file` + `source_location` as a line ("L14") only. There is no column/byte span, snippet, extractor name/version, or per-edge commit. The commit (`built_at_commit`) and `graphify_version` exist only at graph level. **Fact (run):** `merge-graphs` output **drops `built_at_commit`**, so a merged graph has no commit provenance per repo.
- **Fact:** Human feedback exists as a Q&A "work memory": `save-result --outcome useful|dead_end|corrected` writes to `graphify-out/memory/`, and `reflect` aggregates it into LESSONS.md (CLI help). There is **no** mechanism to correct or override nodes and edges with attribution.

### 4. Cross-repository merging

- **Fact:** `graphify merge-graphs g1.json g2.json … --out` loads each graph, prefixes every non-external node id with `<repo_tag>::`, sets `repo` + `local_id`, offsets community ids, and composes everything into one **undirected simple** `nx.Graph` ([cli.py:2730-2900][mergecli], [build.py:2529-2575][prefix]). External stubs (stdlib/third-party) stay unprefixed so they unify across repos.
- **Fact:** `graphify global add|remove|list` keeps `~/.graphify/global-graph.json` + `global-manifest.json`. It works incrementally per repo tag (prune that repo, re-add, skip when the source hash is unchanged) ([global_graph.py:79-170][global]).
- **Fact:** There are only two post-merge cross-repo linkers. (a) `same_type_as` links type declarations with the **same namespace + name** in different repos, which in practice means .NET and similar ([cross_repo_types.py][xtypes]). (b) `calls` links a member call whose receiver type was "parked" as `metadata.unresolved_calls` to a single declaration in another repo; only C++, C#, Java and Swift park such calls ([cross_repo_calls.py:30-50, 151][xcalls]).
- **Fact (run):** for a fixture with a Go `orders` service (chi route `/orders`), a TS `web` client calling `axios.get("http://orders:8080/orders")`, and a `docker-compose.yml` with `depends_on`, the graph had **no service nodes, no route node and no web→orders edge**. Compose was skipped as a document under `--code-only`.
- **Inference:** Package nodes from manifests are not `external`, so they get repo-prefixed on merge. A service's own `pkg_x` and a sibling repo's reference to `x` therefore do not unify across repos.
- **Inference:** Graphify's cross-repo story is "union of symbol graphs + name joins". It does not cover how services talk at runtime (HTTP, gRPC, queues, DB sharing), which is the Macro Map's core.

### 5. Query interface

- **Fact:** CLI commands: `query "<q>" [--dfs] [--budget N] [--context C]`, `path A B`, `explain X`, `affected X [--relation] [--depth]` (reverse traversal), `god-nodes`, `benchmark` (CLI help; [serve.py][serve]).
- **Fact:** Query matches query terms against `label/norm_label/source` to pick seed nodes ([serve.py:548-575][score]), then runs BFS/DFS to a fixed depth ([serve.py:1034][bfs]) and renders a **token-budgeted text** subgraph ([serve.py:1092][subtext]). There are no embeddings or vector store ([how-it-works.md][hiw]).
- **Fact (run):** the output lines carry provenance, e.g. `EDGE handler() --calls [EXTRACTED context=call]--> .List() at=orders/cmd/main.go:L14`.
- **Fact:** Other "views": `GRAPH_REPORT.md` (god nodes, surprising connections, suggested questions, AMBIGUOUS audit), a wiki with one article per community, `tree` (D3), and `callflow-html` (Mermaid architecture/call-flow) ([ARCHITECTURE.md table][arch]).

### 6. MCP integration

- **Fact:** `graphify-mcp <graph.json>` (stdio) and `serve_http` (Starlette, default `127.0.0.1`, path `/mcp`, optional API-key middleware) ([serve.py:2556, 2714][servehttp], [serve.py:2600][serveauth]). Needs the `[mcp]` extra (`mcp>=1,<3`).
- **Fact (run):** the server reports `graphify` / protocol `2025-06-18` and exposes 10 tools: `query_graph, get_node, get_neighbors, get_community, god_nodes, graph_stats, shortest_path, list_prs, get_pr_impact, triage_prs` ([serve.py:1990-2120][tools]). Six resources: `graphify://report|stats|god-nodes|surprises|audit|questions` ([serve.py:2428-2433][resources]). Every tool accepts an optional `project_path` to serve several projects.
- **Fact:** The server reloads when the `(mtime_ns, size)` of graph.json changes ([serve.py:171][reload]). `graph_stats` reports the build commit and whether it is behind HEAD (CHANGELOG 0.9.80).
- **Fact:** Agent integration also goes beyond MCP. `graphify install` / `<agent> install` write CLAUDE.md, AGENTS.md, GEMINI.md, Cursor rules and similar blocks ([always_on/][alwayson]) for about 17 hosts, plus a Claude/CodeBuddy **PreToolUse hook** (`hook-guard`) that nudges or blocks raw file reads in favour of graph queries ([cli.py:2586-2597][hookguard]).

### 7. Refresh / incremental behaviour

- **Fact:** The AST cache is per file, keyed by content hash and namespaced by package version + cache schema (`_AST_CACHE_SCHEMA = 7`), so extractor upgrades invalidate it. The semantic (LLM) cache is keyed by a **prompt fingerprint**, so it survives releases that leave the prompt unchanged ([cache.py:9-39, 101][cache]).
- **Fact:** `graphify update <path>` re-extracts code only (no LLM). Given `changed_paths`, it re-extracts only those files and keeps the other nodes from the existing graph; otherwise it does a full re-extract through the cache ([watch.py:1558-1590][rebuild]). `watch` uses watchdog (optional) with debounce ([watch.py:2543][watch]). `hook install` adds post-commit/post-checkout git hooks and a **git merge driver** for `graph.json`, so the graph is designed to be committed ([hooks.py:935][hooks]). A per-repo flock serialises rebuilds.
- **Fact:** doc/paper/image changes need the AI-assistant `/graphify --update` (or `extract` with an LLM backend). `check-update` reports pending semantic work.
- **Fact (run):** after renaming a TS function, `update` gave a graph identical (node and edge sets) to a clean `extract` of the same tree.
- **Fact:** `to_json` refuses to write a graph smaller than the existing one unless `--force` / `GRAPHIFY_FORCE=1` ([export.py:291-300][tojson]). **Inference:** a legitimate large deletion needs a manual override, which is risky for unattended refresh.

### 8. Export formats

- **Fact:** `graph.json`, `graph.html` (interactive), Obsidian vault + `.canvas`, SVG, GraphML, Cypher for Neo4j and FalkorDB (file or `--push`), wiki markdown, D3 tree HTML, Mermaid callflow HTML, GRAPH_REPORT.md ([export.py][export], CLI help).

### 9. Extension points

- **Fact:** **No runtime plugin mechanism.** There are no `entry_points` and no user extractor registration (grep of the package). Adding a language means editing `_DISPATCH`, `collect_files`, `CODE_EXTENSIONS`, `_WATCHED_EXTENSIONS` and pyproject, then adding tests ([ARCHITECTURE.md:81-89][addlang]).
- **Fact:** There are internal seams that are still moving. `extractors/LANGUAGE_EXTRACTORS` is described as a "registry seed; wiring dispatch through it is a later step" ([extractors/__init__.py][extinit]). `resolver_registry.LanguageResolver(name, suffixes, resolve)` handles cross-file passes ([resolver_registry.py][resreg]). An ongoing migration is splitting `extract.py` one language per PR ([MIGRATION.md][migration]).
- **Fact:** These data-level seams are usable from outside without code changes: any `{nodes, edges}` JSON that passes `validate.py` can be merged (`merge-chunks`, `merge-semantic`, `merge-graphs`). `scip_ingest.py` reads a *simplified* SCIP-like JSON shape (not real SCIP protobuf) and is "Not wired to the CLI" ([scip_ingest.py:1-30][scip]). Custom LLM providers can be added via `graphify provider add`.

### 10. License, maintenance, dependency footprint

- **Fact:** Apache-2.0. It was relicensed from MIT on 2026-07-22 (commit `ba7f9ea`); earlier contributions remain MIT ([NOTICE][notice], [LICENSE-MIT][mit]). The PyPI package is `graphifyy`.
- **Fact:** Activity: first commit 2026-04-03; 2,188 commits; 498 commits in the 30 days to 2026-10-09; a release almost every day (v0.9.75–v0.9.82 between 2026-10-04 and 2026-10-09); 199 distinct authors in the last 3 months. The two top committer identities are the founder (Safi Shamsi). GitHub API on 2026-10-09: about 124.9k stars, 12.1k forks, 1,556 open issues, not archived. About 5.8k test functions across 360 test files. CI runs on Ubuntu only, Python 3.10/3.12/3.13/3.14 ([ci.yml][ci]).
- **Fact:** It is open-core. The README and the v0.9.82 CLI promote Graphify Cloud / Enterprise (cross-repo search, PR review, "always-current"), and the CLI prints a one-time Cloud tip ([README.md:31-40, 888-894][readme], [CHANGELOG 0.9.82][changelog]).
- **Fact:** Runtime is Python ≥3.10. Base dependencies: networkx, numpy, rapidfuzz, tree-sitter 0.25 + about 25 grammar wheels. Many extras: mcp/starlette, leiden (graspologic), neo4j, falkordb, pdf, office, video (faster-whisper, yt-dlp), LLM SDKs and others ([pyproject.toml][pyproject]). **Fact (measured):** a base `uv pip install graphifyy==0.9.82` gives 30 packages / about 158 MB venv; with `[mcp]` it is 58 packages / about 189 MB. `nuitka` is a dev dependency, but no standalone binary is built in CI or release workflows.
- **Inference:** The large heuristic surface (about 80k LOC of Python, per-language resolution, frequent id-scheme fixes) plus the daily release cadence means the **output format and node ids churn**, so any adapter must pin a version.

---

### What a Go or Rust harness could reuse

| Option | Verdict | Why |
|---|---|---|
| Fork or embed | **No** | Python runtime (about 160 MB venv), not a single binary; architecture assumes NetworkX in memory. (Also settled in the map.) |
| **Consume `graph.json` through an adapter** | **Yes, optional, symbol-level evidence only** | The format is documented, validated and versioned (`schema_version`), and has file+line provenance. A Go/Rust adapter could import it as *low-level evidence* (calls/imports inside a Component) and roll it up to Components the harness itself detects. Caveats: pin a `graphify_version`; handle direction via link order / `_src`/`_tgt`; ids are unstable and slug-based (match on `source_file`+`label` instead); one edge per pair loses information; no evidence of service-to-service traffic. |
| Adopt schema ideas | **Yes, selectively** | (1) an extraction-fragment contract with a validator; (2) the three-level confidence label + a **discrete** score rubric (backed by their observation that LLMs collapse continuous scores); (3) an explicit tier/origin marker, kept as a first-class field rather than inferred from shape; (4) deterministic canonical JSON (sorted keys/records) so diffs are clean; (5) cache keyed on content hash + extractor version + schema, with the LLM cache keyed on a prompt fingerprint; (6) "external stub" nodes so edges never dangle; (7) repo-tag prefixing + `local_id` on merge; (8) a token-budgeted, provenance-annotated text output for agents; (9) a git merge driver for a committed map file; (10) agent-host "always-on" instruction blocks + a pre-tool hook as a distribution channel beyond MCP. |
| Avoid | — | Simple-graph edge collapse; a free-string relation vocabulary; LLM allowed to say EXTRACTED; ids that change when unrelated files are added; no per-edge extractor/version/commit; "graph shrank" guard that needs `--force`; communities as a stand-in for architecture. |

**Inference:** Graphify's tree-sitter query logic is per-language Python code. It cannot be called from Go or Rust, but its fixtures (`tests/fixtures/`) and the issues behind its resolution fixes are a free **catalogue of edge cases** a harness extractor will meet: nested classes, CRLF, encodings, TS same-name exports, Python `TYPE_CHECKING` imports, and so on.

### Open / unknown

- **Open:** How accurate its symbol-level edges are on a real multi-service corpus. I ran only toy fixtures; the README's token-reduction numbers (for example "71.5x") were not checked.
- **Open:** Whether the internal registry migration (`LANGUAGE_EXTRACTORS`, `resolver_registry`) will turn into a public plugin API, and whether the MultiDiGraph mode will ship.
- **Open:** Whether Graphify Cloud's "cross-repo search" adds service-level edges that the OSS lacks. It is closed source and outside this review.
- **Open:** How stable `schema_version: 1` is in practice. It was introduced in 0.9.78 (2026-10-06) and has not been bumped yet, while field-level churn continues.
- **Open:** Behaviour on Windows/macOS. CI is Ubuntu-only, though the changelog shows many Windows fixes.

[arch]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/ARCHITECTURE.md
[addlang]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/ARCHITECTURE.md#L81-L89
[hiw]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/docs/how-it-works.md
[dispatch]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L7205
[gen]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L1712-L1717
[promote]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L9178-L9201
[indirect]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L9143-L9158
[recv]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L4150-L4170
[originstamp]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extract.py#L9427
[detect]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/detect.py#L47-L52
[manifest]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/manifest_ingest.py#L1-L29
[cargo]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cargo_introspect.py
[pg]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/pg_introspect.py
[mcpingest]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/mcp_ingest.py
[stubrel]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L82-L96
[genrel]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L71-L80
[origin]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L45-L68
[buildgraph]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L1149
[astwins]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L1180-L1225
[collapse]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L1530-L1600
[prefix]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/build.py#L2529-L2575
[goid]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extractors/go.py#L111
[base]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extractors/base.py
[extinit]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extractors/__init__.py
[migration]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/extractors/MIGRATION.md
[resreg]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/resolver_registry.py
[cluster]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cluster.py#L23-L127
[cache]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cache.py#L9-L101
[validate]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/validate.py#L4-L7
[schemaver]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/export.py#L176
[confdef]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/export.py#L197
[tojson]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/export.py#L286-L416
[tojsonmeta]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/export.py#L451-L460
[export]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/export.py
[mdg]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/multigraph_compat.py#L1-L10
[spec]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/skills/agents/references/extraction-spec.md
[mergecli]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cli.py#L2730-L2900
[hookguard]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cli.py#L2586-L2597
[global]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/global_graph.py#L79-L170
[xtypes]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cross_repo_types.py
[xcalls]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/cross_repo_calls.py#L151
[serve]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py
[score]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L548-L575
[bfs]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L1034
[subtext]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L1092
[reload]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L171
[tools]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L1990-L2120
[resources]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L2428-L2433
[servehttp]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L2714
[serveauth]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/serve.py#L2600
[rebuild]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/watch.py#L1558-L1590
[watch]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/watch.py#L2543
[hooks]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/hooks.py#L935
[scip]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/scip_ingest.py#L1-L30
[alwayson]: https://github.com/Graphify-Labs/graphify/tree/5b74d7d74911cf435c8f1636b6f96ea202cc6246/graphify/always_on
[pyproject]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/pyproject.toml
[notice]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/NOTICE
[mit]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/LICENSE-MIT
[contrib]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/CONTRIBUTING.md#L15
[changelog]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/CHANGELOG.md
[readme]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/README.md
[ci]: https://github.com/Graphify-Labs/graphify/blob/5b74d7d74911cf435c8f1636b6f96ea202cc6246/.github/workflows/ci.yml
