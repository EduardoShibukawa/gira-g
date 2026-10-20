---
schema: task/1
tier: standard
status: todo
spec: 2026-10-20-gira-agent-context-surface@1
parent: 2026-10-20-gira-agent-context-surface
depends_on: ["2026-10-20-gira-agent-context-surface-01-caveman-renderers"]
---

# Guarda de regressão: summary de nó Go com go_intra_service grande cabe em --budget 400 (SC4)

## Routing and approvals

- tier: standard (testes de guarda em `internal/slices`; o corte de código já existe em `cloneNode`/`89fda6f` — D5)
- approved: criação pelo dono via spec v1 aprovada (2026-10-20, instrução unattended)

## Objective

Travar com testes o contrato G1: (a) `slice --tier summary` de um Service cujo índice carrega `go_intra_service` volumoso renderiza em JSON e em caveman sob `--budget 400` sem a chave `go_intra_service` na saída; (b) `cloneNode` strippa `domain.IntraServiceMetadataKey` de focal e vizinhos (teste que falha se o delete sumir); (c) o tier deep continua servindo o detalhe completo via `GetIntraServiceDoc`. Padrão dos testes `Fit`/`SC6` de `internal/slices`, construído in-memory com macro-map sintético.

In scope: SC4; testes em `internal/slices` (arquivo novo `agent_budget_test.go` ou extensão de `intra_service_test.go`/`budget_test.go`).

Out of scope: mudança de comportamento de produção (nenhuma — se um teste falhar por comportamento ausente, implementar o mínimo no mesmo filho e registrar), CLI (filho 02), erros de índice (filho 04).

## Acceptance criteria

- AC1 = SC4(a): summary JSON e caveman sob budget 400 com intra-service grande, sem a chave no output
- AC2 = SC4(b): regressão explícita do strip em `cloneNode`
- AC3 = SC4(c): deep preserva o detalhe completo via `GetIntraServiceDoc`

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| SC4 | `go test ./internal/slices -run SCAgentSummaryIntraBounded -v -count=1` | not run | | |
| R1 | `go test -race ./...` | not run | | |
| V1 | diff lido contra AC1–AC3; nenhum teste altera goldens/fixtures | not run | | |

## Constraints and budget

- Testes determinísticos, sem I/O de rede; fixtures sintéticos inline (não tocar `fixtures/`/`testdata/`).
- Honestidade do envelope (estimator/measured) verificada nas saídas dos testes.
