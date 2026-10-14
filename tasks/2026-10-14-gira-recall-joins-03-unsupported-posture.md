---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
depends_on: [2026-10-14-gira-recall-joins-02-depends-on-corroboration]
verified_at: "cb49c04"
---

# J3 unsupported-language posture (honesty; SC3 open-ended)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-recall-joins; spec v1 approved)

## Objective

No new language extractors in E2. Standardize "every unsupported call-site is
either joined (compose/env wiring proves it) or explicitly unknown with a
locator to the unsupported file — never a silent `missing`." Concretely, the
probe confirms `service:billing`, `service:exchange`, `service:storefront`,
`service:opamp-server`, `contract:ExchangeService` are all absent today with
unsupported needles present for billing/payment/currency/frontend/product-catalog.
Posture is verifiable already, but SC3 as written names a SC test that does
not yet exist. Test-first next: write `TestSC3UnsupportedPosture` codifying the
probe, then either add the explicit join where wiring proves it, or record
`unknown` alongside its unsupported file (never retarget gt).

In scope: SC3, test-first. Failing test first: a hidden miss must fail.
Out of scope: J4, new extractors, ground-truth changes.

Posture law (owner, 2026-10-14): each of billing/exchange/storefront/
opamp-server/ExchangeService must surface either its joined ID with a locator
or an explicit unknown row; no silent missing. Ground truth is not weakened
and `reviewed` stays false.

## Acceptance criteria

- AC1 = SC3
- AC2 = SC5 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/extractors ./internal/eval -count=1` + validator before change | pass | 7ae2447 | 02 verified; probe probe2 ok, validator pass |
| B2 | `go test ./internal/eval -run EndToEnd -v` + bench `go test ./layer2 -run SC15 -v` before change | pass | d2a5d51 | published R82.4/P100.0/TR78.9, perturbed R81.2/P100.0/TR80.6, safety/determinism 100%; SC15 assertiveness 75.0% (15/20) both tours |
| AC1 | `go test ./internal/eval -run SC3UnsupportedPosture -v` | pass | cb49c04 | failed first on d2a5d51 + test (6 silent misses: payment/billing task-03, currency/exchange task-04, storefront task-04 perturbed, opamp-server task-07 both); passes with every watched id an unknown row at the unsupported call-site; gt unread-only, reviewed=false asserted |
| AC2 | `go test -race ./...` (root + bench/) + EndToEnd + validator, no E0 dip | pass | cb49c04 | root + bench race ok; R82.4/P100.0 and R81.2/P100.0 unchanged; safety/determinism 100%; TR 78.4/80.2 (still ≥60, rows added); validator pass |
| R1 | `go test ./internal/extractors ./internal/eval ./internal/slices ./cmd/gira -count=1` | pass | cb49c04 | 4 gira-query goldens regenerated: J3-mandated `unknown_joins` + call-site lines + OpAMP row; impact-checkout budget est. 1060→1364 (<1500) |
| V1 | diff read against AC1–AC2 + spec J3 | pass | cb49c04 | no node/edge from unsupported code (unit `TestUnsupportedPostureIsUnknownNotJoined`); OpAMP row names no peer (I1); literal construct table only; SC15 unchanged vs B2 |

## Plan

1. Author `TestSC3UnsupportedPosture` codifying the probe.
2. Implement join-where-wired, explicit unknown otherwise.
3. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

J3 delivered as explicit-unknown posture (no wiring in the hermetic mini
proves any watched edge, so nothing is joined; recall unchanged by design):

- `UnsupportedBoundary.UnknownJoins` (additive, omitempty).
- `annotateUnknownJoins` (post-resolve): a no-extractor file naming a known
  gRPC Contract (or `<Name>Client`) lists it as an unknown join and cites the
  first mention line (`src/payment/index.js#L46`, `src/currency/src/server.cpp#L141`,
  `src/frontend/gateways/rpc/*.gateway.ts#L5`).
- `unsupportedGoConstructs` literal table (opamp-go import / `*OpAMPClient`
  call): unknown row at `src/product-catalog/main.go#L171`, peer never named.
- Engine relates a boundary to the nodes in its `UnknownJoins`.

Observation for J4 (out of scope here): bench SC15 assertiveness is 75.0%
(15/20) at d2a5d51 and at cb49c04; it was 85.0% before d2a5d51 and the spec
SC4 floor is ≥90.0%. J4 must address it before the parent seal.
