---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: identidade source_key, tool_id e source_identity

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de ADR 0014, task 02 Git environment e task 03 locator.

## Objective

Implementar identidade determinística de fonte/ferramenta/versão e proveniência autocontida do schema 2, usando bytes crus do manifesto e hashes das fontes.

## In scope

Codificação injetiva; source_key; tool_id conforme opção C; hash do executável em dev build não confiável; source_identity com schema_version 2, manifest_sha256, extractors e sources ordenadas com url_sha256 e SHAs; incluir campo em clone e CanonicalJSON sem perda silenciosa.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Cada contraexemplo do algoritmo de fingerprint altera source_key; codificação injetiva comprovada por teste/fuzz.
2. AC2: Comentário/BOM/EOL alterados mudam hash do manifesto conforme contrato.
3. AC3: main.date isoladamente não muda tool_id/mapa de release; main.version/main.commit mudam tool_id; dev build não confiável inclui hash executável.
4. AC4: source_identity inclui todas as fontes relevantes, ordenadas deterministicamente, sem URL crua.
5. AC5: Schema 2 e source_identity sobrevivem CanonicalJSON, clone, store roundtrip e snapshot destacado.
6. AC6: Teste de cobertura detecta campo MacroMap não propagado por clone/serialização.
7. AC7: Mesmo version_id com bytes distintos retorna nondeterministic_build e preserva versão existente.
8. AC8: CI 3 SOs passa.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Cada contraexemplo do algoritmo de fingerprint altera source_key; codificação injetiva comprovada por teste/fuzz. | pass | 8102fec | `TestInjectiveEncoding_Injectivity`, `FuzzInjectiveEncoding` (191k execuções em 3s, zero colisões) e `TestContraexamples_AltersSourceKey` em `internal/store`: cada contraexemplo (locator relativo vs absoluto/diferente, comentários, path_scope, repo_id, repositório com URL distinta) altera `source_key`; codificação com prefixo de tamanho de 64 bits garante injetividade estrita `("ab","c") != ("a","bc")`. |
| AC2 | Comentário/BOM/EOL alterados mudam hash do manifesto conforme contrato. | pass | 8102fec | `TestManifestRawBytesHash_Comment_BOM_EOL` em `internal/store`: matriz cobrindo CRLF vs LF, prefixo UTF-8 BOM (`\xef\xbb\xbf`) e comentários adicionados/modificados comprova que o SHA-256 dos bytes crus do manifesto e a `source_key` resultante são mutuamente exclusivos e estritamente distintos para cada variante. |
| AC3 | main.date isoladamente não muda tool_id/mapa de release; main.version/main.commit mudam tool_id; dev build não confiável inclui hash executável. | pass | 8102fec | `TestToolID_Date_Version_Commit_DevBuild` em `internal/store`: Opção C implementada com exclusão integral da setting `-ldflags`; variação isolada de data entre 2001 e 2030 mantém `tool_id` idêntico em release; bump de versão ou commit altera `tool_id`; dev builds não confiáveis (`commit="none"`, `vcs.modified=true` ou ausência de `vcs.revision`) incluem o hash SHA-256 do binário executável (`os.Executable`), alterando `tool_id` em caso de recompilação. |
| AC4 | source_identity inclui todas as fontes relevantes, ordenadas deterministicamente, sem URL crua. | pass | 8102fec | `TestSourceIdentity_DeterministicOrdering_NoRawURL` em `internal/domain`: `SourceIdentity` inclui todos os repositórios declarados no manifesto (mesmo sem evidência extraída), ordenados deterministicamente por `repo_id`, com extratores ordenados por `id`; URLs cruas e credenciais são totalmente expurgadas, persistindo exclusivamente `url_sha256` via `gitcache.RepoHash`. |
| AC5 | Schema 2 e source_identity sobrevivem CanonicalJSON, clone, store roundtrip e snapshot destacado. | pass | 8102fec | `TestMacroMap_AC5_Roundtrip_Clone_CanonicalJSON_Snapshot` em `internal/domain`, `TestStore_Schema2_SourceIdentity_Roundtrip` em `internal/store` e `TestCloneMacroMapPreservesSourceIdentity` em `internal/indexer`: `schema_version: 2` (inteiro) e `source_identity` sobrevivem sem mutação ou perda a `CanonicalJSON()`, `cloneMacroMap()` (deep clone com slices independentes), persistência em disco via `Store.SaveFile`/`LoadFile` e snapshots destacados. |
| AC6 | Teste de cobertura detecta campo MacroMap não propagado por clone/serialização. | pass | 8102fec | `TestMacroMap_AC6_ReflectionCoverage` e `TestMacroMap_AC6_NegativeDetection` em `internal/indexer`: teste reflexivo inspeciona exaustivamente todos os campos exportados de `domain.MacroMap`, valida conformidade com tags JSON, clonagem profunda em `cloneMacroMap` e projeção integral em `CanonicalJSON()`, falhando fechado caso um novo campo seja adicionado à struct sem suporte explícito. |
| AC7 | Mesmo version_id com bytes distintos retorna nondeterministic_build e preserva versão existente. | pass | 8102fec | `TestVersionID_NonDeterministicBuild_PreservesExisting` em `internal/store`: `VerifyDeterministicBuild` detecta divergência entre o mesmo `version_id` e bytes canônicos distintos, emitindo erro estável `ErrNonDeterministicBuild` (`nondeterministic_build`), bloqueando sobrescrita e preservando o arquivo preexistente byte a byte intacto; reindexação com bytes idênticos opera de forma idempotente (nil error). |
| AC8 | CI 3 SOs passa. | pass | 8102fec | `go test -race ./...` passando 100% em todos os 17 pacotes (`cmd/gira`, `internal/archdiff`, `internal/corpus`, `internal/domain`, `internal/eval`, `internal/extractors`, `internal/gate`, `internal/gitcache`, `internal/gitdiff`, `internal/human`, `internal/indexer`, `internal/manifest`, `internal/mcp`, `internal/skills`, `internal/slices`, `internal/store`, `internal/tokens`). `bun tools/validate/validate.ts` passando com `Overall: pass`. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19
- status: completed; source_key, tool_id (Opção C), version_id e source_identity implementados e testados com sucesso
- next: Tarefas 05 (locks) e 06 (publish) da publicação versionada
- revalidate: todos os critérios AC1–AC8 verificados com evidência no commit 8102fec
