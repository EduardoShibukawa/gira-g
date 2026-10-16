---
schema: task/1
tier: standard
status: todo
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-02-concepts, 2026-10-15-gira-docs-site-03-cli-reference, 2026-10-15-gira-docs-site-04-guides]
---

# 05 — Navigation and search (F3; SC3, SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Wire the sidebar (`sidebars.ts`) with the ordered category list (Concepts →
CLI Reference → Guides → Export) and verify local search indexes all pages.

In scope: sidebar config, search plugin config, no broken internal links.
Out of scope: content creation (02–04), build pipeline (06).

## Acceptance criteria

- AC1 = SC3: sidebar renders in the ordered category list; every internal
  link resolves (no 404 in build).
- AC2 = SC4: local search indexes all pages; search returns results for CLI
  command names and concept terms.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | HEAD | |
| AC1 | bun run build — no broken links | not run | | |
| AC2 | local search returns results for CLI + concept terms | not run | | |
| R1 | go test -race ./... (no Go change) | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |

## Plan

1. Configure `sidebars.ts` with ordered categories.
2. Verify search plugin indexes all docs.
3. Run build; check for broken links.
4. Test search locally.
5. Run AC1–AC2/R1/V1; commit.

## Outcome

pending
