---
schema: task/1
tier: standard
status: completed
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
| AC1 | Nomes não ASCII/tab/newline são parseados corretamente onde materializáveis. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | CI [39258639643](https://github.com/EduardoShibukawa/gira-g/actions/runs/39258639643) passou nos 3 SOs. `TestCacheManager_ListTreePreservesUnusualPathnames` passou em Ubuntu/macOS; Windows pula nomes de controle que não materializa, mas lê o nome não ASCII via `ListTree/ReadFile` e a exportação só publica árvore completa ou retorna erro claro. `TestCacheManager_ExportTreeFailsClosedOnUnmaterializableWindowsPath` passou no Windows sem deixar target/staging. |
| AC2 | `core.quotePath` e `refs/replace` não mudam bytes/entradas de leitura do commit fixado. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | `TestCacheManager_ReadsPinnedObjectsIndependentOfReplaceRefsAndGlobalConfig` passou em Ubuntu/macOS/Windows: compara `ListTree` com quotePath true/false, instala replace ref e verifica leitura do blob do SHA fixado, isolando config global. |
| AC3 | Variáveis Git de redirecionamento, prefixo e config não sequestram probes/leitura local. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | `TestGitProbesIgnoreRepositoryRedirectEnvironment` passou nos 3 SOs; `TestCacheManager_ReadsPinnedObjectsIndependentOfReplaceRefsAndGlobalConfig` também passou sob variáveis adversariais, incluindo `GIT_PREFIX`, `GIT_CONFIG_*` e pathspec/attribute env. |
| AC4 | Clone/fetch continuam respeitando configuração de usuário necessária. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | `TestCacheManager_PreservesUserURLRewriteForCloneAndFetch` passou em Ubuntu/macOS/Windows, comprovando a distinção entre ambiente de leitura estrito e ambiente de rede que preserva `url.insteadOf`. |
| AC5 | No Windows, caminhos não materializáveis falham fechado; extração parcial não é publicada. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | `TestCacheManager_ExportTreeFailsClosedOnUnmaterializableWindowsPath` passou no Windows. `ExportTree` agora valida os membros do archive e sua materialização no staging antes de promover o diretório; Ubuntu/macOS passaram `TestCacheManager_ExportTreeDoesNotPublishPartialArchive`, que faz o tar materializar conteúdo e terminar com erro e confere limpeza/target ausente. |
| AC6 | Testes e CI nos três SOs passam. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | [CI run 39258639643](https://github.com/EduardoShibukawa/gira-g/actions/runs/39258639643): `test (ubuntu-latest)`, `test (macos-latest)`, `test (windows-latest)`, validator, floor guard e gate-lint concluídos em sucesso. [Docs run 39258639578](https://github.com/EduardoShibukawa/gira-g/actions/runs/39258639578): docs build em sucesso; deploy `skipped` em PR. |
| V1 | Auto-revisão do diff integrado contra AC1–AC6 e limites de escopo. | pass | 4162d47d1b207a0a40b77b9602805f3c9366b857 | Revisado o diff de `cmd/gira`, `internal/gitcache`, indexer, testes e task. A PR reutiliza #28 e não implementa schema 2, `source_identity`, store/head versionado, resolver, locks ou durabilidade da publicação. Nenhum teste local foi alegado. |


## Plan

1. Revalidar que esta branch parte de Gira-G `main` `6c66d938148de56d460389f183bf9900f482f50f`; PR #28 já foi incorporada, então preservar os fixes de `ls-tree -z`, subprocesso `tar`, e ordem de publicação do manifesto.
2. Manter clone/fetch com configuração/credenciais do usuário, removendo apenas variáveis que redirecionam o repositório; leituras de objetos fixados usam ambiente isolado e desativam refs/replace.
3. Adicionar regressões para `GIT_PREFIX`, `GIT_CEILING_DIRECTORIES`, todos `GIT_CONFIG_*`, `core.quotePath`, `refs/replace`, preservação de `url.insteadOf` em clone/fetch e extração parcial no Windows com path de tab criado por plumbing Git.
4. Executar testes focados e CI nos três SOs; registrar SHA, URLs de execução e resultados, mantendo todos os checks `not run` até haver execução no head relevante.
5. Revisar o diff integrado por AC1–AC6. Não alterar schema/contrato do MacroMap, store versionado, schema 2, source_identity, locks ou semântica de publicação.

## Checkpoint

- as of: 2026-10-19
- status: completed; AC1–AC6 and V1 passed on `4162d47d1b207a0a40b77b9602805f3c9366b857`
- branch/PR: `work/2026-10-19-gitcache-env-fail-closed`, [PR #32](https://github.com/EduardoShibukawa/gira-g/pull/32), draft/open/no merge
- evidence: [CI 39258639643](https://github.com/EduardoShibukawa/gira-g/actions/runs/39258639643), [docs 39258639578](https://github.com/EduardoShibukawa/gira-g/actions/runs/39258639578)
- next: complete formal task-state verification; keep PR draft until review/merge authorization. Task 03 remains blocked on ADR 0014 and this task reaching `verified`.
- revalidate: no local checkout or local tests; product implementation beyond task 02 remains out of scope
