---
schema: task/1
tier: heavy
status: active
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# Envelopes & hygiene (E3)

## Routing and approvals

- tier: heavy (H1: product calls on ceilings/overflow/CI; H3: spans MCP, CLI, slices, tokens, skills, gate, CI, bench; H5: 5 children, multi-session)
- confirmed: heavy by owner (2026-10-14, Q1–Q7 answers)
- decided: Q1 envelopes opt-in only; Q2 gate-lint warn-only in CI; Q3 bench+validator ubuntu-only; Q4 opt-in floor guard; Q5 refuse with budget_exceeded; Q6 hygiene extras in E3; Q7 F2 deferred
- approved: spec v1 by owner (2026-10-14)

## Objective

Every agent-facing output has a hard, deterministic, honestly labeled envelope. Over-ceiling output is refused with `budget_exceeded` + `escalate_hint`. One tool registry; structured errors. CI hygiene wired (bench race, validator, gate-lint warn-only) on ubuntu-latest. Opt-in floor guard for hermetic R/P and SC15 ≥ 90%. Hygiene fixes: gitignore results.tsv/run.log, fix README.md:22 link, bench/quarantine skips .claude/worktrees/, fix validator §6.1 false positive in non-linear history. E2 F2 stays deferred.

In scope:
- spec `gira-envelopes-hygiene` v1;
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
| B1 | `go test -race ./...` (root) + bench + validator before children | pass | 524f55d | root 16 pkgs ok, bench ok, validator pass (before E3 changes) |
| SC1 | `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -count=1` | not run | | |
| SC2 | `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -count=1` | not run | | |
| SC3 | `go test ./internal/tokens ./internal/slices -run SCEnvelopeHonesty -count=1` | not run | | |
| SC4 | `go test ./internal/mcp -run SCTransportCeiling -count=1` | not run | | |
| SC5 | `go test ./internal/mcp ./internal/skills ./cmd/gira -run SCStructuredErrors -count=1` | not run | | |
| SC6 | `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -count=1` | not run | | |
| SC7 | workflow greps for validate.js, gate-lint, `working-directory: bench`; `go test ./internal/gate -run SCGateLintStrictness -count=1` | not run | | |
| SC8 | EndToEnd published/perturbed R/P unchanged; bench SC15 ≥ 90% both tours | not run | | |
| SC9 | `git diff --exit-code 524f55d -- fixtures/ testdata/ cmd/gira/testdata/gira-query/`; `"reviewed": false` | not run | | |
| SC10 | `go test -race ./...` root and bench/; validator exit 0 | not run | | |
| AC1 | SC1–SC10 integrated | not run | | |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-envelope-all-tools`: R1, R2; SC1, SC2. Envelope on get_tour, diff_maps, explain_evidence, locate_boundary + CLI tour/flows/diff/explain/locate; deterministic drop order.
2. `-02-envelope-honesty-and-ceiling`: R3, R4; SC3, SC4. Honest budget labels (estimator, measured); global MCP ceiling; refuse with budget_exceeded + escalate_hint. Depends on 01.
3. `-03-structured-errors`: R5; SC5. Structured error `{code, message, retryable, suggestions}` on MCP and CLI `--format json`; gira-query reads code first. Depends on 01.
4. `-04-single-registry`: R6; SC6. Single tool registry (or set-equality test); each tool declares `enveloped`. Depends on 01.
5. `-05-ci-hygiene`: R7, R8; SC7. CI: bench race, validator, gate-lint warn-only on ubuntu-latest; opt-in floor guard (hermetic R/P, SC15 ≥ 90%). Hygiene: gitignore, README link, bench/quarantine skip .claude/worktrees, validator §6.1 fix. Needs owner `approved:` before commit. Depends on 01–04.

SC8–SC10 at every child. Rollback: `git revert` per child. Branch linear (cherry-pick) until §6.1 fix lands.

## Review

## Checkpoint

- as of: 2026-10-14, branch e3-envelopes-hygiene, head d20a85d, worktree clean
- done: spec approved; parent active; children 01-03 verified; children 04-05 todo
- next: child 04 (single registry)

## Outcome