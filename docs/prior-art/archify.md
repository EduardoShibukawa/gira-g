# Prior art: archify

[archify](https://github.com/tt-a1i/archify): MIT. Skill + zero-dep Node renderer: LLM-authored typed JSON (10 types) → validated standalone interactive HTML. v3.1.0 `c21c31f` (2026-10-13).

## How we use it

Studied, never depended on; **not** a Layer-0 arm (no deterministic extractor for a NAG adapter; wrapping its LLM pipeline would break determinism + ADR 0012).

## USE / COPY / NORTH STAR

- **USE:** nothing in the pipeline.
- **COPY:** `meta.repository{url, 40-hex revision}` + `sources[≤3]{path,line,end_line}` contract; git-object verification (origin/commit/blob/lines, dirty tree excluded); delivery taxonomy (deterministic receipt vs browser vs captures vs perceptual — never upgrade `skipped`); ID-paired `compare` with classes (semantic/evidence/scope/topology/geometry/provenance/presentation); reach/route/lens deep-links + Semantic Passport.
- **NORTH STAR:** reachability ≠ blast radius (per its DESIGN.md — cite in tours/flows); LLM-authors + checker-validates-shape passes geometry while hallucinating topology — extractor-owns-facts in Gira.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** 10 IRs with JSON-Schema `additionalProperties:false` (`archify/schemas/README.md`); `architecture`: 7 visual types, `from/to` edges + `variant` + free-text label — no verbs, no Contract node, sourceless edges.
- **Fact:** author→`finalize/validate/deliver` (atomic, SHA-256 receipt)→`check --require-provenance`→`browser-check`→`visual-check`→human review; `compare` with `MOVE FROM` phantoms.
- **Fact:** skill distribution (`npx skills add`), no MCP server; `ordinary-model-floor` benchmark measures first-pass usability, not P/R.
- **Open:** full `--help` not fetched; renderer dirs not fully enumerated.
