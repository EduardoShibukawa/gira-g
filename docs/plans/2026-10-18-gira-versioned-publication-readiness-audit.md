# Gira-G / G-Ingá: auditoria de readiness para publicação versionada

**Última atualização da evidência remota registrada:** 2026-10-18T14:25:24Z (UTC). A matriz e o veredito distinguem o estado de `main` dos checks desta PR documental.  
**Escopo:** APIs do GitHub, branches, PRs, workflows e código/documentação dos dois repositórios.  
**Limitação local:** o ambiente desta execução não contém checkout Git (`git rev-parse --show-toplevel` retorna “not a git repository”). Nenhum teste local foi executado; não é possível afirmar branch, HEAD ou limpeza do worktree na máquina do usuário.  
**Alterações propostas neste branch:** correção documental do plano da task 00 e este relatório. Sem alterações de código/testes de produto, sem mudança de spec, sem aprovação de SC, sem merge/release.

Este é um adendo de revalidação. O relatório anterior [gira-schema-consumer-compatibility-audit](2026-10-18-gira-schema-consumer-compatibility-audit.md) é preservado como registro histórico. Suas conclusões anteriores à PR #28 e à aprovação da spec precisam ser lidas à luz dos fatos abaixo.

## Estado verificado

| Repositório | Branch padrão / SHA atual | CI mais recente observado na consulta | Conclusão |
|---|---|---|---|
| [Gira-G](https://github.com/EduardoShibukawa/gira-g) | `main` / `6c66d938148de56d460389f183bf9900f482f50f` | [CI run 39133987360](https://github.com/EduardoShibukawa/gira-g/actions/runs/39133987360): todos os jobs concluídos em sucesso, incluindo Ubuntu, macOS, Windows, validator, floor guard e gate-lint (atualizado em 2026-10-18T14:09:29Z). [Docs run 39133987482](https://github.com/EduardoShibukawa/gira-g/actions/runs/39133987482) em sucesso. | CI integral do SHA atual terminou com sucesso nos três SOs. Isso valida o código atual, não a implementação futura da publicação versionada. |
| [G-Ingá](https://github.com/EduardoShibukawa/G-Inga) | `main` / `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0` | [Run 39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098): jobs `test` e `pre-commit` em sucesso. | CI existente do HEAD atual passou; isso não significa suporte ao schema 2. |

No momento da consulta, as listas de PRs abertas estavam vazias nos dois repositórios. O endpoint de branch protection retornou HTTP 403 para a integração; embora `main` apareça como protegido nas listas de branches, a lista exata de checks obrigatórios não pôde ser inspecionada. Existem branches remotas históricas/experimentais, logo “nenhuma PR aberta” não significa “nenhuma branch”.

### PRs históricas

| PR | Estado remoto observado | Resultado relevante |
|---|---|---|
| [Gira-G #17](https://github.com/EduardoShibukawa/gira-g/pull/17) | merged; merge `61b58c281f5a7fae4092f0cbff2ecfd8794dea77` | Isolamento HOME/CWD em testes de store. |
| [Gira-G #25](https://github.com/EduardoShibukawa/gira-g/pull/25) | merged; merge `5a24a361841fdbe0ef85ce41987d2a17ba87b45b` | Pacote documental de compatibilidade/regressão; não prova que os testes de publicação foram executados. |
| [Gira-G #26](https://github.com/EduardoShibukawa/gira-g/pull/26) | merged; merge `ec436cf331369f9809cb2a5821c31eee926404b6` | Auditoria de consumidores/schema. |
| [Gira-G #28](https://github.com/EduardoShibukawa/gira-g/pull/28) | merged; merge = HEAD atual | Corrigiu clone de campos diagnósticos e também alterou parsing Git, reaping do `tar`, isolamento de ambiente e a ordem da publicação de `gira.yaml`. |
| [Gira-G #29](https://github.com/EduardoShibukawa/gira-g/pull/29) | fechada sem merge | Não reabrir automaticamente. |
| [Gira-G #30](https://github.com/EduardoShibukawa/gira-g/pull/30) | merged; merge `1bde30f36d465277b97e45a997b3a79f52b81130` | Aprovação/planejamento e tarefas; não implementa publicação versionada. |
| [G-Ingá #47](https://github.com/EduardoShibukawa/G-Inga/pull/47) | fechada sem merge | Supersedida; não reabrir nem integrar automaticamente. |
| [G-Ingá #105](https://github.com/EduardoShibukawa/G-Inga/pull/105) | merged; merge `860792469bf0c7b56d37d98c8dc07079ce816928` | Evidência documental. |
| [G-Ingá #106](https://github.com/EduardoShibukawa/G-Inga/pull/106) | merged; merge `887f25adfe9c55ccd1d1b80082ac86ecd53d846d` | Contrato de compatibilidade fallback. |
| [G-Ingá #108](https://github.com/EduardoShibukawa/G-Inga/pull/108) | merged; merge = HEAD atual | Validação estrita do fallback do evaluador. |

## Branch de auditoria e CI da documentação

- PR documental: [Gira-G #31](https://github.com/EduardoShibukawa/gira-g/pull/31), draft, aberta, sem merge; branch `audit/2026-10-18-versioned-publication-readiness`.
- Snapshot em 2026-10-18T14:25:24Z: branch HEAD `0a364bc69143625ca5ab053d3962910be5af73ce`; o run de docs [39135912931](https://github.com/EduardoShibukawa/gira-g/actions/runs/39135912931) estava `in_progress` e CI [39135912845](https://github.com/EduardoShibukawa/gira-g/actions/runs/39135912845) estava `queued`.
- Revisões anteriores desta mesma PR passaram: docs [39135819196](https://github.com/EduardoShibukawa/gira-g/actions/runs/39135819196) em sucesso para o commit `dab35757bd9c0df4e9f6d974044a3435138dc3e8`; CI [39135394343](https://github.com/EduardoShibukawa/gira-g/actions/runs/39135394343) em sucesso para `461d3ae73c5e5e16aaaac9842c3fee41895b5c19`. Estes sucessos não substituem o check do HEAD mais recente.
- A PR contém apenas os dois arquivos preparatórios referidos no corpo: este relatório e a tarefa 00 corrigida para ser estritamente read-only. A tarefa 00 continua `todo`; seus ACs continuam `not run`.
- Nenhum merge, alteração de `main`, mudança de código de produto, mudança de contrato ou suporte ao schema 2 foi realizado nesta fase.

## Evidências de código e contrato

### MacroMap, clone, serialização e store

Fontes atuais: [models.go](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/domain/models.go), [incremental.go](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/indexer/incremental.go), [store.go](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/store/store.go), [main.go](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go).

- O `MacroMap` de Gira-G tem oito campos atuais: `schema_version`, `target_system`, `nodes`, `relationships`, `unsupported`, `unsupported_constructs`, `unresolved_refs` e `conflicts`. `SchemaVersion` segue string e `NewMacroMap` emite `"1.0.0"`. Não há `source_identity`.
- `Validate()` exige versão não vazia e integridade estrutural do grafo; não impõe allowlist estrita de schema. Strings desconhecidas não são rejeitadas só por serem versões desconhecidas.
- `CanonicalJSON()` recompõe os campos atuais explicitamente e os ordena deterministicamente. A futura `source_identity` precisa ser adicionada ao modelo, clone, serializer e consumidores, ou será perdida/não existirá.
- A PR #28 já corrigiu `cloneMacroMap()` para copiar `UnresolvedRefs` e `Conflicts`, inclusive evidências. Os testes `TestCloneMacroMapPreservesUnresolvedRefsAndConflicts` e `TestCloneMacroMapPreservesNilDiagnosticSlices` existem no HEAD atual. A auditoria anterior que descreveu esses dois campos como perdidos está desatualizada em relação ao HEAD após #28.
- O store continua sendo o caminho legado baseado em arquivo: não há ainda publicação versionada imutável, `version_id`, heads por linhagem ou reader estrito com a taxonomia normativa. O `LoadFile` decodifica para o modelo atual e valida; a spec exige distinguir schema não suportado de corrupção/arquivo ausente.
- Os caminhos de CLI para descoberta e `--index` ainda precisam receber a política final do reader versionado. A existência de carregamento explícito não demonstra por si só que o legado seja apenas read-only com aviso nem que não ocorra downgrade.

### Git, ambiente, subprocessos e init

Fontes atuais: [cache.go](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/gitcache/cache.go), [main.go](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go), [testes do gitcache](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/gitcache/gitcache_test.go), [testes da CLI](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main_test.go).

A PR #28 já introduziu:
- `git ls-tree -z` e parsing NUL-delimited; existe `TestCacheManager_ListTreePreservesUnusualPathnames`.
- Reaping de `tar` mesmo quando `git archive` falha; existe `TestCacheManager_ExportTreeReapsTarWhenArchiveFails`.
- Saneamento de variáveis que redirecionam probes Git locais; existe `TestGitProbesIgnoreRepositoryRedirectEnvironment`.
- Staging de manifesto novo e publicação de `gira.yaml` somente depois de uma indexação bem-sucedida; existem `TestInitDoesNotPublishManifestWhenIndexingFails` e `TestInitForceDoesNotRepinManifestWhenIndexingFails`.

Logo, a tarefa 02 deve reutilizar essas mudanças e completar suas lacunas, não duplicá-las. O experimento Windows registrado nas [decisões de gates](2026-10-17-gates-tecnicos-publicacao-versionada.md), [run 39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230), relatou `git archive` com erro 128 em um nome com tab (`invalid path`) e `tar -x` com erro 2. Isso é evidência experimental de um branch/runner separado, não um teste executado nesta auditoria nem prova de correção no HEAD atual. Ainda é necessário provar que caminho não materializável resulta em erro claro, falha fechada e nunca em extração parcial aceita como sucesso pelos chamadores.

### G-Ingá

Fontes atuais: [evaluate_girag_slicing.ts](https://github.com/EduardoShibukawa/G-Inga/blob/main/tools/eval/evaluate_girag_slicing.ts) e [testes](https://github.com/EduardoShibukawa/G-Inga/blob/main/tools/eval/evaluate_girag_slicing.test.ts).

O fallback aceita somente string `"1.0.0"`; valida o objeto superior, arrays `nodes`/ `relationships`, IDs de nós e endpoints usados na travessia. Os testes rejeitam schema ausente, null, vazio, booleano, objeto, array, inteiro 2, inteiro desconhecido, string desconhecida e representação `"2"`; verificam a mesma rejeição quando a CLI está ausente ou falha. JSON malformado, estrutura inválida e nó ausente têm erros distintos.

**Conclusão:** G-Ingá continua v1-only e o schema inteiro `2` deve continuar não suportado. A validação do fallback é uma defesa correta do consumidor atual, não evidência da publicação versionada do produtor.

### Mapa adicional de consumidores e caminhos indiretos

Esta tabela amplia o mapa de código consultado. Ela não afirma que o inventário esteja completo: os consumidores restantes, especialmente o caminho das skills e eventuais leitores adicionais encontrados por busca no checkout, continuam dentro do AC2 da task 00.

| Local / símbolo | Relação com MacroMap/publicação | Comportamento atual / risco a cobrir |
|---|---|---|
| [`cmd/gira/main.go`, `loadIndex`](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go) | Leitor central de comandos de CLI; resolve `--index`, projeto `gira.yaml`, caches `.gira/indexes` e `~/.cache/gira/indexes`. | Usa `store.LoadFile`/store legado. Descoberta por nome de target system e arquivos `macro-map.json` precisam ser adaptados ao resolver estrito; erros ainda não têm a taxonomia final. Chamado por impact/slice/explain/locate/export/view/tour/flows/doc e `serve mcp`. |
| [`cmd/gira/main.go`, `runDiff`](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go) | Carrega dois snapshots explícitos para comparação. | Chama `store.LoadFile` diretamente para ambos; precisa manter snapshots destacados compatíveis e distinguir corrupção/schema não suportado. |
| [`cmd/gira/main.go`, `runServe`](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go) e [`internal/mcp/server.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/mcp/server.go) | Inicialização do engine MCP e ferramentas externas (impact, slice, explain, locate etc.). | `serve mcp` chama `loadIndex` uma vez e cria `slices.NewEngine(mm)`; o engine mantém o mapa carregado. A ferramenta `diff_maps` também chama `store.LoadFile` para `index_a/index_b`, mas converte qualquer falha em `index_missing`, perdendo distinção entre schema desconhecido e corrupção. |
| [`internal/skills/query.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/skills/query.go) | Ponte `gira-query` para MCP ou CLI; encaminha `--index` e padroniza erros. | Não decodifica o mapa principal, mas `isStale` lê o arquivo cru e procura o SHA da HEAD por substring JSON citada. A lógica precisa ser revista para a forma versionada e para proveniência/staleness estrita; erro novo de schema pode cair em `internal` sem código público dedicado. |
| [`internal/skills/install.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/skills/install.go), [`manifest.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/skills/manifest.go) e catálogo [`internal/skills/catalog`](https://github.com/EduardoShibukawa/gira-g/tree/main/internal/skills/catalog) | Instala skills/configuração MCP e gera manifesto de skills. | A opção `Index` é repassada como `gira serve mcp --index`; a instalação não parseia MacroMap. Confirmar que os docs/skills não descrevam um contrato antigo depois da mudança do CLI. |
| [`internal/eval/eval.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/eval/eval.go) / [`internal/indexer/indexer.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/indexer/indexer.go) | Produtor de avaliação e indexação: avalia o corpus por meio do mesmo caminho de build que a CLI. | O comentário de `eval.go` explicita que o avaliador interno não desserializa MacroMap preconstruído; ele indexa o corpus pelo `indexer.Build`. Isso reduz um leitor de arquivo, mas torna determinismo/identidade do produtor uma dependência do scorecard e da avaliação. |
| [`internal/eval/scorecard.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/eval/scorecard.go) | Consumidor em memória de mapas publicado/perturbado para métricas. | Calcula `OutputSHA256` usando `mm.CanonicalJSON()` e métricas pelos campos atuais. Deve acompanhar mudança de canonicalização/schema, sem confundir esse scorecard com um leitor de snapshot em disco. |
| [`bench/internal/benchtest/benchtest.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/bench/internal/benchtest/benchtest.go) e [`bench/layer1/mcpclient.go`](https://github.com/EduardoShibukawa/gira-g/blob/main/bench/layer1/mcpclient.go) | Test harness cria repo Git temporário, compila `gira`, indexa para `--output`; layer 1 chama tools via MCP. | São consumidores indiretos dos contratos de `index --output`, CLI, loader e MCP. Precisam ter regressão que prova que outputs destacados seguem válidos sem ativar store/head canônico. |
| [`bench/layer0/testdata/gira/macro-map.json`](https://github.com/EduardoShibukawa/gira-g/blob/main/bench/layer0/testdata/gira/macro-map.json) e [fixture layer 1](https://github.com/EduardoShibukawa/gira-g/blob/main/bench/layer1/testdata/scorer-fixture.json) | Fixtures de avaliação/score, incluindo um mapa legado v1 no layer 0. | A fixture `"1.0.0"` deve permanecer útil para os casos de compatibilidade v1 explicitamente suportados; localizar testes que a leem e definir a expectativa de aviso/read-only no novo resolver. O fixture scorer contém grafo de resposta/universo, não é leitor direto de MacroMap. |
| [`tools/index-bundle/bootstrap.sh`](https://github.com/EduardoShibukawa/gira-g/blob/main/tools/index-bundle/bootstrap.sh) | Bootstrap de cache Git e snapshot para ambientes offline. | Roda `gira index ... --output macro-map.json` e repete com `--offline`. A nova semântica de snapshot destacado precisa manter esse fluxo; verificar se o script deve também validar versão/identidade no bundle. |
| [G-Ingá canonical MacroMap](https://github.com/EduardoShibukawa/G-Inga/blob/main/docs/architecture/macro-map.json) e fallback do evaluador | Consumidor externo com fixture canônica v1. | O fixture continua `schema_version: "1.0.0"`; fallback em TS recusa inteiro `2`. Não ativar schema 2 antes do producer/consumer gate separado. |

**Nota sobre evidência:** os arquivos acima foram buscados remotamente no `main` corrente; os links são fontes de código, não registros de execução local. O run de CI verde cobre os testes que o workflow já executou, mas não demonstra todos os cenários específicos listados em SC1–SC10. Para fechar AC2, é necessário completar a busca de referências e consumers restantes com um checkout acessível, incluir cada call site identificado e executar/referenciar os testes correspondentes no SHA efetivamente testado.

## Spec e ADR 0014

- [Spec versionada](https://github.com/EduardoShibukawa/gira-g/blob/main/specs/2026-10-17-gira-versioned-publication.md): `status: approved`, versão aprovada 1, hash `sha256:29bfea220a465391d8dae6e4954f01e7777c66df7d80bcf4ff2b553eea20d2d2`. A PR #30 registra autorização heavy e criação das tarefas-filhas.
- N1–N4, D1–D9c e decisões dos Gates 1–5 estão registradas em [decisões arquiteturais](2026-10-17-decisao-arquitetural-publicacao-versionada.md) e [gates técnicos](2026-10-17-gates-tecnicos-publicacao-versionada.md). Não são tratadas como propostas quando o registro do dono as marca como decididas; o que continua pendente é a evidência técnica e os gates.
- Não existe ADR 0014 na árvore atual de `docs/adr` (a série remota vai até 0013). A [task 01](../tasks/2026-10-18-gira-versioned-publication-01-adr-0014.md) está `todo`, depende da task 00 e deve consolidar o contrato sem alterar decisões sem nova aprovação.

## Grafo de dependências confirmado

Os arquivos existem no HEAD atual; a tarefa-mãe registra os children como `todo`.

| Tarefa | Depende de | Papel |
|---|---|---|
| [00 audit-revalidation](../tasks/2026-10-18-gira-versioned-publication-00-audit-revalidation.md) | nenhuma | Inventário e evidência atuais; pré-requisito para 01. |
| [01 adr-0014](../tasks/2026-10-18-gira-versioned-publication-01-adr-0014.md) | 00 | Consolidar decisões/contrato aprovado. |
| [02 gitcache-env](../tasks/2026-10-18-gira-versioned-publication-02-gitcache-env.md) | auditoria 00; incorporar #28 | Parsing de caminhos, ambiente, clone/fetch e erro fail-closed Windows. |
| [03 locator](../tasks/2026-10-18-gira-versioned-publication-03-locator.md) | 01, 02 | Raiz Git e locator do manifesto. |
| [04 keys](../tasks/2026-10-18-gira-versioned-publication-04-keys.md) | 01, 02, 03 | `source_key`, `tool_id`, `source_identity`. |
| [05 locks](../tasks/2026-10-18-gira-versioned-publication-05-locks.md) | 01 | Locks M→S→P; a tarefa permite paralelismo com 03/04. |
| [06 publish](../tasks/2026-10-18-gira-versioned-publication-06-publish.md) | 04, 05 | Store imutável, heads, publicação e durabilidade. |
| [07 resolver](../tasks/2026-10-18-gira-versioned-publication-07-resolver.md) | 06, 04 | Reader estrito, legado, erros e anti-downgrade. |
| [08 cli](../tasks/2026-10-18-gira-versioned-publication-08-cli.md) | 05, 07 | `index`, `init` e fluxos CLI. |
| [09 docs-consumers](../tasks/2026-10-18-gira-versioned-publication-09-docs-consumers.md) | 08 | Consumidores, fixtures e docs. |
| [10 integrated-acceptance](../tasks/2026-10-18-gira-versioned-publication-10-integrated-acceptance.md) | todas as anteriores | SC1–SC10, CI e revisão integrada. |

A [revisão M0–M6](../tasks/2026-10-17-m0-m6-retro-review.md) é separada, read-only e não constitui dependência desta sequência.

## Matriz de SCs

A redação normativa completa permanece na seção Acceptance Criteria da [spec aprovada](../specs/2026-10-17-gira-versioned-publication.md). As obrigações centrais, seu estado remoto atual e as etapas responsáveis são:

| SC | Obrigação original resumida fielmente | Evidência atual / lacuna | Estado |
|---|---|---|---|
| SC1 | Determinismo de invocação via caminhos relativo/absoluto/subdiretório/symlink/worktree/submódulo; matriz adversarial de cwd, env, quotePath, refs/replace e nomes especiais; root Git não sequestrada; falha fechada fora da raiz quando overrides exigem locator. | Existem testes de nomes incomuns e ambiente, mas não evidência da matriz completa nem dos bytes idênticos por todas as invocações/plataformas. | **Não aceito;** task 00 para inventário, 02/03 para implementação/teste. |
| SC2 | Leituras Git independentes de config para tab/não ASCII; quotePath/refs/replace não alteram bytes; clone/fetch mantêm configuração do usuário. | `ls-tree -z` e teste de caminho incomum existem; a matriz completa de clone/fetch não está evidenciada por um run final de três SOs no HEAD atual. | **Não aceito;** task 02. |
| SC3 | Chaves e codificação injetiva; hash de executable em dev build; version_id idêntico com bytes diferentes retorna `nondeterministic_build`; schema inteiro 2 com `source_identity` autocontido; reader estrito; legado v1 somente por `--index` explícito, read-only e com aviso. | Modelo ainda usa string `"1.0.0"`; `source_identity`, `source_key` e store versionado não existem. #28 preserva os campos legados conhecidos. | **Não implementado/não executado;** tasks 04, 06, 07 e 09. |
| SC4 | S1/S2 dão `index_missing`, mismatch de ferramenta/schema têm erros distintos, publicação proíbe downgrade, legado nunca carrega implicitamente por `gira.yaml`, snapshots destacados compatíveis, eval não inventa `gira.yaml` e retorna `manifest_path_required` quando necessário. | Ainda não há reader versionado nem evidência de toda a taxonomia de erros. O fallback G-Ingá implementa somente schema v1. | **Não implementado/não executado;** tasks 07 e 09. |
| SC5 | Falha pré-rename não deixa versão visível; falha pós-rename preserva head e permite retry idempotente; gate rejeitado não grava; head corrompido/ausente falha fechado. | Protocolo de versão imutável/head não existe no store atual; cenários de falha não foram executados para esse contrato. | **Não implementado/não executado;** task 06. |
| SC6 | Unix: falha de sync gera `visible_unconfirmed`, código de saída não zero e manifesto não gravado no init; Windows: `visible_unconfirmed`, saída 0 e aviso. | Política de durabilidade do novo protocolo não existe no store atual; nenhum ensaio de falha/sync ligado ao SC6 foi evidenciado. | **Não implementado/não executado;** tasks 06 e 10. |
| SC7 | Travas em três SOs: exclusão entre processos e grafias; liberação após morte de filho; timeout `index_locked`; erro injetado `lock_unavailable` sem gravação; ordem M→S→P. | Não há protocolo M→S→P nem matriz de concorrência da publicação versionada no código atual. | **Não implementado/não executado;** tasks 05 e 10. |
| SC8 | Init que falha não cria manifesto; init --force rejeitado não altera manifesto/head/P; init --output produz snapshot canônico e P byte a byte iguais; index --output não toca store; mutação externa do manifesto em init dá conflito. | #28 já adicionou staging e testes para não publicar/repinar manifesto em falha de indexação. As semânticas de output/store versionado e conflito concorrente continuam sem evidência. | **Parcial no legado, gate não aceito;** tasks 06, 08 e 10. |
| SC9 | README/CLI/skills/bench/tools descrevem comportamento implementado; gira-arch-diff, bench/phases, index-bundle e benchtest continuam funcionais; validador e docs build passam. | Build de docs atual passou, mas documenta a situação anterior à publicação versionada. Inventário total dos consumidores e fixtures ainda está aberto. | **Parcial, não aceito;** tasks 00 e 09. |
| SC10 | `go test -race ./...` passa em Ubuntu/macOS/Windows; floor guard passa; nenhum teste escreve no cache real de usuário. | Ubuntu/macOS/Windows e floor guard passaram no run base atual. Não há prova de SC integrado nem hash do cache real antes/depois. | **Não aceito;** task 10. |

A presença de um teste não equivale à sua execução nem à aprovação do SC. Os resultados acima nunca foram marcados como aceite de publicação versionada.

## Testes, setup e plataformas

- Gira-G usa [`mise.toml`](https://github.com/EduardoShibukawa/gira-g/blob/main/mise.toml): Go 1.24.0, Bun 1.2.4; `mise run test` executa `go test -v -race ./...`, `mise run check` executa `go test -race ./...`, `mise run docs:build` constrói o site. Nenhum desses comandos foi executado localmente nesta sessão.
- Gira-G CI tem jobs Ubuntu/macOS/Windows, validator e floor guard; gate-lint está configurado como warn-only/continue-on-error. Todos os jobs base observados passaram; CI verde do código atual não equivale a SC1–SC10 aprovados.
- G-Ingá `package.json` define testes Bun para validate/gate/eval/opencode-plugin e scripts de validação/typecheck; os jobs remotos existentes passaram no HEAD atual.
- O endpoint de branch protection inacessível impede confirmar diretamente a lista exata de checks obrigatórios.

## Lacunas P0–P3

### P0: bloqueia declarar readiness para iniciar implementação funcional

**P0-1: Task 00 permanece `todo` e o inventário completo não foi comprovado.** Esta auditoria revalidou fontes centrais e corrigiu o entendimento de #28, mas não terminou a rastreabilidade individual de todo leitor indireto/dinâmico em skills/check, bench/benchtest, bundle, eval/scorecard, fixtures/goldens e docs. Task 01 depende de task 00.  
**Ação de fechamento:** completar a matriz de consumidores em checkout/acesso remoto verificável, registrar arquivo/símbolo/SHA/teste/resultado, concluir AC1–AC5 com auto-revisão e manter `todo` enquanto houver lacunas. É preparatório/read-only.

### P1: bloqueia tarefas específicas posteriores

- **P1-1, ADR 0014 ausente:** task 01 está `todo`; criar ADR documentado depois da auditoria 00, sem alterar decisões aprovadas sem nova aprovação.
- **P1-2, Git for Windows e extração parcial:** acrescentar teste específico de nome com tab e comprovar erro claro/fail-closed, limpeza/isolamento da extração parcial e comportamento de todos os chamadores. Task 02; não duplicar os fixes da PR #28.
- **P1-3, contrato versionado não implementado:** schema 2, `source_identity`, keys, store/heads, reader estrito, locks e durabilidade são trabalho funcional explícito nas tasks 03–09.
- **P1-4, worktree local não observado:** antes de comandos que alterem um checkout local, obter acesso real e registrar branch/HEAD/status sem reset/clean. Não presumir que o worktree do usuário está limpo.

### P2: bloqueia merge/aceitação final

- **P2-1:** confirmar required checks com permissão apropriada, pois API de proteção retornou 403.
- **P2-2:** SC1–SC10 e verificação de cache real intacto seguem sem aceite; exigem testes no head integrado e nas plataformas aplicáveis.

### P3: melhoria não bloqueante

- **P3-1:** revisão M0–M6 é independente, read-only e não bloqueia a sequência de publicação.
- **P3-2:** manter este adendo como revalidação datada; preservar o relatório anterior sem apagar seu histórico.

## Checklist de entrada

- [ ] Task 00 AC1–AC5 evidenciados e inventário total de consumidores revisado.
- [ ] Sem alegação sobre worktree/teste local que não tenha sido observado.
- [ ] ADR 0014 criado em `docs/adr` e consistente com a spec aprovada.
- [ ] Task 02 testa parsing NUL, environment Git, clone/fetch e falha fechada no Windows, reutilizando #28.
- [ ] Grafo tasks 03–10 preservado; entradas, saídas e gates por tarefa claros.
- [ ] Cada SC tem teste, plataforma, resultado, evidência e condição objetiva de aceite; nenhum SC aprovado por inferência.
- [ ] Required checks confirmados antes do merge; o CI de `main` atual está verde, mas a implementação futura precisará de sua própria evidência de plataforma.
- [ ] Nenhuma implementação funcional/schema 2 foi alegada nesta fase.

## Primeiro passo operacional

**Continuar pela task 00, read-only, que permanece `todo`.** Base remota no momento verificado: Gira-G `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`. Revalidar os SHAs ao iniciar trabalho posterior.

1. Terminar inventário de leitores: `store.LoadFile/PublishFile`, CLI discovery e `--index`, MCP, skills install/check, bench, benchtest, bundle, eval/scorecard, fixtures, goldens e docs; mapear campos e erros em cada caminho.
2. Cobrir o contrato legada v1 string, inteiro 2, schema desconhecido, estrutura inválida, arquivo ausente, mismatch de ferramenta e `source_identity` futuro separadamente.
3. Reutilizar as correções e testes da #28; a auditoria não deve editar código nem testes de produto.
4. Registrar testes existentes como existentes e resultados só quando houver execução por comando/plataforma/SHA. O experimento Windows mencionado é evidência externa ao HEAD, não aceite.
5. Fechar AC1–AC5 e auto-revisar. Só então dar a task 01 como entrada para ADR 0014; avaliar task 02 conforme sua dependência explícita, sem reordenar silenciosamente o plano.

## Veredito

**NOT READY para iniciar a implementação funcional como se a preparação estivesse concluída.** Há spec aprovada, decisões registradas e uma sequência de tarefas real, mas task 00 continua `todo`, o inventário integral ainda não está demonstrado, ADR 0014 não existe e todos os SCs continuam sem aceite. É seguro continuar a preparação read-only; não é correto declarar a missão concluída.

**Próxima ação única:** completar a matriz de consumidores da task 00 com evidência por arquivo/símbolo e SHA para os leitores restantes, depois auto-revisar AC1–AC5.
