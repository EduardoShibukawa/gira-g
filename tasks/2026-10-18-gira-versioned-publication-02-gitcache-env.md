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
- dependencies: Auditoria remota atualizada na PR #31; a PR #28 já está incorporada em main (6c66d938). Trabalho restrito a lacunas remanescentes; não duplicar seus fixes.

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

1. Revalidar o estado remoto e trabalhar em branch separado, sem reset/clean nem escrita em `main`.
2. Reutilizar a PR #28: manter `ls-tree -z`, cópia de diagnostics, child reaping e o isolamento atual de probes; não reconstruir as mesmas correções.
3. Estender o saneamento de probes para cobrir `GIT_PREFIX` e variáveis de descoberta/replace relevantes; adicionar testes adversariais com valores reais e inexistentes.
4. Tornar as leituras da cache determinísticas: ignorar `refs/replace` em comandos por SHA e filtrar as variáveis Git que redirecionam a leitura. Deixar clone/fetch com o ambiente do usuário intacto.
5. Reforçar `ExportTree`: extrair para staging isolado, esperar/recolher os subprocessos em todos os caminhos de erro e não deixar um destino parcial visível como resultado bem-sucedido. Preservar um destino preexistente ou falhar antes de tocá-lo.
6. Adicionar regressões para `core.quotePath`, `refs/replace`, variáveis de ambiente e erro de archive depois de emitir bytes; uma prova específica Windows deve demonstrar que um caminho não materializável falha fechado.
7. Executar os testes focados e deixar CI executar suite/validator nos três SOs. Registrar somente resultados realmente observados com commit/run.
8. Revisar os diffs e fechar ACs; não alegar que `ExportTree` é exercitado por produção se não houver chamador no repositório.

## Checkpoint

- as of: 2026-10-19
- status: active; escopo técnico começou em branch isolado
- branch: `work/2026-10-19-gira-versioned-publication-02-gitcache-env`
- base: Gira-G `main` `6c66d938148de56d460389f183bf9900f482f50f`
- evidence: PR #28 merged; remote audit PR #31 has static consumer matrix, while external/dynamic consumers and local checkout remain unobservable
- finding: `ExportTree` has no production callsite in the Go source files inspected so far; method is in `internal/gitcache`. Keep this distinction explicit and retain direct regression coverage for the method itself.
- next: implement focused gap fixes and validate them with remote CI
- revalidate: AC1–AC6 remain not run until exact-head CI completes
