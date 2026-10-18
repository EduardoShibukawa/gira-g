---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: revalidar auditoria de consumidores e limites de plataforma

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Nenhuma; executar antes de ADR 0014 e de decisões que dependam de detalhes do consumidor.

## Objective

Atualizar a evidência remota de compatibilidade após a aprovação da spec, comparando a auditoria existente com o HEAD atual de Gira-G, o estado final da PR #28 e o fallback atual de G-Ingá. Documentar limites reais do Git for Windows para caminhos não materializáveis.

## In scope

Revalidar produtores, cópia/serialização, store, CLI, MCP, skills, bench, eval/scorecard, index-bundle, fixtures, goldens e docs; registrar SHAs atuais e CI; cobrir tab/newline/não ASCII, ambiente Git e comportamento de schema; distinguir testes definidos, executados e não executados.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Report aponta SHAs exatos e URLs das fontes atuais; não afirma checkout local nem testes locais que não ocorreram.
2. AC2: Matriz de consumidores inclui todos os campos atuais e source_identity, com comportamento atual e risco de downgrade.
3. AC3: Comportamento atual do fallback G-Ingá é revalidado no HEAD atual.
4. AC4: Limite do git archive no Windows é documentado com política fail-closed e teste proposto.
5. AC5: Relatório atualizado e auto-revisado; nenhum código de produto modificado.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Report aponta SHAs exatos e URLs das fontes atuais; não afirma checkout local nem testes locais que não ocorreram. | not run | | |
| AC2 | Matriz de consumidores inclui todos os campos atuais e source_identity, com comportamento atual e risco de downgrade. | not run | | |
| AC3 | Comportamento atual do fallback G-Ingá é revalidado no HEAD atual. | not run | | |
| AC4 | Limite do git archive no Windows é documentado com política fail-closed e teste proposto. | not run | | |
| AC5 | Relatório atualizado e auto-revisado; nenhum código de produto modificado. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Nenhuma; executar antes de ADR 0014 e de decisões que dependam de detalhes do consumidor.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
