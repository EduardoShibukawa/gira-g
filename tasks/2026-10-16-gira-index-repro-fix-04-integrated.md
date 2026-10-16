---
schema: task/1
tier: standard
status: verified
verified_at: "26a28e5"
parent: 2026-10-16-gira-index-repro-fix
spec: 2026-10-16-gira-index-repro-fix@1
depends_on:
  - 2026-10-16-gira-index-repro-fix-01-offline-bundle-bootstrap
  - 2026-10-16-gira-index-repro-fix-02-unsupported-declaration
  - 2026-10-16-gira-index-repro-fix-03-governance-note
---

# Index repro fix 04: integrated verification

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-index-repro-fix)

## Objective

Verify SC1–SC3 end to end on the integrated head (SC rows).

In scope: SC1–SC3 checks, full regression, review input.
Out of scope: new changes (fix-forward via new tasks only).

## Acceptance criteria

1. AC1 = SC1: bootstrap `--help` + index.md section live.
2. AC2 = SC2: extractor.md declaration live.
3. AC3 = SC3: slice help + slice.md note live.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | parent B1 (`go test -race`, pre-children) | not run | a82e44b | baseline, kept |
| AC1 | SC1 live check | pass | 26a28e5 | bootstrap `--help` ok; index.md Offline bundle section live |
| AC2 | SC2 live check | pass | 26a28e5 | extractor.md Unsupported-by-design live |
| AC3 | SC3 live check | pass | 26a28e5 | slice help + slice.md:42 note live |
| R1 | `go test -race ./...` + lint + validator | pass | 26a28e5 | suite exit 0 (17 ok); lint 0 new; validator pass |
| V1 | integrated diff read against SC1–SC3 | pass | 26a28e5 | 8 files, +124/−16 vs eec886f; no contract breaks |

## Outcome

SC1–SC3 live on integrated head. No fix-forward changes; follow-ups
via new tasks only.
