# Prior art: context-mode

[context-mode](https://github.com/mksglu/context-mode): Elastic-2.0 (**never vendor, reimplement only**). Window optimizer (TS, MCP SDK, SQLite FTS5). v1.0.169 `dbec980` (2026-10-14).

## How we use it

Studied, never depended on (license bars hosted-service use).

## USE / COPY / NORTH STAR

- **USE:** nothing (ELv2).
- **COPY:** hook-enforced routing with measured delta (~98% vs ~60% file-only); `doctor` + redacted debug bundle + `upgrade/purge` + `heal-*`; batched `queries:[]` + `source` scoping + progressive throttling; `>100KB → FTS5 pointer` overflow; TTL cache; "don't pollute the repo" (runtime injection > auto-written files).
- **NORTH STAR:** opt-in slices, not mandatory interception (`BLOCKED curl` maximalism); byte savings ≠ task success.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** 11 tools (`ctx_execute/batch/file`, `ctx_index/search/fetch_and_index`, `ctx_stats/doctor/upgrade/purge/insight`); 12 sandboxed runtimes (stdout only); FTS5 Porter+trigram RRF; `skills/context-mode/SKILL.md` + `configs/<host>` via `version-sync`.
- **Open:** `docs/platform-support.md`, exact thresholds, OpenCode adapter before copying numbers.
