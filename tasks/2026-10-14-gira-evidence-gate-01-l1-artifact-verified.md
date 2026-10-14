---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
---

# L1 artifact-verified closing line (F1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Every `eval`/`verify` run emits `ARTIFACT-VERIFIED sha256:<hex>` as its final line; the digest covers command/expect/cwd + exit + frozen input fingerprint (1 MiB cap). Writes outside the frozen scope fail closed. Dispatcher 5-point self-check.

In scope: SC1, test-first. Failing test first: a tampered digest must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` and the EndToEnd eval report | not run | | |
