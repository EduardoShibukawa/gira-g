---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: migrar consumidores, fixtures e documentação

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 08 CLI.

## Objective

Migrar e verificar todos os consumidores downstream e atualizar documentação/fixtures para refletir exatamente o contrato implementado.

## In scope

README, website/docs atuais (sem reescrever docs históricas versionadas), skills, bench/phases, benchtest, index-bundle, MCP, eval/scorecard, goldens, fixtures, schema docs e troubleshooting de erros.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Cada consumidor auditado recebe teste de integração ou justificativa documentada quando não aplicável.
2. AC2: bench/phases, benchtest, index-bundle e gira-arch-diff continuam funcionando.
3. AC3: Fixtures/goldens/scorecard atualizados de forma reproduzível; sem atualização silenciosa de snapshot.
4. AC4: Docs descrevem store, --output, legado, schema 2, source_identity, erros e limites de durabilidade/locks.
5. AC5: Validador e docs build passam.
6. AC6: Nenhuma doc versionada histórica é reescrita sem aprovação explícita.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Cada consumidor auditado recebe teste de integração ou justificativa documentada quando não aplicável. | not run | | |
| AC2 | bench/phases, benchtest, index-bundle e gira-arch-diff continuam funcionando. | not run | | |
| AC3 | Fixtures/goldens/scorecard atualizados de forma reproduzível; sem atualização silenciosa de snapshot. | not run | | |
| AC4 | Docs descrevem store, --output, legado, schema 2, source_identity, erros e limites de durabilidade/locks. | not run | | |
| AC5 | Validador e docs build passam. | not run | | |
| AC6 | Nenhuma doc versionada histórica é reescrita sem aprovação explícita. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de task 08 CLI.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
