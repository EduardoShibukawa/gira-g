---
schema: task/1
tier: standard
status: verified
verified_at: "42ba885"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
---

# Eval end to end (SC1–SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Shared indexing unit used by the CLI and the evaluator; a hermetic mini-corpus materialized as a deterministic git repo; a Go perturbation; ground truth v2 (no macro_map, `reviewed` flag); measured grep baseline; real Unsupported Safety; index determinism; gate semantics.

In scope: spec §C items covered by SC1–SC5.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC2
- AC3 = SC3
- AC4 = SC4
- AC5 = SC5

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 6dacb25 | baseline before any edit (code unchanged through 41fd67b) |
| AC1 | `go test -race ./internal/eval -run EndToEnd and -run Mutation` | pass | 42ba885 | the evaluator indexes a git repo through indexer.Build; removing Go AST lowers contract recall 0.33 → 0.17 |
| AC2 | `go test -race ./internal/eval -run Baseline` | pass | 42ba885 | baseline measured on the corpus; the constants are gone |
| AC3 | `go test -race ./internal/eval -run UnsupportedSafety` | pass | 42ba885 | |
| AC4 | `go test -race ./internal/indexer ./internal/eval -run Deterministic and -run EndToEnd` | pass | 42ba885 | |
| AC5 | `go test -race ./internal/eval -run GateSemantics` | pass | 42ba885 | |
| R1 | `go test -race ./...` | pass | 42ba885 | |
| R2 | `node tools/validate/validate.js` | pass | 42ba885 | |
| V1 | read the diff against AC1–AC5 | pass | 42ba885 | scope: eval, corpus, indexer, tokens, cmd/gira index, fixtures, testdata |

## Outcome

Delivered:
- a shared `indexer.Build`, used by `gira index` and by the evaluator;
- `corpus.Materialize` and `corpus.Perturb`;
- a vendored 3.1.0 mini-corpus with a NOTICE file;
- ground truth v2, derived from the real sources (`reviewed: false`), plus a perturbation recipe;
- a measured grep baseline, real Unsupported Safety, index determinism checked over two cold runs, and the Enforce gate semantics.

Found and fixed: evidence `DiscoveredAt` used the wall clock, so indexes were never byte-stable. It is now stamped with the commit time.

Honest gate results (unreviewed ground truth, so reported only):

| Variant | Decision | Recall | Precision | Token reduction | Safety | Determinism |
|---|---|---|---|---|---|---|
| Published | No-Go | 64.6% | 34.4% | 75.2% | 100% | 100% |
| Perturbed | No-Go | 62.3% | 32.2% | 75.8% | 100% | 100% |

Main causes, recorded as follow-ups for the extractors, not fixed here:
- Dockerfile base images and stages emitted as library/service nodes;
- no Kafka topic contracts;
- no provider edges for non-Go services;
- briefings count `library:` nodes as upstream Components.

Spec divergence: SC1 names the Protobuf extractor, but the Go AST extractor also creates the contract nodes. Removing Protobuf alone leaves recall unchanged, so the mutation test removes Go AST instead (recorded on the parent).
