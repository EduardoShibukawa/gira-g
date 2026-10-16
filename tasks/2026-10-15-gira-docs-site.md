---
schema: task/1
tier: heavy
status: verified
verified_at: "d5f774e"
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
- divergence: SC2/F2: spec listed "gira map", "gira serve", "gira index" as CLI commands; the CLI exposes `gira index`, `gira serve mcp`, `gira export/view`. Docs document real commands and add `website/docs/cli/map.md` (alias page) so SC2's per-command page holds without claiming a non-existent flag set.
- approved: exception accept real CLI command names + cli/map.md placeholder by owner (2026-10-15, "1")

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
| B1 | node tools/validate/validate.js (baseline) | pass | 3b01049 | before any E6 edit |
| SC1 | scaffold + build/index.html | pass | d5f774e | clean build, exit 0 |
| SC2 | every CLI cmd + CONTEXT term has a page | pass | d5f774e | 6 cli pages (map alias) + 14 concepts |
| SC3 | sidebar ordered + no broken links | pass | d5f774e | build onBrokenLinks throw, passes |
| SC4 | local search plugin | pass | d5f774e | @easyops configured |
| SC5 | mise run docs:build + CI workflow valid | pass | d5f774e | YAML parses, exit 0 |
| SC6 | quarantine grep | pass | d5f774e | exit 1 (empty) |
| SC7 | go test -race + validator | pass | d5f774e | exit 0 |
| SC8 | Docusaurus in THIRD_PARTY_NOTICES | pass | d5f774e | entry present |
| R1 | integrated build + go test | pass | d5f774e | |
| V1 | parent diff read vs spec | pass | d5f774e | review completed, 5 findings resolved |

## Plan

1. 2026-10-15-gira-docs-site-01-scaffold (no deps)
2. 2026-10-15-gira-docs-site-02-concepts (deps: 01)
3. 2026-10-15-gira-docs-site-03-cli-reference (deps: 01)
4. 2026-10-15-gira-docs-site-04-guides (deps: 02, 03)
5. 2026-10-15-gira-docs-site-05-navigation-search (deps: 02, 03, 04)
6. 2026-10-15-gira-docs-site-06-build-pipeline (deps: 01)
7. 2026-10-15-gira-docs-site-07-licenses-quarantine (deps: 05, 06)

## Review

- F1 (major) SC2 missing "gira map" page → fixed in 2a3e753 (added website/docs/cli/map.md alias page + spec divergence approved as exception)
- F2 (major) SC3 sidebar ordering / missing Export category → fixed in 2a3e753 (sidebars.ts reordered per spec; cli/export labeled Export)
- F3 (major) SC2 serve/slice/export missing output shape or example → fixed in 2a3e753
- F4 (minor) freshness statuses `current`/`stale`/`unverified` not fully documented → fixed in 2a3e753
- F5 (minor) intro "coming soon" stale link → fixed in 2a3e753 (now links to Guides)

## Outcome

All 7 children verified. Docs site builds clean (`mise run docs:build` exit 0),
no broken links, local search wired, quarantine compliant, license notice added.
Review findings (5) resolved in 2a3e753 via an approved exception for the
CLI-command naming divergence. Awaiting owner acceptance.
