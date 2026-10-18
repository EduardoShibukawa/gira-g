---
schema: task/1
tier: standard
status: active
---

# Gira-G: revisão retroativa e verificação dos critérios de saída de M0–M6 (`aa4e6fe`)

## Routing and approvals

- tier: standard (S1 um commit já em `main` a revisar, sem mudança de contrato nesta tarefa; revisão somente leitura; várias sessões possíveis, mas sem H1–H5 provado: reavaliar se a revisão achar divergência de contrato)
- decided: M0–M6 fica fora da revisão adversarial de N2 e vira frente separada by owner (2026-10-17, owner, 2026-10-17 (session decision message))
- approved: revisão retrospectiva read-only de aa4e6fe e verificação dos critérios de saída por owner (2026-10-18, explicit user instruction to unblock remaining tasks)

## Objective

Decidir e planejar a revisão retroativa e a verificação dos critérios de saída de M0–M6 no commit `aa4e6fe` ("implement M0-M6 roadmap": 46 arquivos, +6157/−611, sem registro de tarefa, sem PR; V6 do [plano de execução](../docs/plans/2026-10-17-plano-execucao-governanca.md) §5). A revisão está autorizada e ativa; nenhum trabalho de verificação foi executado ainda.

In scope:
- decidir se e como revisar `aa4e6fe`;
- levantar os critérios de saída de M0–M6 e verificá-los.

Out of scope:
- a revisão da Etapa 1/1.1 (`4c2204c`, `0d5f248`), tratada em N2;
- qualquer mudança de código.

## Acceptance criteria

- AC1 [inferred]: o dono decide escopo e forma da revisão (ou descarta a frente).
- AC2 [inferred]: cada critério de saída de M0–M6 tem uma linha em `## Verification` com resultado e commit.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | baseline (`go test -race ./...`, validador) | not run | | nada executado |
| AC1 | decisão do dono sobre escopo read-only | pass | user authorization 2026-10-18 | Revisão autorizada; execução continua pendente |
| AC2 | critérios de saída verificados | not run | | |
| V1 | self-review | not run | | |

## Checkpoint

- as of: 2026-10-17, branch `work/2026-10-17-gira-n1-n2-o6-decisions`, base `186427e`
- done: registro criado; nenhum trabalho feito
- next: o dono decide se e como revisar `aa4e6fe`
- open: onde estão os critérios de saída de M0–M6 (maturidade/roadmap nas Sources)
- revalidate: n/a


## Owner decision (2026-10-18)

- Scope approved: read-only retrospective review of commit aa4e6fe; no code changes in this task.
- Execution is authorized. Unknown or untraceable exit criteria must remain explicitly unknown rather than inferred.
