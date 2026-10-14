---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
verified_at: "87d9491"
---

# F5 depends_on corroboration by contract reference and accesses (J2 follow-up)

## Routing and approvals

- tier: standard (review follow-up F5 of parent 2026-10-14-gira-recall-joins; spec v1 approved; precision-preserving change inside internal/extractors)

## Objective

Spec J2 (specs/gira-recall-joins.md):

> A `depends_on`-only edge is emitted only with corroboration: a call-site,
> an env wiring, or a contract reference between the same pair. Bare ordering
> edges (`frontend→quote`, `frontend→image-provider`,
> `agent→product-catalog`) disappear; corroborated edges keep their locator,
> now citing both the `depends_on` line and the corroborating line.

`pruneOrderingOnlyEdges` corroborated only through evidence on the same
edge (same verb). This adds the two missing forms: a **contract
reference** (the source consumes a Contract the target provides) and an
**accesses** edge between the same pair resting on a non-depends_on line
(env wiring the resolver classified as accesses). The kept edge cites both
the depends_on line and the corroborating line(s). Recall and precision
must not drop.

In scope: `internal/extractors/resolve.go`, new
`internal/extractors/resolve_corroboration_test.go`. Out of scope:
`pipeline.go`, `internal/domain/`, `internal/slices/`, `cmd/`,
`internal/mcp/`, `layer2_compose*.go`, fixtures, testdata, ground truth.
`internal/eval/joins_test.go` unchanged.

## Acceptance criteria

- AC1 = contract-reference corroboration keeps the edge citing both lines; accesses corroboration keeps the edge citing both lines; an ordering-only edge (contract provided by a third service, accesses resting on the same depends_on line) is still pruned (test-first)
- AC2 = SC2 + SC5 regression guard: published ≥ R82.4/P100.0, perturbed ≥ R81.2/P100.0, SC15 assertiveness ≥ 90%, race suites green, validator exit 0

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/eval -run EndToEnd -v` before change | pass | ad6601b | published R82.4/P100.0 (token reduction 78.4%), perturbed R81.2/P100.0 (80.2%); worker measured at bc37716, orchestrator re-measured on the linear baseline (code == 3a39829) with identical numbers |
| AC1 | `go test ./internal/extractors -run DependsOn -v` (regex DependsOn or Corrobor; new `TestDependsOnCorroboratedByContractOrAccesses`) | pass | 87d9491 | red first at bc37716: "corroborated service:c>service:d was pruned", "corroborated service:a>service:b was pruned"; green after fix; `TestDependsOnNeedsCorroboration`, `TestDependsOnYAMLAliasAndMergeKey` still pass |
| AC2 | `go test ./internal/eval -run SC2DependsOnCorroboration -v` and `-run SC3UnsupportedPosture -v` | pass | 87d9491 | both PASS |
| R1 | `go test ./internal/eval -run EndToEnd -v` | pass | 87d9491 | published R82.4/P100.0, perturbed R81.2/P100.0 (unchanged); token reduction 78.3%/80.1%: agent→flagd and agent→otel-collector consumes edges are now kept, corroborated by env-wiring accesses (compose.agent.yaml L50/L32) and citing both lines |
| R2 | `(cd bench && go test ./layer2 -run SC15 -v)` | pass | 87d9491 | assertiveness 95.0% (focus frontend), 100.0% (unfocused) |
| R3 | `go test -race ./...` (root) | pass | 87d9491 | all packages ok |
| R4 | `go test -race ./...` (bench) | pass | 87d9491 | all ok, with `bench/layer0/testdata/gira/macro-map.json` copied in from the main checkout (gitignored, not committed) |
| R5 | `node tools/validate/validate.js; echo $?` | pass | 87d9491 | 0 (orchestrator, main checkout, linear history). The worker's run at b15bde4 exited 1 on a §6.1 false positive: the history walk linearises `rev-list HEAD -- tasks specs`, so parallel ticket branches read as record deletions. Resolved by linearizing followups/e2-review (cherry-picks), not by changing the validator |
| V1 | diff read against AC1–AC2 + spec J2 + F5 | pass | 87d9491 | only resolve.go + new test; corroborating lines are non-depends_on evidence of the accesses edge / the source's contract-consumption edge, deduplicated by locator; contracts visited in sorted order |

## Plan

skip: one-sentence diff extending the existing prune (failing test, then index accesses/consumed/provided edges and append corroborating evidence).

## Outcome

`pruneOrderingOnlyEdges` now also keeps an ordering-only service→service
consumes edge when an accesses edge between the same pair, or a Contract
the source consumes and the target provides, corroborates it; the
corroborating lines are appended so the edge cites both. Metrics hold:
R82.4/P100.0 published, R81.2/P100.0 perturbed, SC15 95.0%. Verified by the orchestrator after linearization.

- Linearized 2026-10-14 by the orchestrator: originally built at b15bde4 (on the merge-based bc37716) in an isolated worktree; cherry-picked on top of F4+F6+F7 as 87d9491 and re-verified there in the main checkout with no worktrees present: validator exit 0, race root ok, bench race ok incl. quarantine, EndToEnd + SC1–SC3 pass (published R82.4/P100.0, perturbed R81.2/P100.0), SC15 95.0%/100.0%.
