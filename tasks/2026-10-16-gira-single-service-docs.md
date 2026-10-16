---
schema: task/1
tier: heavy
status: active
spec: 2026-10-16-gira-single-service-docs@1
---

# GIRA single-service docs (INTENT: single-repo first-class + doc humano/agente)

## Routing and approvals

- tier: heavy (H1: product calls em granularidade/formato/zero-config; H3: manifest/indexer→slices→CLI/MCP; H5: quatro frentes)
- confirmed: heavy by owner (2026-10-16, "sim")

## Objective

Fazer o gira-g servir um serviço sozinho tão bem quanto serve o multi-service (plus): `index .` zero-config, detalhe intra-serviço determinístico, e `gira doc <service>` gerando o mesmo onboarding para humanos (md/html) e agentes (json/MCP) com budget medido.

In scope: F1–F4 do spec `2026-10-16-gira-single-service-docs@1`.
Out of scope: novas linguagens, rewrites de extratores, skills/release, docs site, ground-truth changes (ver spec).

## Acceptance criteria

SC1–SC5 do spec.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test -race ./...` (root) antes de qualquer child | not run | | |
| B2 | `go test -race ./...` (bench/) antes de qualquer child | not run | | |
| B3 | `go test ./internal/eval -run EndToEnd -v` antes (R/P published/perturbed) | not run | | |
| SC1 | `go test ./cmd/gira -run SCSingleInit` | not run | | F1 init gera manifest + indexa |
| SC2 | `go test ./internal/slices -run SCDocIntraService` | not run | | F2 detalhe com locators |
| SC3 | `go test ./cmd/gira -run SCDocParity` | not run | | F3 md/json/html |
| SC4 | `go test ./internal/slices -run SCDocBudget` + bench economy | not run | | F4 economia medida |
| SC5 | regressão integrada (B1+B2+B3 no head integrado, validator) | not run | | |
| R1 | `go test -race ./...` root+bench no head integrado | not run | | |
| V1 | diff integrado lido contra SC1–SC5 | not run | | |

## Plan

Spec em `draft` v1 — sem children até aprovação (spec gate). Após `approved: spec v1`:

1. `2026-10-16-gira-single-service-docs-01-local-init`: F1 `init` gera manifest + indexa.
2. `2026-10-16-gira-single-service-docs-02-intra-service`: F2 detalhe intra-serviço.
3. `2026-10-16-gira-single-service-docs-03-doc-command`: F3 `doc` CLI + renderers md/html/json.
4. `2026-10-16-gira-single-service-docs-04-doc-mcp-budget`: F3 MCP `get_doc` + F4 budget/bench.
5. Integrado: SC1–SC5 + R1 + V1 + review.

## Review

skip: roda após children verificados
