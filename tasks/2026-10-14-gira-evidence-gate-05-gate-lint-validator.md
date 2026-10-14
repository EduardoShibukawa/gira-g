---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-03-l3-taxonomy-register, 2026-10-14-gira-evidence-gate-04-token-honesty]
---

# gate-lint and validator integration (F5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Port of the unlazy gate contract (idea, Go reimplementation): success-only marker + digest, stale-on-drift fail-closed, `--status`/`--reverify` split, `ABANDON` as qualified handoff that never reports `ALL MET`. Warn-only in CI until E3; fail-closed locally. Integrated SC1–SC6 plus the E0 recall/precision guard.

In scope: SC5 and SC6, test-first.
Out of scope: ground-truth changes.

## Acceptance criteria

- AC1 = SC5
- AC2 = SC6

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
