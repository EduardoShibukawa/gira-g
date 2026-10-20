---
schema: task/1
tier: heavy
status: todo
spec: 2026-10-20-gira-agent-context-surface@1
---

# Agent context surface (ACS): caveman, teto summary com intra-service e erros de índice legíveis

## Routing and approvals

- tier: heavy (H1: formato de saída para agentes é decisão de produto; H3: span `internal/slices`, `cmd/gira`, `internal/store` e docs consumidos pelo pi-harness/bench; H5: 5 filhos, multi-sessão)
- confirmed: heavy by owner (2026-10-20, instrução explícita "faça tudo sem supervisão… quero TODO PLANO EXECUTADO e TODAS TASKS", dono autorizou aprovação da spec v1 sem gate interativo)
- approved: spec v1 by owner (2026-10-20, autorização unattended da sessão de orquestração multi-repo; validação final do resultado fica para o dono ao acordar)

## Objective

Implementar a INTENT 0001 na parte dos gaps G1/G2/G3 do dogfooding Pi Harness × Gira-G: formato `--format=caveman` em `slice`, `impact` e `locate` com gramática 1-fato-por-linha e orçamento via `fit` existente; guarda de regressão do teto do summary para nós Go com `go_intra_service` grande (dropped-able no summary, completo no deep via `GetIntraServiceDoc`); e erros `index_corrupt` de discovery nomeando o arquivo e sugerindo `--index <path>`.

In scope: spec `2026-10-20-gira-agent-context-surface` v1, SC1–SC6, filhos listados no Plan, docs da CLI no site, rebuild local do binário.

Out of scope: `gira symbols`, `--errors-only`, Breaking Changes Gate (INTENT 0002), paridade MCP do formato (D4/F1), ground truth (`fixtures/`, `testdata/`), qualquer modelo (ADR 0001), push/PR/release/tag.

## Acceptance criteria

- AC1 = SC1–SC6 of `specs/2026-10-20-gira-agent-context-surface.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` + validator antes dos filhos | pass | 3541cf9 | root ok; validator pass |
| AC1 | SC1–SC6 integrated at the final head | pass | 876c868 | filhos 01–04 verified; -05 integration: race 17 pkgs, validator pass, baterias SCAgent/SCIndexCorrupt reexecutadas, docs site, gira 0.2.0 instalado no PATH |

## Plan

Filhos criados em @1, em ordem de dependência, test-first, um worktree/branch e um commit por filho, merge `--no-ff` na main local:

1. `-01-caveman-renderers`: SC1 (parcial, library), SC2 (parcial, library). `FormatCaveman`, `RenderCaveman` de `ContextSlice`/`ImpactBriefing`/`BoundaryResult`, wiring nos `Emit`. Sem dependências.
2. `-02-caveman-cli-surface`: SC1 (CLI), SC3. Flags `--format=caveman|llm` em slice/impact/locate com validação estrita (`ParseFormatStrict`) e help. Depende de 01.
3. `-03-summary-intra-regression`: SC4. Guarda de regressão do strip de `go_intra_service` no summary sob `--budget 400` em JSON e caveman; deep preserva via `GetIntraServiceDoc`. Depende de 01.
4. `-04-corrupt-index-errors`: SC5. `internal/store` (resolver Branch 2 e publish) nomeia o arquivo e sugere `--index`; código `index_corrupt` preservado. Sem dependências com 01–03.
5. `-05-integrated-docs-surface`: SC6 e integração. `go test -race ./...`, validator, docs do site (`slice`/`impact`/`locate`), registro de versão da build local. Depende de 01–04.

Rollback: `git revert` por filho.

## Review

- [F1] acceptance do dono pendente (validação ao acordar, 2026-10-20): G1 era do binário stale, sanado na main por `cloneNode` (`89fda6f`) — esta spec o travou por testes; a rebuild local é a correção operacional
- [F2] paridade MCP do formato (F1/D4) e os itens da INTENT 0001 fora do escopo (`symbols`, `--errors-only`, gate da INTENT 0002) permanecem fases futuras
