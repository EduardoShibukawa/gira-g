---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
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
| B1 | `go test ./internal/extractors/ ./internal/slices/` antes da mudança | not run | | |
| AC1 | `go test ./internal/slices -run SCDocIntraService` | not run | | |
| R1 | `go test -race ./...` (root) | not run | | |
| V1 | diff lido contra AC1 | not run | | |
