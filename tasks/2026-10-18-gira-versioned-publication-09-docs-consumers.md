---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: migrar consumidores, fixtures e documentação

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 08 CLI.

## Objective

Migrar e verificar todos os consumidores downstream e atualizar documentação/fixtures para refletir exatamente o contrato implementado.

## In scope

README, website/docs atuais (sem reescrever docs históricas versionadas), skills, bench/phases, benchtest, index-bundle, MCP, eval/scorecard, goldens, fixtures, schema docs e troubleshooting de erros.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Cada consumidor auditado recebe teste de integração ou justificativa documentada quando não aplicável.
2. AC2: bench/phases, benchtest, index-bundle e gira-arch-diff continuam funcionando.
3. AC3: Fixtures/goldens/scorecard atualizados de forma reproduzível; sem atualização silenciosa de snapshot.
4. AC4: Docs descrevem store, --output, legado, schema 2, source_identity, erros e limites de durabilidade/locks.
5. AC5: Validador e docs build passam.
6. AC6: Nenhuma doc versionada histórica é reescrita sem aprovação explícita.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Cada consumidor auditado recebe teste de integração ou justificativa documentada quando não aplicável. | pass | `cmd/gira/consumers_test.go`, `bench/layer0/layer0_test.go` | `TestConsumer_BenchPhases`, `TestConsumer_IndexBundleBootstrap`, `TestConsumer_GiraArchDiff` e `TestFromGIRASchema2` cobrem os consumidores; benchtest, mcp e eval cobertos por suítes existentes. |
| AC2 | bench/phases, benchtest, index-bundle e gira-arch-diff continuam funcionando. | pass | `cmd/gira/consumers_test.go`, `bench/` | Arm B (`index --output` + `serve mcp --index`), `bootstrap.sh`, `gira diff` e benchtest passaram com sucesso na suíte de testes (`bench` e `cmd/gira`). |
| AC3 | Fixtures/goldens/scorecard atualizados de forma reproduzível; sem atualização silenciosa de snapshot. | pass | `internal/eval/scorecard_test.go`, `bench/layer0/layer0_test.go` | `TestPublicScorecardReproducible` passou sem divergências (100% determinismo de bytes); `TestAdaptersGolden` e `TestFromGIRASchema2` preservam contratos sem modificação silenciosa de fixtures. |
| AC4 | Docs descrevem store, --output, legado, schema 2, source_identity, erros e limites de durabilidade/locks. | pass | `README.md`, `website/docs/` | `README.md`, `macro-map.md`, `index.md`, `init.md`, `serve.md`, `quickstart.md`, `intro.md` atualizados descrevendo layout D2, `--output` D1, Schema 2 com `source_identity`, política R2, taxonomia de erros, locks M->S->P e durabilidade. |
| AC5 | Validador e docs build passam. | pass | `bun tools/validate/validate.ts`, `bun run build` | Validador executou com `Overall: pass`; Docusaurus `bun run build` compilou estáticos com êxito. |
| AC6 | Nenhuma doc versionada histórica é reescrita sem aprovação explícita. | pass | `git diff --stat` | Nenhuma documentação histórica versionada (`docs/adr/0001–0013`, `docs/plans/*`, `website/versioned_docs/*`) foi alterada. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed
- next: Task 10 Integrated Acceptance
- revalidate: todos os checks AC1–AC6 passaram com evidência documentada
