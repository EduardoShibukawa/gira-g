# Prior art: unlazy

[unlazy](https://github.com/Leonxlnx/unlazy): MIT. Anti-laziness skill (SKILL.md + Node + gate ledgers + Stop hook). 2.1.0 `1667149` (2026-09-03).

## How we use it

Studied; port the **contract**, not the JS.

## USE / COPY / NORTH STAR

- **USE:** gate contract as spec for skill/map verification.
- **COPY:** success-only marker + exit-0-and-match + SHA-256 digest (never raw logs); stale-on-drift + `--status`/`--reverify` split; `ABANDON` as qualified handoff (never `ALL MET`); `gate-lint --strict`; `.unlazy/<scope>` scopes + `OWNS:` leases; out-of-repo approval + fail-closed.
- **NORTH STAR:** gate only when silent-incompleteness cost justifies it; Depth Tree is decomposition, not a scaling law (v1 retracted).

## Study 1 — 2026-10-14 (subagent)

- **Fact:** gate requires exit 0 **and** `EXPECT:` match; evidence = digest of `CHECK:/EXPECT:/CWD:` + exit + fingerprint; 1 MiB cap; approval in `~/.unlazy/approved` outside the repo; solo/orchestrated/parallel modes + `token-economy.md`.
- **Fact:** research honesty (cited papers with caveats, internal ratio retracted in `research/validation-protocol.md`).
- **Open:** real false-completion reduction vs token/wall-clock overhead on real repos.
