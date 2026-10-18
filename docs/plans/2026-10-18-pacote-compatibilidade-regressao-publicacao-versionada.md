# Pacote de compatibilidade e regressão: publicação versionada

| Campo | Valor |
|---|---|
| Data | 2026-10-18 |
| Tipo | Plano de verificação, somente documentação |
| Estado | Proposta de testes; nenhum teste novo foi implementado ou executado por este documento |
| Base remota | `main` em `22d12ee75a9af87cc44e837492b547dc5f0932e7` |
| Spec | [publicação versionada](../../specs/2026-10-17-gira-versioned-publication.md), ainda `draft` |
| Tarefa-mãe | [publicação versionada](../../tasks/2026-10-17-gira-versioned-publication.md), ainda `todo` |
| Registro de evidência | [Gates técnicos](2026-10-17-gates-tecnicos-publicacao-versionada.md) |
| Decisões | [decisão arquitetural](2026-10-17-decisao-arquitetural-publicacao-versionada.md) |

Este pacote torna explícitos os testes de fronteira necessários para revisar a compatibilidade e preparar a aprovação da spec. Não aprova a spec, não adiciona `confirmed: heavy`, não cria tarefas-filhas e não autoriza implementação.

## 1. Estado da evidência

- **[CÓDIGO INSPECIONADO nesta revisão remota]** `internal/indexer/incremental.go`: `cloneMacroMap` copia campos manualmente e não copia `UnresolvedRefs` nem `Conflicts`.
- **[CÓDIGO INSPECIONADO nesta revisão remota]** `internal/gitcache/cache.go`: `ListTree` usa `git ls-tree` sem `-z` e separa saída por newline e tab.
- **[CÓDIGO INSPECIONADO nesta revisão remota]** `cmd/gira/main.go`: o fluxo de `init` grava/repina `gira.yaml` antes de chamar `indexManifest`.
- **[CI VERIFICADO via GitHub]** no SHA base, o workflow do Gira-G terminou com sucesso. Isso não cobre os cenários novos deste pacote.
- **[PENDENTE]** reprodução independente, testes de regressão, matriz de plataformas e auditoria completa dos consumidores externos.
- **[PENDENTE]** inspeção do consumidor correspondente no G-Ingá. A busca remota executada nesta revisão não resolveu o caminho exato do arquivo consumidor; não presumir que o inventário esteja completo.

Referências de código:
- [cloneMacroMap](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/indexer/incremental.go)
- [ListTree](https://github.com/EduardoShibukawa/gira-g/blob/main/internal/gitcache/cache.go)
- [runInit e probes Git](https://github.com/EduardoShibukawa/gira-g/blob/main/cmd/gira/main.go)

## 2. Regras para executar e registrar

1. Cada teste deve falhar antes da correção e passar depois dela; testes que apenas codificam o comportamento atual não fecham o risco.
2. Isolar `HOME`, `USERPROFILE`, `XDG_CACHE_HOME` e diretórios temporários. Comparar o cache real antes/depois quando a tarefa integrada exigir essa prova.
3. Registrar comando, sistema operacional, versão de Go/Git, SHA testado, exit code, logs e artefatos. Distinguir `not run`, `blocked`, falha esperada do produto e correção verificada.
4. Testes de schema precisam verificar código de erro estável e que nenhum dado existente foi sobrescrito; não basta verificar que o comando terminou com erro.
5. Testes de publicação devem inspecionar o estado em disco após cada ponto de falha: versão, head, snapshot destacado e manifesto.
6. Testes de concorrência devem usar processos reais separados, não apenas goroutines.
7. Não alterar arquivos de produto como parte deste pacote documental. Implementar os testes apenas após autorização e conforme spec aprovada.

## 3. Matriz de compatibilidade do mapa e do schema

| ID | Cenário | Resultado exigido | Evidência que fecha o teste |
|---|---|---|---|
| MAP-01 | `cloneMacroMap` com `UnresolvedRefs` e `Conflicts` preenchidos | Ambos preservados, sem alias mutável inesperado | Teste unitário com igualdade estrutural e mutação isolada |
| MAP-02 | Todo campo exportado de `domain.MacroMap` passa por clone e serialização canônica | Nenhum campo omitido silenciosamente | Teste de cobertura de campos (reflection ou mecanismo equivalente) para clone e `CanonicalJSON` |
| MAP-03 | `CanonicalJSON → LoadFile → CanonicalJSON` com `source_identity` | Roundtrip byte a byte determinístico | Teste de roundtrip com objeto completo e ordenação fixa |
| SCH-01 | Schema 2 válido, `schema_version` inteiro e `source_identity` completo | Aceitar e preservar todos os campos | Teste de leitura/escrita estrita |
| SCH-02 | Schema desconhecido, tipo incorreto ou campo obrigatório ausente | Falhar fechado com `index_schema_unsupported`, não `index_corrupt` | Testes de tabela validando código e mensagem estável |
| SCH-03 | `source_identity` ausente no schema novo ou schema interno divergente | Rejeitar | Testes negativos de cada variante |
| SCH-04 | SHA-256 do manifesto malformado, SHA de fonte inválido, fontes duplicadas ou fora de ordem | Rejeitar ou normalizar apenas conforme contrato explícito; nenhuma ambiguidade silenciosa | Testes de validação estrutural e de ordenação |
| SCH-05 | Binário antigo tenta ler ou republicar schema novo | Falhar fechado; nenhuma aceitação como mapa vazio e nenhum downgrade silencioso | Matriz de comandos e comparação do hash dos arquivos antes/depois |
| SCH-06 | Mapa legado `schema_version: "1.0.0"` | Só leitura explícita via `--index`, com aviso; nunca resolução implícita por `gira.yaml` nem reescrita in-place | Testes CLI para ambos os caminhos e hash de preservação |
| SCH-07 | Mesmo `version_id` publicado com bytes diferentes | `nondeterministic_build`; preservar versão existente | Teste de colisão injetada e comparação dos bytes existentes |

A aceitação de versões deve usar uma lista explícita e estrita. Não aceitar versões apenas porque o campo existe ou porque o JSON é sintaticamente válido.

## 4. Matriz Git, caminhos e identidade

| ID | Cenário | Resultado exigido | Evidência que fecha o teste |
|---|---|---|---|
| GIT-01 | `git ls-tree` com nome não ASCII, tab e newline | Caminho recuperado sem truncamento, divisão ou C-quoting ambíguo | Teste de integração usando saída NUL-delimited (`-z`) |
| GIT-02 | `core.quotePath` padrão e alterado; `refs/replace` presentes/ausentes | Mesmos bytes de índice | Testes pareados com hashes |
| GIT-03 | `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_PREFIX` redirecionando descoberta local | Probes e `init` não são sequestrados; falha fechada antes de publicar artefatos quando a raiz não é válida | Testes subprocess com ambiente controlado |
| GIT-04 | Clone/fetch com configuração/credenciais Git do usuário | Configuração necessária a clone/fetch continua funcional | Testes separados: probes locais sanitizados, operações remotas não sanitizadas indiscriminadamente |
| PATH-01 | Mesmo manifesto via caminho relativo, absoluto, descoberta em subdiretório e symlink | Mesmo locator normalizado e mesmos bytes | Matriz de hashes e igualdade do locator |
| PATH-02 | Manifesto em worktree e submódulo | Locator relativo à raiz correta e bytes determinísticos | Testes em repositórios temporários reais |
| PATH-03 | Manifesto com overrides fora de raiz Git | Erro `manifest_outside_git_root` ou `git_root_unavailable` conforme causa; nada publicado | Verificar ausência/inalteração de store, head, snapshot e manifesto |
| PATH-04 | Env var tenta alterar a raiz Git durante `init` | A raiz escolhida vem do repositório pretendido ou a operação falha fechada | Teste que reproduz o cenário adversarial e compara os artefatos |
| ID-01 | Alterar apenas comentário, BOM ou EOL do manifesto | `source_key` muda, pois a identidade usa bytes crus | Testes com bytes explícitos e hashes esperados |
| ID-02 | Pares ambíguos como (`ab`, `c`) e (`a`, `bc`) | Codificações/fingerprints diferentes | Teste unitário e fuzz da codificação injetiva |
| ID-03 | `main.date`, `main.version`, `main.commit`, `-trimpath` e build dev não confiável | Comportamento de `tool_id` conforme D5, sem variação indevida por data; hash de executável em dev não confiável | Builds controlados, comparação de `tool_id` e documentação do caso `vcs.*` |

## 5. Matriz de resolução, publicação e CLI

| ID | Cenário | Resultado exigido | Evidência que fecha o teste |
|---|---|---|---|
| RES-01 | Head aponta para versão ausente | `index_missing`; nunca retornar mapa vazio como sucesso | Teste de store corrompido e CLI com exit code não zero |
| RES-02 | Só existe versão de outra ferramenta | `index_tool_mismatch`; não carregar mapa | Teste com `tool_id` divergente |
| RES-03 | Head corrompido, versão corrompida ou schema desconhecido | Falhar fechado com código específico | Testes independentes para head, corrupção e schema |
| PUB-01 | Falha antes do rename da versão | Nenhum artefato novo visível | Injeção de falha e inventário do diretório |
| PUB-02 | Falha depois do rename da versão e antes da atualização do head | Versão pode ficar visível; head intacto; nova tentativa idempotente | Injeção de falha em cada fronteira e repetição |
| PUB-03 | Gate rejeita publicação | Store, head e snapshot destacados inalterados | Hash de todos os artefatos antes/depois |
| PUB-04 | `index --output P` | Só snapshot destacado em P; store, heads e manifesto inalterados | Hash de diretórios/arquivos antes/depois |
| PUB-05 | `init --output P` | Publicação canônica e P byte a byte idênticos; manifesto gravado por último | Teste CLI e comparação binária |
| PUB-06 | Primeira inicialização com build/gate falhando | `gira.yaml` não criado | Teste de integração com erro injetado |
| PUB-07 | `init --force` rejeitado | Manifesto existente, head e P inalterados | Snapshot de estado antes/depois |
| PUB-08 | Manifesto muda durante `init` | Conflito explícito; não publicar estado misto | Teste concorrente com sincronização determinística |
| DUR-01 | Sync falha no Unix após publicação visível | Estado `visible_unconfirmed`, exit code não zero; `init` não grava manifesto | Injeção de erro de sync e inspeção do estado |
| DUR-02 | Windows sem confirmação de sync de diretório | Estado `visible_unconfirmed`, exit 0 e aviso; sem alegação de durabilidade contra queda de energia | Teste Windows e captura da mensagem |
| LOCK-01 | Dois processos publicam em paralelo | Exclusão mútua | Teste multi-processo nos três SOs |
| LOCK-02 | Caminhos relativos, absolutos e via symlink para o mesmo manifesto | Mesma trava M | Teste multi-processo com caminhos equivalentes |
| LOCK-03 | Processo morre segurando trava | Trava liberada pelo SO; próxima operação prossegue | Processo filho morto e retry |
| LOCK-04 | Timeout e erro de mecanismo de lock | `index_locked` ou `lock_unavailable`; nenhuma gravação parcial por erro de lock | Injeção de timeout/erro e inspeção do estado |
| LOCK-05 | Aquisição combinada de travas | Ordem global M → S → P sempre respeitada | Teste de concorrência/instrumentação sem deadlock |
| CLI-01 | `eval` com overrides e sem caminho real do manifesto | `manifest_path_required`; não usar fallback silencioso para `gira.yaml` | Testes do chamador e hash dos bytes do manifesto |
| CLI-02 | `skills install --index` com legado, caminho relativo e snapshot destacado | Aviso legado e caminho normalizado conforme D9c | Testes de CLI para cada variante |

## 6. Consumidores e regressão cross-repository

Auditar cada consumidor antes de aprovar o schema final. O inventário inicial na spec cita, no Gira-G:

- CLI: `locate`, `impact`, `slice`, `explain`, `export`, `view`, `tour`, `flows`, `doc`, `diff`, `serve mcp`, `skills nudge/install/check`;
- store: `LoadFile`, `PublishFile`/ `Publish` e gate de perda de símbolos;
- cópia/serialização: `cloneMacroMap`, `CanonicalJSON`, fixtures, goldens e scorecard;
- ferramentas e benchmarks: `bench/layer0`, `bench/phases.json`, `internal/benchtest`, `tools/index-bundle/bootstrap.sh`, skill `gira-arch-diff`;
- documentação: README, CLI docs e `website/docs`.

Para G-Ingá, localizar e auditar o consumidor real que chama ou avalia Gira-G; não assumir o caminho apenas pelo nome citado em relatórios. Registrar arquivo/função, comando ou formato consumido, tratamento de exit code, versão/schema aceitos e testes. Incluir no mínimo:

1. saída vazia com exit 0 não pode ser interpretada como mapa válido;
2. schema desconhecido precisa distinguir incompatibilidade de corrupção;
3. caminho de `--index` e manifesto devem seguir o contrato de publicação destacado/canônico;
4. fixtures, scorecards e hashes de saída precisam de atualização deliberada, não regeneração silenciosa;
5. erro de `nudge` pode manter a política de não quebrar hooks, mas deve produzir observabilidade/teste explícitos sem fingir sucesso de leitura.

O inventário só fecha quando cada item tiver referência de código e teste, ou uma lacuna externa explicitamente aceita pelo dono.

## 7. Critérios de saída para revisão da spec

A spec pode ser submetida à revisão de aprovação quando houver evidência registrada para todos os pontos abaixo, sem alterar seu status automaticamente:

- [ ] Estratégia estrita de clone/serialização cobre todos os campos de `MacroMap`, incluindo `UnresolvedRefs`, `Conflicts` e `source_identity`.
- [ ] Matriz de compatibilidade documenta o comportamento de binários antigos, schemas desconhecidos e tentativa de downgrade.
- [ ] Consumidores internos e cross-repository estão enumerados com arquivos/funções e contratos.
- [ ] Identidade do manifesto e `tool_id` têm testes adversariais e determinísticos.
- [ ] Caminhos com tab/newline/não ASCII, variáveis Git, symlinks, worktrees, submódulos e manifesto fora do Git têm resultados registrados.
- [ ] O caso Windows `git archive | tar -x` com tab tem política explícita: não descartar arquivo silenciosamente nem tratar extração parcial como sucesso.
- [ ] Semântica de falha em cada fronteira de publicação e durabilidade está coberta por testes.
- [ ] Travas multi-processo e ordem M → S → P têm evidência nos três SOs, com skip de symlink justificado quando o runner não permite criá-lo.
- [ ] Comandos de verificação e resultados são registrados no artefato canônico, com SHA exato; cenários não executados permanecem `not run` ou `blocked`.
- [ ] A auditoria adversarial da Etapa 1/1.1 e as demais pré-condições registradas na tarefa-mãe foram revalidadas.

## 8. Não objetivos e limites

- Este documento não é uma implementação nem um plano que substitui o [plano de execução e governança](2026-10-17-plano-execucao-governanca.md).
- Não altera a autoridade da spec, das decisões ou da tarefa-mãe.
- Não afirma que os cenários listados já tenham testes, nem que qualquer teste desta matriz tenha sido executado.
- Não autoriza ADR 0014, tarefas-filhas, mudanças de código, release, merge ou ativação da tarefa heavy.
- O pacote deve ser refinado se a inspeção do consumidor real do G-Ingá ou a revisão do contrato revelar cenários adicionais.
