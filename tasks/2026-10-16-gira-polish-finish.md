---
schema: task/1
tier: heavy
status: accepted
verified_at: "5b0962b"
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
- accepted: by owner (2026-10-16, "Aceito")

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
| B1 | node tools/validate/validate.js (baseline) | pass | 8419f77 | pre-child, exit 0 |
| B2 | mise run docs:build (baseline) | pass | 8419f77 | pre-child, same website tree, exit 0 |
| SC1 | docs-correctness greps + flag review | pass | 5b0962b | --service/--contract/ADR refs empty |
| SC2 | full-surface pages + exit codes | pass | 5b0962b | 6 new pages, 12/12 with exit codes |
| SC3 | sidebar single Export + build + search tuning | pass | 5b0962b | build exit 0, indexBlog false |
| SC4 | identity + versions.json + build/.nojekyll | pass | 5b0962b | real org/project, 0.1.0, .nojekyll ships |
| SC5 | CI grep + frozen-lockfile + THIRD_PARTY | pass | 5b0962b | canonical pattern, empty run, six entries |
| SC6 | go test -race + validator + quarantine grep | pass | 5b0962b | suite pass under mise exec, validator 0 |
| R1 | integrated build + go test + validator | pass | 5b0962b | bare go-test failure was mise-shim env only, pass under mise exec |
| V1 | integrated diff read against SC1–SC6 | pass | 5b0962b | review completed, no findings |

## Review

- Fresh-context adversarial review (prompt per review.md): PASS, no
  findings. All SC1–SC6 verified against `cmd/gira/main.go` flags and
  canonical quarantine regex; zero Go files touched; validator exit 0.

## Outcome

All 3 children verified at `@1`. Integrated head `5b0962b`: docs
examples use real flags, full CLI surface documented with exit codes,
single Export nav, real org/project identity, `0.1.0` snapshot,
canonical quarantine grep, frozen docs install, complete notices.
Awaiting owner acceptance.
