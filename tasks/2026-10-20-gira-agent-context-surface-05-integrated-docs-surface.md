---
schema: task/1
tier: standard
status: todo
spec: 2026-10-20-gira-agent-context-surface@1
parent: 2026-10-20-gira-agent-context-surface
depends_on: ["2026-10-20-gira-agent-context-surface-01-caveman-renderers", "2026-10-20-gira-agent-context-surface-02-caveman-cli-surface", "2026-10-20-gira-agent-context-surface-03-summary-intra-regression", "2026-10-20-gira-agent-context-surface-04-corrupt-index-errors"]
---

# Integração: suíte completa, docs do site e build local reinstallada (SC6)

## Routing and approvals

- tier: standard (integração e docs; nenhum corte novo de comportamento)
- approved: criação pelo dono via spec v1 aprovada (2026-10-20, instrução unattended)

## Objective

Verificação integrada dos quatro filhos na main local: `go test -race ./...` e `node tools/validate/validate.js` verdes; docs do site (`website/docs/cli/`) documentando `--format=caveman` em slice/impact/locate com a gramática e o exemplo 1-fato-por-linha; rebuild do binário com `-ldflags` de versão/commit e instalação em `~/.local/bin/gira` (o PATH que o pi-harness usa), registrando `which gira`/`gira version` antes e depois no registro desta task; confirmação de que `slice service:internal --tier summary --budget 400` passa com o binário instalado.

In scope: SC6; `website/docs/cli/*`; o ato de instalar (fora do git).

Out of scope: push, PR, tag, release, goreleaser; Fase 2 (pi-harness) e Fase 3 (bench) pertencem ao outro repo.

## Acceptance criteria

- AC1 = SC6: root suite race + validator verdes na main local pós-merge dos 4 filhos
- AC2 = docs do site refletem o formato e a validação estrita; exemplo reproduzível
- AC3 = binário do PATH reporta a nova build e serve caveman + summary budget 400

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| R1 | `go test -race ./...` | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| SC1–SC5 | baterias dos filhos, reexecutadas no head integrado | not run | | |
| AC3 | `which gira; gira version; gira slice service:internal --tier summary --budget 400 --format caveman` | not run | | |
| V1 | diff lido contra SC1–SC6 | not run | | |

## Constraints and budget

- Instalação local apenas; nenhum artefato de release tocado.
- Goldens/ground truth intocados (regeneração exigiriam nota explícita — esperado zero).
