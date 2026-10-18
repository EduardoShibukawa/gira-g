---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-01-adr-0014
  - 2026-10-17-gira-versioned-publication-02-gitcache-env
---

# Gira-G: locator canônico do manifesto

## Objective
Derivar locators de overrides de forma determinística e segura a partir dos bytes reais do manifesto e da raiz Git correta.

## In scope
- Resolver symlinks e normalizar caminho relativo à raiz Git com separadores /.
- Implementar manifest_outside_git_root quando houver overrides fora da raiz e git_root_unavailable em falha de descoberta.
- Sem overrides, não falhar só porque o projeto não tem raiz Git.
- Cobrir raiz, subdiretório, worktree, submódulo, symlink e manifesto fora do Git.
- Garantir que falha ocorre antes de qualquer publicação.

## Out of scope
- Fingerprint completo source_key/tool_id e layout de store.

## Acceptance criteria
- [ ] Mesmo manifesto acessado por formas equivalentes produz locator idêntico.
- [ ] Manifesto com overrides fora da raiz falha com erro estável e sem artefatos publicados.
- [ ] Falha de descoberta Git é distinta de manifesto fora da raiz.
- [ ] Sem overrides, ausência de raiz Git não é motivo isolado de falha.
- [ ] Testes cobrem symlink/worktree/submódulo ou documentam skip de plataforma com evidência.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | Matriz de locators e erro fail-closed | not run | | |
| AC2 | Testes de não publicação em falha | not run | | |

## Dependencies
- ADR 0014 e tarefa Git/environment.
