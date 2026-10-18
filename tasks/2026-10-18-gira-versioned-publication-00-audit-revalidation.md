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

1. Revalidar os SHAs remotos de `main`, os workflows recentes, as PRs relevantes e, se houver um checkout local efetivamente acessível, seu branch/HEAD/worktree. Se o checkout local não estiver disponível, registrar essa limitação sem inferir seu estado.
2. Ler a spec aprovada, os registros de decisão, a tarefa-mãe, as tarefas downstream e a auditoria de consumidores anterior.
3. Rastrear em modo read-only o contrato MacroMap do produtor até cópia, serialização, store, leitores CLI/MCP/skills, bench/eval/bundle/fixtures/goldens/docs e o fallback do G-Ingá. Para cada caminho, registrar símbolo, arquivo, SHA/URL, campos preservados/perdidos e comportamento de erro observado.
4. Comparar o diff final da PR #28 com o HEAD atual; identificar correções já integradas e riscos ainda abertos para não duplicar fixes.
5. Montar a matriz de compatibilidade de schema e consumidores, incluindo `source_identity`, schema `2` inteiro, legado `"1.0.0"`, prevenção de downgrade e o limite de Git for Windows para nomes não materializáveis. Separar teste existente, teste executado com evidência, teste planejado e teste ausente.
6. Registrar a matriz de plataformas e CI com URLs, commits e resultado de cada job; não afirmar execução local ou aceitação de SC sem evidência do ambiente e do commit correspondente.
7. Auto-revisar AC1–AC5, classificar gaps e propor testes quando faltarem. Esta tarefa é read-only: não implementar comportamento, não adicionar/alterar testes de produto, não alterar código de produto, não mudar a spec/decisões nem marcar a tarefa como concluída enquanto qualquer critério permanecer sem evidência.
## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Nenhuma; executar antes de ADR 0014 e de decisões que dependam de detalhes do consumidor.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
