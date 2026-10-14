---
schema: task/1
tier: standard
status: verified
verified_at: "ee7b029"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
depends_on: [2026-10-13-gira-impact-precision-02-cross-file-resolution]
---

# Kafka topic Contracts (cause 2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Emit `contract:kafka:<topic>` with the Go producer (extracted), the broker (declared) and the env-wired consumers (inferred); keep the Perturbed Variant coherent with one recipe edit.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC3
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval | pass | dbbf741 | published recall 58.8%, precision 83.3%; perturbed 56.2%, 81.8% (after child 02) |
| AC1 | `go test -race ./internal/eval -run SC3Kafka` and `go test -race ./internal/extractors` (KafkaTopic tests) | pass | ee7b029 | failed first on dbbf741: no `contract:kafka:orders` or `contract:kafka:order-events` |
| AC2 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | ee7b029 | published: recall 58.8% → 77.9%, precision 83.3% → 86.9%; perturbed: recall 56.2% → 76.6%, precision 81.8% → 86.0%; tasks 05 and 06 pass in both |
| AC3 | same EndToEnd run, two cold indexes per variant (SC8) | pass | ee7b029 | safety 100%, determinism 100%; go.mod unchanged |
| AC4 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | ee7b029 | SC6 tightened to fail on a missing target, and it runs 12 of 12 tasks; `TestBudgetOnMissingTargetEmitsNothing` keeps its assertion on a topic that does not exist; SC9 unchanged and passing |
| AC5 | `node tools/validate/validate.js` (SC10) | pass | ee7b029 | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | ee7b029 | full corpus: recall 60.3% → 79.4%, precision 78.8% → 83.1% |
| R2 | `go test ./layer2 -run SC15 -v` in bench/ | pass | ee7b029 | assertiveness unchanged (5.9%); order-flow coverage still 81.8%, see Outcome |
| V1 | read `git diff dbbf741..ee7b029` against SC3 | pass | ee7b029 | origins: producer extracted (plus declared env line), broker declared, consumers inferred; unresolved topic emits nothing (unit test) |

## Outcome

Delivered:
- **Go hints.**
  - Env bindings: `os.Getenv` assignments and `f(&field, "VAR")`.
  - Package-level env variables, including a function that returns `os.Getenv` with a literal fallback.
  - Sarama `ProducerMessage` topics: literal, env, package var or field.
  - Producer broker lists, followed one call into the constructor.
- **Compose hints.** Every `environment` entry, list or map form, each at its own line.
- **The Kafka join.**
  - `contract:kafka:<topic>`, provided by the producer: extracted at the `Topic:` line, plus the declared env line when compose names the topic.
  - Also provided by the broker the producer's address resolves to (declared).
  - Consumed by every other service wired to the same topic variable, with the same value (inferred).
- **Perturbation.** The recipe edits `.env` (`KAFKA_TOPIC=order-events`). `eval.Recipe.Apply` also maps the declared id renames. `gira diff` reports the topic as an inferred rename.
- **Golden diffs.** `impact-checkout` lists the consumers' unsupported code (`Consumer.cs`, `main.kt`). `diff-published-perturbed` gains the topic rename, and some inferred rename scores changed.

Not delivered: tour coverage of accounting and fraud-detection. The tour walks from a stop through a Contract to the Contract's providers only, so from checkout it reaches the broker, not the topic's consumers. Covering them needs a tour traversal rule ("a provided topic leads to its consumers"), which spec v1 does not include. It is raised on the parent for the owner.
