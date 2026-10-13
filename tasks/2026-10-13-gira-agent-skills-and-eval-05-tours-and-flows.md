---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Tours and flows (SC8)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira tour`, `gira flows`, MCP `get_tour`.

In scope: spec §A items covered by SC8.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC8

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC8 | not run | | |
| V1 | read the diff against the ACs | not run | | |
