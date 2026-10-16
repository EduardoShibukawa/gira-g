---
schema: task/1
tier: heavy
status: active
spec: 2026-10-16-gira-polish-finish@1
---

# Polish finish: docs correctness + identity/version + CI hygiene

## Routing and approvals

- tier: heavy (H1 product calls abertas, H3 multi-subsistemas, H5 multi-sessão)
- confirmed: heavy by owner (2026-10-16, "Heavy fatiado (Recomendado)")
- decided: tier → heavy fatiado by owner (2026-10-16, "Heavy fatiado (Recomendado)")
- decided: Q1 hosting URL → manter https://gira.local by owner (2026-10-16, "Manter placeholder (Recomendado)")
- decided: Q2 org/project → alinhar real EduardoShibukawa/gira-g by owner (2026-10-16, "Alinhar real (Recomendado)")
- decided: Q3 versionamento → snapshot 0.1.0 agora by owner (2026-10-16, "Snapshot v0.1.0 (Recomendado)")
- decided: Q4 escopo CLI → full surface by owner (2026-10-16, "Full surface (Recomendado)")
- decided: Q5 extras → .nojekyll agora, CNAME deferred by owner (2026-10-16, ".nojekyll agora (Recomendado)")
- approved: spec v1 by owner (2026-10-16, "Spec aprovada")

## Objective

Finish residual E6 gaps without touching Go behavior or hosting infra:
correct CLI examples, full-surface reference pages with exit codes,
deduped Export nav, real org/project identity, 0.1.0 docs snapshot,
and widened quarantine/frozen/licence hygiene.

In scope: spec F1–F3 (website docs/config/sidebars/versioned,
docs.yml grep, mise docs:build, THIRD_PARTY_NOTICES).
Out of scope: Pages deploy, CNAME, real brand assets, Go changes,
release workflows.

## Acceptance criteria

SC1–SC6 from the spec.

## Plan

1. 2026-10-16-gira-polish-finish-01-docs-correctness: real flags, 6 new CLI pages, exit codes, sidebar Export dedup, search tuning
2. 2026-10-16-gira-polish-finish-02-identity-version: org/project align, .nojekyll, docs:version 0.1.0 snapshot
3. 2026-10-16-gira-polish-finish-03-ci-hygiene: widened quarantine grep, frozen-lockfile, THIRD_PARTY gaps

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | not run | | before any child |
| B2 | mise run docs:build (baseline) | not run | | before any child |
| SC1 | docs-correctness greps + flag review | not run | | |
| SC2 | full-surface pages + exit codes | not run | | |
| SC3 | sidebar single Export + build + search tuning | not run | | |
| SC4 | identity + versions.json + build/.nojekyll | not run | | |
| SC5 | CI grep + frozen-lockfile + THIRD_PARTY | not run | | |
| SC6 | go test -race + validator + quarantine grep | not run | | |
| R1 | integrated build + go test + validator | not run | | |
| V1 | integrated diff read against SC1–SC6 | not run | | |

## Review

skip: runs after child 03 is verified
