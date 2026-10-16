---
schema: task/1
tier: heavy
status: accepted
verified_at: "56c1212"
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
- accepted: by owner (2026-10-16, "(Recommended) Aceitar a entrega integrada (tasks: accept gira-docs-deploy-brand-audit by owner)")

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
| B3 | mise exec -- go test -race ./... (baseline) | pass | 41781d9 | pre-edit baseline exit 0 |
| SC1 | hosting-config (docusaurus.config.ts url/baseUrl + README link) | pass | 56c1212 | url=https://eduardoshibukawa.github.io, baseUrl=/gira-g/, README linked |
| SC2 | pages-workflow (.github/workflows/docs.yml build + deploy-pages) | pass | 56c1212 | docs.yml has docs-build + docs-deploy with deploy-pages@v4 |
| SC3 | brand-assets (logo.svg, favicon.svg, gira-social-card.svg Helio-Graph) | pass | 56c1212 | logo.svg (4.4KB), favicon.svg (1.4KB), gira-social-card.svg (7.6KB), 0 <text> in logo/favicon |
| SC4 | theme-palette (custom.css Solar Graph light/dark) | pass | 56c1212 | custom.css Solar Graph palette (#b45309 light / #f59e0b dark) |
| SC5 | gira-dogfood-audit (index, locate, slice, explain, impact, diff, skills) | pass | 56c1212 | skills check (3 hosts), index, locate, slice, explain, impact, diff pass |
| SC6 | nfr-gates (docs:build, go test -race, quarantine grep, validator) | pass | 56c1212 | docs:build exit 0, go test -race exit 0, quarantine grep exit 1 (0 matches), validator exit 0 |
| R1 | integrated regression suite (go test -race + docs:build + validator) | pass | 56c1212 | full regression suite, docs:build, go test -race, and validator pass cleanly |
| V1 | self-review + adversarial review against SC1–SC6 | pass | 56c1212 | adversarial review passed against SC1–SC6, zero findings |

## Review

- Fresh-context adversarial review against spec v1: PASS, zero findings. All SC1–SC6 verified directly against implementation and git diff `41781d9..HEAD`. No Go behavior touched, no protected skills touched, clean quarantine scan, validator exit 0.

## Outcome

Delivered all three production-readiness frentes (b, c, d) across 3 standard child tasks, verified at commit `56c1212`:
- **Frente (b) Hosting & Deploy:** Configured `website/docusaurus.config.ts` (`url: 'https://eduardoshibukawa.github.io'`, `baseUrl: '/gira-g/'`), upgraded `.github/workflows/docs.yml` with `workflow_dispatch`, `actions/upload-pages-artifact@v3` and `docs-deploy` job (`actions/deploy-pages@v4`), and linked the published Docs Site in `README.md`.
- **Frente (c) Brand & Visual Identity:** Replaced placeholder SVGs with definitive **Helio-Graph** pure-vector assets (`logo.svg`, 16x16-optimized `favicon.svg`, `gira-social-card.svg`, each `< 8 KB`, zero external font dependencies) and aligned `website/src/css/custom.css` with the **Solar Graph** theme palette (`#b45309` light / `#f59e0b` dark).
- **Frente (d) Dogfooding Audit & NFRs:** Audited `gira skills manifest` and `skills check` across all 3 hosts (`agents`, `claude`, `cursor`); indexed baseline and HEAD with `bin/gira index`; verified `locate`, `slice --tier summary`, `explain`, `impact --from-diff`, and `diff` (7 added nodes, 30 added edges, 0 contract regressions); passed `mise run docs:build`, canonical quarantine grep (0 matches), `mise exec -- go test -race ./...`, and `node tools/validate/validate.js`.


