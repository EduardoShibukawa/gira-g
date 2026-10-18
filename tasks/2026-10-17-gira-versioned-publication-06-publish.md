---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
depends_on:
  - 2026-10-17-gira-versioned-publication-04-keys
  - 2026-10-17-gira-versioned-publication-05-locks
---

# Gira-G: store versionado e publicação atômica

## Objective
Implementar store local com versões imutáveis, heads por linhagem e publicação recuperável, com semântica de durabilidade explícita.

## In scope
- Layout de versão imutável, meta.json estrito e head por linhagem.
- version_id = H(source_key, tool_id); mesma versão com bytes diferentes retorna nondeterministic_build sem sobrescrever.
- Escrita temporária, sync, rename e atualização de head com recuperação idempotente.
- Estados confirmed e visible_unconfirmed.
- Resolver XDG_CACHE_HOME → ~/.cache → fallback no projeto; gate consulta também caminho legado conforme D2.
- index --output P como snapshot destacado sem mutar store/head/manifesto.
- Preservar snapshot válido/corrompido conforme --recover-corrupt.
- Implementar Gate 5 como parte obrigatória desta tarefa, sem protótipo descartável.

## Acceptance criteria
- [ ] Falha antes do rename não deixa artefato visível.
- [ ] Falha após rename preserva versão, mantém head intacto e retry é idempotente.
- [ ] Gate rejeitado não altera store, head ou snapshot.
- [ ] Head/versão ausente ou corrompida falha fechado.
- [ ] visible_unconfirmed reflete a semântica Unix/Windows da spec.
- [ ] Mesmo version_id com bytes diferentes retorna nondeterministic_build e preserva bytes existentes.
- [ ] Hashes antes/depois provam preservação de artefatos em cada falha injetada.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes de falha antes/depois do rename | not run | | |
| AC2 | Testes de durabilidade por plataforma | not run | | |
| AC3 | Testes de recuperação e nondeterministic_build | not run | | |

## Dependencies
- Chaves/source_identity e locks.
