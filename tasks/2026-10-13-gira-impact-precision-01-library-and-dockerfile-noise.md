---
schema: task/1
tier: standard
status: todo
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
---

# Library and Dockerfile noise out of the Impact Set (cause 1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Libraries reached through imports move to an additive `upstream_libraries` field; Dockerfile stages and base images become Service metadata instead of nodes.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
