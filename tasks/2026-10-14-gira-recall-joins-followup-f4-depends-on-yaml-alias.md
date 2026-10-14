---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
verified_at: "4e4d73c"
---

# F4 depends_on hints through YAML aliases and merge keys (J2 follow-up)

## Routing and approvals

- tier: standard (review follow-up F4 of parent 2026-10-14-gira-recall-joins; spec v1 approved; precision-only change inside internal/extractors)

## Objective

J2 prunes service→service `consumes` edges resting on compose `depends_on`
alone. `depends_on` reached through a YAML alias (`depends_on: *x`) or a
merge key (`<<: *x`, `<<: [*x]`) produced no hint, so such ordering-only
edges escaped the prune. Resolve aliases and merge keys in the depends_on
hint extraction (list and map form) so they are pruned like any other
ordering-only edge. Recall must not drop.

In scope: `internal/extractors/layer2_compose.go`, a new test in
`internal/extractors`. Out of scope: `resolve.go` (unchanged),
`pipeline.go`, `internal/domain/`, fixtures, testdata, ground truth.

## Acceptance criteria

- AC1 = aliased/merged depends_on (list and map form) yields depends_on hints and the ordering-only edge is pruned; a corroborated aliased edge survives citing both lines (test-first)
- AC2 = SC2 + SC5 regression guard: published ≥ R82.4/P100.0, perturbed ≥ R81.2/P100.0, SC15 assertiveness ≥ 90%, race suites green, validator exit 0

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/extractors -count=1` + validator before change | pass | 9ba80a6 | extractors ok; validator Overall: pass |
| AC1 | `go test ./internal/extractors -run DependsOn -v` (new `TestDependsOnYAMLAliasAndMergeKey`) | pass | 4e4d73c | red first at 9ba80a6: 5 missing hints (alias-list/alias-map/merge-list/merge-map/backend) + 4 ordering-only edges reached the map; green after fix; `TestDependsOnNeedsCorroboration` still pass |
| AC2 | `go test ./internal/eval -run SC2DependsOnCorroboration -v` | pass | 4e4d73c | PASS |
| R1 | `go test ./internal/eval -run EndToEnd -v` | pass | 4e4d73c | published R82.4/P100.0, perturbed R81.2/P100.0 (unchanged) |
| R2 | `(cd bench && go test ./layer2 -run SC15 -v)` | pass | 4e4d73c | assertiveness 95.0% (focus frontend), 100.0% (unfocused) |
| R3 | `go test -race ./...` (root) | pass | 4e4d73c | all packages ok |
| R4 | `go test -race ./...` (bench) | pass | 4e4d73c | ran with `-skip TestAdaptersGolden/gira`: its input `bench/layer0/testdata/gira/macro-map.json` is matched by `.gitignore:16` and absent in a fresh worktree (pre-existing, unrelated to this diff); layer0/layer1/layer2/quarantine ok |
| R5 | `node tools/validate/validate.js; echo $?` | pass | 4e4d73c | 0 |
| V1 | diff read against AC1–AC2 + spec J2 + F4 | pass | 4e4d73c | only layer2_compose.go + new test; merge detected by `!!merge` tag, direct key wins over merged, first merge source wins; resolve.go/pipeline.go/domain/fixtures untouched |

## Plan

skip: one-sentence diff following the existing hint pattern (failing test, then alias/merge-aware lookup).

## Outcome

`composeValue` (alias-following, merge-key-aware lookup) and `deref` in
`layer2_compose.go` now feed both the depends_on hints and the edge evidence
lines, so hint and edge cite the same entry line and J2 prunes aliased/merged
ordering-only edges. `resolve.go` unchanged. Metrics unchanged (no fixture
uses aliases): R82.4/P100.0 published, R81.2/P100.0 perturbed, SC15 95.0%.

- Linearized 2026-10-14 by the orchestrator: originally verified at 6b6c604 in an isolated worktree; cherry-picked onto 9ba80a6 as 4e4d73c with a byte-identical tree (git diff 6b6c604 HEAD empty), so every row above holds at the new SHA. The orchestrator also ran the full bench race in the main checkout (macro-map.json present): layer0/1/2 ok.
