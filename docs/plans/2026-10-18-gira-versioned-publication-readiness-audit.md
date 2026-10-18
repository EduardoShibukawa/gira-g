# Gira-G / G-Ingá: auditoria de readiness para publicação versionada

**Última revalidação remota:** 2026-10-18, nesta retomada. Os SHAs de `main`, o HEAD da PR #31 e os jobs de CI citados foram consultados separadamente.  
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
- HEAD remoto revalidado da branch: `7be227755fba688fac2dae602c352bab67e75408`, base `main` em `6c66d938148de56d460389f183bf9900f482f50f`.
- CI do HEAD atual: [run 39136093491](https://github.com/EduardoShibukawa/gira-g/actions/runs/39136093491), todos os jobs concluídos com sucesso, incluindo `test (ubuntu-latest)`, `test (macos-latest)`, `test (windows-latest)`, `validator`, `floor guard` e `gate-lint (warn-only)`. Docs [run 39136093575](https://github.com/EduardoShibukawa/gira-g/actions/runs/39136093575): `docs build` passou; `docs deploy` foi `skipped`, não falha.
- A PR #31 continua draft, aberta e sem merge. O diff contém os dois documentos preparatórios. Os checks verdes verificam a branch documental e o código-base existente; não executam nem aprovam SC1–SC10.
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

### Matriz de produtores, campos e consumidores conhecidos

**Base de código para os links abaixo:** Gira-G `main` `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `main` `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`. Os links apontam para os arquivos do repositório, não para execuções locais. A API de busca de código do conector não retornou resultados para consultas por símbolos; como não há checkout local, referências dinâmicas e ocorrências em diretórios de goldens não listados ficam marcadas como **não verificadas**, não presumidas ausentes.

#### Fluxo de dados observado

`internal/indexer.Build` produz `*domain.MacroMap` → cópias incrementais passam por `cloneMacroMap` → `CanonicalJSON` projeta e ordena os campos conhecidos → `store.PublishFile/Publish` grava snapshot legado → `store.LoadFile/Load` lê e valida → `cmd/gira/main.go:loadIndex` resolve CLI/servidor MCP → `slices.NewEngine` consome o mapa em memória. Há caminhos que leem arquivos diretamente: CLI `diff` e MCP `diff_maps`. O evaluator interno reconstrói o corpus e indexa novamente; ele não desserializa um MacroMap pré-construído. G-Ingá tem um fallback externo separado que lê o JSON diretamente.

#### Campos: produtor, cópia, serialização, leitura e risco

| Campo JSON atual / futuro | Código de referência | Comportamento remoto observado | Risco / cobertura necessária |
|---|---|---|---|
| `schema_version` | [`MacroMap`, `NewMacroMap`, `Validate`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L146-L253) | Campo Go é `string`; `NewMacroMap` emite `"1.0.0"`. `Validate` exige valor não vazio, não uma allowlist. | Schema inteiro `2` falha ao decodificar na estrutura legada e segue o caminho `ErrCorruptIndex`, não `index_schema_unsupported`. Uma versão desconhecida representada como string não é rejeitada apenas pela versão. |
| `target_system` | [modelo](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L146-L253), [store](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/store/store.go#L35-L80) | Campo obrigatório no modelo e usado no caminho do store por sistema; permanece no clone e na projeção canônica. | A identidade por nome não comprova que o mapa corresponde ao manifesto/binário que o produziu; isso é parte do contrato versionado pendente. |
| `nodes` | [`indexer.Build`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/indexer.go#L34-L130), [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L706-L775), [serializer](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L287-L348) | Construído no indexer; clone copia nodes, evidence e metadata; `CanonicalJSON` ordena nós por ID; engine/slices/export/nudge leem o mapa em memória. | Testes do schema novo devem verificar que atributos e evidências preservam bytes/semântica na publicação, clonagem e roundtrip. |
| `relationships` | [modelo/validação](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L146-L253), [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L706-L775) | Validação confere IDs e endpoints; clone preserva relações, evidence e metadata; projeção canônica ordena por ID. | Cobrir roundtrip e relações com referências inválidas no reader estrito; consumidores CLI/MCP não devem colapsar erros de schema em ausência. |
| `unsupported` | [modelo e serializer](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L146-L348), [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L706-L775) | Mantido no modelo, clonado, normalizado e ordenado na serialização canônica. | Não perder limites/diagnósticos ao adicionar `source_identity` ou criar outra projeção. |
| `unsupported_constructs` | [modelo/serializer](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/domain/models.go#L146-L348), [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L706-L775) | Slice copiada e ordenada deterministicamente na projeção. | Incluir na verificação de cobertura de campos; não inferir que `unsupported` a substitui, pois os campos têm semânticas distintas. |
| `unresolved_refs` | [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L744-L765), [regressões](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental_clone_test.go#L9-L100) | A PR #28 corrigiu perda anterior; a cópia mantém os campos e duplica slices de evidence. `CanonicalJSON` ordena por ID. | Correção legada já incorporada; preservar o campo e adicionar teste com objeto `source_identity` quando o novo modelo for implementado. |
| `conflicts` | [clone](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental.go#L744-L775), [regressões](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/indexer/incremental_clone_test.go#L9-L100) | A PR #28 corrigiu perda anterior e duplica a slice de evidence; serialização normaliza o campo. | Correção legada já incorporada; uma verificação de cobertura deve detectar futuras omissões em clone e serializer. |
| `source_identity` (proposto) | Ausente de [`domain.MacroMap`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/domain/models.go#L146-L348) | Não existe campo ou validação atual. O JSON decoder Go ignora campos extras ao decodificar para uma struct, e a projeção canônica recria apenas os campos conhecidos. | **Inferência do código atual:** um reader legada pode descartar esse campo extra; schema inteiro `2` não é aceito pelo modelo string. Novo schema precisa incluir, validar e preservar `source_identity` deliberadamente em modelo, clone, serializer, store e consumidores. |

#### Consumidores diretos e indiretos

| Local / símbolo | Tipo de consumidor | Comportamento observado e contrato a cobrir |
|---|---|---|
| [`cmd/gira/main.go:loadIndex`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/cmd/gira/main.go#L307-L355) | Leitor CLI central | Resolve explicit `--index` e descoberta por projeto/store legados via `store.LoadFile`. Consumido por comandos de slice/impact/explain/locate/export/view/tour/flows/doc e por `serve mcp`; precisa de reader estrito e erros estáveis. |
| [`cmd/gira/main.go:runDiff`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/cmd/gira/main.go#L1294-L1360) | Leitor CLI direto | Carrega os dois snapshots por `LoadFile`; não passa pela descoberta central. Deve preservar suporte a snapshots destacados e diferenciar ausente/corrompido/schema desconhecido. |
| [`internal/store/store.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f50f/internal/store/store.go#L83-L251) | Persistência e leitura | `PublishFile` faz escrita temporária, sync do arquivo e rename; `LoadFile` lê JSON para o modelo atual. `syncDirectory` retorna sucesso sem sync no Windows. Isso é store mutável legado, não o protocolo imutável de versões/heads; testes existentes de atomicidade não demonstram SC5–SC7. |
| [`internal/mcp/server.go:624-633`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/mcp/server.go#L624-L633) | Leitor MCP direto de snapshots | A ferramenta `diff_maps` usa `LoadFile` para `index_a/index_b`, mas reporta qualquer erro como `index_missing`. A taxonomia `index_schema_unsupported`/`index_corrupt` ainda não está implementada. |
| [`internal/skills/query.go:isStale`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/skills/query.go#L298-L310) | Consumidor indireto / freshness | A ponte de skills encaminha o caminho `--index`; `isStale` lê bytes crus e verifica se a string JSON do HEAD aparece no arquivo. Não interpreta `source_identity` nem metadata/head de versões; a política precisa ser substituída ou demonstrada compatível. |
| [`internal/skills/install.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/skills/install.go#L20-L40) | Instalação e configuração | Encaminha caminho opcional para `gira serve mcp --index`; não decodifica MacroMap. Os manifests/configs de skills e documentação devem acompanhar semântica CLI e avisos de legado. |
| [`internal/indexer/indexer.go:Build`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/indexer/indexer.go#L34-L130) | Produtor interno | Constrói o MacroMap a partir do manifesto/caches; identidade do produtor e lista completa de SHAs de fontes ainda não são parte do contrato legado. |
| [`internal/eval/eval.go:Corpus.Index`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/eval/eval.go#L383-L430) | Avaliação | Reindexa corpus via `indexer.Build`; não lê snapshot pré-construído diretamente. Compara saídas via `CanonicalJSON`. |
| [`internal/eval/scorecard.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/eval/scorecard.go#L98-L160) | Scorecard / hash | Calcula `OutputSHA256` sobre `CanonicalJSON` e métricas em memória; canonicalização nova pode alterar hash e precisa de regressões deliberadas. |
| [`internal/slices/engine.go:NewEngine`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/slices/engine.go#L151-L180), [`export.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/slices/export.go#L15-L45), [`nudge.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/skills/nudge.go#L19-L47) | Consumidores em memória / projeções | Engine, docs/export e hook consultam campos do mapa após carregamento; não são decodificadores, mas dependem da estrutura e da presença de evidência/locators. |
| [`bench/internal/benchtest/benchtest.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/bench/internal/benchtest/benchtest.go#L38-L76), [`bench/layer1/mcpclient.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/bench/layer1/mcpclient.go#L15-L56) | Test harness / MCP | O harness cria repositório temporário e chama `gira index --output <macro-map.json>`; Layer 1 fala com o servidor MCP, sem decodificar diretamente o arquivo. Deve provar semântica de snapshot destacado e integração do loader. |
| [fixture Layer 0](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/bench/layer0/testdata/gira/macro-map.json#L1-L40), [bootstrap offline](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/tools/index-bundle/bootstrap.sh#L1-L51) | Fixture e distribuição | A fixture conhecida é v1 string; bootstrap indexa para `--output macro-map.json` e copia cache/snapshot para uso offline. Precisa preservar ou migrar compatibilidade conforme a política explícita do resolver. |
| [G-Ingá MacroMap canônico](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/docs/architecture/macro-map.json#L1-L8), [fallback](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/tools/eval/evaluate_girag_slicing.ts#L266-L320), [testes](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/tools/eval/evaluate_girag_slicing.test.ts#L185-L280) | Consumidor externo confirmado | O fixture é v1; o fallback exige string `"1.0.0"`, valida top-level/nodes/relationships/IDs e rejeita inteiro 2, versão desconhecida, JSON malformado e estrutura inválida. Testes cobrem CLI indisponível e falha de CLI. Não implementar suporte a schema 2 neste lado. |
| README, website, docs versionadas, fixtures/goldens adicionais e consumidores em skills/plugins dinâmicos | Superfície documental ou dinâmica | Task 09 lista explicitamente README/website, skills, bench, bundle, goldens e fixtures. Não foi possível realizar uma busca global reproduzível de todos os nomes no checkout local nem via ferramenta de code-search nesta execução. **Ocorrências restantes ficam desconhecidas até grep local na task 09; não concluir que não existem.** |

#### Testes e evidência vinculados à matriz

| Cenário | Teste/artefato identificado | Evidência de execução observada |
|---|---|---|
| Clone dos oito campos atuais, inclusive diagnósticos e slices independentes | [`incremental_clone_test.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/indexer/incremental_clone_test.go#L9-L100) | O CI Gira-G no SHA atual base passou em [run 39133987360](https://github.com/EduardoShibukawa/gira-g/actions/runs/39133987360); o CI da branch documental também executou testes em [run 39136093491](https://github.com/EduardoShibukawa/gira-g/actions/runs/39136093491). Não é um teste de schema 2. |
| Parsing NUL, caminhos incomuns e falha de archive | [`gitcache_test.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f482f/internal/gitcache/gitcache_test.go#L280-L400) | Incluído no conjunto `go test -v -race ./...` dos jobs Ubuntu/macOS/Windows; fixture com tab/newline é pulada em Windows porque o filesystem não materializa esses nomes. |
| Leitura/publicação atual do store | [`store_test.go`](https://github.com/EduardoShibukawa/gira-g/blob/6c66d938148de56d460389f183bf9900f/internal/store/store_test.go) | Incluído no CI base. Cobertura existente não prova heads imutáveis, anti-downgrade, protocolo M→S→P ou `visible_unconfirmed`. |
| Fallback externo | [testes de slicing](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/tools/eval/evaluate_girag_slicing.test.ts#L185-L280) | [G-Ingá run 39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098): jobs `test` e `pre-commit`/validação passaram no HEAD `e13d49d`. |
| Matriz SC1–SC10 integrada | Tarefa [10](../tasks/2026-10-18-gira-versioned-publication-10-integrated-acceptance.md) | Não executada: não há implementação do protocolo versionado a testar. SCs continuam sem aceite. |

**Resultado da cobertura:** o inventário remoto identifica o produtor, todos os oito campos atuais, clone, serializer, store, leitores CLI/MCP, ponte/instalação de skills, avaliação/scorecard, engine/slices, bench/harness, fixture Layer 0, bundle e consumidor G-Ingá. A busca global restante, especialmente ocorrências dinâmicas e todos os goldens/docs, está explicitamente marcada **desconhecida** porque não há checkout local e a code-search remota não retornou resultados. Essa limitação é uma entrada de trabalho posterior, não uma afirmação de inexistência.

## Auto-revisão da task 00 (AC1–AC5)

| ID | Resultado da auto-revisão | Evidência / limite |
|---|---|---|
| AC1 | **pass com limite local declarado** | O relatório contém SHAs completos e URLs remotas de Gira-G/G-Ingá, PR #31 e CI. Checkout, branch e worktree local do usuário não estavam disponíveis; o relatório não afirma que foram inspecionados. |
| AC2 | **pass com desconhecidos explicitados** | A matriz acima inclui os oito campos atuais e `source_identity` futuro, caminhos identificados, símbolos, links para os SHAs remotos e comportamento atual/risco de downgrade. Consumidores dinâmicos e referências globais não pesquisáveis estão explicitamente classificados como desconhecidos. |
| AC3 | **pass** | Código e testes do fallback de G-Ingá foram revalidados na `main` `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`; schema string `"1.0.0"` aceito estruturalmente, inteiro 2 rejeitado; CI `39089480098` passou. |
| AC4 | **pass documental; teste proposto, não executado nesta tarefa** | [Experimento Windows 39084138230](https://github.com/EduardoShibukawa/gira-g/actions/runs/39084138230) registrou `git archive` exit 128 com `invalid path` para nome com tab e `tar -x` exit 2. A política é falhar fechado, não aceitar extração parcial, limpar o destino temporário e testar o erro através dos chamadores; não alegar que esse teste foi executado na main atual. |
| AC5 | **pass** | O diff da PR #31 contém apenas relatório e tarefa 00; não altera código, testes, fixtures, spec ou decisões de produto. Auto-revisão desta atualização comparou links/SHAs, distinções entre código e CI, matriz de consumidores e ressalvas. |

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

**P0 de auditoria documental: fechado no escopo verificável.** A task 00 tem a matriz remota e a auto-revisão AC1–AC5 registradas. As referências dinâmicas e a enumeração global de goldens/docs não observáveis sem checkout continuam explicitamente desconhecidas e passam para task 09, em vez de serem declaradas ausentes. Isso fecha a lacuna de planejamento da auditoria, não os gates técnicos do protocolo versionado.

**Continua bloqueado declarar a publicação versionada funcionalmente aceita:** schema 2, source_identity, store/heads, reader anti-downgrade, locks, durabilidade e SC1–SC10 continuam sem implementação/evidência de aceitação. A próxima tarefa é ADR 0014; task 02 é o trabalho técnico paralelo permitido pela sua dependência própria, sempre em branch separado e reutilizando #28.

### P1: bloqueia tarefas específicas posteriores

- **P1-1, ADR 0014 ausente:** task 01 permanece dependente da verificação formal da task 00; depois, criar ADR em um branch separado, sem alterar decisões aprovadas.
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

- [x] Task 00 AC1–AC5 evidenciados; todos os consumidores identificados foram mapeados e os dinâmicos/globais desconhecidos foram marcados explicitamente.
- [x] Sem alegação sobre worktree/teste local que não tenha sido observado.
- [ ] ADR 0014 criado em `docs/adr` e consistente com a spec aprovada.
- [ ] Task 02 testa parsing NUL, environment Git, clone/fetch e falha fechada no Windows, reutilizando #28.
- [ ] Grafo tasks 03–10 preservado; entradas, saídas e gates por tarefa claros.
- [ ] Cada SC tem teste, plataforma, resultado, evidência e condição objetiva de aceite; nenhum SC aprovado por inferência.
- [ ] Required checks confirmados antes do merge; o CI de `main` atual está verde, mas a implementação futura precisará de sua própria evidência de plataforma.
- [ ] Nenhuma implementação funcional/schema 2 foi alegada nesta fase.

## Primeiro passo operacional

**Próxima ordem:** (1) registrar formalmente o resultado e a auto-revisão AC1–AC5 na task 00; (2) após sua verificação formal, criar ADR 0014 em branch separado; (3) iniciar task 02 como trabalho técnico separado, reutilizando PR #28. Bases remotas desta retomada: Gira-G `6c66d938148de56d460389f183bf9900f482f50f`; G-Ingá `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`; HEAD PR #31 `7be227755fba688fac2dae602c352bab67e75408`.

1. Formalizar a transição da task 00 para `verified` usando a revisão documental como evidência e preservando os limites locais.
2. Criar ADR 0014 em branch separado, consolidando apenas decisões já aprovadas e mantendo questões de plataforma/implementação visíveis.
3. Iniciar task 02 em branch técnico separado: completar environment allowlist e provar a falha fechada/isolamento da extração Windows sem duplicar os fixes da #28.
4. Deixar schema 2, source_identity, store versionado/heads, reader, locks, durabilidade e integração CLI para as tarefas 03–10, conforme as dependências existentes.
5. Manter SC1–SC10 sem aceite até testes realmente executados no SHA integrado e nas plataformas exigidas.

## Veredito

**NOT READY para declarar a publicação versionada pronta para implementação integrada.** A auditoria documental remota foi ampliada, mas schema 2, source_identity, store/heads, reader estrito, locks, durabilidade e SC1–SC10 ainda não existem como implementação validada. A matriz também mantém como desconhecidas as referências dinâmicas/globais que não puderam ser consultadas sem checkout local. É possível preparar ADR 0014 e iniciar o trabalho delimitado da task 02 somente conforme suas dependências formais; nada aqui autoriza merge, release ou suporte a schema 2.

**Próxima ação:** registrar o resultado e a auto-revisão AC1–AC5 na task 00, preservando explicitamente suas limitações; depois seguir a ordem de dependências.
