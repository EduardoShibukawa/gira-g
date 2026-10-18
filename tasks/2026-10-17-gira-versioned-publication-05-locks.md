---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-01-adr-0014
---

# Gira-G: locks M, S e P multiplataforma

## Objective
Implementar exclusão mútua fail-closed entre escritores locais, com ordem global M → S → P.

## In scope
- flock em Unix e LockFileEx no Windows.
- Timeout de 60 segundos e erros estáveis index_locked e lock_unavailable.
- Liberar lock quando processo filho morre; garantir que todos os fluxos liberam locks.
- Canonicalizar a identidade do manifesto para que caminhos relativos/absolutos/symlink concorram pelo mesmo lock.
- Documentar limites de NFS/FUSE/SMB, editores externos e hardlinks.

## Acceptance criteria
- [ ] Dois processos concorrentes se excluem nos três SOs.
- [ ] Caminhos relativos, absolutos e symlink do mesmo manifesto compartilham lock.
- [ ] Morte de processo libera lock; timeout retorna index_locked.
- [ ] Falha de backend retorna lock_unavailable sem artefatos parciais.
- [ ] Ordem M→S→P é verificada por testes.
- [ ] Testes não escrevem no cache real do usuário.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Testes multi-processo Linux/macOS/Windows | not run | | |
| AC2 | Testes timeout/falha e ordem dos locks | not run | | |

## Dependencies
- ADR 0014. Pode correr em paralelo com locator/keys.
