---
schema: task/1
tier: standard
status: completed
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
| AC1 | Report points to exact remote SHAs and URLs and makes no unsupported local checkout/test claims. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | Remote Gira-G/G-Ingá SHAs, PR #31 HEAD/base and CI links recorded. Local branch/HEAD/worktree was not accessible and is explicitly described as unobserved. |
| AC2 | Consumer matrix includes all current fields plus proposed source_identity, the identified producer/copy/serialization/store/reader paths, current behavior and downgrade risks. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | Report maps all eight current fields, CLI/MCP/skills/eval/scorecard/slices/bench/bundle/G-Ingá and marks unenumerated dynamic consumers/global fixtures as unknown rather than absent. Global local grep remains unavailable. |
| AC3 | Current G-Ingá fallback is revalidated on current remote default branch. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | Code/tests read at G-Ingá main e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0; strict string "1.0.0" validation and integer schema 2 rejection confirmed. CI run [39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) passed test and pre-commit/validation jobs. |
| AC4 | Git for Windows archive limitation is documented with fail-closed policy and a proposed test. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | Experiment [39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230) records git archive exit 128 for a tab-containing path and tar exit 2. The report specifies an error/cleanup/fail-closed regression through callers; this task did not rerun that experiment. |
| AC5 | Report is updated and self-reviewed; no product code changed. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | Report matrix/links/statuses self-reviewed; PR #31 is documentation-only and no product source, test, fixture, spec or decision file was changed. |
| V1 | Self-review of report and task record against AC1–AC5, evidence provenance and scope boundaries. | pass | 4163a9623cad03964bd0a8f9f64d3eba90b74498 | The matrix distinguishes observed source, existing tests, actual CI evidence and unrun future SC tests; unknown consumers and local execution limits remain explicit. No SC was marked accepted. |

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
- status: completed; AC1–AC5 and V1 recorded against the report commit
- next: transition to verified with verified_at set to the evidence/report commit, then start task 01/ADR 0014 and task 02 in dependency-respecting branches
- revalidate: report is remote-only; no checkout or local command is claimed

## Remote revalidation update (2026-10-18)

The earlier progress note below the initial checkpoint described the then-current partial state and is superseded by the expanded matrix in [the readiness audit](../docs/plans/2026-10-18-gira-versioned-publication-readiness-audit.md).

- Report commit providing the evidence: `4163a9623cad03964bd0a8f9f64d3eba90b74498`.
- Gira-G `main`: `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `main`: `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`.
- PR #31 HEAD at the start of this revalidation: `7be227755fba688fac2dae602c352bab67e75408`. Its CI run [39136093491](https://github.com/EduardoShibukawa/gira-g/actions/runs/39136093491) passed all test OS jobs, validator and floor guard; docs run [39136093575](https://github.com/EduardoShibukawa/gira-g/actions/runs/39136093575) passed docs build. These runs precede the report-only matrix update commit and are not evidence for future product code.
- AC1–AC5 and V1 are now recorded as `pass` for the **read-only audit task**, with explicit limits: no local checkout/tests, no exact required-check rules because branch-protection access returned HTTP 403, and unknown dynamic/global consumers that could not be listed through the available code-search interface.
- This verification does not accept SC1–SC10 or assert schema-2 support. It records the current contract and makes the remaining implementation gates actionable.
- Task 02 must reuse the merged PR #28 changes and separately close `GIT_PREFIX`/environment coverage and Git-for-Windows partial-extraction behavior. ADR 0014 must document already approved decisions without inventing new approvals.

