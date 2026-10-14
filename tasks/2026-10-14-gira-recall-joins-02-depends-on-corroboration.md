---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
depends_on: [2026-10-14-gira-recall-joins-01-health-denylist]
verified_at: "18d79b6"
---

# J2 depends_on corroboration (SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-recall-joins; spec v1 approved)

## Objective

A `depends_on`-only edge survives only with corroboration (call-site,
env wiring, or contract reference between the same pair), citing both
lines. Bare ordering edges (`frontend→quote/image-provider`,
`agent→product-catalog`) disappear with no recall loss.

In scope: SC2, test-first. Failing test first: an ordering-only edge
reaching the Macro Map must fail.
Out of scope: J1/J3/J4; ground-truth changes.

## Acceptance criteria

- AC1 = SC2
- AC2 = SC5 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/extractors -count=1` + validator before change | pass | a30415d | 01 verified; extractors + validator green |
| AC1 | `go test ./internal/extractors -run DependsOnCorroboration -v` + EndToEnd task-02/task-11 check | pass | 18d79b6 | failed first (ordering-only edge reached map), passes after prune; task-02 unexpected=[], task-11 pass=true P/R 1.00; P93.3→100.0 published / 92.9→100.0 perturbed, R unchanged |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | 18d79b6 | root 16 pkgs ok, bench ok, validator pass; recall 82.4/81.2 unchanged, safety/determinism 100% |
| R1 | `go test ./internal/extractors ./internal/eval ./cmd/gira -count=1` | pass | 18d79b6 | pipeline/dogfood/golden updates are J2-mandated (ordering-only out); tour golden now reaches cart via corroborated line |
| V1 | diff read against AC1–AC2 + spec J2 | pass | 18d79b6 | HintDependsOn + pruneOrderingOnlyEdges in resolve; infra accesses untouched |

## Plan

1. Failing pipeline test first (ordering-only edge must not reach the map; corroborated edge must, with both lines).
2. `HintDependsOn` from compose + `pruneOrderingOnlyEdges` in resolve.
3. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

J2 delivered: `HintDependsOn` (exact edge-evidence line) + 
`pruneOrderingOnlyEdges` in resolve (keeps iff non-depends_on evidence
exists; merged evidence cites both lines). task-02 `unexpected=[]`,
task-11 `pass=true`; precision 100.0/100.0, recall unchanged. 03 unblocked.
