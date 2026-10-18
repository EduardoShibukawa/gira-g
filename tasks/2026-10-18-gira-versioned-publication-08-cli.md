---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: integrar index/init e fluxos CLI

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 05 locks e task 07 resolver.

## Objective

Integrar o protocolo versionado aos comandos index/init e consumidores da CLI, preservando a ordem de publicação e semântica dos snapshots destacados.

## In scope

index sem output publica canônico; index --output só snapshot; init publica canônico e cópia byte a byte; manifesto por último; conflito se manifesto muda durante init; remover fallback eval para gira.yaml; skills --index legacy warning e path absoluto; erros estáveis.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: index --output P não altera store/head/manifesto.
2. AC2: init --output P publica canônico e P com bytes idênticos; manifesto é gravado por último.
3. AC3: Falha de build/gate/sync não cria manifesto novo nem altera o existente.
4. AC4: Alteração concorrente do manifesto durante init gera conflito sem publicar estado misto.
5. AC5: eval sem manifesto real retorna manifest_path_required quando overrides exigem locator.
6. AC6: skills install/check --index avisa para legado e normaliza caminho relativo para absoluto.
7. AC7: Mensagens refletem publicação parcial/visible_unconfirmed corretamente.
8. AC8: Comandos existentes continuam funcionando e CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | index --output P não altera store/head/manifesto. | not run | | |
| AC2 | init --output P publica canônico e P com bytes idênticos; manifesto é gravado por último. | not run | | |
| AC3 | Falha de build/gate/sync não cria manifesto novo nem altera o existente. | not run | | |
| AC4 | Alteração concorrente do manifesto durante init gera conflito sem publicar estado misto. | not run | | |
| AC5 | eval sem manifesto real retorna manifest_path_required quando overrides exigem locator. | not run | | |
| AC6 | skills install/check --index avisa para legado e normaliza caminho relativo para absoluto. | not run | | |
| AC7 | Mensagens refletem publicação parcial/visible_unconfirmed corretamente. | not run | | |
| AC8 | Comandos existentes continuam funcionando e CI 3 SOs passa. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de task 05 locks e task 07 resolver.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
