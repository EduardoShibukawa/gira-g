---
schema: task/1
tier: standard
status: todo
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
| AC1 | Invocações relativas/absolutas/subdiretório/symlink/worktree convergem ao mesmo locator quando representam o mesmo manifesto. | not run | | |
| AC2 | Manifesto com overrides fora da raiz falha com manifest_outside_git_root; falha de descoberta retorna git_root_unavailable. | not run | | |
| AC3 | Sem overrides, ausência de raiz Git não bloqueia indexação apenas por si. | not run | | |
| AC4 | Env Git adversarial não sequestra a raiz nem init. | not run | | |
| AC5 | Em todos os erros, store/head/snapshot/manifesto ficam ausentes ou byte a byte inalterados. | not run | | |
| AC6 | Testes determinísticos e CI 3 SOs passam. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de ADR 0014 e task 02 Git environment.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
