# Prior art: groma.md

[groma.md](https://github.com/MrLesk/groma.md): MIT. "Architecture as OKF Markdown in Git + one C4 map". v0.6.6 `9b4ee9c` (2026-10-13). TS/Bun, 12 stable langs + 4 experimental via `groma scanner add`.

## How we use it

Studied, never depended on. TS/Bun stack is incompatible with the Go binary.

## USE / COPY / NORTH STAR

- **USE:** nothing (incompatible); specs as reference only.
- **COPY:** Derived/Authored/Draft sections with precedence (→ declared/extracted/inferred); scoreless abstention ("decline, no threshold"); P1-P5 rules + 7-step HTTP endpoint/request join (best evidence-join spec seen); curation-preserving rescan (`groma.code` only); draft-at-final-path + `accept` after match; `Include/Exclude-as-support/Uncertain` eval + `I/(I+X)`; PR architecture-diff Action.
- **NORTH STAR:** scanners architecture-unaware, core maps; inventory (3200 elements for 3164 files in the OpenClaw case) ≠ Macro Map — aggregate, don't flood.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** tree `groma/: actors/, externals/, systems/<s>/containers/<c>/components/<co>.md + drafts/, flows/`; OKF v0.2 (`docs/component-markdown.md`); "Scanners are language-aware and architecture-unaware… Core maps those facts to ownership" (`docs/relationship-inference.md`).
- **Fact:** no MCP stdio (CLI + `agent-instructions` + AGENTS.md/CLAUDE.md block); flows with steps that must resolve to a directed relationship; no token budgets (only `--max-count/--skip`).
- **Inference:** highest overlap, opposite philosophy on evidence discipline; manual curation covers cohesion Gira wants deterministic.
- **Open:** MCP plans? Jelly-vs-compiler eval detail.
