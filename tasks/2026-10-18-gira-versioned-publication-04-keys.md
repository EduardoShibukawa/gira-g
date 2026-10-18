---
schema: task/1
tier: standard
status: todo
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
| AC1 | Cada contraexemplo do algoritmo de fingerprint altera source_key; codificação injetiva comprovada por teste/fuzz. | not run | | |
| AC2 | Comentário/BOM/EOL alterados mudam hash do manifesto conforme contrato. | not run | | |
| AC3 | main.date isoladamente não muda tool_id/mapa de release; main.version/main.commit mudam tool_id; dev build não confiável inclui hash executável. | not run | | |
| AC4 | source_identity inclui todas as fontes relevantes, ordenadas deterministicamente, sem URL crua. | not run | | |
| AC5 | Schema 2 e source_identity sobrevivem CanonicalJSON, clone, store roundtrip e snapshot destacado. | not run | | |
| AC6 | Teste de cobertura detecta campo MacroMap não propagado por clone/serialização. | not run | | |
| AC7 | Mesmo version_id com bytes distintos retorna nondeterministic_build e preserva versão existente. | not run | | |
| AC8 | CI 3 SOs passa. | not run | | |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-18
- status: todo; ainda não executada
- next: Depende de ADR 0014, task 02 Git environment e task 03 locator.
- revalidate: todos os checks estão not run salvo decisão do dono explicitamente registrada
