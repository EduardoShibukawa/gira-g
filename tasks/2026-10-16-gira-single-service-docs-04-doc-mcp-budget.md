---
schema: task/1
tier: standard
status: verified
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
verified_at: "c6d7465"
depends_on:
  - 2026-10-16-gira-single-service-docs-03-doc-command
---

# Single-service docs 04: MCP `get_doc` + budget/bench (F3 MCP + F4/SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Espelhar `doc` na tool MCP `get_doc` (mesmo engine) e emitir budget honesto (`bytes/4`) com bench comparando contra baseline cru. Sem número caçado: medido e reportado.

In scope: MCP `get_doc` + SC4 do spec.
Out of scope: F1–F3 (pré-requisitos via `depends_on`).

## Acceptance criteria

1. AC1: `get_doc` retorna o mesmo conteúdo do CLI `doc --format json`.
2. AC2 = SC4: `doc` carrega budget honesto e o bench reporta razão vs baseline sem regredir determinismo/segurança.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test ./internal/mcp/ ./internal/slices/` antes da mudança | pass | b959a04 | sem falhas pré-existentes; base limpa |
| B2 | `go test ./...` no módulo bench/ separado antes da mudança | fail | b959a04 | pré-existentes, sem relação com 04: layer0 `TestAdaptersGolden` (falta `testdata/gira/macro-map.json`); quarantine `TestNoReferencesFromGuardedFiles` (2 violações: docs.yml + intra_service.go Graphify) |
| AC1 | MCP `get_doc` parity check | pass | c6d7465 | `TestSCDocMCPParity`: byte-identical ao CLI `doc --format json`; mesmos facts/IDs + twin md; envelope só com budget; erros honestos (unknown service, missing service_id, budget_exceeded em ceiling 5) |
| AC2 | `go test ./internal/slices -run SCDocBudget` + bench economy | pass | c6d7465 | `TestSCDocBudget` (honest_envelope, tight_budget_truncates_honestly, economy_reports_ratio) + `BenchmarkDocEconomy` reporta raw/doc=0.10 (doc 9691B/2423tok vs raw cat 932B/233tok no fixture micro; medido, sem threshold) |
| R1 | `go test -race ./...` (root + bench/) | fail | c6d7465 | baseline: B2; root `go test -race ./...` 100% verde e afetados com -race verdes. Módulo bench/ separado falha identicamente a B2 (nenhuma edição em bench/; meu teste foi redigido para não adicionar violação à quarantine); ficam para o R1 integrado do parent (SC5) |
| V1 | diff lido contra AC1–AC2 | pass | c6d7465 | aditivo puro: +1 tool no registry +1 ToolDefinition +1 handler `get_doc` (espelha `get_tour`); semântica de `Doc.Emit`/budget de 03 intacta; gofmt/vet/build verdes |

## Outcome

`get_doc` exposto no registry único + MCP (mesmo engine do CLI `doc --format json`, envelope orçado só com budget, erros honestos). `TestSCDocBudget` trava o budget honesto (`bytes/4`, `measured:false`, `estimated == tokens.Estimate(output)`) e `BenchmarkDocEconomy` mede doc vs cat cru no fixture: raw/doc=0.10 (doc 9691B contra 932B de fontes — micro-fixture de 4 arquivos; o doc estruturado custa mais que um único `cat`, reportado sem threshold). Sem regressão: root `-race` verde; falhas do módulo bench/ são pré-existentes na base.
