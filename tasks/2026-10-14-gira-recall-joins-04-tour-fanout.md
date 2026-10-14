---
schema: task/1
tier: standard
status: verified
verified_at: "06aaee8"
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
depends_on: [2026-10-14-gira-recall-joins-03-unsupported-posture]
---

# J4 tour fan-out to Kafka consumers (SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-recall-joins; spec v1 approved)
- note: d2a5d51 landed a first J4 fan-out without a task record or an
  `SCTourFanout` test; this record covers it retroactively and test-first.
- decision (owner, 2026-10-14): "Corrigir o bench" — Blockers option 1: the
  layer2 anchor match treats `-` and `_` alike (case-insensitive, both sides).
  Scope widened for this attempt to `bench/layer2/**`; fixtures, ground truth,
  specs and SC15 thresholds/assertions stay untouched; floor stays 90.0%.

## Objective

`GetTour` enqueues the consumers of a visited `contract:kafka:<topic>`
alongside its providers, so `order-flow` covers `accounting` +
`fraud-detection`, while bench Layer 2 assertiveness stays ≥ 90.0% on both
SC15 tours.

In scope: `internal/slices` tour walk, why templates and citations; tests;
gira-query goldens only where J4 changes them; the layer2 anchor match
(owner decision, see log).
Out of scope: bench thresholds/assertions, fixtures, ground truth (`reviewed` stays false), extractors,
eval, transitive fan-out (spec: deferred, "one hop closes order-flow").

## Acceptance criteria

- AC1 = SC4
- AC2 = SC5 (invariants + regression)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | bench `go test ./layer2 -run SC15 -v` before this task | pass | d2a5d51 | runs green but SC4 floor unmet: assertiveness 75.0% (15/20) both tours, order-flow 100%; 1c64693 (= d2a5d51~1): 85.0% (17/20) both, order-flow 81.8% (missed accounting, fraud-detection); same 75.0% at 93f4226 |
| B2 | `TestSCTourFanout` red, before the change | fail | 93f4226 | at 1c64693: accounting/fraud-detection not stops; at 93f4226: fans out to gRPC consumers (`service:admin` via `contract:CartService`) and transitively (`service:audit` via `contract:kafka:ledger`), why sentence `checkout contract:kafka:orders consumed by it`, claim cites only the env line |
| B3 | SC15 history by worktree (scratch harness = bench FromGIRA + Measure rules) | pass | 8bbc42d | 8bbc42d and 863580f (J1): 90.0% (18/20) both; 45a248c (J2): 85.0% (17/20) both; 18d79b6, 1c64693: 85.0%; d2a5d51: 75.0% |
| B4 | `TestMeasureAnchorHyphenUnderscore` red on the old matcher | fail | 93f4226 | `image-provider`/`FRONTEND_PROXY`/`Image-Provider` verified 0, want 1 |
| AC1 | `go test ./internal/slices -run SCTourFanout -v` + bench `go test ./layer2 -run SC15 -v` | pass | 06aaee8 | SCTourFanout pass; assertiveness focused 95.0% (19/20, only `astronomy-db` left), full 100.0% (20/20); order-flow coverage 100.0% both. Before the bench fix the tour change alone gave 85.0% (17/20) both |
| AC2 | `go test -race ./...` root + bench, EndToEnd + SC3, validator | pass | 06aaee8 | root + bench race ok (incl. `TestMeasureAnchorHyphenUnderscore`, `TestMeasureHandChecked`, `TestUATourAdapter`); published R82.4/P100.0, perturbed R81.2/P100.0 (unchanged), safety/determinism 100%; SC3 pass; validator pass |
| R1 | `go test ./internal/slices ./cmd/gira -count=1` (goldens) | pass | 06aaee8 | no gira-query golden changes needed: `tour-frontend` is capped at 5 steps, before any fan-out stop; SC8 tour checks pass |
| V1 | diff read against AC1–AC2 + spec J4 / I1 / I2 | pass | 06aaee8 | fan-out only from Kafka topic contracts, one hop (fan-out stops don't fan out again); inferred consume edges render `(inferred)` and the map is not relabeled; every fan-out claim cites the consume edge (I1); bench matcher changed only in anchor normalization (case fold + `_`→`-` on both sides; separators not dropped, `image-provider` ≠ `IMAGE_PROXY`); SC15 thresholds/assertions, gt, fixtures untouched |

## Plan

1. Measure SC15 history in worktrees to find the 90→85→75 causes.
2. Write `TestSCTourFanout` (red at 1c64693 and 93f4226).
3. Restrict and fix the fan-out in `internal/slices/tour.go`; rerun AC1/AC2/R1/V1.

## Blockers

Resolved by owner decision (2026-10-14, "Corrigir o bench"): option 1
applied in `bench/layer2/proxies.go` (`normAnchor`). Kept below as the
record of why.

AC1's assertiveness floor (≥ 90.0%, 18/20) is not reachable from the tour
alone without changing what the bench measures or picking citations for the
bench's sake. Every 20-step SC15 tour ends exactly on a BFS level boundary,
so the set of stops is fixed; the remaining non-assertive steps are:

| Tour | Step | Claim locator | Why it fails the bench anchor check | Since |
|---|---|---|---|---|
| focused | `astronomy-db` via otel-collector | `compose.yaml#L873` `POSTGRES_HOST` | name absent | pre-E2 (in the 90% baseline) |
| focused + full | `frontend-proxy` via otel-collector | `compose.yaml#L866` `FRONTEND_PROXY_ADDR` | `frontend-proxy` vs `FRONTEND_PROXY` (`-` vs `_`) | pre-E2 (in the 90% baseline) |
| focused + full | `image-provider` via otel-collector | `compose.yaml#L868` `IMAGE_PROVIDER_HOST` | `-` vs `_` | 45a248c (J2 pruned the ordering-only `frontend → image-provider` depends_on edge that verified at L314) |
| full | `product-catalog` entry | `src/product-catalog/Dockerfile#L31` `EXPOSE ${PRODUCT_CATALOG_PORT}` | `-` vs `_` | pre-E2 (in the 90% baseline) |

Root cause chain: 90.0 → 85.0 at 45a248c (J2: `image-provider` moves to the
otel-collector route and an unverifiable env line); 85.0 → 75.0 at d2a5d51
(J4 adds `accounting`/`fraud-detection` citing only `KAFKA_TOPIC`, which never
names the consumer, displacing two verified stops, `chatbot` and `flagd-ui`,
past the 20-step cap). This task fixes the second step (75 → 85); the first
is not a J4 effect.

Options for the owner:

1. Bench: match anchors with `-` and `_` treated alike (bench change, out of
   this task's scope). Measured in scratch with this diff: focused 95.0%
   (19/20, only `astronomy-db` left), full 100.0% (20/20).
2. Record a `divergence:` entry setting the SC4 floor to the post-J2 baseline
   85.0% (owner call; spec SC5 already asks for one on any dip).
3. Tour-side citation changes for non-J4 steps (e.g. entry claims citing the
   compose block instead of the Dockerfile `EXPOSE`; re-attributing
   `astronomy-db` to product-catalog's extracted edge). Rejected here: they
   pick evidence to satisfy the proxy, not because J4 needs them.

## Checkpoint

- as of: 2026-10-14, branch overnight/oct15, head 06aaee8, worktree clean (code sealed by orchestrator; this record transitions to verified)
- done: B1–B4 measured; SCTourFanout and anchor test red then green; AC1/AC2/R1/V1 pass at 06aaee8 (orchestrator gate re-ran all)
- next: parent seal

## Outcome

Delivered at 06aaee8:

- `internal/slices/tour.go`: the consumer fan-out applies only to Kafka topic
  contracts (`protocol: kafka` / `contract:kafka:`), and stops reached by the
  fan-out don't fan out again (one hop). gRPC contract consumers are no longer
  enqueued (d2a5d51 did that for every contract).
- New why template `consumer` (additive): "`<consumer>` is reached because it
  consumes `<topic>` (<tier>), which `<stop>` <verb>." The tier is the consume
  edge's origin, so inferred edges read `(inferred)` (I2). It replaces the
  ungrammatical `` `checkout` contract:kafka:orders consumed by it``.
- The fan-out claim cites the consume edge; when that line sits in the
  consumer's own block of the same file (no other node declared in between),
  the range starts at the consumer's declaration (`compose.full.yaml#L12-L27`,
  `#L47-L66`), so the cited range shows whose `KAFKA_TOPIC` it is. Otherwise
  the edge line alone.
- `internal/slices/tour_test.go`: `TestSCTourFanout`,
  `TestTourFanoutCitationStaysInOwnBlock`.

- `bench/layer2/proxies.go` (owner decision): anchor match folds case and
  treats `-` and `_` alike on both sides (`normAnchor`); nothing else in
  `Measure` changed. `bench/layer2/layer2_test.go`:
  `TestMeasureAnchorHyphenUnderscore` (equivalence both directions, case,
  and non-matches: `image-provider` vs `IMAGE_PROXY`, dropped separators).

SC15 now: focused 95.0% (19/20), full 100.0% (20/20), order-flow 100.0%.
Remaining non-assertive step: focused `astronomy-db` via otel-collector
(`POSTGRES_HOST` never names it) — honest miss, left as is.
