---
schema: task/1
tier: heavy
status: accepted
spec: gira-envelopes-hygiene@2
verified_at: "3369500"
---

# Envelopes & hygiene (E3)

## Routing and approvals

- tier: heavy (H1: product calls on ceilings/overflow/CI; H3: spans MCP, CLI, slices, tokens, skills, gate, CI, bench; H5: 5 children, multi-session)
- confirmed: heavy by owner (2026-10-14, Q1–Q7 answers)
- decided: Q1 envelopes opt-in only; Q2 gate-lint warn-only in CI; Q3 validator/gate-lint/floor-guard ubuntu-only; Q4 opt-in hermetic R/P floor guard; Q5 refuse with budget_exceeded; Q6 hygiene extras in E3; Q7 F2 deferred
- decided: ADR 0012 takes precedence over the v1 CI-bench clause; the quarantined comparative module is not run from the product CI (owner 2026-10-14, "Honrar o ADR: tirar bench do CI")
- approved: spec v1 by owner (2026-10-14)
- approved: spec v2 by owner (2026-10-14, ADR 0012 precedence over the CI bench clause; opt-in hermetic R/P floor guard)
- spec-version: v1 → v2; rerun: SC7 (CI hygiene) only
- accepted: by owner (2026-10-14, "Corrigir as falhas agora")

## Objective

Every agent-facing output has a hard, deterministic, honestly labeled envelope. Over-ceiling output is refused with `budget_exceeded` + `escalate_hint`. One tool registry; structured errors. CI hygiene wired (validator, gate-lint warn-only, floor guard) on ubuntu-latest; ADR 0012 keeps the quarantined comparative module out of the product CI. Opt-in floor guard for hermetic R/P. Hygiene fixes: gitignore results.tsv/run.log, fix README.md:22 link, quarantine skips .claude/worktrees/, fix validator §6.1 false positive in non-linear history. E2 F2 stays deferred.

In scope:
- spec `gira-envelopes-hygiene` v2;
- the children listed in the Plan;
- integrated verification of SC1–SC10.

Out of scope:
- E2 joins (F2 deferred), E4 human surface, E5 skills distribution, E6 docs site;
- ground-truth changes, `reviewed: true`;
- any model client (ADR 0001), compression/summarization;
- push or release.

## Acceptance criteria

- AC1 = SC1–SC10 of `specs/gira-envelopes-hygiene.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before E3 children | pass | 524f55d | root 16 pkgs ok, bench ok, validator pass (original baseline) |
| B2 | root + bench + validator at the E3 merge, before the fix pass | fail | 212e7f5 | pre-existing/undetected: `budget.estimated`≠output, locate object, stale goldens, and the §6.1/quarantine/floor-guard gaps below; fixed in 3369500 |
| SC1 | `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -count=1` | pass | 3369500 | all 12 MCP tools carry the envelope |
| SC2 | `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -count=1` | pass | 3369500 | same index + budget is byte-identical |
| SC3 | `go test ./internal/tokens ./internal/slices -run SCEnvelopeHonesty -count=1` | pass | 3369500 | estimator=bytes/4, measured=false; Estimated now equals the emitted bytes (F4) |
| SC4 | `go test ./internal/mcp -run SCTransportCeiling -count=1` | pass | 3369500 | global ceiling refuses with budget_exceeded + escalate_hint |
| SC5 | `go test ./internal/mcp ./internal/skills ./cmd/gira -run SCStructuredErrors -count=1` | pass | 3369500 | {code, message, retryable, suggestions}; gira-query reads code first |
| SC6 | `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -count=1` | pass | 3369500 | one registry at internal/tools |
| SC7 | workflow greps for validate.js, gate-lint, `E3_FLOOR_GUARD`; `go test ./internal/gate -run SCGateLintStrictness -count=1`; `E3_FLOOR_GUARD=1 go test ./internal/eval -run FloorGuard -count=1` | pass | 3369500 | validator + gate-lint + real floor guard on ubuntu; quarantined module out of CI (ADR 0012) |
| SC8 | EndToEnd published/perturbed R/P; bench SC15 ≥ 90% both tours | pass | 3369500 | published 82.4/100.0, perturbed 81.2/100.0; SC15 95.0% / 100.0% |
| SC9 | `git diff --exit-code 9ba80a6 -- fixtures/ testdata/`; gira-query goldens regenerated; `"reviewed": false` | pass | 3369500 | ground-truth/corpus frozen; 3 budgeted goldens regenerated for F2 (F6) |
| SC10 | `go test -race ./...` root and bench/; validator exit 0 | pass | 3369500 | root + bench green, validator pass |
| AC1 | SC1–SC10 integrated | pass | 3369500 | every row above passes at the integrated head |
| R1 | `go test -race ./...` (root) | pass | 3369500 | all packages green |
| R2 | `go test -race ./...` (bench/) | pass | 3369500 | layer0/1/2/quarantine green |
| R3 | `node tools/validate/validate.js` | pass | 3369500 | manifest-integrity + artifact-records pass |
| V1 | integrated diff read + fresh review | pass | 3369500 | fixes F1–F8 below; no ground-truth/model/paid/push |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-envelope-all-tools`: R1, R2; SC1, SC2. Envelope on get_tour, diff_maps, explain_evidence, locate_boundary + CLI tour/flows/diff/explain/locate; deterministic drop order.
2. `-02-envelope-honesty-and-ceiling`: R3, R4; SC3, SC4. Honest budget labels (estimator, measured); global MCP ceiling; refuse with budget_exceeded + escalate_hint. Depends on 01.
3. `-03-structured-errors`: R5; SC5. Structured error `{code, message, retryable, suggestions}` on MCP and CLI `--format json`; gira-query reads code first. Depends on 01.
4. `-04-single-registry`: R6; SC6. Single tool registry (or set-equality test); each tool declares `enveloped`. Depends on 01.
5. `-05-ci-hygiene`: R7, R8; SC7. CI: validator, gate-lint warn-only, opt-in hermetic R/P floor guard on ubuntu-latest (ADR 0012 keeps the quarantined module out). Hygiene: gitignore, README link, quarantine skip .claude/worktrees, validator §6.1 fix. Needs owner `approved:` before commit. Depends on 01–04.

SC8–SC10 at every child. Rollback: `git revert` per child.

## Review

- [F1] the §6.1 non-linear false positive was never fixed: `validate.js` was byte-identical to the bundle and the "verifying" test only `t.Log`ged → real parent-aware deletion detection + a regression test that fails without it (3369500)
- [F2] the E3 CI bench job and `internal/mcp/ci_hygiene_test.go` broke SC12/ADR 0012 (quarantine passed at E2, broke at E3) → bench removed from the product CI; `bench/` refs removed from the Go test; the `.claude/worktrees` skip is asserted in the quarantined module (3369500)
- [F3] the opt-in floor guard was a no-op (`gira eval --floor-guard` did not exist; CI used `|| true`) → a real `E3_FLOOR_GUARD=1 go test ./internal/eval -run FloorGuard` (3369500)
- [F4] `budget.estimated` did not match `tokens.Estimate(output)`: the no-truncation path settled without the escalate hint → `settleHint` includes the hint, so Estimated matches the emitted bytes (3369500)
- [F5] `locate_boundary` without a budget returned an object, breaking Q1 byte-identity and gira-query MCP/CLI parity → it returns the bare array; only a budgeted call gets the envelope (3369500)
- [F6] the 3 budgeted gira-query goldens were stale against F2 (`estimator`/`measured`) → regenerated; SC9 scoped to the frozen ground-truth/corpus (3369500, spec v2)
- [F7] the dogfood e2e poisoned the fixture commit SHA from a git hook's leaked `GIT_*` environment → commit identity pinned; the pre-commit gate is green (3369500)
- [F8] formatting drift from earlier `--no-verify` commits plus an embedded `.claude/worktrees` gitlink → gofmt applied; the worktree is untracked and ignored (3369500)

## Checkpoint

- as of: 2026-10-14, main, head 3369500, worktree clean
- done: children 01–05 verified at @2; spec v2 approved; §6.1/SC12/floor-guard/budget/locate/goldens fixed in 3369500; gofmt + worktree hygiene in 3369500; root + bench + validator green
- next: open PR / push (owner) — no push yet

## Outcome

E3 Envelopes & Hygiene delivered and accepted at spec v2: all 5 children verified, plus the fix pass that made the delivery honest and green.

- Child 01 (SC1, SC2): Envelope coverage on all 12 MCP tools + CLI; deterministic drop order. `locate_boundary` keeps the bare array without a budget (Q1).
- Child 02 (SC3, SC4): Honest budget labels (estimator="bytes/4", measured=false); global MCP ceiling with budget_exceeded + escalate_hint. The no-truncation estimate now settles with the escalate hint, so `estimated` equals the emitted bytes.
- Child 03 (SC5): Structured errors {code, message, retryable, suggestions} on MCP; gira-query reads code first.
- Child 04 (SC6): Single tool registry at internal/tools/registry.go; enveloped field on all 12 tools.
- Child 05 (SC7): CI hygiene — validator, gate-lint warn-only, and a real opt-in hermetic R/P floor guard on ubuntu-latest. ADR 0012 keeps the quarantined comparative module out of the product CI; its own tests (layer0/1/2 + quarantine) stay green. Hygiene fixes: gitignore, README link, quarantine skip .claude/worktrees/, and the validator §6.1 non-linear fix with a regression guard.

Invariants preserved: hermetic published R82.4/P100.0, perturbed R81.2/P100.0; SC15 ≥ 90% (95.0% / 100.0%); ground-truth/corpus frozen (`fixtures/`, `testdata/`); unbudgeted calls byte-identical; validator passes. The 3 budgeted gira-query goldens were regenerated for F2 (see SC9/F6). Ready for PR (no push yet).