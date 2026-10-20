---
schema: task/1
tier: standard
status: verified
spec: 2026-10-20-gira-agent-context-surface@1
verified_at: "56062a8"
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
| B1 | `go test -race ./...` na base do filho | pass | 923b26a | verde antes do corte |
| SC4 | `go test ./internal/slices -run SCAgent -v -count=1` | pass | 56062a8 | TestSCAgentSummaryIntraBounded + TestSCAgentIntraDocKeepsFullDetail; JSON/caveman/md × summary/contracts/deep sob budget 400 sem vazar go_intra_service |
| AC1 | SC4 acima (fit 400 sem vazamento) | pass | 56062a8 | detalhe sintético de 400 funcs/pkgs/calls (~>400 KB) |
| AC2 | guarda de regressão do strip: cloneNode sem delete → suíte falha | pass | 56062a8 | verificado por mutação temporária (FAIL) e restauração (ok) |
| AC3 | GetIntraServiceDoc deep preserva detalhe completo | pass | 56062a8 | funcs/calls/packages = 400 |
| R1 | `go test -race ./...` | pass | 56062a8 | EXIT=0 |
| V1 | diff lido contra AC1–AC3; nenhum teste altera goldens/fixtures | pass | 56062a8 | diff = apenas arquivo novo de teste; git diff engine.go vazio |

## Outcome

Contrato G1 travado por testes: metadados go_intra_service permanecem dropped-able no summary (fit 400 em JSON/caveman/md), cloneNode é o ponto de strip guardado por mutação, e o deep continua servindo o detalhe via GetIntraServiceDoc. Merge local 56062a8.

## Constraints and budget

- Testes determinísticos, sem I/O de rede; fixtures sintéticos inline (não tocar `fixtures/`/`testdata/`).
- Honestidade do envelope (estimator/measured) verificada nas saídas dos testes.
