---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-01-eval-end-to-end]
---

# Arch diff (SC9)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira diff`, MCP `diff_maps`.

In scope: spec §A items covered by SC9.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC9

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC9 | not run | | |
| V1 | read the diff against the ACs | not run | | |
