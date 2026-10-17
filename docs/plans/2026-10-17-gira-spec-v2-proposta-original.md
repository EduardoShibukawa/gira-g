# Especificação revisada v2: publicação versionada do Gira-G

**Situação:** a Etapa 2 não foi iniciada nem autorizada. Neste turno não alterei código, não toquei em Sources e não criei commits nem arquivos; `git status` está limpo em `0d5f248`. As três ações deste turno foram somente leitura: li o log do CI com `gh run view`, li o código-fonte do Go em `GOROOT` e listei `~/.cache/gira/indexes`.

**Como marquei cada afirmação:**

| Marca | Significado |
|---|---|
| **[código]** | Inspecionei o código-fonte. |
| **[CI]** | Li o log do job. |
| **[sonda]** | Execução manual descartável, relatada no turno anterior. |
| **[listagem]** | Leitura do sistema de arquivos nesta máquina. |
| **[INF]** | Inferência minha. |
| **[DESC]** | Desconhecido. |

Nenhum teste do repositório foi executado por mim.

## 0. Falha do CI no Windows: diagnóstico

1. **[CI]** Os casos 1 e 1b passaram (`index_missing` era o esperado). O caso 2 falhou em `main_test.go:641` com `index_missing … C:\Users\RUNNER~1\AppData\Local\Temp\…\002\gira.yaml`. O caso 3 nem rodou, porque o caso 2 encerra o teste com `Fatalf`.
2. **[código]** No Windows, `os.UserHomeDir()` lê `USERPROFILE`, não `HOME` (Go 1.24, `os/file.go:589-592`). O teste muda só `HOME` (`main_test.go:569`) e grava o índice em `tempHome\.cache\…` (`:578`). O resolvedor procura em `%USERPROFILE%\.cache\…` (`main.go:326,370`).
3. **[código]** Em produção, escrita (`main.go:475`) e leitura (`:326,370`) usam a mesma função, `os.UserHomeDir()`, e portanto concordam entre si. A falha não indica defeito de produção. Isso só fica **confirmado** quando o teste corrigido passar no CI do Windows. Essa execução ainda não aconteceu.
4. **Defeito de isolamento que não estava no relato:**
   - **[código]** `TestCLI_Index` roda `gira index` sem `--output` e sem isolar `HOME` (`main_test.go:518`). Com isso grava no cache real do usuário, em todos os sistemas operacionais.
   - **[listagem]** Nesta máquina existe `~/.cache/gira/indexes/index-test-system`, com data de 17/out 17:34. **[INF]** O nome é o mesmo da fixture (`:420`).
   - No Windows, `TestCLI_LoadIndexResolutionAndFailClosed` lê o `%USERPROFILE%` real do runner. O caso 1 só passa porque os nomes de sistema não coincidem.

**Correção proposta (Etapa 2.0, depende de autorização):**
- usar `t.Setenv` para `HOME` e `USERPROFILE`;
- usar `t.Chdir`, que existe no Go 1.24 (conferi em `testing.go:1628`);
- isolar o passo 2 de `TestCLI_Index` da mesma forma;
- depois disso, rodar o CI nos três sistemas.

Os itens 8 e 9 do seu pedido se contradizem: corrigir o teste é mudança de código, e o item 9 proíbe código sem autorização. Por isso a correção está especificada, mas não aplicada.

## 1. D1: o `--output` de `index` e o de `init` passam a ter semânticas diferentes

| Comando | Semântica | Garantia |
|---|---|---|
| `index` sem `--output` | Publicação **canônica** no store versionado. | É o único alvo que o resolvedor lê a partir de um `gira.yaml`. |
| `index --output P` | **Snapshot destacado**: grava somente P. O gate compara com o conteúdo anterior de P, sob a trava de P. Não grava no store nem nos heads, e nunca escreve `gira.yaml`. | Preserva `gira-arch-diff` (`base.json`/`head.json`), `bench/phases.json`, `tools/index-bundle` e `benchtest`. A saída avisa: "snapshot destacado; não é o índice ativo do projeto". |
| `init` (com ou sem `--output`) | Publicação **canônica** obrigatória. Com `--output P`, grava **além disso** uma cópia exata dos bytes da versão canônica em P. | `init` só termina com sucesso se, nesta ordem, a versão canônica estiver publicada, o snapshot P (se pedido) estiver gravado e o manifesto estiver gravado. Qualquer falha antes de gravar o manifesto retorna erro, e o manifesto ativo fica intacto ou nem é criado. |

**Opções descartadas para `init --output`:**
- Rejeitar a flag: quebraria `init_test.go:149`, o README:50-57 e a página `cli/init.md`.
- Gravar só P, sem criar manifesto: o `init` perderia sentido.

**Sobrescrita de P pelo `init`:** P pode ser sobrescrito se não existir ou se for um macro-map válido; se for outra coisa, a sobrescrita exige `--recover-corrupt`. P não passa por um segundo gate, porque é cópia de uma versão que já passou pelo gate canônico.

**Consequência para testes:** os testes de `init` passam a gravar no store e precisam de isolamento. Em `TestInit_ForceDoesNotAllowSymbolLoss`, a rejeição passa a vir do gate canônico, e o teste deve afirmar que `gira.yaml`, o head e P ficaram inalterados.

## 2. D5: incompatibilidade de ferramenta falha fechado

- O resolvedor só aceita o `version_id` exato, isto é, `source_key` mais o `tool_id` atual.
- Se só existir uma versão com o mesmo `source_key` e outro `tool_id`, o erro é `index_tool_mismatch`. A mensagem mostra o caminho dessa versão e a identidade da ferramenta que a gerou.
- Não há flag para aceitar a divergência. A saída explícita continua sendo `--index <caminho>`, que é uma escolha consciente do usuário.
- **Custo aceito:** toda atualização do gira exige reindexar. Em builds de desenvolvimento com `vcs.modified=true`, o `tool_id` inclui o hash do executável, então **cada rebuild** invalida a resolução. Isso afeta o dogfood deste repositório.

## 3. D6: o que a durabilidade não confirmada faz com a gravação do manifesto

Chamo de "gravação do manifesto" a escrita do `gira.yaml` pelo `init`: arquivo temporário, sync e `rename`.

| Plataforma | Estado depois de publicar a versão e o head | Efeito na gravação do manifesto |
|---|---|---|
| Linux / macOS | `confirmed` | Grava o manifesto. |
| Linux / macOS | `visible_unconfirmed` (algum sync existente falhou) | **Bloqueia** a gravação, código ≠ 0. Rodar de novo cai no caminho idempotente, que repete o sync dos arquivos e diretórios da versão e do head, e só grava o manifesto se tudo for confirmado. |
| Windows | Sempre `visible_unconfirmed`, porque não existe sync de diretório e `MoveFileEx` é chamado sem `WRITE_THROUGH` **[código: GOROOT]** | Grava o manifesto, sai com código 0 e emite um aviso explícito. |

Em nenhuma plataforma prometo recuperação após queda de energia. No Windows, depois de uma queda, o manifesto, o head e a versão podem estar cada um no estado antigo ou no novo, em qualquer combinação **[INF: nenhuma primitiva usada ordena essas gravações]**. O que se garante é só a **detecção** na leitura:
- versão ausente: `index_missing`;
- hash do conteúdo divergente: `index_corrupt`;
- head inválido ou apontando para versão ausente: falha fechada.

Recuperar exige reindexar, e a reindexação depende de o cache git estar íntegro. Esse estado é **[DESC]** após uma queda.

Não estou declarando que o `rename` é atômico quando o *processo* cai no Windows: a Microsoft não documenta isso formalmente **[INF]**.

**Descartado:** usar `MOVEFILE_WRITE_THROUGH`. A documentação só o garante para movimentos do tipo copiar e apagar, então ele não justificaria marcar a publicação como `confirmed`.

**Para decidir:** criar `--require-durable`, que falharia sempre no Windows. Minha recomendação é não criar nesta etapa.

## 4. Travas: manifesto, store e snapshot

Usam `flock` no Unix e `LockFileEx` no Windows, via `golang.org/x/sys`, com polling e timeout. Os arquivos de trava nunca são apagados.

| Trava | Arquivo | Quem usa |
|---|---|---|
| **M** (manifesto) | `<dir(EvalSymlinks(Abs(manifesto)))>/.gira/locks/<lower(basename)>.lock` | `init`, e qualquer comando futuro que grave o manifesto |
| **S** (store) | `<raiz>/<sys>/.lock` | publicação canônica |
| **P** (snapshot) | `<dir(EvalSymlinks(dir de P))>/.<lower(basename(P))>.lock` | `index --output` e o snapshot do `init` |

**Por que M usa um arquivo vizinho do manifesto, e não um hash do caminho:**
- A exclusão passa a depender da identidade física do arquivo de trava. Grafias diferentes do mesmo diretório (relativa ou absoluta, maiúsculas ou minúsculas em sistemas que não diferenciam) abrem o mesmo arquivo.
- Resolver symlinks faz com que um `gira.yaml` que é link para outro arquivo use a mesma trava que o arquivo de destino.
- Usar o basename em minúsculas pode serializar além do necessário quando há dois manifestos distintos por caixa num sistema que diferencia maiúsculas. Isso é inofensivo.
- **Limite:** dois hardlinks com nomes diferentes não se excluem.

**Ordem global: M → S → P.** Uma trava nunca é adquirida enquanto se segura outra que venha depois na ordem, e as travas são liberadas em ordem inversa. Sequência do `init`:
1. adquirir M; ler o manifesto ou confirmar que ele não existe;
2. construir;
3. adquirir S; gate, versão, head, sync; liberar S;
4. adquirir P; gravar o snapshot; liberar P;
5. comparar o hash do manifesto em disco com o lido no passo 1, gravar o manifesto, liberar M.

O `index` canônico usa só S, e o `index --output` usa só P. Nenhuma sequência adquire na ordem inversa, então não há deadlock.

**Cobertura e limitações:**
- M exclui apenas processos gira. Um editor de texto não respeita a trava. A comparação de hash antes do `rename` detecta mudanças feitas por fora, mas sobra uma janela entre a comparação e o `rename`. Esse risco fica registrado como residual.
- **Sistema de arquivos sem suporte a trava**, qualquer erro de trava (`ENOLCK`, `ENOTSUP`, `EOPNOTSUPP`, erros do Windows): falha fechada com `lock_unavailable`, sem gravar nada. Não haverá `--no-lock`.
- **Rede (NFS com `local_lock`, FUSE):** a trava pode ser concedida só dentro da máquina local, e o gira não consegue detectar isso. A exclusão entre máquinas diferentes não é garantida.
- `.gira/locks/` fica dentro do projeto. Se `.gira/` não estiver no `.gitignore` do usuário, a verificação de worktree sujo do `init` vai avisar. O `init` passa a imprimir uma dica sobre isso.

## 5. D7: sem coleta automática nesta etapa

- Não se apaga nenhuma versão, o que inclui órfãs e históricos antigos. `heads/*.json` guardam o histórico completo.
- `--recover-corrupt` apenas move a versão corrompida para `versions/.quarantine-*`.
- A coleta fica adiada até existir um registro de referências (manifestos que apontam para cada `source_key`, com data de última consulta) e uma semântica de recuperação. Ambos ficam fora do escopo.
- **Residual:** o disco cresce sem limite.

## 6. D9: índice legado

- **Com `gira.yaml` presente:** um legado nunca é carregado automaticamente. O legado não registra manifesto nem ferramenta, então não há como demonstrar que corresponde ao `gira.yaml`. Na prática isso significa nunca. O erro `index_missing` cita o caminho do legado e o `--index` explícito.
- **Sem `gira.yaml`:** os arquivos locais explícitos (`.gira/macro-map.json` e os demais de `main.go:338-343`) continuam valendo. A descoberta automática no store passa a considerar só heads versionados e ignora o legado.
- O legado continua servindo como **base do gate** na primeira publicação de um sistema sem head. Isso só torna a publicação mais restritiva: uma rejeição indevida pode acontecer e é resolvida com `--allow-symbol-loss`. A mensagem mostra qual base foi usada.

## 7. Fingerprint: garantia condicional

O algoritmo, as entradas, a validação e a tabela de contraexemplos continuam como no turno anterior. A garantia passa a ser enunciada assim:

> **Se** (P1) `Build` for determinístico e as entradas enumeradas estiverem completas, (P2) o SHA-1 do git fixar o conteúdo, (P3) `tool_id` identificar o binário de forma injetiva e (P4) a codificação for injetiva, **então** mapas diferentes implicam `version_id` diferentes, salvo colisão de SHA-256.

Isso **não é prova absoluta**. P1 é uma afirmação empírica, sustentada por três coisas:
- **(a) testes adversariais com bytes idênticos** variando o cwd, o caminho relativo ou absoluto do manifesto, `HOME`/`USERPROFILE`, `TZ`, `LANG`, `core.quotePath`, `refs/replace` e nomes não-ASCII ou com tab;
- **(b) cobertura de determinismo** além da fixture única de `TestBuildIsDeterministicAcrossColdRuns`;
- **(c) detecção em tempo de execução:** republicar o mesmo id com bytes diferentes gera `nondeterministic_build`.

Uma entrada oculta ainda não descoberta quebra a garantia sem ser detectada até que alguém republique aquele id.

## 8. Etapas e critérios de aceite

Cada etapa exige autorização própria; não avanço automaticamente.

| Etapa | Arquivos | Critério de aceite |
|---|---|---|
| **2.0 Isolamento dos testes** | `cmd/gira/main_test.go` | `t.Setenv` de `HOME`/`USERPROFILE` e `t.Chdir`; passo 2 de `TestCLI_Index` isolado. CI verde em Ubuntu, macOS e Windows. Antes e depois do teste, nenhuma entrada nova aparece no cache real. Os casos 1 a 3 continuam afirmando o mesmo comportamento. |
| 2a. ADR 0014 | `docs/adr/` | Aprovação do dono. |
| 2b. Leituras do git independentes do ambiente | `gitcache` | Bytes idênticos com qualquer gitconfig; nome não-ASCII indexa com a configuração padrão; clone e fetch continuam usando a config do usuário (ADR 0009). |
| 2c. Locator normalizado | `main.go` | Regressão de S3: invocação relativa e absoluta geram bytes idênticos. |
| 2d. Chaves | `store/key.go` | Cada contraexemplo muda a chave; codificação injetiva; o hash do executável entra quando o build está modificado. |
| 2e. Travas M, S e P | `store/lock_*.go` | Dois concorrentes se excluem nos três sistemas. Grafias relativa, absoluta e via symlink do mesmo manifesto se excluem. Timeout produz erro. Processo filho morto libera a trava. Erro de trava injetado resulta em `lock_unavailable` sem nada gravado. A ordem M→S→P é verificada por um hook. |
| 2f. Publicação canônica e destacada | `store.go` | Falha antes do `rename`: nada visível. Falha depois do `rename`: versão visível, head intacto, nova execução idempotente. Gate rejeita: nada gravado. Head corrompido ou apontando para versão ausente: falha fechada. Bytes divergentes no mesmo id: `nondeterministic_build`. Sync falho no Unix: `visible_unconfirmed` com erro. No Windows, o CI afirma `unconfirmed` com código 0. |
| 2g. Resolvedor | `main.go` (`loadIndex`) | S1 e S2 resultam em `index_missing`. Divergência de ferramenta: `index_tool_mismatch`, sem carregar nada. Legado com `gira.yaml`: não é carregado. A descoberta automática ignora o legado. `--index` continua igual. |
| 2h. CLI | `runIndex`, `runInit` | Primeira inicialização com build falhando: `gira.yaml` não é criado. `init --force` rejeitado pelo gate: `gira.yaml`, head e P inalterados. `init --output`: versão canônica e P com bytes idênticos, e o resolvedor encontra a versão sem `--index`. `index --output`: store intacto. Sync falho no Unix: manifesto não gravado; nova execução grava. Manifesto alterado por fora: conflito. Reinício com sobras (`.tmp`, órfã) é limpo pela próxima execução. |
| 2i. Documentação | README, `website/docs/cli/*`, skills, bench, tools | Documentação igual ao código; validador verde. |

**Relatório por etapa:** invariantes; testes executados, com comando e resultado literais; resultado do CI; limitações; e o que fica **[DESC]**. Ao fim de cada etapa paro e espero autorização para a seguinte.

## Decisões pendentes

| # | Decisão | Recomendação |
|---|---|---|
| D1 | Semântica de `--output` | Como na seção 1: `index` destacado, `init` canônico com snapshot opcional |
| D2 | Store respeita `XDG_CACHE_HOME` | Sim |
| D3 | Locator do manifesto vira o basename | Sim |
| D4 | `source_key` sobre os bytes crus do manifesto | Sim |
| D5 | Divergência de ferramenta falha fechado, sem flag de exceção | Sim |
| D6 | Tabela da seção 3; `--require-durable` | Tabela adotada; sem `--require-durable` nesta etapa |
| D7 | Sem coleta automática | Sim |
| D8 | Travas M→S→P, falha fechada sem suporte, sem `--no-lock` | Sim |
| D9 | Legado nunca carregado automaticamente com `gira.yaml` | Sim |

## Riscos residuais

- A completude de P1.
- Ausência de garantia contra queda de energia, em especial no Windows.
- Exclusão entre máquinas em sistemas de arquivos de rede.
- Editores externos e hardlinks fora do alcance da trava M.
- `.gira/locks/` sem estar no `.gitignore` do usuário.
- Cada atualização e cada rebuild de desenvolvimento exigem reindexar.
- Disco crescendo sem limite.
- `serve mcp` continua usando o mapa carregado em memória.
- `nudge` engole os erros do resolvedor.
- `DeletedPaths=nil` na CLI, comportamento que já existe hoje.
- Uma linhagem nova começa sem base para o gate.

**Pedido de autorização:** apenas a **Etapa 2.0**, ajustando os testes num branch com PR e CI nos três sistemas, você faz o merge. Pode ser autorizada sozinha. As decisões D1 a D9 vêm antes da 2a.