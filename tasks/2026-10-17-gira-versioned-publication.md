---
schema: task/1
tier: heavy
status: todo
spec: 2026-10-17-gira-versioned-publication@1
---

# Gira-G: publicação versionada de índices (proposta heavy)

## Routing and approvals

- tier: heavy (H1: contrato de `MacroMap`/proveniência, compatibilidade e algumas políticas de fallback ainda exigem decisão evidenciada; H3: muda o contrato de `--output`, schema, layout do store e erros do resolvedor consumidos por skills, bench, tools e docs; H5: várias etapas em várias sessões). Direções N3, N4 e D1–D9c registradas em `docs/plans/2026-10-17-decisao-arquitetural-publicacao-versionada.md`; D5 sobre `main.date` e a política de manifestos fora da raiz Git continuam abertas até auditoria.
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

Proposta. Este parent só fica `active` depois de `confirmed: heavy by owner`. Os children só existem depois de `approved: spec v1`.

Pré-condições, todas decisões do dono:
1. ~~N1 (nome e sequência)~~: decidida (a), 2026-10-17;
2. N2 (aceite da Etapa 1/1.1 em `61b58c2`): decidida (b); revisão adversarial concluída sem bloqueios materiais; aceite formal registrado em `2026-10-17-etapa-1-1-1-retro-acceptance` (`accepted`), merge da PR #19;
3. N3, N4, D1–D9c: direções escolhidas pelo dono e registradas no [registro arquitetural](../docs/plans/2026-10-17-decisao-arquitetural-publicacao-versionada.md); não autoriza implementação. Bloqueios finais: auditoria dos consumidores/schema, regra para manifesto fora da raiz Git, evidência semântica de `main.date` e testes do runner Windows.

Children previstos, em ordem de dependência:
1. `2026-10-17-gira-versioned-publication-01-adr-0014`: ADR 0014 consolidando as decisões aprovadas (2a).
2. `2026-10-17-gira-versioned-publication-02-gitcache-env`: leituras Git independentes do ambiente (2b). Pode sair como tarefa `standard` independente, antes da aprovação do spec, se o dono autorizar; nesse caso sai deste parent.
3. `2026-10-17-gira-versioned-publication-03-locator`: locator normalizado (2c), depende de 01.
4. `2026-10-17-gira-versioned-publication-04-keys`: chaves (2d), depende de 01, 02 e 03.
5. `2026-10-17-gira-versioned-publication-05-locks`: travas M, S e P (2e), depende de 01; pode correr em paralelo com 03 e 04.
6. `2026-10-17-gira-versioned-publication-06-publish`: publicação canônica e destacada (2f), depende de 04 e 05.
7. `2026-10-17-gira-versioned-publication-07-resolver`: resolvedor (2g), depende de 06.
8. `2026-10-17-gira-versioned-publication-08-cli`: `runIndex`/`runInit` (2h), depende de 05 e 07.
9. `2026-10-17-gira-versioned-publication-09-docs`: documentação e consumidores (2i), depende de 08.
10. Integrado: SC1–SC10, R1, V1 e revisão.

Uma etapa concluída não autoriza a seguinte.

## Review

skip: roda depois que os children estiverem verificados

## Checkpoint

- as of: 2026-10-17, branch `work/2026-10-17-gira-versioned-architecture-decisions`, base `db5a3cfa`
- done: N1 (a), N2 (b) com aceite formal registrado, O6 aplicada, direções N3/N4/D1–D9c registradas em decisão arquitetural
- status da spec: `draft`; implementação não autorizada e nenhuma tarefa-filha de implementação criada
- next: concluir e evidenciar os gates da spec; solicitar autorização heavy explícita só depois dessa evidência
- open: regra de D3 para manifesto fora da raiz Git; subdecisão D5 `main.date`; auditoria dos consumidores e schema; compatibilidade Windows para symlink e `os.DevNull`; M0–M6 em `tasks/2026-10-17-m0-m6-retro-review.md`
- revalidate: B1 e B2 rodaram em `61b58c2`; rodar de novo se `main` andar
- gates (2026-10-17): evidência dos gates 1–4 registrada em `docs/plans/2026-10-17-gates-tecnicos-publicacao-versionada.md`; gate 4 (Windows) evidência registrada no perfil runner admin; gate 5 aberto. Próximo: decisões do dono listadas no documento. Status segue `todo`; spec `draft`; implementação não autorizada.
