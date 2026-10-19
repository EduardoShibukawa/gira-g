---
schema: task/1
tier: standard
status: active
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

1. Revalidar que esta branch parte de Gira-G `main` `6c66d938148de56d460389f183bf9900f482f50f`; PR #28 já foi incorporada, então preservar os fixes de `ls-tree -z`, subprocesso `tar`, e ordem de publicação do manifesto.
2. Manter clone/fetch com configuração/credenciais do usuário, removendo apenas variáveis que redirecionam o repositório; leituras de objetos fixados usam ambiente isolado e desativam refs/replace.
3. Adicionar regressões para `GIT_PREFIX`, `GIT_CEILING_DIRECTORIES`, todos `GIT_CONFIG_*`, `core.quotePath`, `refs/replace`, preservação de `url.insteadOf` em clone/fetch e extração parcial no Windows com path de tab criado por plumbing Git.
4. Executar testes focados e CI nos três SOs; registrar SHA, URLs de execução e resultados, mantendo todos os checks `not run` até haver execução no head relevante.
5. Revisar o diff integrado por AC1–AC6. Não alterar schema/contrato do MacroMap, store versionado, schema 2, source_identity, locks ou semântica de publicação.

## Checkpoint

- as of: 2026-10-19
- status: active; implementação delimitada iniciada nesta branch
- next: executar CI e revisar os resultados por sistema operacional; corrigir qualquer falha antes de considerar ACs aprovados
- branch: `work/2026-10-19-gitcache-env-fail-closed`, base `6c66d938148de56d460389f183bf9900f482f50f`
- revalidate: nenhum teste local foi executado; as verificações ficam `not run` até CI do head integrado
