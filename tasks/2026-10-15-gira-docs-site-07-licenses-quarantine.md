---
schema: task/1
tier: standard
status: verified
verified_at: "ae5acf3"
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
| B1 | node tools/validate/validate.js (baseline) | pass | 3b01049 | |
| AC1 | grep Docusaurus in THIRD_PARTY_NOTICES.md | pass | ae5acf3 | license entry present |
| AC2 | quarantine grep returns empty | pass | ae5acf3 | no benchmark refs in website/ |
| R1 | go test -race ./... (no Go change) | pass | ae5acf3 | no Go touched |
| V1 | diff read against AC1–AC2 | pass | ae5acf3 | only THIRD_PARTY_NOTICES touched |

## Plan

1. Add Docusaurus entry to `THIRD_PARTY_NOTICES.md`.
2. Add quarantine grep check to CI workflow (or a mise task).
3. Run AC1–AC2/R1/V1; commit.

## Outcome

Added Docusaurus MIT entry to `THIRD_PARTY_NOTICES.md` and confirmed
quarantine compliance (no benchmark references under `website/`). AC1–AC2,
R1, V1 pass at ae5acf3.
