# Prior art: codebase-memory-mcp

[codebase-memory-mcp](https://github.com/DeusData/codebase-memory-mcp): MIT. Local structural graph (C + tree-sitter 162 langs, SQLite WAL, loopback Lens UI). v0.11.0 `24d0251` (2026-10-13).

## How we use it

Studied, never depended on. Complements Graphify/UA without duplicating them.

## USE / COPY / NORTH STAR

- **USE:** nothing.
- **COPY:** `CALLS` (proven) vs `CALL_REFERENCE` vs `USAGE` (unproven — honesty) taxonomy; `check_index_coverage` + "coverage shadow" for query-time staleness; Scout/Verify/Auditor tiers with allowlists; structured errors (`retryable`, `serving_index_preserved`); budgets (`max_output_tokens` 4B/token, pagination, `has_more` + cursors); artifact caution (`merge=ours`, LFS costs).
- **NORTH STAR:** fast+broad ≠ evidence-backed; "memory" = durable graph + ADR, no per-edge correction; local-cache default is the opposite of Gira's commit-the-index.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** 17 MCP tools (`index_repository`, `trace_path`, `detect_changes`, read-only openCypher `query_graph`, `compare_graphs`, `manage_adr` 8KB cap…); HTTP route↔callsite confidence; `SIMILAR_TO` MinHash, `SEMANTICALLY_RELATED ≥0.80` without locators.
- **Fact:** path-based qualified-name IDs → rename-unstable; no `origin/discovered_at/why_locator`; no symbol-loss gate; watcher covers the open session's project only.
- **Open:** exact freshness schema; does `USAGE` always carry file/line?
