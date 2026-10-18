---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: leituras Git determinísticas e isolamento de ambiente

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Pode começar após revalidar a auditoria; incorporar/aguardar merge da PR #28 antes de duplicar seus fixes.

## Objective

Garantir que probes locais e leitura de árvores Git não dependam de variáveis/configuração que redirecionem o repositório ou alterem a representação de caminhos, sem remover a configuração necessária para clone/fetch.

## In scope

NUL-delimited ls-tree; sanitização de variáveis GIT_* para probes locais; preservar env de clone/fetch; tratar caminhos que Git for Windows não pode materializar com erro claro e sem considerar extração parcial sucesso.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Nomes não ASCII/tab/newline são parseados corretamente onde materializáveis.
2. AC2: core.quotePath e refs/replace não alteram bytes do índice.
3. AC3: GIT_DIR, GIT_WORK_TREE, GIT_COMMON_DIR, GIT_INDEX_FILE, GIT_OBJECT_DIRECTORY, GIT_ALTERNATE_OBJECT_DIRECTORIES, GIT_NAMESPACE, GIT_PREFIX e GIT_CONFIG_* não sequestram probes locais.
4. AC4: Clone/fetch continuam respeitando configuração/credenciais necessárias.
5. AC5: Windows falha fechado e com diagnóstico claro quando o caminho não pode ser materializado; não aceita extração parcial.
6. AC6: Testes e CI dos três SOs passam.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Nomes não ASCII/tab/newline são parseados corretamente onde materializáveis. | not run | | |
| AC2 | core.quotePath e refs/replace não alteram bytes do índice. | not run | | |
| AC3 | GIT_DIR, GIT_WORK_TREE, GIT_COMMON_DIR, GIT_INDEX_FILE, GIT_OBJECT_DIRECTORY, GIT_ALTERNATE_OBJECT_DIRECTORIES, GIT_NAMESPACE, GIT_PREFIX e GIT_CONFIG_* não sequestram probes locais. | not run | | |
| AC4 | Clone/fetch continuam respeitando configuração/credenciais necessárias. | not run | | |
| AC5 | Windows falha fechado e com diagnóstico claro quando o caminho não pode ser materializado; não aceita extração parcial. | not run | | |
| AC6 | Testes e CI dos três SOs passam. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Pode começar após revalidar a auditoria; incorporar/aguardar merge da PR #28 antes de duplicar seus fixes.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
