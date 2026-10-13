---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Impact from diff (SC7)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira impact --from-diff` and MCP `get_pr_impact`.

In scope: spec §A items covered by SC7.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC7

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC7 | not run | | |
| V1 | read the diff against the ACs | not run | | |
