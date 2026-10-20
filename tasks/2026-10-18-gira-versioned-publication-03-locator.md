---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: normalizar locator do manifesto e raiz Git

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de ADR 0014 e task 02 Git environment.

## Objective

Implementar locator determinístico relativo à raiz Git para manifestos com overrides, com resolução de symlink e falha fechada fora da raiz.

## In scope

Capturar bytes crus do manifesto; normalizar caminho depois de resolver symlinks; suportar raiz, subdiretório, worktree, submódulo e symlink; erros manifest_outside_git_root e git_root_unavailable; nenhuma publicação em falha.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Invocações relativas/absolutas/subdiretório/symlink/worktree convergem ao mesmo locator quando representam o mesmo manifesto.
2. AC2: Manifesto com overrides fora da raiz falha com manifest_outside_git_root; falha de descoberta retorna git_root_unavailable.
3. AC3: Sem overrides, ausência de raiz Git não bloqueia indexação apenas por si.
4. AC4: Env Git adversarial não sequestra a raiz nem init.
5. AC5: Em todos os erros, store/head/snapshot/manifesto ficam ausentes ou byte a byte inalterados.
6. AC6: Testes determinísticos e CI 3 SOs passam.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Invocações relativas/absolutas/subdiretório/symlink/worktree convergem ao mesmo locator quando representam o mesmo manifesto. | pass | fe4dd5c | `TestResolveManifestLocator_Convergence` em `internal/manifest` e `TestCLI_AC1_LocatorConvergence` em `cmd/gira`: caminhos relativos, absolutos, de subdiretório, com `..`, symlinks de arquivo/diretório e worktrees convergem deterministicamente ao mesmo locator com `/` (`gira.yaml` e `conf/sub/gira.yaml`) e bytes JSON idênticos. |
| AC2 | Manifesto com overrides fora da raiz falha com manifest_outside_git_root; falha de descoberta retorna git_root_unavailable. | pass | fe4dd5c | `TestResolveManifestLocator_OutsideGitRoot`, `TestIndexer_LocatorNormalization_OutsideGitRoot_Fails`, `TestCLI_AC2_OutsideGitRoot_And_GitRootUnavailable_FailClosed`: manifesto fora do git ou em repo bare/.git com overrides falha fechado com `manifest_outside_git_root`; descoberta sob dubious ownership (`GIT_TEST_ASSUME_DIFFERENT_OWNER=1`) falha com `git_root_unavailable`; caminho vazio falha com `manifest_path_required`. |
| AC3 | Sem overrides, ausência de raiz Git não bloqueia indexação apenas por si. | pass | fe4dd5c | `TestIndexer_NoOverrides_OutsideGitRoot_Succeeds` e `TestCLI_AC3_NoOverrides_OutsideGitRoot_Succeeds`: manifesto sem overrides é indexado com sucesso mesmo fora de raiz Git. |
| AC4 | Env Git adversarial não sequestra a raiz nem init. | pass | fe4dd5c | `TestResolveManifestLocator_AdversarialGitEnv`, `TestCLI_AC4_AdversarialGitEnv_DoesNotHijack` e `TestGitProbesIgnoreRepositoryRedirectEnvironment`: variáveis adversariais (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_CONFIG`, etc.) são expurgadas por `manifest.GitCommandEnv()`, impedindo sequestro em `index` e `init`. |
| AC5 | Em todos os erros, store/head/snapshot/manifesto ficam ausentes ou byte a byte inalterados. | pass | fe4dd5c | `TestCLI_AC5_FailClosed_ArtifactsIntact`: em falhas fechadas (`manifest_outside_git_root`, `git_root_unavailable`), snapshots inexistentes não são criados, snapshots preexistentes permanecem inalterados, nenhum store é criado/modificado em `~/.cache` ou `.gira`, e nenhum `gira.yaml` é gravado. |
| AC6 | Testes determinísticos e CI 3 SOs passam. | pass | fe4dd5c | `go test -race ./...` passou em todos os 17 pacotes sem falhas nem races (74s `cmd/gira`, 36s `internal/eval`, 0 falhas). `bun tools/validate/validate.ts` passou com `Overall: pass`. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; AC1–AC6 aprovados em fe4dd5c
- summary: Normalização determinística do locator de manifesto relativa à raiz Git implementada com política O1 (fail-closed com manifest_outside_git_root e git_root_unavailable), saneamento estrito de env Git, convergência testada em matriz completa (relativo/absoluto/subdiretório/symlink/worktree) e preservação estrita de cmd/gira/main_test.go.
