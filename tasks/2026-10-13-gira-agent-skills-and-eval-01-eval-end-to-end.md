---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
---

# Eval end to end (SC1–SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Shared indexing unit used by the CLI and the evaluator; a hermetic mini-corpus materialized as a deterministic git repo; a Go perturbation; ground truth v2 (no macro_map, `reviewed` flag); measured grep baseline; real Unsupported Safety; index determinism; gate semantics.

In scope: spec §C items covered by SC1–SC5.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC2
- AC3 = SC3
- AC4 = SC4
- AC5 = SC5

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1 | test for SC1 | not run | | |
| AC2 | test for SC2 | not run | | |
| AC3 | test for SC3 | not run | | |
| AC4 | test for SC4 | not run | | |
| AC5 | test for SC5 | not run | | |
| V1 | read the diff against the ACs | not run | | |
