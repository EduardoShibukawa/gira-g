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
| AC1 | Report points to exact remote SHAs and URLs and makes no unsupported local checkout/test claims. | pass | 30819b42a01cb331f1ec501bfb65921e3c87f4d3 | Gira-G `main` ref `30819b4...` and G-Ingá `main` ref `e13d49d...` rechecked via Git refs. PR #32 is merged; Gira-G CI [39261086261](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086261) and docs [39261086265](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086265) succeeded. PR #31 state/head and CI were checked. No checkout/local tests claimed. |
| AC2 | Consumer matrix includes all current fields plus proposed source_identity, current behavior and downgrade risks for producer/copy/serializer/store/readers, with unknown paths explicit. | blocked | 30819b42a01cb331f1ec501bfb65921e3c87f4d3 | Current tree is 534 blobs (658 tree entries, untruncated), with 186 Go files including 94 tests and 495 text-like candidates. The matrix covers all eight current MacroMap fields, proposed `source_identity`, clone, canonical serializer, store, CLI direct readers, MCP/skills, Layer 0 adapter, bundle, fixtures/goldens/scorecard/docs and G-Ingá. `source_identity` and strict schema-2 reader remain absent. GitHub code-search returned zero results for central symbols, so it is not valid evidence of a complete global search; dynamic/reflection/host-external consumers remain unknown. |
| AC3 | Current G-Ingá fallback is revalidated on current remote default branch. | pass | e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0 | The fallback is typed for schema string `"1.0.0"`; test fixtures reject integer schema `2`, unknown versions, wrong representation, malformed JSON and invalid structures. [CI 39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) passed `test` and `pre-commit`; no open G-Ingá PRs were returned by the current query. |
| AC4 | Git for Windows archive limitation is documented with fail-closed policy and a proposed test. | pass | 30819b42a01cb331f1ec501bfb65921e3c87f4d3 | Historical experiment [39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230) recorded the tab path failure. Current main CI [39261086261](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086261): Windows `TestCacheManager_ExportTreeFailsClosedOnUnmaterializableWindowsPath` passed; partial archive test passed on Ubuntu/macOS and is skipped on Windows by design. Windows skips the general unusual-path test for unmaterializable control characters, which remains a platform limitation rather than a passing materialization claim. |
| AC5 | Report is updated and self-reviewed; no product code changed. | pass (documentary) | f6887252438d9343ecd920da78cf5949bf4a8b1e | New addendum records current remote SHAs, CI, PR states, source-level findings and uncertainty. This audit revision changes only documentation/task record, not product source, tests, fixtures, spec or owner decisions. CI for the new PR #31 head after this edit still needs to be read back. |
| V1 | Self-review of report and task record against AC1–AC5, evidence provenance and scope boundaries. | blocked | 39c92e3d347e7fd383edca56b72495063c909d77 | Self-review keeps AC2 blocked. PR #31 CI [39259495126](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495126) failed validator finding: commit `4f1c7b7` recorded `active → todo`, unreachable in the §3.3 state machine; docs run [39259495034](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495034) succeeded and the three OS jobs passed. That run preceded PR #32's merge and is not current-base CI. No SC1–SC10 acceptance claimed. |

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
- status: todo; task remains open because AC2 and V1 are blocked
- next: obtain a checkout or complete repository-wide symbol search to close AC2; do not start task 01/ADR 0014 until task 00 is verified. Task 02 may proceed only within its independent, explicitly stated dependency on remote audit revalidation and merged PR #28.
- revalidate: report is remote-only; no checkout or local command is claimed

## Remote revalidation update (2026-10-19)

- State correction: AC2 and V1 are explicitly `blocked`, and the task frontmatter is `todo`; earlier table wording that marked AC2/V1 `pass` contradicted the checkpoint and is superseded by this record.

- Current audit report commit: `44a7bb3f3f16db2cb5459f852ddaed1000e18fbb`; it inventories the full Git tree and adds classifications for the Layer 0 adapter, scorecard artefact, skills and versioned docs.
- Gira-G `main`: `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `main`: `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`.
- CI of the previous audit HEAD `94dbe2920b9bb5efe457813064d81bbf95f6209c` completed successfully in runs [39146546866](https://github.com/EduardoShibukawa/gira-g/actions/runs/39146546866) and [39146546868](https://github.com/EduardoShibukawa/gira-g/actions/runs/39146546868). CI for the newer report commit must be rechecked before its documentation changes are treated as validated.
- AC2 remains `blocked`: the static matrix is substantially expanded, but GitHub code-search was incomplete and a local checkout is not available to run `git grep`. No absence claim is made for possible dynamic/reflection consumers.
- Consequently task 00 remains `active`, not `verified`; task 01/ADR 0014 remains gated on task 00. No SC1–SC10 acceptance is claimed.
- Task 02 can continue only in its own branch, reusing merged PR #28 and limiting scope to deterministic Git reads, environment isolation and fail-closed archive handling. No schema 2 or versioned store implementation is in scope.



## Revalidação remota (2026-10-19, após merge da PR #32)

- **Task state:** frontmatter permanece `status: todo`. AC2/V1 estão bloqueados, portanto não converter para `verified`; o estado `heavy/active` da tarefa-mãe e a autorização existente não foram alterados.
- **Gira-G:** main `30819b42a01cb331f1ec501bfb65921e3c87f4d3`; [CI 39261086261](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086261) passou nos jobs Ubuntu/macOS/Windows, validator, floor guard e gate-lint; [docs 39261086265](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086265) passou.
- **G-Ingá:** main `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0); [run 39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) passou `test` e `pre-commit`; consulta atual não encontrou PR aberta.
- **PR #32:** integrada; merge SHA `30819b4...`. O conteúdo da task 02 em main contém um checkpoint desatualizado que ainda diz `draft/open/no merge`; isso deve ser reconciliado documentalmente após a verificação da task 00, sem reabrir a PR nem repetir seus fixes.
- **PR #31:** continua aberta/em draft, HEAD observado antes desta atualização `39c92e3d347e7fd383edca56b72495063c909d77`. O run CI [39259495126](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495126) falhou por transição histórica inválida `active → todo` no commit `4f1c7b7`; docs [39259495034](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495034) passou. Não foi feito force-push/rewrite do histórico. A integração da #32 mudou main depois desses runs, portanto o próximo resultado de CI da #31 precisa ser reconsultado contra a base atual.
- **Read-only / limites:** checkout local não disponível; nenhum teste local executado. Nenhum código de produto, teste, fixture, spec, decisão aprovada, main, merge ou branch histórica foi alterado por esta auditoria.
- **Next:** fechar a busca global de consumidores dinâmicos/reflexivos com evidência reproduzível; só então verificar task 00. ADR 0014 continua dependente da task 00 e permanece `todo`. Manter SC1–SC10 sem aceite.
