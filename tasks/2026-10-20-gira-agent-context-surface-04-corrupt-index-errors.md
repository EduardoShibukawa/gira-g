---
schema: task/1
tier: standard
status: verified
spec: 2026-10-20-gira-agent-context-surface@1
verified_at: "8af2678"
parent: 2026-10-20-gira-agent-context-surface
depends_on: []
---

# Erros index_corrupt de discovery nomeiam o arquivo e sugerem --index (SC5)

## Routing and approvals

- tier: standard (mensagens de erro do `internal/store` na discovery com manifesto; contrato `index_corrupt` preservado)
- approved: criação pelo dono via spec v1 aprovada (2026-10-20, instrução unattended)

## Objective

No ramo de discovery com manifesto do resolver (`internal/store/resolver.go`, Branch 2) e nos carregamentos de versão (`internal/store/publish.go`), todo erro `ErrCorruptIndex` produzindo falha de discovery deve: (a) nomear o caminho completo do arquivo em causa (`.head`, `meta.json`, `macro-map.json`), (b) sugerir `re-index with 'gira index'` e `--index <path>`, (c) manter o código estruturado `index_corrupt` legível por `errors.Is`/prefixo. Testes em `internal/store` (cache sintético corrompido por tipo de arquivo) e um teste de CLI assertando stderr.

In scope: SC5; `internal/store/resolver.go`, `internal/store/publish.go` (apenas onde o caminho não é nomeado), testes correspondentes.

Out of scope: mudança de política de resolução/linhagem, schema, publicação, CLI flags novas, os filhos 01–03.

## Acceptance criteria

- AC1 = SC5: meta.json ilegível, macro-map com hash divergente e `.head` truncado produzem erro com caminho completo + `--index` na mensagem, código `index_corrupt`
- AC2 = regressos existentes do resolver/store continuam passando sem edição de expectativas anteriores (novas expectativas apenas adicionadas)
- AC3 = SC6 (raça/validator)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` na base do filho | pass | 36ba26b | verde antes do corte |
| SC5 | `go test ./internal/store ./cmd/gira -run SCIndexCorrupt -count=1` | pass | 8af2678 | 4 casos store (empty .head, meta ilegível, meta ausente, macro-map corrompido) + CLI stderr |
| AC1 | SC5: caminhos absolutos nomeados + `--index <path>` + código `index_corrupt` preservado | pass | 8af2678 | TestSCIndexCorruptErrorDiscoveryNamesFile + TestSCIndexCorruptErrorCLI |
| AC2 | suíte store/cmd sem edição de expectativas anteriores | pass | 8af2678 | apenas arquivos novos |
| AC3 | SC6: `go test -race ./...` | pass | 8af2678 | EXIT=0 |
| R1 | `go test -race ./...` | pass | 8af2678 | |
| V1 | diff lido contra AC1–AC3 | pass | 8af2678 | corruptHint usa %w: errors.Is(ErrCorruptIndex) preservado |

## Outcome

Discovery com manifesto: todo ErrCorruptIndex do Branch 2 nomeia o arquivo absoluto e sugere `--index <path>`; código estruturado intacto; teste CLI na superfície. Merge local 8af2678.

## Constraints and budget

- Mensagens: sufixo único reutilizável, sem quebrar `errors.Is(ErrCorruptIndex)`.
- O erro ambíguo existente (`index_ambiguous … pass --index <path>`) já está correto — não duplicar.
- Sem ground truth; sem push.
