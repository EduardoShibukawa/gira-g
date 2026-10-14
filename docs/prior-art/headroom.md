# Prior art: headroom

[headroom](https://github.com/headroomlabs-ai/headroom): Apache-2.0. Context-compression layer (Python CLI+lib, Rust PyO3 core, SQLite+HNSW+FTS5, ONNX Kompress fallback). HEAD `f1bc260` (2026-10-13).

## How we use it

Studied, never depended on. Orthogonal complement (compression ⊥ mapping).

## USE / COPY / NORTH STAR

- **USE:** nothing.
- **COPY:** deterministic per-type ContentRouter + fails-open + `transforms[]` audit + skip-set/cache; reversible CCR (hash-indexed SQLite + TTL + marker + retrieve); cache-vs-token mode + `min_tokens_to_compress` + canonical `server.json` + aggregated `stats`; holdout-measured savings; Scanner/Analyzer/Writer adapter shape for NAG.
- **NORTH STAR:** LLM-emitted memory (`<memory>` block) and lossy Kompress without locator = never-guess violated; prompt-append + budget-clamp + hot-sync break determinism/quarantine.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** 3 MCP tools (`headroom_compress/retrieve/stats`), stdio + Streamable HTTP; no query/path/impact; memory `supersede()` chain; seeded benchmarks (`index_proof_table --seed`, `bench_latency`, GSM8K/TruthfulQA/SQuAD/BFCL) but N=100.
- **Open:** TOIN weights; Kompress keep-decisions without provenance.
