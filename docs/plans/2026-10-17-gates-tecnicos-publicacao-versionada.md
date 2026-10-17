# Gates técnicos da publicação versionada: evidência registrada

| Campo | Valor |
|---|---|
| Data | 2026-10-17 |
| Tipo | Registro de evidência, somente documentação |
| Estado | Decisões do dono registradas (2026-10-17). Spec `draft`. Tarefa-mãe `todo`. **Implementação não autorizada** |
| Base | `25882da` (main) |
| Spec | [`2026-10-17-gira-versioned-publication.md`](../../specs/2026-10-17-gira-versioned-publication.md), `## Open questions` itens 1–5 |
| Decisão | [`2026-10-17-decisao-arquitetural-publicacao-versionada.md`](2026-10-17-decisao-arquitetural-publicacao-versionada.md), `## Bloqueios antes da aprovação final`, itens 1–5 |
| Tarefa-mãe | [`2026-10-17-gira-versioned-publication.md`](../../tasks/2026-10-17-gira-versioned-publication.md) |

Este documento registra evidência e, no adendo ao final, as decisões do dono de 2026-10-17. Não aprova a spec, não adiciona `confirmed: heavy`, não cria ADR 0014 nem tarefas-filhas e não altera código de produto. Os achados técnicos continuam sujeitos aos testes de aceitação listados; Gate 5 é verificado dentro da primeira tarefa de implementação autorizada, não antes dela.

## Legenda das classes de evidência

| Tag | Significado |
|---|---|
| `[CÓDIGO INSPECIONADO]` | Lido no código em `25882da`, com `file:line` |
| `[TESTE DEFINIDO]` | Teste descrito, ainda não escrito nem executado |
| `[TESTE EXECUTADO]` | Experimento rodado nesta máquina (Linux, go1.24.0, git 2.55.0), com `HOME`, `USERPROFILE` e `XDG_CACHE_HOME` isolados em `/tmp`. O cache real não foi tocado |
| `[CI]` | Resultado de runner de CI |
| `[RELATO]` | Informação de terceiros ou de fonte não reexecutada |
| `[INFERÊNCIA]` | Conclusão ou recomendação do agente |
| `[DESCONHECIDO]` | Não verificado |
| `[DECISÃO APROVADA]` | Decisão do dono registrada em documento |

Limitação comum: todos os experimentos foram em Linux. LSP não estava disponível, então os inventários usaram `grep` e `git grep`; referências dinâmicas ou por reflexão podem ter escapado `[DESCONHECIDO]`.

## Resumo de status

| Gate | Tema | Status |
|---|---|---|
| 1 | Contrato do mapa e consumidores | direção escolhida; compatibilidade/consumidores ainda exigem implementação e testes |
| 2 | Manifesto fora da raiz Git (D3) | política escolhida (O1); testes e saneamento do ambiente ainda pendentes |
| 3 | `main.date` (D5) | opção C escolhida; regressão de `tool_id` ainda pendente |
| 4 | Windows | W1 escolhido para o perfil de CI testado; limite documentado, critérios de SC7 ainda necessários |
| 5 | Durabilidade e publicação por plataforma | reservado como gate de aceitação da implementação, com testes nos três SOs |

---

## Gate 1: contrato do mapa e consumidores

**Pergunta.** Spec `## Open questions` item 1; decisão §Bloqueios item 1: inventariar os consumidores de `MacroMap`/JSON, definir o valor de `schema_version`, a migração e o comportamento de parsers antigos, e demonstrar que nenhum consumidor perde a proveniência.

### Evidência: contrato atual

| Fato | Evidência |
|---|---|
| `domain.MacroMap` tem 8 campos (`schema_version`, `target_system`, `nodes`, `relationships`, `unsupported`, `unsupported_constructs`, `unresolved_refs`, `conflicts`) | `[CÓDIGO INSPECIONADO]` `internal/domain/models.go:146-157` |
| `schema_version` nasce só em `NewMacroMap` com `"1.0.0"` | `[CÓDIGO INSPECIONADO]` `models.go:199`; usado em `extractors/pipeline.go:78` |
| Nenhum leitor rejeita versão diferente: `Validate` só exige string não vazia; `store.LoadFile` usa `json.Unmarshal` sem `DisallowUnknownFields` | `[CÓDIGO INSPECIONADO]` `models.go:213`, `store.go:217` |
| Único código que compara a versão é `indexer/incremental.go:645` (força `RefreshFull`); só testes chamam `BuildIncremental` | `[CÓDIGO INSPECIONADO]` |
| `CanonicalJSON` e `cloneMacroMap` reenumeram campos à mão: campo novo some em silêncio se não for adicionado. `cloneMacroMap` já descarta `UnresolvedRefs` e `Conflicts` (defeito latente) | `[CÓDIGO INSPECIONADO]` `models.go:309-336`, `incremental.go:706-754` |
| Valores de `schema_version` no repo: `"1.0.0"` em `models.go:199`, `bench/layer0/testdata/gira/macro-map.json:2`, `incremental_test.go` (274, 317, 366, 415), `slices/export_test.go:273`. Fixtures de `cmd/gira/testdata/**` e `extractors/testdata/**` não têm o campo | `[CÓDIGO INSPECIONADO]` `git grep` |
| Goldens de bytes do mapa: só `docs/scorecard.json:9,247` (`output_sha256`). `TestPublicScorecardReproducible` (`scorecard_test.go:106`) falharia até regenerar com `UPDATE_SNAPSHOTS=1` | `[CÓDIGO INSPECIONADO]` |

### Evidência: consumidores e produtores

| Consumidor / produtor | file:line | Comportamento hoje | Impacto de proveniência embutida + bump |
|---|---|---|---|
| `store.LoadFile` (único leitor) | `store/store.go:207-225` | Aceita qualquer `schema_version` não vazio; vazio → `ErrCorruptIndex` (mensagem "corrupt") | Ponto único para validar versão; a mensagem "corrupt" é errada para "schema não suportado" |
| `store.PublishFile`/`Publish`, `CheckSymbolLossGate` | `store.go:83-176`, `incremental.go:434` | Gate compara só nós e arestas | Não vê o schema: binário antigo sobrescreve 2.0.0+proveniência com 1.0.0 (E5) |
| `loadIndex` e comandos `serve mcp`, `impact`, `slice`, `explain`, `locate`, `export`, `view`, `tour`, `flows`, `doc`, `skills nudge` | `cmd/gira/main.go:300-397`; `:659…1376` | Aceitam; com forma incompatível devolvem vazio com rc=0 (E2) | Hoje nenhum imprime proveniência; N3 só é útil se `export`/`view`/`doc` e envelopes MCP a expuserem |
| `gira diff`, MCP `diff_maps` | `main.go:1233-1238`; `mcp/server.go:612-640` | Aceitam; qualquer erro de load vira `index_missing` com `retryable:true` (E3) | Precisa de código próprio para schema não suportado e corrompido |
| `skills install` / `check --index` | `main.go:1281,1342`; `skills/install.go:158,178` | Só gravam o caminho; não leem o arquivo | D9c exige leitura nova do cabeçalho |
| `skills.Nudge` | `skills/nudge.go:19` | `loadIndex` engole todo erro (rc=0, mudo) | Fail-closed silencioso |
| `Querier.isStale` | `skills/query.go:297-309` | `bytes.Contains` do HEAD no JSON cru; sem chamador de produção | SHA dentro da proveniência passaria a satisfazê-lo: ambiguidade a decidir |
| `IndexedCommits`/`HasCommit`/`InspectRepoFreshness`, `humanBanner` | `domain/staleness.go:44,72,119,205`; `main.go:52-65` | Varrem só evidência de nós e arestas (ADR 0001) | Cegos às fontes da proveniência; overrides têm `Repo` e `CommitSHA` vazios (`pipeline.go:360,422`) |
| `slices.ExportMarkdown/HTML` | `slices/export.go:22,487` | Imprimem `mm.SchemaVersion` sem validar | Bytes de export mudam com o bump |
| `eval` hermético | `eval/eval.go:380-403,428-436`; `scorecard.go:147-150` | `OutputSHA256` = hash dos bytes do mapa | Bump invalida `docs/scorecard.json` |
| `indexer.Build` | `indexer/indexer.go:34-97` | Recebe `manifest.Manifest` já parseado, sem bytes crus | D4/N3 exigem hash dos bytes crus: a API muda |
| `bench/layer0/gira.go` | `:5-33` | Subconjunto próprio; não lê `schema_version` | Cego a schema (E7) |
| `bench/phases.json:20-21`, `bench/internal/benchtest/benchtest.go:75-76`, `bench/layer1`, `tools/index-bundle/bootstrap.sh:42,49` | — | Usam `gira index --output` e `serve mcp --index` pela CLI | Seguem válidos com `--output` destacado (D1); precisam de teste. O texto do bootstrap ("saves the bundle") precisa de revisão |
| Skill `gira-arch-diff` | `skills/catalog/gira-arch-diff/SKILL.md:14-15` | `gira index --output` + `gira diff`/`diff_maps` | Fluxo continua; documentar proveniência no diff |
| README, `website/docs` (22 arquivos), ADR 0003 | `README.md:42-59` | Nenhum documenta as chaves do JSON | O schema não está documentado em lugar nenhum; docs versionadas da 0.1.0 não devem ser reescritas |
| ADR 0001 | `docs/adr/0001-…md` | Índice guarda locators e identificadores, nunca código | Hash, SHAs e versões são compatíveis; URL crua do repositório não é segura |

Todos os itens acima: `[CÓDIGO INSPECIONADO]`.

### Evidência: experimentos

`[TESTE EXECUTADO]`. Binário `go build -o /tmp/gates-GateSchema/gira ./cmd/gira` (exit 0, `gira 0.1.0-dev`). Mapa base do corpus `testdata/corpus/otel-demo-3.1.0-mini` (88 nós, 146 relações); três indexações deram o mesmo SHA-256 `ba8ade6f…7650`. Variantes geradas com `jq`: **a** `provenance` extra; **b** `schema_version: "2.0.0"`; **c** a+b; **d** `"2.0.0"` com forma incompatível (`nodes`/`relationships` movidos para `graph`); **e/f/g** controles (vazio, numérico `2`, ausente). Artefatos: `/tmp/gates-GateSchema/{exp.sh,exp.log}`.

| Exp. | Resultado |
|---|---|
| E1 (a, b, c) | `locate`, `export`, `view`, `impact`, `slice`, `explain`, `tour`, `flows`, `doc`, `diff`, `serve mcp`, `skills nudge`: rc=0 em todos. Saída byte a byte igual ao base, exceto `export md/html` em b e c (mostram `2.0.0`). `skills install/check --index`: saída idêntica em todas as variantes |
| E2 (d) | rc=0 sem erro: `locate` devolve `[]`; `export md` documento vazio de 880 B com "Schema Version: 2.0.0"; `tour` 40 B; `diff` base→d reporta toda a arquitetura removida (51 224 B); `diff_maps` idem; `nudge` mudo. `impact`, `slice`, `explain`, `flows`, `doc` falham rc=1 só por "node not found" |
| E3 (controles) | vazio/ausente: rc=1 `corrupt or invalid macro-map index … missing schema_version`; numérico: rc=1 `cannot unmarshal number into … MacroMap.schema_version of type string`; `nudge` rc=0 mudo; `diff_maps` com versão vazia devolve `{"code":"index_missing",…,"retryable":true}` |
| E4 | `nudge` com caminho citado (`src/cart/src/Dockerfile`) produz `additionalContext` em base, a, b, c; mudo em d |
| E5 | Binário antigo: `gira index manifest.yaml --output var/w_overwrite.json` sobre mapa 2.0.0+proveniência: rc=0; resultado passa de `["2.0.0","object"]` para `["1.0.0","null"]` (downgrade silencioso) |
| E6 | Descoberta legada (`~/.cache/gira/indexes/<sys>/macro-map.json`) com variante c: `locate` rc=0 |
| E7 | Adaptador `bench/layer0` compilado em `/tmp`: base/a/b/c dão 26 componentes, 11 contratos, 79 arestas; d dá 0/0/0 sem erro; e e f (schema vazio e numérico) lidos normalmente |

### Achados relevantes

1. Binários atuais aceitam em silêncio schema `2.0.0` e proveniência extra (E1). O requisito SC3 ("parser antigo não serve mapa novo como se fosse válido") **não é cumprido só subindo a string** (variante D abaixo).
2. Uma forma incompatível produz resultado vazio com rc=0, o mesmo modo de falha do S2 da v1 (E2). Renomear `nodes` é pior que não mudar.
3. Binário antigo sobrescreve mapa novo sem aviso (E5).
4. `store.LoadFile` é o único leitor Go, o que centraliza a correção. `loadIndex`, `diff` e `diff_maps` precisam de erro novo distinto de `ErrCorruptIndex`.
5. `cloneMacroMap` já perde campos hoje; `CanonicalJSON` perderia o campo novo.
6. `bench/layer0`, `skills nudge`, `skills install/check` e `isStale` exigem mudança de código, qualquer que seja a opção.
7. `url` crua não é segura para embutir: `gitcache.NormalizeURL` (`cache.go:48`) não remove credenciais e `file://` carrega caminho da máquina `[CÓDIGO INSPECIONADO]`.
8. `SupportMatrix()` tem 17 extratores, todos `"1.0.0"`, mantidos à mão; há outra constante independente `DefaultExtractorVersion` (`incremental.go:35`). Nada obriga o bump de versão quando o extrator muda `[CÓDIGO INSPECIONADO]`.

### Opções para o dono

**Valor de `schema_version`:**

| Opção | O que faz | Efeito sobre binários atuais |
|---|---|---|
| A | `"1.1.0"`, mudança aditiva | Lêem tudo, ignorando a proveniência (E1); downgrade silencioso persiste (E5). Não cumpre a letra de SC3 se o schema for considerado incompatível |
| B | `"2.0.0"` com `schema_version` deixando de ser string | Falham em `LoadFile` nomeando `MacroMap.schema_version` (E3). Foge do semver textual |
| C | `"2.0.0"` string, sem `target_system` no topo | Falham com "missing target_system": mensagem enganosa, rejeição incidental. Mantém semver |
| D | `"2.0.0"` string, sem outro cuidado | **Descartada pela evidência**: aceita em silêncio (E1 b/c) |

**Política do leitor novo:**

| Opção | O que faz |
|---|---|
| R1 | Aceitar conjunto explícito (`"1.0.0"` legado + versões novas); rejeitar o resto com código estável novo (ex. `index_schema_unsupported`), distinto de `index_corrupt` |
| R2 | R1, mas `"1.0.0"` só em `--index` explícito, com aviso (alinhado a D9/D9c); nunca com `gira.yaml` |
| R3 | Aceitar qualquer versão da mesma maior; rejeitar maior desconhecida |

**Nome e forma da proveniência:** nome `provenance` já nomeia a métrica `provenance_resolvable_pct` (`scorecard.go:37`) e textos em `slices/doc.go`, com risco de confusão com evidência por locator `[CÓDIGO INSPECIONADO]`. Alternativas: `snapshot`, `source_identity`. Esboço (ordem fixa pela struct, sem mapas):

```json
"provenance": {
  "schema_version": "<igual ao topo>",
  "manifest_sha256": "<64 hex, bytes crus do manifesto (D4)>",
  "extractors": [ {"id":"…","version":"…"} ],
  "sources": [ {"repo_id":"…","sha":"<40 hex>","path_scope":"…","url_sha256":"<gitcache.RepoHash>"} ]
}
```

Fontes de TODOS os `manifest.Repositories` (não só os com evidência); sem URL crua; sem timestamp, caminho absoluto, hostname nem `main.date`; auto-hash fora do mapa (circular). Incluir `tool_id`/hash do executável no snapshot é opção em aberto e interage com o Gate 3.

### Recomendação `[INFERÊNCIA]` (recomendação, aguarda o dono)

- Só B (ou C) falham fechado em binários já distribuídos; para os consumidores do repo qualquer opção exige código. Entre B e C, a rejeição de C é incidental e a mensagem enganosa.
- Leitor R1 com decodificador estrito (`DisallowUnknownFields`) ao menos na proveniência; validar hash 64 hex, SHAs 40 hex, listas ordenadas sem duplicatas, `provenance.schema_version == schema_version`, proveniência obrigatória a partir da versão nova; `PublishFile` recusar sobrescrever schema maior por menor.
- Não embutir URL crua; usar `url_sha256`.
- Mapas 1.0.0 válidos só para leitura explícita; sem reescrita in-place; Rollback da spec segue válido.

### Testes a definir `[TESTE DEFINIDO]`

1. Roundtrip `CanonicalJSON` → `LoadFile` → `CanonicalJSON` byte a byte, com proveniência; teste por reflexão de que todo campo exportado de `MacroMap` sobrevive a `CanonicalJSON` e `cloneMacroMap`.
2. Determinismo: mesma entrada, mesmos bytes; sem timestamp nem caminho absoluto na proveniência.
3. Leitor novo × variantes: aceitar 1.0.0 conforme R1/R2; rejeitar schema desconhecido, hash e SHA malformados, proveniência ausente em versão nova, `provenance.schema_version` divergente, cada um com código estável.
4. Matriz de comandos de `exp.sh` com binário antigo contra mapas novos (B/C: rc≠0 em todo comando que lê o mapa; A: aceitação documentada).
5. Downgrade: binário antigo sobre mapa novo conforme a decisão.
6. Fontes: manifesto com 2 repositórios, um sem evidência e um com `path_scope` vazio, ambos em `sources`; override só via `manifest_sha256`; credenciais na URL não vazam.
7. `diff` e `diff_maps` com schemas diferentes: aviso ou erro com código estável.
8. `isStale`/staleness: SHAs de proveniência contam ou não, conforme decisão.
9. Regenerar `docs/scorecard.json` e goldens de `export`.
10. `bench/layer0` falha em schema incompatível; fixture atualizado ou mantido como legado explícito.
11. `skills install/check --index` com caminho legado, relativo e snapshot destacado (D9c).
12. Cópia byte a byte de `init --output P` (D1) igual à canônica, com proveniência.

### Residual `[DESCONHECIDO]`

- Comportamento no Windows e macOS das leituras acima.
- Consumidores externos ao repo (hosts MCP configurados, scripts de usuário) que leem `macro-map.json` direto.
- Se `skills nudge` mudo diante de schema inválido é aceitável (hoje é por desenho: "never fails the hook").
- Se `tool_id` entra no objeto embutido ou só em `meta.json`.
- Se os 17 extratores com versão manual bastam como "versões dos extratores" de N3.

**Status do gate: direção escolhida (B/R2/`source_identity`); implementação e auditoria de consumidores/schema ainda pendentes.**

---

## Gate 2: manifesto fora da raiz Git (D3)

**Pergunta.** Spec `## Open questions` item 2; decisão §Bloqueios item 2 e linha D3 (`[DECISÃO APROVADA]`: `Locator.Path` relativo à raiz Git, normalizado com `/`; política para manifesto fora da raiz Git permanece bloqueadora; nunca absoluto como fallback silencioso). Auditar ADR 0001 e consumidores de locators.

### Evidência: fluxo do caminho até `Locator.Path`

| Etapa | Evidência |
|---|---|
| Descoberta (`findProjectManifest`) parte de `os.Getwd()` e sobe: sempre **absoluto**, sem resolver symlinks | `[CÓDIGO INSPECIONADO]` `cmd/gira/main.go:278-297` |
| Posicional e `--manifest` ficam **como digitados** | `[CÓDIGO INSPECIONADO]` `main.go:423-431`, `:439`, `:448`, `:458` |
| `ApplyOverrides(mm, m.Overrides, opts.ManifestPath)` | `[CÓDIGO INSPECIONADO]` `indexer/indexer.go:90`, `indexer/incremental.go:683` |
| Locator: `"gira.yaml"` se vazio; `Locator{Path: manifestPath, …}` com `Repo` e `CommitSHA` vazios | `[CÓDIGO INSPECIONADO]` `extractors/pipeline.go:337-339,360,422-426` |
| `init` usa `absDir/gira.yaml` e `gitToplevel(absDir)`; escreve `url: file://<toplevel>` | `[CÓDIGO INSPECIONADO]` `main.go:524-541,569-575` |
| `eval.go:399` chama `indexer.Build` sem `ManifestPath` (fallback `"gira.yaml"`) | `[CÓDIGO INSPECIONADO]` |
| Helpers Git (`gitToplevel` `:569`, `isGitRepo` `:87`, `gitHead` `:66`, `:1381`) não sanitizam o ambiente (`GIT_DIR` etc.) | `[CÓDIGO INSPECIONADO]` |

### Evidência: matriz de execução

`[TESTE EXECUTADO]`. Binário `/tmp/gates-GateD3/gira` (sha256 `552c2d3b…`, linux/amd64). Cada célula: `gira index [args] --output out/N.json --cache-dir …`, fonte `file://` com `Dockerfile`, manifesto com overrides sobre `ghost-a`/`ghost-b` (viram `unresolved_refs`). Todas rc=0; duas execuções completas deram hashes idênticos. Dados brutos: `/tmp/gates-GateD3/matrix2.tsv`. `ABS` = `/tmp/gates-GateD3/work`; `P` repo git com `gira.yaml` na raiz e em `conf/sub/`; `NG` sem git; `W` worktree de P; `LNK_DIR`/`LNK_FILE` symlinks; `S/M` submódulo.

| # | Célula | `Locator.Path` emitido | sha256 (12) |
|---|---|---|---|
| 1 | raiz, relativo (cwd P) | `gira.yaml` | 1a8c07538fcc |
| 2 | raiz, absoluto | `ABS/P/gira.yaml` | 2e390a7574dc |
| 3 | raiz, descoberta (P/sub/deep) | `ABS/P/gira.yaml` | 2e390a7574dc |
| 4 | subpasta, relativo da raiz | `conf/sub/gira.yaml` | 9d4b5abcd0e9 |
| 5 | subpasta, relativo da própria pasta | `gira.yaml` | 1a8c07538fcc |
| 6 | subpasta, relativo com `..` | `../../conf/sub/gira.yaml` | abb6fc247dc7 |
| 7 | subpasta, absoluto | `ABS/P/conf/sub/gira.yaml` | 0a0c12b1ef19 |
| 8 | subpasta, descoberta | `ABS/P/conf/sub/gira.yaml` | 0a0c12b1ef19 |
| 9 | fora do git, relativo | `gira.yaml` | 1a8c07538fcc |
| 10 | fora do git, absoluto | `ABS/NG/gira.yaml` | 3aa286329957 |
| 11 | fora do git, descoberta | `ABS/NG/gira.yaml` | 3aa286329957 |
| 12 | worktree, relativo | `gira.yaml` | 1a8c07538fcc |
| 13 | worktree, absoluto | `ABS/W/gira.yaml` | 7280f626b83d |
| 14 | worktree, descoberta | `ABS/W/gira.yaml` | 7280f626b83d |
| 15 | worktree, subpasta relativo | `conf/sub/gira.yaml` | 9d4b5abcd0e9 |
| 16 | symlink de pasta, absoluto | `ABS/LNK_DIR/gira.yaml` | 3a7f6df03211 |
| 17 | symlink de pasta, relativo | `gira.yaml` | 1a8c07538fcc |
| 18 | symlink de pasta, descoberta | `ABS/LNK_DIR/gira.yaml` (mantém o symlink) | 3a7f6df03211 |
| 19 | symlink de arquivo, relativo | `gira.yaml` | 1a8c07538fcc |
| 20 | symlink de arquivo, absoluto | `ABS/LNK_FILE/gira.yaml` | 443d45fbe025 |
| 21 | symlink de arquivo, descoberta | `ABS/LNK_FILE/gira.yaml` | 443d45fbe025 |
| 22 | submódulo, relativo do pai | `M/gira.yaml` | b0aceb17be33 |
| 23 | submódulo, relativo interno | `gira.yaml` | 1a8c07538fcc |
| 24 | submódulo, descoberta | `ABS/S/M/gira.yaml` | e3e5d799b553 |

Protótipo do Git root `[TESTE EXECUTADO]` (shell em `/tmp`, não é código do produto): `readlink -f` do arquivo; `git -C <dir real> rev-parse --show-toplevel --show-prefix` com `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE` removidos; caminho = prefixo + basename. Células 1–8, 12–24 convergem para `gira.yaml` ou `conf/sub/gira.yaml` (relativo, absoluto, descoberta e worktree iguais; symlink de arquivo resolvido ao alvo); células 9–11 (fora do git) **bloqueadas** (rc=128 `not a git repository`).

Comportamento do Git `[TESTE EXECUTADO]`:

| Situação | Resultado |
|---|---|
| Fora de qualquer repo | `rev-parse` rc=128 |
| Repo bare / dentro de `.git/` | `--show-toplevel` rc=128; `--show-prefix` vazio com rc=0 (não detecta sozinho) |
| Submódulo, worktree, `--separate-git-dir` | toplevel correto (`.git` arquivo `gitdir:`) |
| Repo aninhado | o interno vence |
| `GIT_TEST_ASSUME_DIFFERENT_OWNER=1` | rc=128 (`dubious ownership`), indistinguível de "não é repo" pelo rc |
| `GIT_DIR=P/.git` no ambiente, manifesto em NG | toplevel = `NG`, **errado, rc=0** |
| `GIT_CEILING_DIRECTORIES` | corta a busca para cima |
| Symlink de pasta / de arquivo | pasta: toplevel e prefixo do alvo real; arquivo: `git -C` na pasta do link falha, alvo real em P |

`gira init` com `GIT_DIR=<P>/.git` apontando para outro repo, em pasta sem git `[TESTE EXECUTADO]`: gravou `gira.yaml` na pasta sem git, fixando o HEAD de P (`ea828d08…`); o erro só veio depois, ao clonar. Sem `GIT_DIR`, falha corretamente com `not a git repository`. `e2e_test.go:19-25` já documenta o risco com hooks `[CÓDIGO INSPECIONADO]`.

### Evidência: ADR 0001 e consumidores

| Item | Evidência |
|---|---|
| ADR 0001 (`docs/adr/0001-…md:3`): o índice guarda locators, sem corpos de código; excertos são renderizados na consulta. Caminho relativo à raiz é compatível e mais fiel que o absoluto | `[CÓDIGO INSPECIONADO]` |
| `gira explain --node service:app --repo <P>`: índice gerado com caminho relativo mostra o excerto de `gira.yaml`; com absoluto mostra `source unavailable` (`slices/source.go:35` rejeita absoluto) | `[TESTE EXECUTADO]` |
| ADR 0008:34 só diz "locator pointing to the line in `gira.yaml`"; o locator de override já viola `repo@commit`. ADR 0009:8: manifesto "na raiz do repositório ou via `--manifest`", sem dizer o que fazer fora de repo | `[CÓDIGO INSPECIONADO]` |
| `main.go:77-125` (`excerptFor`, `explain --repo`) | melhora: excerto só funciona com relativo, resolvendo sob `--repo` (deve ser a raiz do repo do manifesto) `[CÓDIGO INSPECIONADO]` |
| `slices/primpact.go:48-52,83-86,192-195` (`--from-diff`) | sem `--repo-id`, editar `gira.yaml` passa a casar com a evidência de override; com `--repo-id` continua excluído (`Repo==""`) |
| `skills/nudge.go:28-45` | passa a achar o `gira.yaml` do override |
| `archdiff/archdiff.go:49,345-393` | com absoluto, dois checkouts dão `evidence_moved` espúrio; com relativo cessa; o texto do campo `locator` do diff muda |
| `domain/staleness.go:233-236` | `Repo==""` → `StatusUnverified`; sem efeito |
| `extractors/support_matrix.go:323-355`, `belongsToComponent` (`:427`) | `conf/sub/gira.yaml` pode casar prefixo `svcName/` (ex. manifesto em `payments/gira.yaml` com serviço `payments`); `[INFERÊNCIA]` caso de borda não testado |
| `indexer/incremental.go:442-454,473` (`CheckSymbolLossGate`), `extractors/resolve.go:264-267`, `eval/safety.go:55`, `eval/scorecard.go:221`, MCP `skills/query.go`, `bench/layer0` | sem efeito observado |
| `slices/{render,doc,export,tour,engine,intra_service}.go` | mudam as strings de locator só para evidência de override |
| Quebra se a spec mudar: `extractors/pipeline_test.go:326-389,428-453` (usa `custom-gira.yaml` e fallback `gira.yaml`); `cmd/gira/main_test.go:418-449,522-560` indexa manifesto em `t.TempDir()` fora de qualquer repo git (precisaria de `git init`); bytes de qualquer mapa com override mudam (já previsto na spec, linha 112) | `[CÓDIGO INSPECIONADO]` |

### Achados relevantes

1. O path depende de **como** o manifesto é invocado: a mesma pasta dá três valores (relativo, `../..`, absoluto).
2. A descoberta sempre emite absoluto: o mesmo projeto em duas máquinas gera bytes diferentes.
3. Manifestos **diferentes** colidem no mesmo path e hash (`gira.yaml`, `1a8c07538fcc`: células 1, 5, 9, 12, 17, 19, 23).
4. Hoje não existe política de raiz Git: dentro e fora do git o tratamento é idêntico.
5. `GIT_DIR` no ambiente sequestra `gira init` (e qualquer helper Git sem saneamento).
6. Caminho absoluto quebra o excerto de `gira explain`.
7. `--show-prefix` sozinho não detecta bare nem `.git/`; é preciso `--show-toplevel` e checar o rc. O rc=128 também cobre `dubious ownership`, o que impede distinguir "fora do repo" de "git quebrado" só pelo rc.

### Opções para o dono (política de manifesto fora da raiz Git)

| # | Opção | Prós | Contras | Determinismo |
|---|---|---|---|---|
| O1 | Falha fechada com erro estável | Nunca grava path dependente da máquina; simples | Quebra manifesto em pasta sem git quando há override (ex. `t.TempDir()`) | Total |
| O2 | Relativo à pasta do manifesto (basename) | Não falha | Mesmo `gira.yaml` significa "raiz Git" dentro e "pasta do manifesto" fora; colide entre manifestos | Determinístico, ambíguo |
| O3 | Campo `root:` no `gira.yaml` (ou `--root`) | Cobre sem-git | Novo campo de schema; altera `source_key` (D4) e ADR 0009; pode mentir; exige validação | Determinístico se estrito |
| O4 | Marcador opaco (ex. `<outside-git>/gira.yaml`) | Não falha, não vaza | Path que nunca resolve; formato inventado | Determinístico, ambíguo |
| O5 | `Path=""` para override | Sem ambiguidade | Perde a linha citada; contradiz ADR 0008:34 | Determinístico, perde informação |

Subdecisão: `ManifestPath` vazio (`eval/eval.go:399`): manter o fallback `"gira.yaml"`, exigir locator explícito do chamador, ou o `eval` não aplicar override.

### Recomendação `[INFERÊNCIA]` (recomendação, aguarda o dono)

- **O1**, só quando um locator de override seria emitido (manifesto com `overrides`); sem overrides nada é gravado e nada falha. Erro estável `manifest_outside_git_root` quando o Git responde "não é work tree"; segundo código `git_root_unavailable` para git ausente, `dubious ownership` ou erro de exec, também fechado.
- Resolver o arquivo com `EvalSymlinks`; `git -C <dir real> rev-parse --show-toplevel --show-prefix`, exigir rc=0 e montar `prefixo + basename` já com `/`, evitando `filepath.Rel`.
- Sanear o ambiente Git do helper (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`, `GIT_NAMESPACE`, `GIT_PREFIX`; decidir sobre `GIT_CEILING_DIRECTORIES`), reaproveitando `e2e_test.go:19-45`, também em `gitToplevel`/`gitHead`/`isGitRepo`.
- `eval`: sem fallback silencioso `"gira.yaml"`.
- O3 pode vir depois sem quebrar O1.

### Testes a definir `[TESTE DEFINIDO]`

1. Matriz (raiz, subpasta, worktree, submódulo, symlink) × (relativo, absoluto, descoberta, `..`): todas convergem; hashes iguais entre relativo, absoluto e descoberta.
2. Mesmo manifesto em dois clones e em worktree: bytes idênticos.
3. Manifesto em pasta sem git com overrides: `manifest_outside_git_root`, nenhum arquivo de saída ou versão criado; sem overrides indexa normalmente.
4. Repo bare e manifesto dentro de `.git/`: `manifest_outside_git_root`.
5. `GIT_DIR`/`GIT_WORK_TREE` exportados: resultado igual ao sem ambiente; sem-git ainda falha; regressão de `gira init` com `GIT_DIR` apontando para outro repo.
6. `GIT_TEST_ASSUME_DIFFERENT_OWNER=1`: `git_root_unavailable`.
7. Symlink de arquivo e de pasta: path do alvo real; nada vaza para fora da raiz.
8. Nenhum `Locator.Path` de override absoluto, com `\`, vazio ou começando por `..`.
9. `explain --repo <raiz>` mostra o excerto; outra pasta mostra `source unavailable`.
10. `--from-diff` e `nudge`: edição em `gira.yaml` casa sem `--repo-id`; com `--repo-id` continua excluída.
11. `diff` entre índices de dois clones sem `evidence_moved` espúrio.
12. Ajustar `main_test.go:418` (fixture com `git init`) e `pipeline_test.go:326-453`.
13. Fixar o novo hash por fixture com override.

### Residual `[DESCONHECIDO]`

- Separadores, nomes 8.3 e caixa no Windows (Gate 4).
- Caso de borda `belongsToComponent` não testado.
- Quem identifica o repo do manifesto, já que o locator de override tem `Repo==""` (a identidade fica no hash do manifesto, N3).

**Status do gate: política O1 e política do eval escolhidas; implementação de normalização/saneamento e testes de caminhos ainda pendentes.**

---

## Gate 3: `main.date` em `tool_id` (D5)

**Pergunta.** Spec `## Open questions` item 3; decisão §Bloqueios item 3 e linha D5 (`[DECISÃO APROVADA]`: divergência de `tool_id` falha fechado com `index_tool_mismatch`, sem flag; "se for apenas informativo, excluir de `tool_id`; se alterar semântica, manter com justificativa").

### Evidência: rastreio no código

| Fato | Evidência |
|---|---|
| `version="0.1.0-dev"`, `commit="none"`, `date="unknown"` | `[CÓDIGO INSPECIONADO]` `cmd/gira/main.go:128-130` |
| `date` só é lido em `main.go:152` (subcomando `version`); grep de `\bdate\b` em `cmd/` e `internal/` não achou outro uso; `commit` também só em `:152` | `[CÓDIGO INSPECIONADO]` |
| `version` alimenta `printUsage` (`:255`), `mcp.NewServer` (`:665`) e `skills.BuildManifest` (`:1312`, `skills/manifest.go:54`) | `[CÓDIGO INSPECIONADO]` |
| `date` e `commit` não entram em `indexer`, `extractors`, `domain.MacroMap`, scorecards nem em `DiscoveredAt`; não há `runtime/debug.ReadBuildInfo` nem `vcs.*` no código Go | `[CÓDIGO INSPECIONADO]` |
| `.goreleaser.yaml:36-38`: `-trimpath` e `-s -w -X main.version -X main.commit -X main.date={{.Date}}`; `mise.toml:16` não injeta ldflags (`date` = `"unknown"` em dev) | `[CÓDIGO INSPECIONADO]` |
| `DiscoveredAt`: extractors usam `time.Now()` (`pipeline.go:341`), mas `indexer.StampEvidence` (`indexer.go:96-115`) sobrescreve com a data do commit; decisão SC4 comentada em `domain/models.go:83-86` | `[CÓDIGO INSPECIONADO]` |
| Algoritmo de `tool_id` da v1 (`gira-spec-v1-proposta-original.md:63`): `main.version`, `main.commit`, `GoVersion`, `Main`, todas as `Settings` (incluindo `-ldflags`), `Deps`; sha256 do executável só se `vcs.modified`, sem `vcs.revision` ou `commit=="none"`. `main.date` não é campo próprio, entra via `Settings["-ldflags"]` e pelo hash do executável | `[CÓDIGO INSPECIONADO]` |

### Evidência: experimento

`[TESTE EXECUTADO]`, `25882da`, go1.24.0 linux/amd64. Fixture `testdata/corpus/otel-demo-3.1.0-mini/src` copiada para repo git em `/tmp` (HEAD `d63d843`, 69 nós, 68 relações). `gira index <gira.yaml> --output …` por binário, 2 repetições com `sleep 1`. Binários com `-buildvcs=false`; `version=1.2.3`, `commit=abc123` fixos nos dois com data; `none` sem ldflags. Artefatos: `/tmp/gates-maindate/`.

| Binário | exe sha256 (16) | mapa | `export` md | `export` html | `locate checkout` | `skills manifest` |
|---|---|---|---|---|---|---|
| `none` (sem ldflags) | `36083720…` | `af467922…` | `06f24bb2…` | `393cdee4…` | `e456679f…` | `0bd15169…` |
| `d2001` (`date=2001-01-01T00:00:00Z`) | `6254fdd1…` | `af467922…` | `06f24bb2…` | `393cdee4…` | `e456679f…` | `0567b0ae…` |
| `d2030` (`date=2030-12-31T23:59:59Z`) | `31d7a1c1…` | `af467922…` | `06f24bb2…` | `393cdee4…` | `e456679f…` | `0567b0ae…` |

- Mapa, `export` e `locate` idênticos nos 3 binários e nas 2 repetições. `skills manifest` difere só entre `none` e os demais, por `version` (`d2001` = `d2030`).
- `gira version` mostra a data: `gira 1.2.3 (commit: abc123, date: 2001-01-01T00:00:00Z)`.
- O executável muda com qualquer diferença de ldflags, mesmo só de data: `-trimpath -s -w` com `-X main.date=A` e `=B` deram `0e89e13f…` e `f038b99b…`.
- **`BuildInfo.Settings`:** sem `-trimpath`, `-ldflags` aparece em `Settings` (visto com `go version -m`). **Com `-trimpath`, a setting `-ldflags` não é registrada** (testado com `-trimpath -ldflags "-X main.date=A"` e `-trimpath -ldflags "-s -w -X main.date=A"`). O goreleaser usa `-trimpath`.
- `vcs.*`: nenhum build local gerou `vcs.revision`, `vcs.time` ou `vcs.modified`, nem com `-buildvcs=true` em worktree detached em `25882da`; causa não investigada `[DESCONHECIDO]`.

### Achados relevantes

1. `main.date` é puramente informativo no código de extração `[CÓDIGO INSPECIONADO]`; datas extremas não alteram os bytes do mapa nem `export` e `locate` `[TESTE EXECUTADO]`. Pelo critério de D5, a evidência sustenta **excluir**.
2. A afirmação do plano de governança (`2026-10-17-plano-execucao-governanca.md:357`) de que `-ldflags` aparece em `BuildInfo.Settings` e que `main.date` muda o `tool_id` entre builds de release é **falsa para builds com `-trimpath`** (goreleaser). Correção registrada como nota datada naquele plano.
3. Valores injetados por `-X` só entram no `tool_id` se `main.version` e `main.commit` forem lidos diretamente das variáveis, como a proposta faz.
4. O hash do executável continua mudando com a data. No dev build (`commit=="none"` ou `vcs.modified`) um rebuild reindexa em qualquer opção; dois builds de release da mesma tag (`{{.Date}}` é hora de build) geram executáveis diferentes.

### Opções para o dono (subdecisão D5)

| Opção | O que faz | Consequências |
|---|---|---|
| A | Excluir `main.date` do `tool_id`: normalizar `Settings["-ldflags"]` removendo `-X main.date=…` | Estabiliza `tool_id` de release. Exige parsear a string `-ldflags` (só aparece sem `-trimpath`). Não resolve o reindex de dev |
| B | Manter `-ldflags` bruto no `tool_id` (proposta v2) | Builds do mesmo commit em horários diferentes mudam `tool_id` sem mudança semântica; com `-trimpath` o efeito some, sem `-trimpath` persiste |
| C | Excluir a setting `-ldflags` inteira, mantendo `version` e `commit` como campos próprios | Sem parser. Perde distinção de `-X` futuros; aceitável hoje, pois só `version`, `commit` e `date` são injetados (`.goreleaser.yaml:38`) |

Pergunta derivada: o hash do executável continua valendo para dev builds (`vcs.modified`/`commit=="none"`)?

### Recomendação `[INFERÊNCIA]` (recomendação, aguarda o dono)

Opção C (ou A): excluir `-ldflags` e `main.date` do `tool_id`, mantendo `main.version`, `main.commit`, `GoVersion` e `Deps`. A decisão sobre manter o hash do executável em dev builds fica separada e também é do dono.

### Testes a definir `[TESTE DEFINIDO]`

1. Dois binários que diferem só em `-X main.date`: `tool_id` igual (conforme a opção escolhida) e mapa idêntico.
2. Binário com `version` ou `commit` diferente: `tool_id` diferente.
3. Teste com e sem `-trimpath`, para a presença de `-ldflags` em `Settings` (go1.24 e a versão do CI).
4. Dev build (`commit=="none"`): comportamento do hash do executável conforme decisão.
5. Reexecutar este experimento (3 binários) como regressão, com `vcs.*` investigado.

### Residual `[DESCONHECIDO]`

- `vcs.revision`, `vcs.time`, `vcs.modified` não foram observados em nenhum build local.
- Comportamento em outras versões do Go e no runner do release.

**Status do gate: opção C escolhida; testes de regressão de `tool_id`/hash de dev e confirmação no runner de release ainda pendentes.**

---

## Gate 4: Windows

**Pergunta.** Spec `## Open questions` item 4; decisão §Bloqueios item 4: experiências reproduzíveis no runner Windows para symlinks e `os.DevNull` como `GIT_CONFIG_GLOBAL`; até lá, comportamento `[DESCONHECIDO]`.

### Evidência: execuções

Branch de experimento `exp/2026-10-17-windows-gates` @ `b43faed97a076bfa7bab180fad6418ae7e254c02` (não vai para `main`; só citada). Módulo próprio em `experiments/windows-gates/` (com `golang.org/x/sys v0.30.0` só para `LockFileEx`) e workflow `.github/workflows/windows-gates.yml`; Go `1.24.0`, igual a `mise.toml` e ao `go.mod` da raiz.

| Run | Sha | Observação | Resultado |
|---|---|---|---|
| [39079306847](https://github.com/EduardoShibukawa/gira-g/actions/runs/39079306847) | d770a42 | harness v1 | `[CI]` 3 jobs success |
| [39079448819](https://github.com/EduardoShibukawa/gira-g/actions/runs/39079448819) | b340601 | revelou que `update-index` ignora caminho com tab no Windows e que `whoami` caía no do MSYS | `[CI]` 3 jobs success |
| [39079538834](https://github.com/EduardoShibukawa/gira-g/actions/runs/39079538834) | b43faed | **final**; fixtures por `mktree`, `whoami.exe` do System32 | `[CI]` `gates (windows-latest)`, `gates (ubuntu-latest)`, `gates (macos-latest)` success; 6 testes PASS nos 3 SOs |

`[TESTE EXECUTADO]` no Linux local (go1.24.0, git 2.55.0): 6 testes passaram, exit 0, com `HOME`, `USERPROFILE` e `XDG_CACHE_HOME` isolados. Artefatos de CI: `windows-gates-{Windows,Linux,macOS}` (`.log` e `.jsonl`).

Contexto dos runners `[CI]`:

| SO | Detalhes |
|---|---|
| Windows | `windows-latest`, imagem `win25-vs2026` 20261007.264.1, Windows 10.0.26100.33438, Git for Windows 2.56.0.windows.2, usuário `runneradmin`. **Administrador elevado** (`High Mandatory Level`), **Developer Mode ligado** (`AllowDevelopmentWithoutDevLicense = 0x1`), `SeCreateSymbolicLinkPrivilege` Disabled no token (mesmo com `os.Symlink` funcionando), `core.symlinks=true` vindo de `C:/Program Files/Git/etc/gitconfig`; sem esse config, `core.symlinks` fica sem valor (exit 1). `os.TempDir()` devolve nome curto 8.3 (`C:\Users\RUNNER~1\...`) |
| Linux | ubuntu24 20261012.334.1, git 2.55.0, uid 1001 |
| macOS | macOS 26.6.2, arm64, git 2.56.0 (Homebrew) |

**Atenção:** o runner Windows é admin elevado, com Developer Mode e `core.symlinks=true` de sistema. Os resultados de symlink valem para esse perfil. Usuário final padrão: `[DESCONHECIDO]`.

### Evidência: symlinks

`[CI]`, resultado idêntico nos 3 SOs para arquivo e diretório, alvo relativo ou absoluto: `os.Symlink` ok; `Readlink` devolve o alvo literal; `EvalSymlinks` resolve; `Lstat` marca symlink; `Stat` segue o link; `SameFile(Stat(link), alvo)` true; `SameFile(Lstat, Stat)` false; abrir e ler pelo link ok; link pendente: `Stat` falha, `Lstat` ok; cadeia link→link resolve.

Checkout git de symlink commitado (clone local) `[CI]`: `default` e `core.symlinks=true` criam symlink real (arquivo e diretório) nos 3 SOs; `core.symlinks=false` cria arquivo regular com o texto do alvo; `git status --porcelain` vazio no Windows. Leitura em cache bare sem checkout: `ls-tree -r -z` lista `120000 blob …` e `cat-file -p HEAD:link_to_file` devolve `target.txt` nos 3 SOs.

SC7, grafias do mesmo manifesto: segunda aquisição de lock exclusivo não bloqueante (`flock`/`LockFileEx`). Células: SameFile / texto canônico igual / exclusão. `[CI]`

| Grafia | Windows | Linux | macOS |
|---|---|---|---|
| relativa `../proj/./gira.yaml` | T/T/T | T/T/T | T/T/T |
| symlink de arquivo (rel. e abs.) | T/T/T | T/T/T | T/T/T |
| symlink de diretório (rel. e abs.) | T/T/T | T/T/T | T/T/T |
| hardlink | T/**F**/T | T/**F**/T | T/**F**/T |
| temp dir como o SO devolveu (8.3; `/var`) | T/T/T | n/a | T/T/T |
| barras `/` | T/T/T | n/a | n/a |
| prefixo `\\?\` | T/**F**/T | n/a | n/a |
| junção (`mklink /J`) | T/**F**/T | n/a | n/a |
| variante de caixa (`GIRA.YAML`) | T/T/T | arquivo inexistente | T/**F**/T |

`SameFile` e o lock excluem em toda grafia que existe.

### Evidência: `os.DevNull` como `GIT_CONFIG_GLOBAL`

`os.DevNull` é `NUL` no Windows e `/dev/null` nos outros. Com `GIT_CONFIG_NOSYSTEM=1`, home isolado com `.gitconfig` de usuário (`core.quotepath=false`, `x.y=from-user`). Fixture: bare clone com `proto/café.proto`, `proto/a<TAB>b.proto`, `plain.proto`. Todos `[CI]`, iguais nos 3 SOs salvo indicação:

| Verificação | Resultado |
|---|---|
| `config --global --list` | exit 0, saída vazia |
| `config --global --get core.quotepath` | exit 1, saída vazia |
| Config de usuário visível? | Não, em `os.DevNull`+`NOSYSTEM`, no literal `/dev/null`+`GIT_CONFIG_SYSTEM=/dev/null`, em `os.DevNull` sem NOSYSTEM e com arquivo regular vazio. Controle sem `GIT_CONFIG_GLOBAL`: visível |
| Windows com `/dev/null` literal | leitura funciona, como o git do runner o interpreta |
| `ls-tree -z HEAD`, `ls-tree -r -z --full-name HEAD` (com pathspec), `cat-file -p <oid>` | exit 0, stderr vazio, inclusive nome não ASCII e nome com tab |
| `cat-file -p HEAD:proto/a<TAB>b.proto` e `HEAD:proto/café.proto` | exit 0 |
| Bytes de `ls_tree_z` | mesmo hash nos 3 SOs e em todas as variantes de ambiente |
| Sem `-z`, `core.quotePath` padrão/true (Windows) | `"proto/caf\303\251.proto"`; com `quotePath=false`: `proto/café.proto`; tab: `"proto/a\tb.proto"` nos dois → `-z` obrigatório |
| Escrita `git config --global x.y z` em `os.DevNull` | exit 255, nada alterado. Windows: `could not lock config file NUL: Invalid argument`; Linux/macOS: `… /dev/null: Permission denied`; Windows com `/dev/null` literal: `No such file or directory` |
| Windows, escrita em arquivo regular vazio | exit 0 (só o dispositivo nulo protege de escrita) |
| `git clone --bare` com config de usuário, sem `GIT_CONFIG_GLOBAL` | exit 0 |

Nome com tab no Windows `[CI]`: `git update-index --index-info -z` com tab retorna exit 0 mas imprime `Ignoring path a\tb.proto`; índice vazio e `commit` falha exit 1. `git mktree -z` + `commit-tree` cria o objeto; `fsck --strict` exit 0; `clone --bare` e `ls-tree -r -z` o veem. Fixtures devem usar `mktree`.

`[CÓDIGO INSPECIONADO]`: `cache.go:227-232` (a `25882da`) extrai por `git archive | tar -x` (relevante para tab no Windows, não testado); `internal/corpus/corpus.go:26-27` fixa `GIT_CONFIG_GLOBAL=/dev/null` e `GIT_CONFIG_SYSTEM=/dev/null` literais, o que funciona na leitura no Windows.

### Achados relevantes

1. Comportamento de symlink idêntico nos 3 SOs `[CI]`, no perfil admin/Developer Mode.
2. `SameFile` e lock excluem em todas as grafias; o texto canônico (`Abs`+`EvalSymlinks`) difere em hardlink, `\\?\`, junção e variante de caixa no macOS `[CI]`. `[INFERÊNCIA]` o desenho do lock deve usar identidade de arquivo (lock no próprio arquivo ou sidecar por identidade), não o texto do caminho.
3. `os.DevNull` + `GIT_CONFIG_NOSYSTEM=1` isola a leitura nos 3 SOs; escrita falha (bom isolamento); `-z` é obrigatório porque `quotePath` altera a saída sem `-z`.
4. Tab em nome não entra no índice no Windows; fixtures por `mktree -z`.

### Opções para o dono

| Opção | O que faz | Consequência |
|---|---|---|
| W1 | Aceitar a evidência do perfil runner (admin + Developer Mode) e documentar o limite | Gate fecha; usuário não elevado segue `[DESCONHECIDO]` |
| W2 | Exigir um experimento extra sem Developer Mode e com processo não elevado antes de aprovar a spec | Mais um ciclo de CI; reduz o `[DESCONHECIDO]` principal |
| W3 | Exigir `git archive \| tar -x` com tab no Windows antes da aprovação | Cobre o último ponto aberto de extração |

### Recomendação `[INFERÊNCIA]` (recomendação, aguarda o dono)

W1 para symlinks e `os.DevNull`, com W2 como teste de aceitação do SC7 no Windows (pular e registrar o motivo quando a criação de symlink falhar). Lock por identidade de arquivo. Leitura Git sempre com `-z`.

### Testes a definir `[TESTE DEFINIDO]`

SC2 (leituras Git): ler com `ls-tree -r -z --full-name` e parsing por NUL; fixture `café` e tab via `mktree -z` + `commit-tree`; `GIT_CONFIG_GLOBAL=os.DevNull` + `GIT_CONFIG_NOSYSTEM=1` com `HOME`/`USERPROFILE` apontando para `.gitconfig` diferente (config de usuário invisível; `sha256` do mapa igual com `core.quotePath` true e false); `clone`/`fetch` continuam lendo a config do usuário (ADR 0009, ex. `url.<x>.insteadOf`); escrita em `os.DevNull` falha (exit 255).

SC7 (travas), nos 3 SOs: grafias parametrizadas (relativa, absoluta, symlink de arquivo e de diretório, hardlink, variante de caixa; no Windows também `\\?\`, junção, 8.3), assertando exclusão e não igualdade de texto; se a chave for derivada do caminho, assertar igualdade de chave entre grafias; pular com motivo quando symlink não puder ser criado; processo filho morto libera o lock, timeout `index_locked`, erro injetado `lock_unavailable`, ordem M→S→P (nada disso exercitado aqui, só `LockFileEx` básico em processo único).

### Residual `[DESCONHECIDO]`

- Usuário final não admin, sem Developer Mode, token filtrado, `core.symlinks=false` de sistema; não se isolou se admin ou Developer Mode habilita o symlink.
- NTFS vs ReFS, SMB e OneDrive (só o NTFS do runner foi exercitado).
- Lock sobre o manifesto ou sobre sidecar.
- `git archive | tar -x` com tab no Windows.
- Outras versões ou instalações do Git for Windows.
- Escrita em `NUL`/`/dev/null` como administrador no Linux/macOS.

**Status do gate: W1 aceito para o perfil do runner hospedado; não se generaliza para usuário comum. O teste `git archive | tar -x` com tab e a célula SC7 em usuário não elevado/sem Developer Mode permanecem pendentes, com skip explícito se symlink não puder ser criado.**

---

## Gate 5: durabilidade e publicação por plataforma

**Pergunta.** Spec `## Open questions` item 5; decisão §Bloqueios item 5: falhas antes e depois do rename, fsync de diretório no Unix, semântica `visible_unconfirmed`, publicação parcial de `init --output`, preservação de artefatos válidos e corrompidos, validação de hash nos três sistemas.

**Evidência nesta rodada: nenhuma.** O gate não foi coberto. Não há `[TESTE EXECUTADO]` nem `[CI]` para ele `[DESCONHECIDO]`. Os comportamentos pedidos pertencem ao store versionado que ainda não existe; testes por plataforma só existem depois da implementação, que não está autorizada. Cada comportamento acima já consta como requisito da spec, mas nenhum teste específico foi definido neste documento além dos requisitos da própria spec.

**O que o gate precisa:** (a) implementação do store versionado e de `init --output` em ramo de trabalho, com testes de injeção de falha antes e depois do rename; (b) execução dessas suítes nos runners Linux, macOS e Windows `[CI]`; (c) para o Windows, o resultado do Gate 4 informa o desenho do sync de diretório (hoje nunca confirmado, spec Problem).

**Status do gate: sem evidência antes da implementação; decidido que os testes serão parte do primeiro child heavy autorizado e bloqueiam o aceite/merge dessa implementação.**

---

## Decisões do dono registradas em 2026-10-17

Estas escolhas encerram as alternativas de produto nesta rodada. Elas **não** aprovam a spec e **não** autorizam o tier heavy. Quando a decisão depende de mudança de schema ou contrato, o desenho e os testes abaixo são obrigatórios antes de marcar a spec como aprovada.

1. **Gate 1, forma de `schema_version`: opção B.** O novo formato usa `schema_version: 2` como inteiro (geração principal do contrato, não semver textual). A escolha é deliberada: parsers antigos esperam string e falham fechado em vez de servir silenciosamente um mapa incompatível. O leitor novo deve traduzir tipo/versão inválidos para erro estável `index_schema_unsupported`, sem classificá-los como mapa vazio.
2. **Gate 1, política do leitor: opção R2.** Lista explícita e estrita de schemas suportados. Schema 2 é decodificado estritamente, inclusive `source_identity`; schema legado `"1.0.0"` só pode ser lido por `--index` explícito, com aviso. Nunca carregá-lo implicitamente com `gira.yaml`, nunca reescrever v1 in-place e nunca permitir downgrade de schema em publicação. Usar `index_schema_unsupported` para versão desconhecida e distinguir schema incompatível de `index_corrupt`.
3. **Gate 1, nome e forma da proveniência: `source_identity`.** Campo embutido no mapa, com `schema_version: 2`, `manifest_sha256` sobre os bytes crus, `extractors` e `sources`. A lista inclui todos os repositórios declarados no manifesto, mesmo os sem evidência; entradas ordenadas deterministicamente, sem URL crua, usando `url_sha256`. `tool_id` e hash do executável não entram nesse campo embutido. SHA das fontes participa da avaliação de atualidade/staleness, separado da evidência de locators.
4. **Gate 2, manifesto fora da raiz Git: opção O1.** Quando houver overrides que gerem locators, bloquear manifesto fora de uma worktree Git com `manifest_outside_git_root`; se a consulta Git falhar por Git ausente, dubious ownership ou erro inesperado, falhar fechado com `git_root_unavailable`. Sem overrides que gerem locators, indexação não deve falhar apenas por não haver raiz Git. Resolver symlinks, derivar caminho relativo à raiz com barras `/`, e sanear variáveis de redirecionamento Git (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`, `GIT_NAMESPACE`, `GIT_PREFIX` e `GIT_CEILING_DIRECTORIES`) nos probes de raiz e em `init`. Preservar a configuração do usuário para clone/fetch conforme ADR 0009.
5. **Gate 2, `ManifestPath` vazio em eval: remover o fallback silencioso.** O chamador deve passar o manifesto real, incluindo os bytes crus usados pela chave. Se eval precisar gerar locator e não tiver caminho/bytes do manifesto, retornar erro estável `manifest_path_required`; não inventar `gira.yaml`. Quando não há overrides/locators, não criar locator fictício nem falhar por esse único motivo.
6. **Gate 3, `main.date` em `tool_id`: opção C.** Excluir a setting `-ldflags` inteira do cálculo de `tool_id`; manter `main.version` e `main.commit` como campos próprios, além de GoVersion e dependências. O fato de `-trimpath` já remover `-ldflags` de BuildInfo reforça que depender da string crua não é estável. Se novos flags de build alterarem semântica, registrá-los como campos explícitos de identidade e adicionar teste.
7. **Gate 3, hash do executável em builds de desenvolvimento: manter.** Se `vcs.modified` ou a ausência de `vcs.revision`/commit confiável indicar build de desenvolvimento, incluir o hash do executável. Aceita-se reindexação depois de rebuild para não fazer binários diferentes parecerem a mesma ferramenta. Build de release limpo não deve mudar `tool_id` só por `main.date`.
8. **Gate 4, Windows: opção W1 com limite explícito.** Aceitar a evidência existente para o perfil do runner hospedado (admin elevado, Developer Mode e `core.symlinks=true`). Não generalizar para máquinas de usuários comuns. No SC7, tentar os testes de symlink e, quando o ambiente não permitir criá-lo, pular apenas essa célula com motivo registrado. O comportamento de `git archive | tar -x` com nomes que contêm tab no Windows continua uma verificação de compatibilidade a adicionar antes da aprovação final da spec.
9. **Gate 5: adiar a execução para a primeira tarefa de implementação autorizada.** Não construir um protótipo descartável antes da autorização heavy. A tarefa inicial de publicação terá testes de falha antes/depois do rename, fsync de diretório, `visible_unconfirmed`, publicação parcial de `init --output`, preservação de snapshot anterior e hash, em Linux/macOS/Windows. Esses testes são critério obrigatório para aceitar a implementação e mesclar sua PR, não uma pré-condição circular para começar a implementação autorizada.

## Status após as decisões

- Gates 1–3 têm direção decidida, mas auditoria de consumidores/schema, testes de compatibilidade, política do eval e testes de regressão de `tool_id` ainda precisam ser concluídos.
- Gate 4 é aceito apenas para o perfil de runner testado; o teste de `git archive | tar -x` com tab permanece aberto.
- Gate 5 pertence à verificação da implementação, se e quando houver autorização heavy.
- A spec segue `draft`; a tarefa-mãe segue `todo`; nenhuma implementação, ADR 0014 ou tarefa-filha foi autorizada.
