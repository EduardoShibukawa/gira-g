---
schema: task/1
tier: standard
status: todo
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: consolidar contrato aprovado em ADR 0014

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende da task 00; não muda decisões de produto já registradas sem nova aprovação do dono.

## Objective

Criar ADR 0014 consolidando o contrato aprovado da publicação versionada, schema 2, source_identity, reader policy R2, locator, tool_id, locks, durabilidade e comportamento de snapshots destacados.

## In scope

Documentar contexto, decisão, alternativas rejeitadas, consequências, erros estáveis, política de legado e limitações de plataforma. Corrigir inconsistências de nomenclatura Etapa 1.2 vs 2a–2i somente nesta documentação.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: ADR cita spec aprovada e decisões arquiteturais/gates existentes.
2. AC2: Schema 2 inteiro e source_identity são definidos; legado v1 só em --index explícito, read-only, com aviso.
3. AC3: Semântica index --output, init --output, store/head, erros e ordem de locks coincide com a spec.
4. AC4: Limites Windows/durabilidade e fail-closed são explícitos.
5. AC5: Validador passa e nenhum comportamento de produto é alegado como implementado.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | ADR cita spec aprovada e decisões arquiteturais/gates existentes. | not run | | |
| AC2 | Schema 2 inteiro e source_identity são definidos; legado v1 só em --index explícito, read-only, com aviso. | not run | | |
| AC3 | Semântica index --output, init --output, store/head, erros e ordem de locks coincide com a spec. | not run | | |
| AC4 | Limites Windows/durabilidade e fail-closed são explícitos. | not run | | |
| AC5 | Validador passa e nenhum comportamento de produto é alegado como implementado. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende da task 00; não muda decisões de produto já registradas sem nova aprovação do dono.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
