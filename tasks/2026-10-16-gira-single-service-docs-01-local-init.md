---
schema: task/1
tier: standard
status: verified
parent: 2026-10-16-gira-single-service-docs
spec: 2026-10-16-gira-single-service-docs@1
verified_at: "8994ece"
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
| B1 | `go test ./cmd/gira/` antes da mudança | fail | 03f12f9 | pré-existente: TestSC11GiraQueryFallbackGolden (impact-checkout golden difere); sem relação com 01, nenhuma edição feita ainda |
| AC1 | `go test ./cmd/gira -run SCSingleInit` | pass | 8994ece | gera manifest pinando HEAD + indexa; reuso sem --force; repin com --force; erro honesto fora de git |
| R1 | `go test -race` nos pacotes afetados (`cmd/gira manifest indexer gitcache`) | fail | 8994ece | baseline: B1; única falha é TestSC11, a mesma de B1. manifest/indexer/gitcache passam. Full-root tem ainda 2 falhas fora do escopo deste child (build de internal/extractors quebrado pelo worktree sujo do outro task + cosign ausente no env) — ficam para o R1 integrado do parent (SC5) |
| V1 | diff lido contra AC1 | pass | 8994ece | `init` novo + refactor puro de `runIndex` para `indexManifest`; `runServe` intacto; build OK |

## Outcome

`gira init [dir] [--force]` gera `gira.yaml` (repo `local`, HEAD pinado) e indexa pelo mesmo caminho de `index`. Sem regressão nos pacotes afetados além da falha pré-existente B1.
