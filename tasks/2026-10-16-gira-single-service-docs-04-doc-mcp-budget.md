---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
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
| B1 | `go test ./internal/mcp/ ./internal/slices/` antes da mudança | not run | | |
| AC1 | MCP `get_doc` parity check | not run | | mesmo engine do CLI |
| AC2 | `go test ./internal/slices -run SCDocBudget` + bench economy | not run | | |
| R1 | `go test -race ./...` (root + bench/) | not run | | |
| V1 | diff lido contra AC1–AC2 | not run | | |
