---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
depends_on:
  - 2026-10-17-gira-versioned-publication-08-cli
---

# Gira-G: documentação e consumidores

## Objective
Atualizar consumidores e documentação para refletir o comportamento efetivamente implementado, sem prometer comportamento ainda não verificado.

## In scope
- README, website docs/versioned_docs, documentação CLI.
- gira-arch-diff, bench/phases.json, internal/benchtest, tools/index-bundle/bootstrap.sh, skills install/check, MCP serve, eval/scorecard, goldens e fixtures relevantes.
- Corrigir consumidores estritamente incompatíveis e adicionar fixtures/regressões necessárias.
- Atualizar compatibilidade de schema, avisos de legado e mensagens de erro.

## Acceptance criteria
- [ ] Inventário de consumidores é reconciliado com a auditoria de compatibilidade; cada consumidor alterado tem teste/fixture quando aplicável.
- [ ] Documentação descreve comportamento real de --output, store, legado, locks e durabilidade.
- [ ] Bench, bundle, skill, MCP e gira-arch-diff continuam funcionando.
- [ ] Nenhum consumidor passa a aceitar schema 2 implicitamente sem política explícita.
- [ ] Validator e docs build passam.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes de consumidores/fixtures | not run | | |
| AC2 | Validator e docs build | not run | | |

## Dependencies
- CLI integrado.
