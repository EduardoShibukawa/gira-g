---
schema: task/1
tier: standard
status: active
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
depends_on: [2026-10-14-gira-recall-joins-02-depends-on-corroboration]
verified_at: "7ae2447"
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

## Acceptance criteria

- AC1 = SC3
- AC2 = SC5 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/extractors ./internal/eval -count=1` + validator before change | pass | 7ae2447 | 02 verified; probe probe2 ok, validator pass |
| AC1 | `go test ./internal/eval -run SC3UnsupportedPosture -v` | not run | | needs test authored (see above) |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | not run | | |
| R1 | `go test ./internal/extractors ./internal/eval -count=1` | not run | | |
| V1 | diff read against AC1–AC2 + spec J3 | not run | | |

## Plan

1. Author `TestSC3UnsupportedPosture` codifying the probe.
2. Implement join-where-wired, explicit unknown otherwise.
3. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

skip: active, implementation pending
