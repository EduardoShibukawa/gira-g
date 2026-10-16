---
schema: task/1
tier: standard
status: active
parent: 2026-10-16-gira-docs-deploy-brand-audit
spec: 2026-10-16-gira-docs-deploy-brand-audit@1
depends_on:
  - 2026-10-16-gira-docs-deploy-brand-audit-02-brand-identity
---

# Docs deploy 03: Integrated NFR verification & Gira dogfooding audit

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-docs-deploy-brand-audit)

## Objective

Run the integrated NFR verification and `gira` dogfooding audit across
the repository after children 01 and 02 land: verify `gira skills
manifest` and `gira skills check` (against the 6-skill embedded catalog
in `internal/skills/catalog`), execute `gira index`, `gira locate`,
`gira slice --tier summary`, `gira explain`, `gira impact --from-diff
origin/main..HEAD`, and `gira diff` to confirm architectural map and
envelope integrity, and verify `mise run docs:build`, `mise exec -- go
test -race ./...`, the canonical quarantine grep, and `node
tools/validate/validate.js`.

In scope: Dogfooding audit execution, verification evidence recording,
and NFR gate validation.
Out of scope: Modifying protected `.agents/skills/work` or
`.claude/skills/work` files.

## Acceptance criteria

1. AC1 = SC5: `bin/gira` executes `index`, `locate`, `slice --tier
   summary`, `explain`, `impact --from-diff origin/main..HEAD`, `diff`,
   `skills manifest --format json`, and `skills check` (verified against
   the 6-skill catalog) with exit `0` and valid JSON/Markdown envelopes.
2. AC2 = SC6: `mise run docs:build` exits `0` (`website/build/index.html`
   and `website/build/.nojekyll` present), `mise exec -- go test -race
   ./...` exits `0`, canonical quarantine grep on `website/` returns
   empty, and `node tools/validate/validate.js` exits `0`.

## Constraints and budget

- Subagents: 1 (delegated `self` subagent for isolated execution of the Gira dogfooding audit and NFR verification suite per `artifact-contracts.md` §3.5).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| B2 | mise run docs:build (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| AC1 | gira dogfooding audit (index/locate/slice/explain/impact/diff/skills) | not run | - | pending execution |
| AC2 | NFR gates (docs:build, go test -race, quarantine grep, validator) | not run | - | pending execution |
| R1 | full regression suite + golangci-lint --new-from-rev=origin/main | not run | - | pending execution |
| V1 | audit output and git diff review against AC1–AC2 | not run | - | pending execution |

## Plan

1. Build `bin/gira` (`mise exec -- go build -o bin/gira ./cmd/gira`) and run `bin/gira skills manifest --format json` plus `bin/gira skills install` / `bin/gira skills check` in an isolated scratch directory for `--host agents`, `--host claude`, and `--host cursor`.
2. Index baseline (`origin/main`) and current `HEAD` of `gira-g` via `bin/gira index`, then run `locate`, `slice --tier summary`, `explain`, `impact --from-diff origin/main..HEAD`, and `diff`.
3. Run `mise run docs:build`, `mise exec -- go test -race ./...`, canonical quarantine grep on `website/`, and `mise exec -- golangci-lint run --new-from-rev=origin/main`.
4. Record all verification evidence in `tasks/2026-10-16-gira-docs-deploy-brand-audit-03-nfr-gira-audit.md` and validate with `node tools/validate/validate.js`.
