---
schema: task/1
tier: standard
status: verified
spec: 2026-10-20-gira-agent-context-surface@1
verified_at: "876c868"
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
| B1 | baseline na base do filho (docs only) | pass | 2917515 | suíte e validator verdes antes do merge docs |
| R1 | `go test -race ./...` | pass | 876c868 | 17 pacotes ok no head integrado |
| SC1–SC5 | baterias dos filhos reexecutadas no head integrado | pass | 876c868 | `go test ./internal/slices ./cmd/gira ./internal/store -run 'SCAgent\|SCIndexCorrupt'` ok |
| R2 | `node tools/validate/validate.js` | pass | 876c868 | manifest-integrity + artifact-records pass |
| SC1–SC5 | baterias dos filhos, reexecutadas no head integrado | not run | | |
| AC1 | SC6 integrado (R1+R2+goldens) | pass | 876c868 | `git diff fixtures/ testdata/` vazio |
| AC2 | docs do site atualizadas (slice/impact/locate) | pass | 876c868 | exemplo caveman + validação estrita documentados |
| AC3 | `which gira; gira version; gira slice service:internal --tier summary --budget 400 --format caveman` | pass | 876c868 | 0.1.0 f5dad05 → 0.2.0 876c868 no ~/.local/bin/gira; summary de service:internal (G1) verde no PATH do pi-harness; `--format xml` recusa com `json\|md\|caveman`; impact/locate caveman reais gravados |
| V1 | diff lido contra SC1–SC6 | pass | 876c868 | nenhuma mudança de ground truth; cortes só aditivos |

## Constraints and budget

- Instalação local apenas; nenhum artefato de release tocado.
- Goldens/ground truth intocados (regeneração exigiriam nota explícita — esperado zero).

## Outcome

Spec 2026-10-20-agent-context-surface integrada na main local: 01 renderers+fit, 02 CLI+caveman+validação estrita, 03 guarda G1, 04 erros G3, 05 docs+build. Binário instalado: gira 0.2.0 (commit 876c868) em ~/.local/bin/gira (backup 0.1.0 em gira.bak-0.1.0); AC3 observado com o índice dogfood real. Nenhum push/tag.
