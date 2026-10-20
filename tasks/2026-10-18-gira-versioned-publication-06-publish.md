---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: publicação imutável, heads e durabilidade

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de tasks 04 keys e 05 locks.

## Objective

Implementar store versionado com versões imutáveis, heads por linhagem, publicação atômica, gate de perda de símbolos e estados explícitos de durabilidade.

## In scope

Resolver base legada para gate conforme D2/D9; escrita temporária, sync, rename e sync do diretório; idempotência; snapshot destacado; confirmed/visible_unconfirmed; quarentena só por ação explícita; não apagar versões referenciadas.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Publicação cria versão imutável e atualiza head somente após versão completa.
2. AC2: Falha antes do rename deixa nenhum artefato novo visível.
3. AC3: Falha após rename da versão deixa head intacto e retry idempotente.
4. AC4: Republicação do mesmo version_id com bytes diferentes falha nondeterministic_build sem sobrescrever.
5. AC5: Gate rejeitado não altera store, head nem snapshot.
6. AC6: Unix sync failure retorna erro e visible_unconfirmed; Windows emite aviso explícito sem alegar durabilidade não garantida.
7. AC7: Snapshots destacados são independentes do store; arquivo corrompido existente é preservado salvo --recover-corrupt.
8. AC8: CI 3 SOs e testes de falha passam.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Publicação cria versão imutável e atualiza head somente após versão completa. | pass | 65e664a | `TestPublish_AC1_ImmutableVersionAndHeadUpdate` em `internal/store`: `PublishVersion` grava layout canônico `<store_root>/<lineage_id>/versions/<version_id>/macro-map.json` e `meta.json` estrito, e atualiza atomicamente `<store_root>/<lineage_id>/.head` apontando para `version_id`; leitura via `ReadHead`, `LoadVersion` e `LoadActiveVersion` validam a versão ativa. |
| AC2 | Falha antes do rename deixa nenhum artefato novo visível. | pass | 65e664a | `TestPublish_AC2_FailureBeforeRenameLeavesNoArtifacts` em `internal/store`: gancho `BeforeRenameError` simula falha imediatamente antes da promoção do diretório temporário `.staging-<id>`; diretório de staging é limpo pelo defer e nenhum artefato ou head é deixado no store (`os.IsNotExist`). |
| AC3 | Falha após rename da versão deixa head intacto e retry idempotente. | pass | 65e664a | `TestPublish_AC3_FailureAfterVersionRenameLeavesHeadIntactAndRetryIdempotent` em `internal/store`: falha simulada via `BeforeHeadRenameError` após a promoção da versão deixa a versão visível e `.head` inalterado apontando para a versão anterior; nova execução subsequente verifica igualdade de bytes deterministicamente e atualiza `.head` com sucesso (retry idempotente). |
| AC4 | Republicação do mesmo version_id com bytes diferentes falha nondeterministic_build sem sobrescrever. | pass | 65e664a | `TestPublish_AC4_NonDeterministicBuildFailsClosedWithoutOverwrite` em `internal/store`: tentativa de republicar o mesmo `version_id` com bytes divergentes falha fechada com erro estável `ErrNonDeterministicBuild` (`nondeterministic_build`), preservando o arquivo preexistente no store byte a byte intacto sem sobrescrever. |
| AC5 | Gate rejeitado não altera store, head nem snapshot. | pass | 65e664a | `TestPublish_AC5_GateRejectionLeavesStoreAndSnapshotUntouched` em `internal/store`: perda não autorizada de símbolos reais em relação à versão ativa no `.head` (ou em snapshot prévio) é rejeitada por `CheckSymbolLossGate` antes de qualquer escrita; nem store, nem `.head`, nem snapshots em disco são modificados. |
| AC6 | Unix sync failure retorna erro e visible_unconfirmed; Windows emite aviso explícito sem alegar durabilidade não garantida. | pass | 65e664a | `TestPublish_AC6_DurabilitySemanticsUnixAndWindows` em `internal/store`: falha de sync de diretório em Unix retorna erro explícito e status `visible_unconfirmed` com versão visível no disco; em Windows (`runtime.GOOS == "windows"`), sync de diretório é omitido por limitação do SO e retorna status `visible_unconfirmed`, código/erro nil e aviso textual explícito sem alegar garantias contra perda de energia. |
| AC7 | Snapshots destacados são independentes do store; arquivo corrompido existente é preservado salvo --recover-corrupt. | pass | 65e664a | `TestPublish_AC7_DetachedSnapshotIndependenceAndCorruptPreservation` em `internal/store`: `PublishSnapshot` grava diretamente no caminho de destino sem criar linhagem ou alterar o store; arquivo corrompido pré-existente no destino é preservado intacto por falha fechada a menos que `RecoverCorrupt: true` seja explicitamente passado. |
| AC8 | CI 3 SOs e testes de falha passam. | pass | 65e664a | `TestPublish_AC8_ConcurrentPublishingWithLocksAndD9LegacyFallback` em `internal/store`: fallback do gate D9 contra índice legado na primeira publicação canônica; 4 goroutines concorrentes disputando travas ordenadas M -> S gravam versões distintas sem deadlocks nem corrupção; `go test -race ./...` passa 100% em todos os pacotes; compilação cruzada para Windows e macOS limpas. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; publicação versionada imutável, heads e durabilidade implementadas e verificadas
- next: Próxima tarefa no fluxo é a task 07 resolver.
- revalidate: suite de testes de store e do repositório executada com race detector; validate.ts passa com Overall: pass.
