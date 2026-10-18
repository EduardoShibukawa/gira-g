---
schema: task/1
tier: heavy
status: active
confirmed: heavy
spec: 2026-10-17-gira-versioned-publication@1
---

# Gira-G: publicação versionada de índices (proposta heavy)

## Routing and approvals

- confirmed: heavy by owner (2026-10-18, explicit instruction in this session to unblock the heavy work, approve the spec for implementation, and create the implementation tasks).

- tier: heavy (H1/H3/H5 conforme avaliação original). Confirmado pelo dono em 2026-10-18, por instrução explícita nesta sessão; a spec v1 está aprovada para implementação em etapas. Isso autoriza criar/iniciar os children, não aceitar os gates nem mesclar implementação sem evidência.
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

Autorizado pelo dono em 2026-10-18: parent `active`, `confirmed: heavy`; spec v1 `approved`. Os children abaixo estão criados nesta mudança e devem ser executados pela ordem de dependência. SC1–SC10, a política Windows e a revisão integrada continuam gates obrigatórios para aceite final.

Pré-condições, todas decisões do dono:
1. ~~N1 (nome e sequência)~~: decidida (a), 2026-10-17;
2. N2 (aceite da Etapa 1/1.1 em `61b58c2`): decidida (b); revisão adversarial concluída sem bloqueios materiais; aceite formal registrado em `2026-10-17-etapa-1-1-1-retro-acceptance` (`accepted`), merge da PR #19;
3. N3, N4, D1–D9c e decisões dos Gates 1–5: direção escolhida e registrada; a auditoria dos consumidores/schema, testes de caminhos/ambiente Git, regressões de `tool_id`, teste de `git archive | tar -x` com tab no Windows e validação por plataforma ainda precisam de evidência. Gate 5 é critério obrigatório de aceitação da implementação após autorização, não pré-condição para começar a primeira tarefa autorizada.

Children criados nesta branch, em ordem de dependência:
0. `2026-10-18-gira-clone-macro-map-diagnostics`: correções de clonagem, gitcache, ambiente e ordem de manifesto; implementação está na PR #28 e precisa de CI verde/merge antes do child 02.
1. `2026-10-17-gira-versioned-publication-01-adr-0014`: ADR 0014 consolidando as decisões aprovadas (2a); pode correr em paralelo com 0.
2. `2026-10-17-gira-versioned-publication-02-gitcache-env`: completar matriz de caminhos/ambiente Git e política Windows (2b); depende de 0.
3. `2026-10-17-gira-versioned-publication-03-locator`: locator normalizado (2c), depende de 1 e 2.
4. `2026-10-17-gira-versioned-publication-04-keys`: chaves e source_identity (2d), depende de 1, 2 e 3.
5. `2026-10-17-gira-versioned-publication-05-locks`: travas M, S e P (2e), depende de 1; pode correr em paralelo com 3 e 4.
6. `2026-10-17-gira-versioned-publication-06-publish`: publicação canônica e destacada (2f), depende de 4 e 5.
7. `2026-10-17-gira-versioned-publication-07-resolver`: resolvedor (2g), depende de 6.
8. `2026-10-17-gira-versioned-publication-08-cli`: runIndex/runInit (2h), depende de 5 e 7.
9. `2026-10-17-gira-versioned-publication-09-docs`: documentação e consumidores (2i), depende de 8.
10. `2026-10-17-gira-versioned-publication-10-integration`: SC1–SC10, R1, V1 e revisão adversarial final.

Uma etapa concluída não autoriza a seguinte.

## Review

skip: roda depois que os children estiverem verificados

## Checkpoint

- as of: 2026-10-17, branch `work/2026-10-17-gira-gate-owner-decisions`, base `731cf4ba`
- done: N1/N2 e aceite formal Etapa 1/1.1; O6 aplicada; direções N3/N4/D1–D9c; decisões suplementares dos Gates 1–5 registradas
- status atualizado em 2026-10-18: spec `approved`; tarefa-mãe `active`, `confirmed: heavy`; autorização explícita do dono registrada nesta sessão
- next: executar children pela ordem de dependência; não marcar SCs como pass sem testes/CI reais; política Windows para caminho não materializável é fail-closed
- open: evidência de compatibilidade de consumidores/schema; execução dos testes definidos pelas decisões; comportamento de tab em `git archive | tar -x` no Windows; W2 (symlink com usuário não elevado/sem Developer Mode, com skip justificado se não permitido); M0–M6 em `tasks/2026-10-17-m0-m6-retro-review.md`
- revalidate: CI PR #21 (head `4c59c6a`) passou `test` nos três SOs, `validator`, `floor guard`, `gate-lint` e `docs build`; essas são verificações de documentação, não testes da implementação versionada
- gates (2026-10-17): decisões registradas em `docs/plans/2026-10-17-gates-tecnicos-publicacao-versionada.md`. Gate 5 será testado no primeiro child autorizado e precisa passar antes de aceitar/mesclar a implementação. Status segue `todo`; spec `draft`; implementação não autorizada.
