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
| AC1 | Report points to exact remote SHAs and URLs and makes no unsupported local checkout/test claims. | pass | c83c5891166267ec26816d75c305daf0a2dcda8e | Gira-G `main` ref `c83c589...` e G-Ingá `main` ref `e13d49d...` confirmados por Git refs e fetches remotos. Worktree isolada `../gira-g-wt-audit00-local` na branch `audit/2026-10-19-task00-local-grep`. Checkout original preservado sem alterações em `cmd/gira/main_test.go`. |
| AC2 | Consumer matrix includes all current fields plus proposed source_identity, current behavior and downgrade risks for producer/copy/serializer/store/readers, with unknown paths explicit. | pass | c83c5891166267ec26816d75c305daf0a2dcda8e | Matriz exaustiva concluída via 3 passagens de `git grep` reproduzíveis sobre a árvore rastreada integral (535 arquivos em Gira-G, 347 em G-Ingá). Cobre os 8 campos atuais, o campo futuro `source_identity` (`source_key`, `tool_id`), clone, serialização, persistência, leitor central CLI `loadIndex`, `runDiff`, MCP `diff_maps`, `registry`, `skills` (`isStale`, `nudge`, `install`), adaptadores de bench (`layer0/gira.go`, `cmd/layer1`, `phases.json`, `layer2/adapters.go`), metadata roundtrip (`tour.go`, `intra_service.go`), `eval/scorecard`, `index-bundle` e o consumidor externo G-Ingá (`evaluate_girag_slicing.ts`). Comportamentos de erro (schema int 2, string desconhecida, JSON malformado, dados estruturalmente inválidos e campo extra) verificados por probes black-box observados em diretório temporário isolado. Limite de hosts externos delimitado no relatório e ressalva explícita de que zero ocorrências de `reflect.` constituem evidência textual do padrão pesquisado, não prova absoluta de ausência de qualquer consumidor dinâmico indireto. |
| AC3 | Current G-Ingá fallback is revalidated on current remote default branch. | pass | e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0 | Fallback revalidado na árvore do commit `e13d49d...`. Execução real da suíte `bun test tools/eval/evaluate_girag_slicing.test.ts` observada com sucesso (12 testes pass, 0 fail, 395 asserts). Fallback rejeita estritamente schema int 2 (`Unsupported Macro Map schema_version: 2`) e versões desconhecidas. Sem suporte a schema 2 em G-Ingá. |
| AC4 | Git for Windows archive limitation is documented with fail-closed policy and a proposed test. | pass | 30819b42a01cb331f1ec501bfb65921e3c87f4d3 | Experimento histórico [39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230) registrou a falha de tab path. CI de main [39261086261](https://github.com/EduardoShibukawa/gira-g/actions/runs/39261086261): Windows `TestCacheManager_ExportTreeFailsClosedOnUnmaterializableWindowsPath` passou; teste de arquivo parcial passou em Ubuntu/macOS e é ignorado em Windows por design. |
| AC5 | Report is updated and self-reviewed; no product code changed. | pass | c83c5891166267ec26816d75c305daf0a2dcda8e | Critério documental: o relatório de prontidão e este registro foram atualizados com a auditoria local completa, comandos exatos, hashes e matriz AC2. Nenhum código de produto, teste de produto, fixture, spec ou decisão de dono foi alterado. |
| V1 | Self-review of report and task record against AC1–AC5, evidence provenance and scope boundaries. | pass | c83c5891166267ec26816d75c305daf0a2dcda8e | Auto-revisão contra AC1–AC5 concluída. Comandos de busca reproduzíveis registrados com hashes SHA-256 das saídas brutas. Probes black-box executados com saídas e códigos de retorno verificados. Testes de pacotes relevantes de Gira-G e G-Ingá executados com sucesso em ambiente isolado. Validador executado com Overall: pass. |

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
- **G-Ingá:** main `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`; [run 39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) passou `test` e `pre-commit`; consulta atual não encontrou PR aberta.
- **PR #32:** integrada; merge SHA `30819b4...`. O conteúdo da task 02 em main contém um checkpoint desatualizado que ainda diz `draft/open/no merge`; isso deve ser reconciliado documentalmente após a verificação da task 00, sem reabrir a PR nem repetir seus fixes.
- **PR #31:** continua aberta/em draft, HEAD observado antes desta atualização `39c92e3d347e7fd383edca56b72495063c909d77`. O run CI [39259495126](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495126) falhou por transição histórica inválida `active → todo` no commit `4f1c7b7`; docs [39259495034](https://github.com/EduardoShibukawa/gira-g/actions/runs/39259495034) passou. Não foi feito force-push/rewrite do histórico. A integração da #32 mudou main depois desses runs, portanto o próximo resultado de CI da #31 precisa ser reconsultado contra a base atual.
- **Read-only / limites:** checkout local não disponível; nenhum teste local executado. Nenhum código de produto, teste, fixture, spec, decisão aprovada, main, merge ou branch histórica foi alterado por esta auditoria.
- **Next:** fechar a busca global de consumidores dinâmicos/reflexivos com evidência reproduzível; só então verificar task 00. ADR 0014 continua dependente da task 00 e permanece `todo`. Manter SC1–SC10 sem aceite.


## Revalidação local reproduzível AC2 (2026-10-19)

- **Task state:** transição formal de `todo → active` executada na worktree isolada `../gira-g-wt-audit00-local` (branch `audit/2026-10-19-task00-local-grep`, base `c83c5891166267ec26816d75c305daf0a2dcda8e`).
- **Gira-G:** `main` remoto em `c83c5891166267ec26816d75c305daf0a2dcda8e`. Checkout original preservado intacto.
- **G-Ingá:** `main` remoto em `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`. Suíte `evaluate_girag_slicing.test.ts` executada localmente via bun test sobre a árvore `e13d49d` com 12 pass e 395 asserções.
- **AC2 concluído:** busca global via `git grep` em 535 arquivos de Gira-G e 347 de G-Ingá classificou exaustivamente todos os produtores, clonadores, serializadores, persistência e leitores.
- **Probes black-box observados:** executados com binário compilado em `/tmp/gira-audit00-DWtV3u` e isolamento estrito de variáveis de ambiente (`HOME`, `USERPROFILE`, `XDG_CACHE_HOME`, `GOCACHE`). Comprovado que schema inteiro 2 gera `ErrCorruptIndex` na CLI e `index_missing` no MCP; que string desconhecida é aceita por falta de allowlist na v1; que JSON malformado e estruturas com endpoints ausentes são rejeitados; e que campo extra `source_identity` é ignorado no decode mas perdido na serialização canônica v1.
- **Testes executados nesta sessão:**
  - `go test -v -count=1 ./internal/store ./internal/indexer ./internal/mcp ./cmd/gira` (PASS)
  - `cd bench && go test -v -count=1 ./layer0 ./layer1 ./layer2` (PASS)
  - `go test -count=1 ./internal/slices ./internal/domain ./internal/archdiff ./internal/skills ./internal/eval` (PASS)
  - `bun test tools/eval/evaluate_girag_slicing.test.ts` em G-Ingá `@e13d49d` (PASS)
- **Veredito:** AC1–AC5 e V1 satisfeitos com evidências reproduzíveis. A tarefa 00 está com auditoria integralmente concluída e pronta para transição formal para verified/closeout após revisão do commit local. ADR 0014 continua aguardando o gate da tarefa 00.
