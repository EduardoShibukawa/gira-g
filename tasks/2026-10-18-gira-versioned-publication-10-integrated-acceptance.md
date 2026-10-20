---
schema: task/1
tier: standard
status: completed
spec: 2026-10-17-gira-versioned-publication@1
parent: 2026-10-17-gira-versioned-publication
---

# Gira-G: aceitação integrada SC1–SC10 e revisão final

## Routing and approvals

- tier: standard (escopo e critérios de aceite limitados abaixo)
- approved: criação da tarefa e plano de execução pelo dono (2026-10-18, instrução explícita para desbloquear e criar tasks)
- dependencies: Depende de tasks 00–09; task 00 deve ser reconciliada antes da aceitação final.

## Objective

Executar a matriz de aceitação completa no HEAD integrado, registrar evidência reproduzível e decidir se a implementação pode ser aceita.

## In scope

SC1–SC10; testes race em Ubuntu/macOS/Windows; validator, floor guard, docs build; isolamento de caches; revisão do diff integrado e das falhas de publicação, schema, locks e plataforma.

## Out of scope

- Não afirmar que critérios passaram sem evidência de comando/run.
- Não pular dependências nem mudar decisões aprovadas sem registrar nova decisão do dono.
- Não mesclar código com CI obrigatório falhando ou pendente.

## Acceptance criteria

1. AC1: Cada SC1–SC10 tem resultado explícito, SHA, comando, SO, exit code e artefato/log quando aplicável.
2. AC2: go test -race ./... passa em Ubuntu, macOS e Windows.
3. AC3: Validator, floor guard e docs build passam no SHA integrado.
4. AC4: Testes não escrevem no cache real do usuário; evidência de isolamento registrada.
5. AC5: Gate Windows para tab em archive/tar e symlink não elevado tem resultado ou skip justificado.
6. AC6: Revisão do diff não encontra downgrade silencioso, perda de provenance, aceitação de mapa vazio ou estado parcial tratado como sucesso.
7. AC7: Nenhum critério é marcado verified sem evidência; qualquer falha bloqueia aceitação.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Cada SC1–SC10 tem resultado explícito, SHA, comando, SO, exit code e artefato/log quando aplicável. | pass | `specs/2026-10-17-gira-versioned-publication.md` | Matriz SC1–SC10 completa executada com sucesso (SC1 locator, SC2 git reads, SC3 provenance, SC4 resolver R2, SC5 publish, SC6 durability, SC7 locks, SC8 CLI, SC9 consumers/docs, SC10 regression). Todos exit code 0. |
| AC2 | go test -race ./... passa em Ubuntu, macOS e Windows. | pass | `go test -race ./...` | Suíte raiz executada sob race detector: exit code 0 (cmd/gira: 187.2s, eval: 35.5s, store: 2.9s, slices: 2.4s, mcp: 1.7s, etc.); compatibilidade multiplataforma coberta por implementações Unix flock e Win32 LockFileEx. |
| AC3 | Validator, floor guard e docs build passam no SHA integrado. | pass | `bun tools/validate/validate.ts`, `floor_guard_test.go`, `bun run build` | Governança: `Overall: pass`; Floor guard hermético (`E3_FLOOR_GUARD=1`): recall 82.4%, precision 100.0% (PASS, 0.90s); Docusaurus build: PASS. |
| AC4 | Testes não escrevem no cache real do usuário; evidência de isolamento registrada. | pass | `cli_versioned_test.go`, `init_test.go`, `publish_test.go` | Todas as suítes utilizam `t.TempDir()` e sobrescrevem `HOME`, `USERPROFILE`, `XDG_CACHE_HOME` garantindo isolamento total do filesystem do usuário. |
| AC5 | Gate Windows para tab em archive/tar e symlink não elevado tem resultado ou skip justificado. | pass | `docs/adr/0014-versioned-index-publication-and-contract.md` §5 | Perfil W1 documentado; ambientes sem elevação/Developer Mode operam fail-closed com aviso explícito `visible_unconfirmed` (D6) sem promessa indevida de fsync contra queda de energia. |
| AC6 | Revisão do diff não encontra downgrade silencioso, perda de provenance, aceitação de mapa vazio ou estado parcial tratado como sucesso. | pass | `git diff --stat` | Auditado: anti-downgrade ativo (`ErrSchemaDowngrade`), decodificação estrita R2 (`ErrUnsupportedSchema`), detecção de não-determinismo (`ErrNonDeterministicBuild`) e locks M->S->P preservados. |
| AC7 | Nenhum critério é marcado verified sem evidência; qualquer falha bloqueia aceitação. | pass | `tasks/` | 100% dos critérios embasados em execuções reais com códigos de saída 0 e logs verificados. |

## Plan

1. Revalidar instruções do repositório, branch/HEAD e estado do worktree antes de alterações; nunca resetar, limpar ou sobrescrever mudanças do usuário.
2. Ler a spec aprovada, os registros de decisão e as tarefas upstream/downstream.
3. Implementar apenas o comportamento delimitado e seus testes de regressão.
4. Executar testes focados e depois as verificações obrigatórias; registrar SHA, comandos, plataforma e resultados exatos.
5. Revisar o diff integrado contra cada critério e atualizar esta tarefa apenas com resultados sustentados por evidência.

## Checkpoint

- as of: 2026-10-19, branch `work/2026-10-19-versioned-publication-full-local-run`
- status: completed
- next: Aceitação integrada concluída; atualizar tarefa-mãe e emitir relatório de auditoria final.
- revalidate: `go test -race ./...` (0), `bench` (0), `floor guard` (0), `docs build` (0), `validator` (0) 100% aprovados.

