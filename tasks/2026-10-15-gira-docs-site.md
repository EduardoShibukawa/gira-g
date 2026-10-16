---
schema: task/1
tier: heavy
status: active
spec: 2026-10-15-gira-docs-site@1
---

# E6: documentation website (gira-docs-site)

## Routing and approvals

- tier: heavy
- confirmed: heavy by owner (2026-10-15, "confirmado heavy")
- decided: ferramenta → Docusaurus by owner (2026-10-15, "ok")
- decided: escopo de conteúdo → funcionalidade-primeiro (CLI reference, conceitos, guides, export) by owner (2026-10-15, "focando em funcionalidade e nao em ADRs, decisions, etc..")
- decided: navegação → site único com categorias ordenadas by owner (2026-10-15, "Site único com categorias ordenadas (Recomendado)")
- decided: versionamento → ativo, seguindo tags de release by owner (2026-10-15, "versao agora temos podemos verificar seguindo a tag da release nao?")
- approved: spec v1 by owner (2026-10-15, "ok")

## Objective

Build a static documentation website for the `gira` CLI using Docusaurus, covering concepts, CLI reference, guides, and export — focused on functionality, not internal ADRs/decisions. Versioning follows release tags.

## Acceptance criteria

- SC1 scaffold: `website/` builds clean with Docusaurus v3
- SC2 content: every CLI command and every CONTEXT.md term has a page
- SC3 navigation: sidebar ordered categories, no broken links
- SC4 search: local search indexes all pages
- SC5 build pipeline: `mise run docs:build` + CI workflow
- SC6 quarantine: no benchmark references in `website/`
- SC7 regression: Go tests unchanged, validator exit 0
- SC8 licenses: Docusaurus entry in THIRD_PARTY_NOTICES.md

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | mise run docs:build (clean checkout) | not run | | pending children |
| B2 | content completeness | not run | | pending children |
| B3 | CI workflow valid | not run | | pending children |
| B4 | quarantine grep | not run | | pending children |
| B5 | go test -race ./... | not run | | pending children |
| B6 | node tools/validate/validate.js | not run | | pending children |

## Plan

1. 2026-10-15-gira-docs-site-01-scaffold (no deps)
2. 2026-10-15-gira-docs-site-02-concepts (deps: 01)
3. 2026-10-15-gira-docs-site-03-cli-reference (deps: 01)
4. 2026-10-15-gira-docs-site-04-guides (deps: 02, 03)
5. 2026-10-15-gira-docs-site-05-navigation-search (deps: 02, 03, 04)
6. 2026-10-15-gira-docs-site-06-build-pipeline (deps: 01)
7. 2026-10-15-gira-docs-site-07-licenses-quarantine (deps: 05, 06)

## Review

pending adversarial review

## Outcome

pending
