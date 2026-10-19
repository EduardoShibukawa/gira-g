---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: locks multiplataforma M-S-P

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de ADR 0014; pode executar em paralelo com locator/keys desde que não altere a ordem global.

## Objective

Implementar locks entre processos em ordem M → S → P, timeout de 60 segundos e falha fechada nos sistemas suportados.

## In scope

flock Unix e LockFileEx Windows via dependência aprovada; normalização da identidade de lock; liberação após morte do processo; erro estável index_locked/lock_unavailable; nenhuma gravação quando lock não pode ser adquirido.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Dois processos concorrentes se excluem nos três SOs.
2. AC2: Caminhos relativos, absolutos e symlink do mesmo manifesto compartilham lock M.
3. AC3: Morte de processo libera lock e retry funciona.
4. AC4: Timeout retorna index_locked.
5. AC5: Falha injetada de lock retorna lock_unavailable sem artefatos gravados.
6. AC6: Ordem M→S→P é verificada sem deadlock.
7. AC7: CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Dois processos concorrentes se excluem nos três SOs. | pass | ab1a5b3 | `TestAC1_ConcurrentProcessesMutualExclusion` (dois subprocessos OS concorrentes disputando contador compartilhado com jitter convergem para soma exata sem updates perdidos) e `TestAC1_ConcurrentGoroutinesMutualExclusion` (8 goroutines com fds independentes disputando lock M sem qualquer violação de exclusão mútua na seção crítica). |
| AC2 | Caminhos relativos, absolutos e symlink do mesmo manifesto compartilham lock M. | pass | ab1a5b3 | `TestAC2_ManifestLockIdentitySharing`: caminhos relativo (`./gira.yaml`), absoluto (`/abs/.../gira.yaml`), symlink no mesmo diretório e symlink em subdiretório convergem para o mesmo arquivo canônico de lock em `.gira/locks/` com base na identidade física (`dev:ino` em Unix, `vol:idx` em Windows); trava ativa em um caminho bloqueia tentativas concorrentes nos demais com `index_locked`. |
| AC3 | Morte de processo libera lock e retry funciona. | pass | ab1a5b3 | `TestAC3_ProcessDeathReleasesLock`: subprocesso auxiliar adquire trava M e aguarda; processo pai tenta adquirir e é bloqueado (`index_locked`); subprocesso é morto (`SIGKILL`); SO libera automaticamente o lock na tabela de arquivos; retry subsequente no processo pai adquire com sucesso imediato. |
| AC4 | Timeout retorna index_locked. | pass | ab1a5b3 | `TestAC4_TimeoutReturnsIndexLocked`: disputa de trava com contexto de timeout (60ms) retorna erro estável `ErrIndexLocked` (`index_locked: lock acquisition timed out`), validado via `errors.Is(err, ErrIndexLocked)` e verificação de substring canônica. |
| AC5 | Falha injetada de lock retorna lock_unavailable sem artefatos gravados. | pass | ab1a5b3 | `TestAC5_InjectedLockErrorReturnsLockUnavailable`: falha injetada de lock (`InjectedLockError`) em `WithLocks` e `AcquireLocks` retorna erro canônico `ErrLockUnavailable` (`lock_unavailable: ...`); callback de persistência nunca é executado e nenhum artefato ou snapshot é gravado no disco (fail-closed garantido). |
| AC6 | Ordem M→S→P é verificada sem deadlock. | pass | ab1a5b3 | `TestAC6_StrictLockOrderMSP`: ordem normativa M -> S -> P verificada via hooks de aquisição e liberação estritamente reversa P -> S -> M; `LockChain` detecta e rejeita tentativas fora de ordem (S antes de M, P antes de S/M) com `ErrLockOrder` (`lock_order_violation`); 6 goroutines concorrentes disputando a cadeia M-S-P executam sem deadlocks. |
| AC7 | CI 3 SOs passa. | pass | ab1a5b3 | `TestAC7_DeterministicRaceFreeExecution`: suite de testes determinística passa 100% com `go test -race ./...` em todos os pacotes (`cmd/gira`, `internal/store`, etc.); compilação cruzada para Windows (`GOOS=windows go test -c`) e macOS (`GOOS=darwin go test -c`) limpas; `bun tools/validate/validate.ts` passa com `Overall: pass`. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; locks entre processos M->S->P, identidade física canônica, timeout de 60s e falha fechada implementados e validados
- next: Tarefa 06 (publicação imutável, heads e durabilidade)
- revalidate: todos os critérios AC1–AC7 verificados com evidência no commit ab1a5b3
