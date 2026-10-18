# Decisões de arquitetura para a publicação versionada do Gira-G

| Campo | Valor |
|---|---|
| Data | 2026-10-17 |
| Tipo | Registro de decisões do dono, somente documentação |
| Estado | Direções e spec aprovadas; implementação heavy autorizada pelo dono em 2026-10-18 |
| Base | `731cf4ba4fc10ad8d03e4571696975251ba8940c` |
| Spec | [`2026-10-17-gira-versioned-publication.md`](../../specs/2026-10-17-gira-versioned-publication.md) |
| Tarefa-mãe | [`2026-10-17-gira-versioned-publication.md`](../../tasks/2026-10-17-gira-versioned-publication.md) |

## Limite desta decisão

Este registro escolhe direções de produto para fechar as alternativas N3, N4 e D1–D9c. Não aprova a spec v1, não registra `confirmed: heavy`, não ativa as etapas 2a–2i e não autoriza alterações de código. O trabalho de implementação continua bloqueado até os pré-requisitos em `## Bloqueios antes da aprovação final` serem verificados e documentados.

## Decisões do dono

| ID | Decisão | Condição ou consequência |
|---|---|---|
| N3 | **Metadados de proveniência embutidos no snapshot**, não em sidecar e não ausentes. | O snapshot destacado precisa ser autocontido e identificar hash SHA-256 do manifesto, versão do schema, versões dos extratores e SHAs de todas as fontes relevantes, não apenas fontes presentes em `Locator.CommitSHA`. Exige uma evolução explícita do schema e auditoria de todos os consumidores antes de aprovar a spec. |
| N4 | O store é um **cache derivado local, regenerável e não compartilhado**. | Git continua sendo o mecanismo de versionamento de conhecimento compartilhado: manifestos e snapshots exportados são versionados pelo Git. O store não vira fonte de verdade nem substituto de Git. |
| D1 | Aprovar a semântica **A′**, condicionada ao snapshot autocontido de N3. | `index --output P` produz apenas um snapshot destacado e não altera o store canônico nem o manifesto. `init --output P` publica a versão/snapshot canônicos e grava em P uma cópia byte a byte do mesmo snapshot autocontido. Ordem canônica: versão, snapshot, manifesto por último. Um P válido pode ser substituído; um P corrompido deve ser preservado, exceto com `--recover-corrupt`. Se a cópia P falhar depois da publicação canônica, a mensagem deve declarar o sucesso parcial e identificar o estado canônico, nunca afirmar que nada foi publicado. |
| D2 | Um resolver único compartilhado por leitura e escrita usa **XDG_CACHE_HOME → ~/.cache → fallback no projeto**. | Quando XDG estiver definido, a busca da base do gate ainda precisa consultar o caminho legado `~/.cache/gira/indexes/<sys>` para proteger a primeira publicação. Testar precedência e isolamento de `XDG_CACHE_HOME`. |
| D3 | `Locator.Path` relativo à raiz Git, normalizado com `/`; política O1. | Manifesto com overrides fora da raiz falha com `manifest_outside_git_root`; falhas de descoberta Git com `git_root_unavailable`. Saneamento das variáveis de ambiente Git em probes e em `init`; auditoria/testes de locators continuam obrigatórios. |
| D4 | `source_key` é calculada sobre os **bytes crus do manifesto**. | Comentários, espaços, BOM e finais de linha diferentes podem produzir chave diferente e exigir reindexação. A implementação deve parsear exatamente os bytes representados pela chave; documentar e testar a regra. |
| D5 | Divergência de `tool_id` falha fechado com `index_tool_mismatch`, sem flag de exceção; opção C para a identidade de build. | Excluir a setting `-ldflags` inteira de `tool_id`, mantendo `main.version` e `main.commit` explícitos. Manter hash do executável em dev builds não confiáveis; futuros flags semânticos viram campos explícitos e exigem testes. |
| D6 | Adotar a tabela de durabilidade da proposta v2, sem `--require-durable`. | Em Unix, falha de sync após publicação visível deve retornar erro e impedir a gravação do manifesto quando o protocolo exigir confirmação. Em Windows, publicar com aviso explícito de que durabilidade contra queda de energia não é garantida, sem alegar garantia inexistente. `visible_unconfirmed` deve distinguir visibilidade de durabilidade. Testes específicos por sistema operacional são obrigatórios. |
| D7 | **Sem coleta automática de versões**. | Limpeza manual documentada; quarentena só por ação explícita `--recover-corrupt`. Não apagar versões referenciadas por heads ou manifestos conhecidos; documentar desligamento do Gira-G durante limpeza. |
| D8 | Travas em ordem **M → S → P**, `flock` em Unix e `LockFileEx` em Windows, timeout de 60 segundos, falha fechada e sem `--no-lock`. | Erros estáveis `index_locked` e `lock_unavailable`. Documentar limitações de NFS/FUSE/SMB, editores externos, hardlinks e a janela entre comparar hash do manifesto e renomear. |
| D9 | Com `gira.yaml`, nunca carregar automaticamente o índice legado. | Um índice legado pode servir como base estrita do gate na primeira publicação canônica, com mensagem explícita da base usada. Sem manifesto, caminhos legados continuam somente nos modos explícitos já definidos. |
| D9b | Descoberta sem manifesto exige o `tool_id` atual e falha fechado quando ambígua. | Introduzir o código de erro estável `index_ambiguous`, preservando mensagem humana descritiva. |
| D9c | `skills install --index` apontando para formato legado emite aviso e normaliza o caminho para absoluto. | Testar caminho legado, caminho relativo normalizado e snapshot destacado que não deve gerar o aviso. |

## Decisões suplementares dos gates técnicos (2026-10-17)

Estas decisões complementam N3/N4 e D1–D9c. São escolhas do dono, não autorização para implementar.

| ID | Decisão | Limite |
|---|---|---|
| G1-Schema | Novo `schema_version` é o inteiro `2`, opção B, para que binários antigos que esperam string rejeitem o novo mapa em vez de ignorar proveniência. | É uma geração principal de schema, não SemVer textual. O leitor novo precisa suportar explicitamente o legado `"1.0.0"` apenas por `--index` explícito, com aviso, e schema 2 com decodificação estrita. Erros de tipo/schema desconhecido usam `index_schema_unsupported`; corrupção permanece distinta. |
| G1-Reader | Política R2: conjunto de schemas explicitamente aceitos; legado `"1.0.0"` somente por leitura explícita, com aviso; nunca carregar legado implicitamente com `gira.yaml`; não reescrever v1 in-place. | Publicação deve recusar downgrade de schema. |
| G1-Identity | Campo embutido `source_identity`, com `schema_version: 2`, `manifest_sha256`, `extractors` e `sources`. | Incluir todas as fontes do manifesto, ordenadas deterministicamente; usar `url_sha256`, nunca URL crua. Não incluir `tool_id` nem hash do executável no objeto de proveniência exportado. SHAs das fontes participam da avaliação de staleness separada dos locators. |
| G2-Root | Escolher O1: manifesto com overrides fora da raiz Git falha fechado com `manifest_outside_git_root`; falhas de Git discovery fecham com `git_root_unavailable`. | Sem overrides não há locator de override, logo a ausência de raiz Git sozinha não bloqueia indexação. Saneamento do ambiente Git nos probes de raiz/`init`; caminhos relativos à raiz resolvidos após symlink, barras `/`. |
| G2-Eval | Remover o fallback silencioso `ManifestPath = "gira.yaml"` em eval. | O chamador passa caminho/bytes reais; se houver overrides e não houver manifesto real, retornar `manifest_path_required`. Sem overrides, não fabricar locator. |
| G3-ToolID | Escolher C: excluir `-ldflags` do `tool_id`, mantendo `main.version`, `main.commit`, GoVersion e dependências como campos explícitos. | Novos flags que alterem semântica precisam virar campos explícitos de identidade e testes; não depender da string crua. |
| G3-Dev | Manter hash do executável em builds de desenvolvimento quando `vcs.modified` ou revisão/commit confiável estiver ausente. | Aceita-se reindexação em rebuilds de desenvolvimento para não colapsar executáveis semanticamente diferentes. |
| G4-Windows | Aceitar W1 para o perfil testado do runner hospedado (admin elevado, Developer Mode e symlink config de sistema). | Não generalizar para usuário comum. No SC7, tentar symlink e pular só a célula que não puder ser criada, com motivo. O teste `git archive | tar -x` com tab permanece evidência necessária antes da aprovação final da spec. |
| G5-Durability | Gate 5 será verificado como critério de aceitação da primeira tarefa de implementação autorizada, não com protótipo descartável antes da autorização. | Testes de falha/rename/fsync/`visible_unconfirmed`/publicação parcial/hash em Linux, macOS e Windows são obrigatórios antes de aceitar/mesclar a implementação. |

## Bloqueios antes da aprovação final da spec e da implementação

1. **Contrato do mapa e consumidores:** implementar e auditar o schema 2 inteiro, incluindo a forma do campo `source_identity`; preservar todos os campos em `CanonicalJSON` e `cloneMacroMap`; rejeitar schemas incompatíveis com erro estável; atualizar readers/writers, MCP, CLI, bench, skills, goldens, fixtures e docs; testar contra downgrade por binário antigo.
2. **Fontes fora da raiz Git:** implementar O1 e os erros `manifest_outside_git_root`/`git_root_unavailable`; testar matrix relativa/absoluta/descoberta/subpasta/symlink/worktree/submódulo; provar que `GIT_DIR`/variáveis semelhantes não sequestram `init` nem locator.
3. **Identidade e eval:** implementar `source_identity`, política R2 e o fim do fallback de eval; testar `manifest_path_required`, bytes crus do manifesto, fontes sem evidência, ordenação e SHA/URL sem credenciais. Excluir `-ldflags` de `tool_id`, manter hash de exe em dev conforme decidido, testar builds com/sem `-trimpath`.
4. **Windows:** testar SC7 em usuário não elevado/sem Developer Mode quando possível, com skip explícito de symlink se não puder criá-lo; testar `git archive | tar -x` com nome contendo tab no Windows antes da aprovação final da spec. Registrar as limitações do runner atual.
5. **Gate 5, pós-autorização:** testes de falhas antes/depois do rename, fsync de diretório Unix, `visible_unconfirmed`, publicação parcial de `init --output`, preservação de artefatos e validação de hash em Linux/macOS/Windows. São critérios obrigatórios de aceitação da implementação.
6. **Gate final (atualizado em 2026-10-18):** o dono aprovou a spec e autorizou explicitamente heavy. As evidências 1–4 são obrigatórias nas tarefas-filhas e Gate 5 é obrigatório para aceitar a implementação. A autorização permite começar; não dispensa nem pré-aprova os testes.

Status da evidência (2026-10-17): decisões dos gates 1–5 registradas no documento [gates técnicos](2026-10-17-gates-tecnicos-publicacao-versionada.md). A auditoria de consumidores e as provas por plataforma ainda não foram concluídas; Gate 5 será testado após autorização heavy, antes do aceite de implementação.

## Auditoria de estado

- N1 e N2 já foram registradas em PR #19, agora mesclada.
- O aceite formal da Etapa 1/1.1 em `61b58c2` foi registrado depois de revisão adversarial sem bloqueios materiais; os nove achados residuais seguem como follow-ups.
- Nenhum código da publicação versionada foi alterado nesta decisão.
- Nenhuma worktree, stash, branch ou cache local foi removido ou reescrito.

## Changelog

- 1.0 (2026-10-17): direções arquiteturais iniciais registradas; implementação bloqueada por auditorias de compatibilidade e verificações de plataforma.
- 1.1 (2026-10-17): registra as nove decisões dos gates técnicos; spec continua `draft`, tarefa-mãe `todo` e implementação heavy não autorizada.


## Owner authorization update (2026-10-18)

- spec: approved
- implementation: confirmed heavy by owner
- child tasks: created in dependency order in the governance PR
- remaining evidence: acceptance gates only; no test is considered passed by this authorization
