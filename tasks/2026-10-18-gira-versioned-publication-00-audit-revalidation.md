---
schema: task/1
tier: standard
status: active
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
| AC1 | Report points to exact remote SHAs and URLs and makes no unsupported local checkout/test claims. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | Remote Gira-G/G-Ingá SHAs and PR #31 refs recorded; local branch/HEAD/worktree not available and not claimed. |
| AC2 | Consumer matrix includes all current fields plus proposed source_identity, current behavior and downgrade risks for producer/copy/serializer/store/readers, with unknown paths explicit. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | At Gira-G main tree 6c66d938148de56d460389f183bf9900f482f50f, all 96 first-party source/script files and 94 test files were fetched remotely and searched for MacroMap/schema/store/consumer signatures. Matrix covers identified CLI/MCP/skills/eval/scorecard/slices/bench/Layer0/bundle/fixtures/docs and all eight current fields plus future source_identity. Reflection/dynamic references remain explicitly unknown because code-search is incomplete; the criterion asks to record such unknowns, not claim there are none. |
| AC3 | Current G-Ingá fallback is revalidated on current remote default branch. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | Source/tests at G-Ingá main e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0 confirm strict string "1.0.0", integer schema 2 rejection and malformed/invalid structure cases. CI run [39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) passed test and pre-commit/validation. |
| AC4 | Git for Windows archive limitation is documented with fail-closed policy and a proposed test. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | Experiment [39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230) recorded archive exit 128 for a tab-containing path and tar exit 2. Regression proposal requires diagnostic failure and no published/partial target; this task did not rerun that experiment. |
| AC5 | Report is updated and self-reviewed; no product code changed. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | PR #31 changes only the report and task record. No product source/test/fixture/spec/decision file changed. The docs build run [39255348675](https://github.com/EduardoShibukawa/gira-g/actions/runs/39255348675) passed. |
| V1 | Self-review of report and task record against AC1–AC5, evidence provenance and scope boundaries. | pass | 53f7e22b6ecd4b0787ee167c69b5189cad1ce3e9 | Reviewed the full consumer crosswalk, source/test scan coverage, report links/SHAs, CI evidence, distinction between observed code/tests and unrun future SCs, and explicit dynamic-consumer uncertainty. No SC1–SC10 acceptance claimed. |

## Plan

1. Revalidar os SHAs remotos de `main`, os workflows recentes, as PRs relevantes e, se houver um checkout local efetivamente acessível, seu branch/HEAD/worktree. Se o checkout local não estiver disponível, registrar essa limitação sem inferir seu estado.
2. Ler a spec aprovada, os registros de decisão, a tarefa-mãe, as tarefas downstream e a auditoria de consumidores anterior.
3. Rastrear em modo read-only o contrato MacroMap do produtor até cópia, serialização, store, leitores CLI/MCP/skills, bench/eval/bundle/fixtures/goldens/docs e o fallback do G-Ingá. Para cada caminho, registrar símbolo, arquivo, SHA/URL, campos preservados/perdidos e comportamento de erro observado.
4. Comparar o diff final da PR #28 com o HEAD atual; identificar correções já integradas e riscos ainda abertos para não duplicar fixes.
5. Montar a matriz de compatibilidade de schema e consumidores, incluindo `source_identity`, schema `2` inteiro, legado `"1.0.0"`, prevenção de downgrade e o limite de Git for Windows para nomes não materializáveis. Separar teste existente, teste executado com evidência, teste planejado e teste ausente.
6. Registrar a matriz de plataformas e CI com URLs, commits e resultado de cada job; não afirmar execução local ou aceitação de SC sem evidência do ambiente e do commit correspondente.
7. Auto-revisar AC1–AC5, classificar gaps e propor testes quando faltarem. Esta tarefa é read-only: não implementar comportamento, não adicionar/alterar testes de produto, não alterar código de produto, não mudar a spec/decisões nem marcar a tarefa como concluída enquanto qualquer critério permanecer sem evidência.
## Checkpoint

- as of: 2026-10-19
- status: active; task remains open because AC2 and V1 are blocked
- next: obtain a checkout or complete repository-wide symbol search to close AC2; do not start task 01/ADR 0014 until task 00 is verified. Task 02 may proceed only within its independent, explicitly stated dependency on remote audit revalidation and merged PR #28.
- revalidate: report is remote-only; no checkout or local command is claimed

## Remote revalidation update (2026-10-19)

- Current audit report commit: `44a7bb3f3f16db2cb5459f852ddaed1000e18fbb`; it inventories the full Git tree and adds classifications for the Layer 0 adapter, scorecard artefact, skills and versioned docs.
- Gira-G `main`: `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `main`: `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`.
- CI of the previous audit HEAD `94dbe2920b9bb5efe457813064d81bbf95f6209c` completed successfully in runs [39146546866](https://github.com/EduardoShibukawa/gira-g/actions/runs/39146546866) and [39146546868](https://github.com/EduardoShibukawa/gira-g/actions/runs/39146546868). CI for the newer report commit must be rechecked before its documentation changes are treated as validated.
- AC2 remains `blocked`: the static matrix is substantially expanded, but GitHub code-search was incomplete and a local checkout is not available to run `git grep`. No absence claim is made for possible dynamic/reflection consumers.
- Consequently task 00 remains `active`, not `verified`; task 01/ADR 0014 remains gated on task 00. No SC1–SC10 acceptance is claimed.
- Task 02 can continue only in its own branch, reusing merged PR #28 and limiting scope to deterministic Git reads, environment isolation and fail-closed archive handling. No schema 2 or versioned store implementation is in scope.

