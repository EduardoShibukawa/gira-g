# Prior art: token-optimizer-mcp

[token-optimizer-mcp](https://github.com/ooples/token-optimizer-mcp): MIT. MCP stdio server (~90 `smart_*` tools) + `:3100` dashboard. 7.5.0 `e8013a7` (2026-10-13).

## How we use it

Studied, never depended on. Most relevant to SC6 on the **measurement** side; complements Graphify.

## USE / COPY / NORTH STAR

- **USE:** the `ToolSearch("smart_read smart_grep smart_glob")` deferred-tool pattern; nothing else (no hard envelope).
- **COPY:** tiktoken-only honesty (killed `len/4` with error table: +130% whitespace, +14.7% TS, −75% base64); `net=gross−expansions` accounting with quarantine of modeled claims; single-source `TOOL_DEFINITIONS`→validation; zero-turn refusal; disclosure-must-never-fail; `liftTextPart` before measuring.
- **NORTH STAR:** measured ≠ estimated; substitution (spill-to-disk) ≠ reduction; 90 tools = sprawl (its own server warns); opt-out semantic harvest violates ADR 0001 — pin truncation-only + evidence-derived.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** `TokenCounter` tiktoken + 8KB slices + memo (`src/core/token-counter.ts`); `docs/TOKEN_ACCOUNTING.md`: observed vs gross vs expansion-debit vs net; reduction only with schema v2 + SHA-256 of both payloads.
- **Fact:** stdio via SDK; `expand{ref,section,claim,anchor}`; ad-hoc errors, no codes (gap vs `index_missing/stale/node_not_found/budget_exceeded`); no 500/1500/3500 budgets, no dropped-counts/escalate-hint.
- **Fact:** `smart-summarization` is extractive-only (no LM); head-40/tail-60 truncator; model-based harvest opt-out via `TOKEN_OPTIMIZER_HARVEST`.
- **Open:** `intelligent-assistant`/`recommendation-engine` implementation unverified.
