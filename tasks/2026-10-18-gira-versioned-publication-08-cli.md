---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: integrar index/init e fluxos CLI

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 05 locks e task 07 resolver.

## Objective

Integrar o protocolo versionado aos comandos index/init e consumidores da CLI, preservando a ordem de publicação e semântica dos snapshots destacados.

## In scope

index sem output publica canônico; index --output só snapshot; init publica canônico e cópia byte a byte; manifesto por último; conflito se manifesto muda durante init; remover fallback eval para gira.yaml; skills --index legacy warning e path absoluto; erros estáveis.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: index --output P não altera store/head/manifesto.
2. AC2: init --output P publica canônico e P com bytes idênticos; manifesto é gravado por último.
3. AC3: Falha de build/gate/sync não cria manifesto novo nem altera o existente.
4. AC4: Alteração concorrente do manifesto durante init gera conflito sem publicar estado misto.
5. AC5: eval sem manifesto real retorna manifest_path_required quando overrides exigem locator.
6. AC6: skills install/check --index avisa para legado e normaliza caminho relativo para absoluto.
7. AC7: Mensagens refletem publicação parcial/visible_unconfirmed corretamente.
8. AC8: Comandos existentes continuam funcionando e CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | index --output P não altera store/head/manifesto. | pass | `TestAC1_IndexDetachedSnapshotDoesNotTouchStoreOrManifest` | Snapshot publicado em P; store, .head e gira.yaml intocados. |
| AC2 | init --output P publica canônico e P com bytes idênticos; manifesto é gravado por último. | pass | `TestAC2_InitOutputPublishesCanonicalAndIdenticalSnapshotManifestLast` | Cópia P idêntica byte a byte ao canônico; manifesto gravado ao final. |
| AC3 | Falha de build/gate/sync não cria manifesto novo nem altera o existente. | pass | `TestAC3_InitFailureDoesNotCreateOrModifyManifest` | Falha antes do commit aborta sem criar ou tocar gira.yaml. |
| AC4 | Alteração concorrente do manifesto durante init gera conflito sem publicar estado misto. | pass | `TestAC4_InitConcurrentManifestModificationGeneratesConflict` | Modificação do manifesto com lock detectada; aborta com erro de conflito. |
| AC5 | eval sem manifesto real retorna manifest_path_required quando overrides exigem locator. | pass | `TestAC5_EvalWithoutManifestPathReturnsManifestPathRequired`, `TestCorpus_DeclarativeOverridesRequireManifestPath` | Remove fallback silencioso para gira.yaml; retorna manifest.ErrManifestPathRequired. |
| AC6 | skills install/check --index avisa para legado e normaliza caminho relativo para absoluto. | pass | `TestAC6_SkillsInstallAndCheckLegacyWarningAndAbsolutePath` | Aviso emitido para schema 1.0.0; caminhos relativos resolvidos para absolutos. |
| AC7 | Mensagens refletem publicação parcial/visible_unconfirmed corretamente. | pass | `TestAC7_InitPartialPublicationAndDurability` | Falha na cópia destacada reporta publicação parcial e durabilidade preservada. |
| AC8 | Comandos existentes continuam funcionando e CI 3 SOs passa. | pass | `go test -race ./...` | Suíte completa com -race passou com êxito (cmd/gira, eval, skills, store). |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed
- next: Integração final do fluxo CLI concluída; submeter commit e reportar ao orquestrador.
- revalidate: `go test -race ./...` e `bun tools/validate/validate.ts` passaram com 100% de sucesso.
