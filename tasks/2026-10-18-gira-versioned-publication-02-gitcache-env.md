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
- dependencies: Revalidação remota da auditoria 00; PR #28 já está integrada em `main` (`6c66d938148de56d460389f183bf9900f482f50f`). Reutilizar os fixes existentes, sem duplicá-los.

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
2. Reler o Gate 2 da auditoria técnica, ADR 0009 e testes/changes da PR #28 antes de editar.
3. Manter probes locais separados de clone/fetch: filtrar `GIT_PREFIX` e `GIT_CEILING_DIRECTORIES` nos probes locais, mas deixar clone/fetch herdarem a configuração/credenciais do usuário.
4. Extrair `git archive` para diretório irmão temporário; publicar/substituir o destino só quando archive e tar terminarem com sucesso, restaurando o destino anterior se a promoção falhar.
5. Adicionar regressões de variáveis Git/config, URL rewrite de usuário para clone/fetch, falha após emissão parcial do tar e um tree Git sintético com pathname contendo tab, sem criar esse nome no filesystem Windows.
6. Rodar CI no SHA exato desta branch (Linux/macOS/Windows, validator e floor guard); nenhum critério deve ser marcado `pass` a partir apenas de existência do teste ou leitura de código.
7. Revisar o diff integrado contra AC1–AC6. Não implementar schema 2, `source_identity`, store/head versionado, resolver ou locks nesta tarefa.

## Checkpoint

- as of: 2026-10-19
- status: active; branch separado da auditoria, sem alterações em `main`
- base: PR #31 audit head `85312f52cb3baa3afd8360c6c57feee85a9c30b1`; `main` base `6c66d938148de56d460389f183bf9900f482f50f`
- changes so far: filtros de probes incluem `GIT_PREFIX` e `GIT_CEILING_DIRECTORIES`; testes de env/clone-fetch/partial archive/tab pathname estão sendo adicionados; ExportTree usa staging antes de substituir o destino.
- verification: todos AC1–AC6 continuam `not run` até CI do SHA final; não há teste local porque não existe checkout local acessível nesta sessão.
- next: terminar auto-revisão dos testes e código, abrir/atualizar PR draft empilhada sobre a branch da auditoria e registrar links dos checks do head exato.
