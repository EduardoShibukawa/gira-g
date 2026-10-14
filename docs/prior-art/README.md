# Prior art — index

Studied references for Gira-G. Policy: **studied, never depended on**.

- Run by hand only, in a throwaway environment outside the toolchain (`mise.toml`, CI, hooks, `AGENTS.md`).
- Their output is never committed (see `.gitignore`); only normalized score reports with versions/digests/SHAs (ADR 0012).
- Reimplement ideas, don't copy. Verbatim copies keep the license header and get a `THIRD_PARTY_NOTICES` entry.
- Each file carries a dated study section with findings marked **Fact** (code/docs/run with link), **Inference** (our judgement), **Open** (unresolved).
- Benchmarks against Graphify/Understand Anything run in quarantine only (ADR 0012); they never feed the gate (ADR 0007).

## USE / COPY / NORTH STAR

- **USE** = use the tool/service as-is via adapter/quarantine/container, no fork, no runtime dependency.
- **COPY** = steal the *idea* and reimplement it our way in Go. Never paste their code; verbatim snippets only as a rare exception with header + `THIRD_PARTY_NOTICES`.
- **NORTH STAR** = principle/direction guiding a decision, no reuse at all.

One item can have all three (e.g. USE the format, COPY the mechanism, follow the NORTH STAR).

## Map

| Reference | Status | License | Studied pin | USE | COPY | NORTH STAR |
|---|---|---|---|---|---|---|
| [graphify](graphify.md) | core | Apache-2.0 | v0.9.82 `5b74d7d` (2026-10-09) | `graph.json` via optional adapter | validated fragment, discrete rubric, canonical JSON, hash cache, external stub, repo prefix+local_id, budgeted output, always-on+hook | input adapter, never fork; one edge per pair loses evidence |
| [understand-anything](understand-anything.md) | core | MIT | plugin 2.9.7 `ffa2f0a` (2026-10-09) | nothing (ideas only) | fingerprints SKIP/PARTIAL/ARCH/FULL, symbol-loss gate, staleness model, committed map+viewer, tour as IDs, tolerant ingestion, nearest-manifest resolution | deterministic first, LLM second; no edge without evidence |
| [posthog](posthog.md) | strong | MIT (mostly) + proprietary `ee/` | `master` ~67k commits (2026-10-14, SHA to pin) | hard extraction fixture in quarantine | facade+contracts, `register_routes`, `mcp/tools.yaml` codegen, CLI/exec+feature filters, `db_routing.yaml`/`topics.txt`, crossings baselines | Product≈Component, explicit wiring; LLM-everywhere as counter-example |
| [groma](groma.md) | strong | MIT | v0.6.6 `9b4ee9c` (2026-10-13) | nothing (TS/Bun incompatible) | Derived/Authored/Draft precedence, scoreless abstention, P1-P5 rules + 7-step HTTP join, curation-preserving rescan, `Include/Exclude/Uncertain` eval | scanners architecture-unaware, core maps; inventory ≠ Macro Map |
| [archify](archify.md) | studied, not a Layer-0 arm | MIT | v3.1.0 `c21c31f` (2026-10-13) | nothing (no deterministic extractor for NAG) | `revision 40-hex + sources[≤3]` contract, git-object verification, delivery taxonomy, ID-paired `compare`, reach/route/lens deep-links | reachability ≠ blast radius; LLM-authors + checker-validates-shape |
| [codebase-memory](codebase-memory.md) | contrast | MIT | v0.11.0 `24d0251` (2026-10-13) | nothing | `CALLS/CALL_REFERENCE/USAGE` taxonomy, `check_index_coverage`, Scout/Verify/Auditor tiers, retryable errors, byte-ceiling budgets + cursors | fast+broad ≠ evidence-backed; path-based IDs unstable |
| [token-optimizer](token-optimizer.md) | constructive+cautionary | MIT | 7.5.0 `e8013a7` (2026-10-13) | nothing (no hard envelope) | tiktoken-only honesty, `net=gross−expansions` with quarantine, single-source tool defs, deferred `ToolSearch`, zero-turn refusal | measured ≠ estimated; substitution ≠ reduction; 90 tools = sprawl |
| [headroom](headroom.md) | narrow optional | Apache-2.0 | `f1bc260` (2026-10-13) | nothing | ContentRouter + `transforms[]`, reversible CCR, cache-vs-token mode, holdout-measured savings | compression ⊥ mapping; LLM memory without locator violates never-guess |
| [context-mode](context-mode.md) | narrow | Elastic-2.0 (never vendor) | v1.0.169 `dbec980` (2026-10-14) | nothing (ELv2) | hook-enforced routing (measured), `doctor/upgrade/purge`, batched `queries:[]`, `>100KB→pointer` overflow, TTL | opt-in slices, not mandatory interception |
| [docusaurus](docusaurus.md) | optional | MIT (+CC-BY docs) | v3.10.2, main ~6.9k commits (2026-10-14) | nothing (JS stack) | serializable data/render split, sidebar-as-IDs, `current vs latest` versioning, preset + wrap + Safe/Unsafe/Forbidden | LLM-free viewer; no extractor value |
| [unlazy](unlazy.md) | harness | MIT | 2.1.0 `1667149` (2026-09-03) | gate contract as spec (not the JS) | success-only marker + digest, stale-fail-closed, `ABANDON`=handoff, `gate-lint`, token-economy | gate only when silent-incompleteness cost justifies it |
| [agent-review](agent-review.md) | methodology, partial adopt | MIT | v1.0.0 `031841c` (2026-10-11) + article 2026-10-13 | templates as spec | `ARTIFACT-VERIFIED` closing line, control total + `max(20,5%)`, claim taxonomy, findings register, independent fix-check | S3 fresh-eyes on heavy only; personas as fixed templates |
| Strata | rejected | MIT (engine) | v0.1.44.1 (2026-10-13) | nothing | MCP hygiene only (stdlib-only, allowlist, `connect_info`) | generative runtime ≠ evidence |
| ponytail | rejected | MIT | v5.2.0 (2026-10-14) | nothing | `If-we-skip/Not-checked ≈ unknown` footer and `shortcut:` ledger only | prompt-only is not an impact set |

Rejected items stay in this table only (no file of their own).
