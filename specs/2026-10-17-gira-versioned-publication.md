---
schema: spec/1
status: draft
version: 1
---

# Publicação versionada de índices do Gira-G

Rascunho para revisão do dono. Nenhuma decisão deste documento está aprovada. As recomendações vêm de uma proposta de agente e do plano de governança, e cada uma está marcada como **proposta**. A proposta original do agente (v2) está anexada sem alterações em [docs/plans/2026-10-17-gira-spec-v2-proposta-original.md](../docs/plans/2026-10-17-gira-spec-v2-proposta-original.md). A v1, que traz o algoritmo de fingerprint, está em [docs/plans/2026-10-17-gira-spec-v1-proposta-original.md](../docs/plans/2026-10-17-gira-spec-v1-proposta-original.md). O plano de execução fica em [docs/plans/2026-10-17-plano-execucao-governanca.md](../docs/plans/2026-10-17-plano-execucao-governanca.md).

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
- **SC3 Identidade.**
  - Cada contraexemplo da tabela da v1 muda a chave.
  - A codificação é injetiva: `("ab","c") ≠ ("a","bc")`.
  - Um build modificado inclui o hash do executável.
  - Mesmo `version_id` com bytes diferentes gera `nondeterministic_build`, sem sobrescrever nada.
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

N1 e N2 estão **decididas** (dono, 2026-10-17); as demais estão **abertas**. A recomendação é uma proposta.

| ID | Questão | Alternativas | Proposta |
|---|---|---|---|
| N1 | Nome e sequência da etapa: o item 2 aprovado é "atualidade e sync multi-repo" | continuação do item 1 com rótulo novo; renumerar; fazer primeiro o item 2 | **decidido: continuação do item 1** com rótulo próprio (ex.: "Etapa 1.2"), por dono, 2026-10-17 (mensagem de decisão da sessão). Relabel dos documentos é follow-up separado |
| N2 | Aceite da Etapa 1/1.1 em `61b58c2` antes da 2a | aceitar; aceitar depois de uma revisão adversarial de `aa4e6fe..0d5f248`; não aceitar | **decidido: (b)** por dono, 2026-10-17 (mensagem de decisão da sessão): revisão adversarial de `4c2204c` e `0d5f248` (exclui `aa4e6fe`), depois aceitar se não houver bloqueio material |
| N3 | O snapshot destacado carrega identidade? (decisão aprovada nas Sources: snapshots identificam hash do manifesto, schema, extratores e SHAs) | sidecar `P.meta.json`; meta embutido; sem identidade | sidecar |
| N4 | Escopo do ADR 0014 em relação a "Git faz o versionamento" | store = cache derivado, local e regenerável | sim |
| D1 | Semântica de `--output` | manter; destacada (A′); cópia mais store; diretório-raiz; remover | A′: `index --output` destacado; `init` sempre canônico com cópia; ordem versão → snapshot → manifesto; P sobrescrito só se ausente ou válido, senão `--recover-corrupt` |
| D2 | Localização do store | XDG → `~/.cache` → projeto; `os.UserCacheDir`; manter | XDG → `~/.cache` → projeto, numa função única. A base do gate também procura o caminho legado `~/.cache` quando XDG está definido |
| D3 | Locator do manifesto | basename; relativo à raiz Git; manter | `Path = basename` com barra `/`. Para esse locator, "relativo à pasta do manifesto" produz o mesmo valor |
| D4 | `source_key` | bytes crus; YAML canônico; normalização de EOL/BOM | bytes crus. Comentário, BOM e EOL exigem reindexar |
| D5 | Ferramenta divergente | falhar fechado; servir com aviso (v1, **revogada**) | falhar fechado, sem flag. Sub-decisão: excluir `main.date` do `tool_id` |
| D6 | Durabilidade e manifesto | tabela da v2; criar `--require-durable` | tabela da v2, sem `--require-durable` |
| D7 | Retenção | sem GC; K=3 (v1, **revogada**) | sem GC. Quarentena só em `--recover-corrupt`; procedimento manual documentado |
| D8 | Travas | M→S→P com `flock`/`LockFileEx`, 60 s, falha fechada | como proposto, sem `--no-lock` |
| D9 | Legado | nunca com `gira.yaml`; leitura transitória (v1, **revogada**) | nunca com `gira.yaml`. Serve como base do gate na primeira publicação |
| D9b | Descoberta sem manifesto | exigir `tool_id` atual; não exigir | exigir. Sub-decisão: introduzir `index_ambiguous` |
| D9c | `skills install --index` aponta para legado | aviso; aviso e caminho absoluto; rejeitar | aviso e caminho absoluto |

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

Todas bloqueiam a aprovação. Nenhuma foi adiada.

- N1, N2, N3, N4, D1, D2, D3, D4, D5 (inclusive `main.date`), D6, D7, D8, D9, D9b (inclusive `index_ambiguous`) e D9c, conforme a tabela de decisões.
- A spec precisa de critérios mais finos para o symlink no runner Windows (exige privilégio) e para `os.DevNull` como `GIT_CONFIG_GLOBAL` no Windows. Hoje os dois são desconhecidos.

## Revisions

- v1 (2026-10-17): primeiro rascunho no repositório. Consolida a proposta v2 do agente (anexo), os requisitos acrescentados pelo dono (D9b, D9c, `main.date`, não usar `os.UserCacheDir`) e as decisões N1–N4 que o plano de governança encontrou.
