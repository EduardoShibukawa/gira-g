---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-02-l2-control-sample]
---

# L3 taxonomy and findings register (F3)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Every finding carries a Gallon S1 verbatim label (measured/checked/inferred/assumed) mapped to Gira tiers, plus coordinator/reviewer severity. Code-reading without a real-path repro is `inferred` at most. Additive `reviews/findings.md` register (pointer only, outside slices); fix-check re-runs the real path independently.

In scope: SC3, test-first. Failing test first: an unlabeled finding, and a fix-check that reads the fixer's log, must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC3
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
