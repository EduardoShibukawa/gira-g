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

- **SC1 Determinismo de invocação.** O mesmo manifesto dentro de uma raiz Git, indexado por caminho relativo, absoluto, a partir de subdiretório (descoberta), via symlink e via worktree/submódulo, converge para o mesmo `sha256`. A matriz adversarial (cwd, `HOME`/`USERPROFILE`/`XDG_CACHE_HOME`, `TZ`, `LANG`, `core.quotePath`, `refs/replace`, nomes não ASCII e com tab, três execuções frias, dois ambientes isolados) não muda os bytes. Variáveis de ambiente Git que redirecionam descoberta não podem sequestrar `init` nem a raiz do locator. Manifesto fora da raiz Git com overrides falha fechado antes de publicar qualquer artefato; sem overrides, indexação não falha apenas por ausência de raiz.
- **SC2 Leituras Git independentes do ambiente.** Arquivos com nome não ASCII ou com tab indexam com a configuração padrão. `quotePath` e `refs/replace` não mudam os bytes. Clone e fetch continuam usando a configuração do usuário (ADR 0009).
- **SC3 Identidade e proveniência.**
  - Cada contraexemplo da tabela da v1 muda a chave.
  - A codificação é injetiva: `("ab","c") ≠ ("a","bc")`.
  - Um build modificado inclui o hash do executável.
  - Mesmo `version_id` com bytes diferentes gera `nondeterministic_build`, sem sobrescrever nada.
  - O novo mapa usa `schema_version: 2` (JSON inteiro). O snapshot canônico e o destacado incluem um objeto `source_identity` autocontido com `schema_version: 2`, SHA-256 dos bytes exatos do manifesto, versões dos extratores e SHAs de todas as fontes relevantes.
  - O nome do campo não é `provenance`; fontes incluem todos os repositórios declarados no manifesto, ordenados deterministicamente, sem URLs cruas; usar `url_sha256`.
  - A decodificação do schema 2 e de `source_identity` é estrita. Binários antigos que esperam `schema_version` string devem falhar fechado; o leitor novo usa um conjunto explícito de versões, emite `index_schema_unsupported` para tipo/schema desconhecido e não pode descartar campos em `CanonicalJSON` ou `cloneMacroMap`.
  - O legado `schema_version: "1.0.0"` só pode ser lido através de `--index` explícito com aviso; nunca é carregado implicitamente com `gira.yaml` e nunca é reescrito in-place.
- **SC4 Resolução segura.**
  - Os cenários S1 e S2 da v1 retornam `index_missing`, e não `[]`.
  - Versão só de outra ferramenta retorna `index_tool_mismatch`, sem carregar nada.
  - Tipo/schema desconhecido retorna `index_schema_unsupported`, distinto de `index_corrupt`.
  - Publicação recusa downgrade de schema; uma versão legada explicitamente indicada por `--index` é somente para leitura e gera aviso.
  - Com `gira.yaml`, o legado nunca é carregado.
  - `--index` continua a funcionar para snapshots destacados compatíveis.
  - O fallback de eval para `ManifestPath = "gira.yaml"` é removido; se overrides exigem locator e o chamador não fornecer manifesto real, retornar `manifest_path_required`.
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

N1, N2, N3, N4 e D1–D9c e as nove decisões técnicas dos Gates 1–5 foram escolhidas pelo dono em 2026-10-17. `main.date` sai de `tool_id` pela opção C; o hash do executável continua em builds de desenvolvimento não confiáveis. As escolhas estão registradas em [decisão arquitetural](../docs/plans/2026-10-17-decisao-arquitetural-publicacao-versionada.md) e [gates técnicos](../docs/plans/2026-10-17-gates-tecnicos-publicacao-versionada.md). A spec continua `draft`: faltam auditoria de consumidores/schema, testes dos comportamentos escolhidos e verificação de `git archive | tar -x` com tab no Windows. Gate 5 vira critério de aceitação da implementação após autorização heavy; não é pré-condição para iniciar a primeira tarefa autorizada.

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

A spec continua `draft` e implementação heavy não está autorizada. As direções de produto foram decididas; os pontos abaixo são evidências de compatibilidade e aceitação que ainda faltam.

1. **Contrato do mapa e consumidores (decisões B/R2/`source_identity` registradas):** implementar e testar a forma exata do schema 2 e o leitor estrito; garantir que `CanonicalJSON` e `cloneMacroMap` preservam todos os campos (incluindo os campos já perdidos atualmente); auditar CLI, `store.LoadFile/PublishFile`, MCP, skills, bench, bundle, eval/scorecard, goldens, fixtures e documentação. Provar que um binário antigo falha fechado e não faz downgrade do schema.
2. **Manifesto fora da raiz Git (O1) e ambiente Git:** implementar `manifest_outside_git_root` e `git_root_unavailable`; normalizar locator relativo à raiz Git depois de resolver symlinks; sanear as variáveis de redirecionamento em probes de raiz e `init`; executar a matriz raiz/subpasta/worktree/submódulo/symlink/fora do Git e provar que nenhum output é publicado em falha.
3. **Eval e identidade do manifesto:** remover o fallback silencioso `"gira.yaml"`; garantir que o chamador passa o manifesto real e os bytes que alimentam `manifest_sha256`; com overrides sem manifesto real, retornar `manifest_path_required`. Confirmar que as fontes e seus SHAs participam da semântica de staleness como registrado.
4. **`tool_id` (opção C + hash de desenvolvimento):** testes com e sem `-trimpath`; mudar apenas `main.date` deve manter `tool_id`/mapa de release estáveis; mudar `main.version` ou `main.commit` deve mudar `tool_id`; dev builds não confiáveis incluem hash do executável. Investigar/registrar o comportamento `vcs.*` que não apareceu nos builds locais.
5. **Windows antes da aprovação final:** o Gate 4 aceitou a evidência para o perfil do runner hospedado (admin + Developer Mode), não para máquinas de usuários comuns. Adicionar o teste `git archive | tar -x` com nome de arquivo contendo tab no Windows. No SC7, tentar symlink em contexto não elevado/sem Developer Mode e registrar um skip explícito quando o ambiente não permitir criar o link.
6. **Durabilidade e publicação, após autorização heavy:** a primeira tarefa de implementação deve trazer testes injetando falhas antes/depois do rename, fsync de diretório no Unix, `visible_unconfirmed`, publicação parcial de `init --output`, preservação de snapshots válidos/corrompidos e validação de hash nos três SOs. Esses testes precisam passar antes de aceitar/mesclar a implementação, mas não podem ser exigidos antes de começar o primeiro child autorizado.
7. **Autorização:** nenhuma tarefa-filha, ADR 0014 ou mudança de código de publicação versionada até as evidências 1–5 serem documentadas e o dono confirmar explicitamente a autorização heavy. A decisão de hoje escolhe contratos, mas não é essa autorização.

## Revisions

- v1 (2026-10-17): primeiro rascunho no repositório. Consolida a proposta v2 do agente (anexo), os requisitos acrescentados pelo dono (D9b, D9c, `main.date`, não usar `os.UserCacheDir`) e as decisões N1–N4 que o plano de governança encontrou.
- v1.1 (2026-10-17): registra as direções N3, N4 e D1–D9c aprovadas pelo dono e explicita as evidências ainda necessárias antes da aprovação final; N1/N2 permanecem registradas pelo histórico de governança.
- v1.2 (2026-10-17): registra decisões suplementares dos Gates 1–5 (schema inteiro 2, política R2, `source_identity`, O1, eval sem fallback, `tool_id` opção C, hash de dev, W1 e Gate 5 como aceitação pós-implementação); spec segue `draft` e implementação continua não autorizada.
