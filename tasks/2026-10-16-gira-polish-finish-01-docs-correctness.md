---
schema: task/1
tier: standard
status: active
parent: 2026-10-16-gira-polish-finish
spec: 2026-10-16-gira-polish-finish@1
---

# Polish 01: docs correctness + full CLI surface

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Fix wrong CLI examples and complete the reference surface: real
flags in `intro.md` / `cli/impact.md`, six new `cli/*.md` pages with
exit codes, exit codes on all CLI pages, dead ADR link removed,
single Export nav entry, search `indexBlog: false`.

In scope: `website/docs/**`, `website/sidebars.ts` (Export dedup +
new pages), search tuning in `docusaurus.config.ts`.
Out of scope: org/project identity, versioning, CI files, notices.

## Acceptance criteria

1. AC1 = SC1: real flags (`gira index`, positional ids), no
   `--service`/`--contract` in cli/intro docs, no `ADR 0004` link.
2. AC2 = SC2: `cli/explain.md`, `tour.md`, `diff.md`, `flows.md`,
   `locate.md`, `skills.md` exist (flags, inputs, output, example,
   exit codes); every `cli/*.md` documents exit codes.
3. AC3 = SC3: single top-level `Export`; `docs:build` exits 0;
   search has `indexBlog: false`.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 8419f77 | pre-edit, exit 0 |
| B2 | mise run docs:build (baseline) | pass | 8419f77 | pre-edit, same website tree as badf4db build |
| AC1 | flag greps + serve.md link check | not run | | |
| AC2 | six new pages + exit-code grep | not run | | |
| AC3 | sidebar read + docs:build + search config | not run | | |
| R1 | mise run docs:build + validator | not run | | |
| V1 | diff read against AC1–AC3 | not run | | |

## Plan

1. Fix `intro.md` + `cli/impact.md` examples; drop dead ADR link.
2. Write six new `cli/*.md` pages with exit codes; add exit codes
   to existing pages.
3. Dedup Export in `sidebars.ts`; set `indexBlog: false`; build.
