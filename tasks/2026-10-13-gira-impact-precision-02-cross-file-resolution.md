---
schema: task/1
tier: standard
status: verified
verified_at: "fa772f7"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
depends_on: [2026-10-13-gira-impact-precision-01-library-and-dockerfile-noise]
---

# Cross-file resolution step and compose naming of Dockerfiles (cause 6)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Add the deterministic resolution step of ADR 0013, vendor the upstream `.env` into the hermetic corpus, and attribute each Dockerfile to the compose service that builds it.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC2
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval | pass | 2319893 | published recall 58.8%, precision 83.3%; perturbed 56.2%, 81.8% (after child 01) |
| AC1 | `go test -race ./internal/eval -run SC2Dockerfiles` and `go test -race ./internal/extractors` (EnvFile, Interpolate, Attributes, Unresolved tests) | pass | fa772f7 | failed first on 2319893: `service:src` in both variants, cart and basket not citing their Dockerfile |
| AC2 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | fa772f7 | published: recall 58.8% → 58.8%, precision 83.3% → 83.3%; perturbed: 56.2% → 56.2%, 81.8% → 81.8%; token reduction 75.6% → 76.2% and 78.1% → 78.6%, because the grep baseline grew with `.env` |
| AC3 | same EndToEnd run, two cold indexes per variant (SC8) | pass | fa772f7 | safety 100%, determinism 100%; go.mod unchanged |
| AC4 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | fa772f7 | the extractor-count test now counts 12 (EnvFile added); goldens regenerated, diff explained below |
| AC5 | `node tools/validate/validate.js` (SC10) | pass | fa772f7 | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | fa772f7 | full corpus: recall 60.3%, precision 78.8% (unchanged); the vendored `.env` is checked byte-identical to upstream |
| R2 | `go test ./layer2 -run SC15 -v` in bench/ | pass | fa772f7 | focused 5.9% (1/17) unchanged; full tour 17 steps instead of 18 (`service:src` gone) |
| V1 | read `git diff 2319893..fa772f7` against SC2 and ADR 0013 | pass | fa772f7 | resolution step pure and sorted; unresolved variables attribute nothing |

## Outcome

Delivered:
- **The resolution step** (`resolve`, ADR 0013). Extractors that implement `HintEmitter` emit hints next to their facts. The pipeline runs one sorted, pure join after per-file extraction.
- **`EnvFileExtractor`.** It claims `.env` and emits `env_value` hints, read the way Compose reads them (`export`, quotes, inline comments). Compose-style interpolation handles `${VAR}`, `${VAR:-d}`, `${VAR-d}`, `$VAR` and nested `.env` references. A value with any undefined variable is refused whole.
- **Build hints.** Compose emits `compose_build` hints, located at the `dockerfile:` line. A Dockerfile emits the Service it produced. That Service merges into the compose service that builds it, so `src/cart/src/Dockerfile` becomes `service:cart` (and `service:basket` in the perturbed variant).
- **The vendored `.env`.** It is the upstream `.env`, blob `3b80f698`, identical at `upstream_sha`. The perturbation renames it consistently.

Golden diffs:
- **Corpus SHA.** It changed in every golden, because the materialized corpus gained `.env`.
- **Cart.** `cart` carries its Dockerfile metadata. The 1500-token `slice-cart-contracts` drops one more neighbor (`service:flagd`).
- **Rename score.** One inferred rename score changed (`cart → basket`, 2.25 → 2.75), because the node now carries the Dockerfile evidence too.
