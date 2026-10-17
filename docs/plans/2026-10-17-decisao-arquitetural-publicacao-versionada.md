# Decisões de arquitetura para a publicação versionada do Gira-G

| Campo | Valor |
|---|---|
| Data | 2026-10-17 |
| Tipo | Registro de decisões do dono, somente documentação |
| Estado | Direções aprovadas; spec permanece `draft`; implementação não autorizada |
| Base | `db5a3cfa1675daf42f81699990875eb66f8982ca` |
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
| D3 | Direção escolhida: `Locator.Path` relativo à raiz Git, normalizado com barras `/`, em vez de caminho absoluto dependente do clone. | A política para manifesto fora da raiz Git permanece bloqueadora: não serializar silenciosamente um caminho absoluto. A spec deve definir explicitamente a representação ou uma falha fechada antes da aprovação final. Auditar ADR 0001 e consumidores de locators. |
| D4 | `source_key` é calculada sobre os **bytes crus do manifesto**. | Comentários, espaços, BOM e finais de linha diferentes podem produzir chave diferente e exigir reindexação. A implementação deve parsear exatamente os bytes representados pela chave; documentar e testar a regra. |
| D5 | Divergência de `tool_id` falha fechado com `index_tool_mismatch`, sem flag de exceção. | A subdecisão sobre `main.date` continua aberta até uma auditoria demonstrar se esse valor altera o comportamento semântico dos extratores. Se for apenas informativo, excluir de `tool_id`; se alterar semântica, manter com justificativa. Não escolher por suposição. |
| D6 | Adotar a tabela de durabilidade da proposta v2, sem `--require-durable`. | Em Unix, falha de sync após publicação visível deve retornar erro e impedir a gravação do manifesto quando o protocolo exigir confirmação. Em Windows, publicar com aviso explícito de que durabilidade contra queda de energia não é garantida, sem alegar garantia inexistente. `visible_unconfirmed` deve distinguir visibilidade de durabilidade. Testes específicos por sistema operacional são obrigatórios. |
| D7 | **Sem coleta automática de versões**. | Limpeza manual documentada; quarentena só por ação explícita `--recover-corrupt`. Não apagar versões referenciadas por heads ou manifestos conhecidos; documentar desligamento do Gira-G durante limpeza. |
| D8 | Travas em ordem **M → S → P**, `flock` em Unix e `LockFileEx` em Windows, timeout de 60 segundos, falha fechada e sem `--no-lock`. | Erros estáveis `index_locked` e `lock_unavailable`. Documentar limitações de NFS/FUSE/SMB, editores externos, hardlinks e a janela entre comparar hash do manifesto e renomear. |
| D9 | Com `gira.yaml`, nunca carregar automaticamente o índice legado. | Um índice legado pode servir como base estrita do gate na primeira publicação canônica, com mensagem explícita da base usada. Sem manifesto, caminhos legados continuam somente nos modos explícitos já definidos. |
| D9b | Descoberta sem manifesto exige o `tool_id` atual e falha fechado quando ambígua. | Introduzir o código de erro estável `index_ambiguous`, preservando mensagem humana descritiva. |
| D9c | `skills install --index` apontando para formato legado emite aviso e normaliza o caminho para absoluto. | Testar caminho legado, caminho relativo normalizado e snapshot destacado que não deve gerar o aviso. |

## Bloqueios antes da aprovação final da spec e da implementação

1. **Contrato do mapa e consumidores:** inventariar e testar todos os consumidores de `MacroMap` e do JSON, incluindo CLI, `gira-arch-diff`, bench, tools, skills, exemplos, fixtures, testes e documentação. Definir o novo valor de `schema_version`, migração/compatibilidade e comportamento de parsers antigos. Demonstrar que nenhum consumidor perde os campos de proveniência.
2. **Fontes fora da raiz Git:** fechar a política de D3 e os testes para manifests dentro e fora de repositórios, além de invocação relativa, absoluta e descoberta por subdiretório. Não usar caminho absoluto dependente da máquina como fallback silencioso.
3. **`main.date` em `tool_id`:** auditar build info e pipeline de extração para decidir se a data afeta conteúdo semântico ou só metadados. Registrar evidência da decisão.
4. **Windows no CI:** executar experiências reproduzíveis no runner Windows para symlinks e `os.DevNull` como `GIT_CONFIG_GLOBAL`. Até isso ser verificado, registrar o comportamento como desconhecido.
5. **Durabilidade e publicação:** testar falhas antes e depois do rename, fsync de diretório no Unix, semântica `visible_unconfirmed`, publicação parcial de `init --output`, preservação do arquivo anterior e validação de hash em todos os três sistemas.
6. **Gate final:** a spec continua `draft`, a tarefa-mãe continua `todo` e nenhum child de implementação deve começar até que os cinco pontos anteriores tenham evidência registrada e o dono faça uma autorização heavy explícita.

## Auditoria de estado

- N1 e N2 já foram registradas em PR #19, agora mesclada.
- O aceite formal da Etapa 1/1.1 em `61b58c2` foi registrado depois de revisão adversarial sem bloqueios materiais; os nove achados residuais seguem como follow-ups.
- Nenhum código da publicação versionada foi alterado nesta decisão.
- Nenhuma worktree, stash, branch ou cache local foi removido ou reescrito.

## Changelog

- 1.0 (2026-10-17): decisões do dono registradas como direções de arquitetura; implementação bloqueada por auditorias de compatibilidade e verificações de plataforma.
