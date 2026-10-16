---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-polish-finish
spec: 2026-10-16-gira-polish-finish@1
depends_on: [2026-10-16-gira-polish-finish-01-docs-correctness]
---

# Polish 02: site identity + 0.1.0 version snapshot

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Align site identity with real coordinates and snapshot released
docs: org/project `EduardoShibukawa`/`gira-g` (url stays placeholder),
`.nojekyll` shipped, `docs:version 0.1.0` (`versions.json` +
`versioned_docs/`).

In scope: `website/docusaurus.config.ts` identity fields,
`website/static/.nojekyll`, version snapshot.
Out of scope: CNAME, deploy action, docs prose, CI files.

## Acceptance criteria

1. AC1 = SC4a: org/project match real coordinates, url stays
   `https://gira.local`.
2. AC2 = SC4b: `versions.json` contains `0.1.0`;
   `build/.nojekyll` present after `docs:build`.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 8419f77 | pre-edit, exit 0 |
| B2 | mise run docs:build (baseline) | pass | 8419f77 | pre-edit |
| AC1 | config identity read | not run | | |
| AC2 | versions.json + build/.nojekyll | not run | | |
| R1 | mise run docs:build + validator | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |

## Plan

1. Edit org/project; add `static/.nojekyll`.
2. Run version snapshot for `0.1.0`; full build; verify.
