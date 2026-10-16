---
schema: task/1
tier: heavy
status: accepted
spec: 2026-10-16-gira-single-service-docs@1
verified_at: "4d39e54"
---

# GIRA single-service docs (INTENT: single-repo first-class + doc humano/agente)

## Routing and approvals

- tier: heavy (H1: product calls em granularidade/formato/zero-config; H3: manifest/indexer→slices→CLI/MCP; H5: quatro frentes)
- confirmed: heavy by owner (2026-10-16, "sim")
- approved: spec v1 by owner (2026-10-16, "ok")
- accepted: by owner (2026-10-16, "vamos de 1")

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
| B2 | `go test ./...` (bench/) na base pré-children | fail | b959a04 | baseline real, registrada pelo child 04: layer0 `TestAdaptersGolden` (falta `testdata/gira/macro-map.json`) + quarantine `TestNoReferencesFromGuardedFiles` (docs.yml + `intra_service.go`). Ambas corrigidas em 04262da e re-verificadas verdes em SC5/R1 |
| B3 | `go test ./internal/eval -run EndToEnd` na base 03f12f9 (worktree limpa pré-children) | pass | 03f12f9 | published hermetic recall 82.4 / precision 100.0 — âncora contra a qual o head foi comparado (sem queda) |
| SC1 | `go test ./cmd/gira -run SCSingleInit` | pass | 4d39e54 | init gera manifest pinando HEAD + indexa; `--force` repina; erro honesto fora de git (re-rodado no head mergeado) |
| SC2 | `go test ./internal/slices -run SCDocIntraService` | pass | 4d39e54 | entrypoints/símbolos/contratos/imports intra-vs-externos com locator `[repo@commit:path#Lstart-Lend]`; helper.py segue em Known Unknowns com unknown joins preservados (re-rodado no head mergeado) |
| SC3 | `go test ./cmd/gira -run SCDocParity` | pass | 4d39e54 | md cobre todo ID/locator do json (`docFactTokens`); repeats byte-identical; golden plain `doc-acme.md`; banner só em stderr/md (re-rodado no head mergeado) |
| SC4 | `go test ./internal/slices -run SCDocBudget` + bench economy | pass | 4d39e54 | envelope honesto (`bytes/4`, `measured:false`); `BenchmarkDocEconomy` mede raw/doc=0.10 (doc 9691B/2423tok vs raw cat 932B/233tok no fixture micro — medido, sem threshold; aceite do owner em 2026-10-16) |
| SC5 | regressão integrada (B1+B2+B3 no head integrado, validator) | pass | 4d39e54 | root `-race` verde, bench/ verde (fix 04262da), eval sem queda (82.4/100.0 published, 81.2/100.0 perturbed vs 82.4/90.3 e 81.2/89.7), validator pass; `fixtures/ testdata/` intocados salvo goldens novos |
| R1 | `go test -race ./...` root+bench no head integrado | pass | 4d39e54 | root exit 0 (17 pkgs; `TestSCSBOMAndCosign` precisou de cosign no env local via `mise use -g cosign@3.1.3`, verde em seguida); bench exit 0; CI da PR verde nos 3 OSes + validator |
| V1 | diff integrado lido contra SC1–SC5 | pass | 4d39e54 | leitura integral c31c69c...dbdc40a no seal + delta 04262da relido (3 arquivos: fixture golden restaurada via `git add -f`, quarantine movido p/ `website/scripts/` sem enfraquecer o guard); `fixtures/ testdata/` intocados salvo goldens novos; sem asserts enfraquecidos |

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

Aceito pelo owner em 2026-10-16 ("vamos de 1"), mergeado em 4d39e54 (PR #16, CI verde nos 3 OSes + validator). Entregue: `gira init`, detalhe intra-serviço Go, `gira doc md|json`, MCP `get_doc`, budget honesto + bench economy (raw/doc=0.10 medido, aceite Q5). Follow-ups (fora deste epic): INTENT CLI-contracts (deferido "Depois"), polish do doc (`html`, `--format` default, snippet em `skills install`), dogfooding com corpus real para rever a economia.
