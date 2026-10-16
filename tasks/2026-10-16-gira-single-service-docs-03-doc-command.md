---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
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
| B1 | `go test ./cmd/gira/ ./internal/slices/` antes da mudança | not run | | |
| AC1 | `go test ./cmd/gira -run SCDocParity` | not run | | |
| R1 | `go test -race ./...` (root) | not run | | |
| V1 | diff lido contra AC1 | not run | | |
