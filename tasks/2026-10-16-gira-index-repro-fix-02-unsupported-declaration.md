---
schema: task/1
tier: standard
status: verified
verified_at: "5b5deab"
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
| B1 | parent B1 (`go test -race`, pre-children) | not run | a82e44b | cited from parent; rerun in 04 |
| AC1 | section present + `mise run docs:build` | pass | 5b5deab | Unsupported-by-design section live; build SUCCESS |
| V1 | diff read against AC1 | pass | 5b5deab | 1 file, 15 insertions |

## Outcome

Declaration shipped in extractor.md. No follow-ups.
