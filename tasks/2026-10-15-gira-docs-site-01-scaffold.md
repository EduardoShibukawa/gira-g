---
schema: task/1
tier: standard
status: todo
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
---

# 01 — Docusaurus scaffold (F1; SC1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Create the `website/` directory with a Docusaurus v3 project scaffold:
`package.json`, `docusaurus.config.ts`, `sidebars.ts`, `docs/` structure,
and static assets. Build produces a self-contained static bundle.

In scope: project init, config, empty docs structure, build scripts.
Out of scope: content pages (children 02–04), navigation wiring (05),
CI pipeline (06), licenses (07).

## Acceptance criteria

- AC1 = SC1: `bun install && bun run build` in `website/` exits 0 and
  produces `build/index.html` with no external asset references.
- AC2: `package.json` has `build`, `start`, and `serve` scripts.
- AC3: `docusaurus.config.ts` sets `title`, `url`, `baseUrl`, `preset-classic`,
  and local search plugin (no Algolia).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | HEAD | repo clean before child |
| AC1 | bun install && bun run build in website/ | not run | | |
| AC2 | package.json scripts present | not run | | |
| AC3 | docusaurus.config.ts valid | not run | | |
| R1 | go test -race ./... (no Go change) | not run | | |
| V1 | diff read against AC1–AC3 | not run | | |

## Plan

1. `bun create docusaurus@latest website classic` in repo root.
2. Configure `docusaurus.config.ts` (title, url, baseUrl, preset-classic).
3. Add local search plugin (`@easyops-cn/docusaurus-search-local`).
4. Verify build: `bun install && bun run build`.
5. Run AC1–AC3/R1/V1; commit.

## Outcome

pending
