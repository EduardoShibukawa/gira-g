---
schema: task/1
tier: standard
status: verified
verified_at: "afd1eaf"
parent: 2026-10-16-gira-docs-deploy-brand-audit
spec: 2026-10-16-gira-docs-deploy-brand-audit@1
depends_on:
  - 2026-10-16-gira-docs-deploy-brand-audit-01-docs-hosting
---

# Docs deploy 02: Helio-Graph vector brand assets & Solar Graph theme palette

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-docs-deploy-brand-audit)

## Objective

Replace placeholder SVG files (`website/static/img/logo.svg`,
`website/static/img/favicon.svg`,
`website/static/img/gira-social-card.svg`) with definitive **Helio-Graph**
pure vector brand assets (*Girassol* = *Gira* + musical note *G* / Sun
+ architectural graph) and align `website/src/css/custom.css` with the
**Solar Graph** primary color palette.

In scope: `website/static/img/logo.svg`,
`website/static/img/favicon.svg`,
`website/static/img/gira-social-card.svg`,
`website/src/css/custom.css`.
Out of scope: `docusaurus.config.ts` hosting URLs, CI workflows, Go code.

## Acceptance criteria

1. AC1 = SC3: `website/static/img/logo.svg`,
   `website/static/img/favicon.svg`, and
   `website/static/img/gira-social-card.svg` are valid SVGs implementing
   the Helio-Graph visual identity without external dependencies, each
   `< 8 KB`, with `logo.svg` and `favicon.svg` free of `<text>` elements
   and distinct (16x16-optimized `favicon.svg`).
2. AC2 = SC4: `website/src/css/custom.css` replaces stock Infima green
   (`#2e8550` / `#25c2a0`) with the Solar Graph amber/gold palette
   (`#b45309` light / `#f59e0b` dark).

## Constraints and budget

- Subagents: 1 (delegated `self` subagent for isolated implementation and verification of child 02 per `artifact-contracts.md` §3.5).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| B2 | mise run docs:build (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| AC1 | SVG validity, < 8 KB size, no <text> in logo/favicon, distinct favicon | pass | afd1eaf | logo.svg (4410 B), favicon.svg (1444 B), gira-social-card.svg (7582 B); 0 <text> in logo/favicon; distinct 16x16 favicon |
| AC2 | custom.css Solar Graph light (#b45309) and dark (#f59e0b) palette | pass | afd1eaf | replaced stock Infima green with Solar Graph amber/gold scale in :root and [data-theme='dark'] |
| R1 | mise run docs:build + quarantine grep + validator | pass | afd1eaf | docs:build exit 0; quarantine grep exit 1 (0 matches); go test -race exit 0; validate.js exit 0 |
| V1 | diff + gira impact --from-diff review against AC1–AC2 | pass | afd1eaf | 4 files changed within exact scope; gira impact --from-diff verified |

## Plan

1. Design and write pure-path `website/static/img/logo.svg` (Helio-Graph sunflower crown + architectural graph G core).
2. Design and write 16x16-optimized `website/static/img/favicon.svg` (high-contrast simplified Helio-Graph mark, no `<text>`).
3. Design and write `website/static/img/gira-social-card.svg` (`1200x630` Helio-Graph lockup and architectural graph constellation).
4. Update `website/src/css/custom.css` with the Solar Graph palette (`#b45309` light / `#f59e0b` dark).
5. Verify with `mise run docs:build`, canonical quarantine grep, `bin/gira impact --from-diff`, and `node tools/validate/validate.js`.

## Outcome

- Delivered definitive **Helio-Graph** pure vector brand assets at commit `afd1eaf`:
  - `website/static/img/logo.svg` (4,410 bytes, `viewBox="0 0 64 64"`, zero `<text>` elements): geometric sunflower petal crown (`#f59e0b` / `#fbbf24`) around a dark slate core (`#0f172a`) featuring an architectural node-and-edge graph tracing the letter **G** with `#38bdf8` evidence highlights.
  - `website/static/img/favicon.svg` (1,444 bytes, `viewBox="0 0 32 32"`, zero `<text>` elements): distinct 16×16-optimized high-contrast Helio-Graph badge with simplified 8-point crown and thicker G-graph strokes.
  - `website/static/img/gira-social-card.svg` (7,582 bytes, `1200×630`): OpenGraph card with architectural constellation background, Helio-Graph emblem, geometric `<path>` wordmark `GIRA`, positioning copy, and `gira index` / `gira impact` / `gira slice` workflow pills.
- Updated `website/src/css/custom.css` to replace stock Infima green (`#2e8550` / `#25c2a0`) with the **Solar Graph** primary color palette (`#b45309` in light mode `:root` and `#f59e0b` in `[data-theme='dark']`).
