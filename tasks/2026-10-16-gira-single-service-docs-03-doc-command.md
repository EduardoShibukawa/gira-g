---
schema: task/1
tier: standard
status: verified
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
verified_at: "9bb3606"
depends_on:
  - 2026-10-16-gira-single-service-docs-02-intra-service
---

# Single-service docs 03: comando `doc` md + json (F3/SC3)

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Novo `gira doc <service-id> [--format md|json]` com o mesmo engine: `md` legível para humanos, `json` em envelope para agentes, ambos cobrindo os mesmos fatos (mesmos IDs). Plain path golden-tested; `html` fica como follow-up.

In scope: SC3 do spec.
Out of scope: F1–F2, F4; `html`.

## Acceptance criteria

1. AC1 = SC3: `doc --format md` e `--format json` cobrem os mesmos fatos (mesmos IDs de nodes/edges/evidence).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test ./cmd/gira/ ./internal/slices/` antes da mudança | pass | 23d5ea3 | base limpa; a falha pré-existente TestSC11 da época de 01 não reproduz |
| AC1 | `go test ./cmd/gira -run SCDocParity` | pass | 9bb3606 | md/json cobrem os mesmos fatos (mesmos IDs/locators); golden plain md byte-for-byte; banner+stale, usage errors e budget envelope em testes irmãos |
| R1 | `go test -race ./...` (root) | pass | 9bb3606 | todos os pacotes ok, sem falhas para atribuir a B1 |
| V1 | diff lido contra AC1 | pass | 9bb3606 | aditivo: `runDoc` + `Engine.GetDoc/Emit/RenderMarkdown` + goldens; `html` fora recusa honesta; aliasing de slice no `Emit` corrigido antes do commit |

## Outcome

`gira doc <service-id> --format md|json [--index <path>] [--checkout <dir>] [--budget <tokens>]`: md onboarding (entrypoints, símbolos, contratos providos/consumidos, imports intra/externo/stdlib, calls, run/test de manifests, where-to-change, tour, flows, Known Unknowns) e json no mesmo envelope orçado, mesmos fatos. Sem regressão; `html` segue follow-up.
