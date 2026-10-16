---
schema: task/1
tier: standard
status: todo
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-05-navigation-search, 2026-10-15-gira-docs-site-06-build-pipeline]
---

# 07 — Licenses and quarantine compliance (F5; SC6, SC8)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Add the Docusaurus MIT entry to `THIRD_PARTY_NOTICES.md` and enforce
quarantine compliance: no file under `website/` references the quarantined
comparative benchmark module (ADR 0012 rule 3), enforced by a static grep
check in CI.

In scope: license notice, quarantine grep check in CI.
Out of scope: content, build, navigation.

## Acceptance criteria

- AC1 = SC8: `THIRD_PARTY_NOTICES.md` has a Docusaurus entry (MIT, version,
  license text reference).
- AC2 = SC6: no file under `website/` references the quarantined comparative
  benchmark (grep for benchmark module identifiers returns empty); a CI check
  enforces this.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | HEAD | |
| AC1 | grep Docusaurus in THIRD_PARTY_NOTICES.md | not run | | |
| AC2 | grep -r <benchmark-ids> website/ returns empty | not run | | |
| R1 | go test -race ./... (no Go change) | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |

## Plan

1. Add Docusaurus entry to `THIRD_PARTY_NOTICES.md`.
2. Add quarantine grep check to CI workflow (or a mise task).
3. Run AC1–AC2/R1/V1; commit.

## Outcome

pending
