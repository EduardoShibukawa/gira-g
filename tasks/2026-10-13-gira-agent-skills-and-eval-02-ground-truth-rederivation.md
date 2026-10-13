---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-01-eval-end-to-end]
---

# Ground truth rederivation (GT review)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Re-derive the ground truth against the full pinned OpenTelemetry Demo (opt-in run through the git cache). A human review sets `reviewed: true`.

In scope: spec §C items covered by GT review.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1: an opt-in full-corpus test exists and skips with a visible reason when it is not enabled
- AC2: the ground-truth locator anchors resolve at the pinned SHA [inferred]

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | not run | | baseline is taken at start |
| AC1: | test for an opt-in full-corpus test exists and skips with a visible reason when it is not enabled | not run | | |
| AC2: | test for the ground-truth locator anchors resolve at the pinned SHA [inferred] | not run | | |
| V1 | read the diff against the ACs | not run | | |
