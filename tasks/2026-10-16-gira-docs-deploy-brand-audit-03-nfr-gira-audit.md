---
schema: task/1
tier: standard
status: verified
verified_at: "4aef216"
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
| AC1 | gira dogfooding audit (index/locate/slice/explain/impact/diff/skills) | pass | 4aef216 | skills manifest & check pass across agents/claude/cursor; index/locate/slice/explain/impact/diff verified |
| AC2 | NFR gates (docs:build, go test -race, quarantine grep, validator) | pass | 4aef216 | docs:build exit 0 (index.html + .nojekyll); quarantine grep exit 1 (0 matches); go test -race exit 0; validator exit 0 |
| R1 | full regression suite + golangci-lint --new-from-rev=origin/main | pass | 4aef216 | go test -race exit 0 across all pkgs; golangci-lint zero new issues |
| V1 | audit output and git diff review against AC1–AC2 | pass | 4aef216 | diff confirms 7 added nodes, 30 added edges, 0 contract breaks |

## Plan

1. Build `bin/gira` (`mise exec -- go build -o bin/gira ./cmd/gira`) and run `bin/gira skills manifest --format json` plus `bin/gira skills install` / `bin/gira skills check` in an isolated scratch directory for `--host agents`, `--host claude`, and `--host cursor`.
2. Index baseline (`origin/main`) and current `HEAD` of `gira-g` via `bin/gira index`, then run `locate`, `slice --tier summary`, `explain`, `impact --from-diff origin/main..HEAD`, and `diff`.
3. Run `mise run docs:build`, `mise exec -- go test -race ./...`, canonical quarantine grep on `website/`, and `mise exec -- golangci-lint run --new-from-rev=origin/main`.
4. Record all verification evidence in `tasks/2026-10-16-gira-docs-deploy-brand-audit-03-nfr-gira-audit.md` and validate with `node tools/validate/validate.js`.

## Outcome

- Executed full dogfooding audit of `gira-g` at commit `4aef216`:
  - `bin/gira skills manifest --format json` exited `0`, validating the 6-skill embedded catalog (`gira-arch-diff`, `gira-context`, `gira-explain-component`, `gira-impact-briefing`, `gira-query`, `gira-tour`).
  - `bin/gira skills install` and `bin/gira skills check` executed cleanly in an isolated test host directory for `--host agents`, `--host claude`, and `--host cursor`, confirming `in_sync: true` and `ok: true` on all 3 hosts.
  - Indexed baseline (`origin/main` at `41781d9`) and current HEAD (`4aef216`): baseline indexed 300 nodes and 703 relationships; HEAD indexed 307 nodes and 733 relationships.
  - Tested `locate docs-deploy`, `slice "policy:2026-10-16-gira-docs-deploy-brand-audit" --tier summary`, `explain --node "service:job:docs-deploy"`, `impact --from-diff origin/main..HEAD`, and `diff` between base and head indexes: confirmed 7 added nodes, 30 added edges, 0 removed edges, 0 changed contracts, 0 newly unsupported boundaries, and zero broken envelopes.
- Verified all integrated NFR gates:
  - `mise run docs:build` exited `0` with `baseUrl: '/gira-g/'` and produced `website/build/index.html` and `website/build/.nojekyll`.
  - Canonical case-insensitive quarantine grep on `website/` returned exit `1` (zero matches).
  - `mise exec -- go test -race ./...` exited `0` across all Go packages.
  - `mise exec -- golangci-lint run --new-from-rev=origin/main` exited `0` with zero new issues.
  - `node tools/validate/validate.js` exited `0`.

