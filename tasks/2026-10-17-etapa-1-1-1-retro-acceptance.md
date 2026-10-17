---
schema: task/1
tier: standard
status: accepted
verified_at: "61b58c2"
---

# Gira-G: aceite retroativo da Etapa 1 e 1.1 (`4c2204c`, `0d5f248`) em `61b58c2`

## Routing and approvals

- tier: standard (S1 revisão retroativa de dois commits já em `main`, sem mudança de código nesta tarefa; sem sinais heavy)
- decided: N2 (aceite da Etapa 1/1.1) → (b) revisão adversarial somente leitura de `aa4e6fe..0d5f248` (exclui `aa4e6fe`), depois aceite (a) se não houver bloqueio material by owner (2026-10-17, owner, 2026-10-17 (session decision message))
- accepted: by owner (2026-10-17, revisão adversarial concluída sem bloqueios materiais)

O aceite formal do dono está registrado. Os achados não bloqueantes permanecem como riscos residuais e follow-ups separados; este aceite não autoriza as etapas posteriores nem altera decisões arquiteturais abertas.

## Objective

Reunir a evidência para o aceite formal da Etapa 1 (`4c2204c`) e da Etapa 1.1 (`0d5f248`) no estado de `main` em `61b58c2`. Escopo do diff revisado: `aa4e6fe..0d5f248` (`4c2204c` e `0d5f248`), mais `0d5f248..61b58c2`, que só muda `cmd/gira/main_test.go` (PR #17).

In scope:
- critérios (a)–(e) abaixo, revisados por adversarial read-only (agente `GiraN2Review`, veredito: sem bloqueios materiais).

Out of scope:
- `aa4e6fe` (M0–M6): frente separada, `tasks/2026-10-17-m0-m6-retro-review.md`;
- correção dos achados não bloqueantes (ver `## Outcome`).

## Acceptance criteria

- AC1 [inferred]: (a) `init --force` não autoriza perda de símbolos.
- AC2 [inferred]: (b) `Save` e `SaveFile` são seguros por padrão (gate e proteção contra índice corrompido).
- AC3 [inferred]: (c) publicação atômica e durável (temporário no mesmo diretório, `Sync`, `Close`, `Rename`, fsync do diretório pai).
- AC4 [inferred]: (d) o manifesto é descoberto a partir de subdiretório.
- AC5 [inferred]: (e) testes verdes no SHA final.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | CI de `main` no commit `0d5f248`, run 39067269222 | fail | 0d5f248 | test (windows-latest) `main_test.go:641`: defeito de isolamento do teste (`HOME` sem `USERPROFILE`), corrigido na PR #17 |
| AC1 | `main.go:563-566`, `main.go:541-546`; `init_test.go:132 TestInit_ForceDoesNotAllowSymbolLoss` | pass | 61b58c2 | revisão de código e teste; ver nota de Ref abaixo |
| AC2 | `store.go:49-52,67-70,106-116`; `TestStore_PublishEnforcesSymbolLossGateAndPreservesPrevious`, `TestStore_CorruptIndexPreservedUntilRecoverCorrupt` | pass | 61b58c2 | idem |
| AC3 | `store.go:132,136-141,145,149,162,171,182-184`; `TestStore_PublishAtomicPreservesPreviousAndCleansTmp` | pass | 61b58c2 | idem; lacuna de teste em N3 |
| AC4 | `main.go:278-297,314-334`; `main_test.go:574 TestCLI_LoadIndexResolutionAndFailClosed` (casos 1b e 2, de `projDir/cmd/subpkg`) | pass | 61b58c2 | idem; lacuna de teste em N3 |
| AC5 | `go test -race -count=1 ./...` (go1.24.0 linux/amd64): exit 0, 17 pacotes ok, 1 sem testes; CI 3 SO verde em `61b58c2` (plano §3.3, run 39072167231) | pass | 61b58c2 | local rodado só no Linux, no HEAD `186427e`; entre `61b58c2` e `186427e` mudaram só 5 arquivos de documentação, o código e os testes são os mesmos |
| R1 | `go vet ./...`: exit 0 | pass | 61b58c2 | rodado no HEAD `186427e` (diff só de documentação) |
| V1 | diff `aa4e6fe..0d5f248` e `0d5f248..61b58c2` lidos contra AC1–AC5; `0d5f248..61b58c2` só altera `cmd/gira/main_test.go` e as asserções dos casos 1, 1b, 2 e 3 não mudaram | pass | 61b58c2 | revisão `GiraN2Review`; macOS e Windows só via CI |

## Outcome

Veredito da revisão adversarial: **sem bloqueios materiais**. O dono aceita formalmente a Etapa 1 e a Etapa 1.1 no estado `61b58c2`; status `accepted`. Os nove achados não bloqueantes abaixo permanecem follow-ups e não impedem este aceite.

Achados não bloqueantes, tratados como riscos residuais e follow-ups. Nada foi corrigido e nenhuma decisão aberta foi mudada:

| # | Sev. | Achado | Menor correção | Dependência |
|---|---|---|---|---|
| N1 | P2 | `loadIndex` lê `.gira/indexes/<sys>` (raiz e CWD) antes de `~/.cache/gira/indexes/<sys>` (`main.go:321-331`), mas `indexManifest` grava no home (`main.go:474-479`): cópia local antiga esconde o índice novo, saída 0 | pôr o candidato do home antes dos locais | ordem final depende de **D2** (aberta) |
| N2 | P2 | `gira index` sem argumento grava caminho absoluto em `Locator.Path` (`main.go:288-290,428-429`; `pipeline.go:356,422`); mapa varia com o diretório do clone | usar o caminho relativo/basename | **D3** (aberta) |
| N3 | P3 | lacunas de teste: (c) nada falha se o caminho padrão deixar de chamar `syncDirectory`; (d) descoberta por subdiretório dentro de `runIndex` (`main.go:428-429`) sem teste | testes dedicados | — |
| N4 | P3 | `DeletedPaths` nunca preenchido pela CLI (falha fechado); `AllowShrink` repassado a `indexer.Build` (`main.go:459`) não é usado por `Build` | preencher `DeletedPaths` ou documentar; remover o repasse | — |
| N5 | P3 | diretórios criados por `MkdirAll` (`store.go:123-126`) sem fsync no avô; só na primeira publicação | fsync do avô | — |
| N6 | P3 | falha de fsync do diretório após o rename devolve erro com o índice novo já no lugar (`store.go:171-173`; documentado em `80-82`) | manter e documentar | — |
| N7 | P3 | `--recover-corrupt` vale para qualquer falha de `LoadFile` (EACCES, erro de leitura), sem gate (`store.go:107-110`) | restringir a corrupção | — |
| N8 | P3 | busca de `gira.yaml` sobe até `/` (`~/gira.yaml` esquecido captura subdiretórios); URLs relativas resolvem do CWD (`cache.go:99`), preexistente | parar na fronteira do repositório | — |
| N9 | P3 | `cwdCandidate != sysCandidates[0]` (`main.go:323`) compara relativo com absoluto: sempre verdadeiro, `Stat` duplicado | normalizar antes de comparar | — |

Windows [INFERÊNCIA]: sem trava, um leitor concorrente pode fazer o `os.Rename` falhar; o arquivo anterior fica intacto.

Próximo passo: resolver as questões arquiteturais ainda abertas, começando por N3 e D1, sem iniciar implementação até autorização heavy específica.

## Checkpoint

- as of: 2026-10-17, branch `work/2026-10-17-gira-n1-n2-o6-decisions`, base `186427e`
- done: revisão adversarial concluída (sem bloqueios materiais); aceite formal registrado
- next: decidir N3/D1 e demais questões arquiteturais abertas; implementação continua não autorizada
- open: N1–N9 da revisão como follow-ups (N1 depende de D2, N2 de D3)
- revalidate: as linhas citam `61b58c2`; rodar de novo se o código de `main` mudar
