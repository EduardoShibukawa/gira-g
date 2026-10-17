# Especificação revisada: publicação versionada de índices do Gira-G (Etapa 2, não iniciada)

Não alterei código e não criei arquivos nem commits; `git status` está limpo em `main` (`0d5f248`). Para testar alguns pontos, compilei o HEAD em `/tmp` e rodei um script descartável com `HOME`, `XDG` e `gitconfig` isolados. Já apaguei os dois. Os resultados aparecem como **[sonda]**. O que veio só da leitura do código está marcado **[código]**, e o que é inferência está marcado **[INFERÊNCIA]**. Nenhum teste do repositório foi executado.

O PDF *Gira-G_Analise_Maturidade* é uma análise de documentos, sem execução de código, e não serve como prova. Usei dele quatro critérios como requisitos:
- **Fase M2:** reindexar menos exige motivo comprovado ou bloqueio.
- **P2:** a chave do cache combina conteúdo, versão do extractor e versão do schema.
- **Métrica de frescor:** nenhum mapa desatualizado pode ser apresentado como atual.
- **P0:** um nome canônico de saída. Entre os testes adversariais sugeridos está "nome idêntico em repositórios diferentes".

## 0. O que a sonda mostrou no HEAD

| # | Cenário | Resultado |
|---|---|---|
| S1 | `gira init . --output .gira/macro-map.json` e depois `gira locate` sem `--index` | `index_missing`, saída 1. Quando existe `gira.yaml`, o resolvedor nunca lê o arquivo de `--output` (`main.go:312-334`). O README:43 afirma o contrário. |
| S2 | `gira index` no cache, novo commit, `gira init --force --output …`, depois `locate PayService` sem `--index` | **Retorna `[]` com saída 0.** O mapa desatualizado do cache foi servido como atual. Com `--index` o resultado é o correto. |
| S3 | Mesmo `gira.yaml` com overrides, rodando `gira index gira.yaml` e depois `gira index /abs/gira.yaml` | Os mapas diferem (sha256 `6728…` contra `3deb…`): `locator.path` vale `"gira.yaml"` num e o caminho absoluto no outro. |
| S4 | Uma linha de comentário inserida antes dos overrides | Os mapas diferem: `start_line` e `end_line` mudam de 10/12 para 11/13. |
| S5 | Arquivo `proto/café.proto` com `core.quotePath` no padrão do git | A indexação falha: `ls-tree` sem `-z` devolve o caminho entre aspas (`cache.go:149`). Com `quotePath=false`, passa. |
| S6 | O mapa contém a URL do repositório? | 0 ocorrências da URL. |
| S7 | Onde fica o store de índices | Em `~/.cache/gira/indexes`, ignorando `XDG_CACHE_HOME`. O cache de repositórios do `gitcache` respeita `XDG_CACHE_HOME` (`cache.go:31-40`). |

## 1. Contrato de `--output`

| Opção | Efeito | Problema |
|---|---|---|
| A. Manter como está | `--output P` é o destino que vale, e o store só é usado sem a flag | S1 e S2 continuam. O gate compara só com o P anterior. |
| B. P vira cópia exportada e o store é sempre gravado | O resolvedor sempre encontra o índice | **Quebra fluxos que indexam commits antigos**: `gira-arch-diff/SKILL.md:14` (`--output base.json` e `head.json`), variantes perturbadas do bench e `tools/index-bundle`. O commit base seria barrado pelo gate contra o head mais novo da mesma linhagem. Testes que usam `--output` também passariam a escrever no store do usuário. |
| C. `--output DIR` como raiz alternativa do store | — | Quebra todos os chamadores que passam um caminho de arquivo. |
| D. Remover a flag | — | Quebra bench, ferramentas, skill, README, site e cerca de 10 testes. |

**Recomendação: A′, publicação destacada.**
- **Sem `--output`:** publica no store versionado. É o único lugar que o resolvedor lê a partir do `gira.yaml`.
- **Com `--output P`:** grava somente o arquivo P, de forma atômica. O gate compara com o conteúdo anterior de P, sob trava em `P.lock`. Nada é gravado no store nem em `heads`. P só é lido com `--index P`. P não é imutável; isso fica documentado.
- **Ordem em `gira index`:** ler os bytes do manifesto **uma única vez**, calcular as chaves, construir, travar, rodar o gate, publicar e destravar. Esse comando nunca escreve `gira.yaml`.
- **`gira init` sem `gira.yaml`:** gera os bytes em memória, constrói com eles, publica no store ou em P, e **só depois** cria `gira.yaml` com `O_EXCL`, fsync do arquivo e fsync do diretório. Se o build ou o gate falhar, nenhum `gira.yaml` é criado. Isso muda o comportamento de hoje (`main.go:551-556`).
- **`gira init --force`:** gera os bytes repinados em memória e publica. Antes do `rename` do `gira.yaml`, compara o hash do arquivo em disco com o que foi lido; se divergirem, aborta com conflito e a versão já publicada fica órfã. Se o gate rejeitar, `gira.yaml` fica idêntico byte a byte e o resolvedor continua na versão anterior. **Não há rollback, porque o arquivo nunca foi trocado.**
- **Efeito sobre S2:** depois de `init --force --output`, a chave do `gira.yaml` novo não existe no store, então o resultado é `index_missing` com uma dica para usar `--index`. Deixa de ser `[]` silencioso.
- **Raiz do store (decisão D2):** `$XDG_CACHE_HOME/gira/indexes`, senão `~/.cache/gira/indexes`, senão `<raiz do projeto>/.gira/indexes`. Uma única função, compartilhada por quem escreve e por quem resolve.

## 2. Durabilidade por plataforma

Cada publicação devolve `Durability ∈ {confirmed, unconfirmed}` e o motivo.
- **Visível:** o `rename` final deu certo e um leitor já enxerga a versão nova.
- **Confirmada:** todos os syncs disponíveis retornaram nil: arquivo, diretório temporário, `versions/`, arquivo de head e `heads/`.
- **Não confirmada:** a versão está visível, mas algum sync não existe na plataforma ou falhou.

| Plataforma | Primitivas (Go 1.24, conferido no GOROOT) | Estado | Saída |
|---|---|---|---|
| Linux | `fsync` no arquivo e no diretório | confirmed se nada falhar | Se um sync que existe falhar depois do `rename`: código ≠ 0, "visível, durabilidade não confirmada". |
| macOS | `File.Sync` usa `F_FULLFSYNC`; com `ENOTSUP`, por exemplo em SMB, **cai silenciosamente** para `fsync` (`fd_fsync_darwin.go:22-28`) | confirmed, com a ressalva de que a queda para `fsync` não é detectável | Igual ao Linux. Não verifiquei se `F_FULLFSYNC` em diretório funciona no APFS **[INFERÊNCIA]**; o CI `macos-latest` existe, mas não olhei o resultado dele. |
| Windows | `FlushFileBuffers` no arquivo. Não há sync de diretório: hoje `syncDirectory` só retorna nil. `os.Rename` chama `MoveFileEx(MOVEFILE_REPLACE_EXISTING)` sem `WRITE_THROUGH` (`syscall_windows.go:352-361`). | **sempre unconfirmed** | Código 0 com aviso. Nunca reportar "confirmed". O comentário atual em `store.go:178-180` dá a entender que o caminho é durável; precisa ser corrigido. |

No Windows, `os.Open` não usa `FILE_SHARE_DELETE` (`syscall_windows.go:374`). Por isso um `rename` que substitui um arquivo aberto por um leitor falha. Mitigação: as versões nunca são substituídas, e a troca do head tenta de novo por tempo limitado, sob a trava.

## 3. Fingerprint

**Algoritmo:** SHA-256 sobre registros com tag de domínio e prefixo de tamanho (`u64be(len) ‖ bytes` por campo), o que torna a codificação injetiva.

| Chave | Entradas |
|---|---|
| `source_key` | `"gira.index.source/v1"`, `schema_version` (`"1.0.0"`), **bytes crus** do manifesto (os mesmos que foram parseados) e o locator normalizado (D3: o basename, como `gira.yaml`) |
| `tool_id` | `"gira.index.tool/v1"`; `main.version` e `main.commit`; da `debug.BuildInfo`: GoVersion, Main (Path, Version, Sum), **todas** as Settings (`vcs.revision`, `vcs.modified`, `CGO_ENABLED`, GOOS, GOARCH, `-ldflags`, `-trimpath`) e Deps (path, version, sum). Se `vcs.modified=true`, se não houver `vcs.revision`, ou se `commit=="none"`, entra também o sha256 de `os.Executable()`. |
| `version_id` | H(`"gira.index.version/v1"`, `source_key`, `tool_id`). É o nome do diretório. |
| `lineage_key` | H(`target_system.name`, pares (id, `NormalizeURL(url)`) ordenados). Escolhe o head usado como base do gate. |

**Tudo o que pode alterar o mapa**, segundo `indexer.Build` e as funções que ele chama:
1. **Bytes do manifesto:** `target_system.name`, e por repositório `id` (vai para `locator.repo`), `sha`, `path_scope`, `url` (só serve para buscar o repo; entra na chave sem custo além de um cache miss) e `ref`; a ordem dos repositórios; overrides de relationships e metadata, com os valores já tipados pelo YAML e as **linhas** (`manifest.go:65-91`).
2. **Locator do manifesto:** em `ApplyOverrides`, `Locator.Path` recebe `manifestPath` exatamente como foi passado (`pipeline.go:360,423`). Ver S3.
3. **Conteúdo do git no SHA**, dado por `ls-tree` + `show`, e a data do commit (`%cI`).
4. **Binário:** extractors, `DefaultExtractorVersion="1.0.0"` (uma constante que não acompanha o código), toolchain Go (`go/scanner` e `go/parser` em `normalizeGoSource` e `layer3_go_hints`), versão do `yaml.v3`, GOOS **[INFERÊNCIA: separadores de caminho]**.
5. **Ambiente do git na leitura:** `core.quotePath` (S5) e `refs/replace` **[INFERÊNCIA, não testado]**. Isso é eliminado em vez de entrar no hash: `ls-tree -z`, `cat-file blob <oid>`, e nas operações de leitura `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=os.DevNull` e `GIT_NO_REPLACE_OBJECTS=1`. **Clone e fetch continuam usando a config do usuário**, porque o ADR 0009:36 depende dela para credenciais.
6. **O que não altera o mapa:** `--offline`; o relógio (`StampEvidence` sobrescreve `time.Now`); `AllowShrink`, que `Build` não usa.

**Validar uma versão existente** (quem lê e quem escreve):
- `meta.json` decodificado em modo estrito, com formato `gira.index.version/v1`;
- `meta.version_id` igual ao nome do diretório e igual a H(`meta.source_key`, `meta.tool_id`);
- sha256 de `macro-map.json` igual a `meta.map_sha256`;
- `LoadFile` sem erro;
- `target_system` coerente com o meta.

Qualquer falha resulta em `index_corrupt` e falha fechada, sem cair para outra versão.

**Republicar o mesmo `version_id`:** se os bytes forem iguais, a operação é idempotente: o gate roda de novo contra o head atual e só então o head avança. Se forem diferentes, o erro é `nondeterministic_build` e nada é sobrescrito.

**Índices legados:** não são importados para `versions/`. Eles não registram ferramenta nem manifesto, e inventar um fingerprint para eles seria exatamente uma colisão semântica. Uso permitido:
- (a) como base do gate na primeira publicação, quando o sistema não tem nenhum head;
- (b) na leitura, só enquanto o sistema não tem nenhum estado versionado, sempre com aviso "índice legado sem proveniência".

O arquivo legado não é mais gravado. Os caminhos legados são `~/.cache/gira/indexes/<sys>/macro-map.json`, `<raiz>/.gira/indexes/<sys>/…` e `./.gira/indexes/<sys>/…`.

**Retenção (D7):** por linhagem, ficam o head e as K=3 versões anteriores em `head.history`. A remoção acontece sob a trava: primeiro `rename` para `versions/.trash-*`, depois `RemoveAll`. **Versões órfãs, fora de qualquer `history`, nunca são apagadas automaticamente**: um `gira.yaml` pode apontar para elas, por exemplo depois de uma falha antes de gravar o head. Se uma versão já removida for pedida de novo, o rebuild determinístico recria o mesmo id.

**Prevenção de colisões semânticas:**
- Dois projetos com o mesmo `name` hoje colidem em `indexes/<name>/macro-map.json` (`main.go:476-482`). Passam a ter `version_id` e `lineage_key` diferentes.
- Uma linhagem nova de um sistema que já existe mostra "baseline: none (nova linhagem)" e não fica em silêncio.

### Prova de que o argumento mais fraco se sustenta

**Afirmação:** se mapa(X) ≠ mapa(Y), então `version_id`(X) ≠ `version_id`(Y), salvo colisão de SHA-256. Vale sob quatro premissas:
- **P1:** `Build` é uma função determinística das entradas 1 a 4 acima, com o ambiente do item 5 eliminado.
- **P2:** um SHA-1 do git fixa o conteúdo.
- **P3:** `tool_id` identifica o binário de forma injetiva.
- **P4:** a codificação é injetiva.

Pela contrapositiva: ids iguais implicam pré-imagens iguais, logo entradas iguais, logo mapas iguais.

**P1 é a premissa fraca.** Ela é empírica, não decorre de nenhum mecanismo. Contraexemplos, cada um com o elemento que o resolve:

| Par de builds com mapas diferentes | Chave ingênua que colide | O que separa |
|---|---|---|
| Caminho relativo contra absoluto (S3) | H(bytes) | Locator normalizado, o que torna os mapas iguais, mais locator na chave |
| Comentário que desloca linhas (S4) | Hash semântico | Bytes crus |
| gira vN contra vN+1 com o mesmo manifesto | H(manifesto) | `tool_id` |
| Dois builds de desenvolvimento, ambos `0.1.0-dev`/`none` (`main.go:128-129`) | `main.version` | `vcs.modified` mais hash do executável |
| Go 1.24 contra 1.25, ou outra versão do `yaml.v3` | Versão do gira | GoVersion e Deps |
| Mesmo SHA com `path_scope` ou `id` diferentes | H(SHAs) | Bytes crus |
| Mesmo `name` em projetos diferentes | Diretório por nome | Bytes crus e linhagem |
| `quotePath` (S5) e `refs/replace` | Não chegam a colidir: geram falha ou dependência do ambiente | Eliminação (item 5) |
| **Mesmas entradas com bytes diferentes** (não determinismo) | **Qualquer chave** | Só é detectado, via `nondeterministic_build`. `TestBuildIsDeterministicAcrossColdRuns` cobre apenas uma fixture. |

## 4. Concorrência

**Mecanismo:** trava consultiva de SO num arquivo `<raiz>/<sys>/.lock`, ou `P.lock` no modo destacado. O arquivo de trava nunca é apagado.
- Unix: `unix.Flock(LOCK_EX|LOCK_NB)` com polling.
- Windows: `windows.LockFileEx(EXCLUSIVE|FAIL_IMMEDIATELY)` com polling.
- O pacote `golang.org/x/sys v0.30.0` já está no `go.mod` como indireto e passa a ser direto.
- Timeout de 60 s, com erro `index_locked`.

A trava cobre esta sequência: ler o head, carregar a base, rodar o gate, materializar a versão, gravar o head, fazer a retenção e limpar `.tmp-*` e `.trash-*`. Lixo encontrado por quem segura a trava só pode ser de um processo que morreu. O build fica fora da trava. Os leitores não travam, porque as versões são imutáveis.

| Plataforma | Garantia | Limitação |
|---|---|---|
| Linux | Exclusão entre processos e dentro do processo (por descrição de arquivo aberta). Liberada pelo kernel quando o processo morre. | Em NFS vira trava POSIX via NLM; sem `lockd`, dá `ENOLCK` e **falha fechada**. |
| macOS | Mesma garantia, em APFS e HFS+ | Em NFS e SMB, `flock` dá `ENOTSUP` e **falha fechada**. |
| Windows | Exclusão por handle, inclusive dentro do mesmo processo | A liberação depois de um crash "depende de recursos do sistema", conforme a documentação da Microsoft; por isso o polling tem timeout. |
| Todas | — | Só serializa binários que seguem este protocolo. Um gira antigo continua gravando o caminho legado sem trava. |

## 5. Inventário de compatibilidade

Busca feita por um scout com os padrões listados em `agent://IndexInventory`, mais duas buscas minhas no site, nas skills e em `tools`. Limite: `website/versioned_docs` só foi coberto pelo scout, que encontrou `diff.md` e `doc.md`.

- **Resolvedor:** `loadIndex` (`main.go:300-397`). Recebe chamadas de serve, impact, slice, explain, locate, export, view, tour, flows e doc (cada um com `--index`). Também é usado por `skills nudge` (`:1376-1379`), que **engole erros**.
- **Outros leitores:** `diff` (`:1232-1241`); MCP `diff_maps` (`internal/mcp/server.go:613-632`); `Store.Load`, `LoadFile`, `List` (`List` identifica um sistema pelo nome de arquivo `macro-map.json`); `bench/layer0/gira.go`; `cmd/layer1` e `bench/layer2` via `--index`; `internal/skills/install.go` e `query.go`, que gravam `serve mcp --index <path>` na configuração do host.
- **Escritores:** `indexManifest` (`main.go:466-485`), `Store.Save`, `SaveFile`, `Publish`, `PublishFile`. `internal/eval` chama `Build` diretamente e não publica.
- **`--output` de index e init:** `main.go:412,509`; `bench/phases.json:20`; `internal/benchtest/benchtest.go:76`; `tools/index-bundle/bootstrap.sh:42,49`; skill `gira-arch-diff/SKILL.md:14`; README:41-59; site em `cli/index.md`, `init.md`, `map.md` e `guides/change-maker-workflow.md`. O `--output` de `export`, `view` e `skills manifest` é outra coisa e não muda.
- **`--index` na documentação:** `doc.md:26` e `serve.md:25` descrevem uma ordem de resolução errada, a mesma divergência de S1.
- **Testes:** `store_test.go` (7 testes), `incremental_test.go` (gate), `main_test.go` (`TestCLI_Index`, `TestCLI_LoadIndexResolutionAndFailClosed`, ErrorHandling, ServeMCP, Export, View), `init_test.go` (`TestSCSingleInit`, `TestInit_ForceDoesNotAllowSymbolLoss`), `e2e_test.go`, testes de `cmd/gira` que usam `--index`, `mcp/envelope_test.go`, `skills_test.go:254,266`, `bench/layer0_test.go`.
- **ADRs:** **nenhum ADR trata de onde o índice fica ou de como é publicado.** O ADR 0009 fala do cache de repositórios e de usar a config git do usuário. O ADR 0001 fala de locators. O ADR 0007:55 exige determinismo de 100% e sustenta a premissa P1. Será preciso um ADR 0014 novo.
- **`.gira/index.json` na raiz:** é gitignored; o scout inferiu que não é rastreado a partir de `.git/index` **[INFERÊNCIA]**. É o índice de dogfood e estava 3 commits atrasado segundo o registro de tarefa.

## 6. Plano incremental (nada iniciado)

| Etapa | Arquivos | Invariante | Testes de aceitação |
|---|---|---|---|
| 2a. ADR 0014 e esta especificação | `docs/adr/0014-*` | — | Aprovação do dono |
| 2b. Leituras do git independentes do ambiente | `internal/gitcache/cache.go`, `gitcache_test.go` | Bytes iguais com qualquer gitconfig | Nomes com não-ASCII e tab indexam com `quotePath` padrão; `quotePath` true contra false dá bytes idênticos; `refs/replace` no cache não muda nada |
| 2c. Locator normalizado | `main.go` (`indexManifest`), doc de `indexer.Options` | Invocação relativa e absoluta dão os mesmos bytes | Regressão de S3; snapshots atualizados se forem afetados |
| 2d. Chaves | `internal/store/key.go` | Codificação injetiva | Cada linha da tabela de contraexemplos muda a chave; `("ab","c")` ≠ `("a","bc")`; build modificado inclui o hash do executável |
| 2e. Trava | `internal/store/lock_{unix,windows}.go`, `go.mod` | Um único escritor | Dois que tentam adquirir no mesmo processo se excluem (nos 3 sistemas do CI); timeout gera `index_locked`; processo filho morto libera a trava |
| 2f. `PublishVersion` e `PublishDetached`, `PublishFile` removido | `internal/store/store.go`, `store_test.go` | Versão existe ⇒ passou pelo gate ou foi autorizada; versões imutáveis | Falha **antes** do `rename`: nenhuma versão, head intacto, temporários limpos na próxima trava. Falha **depois** do `rename` da versão e antes do head: versão válida e visível, head intacto, nova execução idempotente que avança o head. **Gate rejeita:** nada é gravado. **Falha ao publicar o head:** versão visível e erro. Head corrompido ou apontando para versão ausente: falha fechada até `--recover-corrupt`. Bytes diferentes no mesmo id: `nondeterministic_build`. Durabilidade: sync com erro no Unix dá erro, Windows dá unconfirmed. |
| 2g. Resolvedor | `main.go` (`loadIndex`), `List` | Nunca apresentar desatualizado como atual | S1 com store e S2 resultam em `index_missing`, não em `[]`; ferramenta divergente gera aviso (D5); legado só sem estado versionado; ambiguidade; `--index` inalterado |
| 2h. CLI | `main.go` (`runIndex`, `runInit`) | `gira.yaml` é gravado por último | **Primeira inicialização:** build falha e `gira.yaml` não existe. `--force` com gate rejeitado: `gira.yaml` idêntico e a resolução continua na versão anterior. **Concorrência:** dois `Run(index)` na mesma linhagem; cada versão aceita passou pelo gate contra a predecessora no `history`. **Reinício:** `.tmp`, órfã e `.trash` sobrando, e a execução seguinte passa e limpa. Conflito no `--force` |
| 2i. Retenção | `store.go` | Head, K anteriores e órfãs preservados | K+2 publicações deixam head + K; órfã sobrevive; `rename` falhando adia a remoção |
| 2j. Chamadores e documentação | README, `website/docs/cli/*`, `skills/install.go`, catálogo de skills, `bench`, `tools`, `benchtest` (`t.Setenv` de `XDG_CACHE_HOME`) | Documentação igual ao código | Validador e `mise run check` |

## Decisões que preciso de você

- **D1:** `--output` como publicação destacada (A′). Recomendado.
- **D2:** o store passa a respeitar `XDG_CACHE_HOME`. Recomendado.
- **D3:** locator do override vira o basename. Recomendado. Muda os bytes de mapas que têm overrides.
- **D4:** `source_key` sobre os bytes crus. Recomendado. Consequência: editar um comentário no `gira.yaml` exige reindexar.
- **D5:** quando existe versão com o mesmo `source_key` mas de outro `tool_id`, servir com aviso (recomendado) ou falhar.
- **D6:** se um sync disponível falhar, código ≠ 0. No Windows, código 0 com "não confirmado".
- **D7:** K = 3, sem remoção automática de órfãs.
- **D8:** trava com timeout de 60 s e falha fechada em NFS ou SMB sem suporte a trava.
- **D9:** sem importar legado; leitura transitória do legado com aviso.
- **D10:** `init` não deixa `gira.yaml` para trás quando falha.

## Riscos residuais

- P1 só se verifica por testes adversariais e por detecção quando uma versão é republicada; não há garantia formal.
- A identidade de um SHA-1 no git é uma premissa teórica.
- Atualizar o binário muda o `tool_id` (ver D5).
- Windows: durabilidade nunca confirmada, `rename` bloqueado por arquivo aberto, liberação lenta da trava depois de crash.
- macOS em SMB: a queda silenciosa para `fsync`.
- Binários antigos e novos convivendo sem trava comum.
- Corrida no `gira.yaml` durante `--force`: o arquivo do projeto não fica sob trava.
- Órfãs acumulando espaço em disco.
- `serve mcp` mantém o mapa em memória.
- `nudge` engole os erros do resolvedor mais estrito.
- Hash do executável a cada consulta em builds de desenvolvimento (cerca de 7 MB).
- O gate na CLI recebe `DeletedPaths=nil`: hoje, apagar um arquivo já exige `--allow-symbol-loss`. Não muda nesta etapa.
- Uma linhagem nova (por exemplo, checkout movido de lugar) começa sem base para o gate.

Aguardo sua autorização explícita antes de qualquer etapa.