---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: resolver estrito e política de compatibilidade

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 06 publish e task 04 keys.

## Objective

Implementar resolução fail-closed para versões publicadas e leitura explícita de snapshots legados.

## In scope

Schema allowlist; schema 2 estrito; legado string 1.0.0 apenas via --index explícito com aviso; nunca legado implícito com gira.yaml; erros separados missing/corrupt/unsupported/tool mismatch/ambiguous; impedir downgrade.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Schema 2 válido é aceito e source_identity preservado.
2. AC2: Schema desconhecido, tipo incorreto e campos obrigatórios ausentes retornam index_schema_unsupported; corrupção retorna index_corrupt.
3. AC3: Legado v1 só via --index explícito, read-only e com aviso.
4. AC4: Com gira.yaml, legado nunca é carregado implicitamente.
5. AC5: Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch; descoberta ambígua retorna index_ambiguous.
6. AC6: Binário/fluxo antigo não pode republicar schema 2 como schema 1 sem erro e sem alterar bytes.
7. AC7: Erros de load não são colapsados para index_missing em MCP/CLI.
8. AC8: CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Schema 2 válido é aceito e source_identity preservado. | not run | | |
| AC2 | Schema desconhecido, tipo incorreto e campos obrigatórios ausentes retornam index_schema_unsupported; corrupção retorna index_corrupt. | not run | | |
| AC3 | Legado v1 só via --index explícito, read-only e com aviso. | not run | | |
| AC4 | Com gira.yaml, legado nunca é carregado implicitamente. | not run | | |
| AC5 | Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch; descoberta ambígua retorna index_ambiguous. | not run | | |
| AC6 | Binário/fluxo antigo não pode republicar schema 2 como schema 1 sem erro e sem alterar bytes. | not run | | |
| AC7 | Erros de load não são colapsados para index_missing em MCP/CLI. | not run | | |
| AC8 | CI 3 SOs passa. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de task 06 publish e task 04 keys.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
