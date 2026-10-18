---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on:
  - 2026-10-17-gira-versioned-publication-02-gitcache-env
  - 2026-10-17-gira-versioned-publication-03-locator
  - 2026-10-17-gira-versioned-publication-04-keys
  - 2026-10-17-gira-versioned-publication-05-locks
  - 2026-10-17-gira-versioned-publication-06-publish
  - 2026-10-17-gira-versioned-publication-07-resolver
  - 2026-10-17-gira-versioned-publication-08-cli
  - 2026-10-17-gira-versioned-publication-09-docs
---

# Gira-G: validação integrada e aceite da publicação versionada

## Routing and approvals
- Esta é a tarefa de aceitação final do parent heavy. Não significa aprovação antecipada de merge/release.

## Objective
Executar SC1–SC10 da spec sobre o head integrado, consolidar evidência real e fazer revisão adversarial do diff completo.

## In scope
- Executar matriz de determinismo e ambientes isolados.
- Executar go test -race ./... em Ubuntu/macOS/Windows, validator, floor guard, gate-lint e docs build.
- Validar que o cache real do usuário permanece intacto (hash/listagem antes/depois).
- Revisar todas as fronteiras de falha/rename/fsync/head/manifesto, ordem M→S→P, downgrade, erro de schema e comportamento Windows.
- Atualizar tabela de Verification na spec e no parent com run IDs e SHAs reais.

## Acceptance criteria
- [ ] SC1–SC10 têm evidência real no head final; resultados não executados continuam not run.
- [ ] CI completo passa nos três sistemas e todos os jobs exigidos.
- [ ] Nenhum teste toca o cache real do usuário.
- [ ] Revisão adversarial não encontra bloqueio material aberto; qualquer exceção é documentada e aprovada explicitamente.
- [ ] Spec/parent só mudam para accepted/verified depois de toda evidência aprovada pelo dono.
- [ ] Diff final e rollback são revisados.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| SC1-SC10 | Spec acceptance matrix | not run | | |
| R1 | CI full matrix | not run | | |
| V1 | Adversarial diff review | not run | | |

## Dependencies
- Todos os children de implementação concluídos.
