---
schema: task/1
tier: heavy
status: active
spec: 2026-10-17-gira-versioned-publication@1
---

# Gira-G: publicação versionada de índices (heavy autorizado)

## Routing and approvals

- tier: heavy
- confirmed: heavy by owner (2026-10-18, explicit user authorization in conversation; spec approval and implementation authorization recorded in `specs/2026-10-17-gira-versioned-publication.md`). Open technical evidence is retained as mandatory child-task acceptance criteria, not as a prohibition on starting implementation.
- approved: spec v1 by owner (2026-10-18, explicit spec approval and heavy implementation authorization)
- decided: N1 (publicação versionada = continuação do item 1, rótulo próprio, ex. "Etapa 1.2"; item 2 aprovado reservado a atualidade e sync) → (a) by owner (2026-10-17, owner, 2026-10-17 (session decision message))
- decided: N2 (aceite da Etapa 1/1.1) → (b) revisão adversarial somente leitura de `4c2204c` e `0d5f248` (`aa4e6fe..0d5f248`, exclui `aa4e6fe`), depois aceite se não houver bloqueio material by owner (2026-10-17, owner, 2026-10-17 (session decision message))

## Objective

Publicar índices como versões imutáveis com identidade determinística (`source_key`, `tool_id`, `version_id`), resolução que falha fechada, travas entre escritores e semântica de durabilidade explícita, conforme o spec `2026-10-17-gira-versioned-publication@1`.

In scope:
- etapas 2a–2i do [plano de execução](../docs/plans/2026-10-17-plano-execucao-governanca.md).

Out of scope:
- coleta automática de versões;
- `--require-durable` e `--no-lock`;
- nome canônico do arquivo de índice;
- itens 2–5 do plano de evolução aprovado.

## Acceptance criteria

- AC1 = SC1 … AC10 = SC10 do spec.

Pacote de preparação para revisão: [compatibilidade e regressão da publicação versionada](../docs/plans/2026-10-18-pacote-compatibilidade-regressao-publicacao-versionada.md). É um plano documental; os testes listados permanecem não executados até haver autorização e implementação.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test -race ./...` (CI: ubuntu, macos, windows) | pass | 61b58c2 | CI run 39072167231, push do merge da PR #17 |
| B2 | `node tools/validate/validate.js` | pass | 61b58c2 | rodado no worktree deste branch antes da primeira edição |
| SC1 | matriz adversarial de invocação e ambiente | not run | | 2c, 2d |
| SC2 | testes de `gitcache` com `quotePath`, `refs/replace`, não ASCII e config de clone/fetch | not run | | 2b |
| SC3 | testes de chave, contraexemplos e fuzz de codificação | not run | | 2d |
| SC4 | testes do resolvedor (S1, S2, tool mismatch, legado) | not run | | 2g |
| SC5 | testes de publicação (falhas antes e depois do rename, gate, head corrompido) | not run | | 2f |
| SC6 | durabilidade por plataforma (CI 3 SO) | not run | | 2f, 2h |
| SC7 | testes de trava nos 3 SO | not run | | 2e |
| SC8 | testes de integração de `init` e `index` | not run | | 2h |
| SC9 | validador e `mise run docs:build` | not run | | 2i |
| SC10 | `go test -race ./...` 3 SO, floor guard, cache real intacto | not run | | head integrado |
| R1 | CI completo no head integrado | not run | | |
| V1 | diff integrado lido contra SC1–SC10 | not run | | |

## Plan

Autorizado pelo dono em 2026-10-18: spec aprovada e parent `active`/`confirmed: heavy`. Os children são criados em ordem de dependência; cada etapa continua sujeita aos próprios critérios de aceite e gates.

Pré-condições, todas decisões do dono:
1. ~~N1 (nome e sequência)~~: decidida (a), 2026-10-17;
2. N2 (aceite da Etapa 1/1.1 em `61b58c2`): decidida (b); revisão adversarial concluída sem bloqueios materiais; aceite formal registrado em `2026-10-17-etapa-1-1-1-retro-acceptance` (`accepted`), merge da PR #19;
3. N3, N4, D1–D9c e decisões dos Gates 1–5: direção escolhida e registrada; a auditoria dos consumidores/schema, testes de caminhos/ambiente Git, regressões de `tool_id`, teste de `git archive | tar -x` com tab no Windows e validação por plataforma ainda precisam de evidência. Gate 5 é critério obrigatório de aceitação da implementação após autorização, não pré-condição para começar a primeira tarefa autorizada.

Children criados, em ordem de dependência (todos inicialmente `todo`, ainda sem execução):

0. `2026-10-18-gira-versioned-publication-00-audit-revalidation`: revalidar consumidores e limites de plataforma; antes da aceitação final.
1. `2026-10-18-gira-versioned-publication-01-adr-0014`: ADR 0014 e consolidação do contrato; depende de 00.
2. `2026-10-18-gira-versioned-publication-02-gitcache-env`: parsing NUL e isolamento Git; incorporar PR #28 sem duplicar mudanças.
3. `2026-10-18-gira-versioned-publication-03-locator`: locator/raiz Git; depende de 01 e 02.
4. `2026-10-18-gira-versioned-publication-04-keys`: source_key, tool_id e source_identity; depende de 01, 02 e 03.
5. `2026-10-18-gira-versioned-publication-05-locks`: locks M→S→P; depende de 01, pode avançar em paralelo com 03/04.
6. `2026-10-18-gira-versioned-publication-06-publish`: store versionado e durabilidade; depende de 04 e 05.
7. `2026-10-18-gira-versioned-publication-07-resolver`: reader estrito; depende de 06 e 04.
8. `2026-10-18-gira-versioned-publication-08-cli`: index/init/eval/skills; depende de 05 e 07.
9. `2026-10-18-gira-versioned-publication-09-docs-consumers`: consumidores, fixtures e docs; depende de 08.
10. `2026-10-18-gira-versioned-publication-10-integrated-acceptance`: SC1–SC10, CI e revisão final; depende de todas as anteriores.

A tarefa separada `2026-10-17-m0-m6-retro-review` também foi autorizada como revisão read-only e ativa; não é dependência da publicação versionada.
## Review

skip: roda depois que os children estiverem verificados

## Checkpoint

- as of: 2026-10-19, branch `work/2026-10-19-versioned-publication-full-local-run`
- done: N1/N2 e aceite formal Etapa 1/1.1; O6 aplicada; direções N3/N4/D1–D9c; decisões suplementares dos Gates 1–5 registradas; Tarefa 00 concluída; Tarefa 02 integrada em main (PR #32) e reconciliada; Tarefa 01 (ADR 0014) concluída e integrada; Tarefa 03 (locator e raiz Git) concluída e integrada; Tarefa 04 (source_key, tool_id e source_identity) concluída e integrada.
- next: Executar Tarefa 05 (`2026-10-18-gira-versioned-publication-05-locks`).
- revalidate: Validador standalone (`bun tools/validate/validate.ts`) em sucesso; suíte de testes passando com `-race`.





## Owner authorization record (2026-10-18)

- approved: `specs/2026-10-17-gira-versioned-publication.md` v1.3.
- confirmed: heavy by owner; implementation may begin through the child tasks below.
- scope remains limited to SC1–SC10 and steps 2a–2i. No task may claim unrun tests as passing or merge with required CI failing.
- children are ordered by dependencies; a later stage does not become accepted merely because an earlier stage merged.
