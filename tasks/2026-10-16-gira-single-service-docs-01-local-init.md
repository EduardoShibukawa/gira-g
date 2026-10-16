---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
---

# Single-service docs 01: `init` gera manifest + indexa (F1/SC1)

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

`gira init [dir] [--force]` gera `gira.yaml` a partir do repo local (HEAD como SHA pinado) quando ausente — `--force` repina — e indexa em seguida. `gira index` com manifest explícito funciona como antes e nunca escreve arquivos.

In scope: SC1 do spec.
Out of scope: F2–F4.

## Acceptance criteria

1. AC1 = SC1: `gira init <dir>` sem `gira.yaml` gera manifest pinando HEAD + produz `macro-map.json`; com manifest explícito, `gira index` inalterado; `--force` repina o SHA.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test ./cmd/gira/` antes da mudança | not run | | |
| AC1 | `go test ./cmd/gira -run SCSingleInit` | not run | | |
| R1 | `go test -race ./...` (root) | not run | | |
| V1 | diff lido contra AC1 | not run | | |
