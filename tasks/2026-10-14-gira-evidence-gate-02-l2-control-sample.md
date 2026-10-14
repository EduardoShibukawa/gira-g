---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-01-l1-artifact-verified]
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
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
