---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
---

# Bench quarantine (SC12)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Benchmark workspace outside the Go module: container recipes, two-phase runs, ignored output, static isolation check.

In scope: spec §B items covered by SC12.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC12

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC12 | not run | | |
| V1 | read the diff against the ACs | not run | | |
