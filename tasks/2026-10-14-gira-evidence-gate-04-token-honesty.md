---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-02-l2-control-sample]
---

# Honest token accounting (F4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Token report shows `{gross, expansions, net, measured_bool}` with `net = gross − expansions`. Estimated-only claims are quarantined with an `estimated` label and never enter headline numbers. No `len/4`. Escape-tax lift before measuring. Measurement only — envelopes stay E3's scope.

In scope: SC4, test-first. Failing test first: a `len/4` estimate passing as measured must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC4
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
