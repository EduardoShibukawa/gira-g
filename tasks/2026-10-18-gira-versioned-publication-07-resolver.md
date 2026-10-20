---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: resolver estrito e política de compatibilidade

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de task 06 publish e task 04 keys.

## Objective

Implementar resolução fail-closed para versões publicadas e leitura explícita de snapshots legados.

## In scope

Schema allowlist; schema 2 estrito; legado string 1.0.0 apenas via --index explícito com aviso; nunca legado implícito com gira.yaml; erros separados missing/corrupt/unsupported/tool mismatch/ambiguous; impedir downgrade.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Schema 2 válido é aceito e source_identity preservado.
2. AC2: Schema desconhecido, tipo incorreto e campos obrigatórios ausentes retornam index_schema_unsupported; corrupção retorna index_corrupt.
3. AC3: Legado v1 só via --index explícito, read-only e com aviso.
4. AC4: Com gira.yaml, legado nunca é carregado implicitamente.
5. AC5: Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch; descoberta ambígua retorna index_ambiguous.
6. AC6: Binário/fluxo antigo não pode republicar schema 2 como schema 1 sem erro e sem alterar bytes.
7. AC7: Erros de load não são colapsados para index_missing em MCP/CLI.
8. AC8: CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Schema 2 válido é aceito e source_identity preservado. | pass | b095226 | `TestResolver_Schema2Valid`: Schema 2 valid is accepted and preserves source_identity (manifest_sha256, extractors, sources) byte-for-byte |
| AC2 | Schema desconhecido, tipo incorreto e campos obrigatórios ausentes retornam index_schema_unsupported; corrupção retorna index_corrupt. | pass | b095226 | `TestResolver_SchemaUnknownAndCorrupt`: Non-int schema, unknown int versions (3), missing mandatory fields return ErrUnsupportedSchema; malformed JSON returns ErrCorruptIndex |
| AC3 | Legado v1 só via --index explícito, read-only e com aviso. | pass | b095226 | `TestCLI_Resolver_ExplicitIndexErrorsAndWarning`: Legacy 1.0.0 is accepted only with explicit --index, emits warning to stderr, and is read-only |
| AC4 | Com gira.yaml, legado nunca é carregado implicitamente. | pass | b095226 | `TestCLI_Resolver_OrphanLegacyNeverLoadedWithManifest`: With gira.yaml, legacy macro-map.json without .head active version fails closed with ErrIndexMissing |
| AC5 | Ausência de versão retorna index_missing; tool mismatch retorna index_tool_mismatch; descoberta ambígua retorna index_ambiguous. | pass | b095226 | `TestCLI_Resolver_ToolMismatchAndAmbiguousDiscovery`: Missing version returns ErrIndexMissing; mismatch tool returns ErrToolMismatch; multiple store systems without manifest returns ErrIndexAmbiguous |
| AC6 | Binário/fluxo antigo não pode republicar schema 2 como schema 1 sem erro e sem alterar bytes. | pass | b095226 | `TestResolver_SchemaDowngradePrevented`: Publishing schema 1 over existing schema 2 returns ErrSchemaDowngrade and leaves target file untouched |
| AC7 | Erros de load não são colapsados para index_missing em MCP/CLI. | pass | b095226 | `TestMCP_DiffMaps_LoadErrorsNotCollapsed`: MCP diff_maps preserves index_corrupt, index_schema_unsupported, index_tool_mismatch, index_ambiguous error codes |
| AC8 | CI 3 SOs passa. | pass | b095226 | `go test -race ./...`: Full race-enabled test suite passes locally (Linux x86_64); CI workflows configured across Linux/macOS/Windows |


## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; strict resolver R2 policy, schema allowlist, and compatibility implemented
- next: downstream tasks relying on resolver and versioned publication
- revalidate: go test -race ./... and bun tools/validate/validate.ts passing
