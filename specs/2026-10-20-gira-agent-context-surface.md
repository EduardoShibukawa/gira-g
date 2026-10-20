---
schema: spec/1
status: draft
version: 1
---

# Agent context surface: formato caveman, teto do summary com intra-service e erros de índice legíveis

Derivada da INTENT 0001 (`intents/0001-llm-semantic-slices-and-micro-extractors.md`) para fechar o loop de dogfooding Pi Harness × Gira-G (SPEC 0007 do pi-harness). Cobrindo os gaps G1, G2 e G3 descobertos em dogfooding real, gravado em `pi-harness/benchmarks/*/results.jsonl`.

## Problem

- **G2 — `--format=caveman` não existe.** A INTENT 0001 especifica um formato 1-fato-por-linha para `slice`, `impact` e `locate`; hoje a CLI aceita apenas `json|md` (`cmd/gira/main.go`, flags `format`), e `slices.ParseFormat` (`internal/slices/budget.go`) mapeia **qualquer** valor desconhecido silenciosamente para JSON. O Filtro Caveman vive client-side no pi-harness (`.pi/extensions/gira.ts`, `formatCavemanContext`), duplicando no consumidor uma responsabilidade que é do produtor do contexto. A queda silenciosa em JSON impede feature-detect confiável no bridge.
- **G1 — `budget_exceeded` no summary para nós Go com intra-service grande.** O binário instalado (0.1.0, `f5dad05`, 2026-10-15) reproduz: `gira slice service:internal --tier summary --budget 3000` falha com `budget_exceeded` porque o bloco never-dropped do `fit` (`internal/slices/budget.go`) carregava os metadados `go_intra_service` do nó focal. Revalidado na `main` em código: `cloneNode` (`internal/slices/engine.go`) já remove `domain.IntraServiceMetadataKey` de toda view derivada desde `89fda6f` (2026-10-16), e `slice --tier summary --budget 400` sobre `service:internal` (intra-service de ~1,6 MB no índice dogfood) passa. O comportamento está correto mas **não tem guarda de regressão**: nada nos testes de `internal/slices` garante que `go_intra_service` permanece dropped-able no summary e completo apenas via `GetIntraServiceDoc`/deep, nem preserva a honestidade do ceiling documentada em `tasks/2026-10-14-gira-envelopes-hygiene-02-*`.
- **G3 — erro de índice corrompido no cache não nomeia o arquivo nem sugere `--index`.** Com binário antigo, o cache `~/.cache/gira/indexes/005` produzia `json: cannot unmarshal number into Go struct field MacroMap.schema_version` puro. Na `main`, `schema_version` é `any` e a discovery ambígua já sugere `--index <path>`; mas no ramo de discovery com manifesto (`internal/store/resolver.go`, Branch 2) e em `internal/store/publish.go`, vários erros `ErrCorruptIndex` ("corrupted meta.json: %v", "cannot read meta.json: %v", "version %s missing meta.json: %v") nomeiam apenas a versão ou nada, sem caminho completo e sem sugestão de `--index`/reindexação.

## Expected behavior

- `gira slice`, `gira impact` e `gira locate` aceitam `--format=caveman` (alias `--format=llm`), renderizando texto puro 1-fato-por-linha com a gramática exata:
  - slice: `Component: <Kind>[<id>] @ <path>:L<n>` (nó focal; sem locator quando não há evidência com path), depois seções `Contracts:` (definições de contrato do nó focal quando presentes nos metadados: método/rota/assinatura/linha), `Dependencies:` (`- <verb> <target-id> @ <path>:L<n>` por relacionamento incidente ao focal) e `Impact:` (`- consumed by <Kind>[<id>] @ <path>:L<n>` para vizinhos que consomem o focal), cada item numa linha, sem markdown, sem AST, sem warnings;
  - impact: `Target: <Kind>[<id>]`, `Depth: <n>`, depois `Upstream:`, `Downstream:` e `Contracts:` com `- <Kind>[<id>] @ <path>:L<n>` por linha; modo `--from-diff` mantém `Target: diff <base>..<head>`;
  - locate: uma linha `Match: <Kind>[<id>] @ <path>:L<n>` por correspondência, na ordem canônica de `BoundaryResult`, e `Boundary: <path> <reason>` para unsupported boundaries.
  - fechando qualquer saída com `Budget: limit=<n> estimated=<n> truncated=<t>` (e `- escalate: <hint>` quando houver), o marker nunca é truncado.
- O formato caveman passa pelo **mesmo** `fit`/drop-order dos demais formatos: `--budget` é respeitado com truncatura por remoção de vizinhos/relacionamentos/evidência (nunca no meio de um fato), com envelope honesto (`estimator: "bytes/4"`, `measured: false`); a saída de um componente médio (nó `service:checkout` dos fixtures dogfood, tier summary) cabe em 400 tokens com o orçamento padrão do tier.
- `--format` com valor desconhecido em `slice`, `impact` e `locate` falha com erro nomeando os formatos aceitos (`json|md|caveman`); nenhum valor desconhecido cai silenciosamente em JSON nesses três comandos. Os demais comandos (`explain`, `tour`, `flows`, `diff`, `export`) mantêm o comportamento atual.
- Um `slice --tier summary` (JSON e caveman) de um nó Go com `go_intra_service` grande **deve caber em `--budget 400`**: a metadados intra-service continua dropped-able/ausente no summary, o tier deep continua servindo o detalhe completo via `GetIntraServiceDoc`, e uma regressão em `internal/slices` (padrão dos testes `Fit`/`SC6`) trava os três fatos.
- Erro de índice inválido/corrompido no cache de índices, em qualquer caminho de discovery (manifesto presente ou não), nomeia o arquivo exato em causa e sugere reindexar ou passar `--index <path>`; o código estruturado permanece `index_corrupt`.

## Scope

**Dentro:** `internal/slices` (formato caveman em slice/impact/locate + guarda G1), `cmd/gira` (flags, validação, help), `internal/store` (mensagens de `ErrCorruptIndex` na discovery), testes das três áreas, documentação do site na página de `slice`.

**Fora do escopo (fases futuras da INTENT 0001, não implementar):** `gira symbols <file>`; `--errors-only`; o Breaking Changes Gate da INTENT 0002; MCP `format=caveman` (o bridge hoje consome JSON client-side — paridade MCP fica como fase futura documentada em Decisions); PetClinic e qualquer repo fora de gira-g/pi-harness/llm-bench; push, PR ou release.

## Acceptance criteria

- SC1 (G2, gramática): as saídas caveman de `slice`, `impact` e `locate` sobre o índice dogfood seguem a gramática acima, byte-determinísticas entre execuções com o mesmo índice e orçamento.
- SC2 (G2, orçamento): `slice service:checkout --tier summary --format caveman --budget 400` do índice dogfood termina 0, com `estimated ≤ 400`; `--budget 100` trunca por remoção de vizinhos (nunca por corte no meio de linha) e mantém o marker `Budget:`.
- SC3 (G2, validação): `--format=caveman` é aceito nos três comandos; `--format=xml` (ou qualquer valor inválido) falha com mensagem contendo `json|md|caveman`; `llm` é alias de `caveman`.
- SC4 (G1, regressão): com um nó Service portador de `go_intra_service` de grande volume no índice, `slice --tier summary` renderiza em JSON e caveman sob `--budget 400` sem a chave `go_intra_service` na saída; o deep preserva o detalhe por `GetIntraServiceDoc`; testes em `internal/slices` falham se `cloneNode` deixar de stripar a chave (guarda de regressão explícita).
- SC5 (G3, erros legíveis): com `gira.yaml` no cwd e cache corrompido (meta.json ilegível, macro-map com hash divergente ou `.head` truncado), o stderr da CLI nomeia o caminho completo do arquivo corrupto e contém `--index`; o erro estruturado mantém `index_corrupt`.
- SC6 (higiene geral): `go test -race ./...` verde, goldens não tocados exceto regeneração explícita onde exigida, envelope honesty (`estimator`/`measured`) inalterada, zero mudança em `fixtures/`/`testdata/` de ground truth, validator pass.

## Interfaces and dependencies

- `internal/slices/budget.go`: `FormatCaveman Format = "caveman"`; `ParseFormat` mapeia `caveman|llm` → `FormatCaveman`; nova `ParseFormatStrict(s string) (Format, error)` para uso dos três comandos (aceita `json|md|markdown|caveman|llm`; `html` só via comandos próprios de export).
- `internal/slices/render.go`: `(*ContextSlice) RenderCaveman() string`, `(*ImpactBriefing) RenderCaveman() string`, `(*BoundaryResult) RenderCaveman() string` (BoundaryResult vive em `engine.go`; os renderers vão em `render.go`).
- `internal/slices/budget.go` / `engine.go`: os closures `render` de `ContextSlice.Emit`, `ImpactBriefing.Emit` e `BoundaryResult.Emit` reconhecem `FormatCaveman`; ordem de drop e marker de budget inalterados.
- `cmd/gira/main.go`: help `Output format (json|md|caveman)` em slice/impact/locate; validação estrita via `ParseFormatStrict`; demais comandos mantêm `ParseFormat`.
- `internal/store/resolver.go` (Branch 2) e `internal/store/publish.go`: mensagens de falha de carga na discovery incluem `: <caminho>` (quando ainda não incluído) e sufixo `; re-index with 'gira index', or read a specific file with --index <path>`; código `index_corrupt` preservado (`errors.Is`/prefixo `index_corrupt:`).
- Consumidor: pi-harness `.pi/extensions/gira.ts` fará feature-detect deste formato (SPEC 0007 Fase 2; fora deste repo).

## Decisions

- D1 — gramática é a da INTENT 0001, fixada pelo prompt de orquestração (`Component:/Contracts:/Dependencies:/Impact:`); o filtro client-side do pi-harness serve de referência de compactação mas o layout canônico é o do produtor.
- D2 — `ParseFormat` mantém default-JSON (legado para explain/tour/flows/diff/export e chamadas internas); a validação estrita é nova e alcança **apenas** slice/impact/locate, para não quebrar consumidores existentes (envelopes-hygiene 01 cobriu os demais).
- D3 — caveman não ganha JSON embutido nem campos novos: tudo que ele mostra já existe no envelope JSON; metadados ausentes (ex.: contracts em summary) simplesmente não aparecem, sem linhas vazias.
- D4 — paridade MCP (`format` nos tools do server) é fase futura: o bridge lê a CLI. Registrar em Open questions como F1.
- D5 — G1 é realizado na `main` (strip em `cloneNode`, `89fda6f`); o trabalho desta spec é a **guarda de regressão** e a documentação do contrato (summary dropped-able, deep completo via `GetIntraServiceDoc`), não um novo corte.
- D6 — versão do binário: rebuild local com `-ldflags` registrando a nova build; a instalação no PATH do pi-harness (`~/.local/bin/gira`) é parte da task integrada. Sem tag/release.

## Risks and constraints

- Goldens de `cmd/gira` podem conter bytes de saída JSON; mudanças aqui são aditivas de formato (caveman novo, validação estrita só nos 3 comandos) e não devem tocar goldens — se tocarem, regeneração explícita na task com diff revisado.
- Honestidade do ceiling (envelopes-hygiene 02): `Estimated` do caveman continua sendo `tokens.Estimate` dos bytes emitidos, com settle igual ao JSON/MD.
- Sem mudanças em ground truth (`fixtures/`, `testdata/`), sem `"reviewed": true`, sem modelo (ADR 0001), sem push/PR/release.
- Commits via `mise exec -- git commit` (lefthook fail-closed, CLAUDE.md); checks executados sozinhos, nunca encadeados com commit.

## Verification

- B1 — baseline: `go test -race ./...` + `node tools/validate/validate.js` antes de qualquer edição.
- SC1 — `go test ./internal/slices -run SCAgentCavemanGrammar -v -count=1` + saída real da CLI sobre o índice dogfood (`slice service:checkout`, `impact service:checkout`, `locate checkout`, `--format caveman`) inspecionada na task.
- SC2 — `go test ./internal/slices ./cmd/gira -run SCAgentCavemanBudget -v -count=1`.
- SC3 — `go test ./cmd/gira -run SCAgentFormatValidation -v -count=1`.
- SC4 — `go test ./internal/slices -run SCAgentSummaryIntraBounded -v -count=1`.
- SC5 — `go test ./internal/store ./cmd/gira -run SCIndexCorruptError -v -count=1`.
- R1 — `go test -race ./...` (root) ao fim de cada task; R2 — validator `node tools/validate/validate.js` na task integrada; V1 — leitura do diff contra os SCs por task.

## Rollback

`git revert` por task (uma por commit, merges `--no-ff` locais); o formato novo é aditivo — reverter os merges remove caveman e restaura a validação antiga sem tocar envelopes existentes.

## Open questions

- F1 (fase futura): expor `format=caveman` nos tools MCP para paridade com a CLI? — adiado, bridge consome CLI.
- F2 (fase futura, INTENT 0001): `gira symbols` e `--errors-only` permanecem fora desta spec por decisão do dono (rodada de benchmark).

## Revisions

- v1 (2026-10-20): draft inicial derivado da INTENT 0001 após revalidação de G1/G2/G3 na `main` (G1 já sanado em `89fda6f`; G3 parcialmente sanado pós-`8102fec`).
