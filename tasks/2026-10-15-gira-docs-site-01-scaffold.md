---
schema: task/1
tier: standard
status: verified
verified_at: "d6e6324"
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
| B1 | node tools/validate/validate.js (baseline) | pass | 3b01049 | repo clean before child |
| AC1 | bun install && bun run build in website/ | pass | d6e6324 | build/index.html exists, no external asset refs |
| AC2 | package.json scripts present | pass | d6e6324 | build/start/serve present |
| AC3 | docusaurus.config.ts valid | pass | d6e6324 | build passes with onBrokenLinks throw |
| R1 | go test -race ./... (no Go change) | pass | d6e6324 | no Go touched |
| V1 | diff read against AC1–AC3 | pass | d6e6324 | only website/ + .gitignore changed |

## Plan

1. `bun create docusaurus@latest website classic` in repo root.
2. Configure `docusaurus.config.ts` (title, url, baseUrl, preset-classic).
3. Add local search plugin (`@easyops-cn/docusaurus-search-local`).
4. Verify build: `bun install && bun run build`.
5. Run AC1–AC3/R1/V1; commit.

## Outcome

Scaffolded `website/` with Docusaurus v3 (package.json, docusaurus.config.ts,
sidebars.ts, tsconfig.json, .css, intro.md). Built a clean static bundle
(`build/index.html` exists, exit 0). Added `.gitignore` exclusions for
`website/node_modules/`, `build/`, `.docusaurus/`. AC1–AC3, R1, V1 all pass
at d6e6324. Search/config wiring deferred to child 05.
