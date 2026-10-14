---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
verified_at: "a30415d"
---

# J1 Health denylist (SC1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-recall-joins; spec v1 approved)

## Objective

Stdlib Health registrations (`RegisterHealthServer`,
`UnimplementedHealthServer`) never become `provides` edges or
`contract:Health` nodes. `task-01`/`task-07` lose the `contract:Health`
`unexpected` with no recall loss.

In scope: SC1, test-first. Failing test first: a Health registration
emitting `contract:Health` must fail.
Out of scope: J2–J4; ground-truth changes.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC5 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/extractors -count=1` + validator before change | pass | 8bbc42d | extractors ok, validator pass; E2 parent B1 green at 9050e14 |
| AC1 | `go test ./internal/extractors -run HealthDenylist -v` + EndToEnd `unexpected` check | pass | a30415d | failed first (contract:Health emitted), passes after denylist; task-01/task-07 lose contract:Health, R82.4 unchanged, P90.3→93.3 published / 89.7→92.9 perturbed |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | a30415d | root 16 pkgs ok, bench ok, validator pass; recall unchanged, safety/determinism 100% |
| R1 | `go test ./internal/extractors ./internal/eval ./cmd/gira -count=1` | pass | a30415d | goldens updated (Health-only diff: node/edge/locator + scores); task-08 missing unchanged (pre-existing) |
| V1 | diff read against AC1–AC2 + spec J1 | pass | a30415d | literal-only denylist (Health), register+embed+client paths; no business name affected |

## Plan

1. Failing `TestGoASTExtractor_HealthDenylist` first.
2. Denylist stdlib Health literals in `layer3_go.go` (register + embed + client-node paths).
3. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

J1 delivered: stdlib `Health` never becomes a contract node or provides
edge (register + embed + client paths, literal-only). task-01
`unexpected=[]`, task-07 keeps only `[service:agent]`; published
P90.3→93.3, perturbed P89.7→92.9, recall unchanged. 02 unblocked.
