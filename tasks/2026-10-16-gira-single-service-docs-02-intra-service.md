---
schema: task/1
tier: standard
status: verified
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
verified_at: "89fda6f"
depends_on:
  - 2026-10-16-gira-single-service-docs-01-local-init
---

# Single-service docs 02: detalhe intra-serviço Go (F2/SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Trazer o subset determinístico do Graphify para Go: símbolos (packages, types, funcs) + `imports`/`calls` com locator file+line, `origin` discreta e IDs estáveis, sem colapso de edges. Sem extrator vira `unsupported`/`unknown`, nunca chute.

In scope: SC2 do spec.
Out of scope: F1, F3–F4; outras linguagens (follow-up).

## Acceptance criteria

1. AC1 = SC2: `doc` (engine) de um serviço Go lista entrypoints, símbolos, contratos e imports intra-módulo vs externos, cada item com locator; resto em Known Unknowns.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test ./internal/extractors/ ./internal/slices/` antes da mudança | pass | c31c69c | árvore limpa; `go test ./...` root também verde (17 pkgs, exit 0) — sem falhas pré-existentes neste worktree |
| AC1 | `go test ./internal/slices -run SCDocIntraService` | pass | 89fda6f | entrypoints, símbolos, contratos providos/consumidos, imports intra vs externos, locator `[repo@commit:path#Lstart-Lend]` em tudo, unknowns em Known Unknowns |
| R1 | `go test -race ./...` (root) | pass | 89fda6f | 17 pkgs, exit 0. No caminho achou 3 regressões do primeiro corte (SC6 budget, SC11 golden, eval baseline) pelo blob vazar nas views derivadas; fix servindo o detalhe só via GetIntraServiceDoc (cloneNode omite a chave); re-rodado verde no código final |
| V1 | diff lido contra AC1 | pass | 89fda6f | hints+resolve aditivos, nenhum nó/edge novo no pipeline default; `gofmt` limpo (drift pré-existente em self_fidelity_golden_test.go intocado); `go build ./...` OK |

## Outcome

F2 fase 1 (Go) pronta: hints de símbolos no extrator Go + montagem no resolve step gravando `metadata["go_intra_service"]` validado por serviço; engine expõe `GetIntraServiceDoc`. Sem regressão: nenhum nó/edge novo, envelopes e goldens intactos.
