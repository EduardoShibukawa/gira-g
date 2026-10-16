---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-index-repro-fix
spec: 2026-10-16-gira-index-repro-fix@1
---

# Index repro fix 02: unsupported-by-design declaration

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-index-repro-fix)

## Objective

Declare site/docs sources unsupported by design in
`website/docs/concepts/extractor.md` (SC2).

In scope: extractor.md section.
Out of scope: new extractors.

## Acceptance criteria

1. AC1: extractor.md names `.md`/TS/CSS/SVG site sources as unsupported
   by design with `locate` pointer.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | parent B1 (`go test -race`, pre-children) | not run | | cited from parent; rerun in 04 |
| AC1 | section present + `mise run docs:build` | not run | | |
| V1 | diff read against AC1 | not run | | |
