---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-04-impact-from-diff, 2026-10-13-gira-agent-skills-and-eval-05-tours-and-flows, 2026-10-13-gira-agent-skills-and-eval-06-arch-diff]
---

# Skills distribution (SC10, SC11)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Canonical skill sources; `gira skills install` for Claude Code, Cursor and agents hosts; MCP/CLI fallback contract.

In scope: spec §A items covered by SC10, SC11.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC10
- AC2 = SC11

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC10 | not run | | |
| AC2 | test for SC11 | not run | | |
| V1 | read the diff against the ACs | not run | | |
