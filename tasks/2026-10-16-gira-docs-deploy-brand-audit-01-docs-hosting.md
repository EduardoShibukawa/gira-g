---
schema: task/1
tier: standard
status: active
parent: 2026-10-16-gira-docs-deploy-brand-audit
spec: 2026-10-16-gira-docs-deploy-brand-audit@1
---

# Docs deploy 01: GitHub Pages hosting config & deployment workflow

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-docs-deploy-brand-audit)

## Objective

Configure Docusaurus production URL and baseUrl for GitHub Pages
(`https://eduardoshibukawa.github.io` + `/gira-g/`), upgrade
`.github/workflows/docs.yml` to a two-job (`docs-build` +
`docs-deploy`) GitHub Pages deployment pipeline, and link the
published documentation site in `README.md` (`## Documentação`).

In scope: `website/docusaurus.config.ts`, `.github/workflows/docs.yml`,
`README.md`.
Out of scope: SVG brand assets, `custom.css`, Go code.

## Acceptance criteria

1. AC1 = SC1: `website/docusaurus.config.ts` sets
   `url: 'https://eduardoshibukawa.github.io'` and
   `baseUrl: '/gira-g/'`; no `https://gira.local` remains in
   `website/docusaurus.config.ts`; `README.md` links to
   `https://eduardoshibukawa.github.io/gira-g/`.
2. AC2 = SC2: `.github/workflows/docs.yml` defines `docs-build`
   (`contents: read`, `mise run docs:build`, canonical quarantine grep,
   `actions/upload-pages-artifact@v3` on `website/build/`) and
   `docs-deploy` (`needs: docs-build`, `pages: write`,
   `id-token: write`, `environment: github-pages`, `concurrency: pages`,
   `actions/deploy-pages@v4` on `main` push / `workflow_dispatch`).

## Constraints and budget

- Subagents: 1 (delegated `self` subagent for isolated implementation and verification of child 01 per `artifact-contracts.md` §3.5).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| B2 | mise run docs:build (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| AC1 | docusaurus.config.ts url/baseUrl + README link check | not run | - | pending implementation |
| AC2 | .github/workflows/docs.yml build + deploy-pages check | not run | - | pending implementation |
| R1 | mise run docs:build + quarantine grep + validator | not run | - | pending implementation |
| V1 | diff + gira impact --from-diff review against AC1–AC2 | not run | - | pending implementation |

## Plan

1. Update `url` and `baseUrl` in `website/docusaurus.config.ts`.
2. Update `.github/workflows/docs.yml` with `workflow_dispatch`, `actions/upload-pages-artifact@v3`, and `docs-deploy` job (`actions/deploy-pages@v4`).
3. Add the published Docs Site link to `README.md` under `## Documentação`.
4. Verify with `mise run docs:build`, canonical quarantine grep, `bin/gira impact --from-diff`, and `node tools/validate/validate.js`.
