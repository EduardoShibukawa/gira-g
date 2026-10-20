---
schema: task/1
tier: standard
status: todo
spec: 2026-10-20-gira-agent-context-surface@1
parent: 2026-10-20-gira-agent-context-surface
depends_on: ["2026-10-20-gira-agent-context-surface-01-caveman-renderers"]
---

# CLI: --format=caveman em slice/impact/locate com validação estrita (SC1-CLI, SC3)

## Routing and approvals

- tier: standard (uma superfície de flag da CLI; sem mudança de contrato nos demais comandos por D2)
- approved: criação pelo dono via spec v1 aprovada (2026-10-20, instrução unattended)

## Objective

Aceitar `--format caveman` (alias `llm`) em `gira slice`, `gira impact` e `gira locate`; rejeitar valor inválido com erro nomeando `json|md|caveman`; atualizar usage/help dessas três flags. Demais comandos mantêm `ParseFormat` legado (D2). Testes de CLI no padrão `cmd/gira` (runCLI) para SC3 e para a saída real SC1/SC2 na superfície da CLI.

In scope: SC1 e SC2 na CLI; SC3 (validação + aliases); `cmd/gira/main.go` (runSlice/runImpact/runLocate), `internal/slices/budget.go` (`ParseFormatStrict`).

Out of scope: renderers (filho 01), regressão G1 (filho 03), erros de índice (filho 04), MCP, tour/flows/diff/explain/export.

## Acceptance criteria

- AC1 = SC3 (caveman aceito; inválido falha com `json|md|caveman`; `llm` é alias)
- AC2 = SC1/SC2 na superfície da CLI (saída sobre o índice dogfood fixture)
- AC3 = SC6 (raça/validator, goldens intocados)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| SC3 | `go test ./cmd/gira -run SCAgentFormatValidation -v -count=1` | not run | | |
| SC1/SC2 | `go test ./cmd/gira -run SCAgentCavemanCLI -v -count=1` | not run | | |
| R1 | `go test -race ./...` | not run | | |
| V1 | diff lido contra AC1–AC3 | not run | | |

## Constraints and budget

- `explain|tour|flows|diff|export` continuam com `ParseFormat` default-JSON (D2).
- humanBanner/emitHuman: caveman é texto plano, sem HTML; tratar como `md` na emissão.
- Sem mudança em ground truth; sem push.
