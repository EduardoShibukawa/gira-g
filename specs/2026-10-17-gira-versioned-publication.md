---
schema: spec/1
status: draft
version: 1
---

# Publicação versionada de índices do Gira-G

Rascunho de especificação. As direções N1–N4 e D1–D9c foram decididas pelo dono em 2026-10-17 e registradas em [decisão arquitetural](../docs/plans/2026-10-17-decisao-arquitetural-publicacao-versionada.md). O status continua `draft`: ainda faltam auditorias de compatibilidade e verificações de plataforma listadas em `## Open questions`. Essas decisões de direção **não autorizam implementação**. A proposta original do agente (v2) está anexada sem alterações em [docs/plans/2026-10-17-gira-spec-v2-proposta-original.md](../docs/plans/2026-10-17-gira-spec-v2-proposta-original.md). A v1, que traz o algoritmo de fingerprint, está em [docs/plans/2026-10-17-gira-spec-v1-proposta-original.md](../docs/plans/2026-10-17-gira-spec-v1-proposta-original.md). O plano de execução fica em [docs/plans/2026-10-17-plano-execucao-governanca.md](../docs/plans/2026-10-17-plano-execucao-governanca.md).

## Problem

Comportamento atual, lido em `61b58c2`:

- O índice publicado é identificado só pelo nome do sistema. Nada liga o mapa ao manifesto nem ao binário que o gerou. Uma sonda de agente (relato, v1 S2) mostrou um mapa desatualizado servido como atual: `locate` retornou `[]` com saída 0.
- `index --output P` e `init --output P` substituem o store. Quando existe `gira.yaml`, o resolvedor nunca lê P. O README diz o contrário.
- `init` grava `gira.yaml` antes do build. Se o build falha, o manifesto fica no disco.
- O store de índices ignora `XDG_CACHE_HOME`, enquanto o cache de repositórios respeita.
- O locator dos overrides guarda o caminho do manifesto tal como foi digitado. Invocar o mesmo manifesto por caminho relativo e por caminho absoluto gera bytes diferentes.
- As leituras do Git herdam a configuração do usuário. `ls-tree` sem `-z` falha com nomes não ASCII sob o `core.quotePath` padrão.
- Não existe trava entre processos.
- No Windows, a durabilidade nunca é confirmada, porque não há sync de diretório.

Afetados: usuários da CLI, hosts MCP configurados por `skills install`, bench, `tools/index-bundle`, a skill `gira-arch-diff` e a documentação.

Gatilhos `heavy`:
- **H1:** as escolhas de produto D1–D9c e N1–N4 estão pendentes.
- **H3:** mudam o contrato de `--output`, o layout do store e os erros do resolvedor, que são consumidos por skills, bench, tools e documentação.
- **H5:** são nove etapas, em várias sessões.

## Expected behavior

Proposta, condicionada às decisões abaixo:

- `gira index` sem `--output` publica uma versão imutável num store versionado. A versão é identificada por `version_id = H(source_key, tool_id)`, e um head por linhagem aponta para ela. O resolvedor só lê o store a partir de um `gira.yaml`.
- `gira index --output P` grava apenas um snapshot destacado em P. Ele não altera o store, os heads nem o manifesto, e a saída avisa que o snapshot não é o índice ativo.
- `gira init` sempre publica no store canônico. Com `--output P`, grava também em P uma cópia byte a byte. O manifesto é gravado por último: um build ou gate que falha não cria nem altera `gira.yaml`.
- A mesma origem e a mesma ferramenta produzem os mesmos bytes, independentemente de cwd, de como o manifesto foi escrito, de `HOME`, `USERPROFILE`, `XDG_CACHE_HOME`, `TZ`, `LANG` e da configuração Git de leitura.
- Republicar o mesmo `version_id` com bytes diferentes gera `nondeterministic_build`.
- O resolvedor nunca apresenta um mapa desatualizado como atual:
  - versão ausente: `index_missing`;
  - versão só de outra ferramenta: `index_tool_mismatch`;
  - versão corrompida: `index_corrupt`.
- Os escritores se serializam pelas travas M (manifesto), S (store) e P (snapshot), adquiridas nessa ordem global:
  - timeout: `index_locked`;
  - trava sem suporte ou com erro: `lock_unavailable`, sem gravar nada.
- A publicação informa `confirmed` ou `visible_unconfirmed`. Um código diferente de zero por sync não confirmado **não** significa que nada foi publicado.

## Scope

**Dentro do escopo:** as etapas 2a–2i do plano: ADR 0014, leituras Git independentes do ambiente, locator normalizado, chaves, travas, publicação canônica e destacada, resolvedor, CLI, documentação e consumidores.

**Fora do escopo:**
- coleta automática de versões;
- `--require-durable` e `--no-lock`;
- exclusão entre máquinas em sistemas de arquivos de rede;
- nome canônico do arquivo de índice;
- os itens 2–5 do plano de evolução aprovado (atualidade, incremental, doctor/eval, robustez da CLI).

**Não objetivos:**
- garantia contra queda de energia;
- prova formal de determinismo.

## Acceptance criteria

- **SC1 Determinismo de invocação.** O mesmo manifesto indexado por caminho relativo, por caminho absoluto, a partir de um subdiretório (descoberta) e via `init` produz um mapa com o mesmo `sha256`. A matriz adversarial (cwd, `HOME`/`USERPROFILE`/`XDG_CACHE_HOME`, `TZ`, `LANG`, `core.quotePath`, `refs/replace`, nomes não ASCII e com tab, três execuções frias, dois ambientes isolados) não muda os bytes.
- **SC2 Leituras Git independentes do ambiente.** Arquivos com nome não ASCII ou com tab indexam com a configuração padrão. `quotePath` e `refs/replace` não mudam os bytes. Clone e fetch continuam usando a configuração do usuário (ADR 0009).
- **SC3 Identidade e proveniência.**
  - Cada contraexemplo da tabela da v1 muda a chave.
  - A codificação é injetiva: `("ab","c") ≠ ("a","bc")`.
  - Um build modificado inclui o hash do executável.
  - Mesmo `version_id` com bytes diferentes gera `nondeterministic_build`, sem sobrescrever nada.
  - Cada snapshot canônico ou destacado carrega metadados autocontidos: hash SHA-256 dos bytes exatos do manifesto, `schema_version`, versões dos extratores e SHAs de todas as fontes relevantes. Consumidores antigos não podem ignorar silenciosamente um schema incompatível.
- **SC4 Resolução segura.**
  - Os cenários S1 e S2 da v1 retornam `index_missing`, e não `[]`.
  - Versão só de outra ferramenta retorna `index_tool_mismatch`, sem carregar nada.
  - Com `gira.yaml`, o legado nunca é carregado.
  - `--index` continua igual.
- **SC5 Publicação.**
  - Falha antes do rename: nada fica visível.
  - Falha depois do rename da versão: a versão fica visível, o head fica intacto e uma nova execução é idempotente.
  - Gate rejeita: nada é gravado.
  - Head corrompido ou apontando para versão ausente: falha fechada.
- **SC6 Durabilidade.**
  - Unix: um sync que falha gera `visible_unconfirmed`, saída diferente de zero e, no `init`, o manifesto não é gravado; uma nova execução grava.
  - Windows: `visible_unconfirmed`, saída 0 e aviso.
- **SC7 Travas.** Nos três sistemas:
  - dois processos se excluem;
  - grafias relativa, absoluta e via symlink do mesmo manifesto se excluem;
  - um processo filho morto libera a trava;
  - timeout gera `index_locked`;
  - um erro injetado gera `lock_unavailable` sem nada gravado;
  - a ordem M→S→P é verificada.
- **SC8 Fluxos da CLI.**
  - Primeira inicialização com build falhando: `gira.yaml` não é criado.
  - `init --force` rejeitado pelo gate: manifesto, head e P ficam inalterados.
  - `init --output P`: a versão canônica e P têm bytes idênticos, e o resolvedor encontra a versão sem `--index`.
  - `index --output P`: o store fica intacto.
  - Manifesto alterado por fora durante o `init`: conflito.
- **SC9 Documentação e consumidores.**
  - O README, a documentação da CLI, as skills, o bench e as tools descrevem o comportamento implementado.
  - `gira-arch-diff`, `bench/phases.json`, `tools/index-bundle` e o benchtest continuam funcionando.
  - O validador e o build da documentação passam.
- **SC10 Regressão.** `go test -race ./...` passa em Ubuntu, macOS e Windows; o floor guard passa; nenhum teste escreve no cache real do usuário.

## Interfaces and dependencies

- **Contrato da CLI:** a semântica de `--output` em `index` e `init` (D1). `--output` de `export`, `view` e `skills manifest` não muda.
- **Localização do store** (D2): `$XDG_CACHE_HOME/gira/indexes`, senão `~/.cache/gira/indexes`, senão `<raiz do projeto>/.gira/indexes`.
- **Layout do store:** versões imutáveis com `meta.json` estrito e heads por linhagem.
- **Erros novos:** `index_tool_mismatch`, `nondeterministic_build`, `index_locked`, `lock_unavailable`. Possivelmente `index_ambiguous` (D9b). Estado `visible_unconfirmed`.
- **Locators de override no mapa** (D3): os bytes de mapas que têm overrides mudam.
- **Consumidores:**
  - `gira-arch-diff`;
  - `bench/phases.json`;
  - `internal/benchtest`;
  - `tools/index-bundle/bootstrap.sh`;
  - `skills install`/`check` com `--index` (D9c);
  - MCP `serve`;
  - README e `website/docs` (incluindo `versioned_docs`).
- **ADRs:**
  - ADR 0001 (locators);
  - ADR 0007 (determinismo de 100%);
  - ADR 0009 (configuração Git do usuário para clone e fetch);
  - ADR 0014, a ser criado.
- **Dependência:** `golang.org/x/sys` passa de indireta a direta.

## Decisions

N1, N2, N3, N4 e D1–D9c estão decididas como direções pelo dono em 2026-10-17, com a subdecisão `main.date` em D5 ainda aberta e as condições listadas no registro arquitetural. A spec permanece `draft` até os bloqueios de aprovação serem resolvidos; decisões de direção não autorizam implementação.

| ID | Questão | Alternativas | Decisão do dono / condição |
|---|---|---|---|
| N1 | Nome e sequência da etapa: o item 2 aprovado é "atualidade e sync multi-repo" | continuação do item 1 com rótulo novo; renumerar; fazer primeiro o item 2 | **decidido: continuação do item 1** com rótulo próprio (ex.: "Etapa 1.2"), por dono, 2026-10-17 (mensagem de decisão da sessão). Relabel dos documentos é follow-up separado |
| N2 | Aceite da Etapa 1/1.1 em `61b58c2` antes da 2a | aceitar; aceitar depois de uma revisão adversarial de `aa4e6fe..0d5f248`; não aceitar | **decidido: (b)** por dono, 2026-10-17 (mensagem de decisão da sessão): revisão adversarial de `4c2204c` e `0d5f248` (exclui `aa4e6fe`), depois aceitar se não houver bloqueio material |
| N3 | O snapshot destacado carrega identidade? (decisão aprovada nas Sources: snapshots identificam hash do manifesto, schema, extratores e SHAs) | sidecar `P.meta.json`; meta embutido; sem identidade | **Decidido: metadados embutidos** no snapshot autocontido. Incluir hash SHA-256 do manifesto, versão de schema, versões dos extratores e SHAs de todas as fontes relevantes, inclusive quando não houver `Locator.CommitSHA`. Exige bump explícito do schema e auditoria de consumidores antes da aprovação final. |
| N4 | Escopo do ADR 0014 em relação a "Git faz o versionamento" | store = cache derivado, local e regenerável | **Decidido: sim.** O store é cache derivado local, regenerável e não compartilhado; Git permanece a fonte de versionamento compartilhado. |
| D1 | Semântica de `--output` | manter; destacada (A′); cópia mais store; diretório-raiz; remover | **Decidido: A′, condicionado a N3.** `index --output P` grava apenas o snapshot destacado; `init --output P` publica o estado canônico e grava em P uma cópia byte a byte do snapshot autocontido. Ordem canônica: versão → snapshot → manifesto. P corrompido é preservado salvo `--recover-corrupt`. Se a cópia falhar depois da publicação canônica, reportar sucesso parcial e o estado canônico. |
| D2 | Localização do store | XDG → `~/.cache` → projeto; `os.UserCacheDir`; manter | **Decidido:** resolver único para leitura/escrita com XDG → `~/.cache` → fallback no projeto. A base do gate também consulta `~/.cache/gira/indexes/<sys>` quando XDG está definido, para compatibilidade legada. |
| D3 | Locator do manifesto | basename; relativo à raiz Git; manter | **Direção decidida:** usar caminho relativo à raiz Git e normalizado com `/`, para evitar caminho absoluto dependente do clone e manter identidade de subpastas. **Bloqueio:** definir comportamento para manifesto fora da raiz Git e auditar ADR 0001 e consumidores antes da aprovação final; não usar fallback absoluto silencioso. |
| D4 | `source_key` | bytes crus; YAML canônico; normalização de EOL/BOM | **Decidido:** bytes crus do manifesto. Comentários, BOM e EOL diferentes podem exigir reindexação; chave e parser devem observar os mesmos bytes. |
| D5 | Ferramenta divergente | falhar fechado; servir com aviso (v1, **revogada**) | **Decidido:** falhar fechado com `index_tool_mismatch`, sem flag de exceção. **Aberto:** auditar se `main.date` altera a semântica da extração. Excluir do `tool_id` apenas se for informativo; manter se alterar semântica, com justificativa. |
| D6 | Durabilidade e manifesto | tabela da v2; criar `--require-durable` | **Decidido:** tabela da v2, sem `--require-durable`. Unix: falha de sync pode deixar a versão visível, retorna erro e impede gravar o manifesto em `init`; Windows: `visible_unconfirmed`, saída 0 e aviso explícito, sem promessa contra perda de energia. Requer testes por plataforma. |
| D7 | Retenção | sem GC; K=3 (v1, **revogada**) | **Decidido:** sem GC; limpeza manual documentada. Quarentena só em `--recover-corrupt`, sem remoção automática de versões. |
| D8 | Travas | M→S→P com `flock`/`LockFileEx`, 60 s, falha fechada | **Decidido:** M→S→P, `flock`/`LockFileEx`, timeout de 60 s, `index_locked`/`lock_unavailable` com falha fechada e sem `--no-lock`. Documentar limitações de FS de rede e editores externos. |
| D9 | Legado | nunca com `gira.yaml`; leitura transitória (v1, **revogada**) | **Decidido:** legado nunca é carregado automaticamente com `gira.yaml`; pode servir de base estrita do gate na primeira publicação canônica, com mensagem explícita. |
| D9b | Descoberta sem manifesto | exigir `tool_id` atual; não exigir | **Decidido:** exigir `tool_id` atual e falhar fechado na ambiguidade com código estável `index_ambiguous`, preservando mensagem legível. |
| D9c | `skills install --index` aponta para legado | aviso; aviso e caminho absoluto; rejeitar | **Decidido:** avisar no caminho legado e normalizar o caminho para absoluto. Cobrir com testes para caminho legado, relativo e snapshot destacado. |

## Risks and constraints

- **Determinismo:** a garantia de fingerprint é condicional a quatro premissas:
  - P1: build determinístico com entradas completas;
  - P2: o SHA do Git fixa o conteúdo;
  - P3: `tool_id` injetivo;
  - P4: codificação injetiva.

  Uma entrada oculta não enumerada quebra a garantia sem ser detectada até alguém republicar o mesmo id.
- **Durabilidade:** não há garantia contra queda de energia, sobretudo no Windows. O que se garante é a detecção na leitura.
- **Travas:**
  - sem exclusão entre máquinas (NFS, FUSE, SMB);
  - editores externos e hardlinks ficam fora da trava M;
  - resta uma janela entre a comparação de hash do manifesto e o rename;
  - `.gira/locks/` aparece no projeto do usuário se `.gira/` não estiver no `.gitignore`.
- **Compatibilidade:**
  - reindexar a cada atualização do gira e a cada rebuild de desenvolvimento;
  - disco crescendo sem limite (o mapa do próprio gira-g tem cerca de 4,9 MB por versão);
  - binários antigos continuam gravando o caminho legado sem trava.
- **Herdados:**
  - `serve mcp` mantém o mapa em memória;
  - `nudge` engole os erros do resolvedor;
  - o gate na CLI recebe `DeletedPaths=nil`;
  - uma linhagem nova começa sem base para o gate.
- **Ações protegidas:** nenhuma; não há release nem publicação nesta etapa.

## Verification

- **Regressão integrada:** `go test -race ./...` em Ubuntu, macOS e Windows, mais os jobs validator, floor guard e gate-lint do CI, no head integrado.
- **Fim a fim:** SC1 e SC2 por matriz adversarial; SC3 com testes de chave e fuzz de codificação; SC4–SC8 com testes de integração da CLI e do store; SC9 com validador e build da documentação; SC10 com CI e hash da listagem do cache real antes e depois.
- **Foco da revisão:**
  - completude das entradas determinantes;
  - janelas de crash entre rename, sync, head e manifesto;
  - ordem das travas;
  - mensagens de erro estáveis;
  - consumidores de `--output`.

## Rollback

Os mapas e o store versionado são caches derivados. Voltar o binário restaura o comportamento anterior: o binário antigo lê e grava o caminho legado. Isso tem dois efeitos:
- versões gravadas pelo binário novo ficam ignoradas;
- o caminho legado fica com o último mapa que o binário antigo gravou, então é preciso reindexar depois do rollback.

Nenhuma etapa apaga o caminho legado. Mudanças de bytes por D3 e D4 se revertem junto com o binário.

## Open questions

A spec continua `draft`; os pontos abaixo bloqueiam sua aprovação e qualquer implementação. As direções da tabela de decisões foram escolhidas, mas a evidência técnica ainda precisa ser registrada.

1. **Contrato do mapa e consumidores:** auditar todos os consumidores de `MacroMap`/JSON e decidir o valor de `schema_version` depois de identificar o contrato atual. Demonstrar que o snapshot autocontido e seus metadados são preservados por CLI, `gira-arch-diff`, bench, tools, skills, exemplos, fixtures, testes e documentação.
2. **D3 fora da raiz Git:** definir comportamento explícito para manifesto fora da raiz do repositório; nunca gravar caminho absoluto dependente da máquina como fallback silencioso. Auditar compatibilidade com ADR 0001 e consumers de locators.
3. **D5 `main.date`:** verificar se o valor afeta a extração semântica. A subdecisão fica aberta até haver evidência.
4. **Windows:** executar testes reprodutíveis no runner Windows para symlinks e `os.DevNull` usado como `GIT_CONFIG_GLOBAL`; hoje o comportamento é desconhecido.
5. **Publicação/durabilidade:** testes por plataforma para falhas antes/depois do rename, fsync de diretório, `visible_unconfirmed`, publicação parcial de `init --output`, preservação de artefatos válidos/corrompidos e validação de hash.
6. **Autorização:** não iniciar ADR 0014, tarefas-filhas ou implementação heavy até os pontos 1–5 estarem evidenciados e o dono confirmar explicitamente a autorização de implementação.

## Revisions

- v1 (2026-10-17): primeiro rascunho no repositório. Consolida a proposta v2 do agente (anexo), os requisitos acrescentados pelo dono (D9b, D9c, `main.date`, não usar `os.UserCacheDir`) e as decisões N1–N4 que o plano de governança encontrou.
- v1.1 (2026-10-17): registra as direções N3, N4 e D1–D9c aprovadas pelo dono e explicita as evidências ainda necessárias antes da aprovação final; N1/N2 permanecem registradas pelo histórico de governança.
