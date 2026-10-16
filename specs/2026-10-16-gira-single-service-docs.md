---
schema: spec/1
status: draft
version: 1
---

# GIRA single-service docs (single-repo first-class + doc humano/agente)

## Problem

Gira-G (`gira`) prova valor no multi-service (Macro Map + slices com budget), mas sozinho num serviço só ele é raso: o mapa para em `Service/Library/Contract` e não contextualiza quem abre o repo pela primeira vez — humano ou agente.

Quem chega num serviço sozinho quer: o que ele faz, entrypoints, contratos que provê/consome, como rodar, onde mexer. Hoje isso exige grep + leitura manual (tokens altos, lento). Graphify entrega symbol/file-graph e Understand Anything entrega inventory + tour/layers — ambos cobrem o intra-serviço que a gira não cobre.

Heavy triggers:

- **H1:** product calls em granularidade intra-serviço, formato do doc, zero-config local.
- **H3:** atravessa manifest/indexer, extrator Go, slices/engine, CLI/MCP, export.
- **H5:** quatro frentes (modo local, detalhe intra-serviço, doc gerado, prova de token).

## Expected behavior

Um serviço sozinho indexado gera um doc útil para humanos e o mesmo conteúdo como slice para agentes, com budget medido. Multi-service continua como plus, sem regressão.

### F1 — Modo single-repo via `init`

`gira init [dir] [--force]` é a porta zero-config: gera `gira.yaml` a partir do repo local (HEAD como SHA pinado) se ausente — com `--force` repina — e em seguida indexa. `gira index` continua lendo o manifest como hoje (nenhuma mudança de comportamento com manifest explícito). Nada de escrita silenciosa fora do `init`.

### F2 — Detalhe intra-serviço determinístico (subset do Graphify, fase 1 = Go)

Fase 1 traz o subset determinístico do Graphify para Go: símbolos (packages, types, funcs) + `imports`/`calls` com locator file+line, `origin` discreta (`declared/extracted/inferred`), IDs estáveis, uma edge por evidência (sem colapso). Reusa o que o prior-art manda copiar (fragmento validado, canonical JSON, cache por hash+versão, external stubs, budget com proveniência) e diverge onde o prior-art manda evitar: sem LLM emitindo `EXTRACTED`, sem 1-edge-por-par, sem confidence contínuo, sem communities como arquitetura (`docs/prior-art/graphify.md:138`). Outras linguagens e wiki semântica são follow-up, nunca chute: sem extrator vira `unsupported`/`unknown`.

### F3 — `gira doc <service>` humano + agente, mesmo conteúdo (MVP: md + json)

Novo comando `gira doc <service-id> [--format md|json]`:

- `md`: onboarding legível (o que é, entrypoints, símbolos, contratos, dependências, como rodar/testar a partir de manifests, onde mexer, known unknowns). Reusa `tour`/`flows`/`explain` por baixo.
- `json`: o mesmo conteúdo em envelope orçado para MCP/agente (`get_doc` espelha o comando).
- Markdown e JSON contam a mesma história. `html` single-file fica como follow-up (padrão ADR 0008 quando entrar).

### F4 — Prova de economia

Cada `doc` emite budget (tokens estimados, bytes/4) e o bench compara contra baseline grep/cat: mesmo onboarding via arquivos crus custa N× mais. Sem número caçado — medido e reportado.

## Scope

**In scope**

- F1–F4 acima, tudo aditivo (sem mudar significado de flag/campo existente; envelopes e `--format json` intactos).
- Golden(s) para `doc` plain + teste de paridade md/json.
- MCP tool `get_doc` espelhando o CLI.

**Out of scope**

- Novos extratores de linguagem (TS/Python/etc são postura `unsupported` honesta, como hoje).
- Reescrita de extratores; joins novos além do necessário para F2.
- Skills distribution/release, docs site, push/release.
- Ground-truth changes, `reviewed: true`.

**Non-goals**

- Paridade com Graphify no nível símbolo/função; communities/Louvain; resumos LLM.
- Meta de recall/precision; números são reportados, nunca caçados enfraquecendo asserts.

## Acceptance criteria

- **SC1 (init).** `gira init <dir>` sem `gira.yaml` gera manifest pinando HEAD + indexa; com manifest explícito, `gira index` funciona como antes; `--force` repina. Teste: `go test ./cmd/gira -run SCSingleInit`.
- **SC2 (intra-serviço).** `doc` de um serviço Go lista entrypoints, símbolos, contratos providos/consumidos e imports intra-módulo vs externos, cada item com locator `[repo@commit:path#Lstart-Lend]`; o que não tem extrator aparece em Known Unknowns, nunca inventado. Teste: `go test ./internal/slices -run SCDocIntraService`.
- **SC3 (paridade humano/agente).** `doc --format md` e `--format json` cobrem os mesmos fatos (mesmos IDs de nodes/edges/evidence). Teste: `go test ./cmd/gira -run SCDocParity`.
- **SC4 (economia).** `doc` carrega budget honesto e o bench reporta razão vs baseline cru sem regredir determinismo/segurança. Teste: `go test ./internal/slices -run SCDocBudget` + bench doc-economy.
- **SC5 (regressão).** `go test -race ./...` root e bench passam; EndToEnd published/perturbed R/P não caem; `fixtures/ testdata/` intocados salvo goldens novos; validator exit 0.

## Interfaces and dependencies

Tudo aditivo. Nenhum flag/campo/tool existente muda de significado.

| Interface | Change |
|---|---|
| CLI | novo `init [dir] [--force]` (gera manifest + indexa); novo `doc <service> [--format md\|json]` |
| MCP | nova tool `get_doc` (mesmo engine do CLI) |
| Slices/engine | `GetDoc(serviceID)` + renderers md/json sobre dados já existentes + F2 |
| Indexer/manifest | construtor de manifest local a partir do HEAD (`init`); `index` inalterado |
| Eval/bench | novo bench doc-economy (quarentena respeitada); sem mudar ground truth |

**ADRs.** ADR 0001 vale (nada de model client; doc é template determinístico). ADR 0008 vale (CLI rico + static export, sem daemon/TUI).

## Decisions

- Q1: template determinístico em Go, sem LLM no harness (ADR 0001). Rejeitado: resumos gerados por modelo.
- Q2: granularidade fase 1 = símbolos + calls/imports Go com evidência (subset determinístico do Graphify). Rejeitado: symbol-graph completo multi-lang e communities nesta spec (follow-up).
- Q3: `doc` reutiliza `tour`/`flows`/`explain`/slices em vez de pipeline própria. Rejeitado: gerador paralelo que diverge.
- Q4: `init` gera manifest + indexa; `index` não escreve nada. Rejeitado: escrita silenciosa fora do `init` e `index .` com manifest efêmero.
- Q5: economia via budget `bytes/4` + bench comparativo, não promessa de N×. Rejeitado: número de marketing sem medição.

## Risks and constraints

- **Escopo rastejante para symbol-graph full.** Trava: F2 é subset Go determinístico; outras langs e semântica LLM precisam de spec nova.
- **Doc que parece verdade.** Mitiga: tiers discretos, commit indexado, `stale`, Known Unknowns sempre visíveis.
- **Styling churn.** Plain path golden-tested byte a byte (padrão E4); styling nunca vaza para JSON.
- **Compatibilidade.** Tudo aditivo; `git revert` por child basta. Sem migração ou estado externo.

## Verification

- Por child + integrado: B1 `go test -race ./...` (root); B2 `go test -race ./...` (bench/); B3 `go test ./internal/eval -run EndToEnd -v` (published/perturbed R/P); B4 os SC acima; B5 `node tools/validate/validate.js`.
- AC rows mapeiam SC1–SC5 para os `go test` acima; `not run` proibido no seal do parent.
- **Review focus:** cada fato do doc tem locator? `md` e `json` divergem em algum ID? `index .` toca em comportamento com manifest? Algum assert enfraquecido ou chute nomeado?

## Rollback

`git revert` por child. Sem migração, publicação ou estado externo. Aditivo; reverter restaura comportamento macro-only.

## Open questions

- deferred: `html` single-file do `doc` (ADR 0008 quando entrar). Safe: md + json cobrem humano/agente no MVP.
- deferred: profundidade do bench economy (fixture própria vs corpus existente). Safe: decidir no child, sem tocar SC.
- deferred: `doc` default `--format` (md humano vs json agente). Safe: exigir flag explícita no MVP.

## Revisions

- v1 (2026-10-16, draft, unapproved): primeiro draft a partir da discussão single-service vs multi-service plus.
