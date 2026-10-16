---
schema: task/1
tier: heavy
status: active
spec: 2026-10-16-gira-docs-deploy-brand-audit@1
---

# Docs hosting deploy, real brand visual identity, and integrated Gira NFR audit

## Routing and approvals

- tier: heavy (H1 product/brand/hosting decisions, H2 external GitHub Pages deploy publication, H3 multi-subsystem across website/CI/docs/audit, H5 decomposed into 3 standard child tasks)
- confirmed: heavy by owner (2026-10-16, "(Recommended) Confirmar tier: heavy com decomposição em 3 tasks filhas standard (01-docs-hosting, 02-brand-identity, 03-nfr-gira-audit) e subagents habilitados no budget")
- decided: Q1 hosting URL & baseUrl → https://eduardoshibukawa.github.io + /gira-g/ without CNAME + README link by owner (2026-10-16, "(Recommended) GitHub Pages padrão do projeto: url 'https://eduardoshibukawa.github.io', baseUrl '/gira-g/', sem CNAME, adicionando link em README.md")
- decided: Q2 docs.yml workflow → two jobs (docs-build + docs-deploy on main push / workflow_dispatch with deploy-pages@v4) by owner (2026-10-16, "(Recommended) Dois jobs (docs-build + docs-deploy em push na main / workflow_dispatch, com deploy-pages@v4 incondicional na main)")
- decided: Q3 SVG brand direction → Helio-Graph pure path vector mark with dedicated 16x16 favicon.svg by owner (2026-10-16, "(Recommended) Helio-Graph (coroa de pétalas de girassol + núcleo em grafo formando G em <path> puro) com favicon.svg dedicado otimizado para 16x16")
- decided: Q4 theme palette → Solar Graph amber/gold (#b45309 light / #f59e0b dark) in custom.css by owner (2026-10-16, "(Recommended) Paleta Solar Graph: atualizar custom.css do verde padrão Infima para Âmbar/Ouro Girassol acessível (#b45309 light / #f59e0b dark) coeso com os SVGs")
- approved: spec v1 by owner (2026-10-16, "(Recommended) Aprovar spec v1 (specs/2026-10-16-gira-docs-deploy-brand-audit.md) e prosseguir com a decomposição e execução das 3 tasks filhas via subagents")

## Objective

Execute the three remaining production-readiness fronts for `gira-g`:
(b) configure GitHub Pages hosting and deployment for the Docusaurus
site (`url: 'https://eduardoshibukawa.github.io'`, `baseUrl: '/gira-g/'`,
two-job `.github/workflows/docs.yml`, and `README.md` link);
(c) replace placeholder SVGs and stock green CSS with the definitive
**Helio-Graph** vector brand assets (`logo.svg`, `favicon.svg`,
`gira-social-card.svg`) and **Solar Graph** theme palette;
(d) execute a comprehensive NFR and architectural dogfooding audit
using `gira` (`index`, `locate`, `slice`, `explain`,
`impact --from-diff`, `diff`, `skills check`), `mise run docs:build`,
`mise exec -- go test -race ./...`, the canonical quarantine grep, and
the G-Ingá validator.

## Acceptance criteria

SC1–SC6 as defined in `specs/2026-10-16-gira-docs-deploy-brand-audit.md`.

## Constraints and budget

- Subagents: 7 (expanded from default `0` per `artifact-contracts.md` §3.5 to allow parallel context-isolated technical inspection via 3 `research` subagents, isolated implementation of the 3 decomposed `standard` child tasks via 3 `self` subagents, and 1 fresh-context adversarial review subagent per `review.md`).

## Plan

1. `2026-10-16-gira-docs-deploy-brand-audit-01-docs-hosting`: Configure `website/docusaurus.config.ts` (`url`, `baseUrl`), `.github/workflows/docs.yml` (`docs-build` + `docs-deploy` via GitHub Pages actions), and `README.md` documentation link (SC1, SC2).
2. `2026-10-16-gira-docs-deploy-brand-audit-02-brand-identity`: Replace `website/static/img/logo.svg`, `website/static/img/favicon.svg`, and `website/static/img/gira-social-card.svg` with Helio-Graph pure vector assets and update `website/src/css/custom.css` to the Solar Graph palette (SC3, SC4).
3. `2026-10-16-gira-docs-deploy-brand-audit-03-nfr-gira-audit`: Run full `gira` dogfooding audit (`index`, `locate`, `slice`, `explain`, `impact --from-diff`, `diff`, `skills check`) and NFR verification suite across the integrated repository (SC5, SC6).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| B2 | mise run docs:build (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| B3 | mise exec -- go test -race ./... (baseline) | not run | 41781d9 | pending child/parent execution |
| SC1 | hosting-config (docusaurus.config.ts url/baseUrl + README link) | not run | - | pending child 01 |
| SC2 | pages-workflow (.github/workflows/docs.yml build + deploy-pages) | not run | - | pending child 01 |
| SC3 | brand-assets (logo.svg, favicon.svg, gira-social-card.svg Helio-Graph) | not run | - | pending child 02 |
| SC4 | theme-palette (custom.css Solar Graph light/dark) | not run | - | pending child 02 |
| SC5 | gira-dogfood-audit (index, locate, slice, explain, impact, diff, skills) | not run | - | pending child 03 |
| SC6 | nfr-gates (docs:build, go test -race, quarantine grep, validator) | not run | - | pending child 03 |
| R1 | integrated regression suite (go test -race + docs:build + validator) | not run | - | pending parent verification |
| V1 | self-review + adversarial review against SC1–SC6 | not run | - | pending parent verification |

## Review

Pending child implementation and fresh-context adversarial review per `.agents/skills/work/review.md`.

