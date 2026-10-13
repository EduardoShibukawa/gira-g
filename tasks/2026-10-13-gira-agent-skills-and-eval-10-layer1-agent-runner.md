---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-08-bench-quarantine, 2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Layer1 agent runner (SC14)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Agent runner with an MCP client, mock-model mode, telemetry schema, deterministic scorer.

In scope: spec §B items covered by SC14.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC14

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC14 | not run | | |
| V1 | read the diff against the ACs | not run | | |
