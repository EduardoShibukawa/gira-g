---
schema: task/1
tier: heavy
status: todo
spec: 2026-10-16-gira-index-repro-fix@1
---

# Gira index reproducibility fix

## Routing and approvals

- tier: heavy (H3: S3 fails, manifest+cache+docs span subsystems; H5: S8 fails, fetch network-blocked, multi-session)

## Objective

Make `bin/gira index` reproducible per `specs/2026-10-16-gira-index-repro-fix.md`.

In scope: spec approval, children, integrated verification.
Out of scope: new extractors, remote repinning without owner call.

## Acceptance criteria

SC1–SC3 from the spec.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test -race ./...` + validator before children | not run | | |
| SC1 | offline index path reproducible where bundle exists | not run | | |
| SC2 | unsupported posture declared | not run | | |
| SC3 | governance behavior documented | not run | | |
| R1 | full suite + lint on integrated head | not run | | |
| V1 | integrated diff read against SC1–SC3 | not run | | |

## Plan

1. Question round + owner confirms heavy.
2. Approve spec v1.
3. Decompose into standard children, execute, integrate, review.

## Review

skip: runs after children are verified.
