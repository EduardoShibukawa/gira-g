---
schema: task/1
tier: standard
status: todo
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
| AC1 | Publicação cria versão imutável e atualiza head somente após versão completa. | not run | | |
| AC2 | Falha antes do rename deixa nenhum artefato novo visível. | not run | | |
| AC3 | Falha após rename da versão deixa head intacto e retry idempotente. | not run | | |
| AC4 | Republicação do mesmo version_id com bytes diferentes falha nondeterministic_build sem sobrescrever. | not run | | |
| AC5 | Gate rejeitado não altera store, head nem snapshot. | not run | | |
| AC6 | Unix sync failure retorna erro e visible_unconfirmed; Windows emite aviso explícito sem alegar durabilidade não garantida. | not run | | |
| AC7 | Snapshots destacados são independentes do store; arquivo corrompido existente é preservado salvo --recover-corrupt. | not run | | |
| AC8 | CI 3 SOs e testes de falha passam. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de tasks 04 keys e 05 locks.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
