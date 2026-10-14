---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-01-l1-artifact-verified]
verified_at: "a389d21"
---

# L2 control-total and seeded sampling (F2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Each run emits `control:{total, matched, mismatched, sampled_n, seed}` with the seeded `max(20,5%)` rule. Blind re-derivation on the sample matches 100% or reports an unplugged diff — never plug, never drop, never force-match. Frozen digests compared before reconciliation.

In scope: SC2, test-first. Failing test first: a forced-match/plug attempt must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC2
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before change | pass | 4f8261f | 01 verified; root/bench/validator green at 4f8261f |
| AC1 | `go test ./internal/eval -run SCL2ControlSample -v` | pass | a389d21 | re-verified after F3/F4 fix batch: plugs rejected both sides, ControlTotal.Validate, CLI seeded sample+blind round, MCP consistency gate |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | a389d21 | root/bench green, validator pass; R82.4/P90.3 unchanged, no extractor change |
| R1 | `go test ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | a389d21 | touched pkgs pass; `gira eval --control-total --seed` runs genuine blind round + closing line |
| V1 | diff read against AC1–AC2 + spec F2 | pass | a389d21 | blind round over frozen inputs, Validate before emit; no existing meaning change |

## Constraints and budget

- Additive only: new `internal/eval/control.go`, extend `gira eval` with `--control-total/--sample/--seed`, new MCP `get_control_total`; no existing meaning change.
- No model client (ADR 0001), no paid, quarantine ADR 0012, COPY idea in Go, no ground-truth changes, no push/release.

## Plan

1. Failing `TestSCL2ControlSample` first (forced-match/plug must fail).
2. Implement control-total + seeded sampler + blind re-derivation + frozen compare.
3. Wire CLI/MCP additively.
4. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

F2 delivered + F3/F4 sealed: control-total `{total,matched,mismatched,sampled_n,seed}`, seeded `max(20,5%)` sampler, blind re-derivation with never-plug/drop/force-match both sides, frozen compare first, `ControlTotal.Validate` gating CLI/MCP. Re-verified at a389d21; 03+04 unblocked.
