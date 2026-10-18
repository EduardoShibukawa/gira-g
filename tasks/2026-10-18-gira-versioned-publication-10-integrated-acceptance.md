---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: aceitação integrada SC1–SC10 e revisão final

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de tasks 00–09; task 00 deve ser reconciliada antes da aceitação final.

## Objective

Executar a matriz de aceitação completa no HEAD integrado, registrar evidência reproduzível e decidir se a implementação pode ser aceita.

## In scope

SC1–SC10; testes race em Ubuntu/macOS/Windows; validator, floor guard, docs build; isolamento de caches; revisão do diff integrado e das falhas de publicação, schema, locks e plataforma.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Cada SC1–SC10 tem resultado explícito, SHA, comando, SO, exit code e artefato/log quando aplicável.
2. AC2: go test -race ./... passa em Ubuntu, macOS e Windows.
3. AC3: Validator, floor guard e docs build passam no SHA integrado.
4. AC4: Testes não escrevem no cache real do usuário; evidência de isolamento registrada.
5. AC5: Gate Windows para tab em archive/tar e symlink não elevado tem resultado ou skip justificado.
6. AC6: Revisão do diff não encontra downgrade silencioso, perda de provenance, aceitação de mapa vazio ou estado parcial tratado como sucesso.
7. AC7: Nenhum critério é marcado verified sem evidência; qualquer falha bloqueia aceitação.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Cada SC1–SC10 tem resultado explícito, SHA, comando, SO, exit code e artefato/log quando aplicável. | not run | | |
| AC2 | go test -race ./... passa em Ubuntu, macOS e Windows. | not run | | |
| AC3 | Validator, floor guard e docs build passam no SHA integrado. | not run | | |
| AC4 | Testes não escrevem no cache real do usuário; evidência de isolamento registrada. | not run | | |
| AC5 | Gate Windows para tab em archive/tar e symlink não elevado tem resultado ou skip justificado. | not run | | |
| AC6 | Revisão do diff não encontra downgrade silencioso, perda de provenance, aceitação de mapa vazio ou estado parcial tratado como sucesso. | not run | | |
| AC7 | Nenhum critério é marcado verified sem evidência; qualquer falha bloqueia aceitação. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de tasks 00–09; task 00 deve ser reconciliada antes da aceitação final.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
