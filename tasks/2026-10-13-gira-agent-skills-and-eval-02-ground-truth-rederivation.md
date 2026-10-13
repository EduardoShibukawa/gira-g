---
schema: task/1
tier: standard
status: completed
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-01-eval-end-to-end]
---

# Ground truth rederivation (GT review)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Re-derive the ground truth against the full pinned OpenTelemetry Demo (opt-in run through the git cache). A human review sets `reviewed: true`.

In scope: spec §C items covered by GT review.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1: an opt-in full-corpus test exists and skips with a visible reason when it is not enabled
- AC2: the ground-truth locator anchors resolve at the pinned SHA [inferred]

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | d627d93 | all 12 packages ok before any edit; validator pass |
| AC1 | `go test -race ./internal/eval -run FullCorpus -v` without the opt-in env var | not run | | |
| AC2 | `go test -race ./internal/eval ./internal/corpus -run Anchor and -run FetchUpstream` (hermetic snapshot) | not run | | |
| R1 | `go test -race ./...` | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| R3 | opt-in `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus -v` over the upstream pinned SHA | not run | | free run, no models; network fetch from github.com/open-telemetry pinned by SHA |
| V1 | read the diff against the ACs | not run | | |

## Plan

skip: two independent additions (opt-in full-corpus fetch, anchor checker); order does not matter.

## Outcome

Delivered:
- `corpus.FetchUpstream` and `corpus.Leaks`: fetch the pinned upstream through the git cache and drop `README*`, `AGENTS.md` and `docs/` before indexing;
- `eval.FullCorpusSkipReason`, `eval.FetchFullCorpus` and `TestFullCorpusOptIn`: opt-in with `GIRA_EVAL_FULL_CORPUS=1` (cache dir in `GIRA_EVAL_CACHE_DIR`), otherwise a visible skip naming the variable, URL and SHA;
- `eval.CheckAnchors`: every anchor resolves in the vendored snapshot (always on) and, opted in, at the upstream SHA;
- the ground truth pins `upstream_url`, and its `review_note` records the re-check. It stays `reviewed: false`.

Opt-in full-corpus run (R3), free, no models: 596 files at `dedc017`. Every anchor resolves upstream. The vendored snapshot is byte-identical to upstream. Every expected id names a Component, proto service or topic that exists in the full tree.

| Variant | Decision | Recall | Precision | Token reduction | Safety | Determinism |
|---|---|---|---|---|---|---|
| Published, full corpus | No-Go | 64.6% | 31.1% | 99.3% | 100% | 100% |

The full-corpus grep baseline is about 2.2M tokens per task, because the one-hop rule reads every Component whose name appears in a matched file. So the 99.3% token reduction overstates the gain. Layer 1, arm A, is the real agent baseline.

For the owner to review before `reviewed: true` (not set by the agent):
1. All 12 tasks: confirm each expected set against the sources at `dedc017`, under the evidence rule in `review_note`.
2. task-07 (product-catalog): `src/product-catalog/opamp.go` connects to `service:opamp-server`. That edge is not in the ground truth. Add it, or confirm it is ambient like flagd and add it to `ignore`.
3. task-02 (frontend): `compose.yaml` makes frontend depend on `quote`, but no frontend source calls quote. The ground truth leaves quote out, which is consistent with its rule. Confirm.
4. task-02: the hermetic snapshot vendors no source for ad, recommendation or quote, so hermetic recall can't reach 100% on this task. Decide whether that is acceptable or whether to vendor one file each.
5. task-05/06: `contract:kafka:orders` is never emitted, which is a known extractor gap from child 01. The ground truth is right.

Follow-up, outside this spec: upstream also ships `CLAUDE.md`, an agent instruction file. The leak rule in spec §B lists only `README*`, `AGENTS.md` and `docs/`, so `CLAUDE.md` is kept. Recommend adding it to the rule (it matters for Layer 1 agents).

