---
schema: task/1
tier: heavy
status: active
spec: 2026-10-16-gira-single-service-docs@1
---

# GIRA single-service docs (INTENT: single-repo first-class + doc humano/agente)

## Routing and approvals

- tier: heavy (H1: product calls em granularidade/formato/zero-config; H3: manifest/indexer→slices→CLI/MCP; H5: quatro frentes)
- confirmed: heavy by owner (2026-10-16, "sim")
- approved: spec v1 by owner (2026-10-16, "ok")

## Objective

Fazer o gira-g servir um serviço sozinho tão bem quanto serve o multi-service (plus): `index .` zero-config, detalhe intra-serviço determinístico, e `gira doc <service>` gerando o mesmo onboarding para humanos (md/html) e agentes (json/MCP) com budget medido.

In scope: F1–F4 do spec `2026-10-16-gira-single-service-docs@1`.
Out of scope: novas linguagens, rewrites de extratores, skills/release, docs site, ground-truth changes (ver spec).

## Acceptance criteria

SC1–SC5 do spec.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test -race ./...` (root) antes de qualquer child | pass | c31c69c | base limpa registrada pelo child 02: `go test ./...` root verde (17 pkgs, exit 0); falha TestSC11 vista pelo child 01 em 03f12f9 não reproduz (transiente; child 03 confirma) |
| B2 | `go test -race ./...` (bench/) antes de qualquer child | fail | b959a04 | pré-existentes registrados pelo child 04, sem relação com o stack: layer0 `TestAdaptersGolden` (falta `testdata/gira/macro-map.json`); quarantine `TestNoReferencesFromGuardedFiles` (2 violações: docs.yml + `intra_service.go` Graphify) |
| B3 | `go test ./internal/eval -run EndToEnd -v` antes (R/P published/perturbed) | not run | | nenhum child registrou eval na base; âncora é a tabela Problem de `specs/gira-recall-joins.md` (published hermetic 82.4/90.3) — head comparado direto contra ela em B3-equivalente |
| SC1 | `go test ./cmd/gira -run SCSingleInit` | pass | dbdc40a | init gera manifest pinando HEAD + indexa; `--force` repina; erro honesto fora de git |
| SC2 | `go test ./internal/slices -run SCDocIntraService` | pass | dbdc40a | entrypoints/símbolos/contratos/imports intra-vs-externos com locator `[repo@commit:path#Lstart-Lend]`; helper.py segue em Known Unknowns com unknown joins preservados |
| SC3 | `go test ./cmd/gira -run SCDocParity` | pass | dbdc40a | md cobre todo ID/locator do json (`docFactTokens`); repeats byte-identical; golden plain `doc-acme.md`; banner só em stderr/md |
| SC4 | `go test ./internal/slices -run SCDocBudget` + bench economy | pass | dbdc40a | envelope honesto (`bytes/4`, `measured:false`); `BenchmarkDocEconomy` mede raw/doc=0.10 (doc 9691B/2423tok vs raw cat 932B/233tok no fixture micro — medido, sem threshold; ver Review finding 2) |
| SC5 | regressão integrada (B1+B2+B3 no head integrado, validator) | fail | dbdc40a | só pelo bench/: root `-race` verde, eval sem queda, validator pass; bench/ falha só nos pré-existentes de B2 (layer0 + quarantine docs.yml) — ver R1 |
| R1 | `go test -race ./...` root+bench no head integrado | fail | dbdc40a | root: exit 0, 17 pkgs ok, NO-FAIL com `-race` (pass isolado). bench `-race`: layer1+layer2 ok; layer0 e quarantine falham identicamente a B2 (arquivos intocados pelo stack), MENOS 1 violação que o stack introduzia (`intra_service.go:22` Graphify, do child 02) — corrigida em dbdc40a; quarantine agora só acusa docs.yml |
| V1 | diff integrado lido contra SC1–SC5 | pass | dbdc40a | lido c31c69c...dbdc40a: 02 (hints+resolve+engine intra) + 03 (runDoc+GetDoc/Emit/RenderMarkdown+golden) + 04 (get_doc MCP+budget/bench) + fix quarantine; `fixtures/ testdata/` intocados salvo 1 golden novo (`cmd/gira/testdata/human/doc-acme.md`); sem asserts enfraquecidos |

## Plan

Spec em `draft` v1 — sem children até aprovação (spec gate). Após `approved: spec v1`:

1. `2026-10-16-gira-single-service-docs-01-local-init`: F1 `init` gera manifest + indexa.
2. `2026-10-16-gira-single-service-docs-02-intra-service`: F2 detalhe intra-serviço.
3. `2026-10-16-gira-single-service-docs-03-doc-command`: F3 `doc` CLI + renderers md/html/json.
4. `2026-10-16-gira-single-service-docs-04-doc-mcp-budget`: F3 MCP `get_doc` + F4 budget/bench.
5. Integrado: SC1–SC5 + R1 + V1 + review.

## Review

Fresh-context review do diff integrado c31c69c...dbdc40a contra o spec v1
(foco da spec: locators, divergência md/json, `index` com manifest, asserts
enfraquecidos, chutes). Lidos na íntegra: `internal/slices/doc.go`,
`internal/slices/intra_service.go`, diff de `cmd/gira/main.go` (runDoc),
`internal/mcp/server.go` (get_doc), `internal/slices/engine.go`
(cloneNode), `internal/tools/registry.go`, `TestSCDocParity` +
`docFactTokens`, `runInit`.

1. **Locators (SC2).** Todo item da lista do SC2 (entrypoints, packages,
   types, funcs, contracts provided/consumed, imports intra/externos/
   stdlib, calls) carrega locator file+line; `RenderMarkdown` imprime todos
   e `docFactTokens` exige cada ID+locator no md. `ChangeDirs` e
   `Description` não têm locator próprio, mas não estão na lista de itens
   do SC2 — são derivados de registros que têm (entrypoints/packages) ou
   metadado do nó. Aceito sem mudança.
2. **Paridade md/json (SC3).** Genuína: `docFactTokens` cobre calls,
   unknowns (+`UnknownJoins`), run tasks, change dirs, tour e flows — não
   só o miolo; repeats byte-identical nos dois formatos; golden plain
   byte-for-byte. JSON sai cru no stdout (`humanBanner` só dispara p/
   md e escreve no stderr; `emitHuman` só decora md). Sem divergência.
3. **`index` com manifest (F1/Q4).** O stack não toca o caminho do
   `index`; `runInit` (base c31c69c, child 01) só escreve `gira.yaml`
   quando ausente (`--force` repina), erra honestamente fora de git e
   avisa de worktree suja. Sem escrita silenciosa.
4. **Sem chute, sem assert fraco.** `helper.py` (sem extrator) fica em
   Known Unknowns com unknown joins preservados e nenhum fato cunhado a
   partir dele (teste dedicado). `html` e `--format` ausente erram
   honestamente. `Emit` com budget estourado retorna `ErrBudgetExceeded`
   em vez de truncar silenciosamente. Nenhum assert enfraquecido
   encontrado.
5. **Finding 1 (corrigido em dbdc40a).** O child 02 introduziu violação
   NOVA de quarantine: `internal/domain/intra_service.go:22` citava o
   nome literal da ferramenta comparativa ("Graphify's confidence_score"),
   que o `forbidden` de `bench/quarantine/check.go` barra em qualquer
   `.go` do módulo. O B2 do child 04 a registrou como pré-existente (era,
   na base dele), mas contra `main` ela é deste stack. Reword para
   "unlike symbol-graph prior art"; quarantine agora só acusa o
   pré-existente `docs.yml:33` (arquivo intocado pelo stack).
   Re-verificado: `internal/domain` ok + `TestNoReferencesFromGuardedFiles`
   sem a violação.
6. **Finding 2 (aberto, sem mudança).** A prosa de F4 ("via arquivos crus
   custa N× mais") não se confirma na medição: `BenchmarkDocEconomy`
   reporta raw/doc=0.10 — o doc estruturado (9691B) custa ~10× MAIS que o
   `cat` do fixture micro (932B, 4 arquivos). O teste trava medição, não
   threshold (Q5: sem número de marketing), então SC4 passa honestamente;
   mas o baseline "mesmo onboarding via crus" com 4 arquivos subestima o
   custo real (onboarding de verdade exigiria ler o serviço inteiro).
   É a open question "profundidade do bench economy" já deferida na spec.
7. **Eval (B3-equivalente).** `TestSuitePublishedEndToEnd`: recall 82.4%,
   precision 100.0%; `TestSuitePerturbedEndToEnd`: recall 81.2%,
   precision 100.0%. Contra a tabela Problem (`gira-recall-joins.md`:
   published 82.4/90.3, perturbed 81.2/89.7): recalls idênticos,
   precisions acima. Sem queda. (Modo reportado: "No-Go reviewed=false".)

## Outcome

Integrado em dbdc40a: SC1–SC4 passam no head; root `-race` verde (17 pkgs);
eval sem queda vs a tabela ancorada; validator pass; `fixtures/ testdata/`
intocados salvo o golden novo. Status segue `active` — só o owner aceita.
Decisões do owner antes do merge: (a) bench/ segue vermelho nos 2
pré-existentes de B2 (layer0 golden sem `testdata/gira/macro-map.json`;
quarantine docs.yml) — mergear com bench vermelho ou corrigir fora deste
stack; (b) Finding 2: aceitar a razão medida invertida como report honesto
(Q5) ou pedir bench mais fundo (fixture própria) em follow-up; (c) ordem de
merge: 01/spec/tasks estão na main local como commits não-pusheados
f117473..c31c69c — o PR contra `origin/main` carrega o stack inteiro.
