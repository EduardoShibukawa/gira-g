---
schema: task/1
tier: standard
status: verified
spec: 2026-10-20-gira-agent-context-surface@1
verified_at: "2ae7757"
parent: 2026-10-20-gira-agent-context-surface
depends_on: []
---

# Caveman renderers e wiring de orçamento em internal/slices (SC1, SC2 — library)

## Routing and approvals

- tier: standard (escopo em `internal/slices`, uma gramática definida pela spec aprovada; abaixo dos triggers heavy)
- approved: criação pelo dono via spec v1 aprovada (2026-10-20, instrução unattended)

## Objective

Adicionar `FormatCaveman` (`caveman`, alias CLI `llm` tratado no filho 02) e os renderers 1-fato-por-linha `RenderCaveman` de `ContextSlice`, `ImpactBriefing` e `BoundaryResult` conforme a gramática da spec, e ligá-los nos closures `render` de `Emit` para que `--budget`/`fit` (drop order, marker `Budget:`, honestidade `estimator: "bytes/4"`/`measured: false`) se apliquem ao novo formato. Testes SC1/SC2 no nível de biblioteca, padrão dos testes `Fit`/`SC6` existentes.

In scope: SC1 (gramática determinística), SC2 (orçamento via fit) na biblioteca; `internal/slices/budget.go`, `internal/slices/render.go`, `internal/slices/engine.go` (Emit de BoundaryResult).

Out of scope: flags da CLI (filho 02), regressão G1 (filho 03), erros de índice (filho 04), MCP, docs, ground truth, push.

## Acceptance criteria

- AC1 = SC1 (gramática e determinismo biblioteca)
- AC2 = SC2 (fit/orçamento no formato)
- AC3 = SC6 (raça/validator sem regressão)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` antes da mudança | pass | 3541cf9 | verde antes do corte |
| SC1 | `go test ./internal/slices -run SCAgentCavemanGrammar -v -count=1` | pass | 2ae7757 | gramática exata byte-a-byte + determinismo slice/briefing/locate |
| SC2 | `go test ./internal/slices -run SCAgentCavemanBudget -v -count=1` | pass | 2ae7757 | truncagem por fato inteiro; marker honesto (estimated==bytes); componente médio <400 sem budget estourado |
| AC1 | SC1 coberto por TestSCAgentCavemanGrammar (go test ./internal/slices -run SCAgentCaveman -count=1) | pass | 2ae7757 | gramática exata slice/briefing/locate + determinismo |
| AC2 | SC2 coberto por TestSCAgentCavemanBudget | pass | 2ae7757 | truncagem por fato inteiro; marker honesto; componente médio <400 |
| AC3 | R1 `go test -race ./...` + validator | pass | 2ae7757 | 17 pacotes ok |
| R1 | `go test -race ./...` | pass | 2ae7757 | 17 pacotes ok |
| V1 | diff lido contra AC1–AC3 e Interfaces da spec | pass | 2ae7757 | aditivo; JSON/MD intocados; ParseFormat legado preservado (D2) |

## Constraints and budget

- Aditivo: nenhum byte muda em JSON/MD; `ParseFormat` legado intocado para os demais formatos.
- `fit`, settle do envelope e marker `Budget:` reutilizados sem fork de lógica.
- Sem mudança em `fixtures/`/`testdata/`; sem push/release.

## Outcome

Renderers caveman (1-fato-por-linha) de ContextSlice/ImpactBriefing/BoundaryResult ligados ao fit existente; FormatCaveman + alias no ParseFormat sem quebrar consumidores legados. Merge local `2ae7757` (--no-ff da branch acs/01-caveman-renderers, 1 commit).
