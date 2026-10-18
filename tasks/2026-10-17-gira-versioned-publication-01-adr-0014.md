---
schema: task/1
tier: standard
status: todo
parent: 2026-10-17-gira-versioned-publication
spec: 2026-10-17-gira-versioned-publication@1
depends_on: []
---

# Gira-G: ADR 0014 para publicação versionada

## Routing and approvals
- Parent heavy autorizado pelo dono em 2026-10-18; esta tarefa documental está liberada.
- A spec aprovada é a fonte normativa; não inventar alternativas novas nem contradizer N1–N4, D1–D9c e Gates 1–5.

## Objective
Consolidar em ADR 0014 as decisões já registradas para o store versionado, identidade da fonte/ferramenta, locks, durabilidade, compatibilidade e semântica CLI.

## In scope
- Criar ADR 0014 usando o formato existente de ADRs.
- Registrar contexto, decisões, consequências, limites de plataforma, rollback e links para a spec e os gates.
- Registrar explicitamente que o store é cache local derivado, não fonte de verdade compartilhada.
- Incluir a política aprovada: caminhos Git que não possam ser materializados no Windows causam falha explícita, sem aceitar extração parcial como sucesso.

## Out of scope
- Alterar as decisões de produto aprovadas.
- Implementar código ou declarar gates técnicos satisfeitos.

## Acceptance criteria
- [ ] ADR 0014 existe no diretório canônico de ADRs e segue o padrão local.
- [ ] Cada decisão aponta para a seção correspondente da spec e não contradiz decisões anteriores.
- [ ] Durabilidade, limites de locks, schema legado, rollback e política Windows estão explícitos.
- [ ] Links e validador de documentação passam.

## Verification
| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | ADR revisado contra spec e decisões | not run | | |
| AC2 | Validador e docs build | not run | | |

## Dependencies
- Nenhuma. Pode executar em paralelo com a tarefa 02.
