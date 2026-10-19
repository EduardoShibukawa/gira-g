---
schema: task/1
tier: standard
status: completed
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
| AC1 | ADR cita spec aprovada e decisões arquiteturais/gates existentes. | pass | d0d54ec | ADR 0014 cita expressamente specs/2026-10-17-gira-versioned-publication.md, ADR 0001, ADR 0003, ADR 0007, ADR 0009, decisões N1-N4, D1-D9c e Gates 1-5. |
| AC2 | Schema 2 inteiro e source_identity são definidos; legado v1 só em --index explícito, read-only, com aviso. | pass | d0d54ec | schema_version: 2 inteiro definido conforme Gate 1 (Opção B); source_identity estruturado com manifest_sha256 sobre bytes crus, extratores e fontes ordenadas com url_sha256; leitor R2 aceita legado apenas com flag explícito --index, em modo leitura e com aviso. |
| AC3 | Semântica index --output, init --output, store/head, erros e ordem de locks coincide com a spec. | pass | d0d54ec | index --output grava snapshot destacado sem alterar store/head; init --output grava versão canônica e cópia P com manifesto por último; ordem M -> S -> P e erros index_missing, index_tool_mismatch, index_schema_unsupported, index_corrupt, nondeterministic_build documentados. |
| AC4 | Limites Windows/durabilidade e fail-closed são explícitos. | pass | d0d54ec | Limitações do Git for Windows documentadas com fail-closed para nomes não materializáveis; semântica visible_unconfirmed, fsync Unix e limitações de diretório Windows explicitados. |
| AC5 | Validador passa e nenhum comportamento de produto é alegado como implementado. | pass | d0d54ec | Validador standalone passa com sucesso; ADR 0014 consolida contrato arquitetural e especificações sem alegar implementação de código de produto. |
| V1 | Auto-revisão contra critérios de aceitação AC1–AC5 e fronteiras de governança. | pass | d0d54ec | Auto-revisão concluiu conformidade total com a spec aprovada, decisões N1–N4, D1–D9c e Gates 1–5, sem modificação de código de produto. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; ADR 0014 criado e critérios AC1–AC5 e V1 verificados com sucesso
- branch: work/2026-10-19-task01-adr
- evidence: docs/adr/0014-versioned-index-publication-and-contract.md criado; validador standalone (bun tools/validate/validate.ts) executado com Overall: pass; nenhum código de produto alterado
- next: Tarefa 01 concluída; desbloquear tarefas dependentes conforme a ordem de dependências do projeto
- revalidate: todos os critérios de aceitação AC1–AC5 e auto-revisão V1 validados contra o artefato ADR 0014 e a governança do repositório
