---
schema: task/1
tier: standard
status: todo
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
| AC1 | Dois processos concorrentes se excluem nos três SOs. | not run | | |
| AC2 | Caminhos relativos, absolutos e symlink do mesmo manifesto compartilham lock M. | not run | | |
| AC3 | Morte de processo libera lock e retry funciona. | not run | | |
| AC4 | Timeout retorna index_locked. | not run | | |
| AC5 | Falha injetada de lock retorna lock_unavailable sem artefatos gravados. | not run | | |
| AC6 | Ordem M→S→P é verificada sem deadlock. | not run | | |
| AC7 | CI 3 SOs passa. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de ADR 0014; pode executar em paralelo com locator/keys desde que não altere a ordem global.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
