---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
verified_at: "829300b"
---

# F6 SC3 unknown rows cite the observed id (review follow-up)

## Routing and approvals

- tier: standard (review follow-up F6, low, from the `## Review` of the parent
  2026-10-14-gira-recall-joins; test-only, one file; persisted record required)

## Objective

`TestSC3UnsupportedPosture` matched an unknown row by directory only: any
unsupported row under the watched id's directory passed, even one citing an
unrelated Contract or nothing. Tighten it so each expected unknown row cites
the observed id: its `unknown_joins` contains the watched id, or (for a
Service, whose row lives in its own code) a Contract the task expects; a
construct that names no Contract (OpAMP) cites through a specific reason.

In scope: `internal/eval/posture_test.go` only. Tighter, never looser: every
existing assertion stays (reviewed=false, joined-with-locator, real corpus
file under the directory, non-empty reason, non-vacuous watch list).
Out of scope: production code, fixtures, testdata, ground truth.

## Acceptance criteria

- AC1 = SC3 tightened: each unknown row cites the observed id; non-vacuous
  (fails against a weakened production state), passes at HEAD.
- AC2 = no regression: EndToEnd R/P not lower, SC15 assertiveness ≥ 90%,
  race suites, validator exit 0.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/eval -run SC3UnsupportedPosture -v` before change | pass | 9ba80a6 | 14 unknown rows matched by dir only (e.g. published task-04 frontend matched `Checkout.gateway.ts` citing CheckoutService) |
| RF1 | red-first: tightened test vs `unknown_joins.go` with the `append(ub.UnknownJoins, c.id)` stripped (reverted) | pass | 829300b | tightened test FAILS: 12 silent misses (every Contract-citing row, both variants); old dir-only test under same mutation: 5 fail, 7 pass vacuously (e.g. task-02 CurrencyService, task-08 frontend/storefront) |
| RF2 | red-first: tightened test vs OpAMP reason in `layer3_go.go` replaced by `unsupported construct` (reverted) | pass | 829300b | tightened test FAILS: 2 silent misses (opamp-server task-07, both variants); dir-only test only requires a non-empty reason |
| AC1 | `go test ./internal/eval -run SC3UnsupportedPosture -v` | pass | 829300b | 14 rows cite: CurrencyService/ExchangeService by id; payment→PaymentService, billing→BillingService, frontend task-04→`Currency.gateway.ts` CurrencyService, task-08→CheckoutService, storefront→ExchangeService/OrderingService; opamp-server via reason `OpAMP client` |
| AC2 | `go test ./internal/eval -run EndToEnd -v` | pass | 829300b | published R82.4/P100.0 (TR78.4), perturbed R81.2/P100.0 (TR80.2); safety/determinism 100% |
| R1 | `(cd bench && go test ./layer2 -run SC15 -v)` | pass | 829300b | assertiveness 95.0% (19/20) focus frontend, 100.0% (20/20) full tour |
| R2 | `go test -race ./...` root | pass | 829300b | all packages ok |
| R3 | `go test -race ./...` bench/ | pass | 829300b | in the worktree layer0 cannot open gitignored `bench/layer0/testdata/gira/macro-map.json` (absent from fresh worktrees; unrelated to this test-only diff); run in a scratch copy of the tree with the main checkout's ignored file supplied: layer0/1/2/quarantine ok |
| R4 | `node tools/validate/validate.js; echo $?` | pass | 829300b | exit 0 before the code commit and before this docs commit |
| V1 | diff read against AC1–AC2 + ticket F6 | pass | 829300b | only `internal/eval/posture_test.go` changed; every prior assertion kept (reviewed=false, joined-with-locator, regular corpus file under dir, non-empty reason, vacuity guard); no fixture/testdata/gt change |

## Plan

skip: one-file test tightening; order is fixed by the red-first contract.

## Outcome

`TestSC3UnsupportedPosture` now requires each unknown row to cite the
observed id (`citer` predicate passed to `unknownRow`; `postureWatch.reason`
for the OpAMP construct). Failure output lists each row's path with its
`unknown_joins`. All prior assertions kept; production code, fixtures and
ground truth untouched; `reviewed` stays false. Red-first evidence shows the
tightened law catches 12 + 2 weakened states the dir-only law let through
partially (7 vacuous passes) or entirely (OpAMP reason).

- Linearized 2026-10-14 by the orchestrator: originally verified at 3b9bbfa (on 9ba80a6) in an isolated worktree; cherry-picked on top of F4 as 829300b and re-verified there by the orchestrator in the main checkout: validator exit 0, race root ok, bench layer0/1/2 race ok, EndToEnd + SC1–SC3 pass (published R82.4/P100.0, perturbed R81.2/P100.0), SC15 95.0%/100.0%. bench/quarantine re-run at the final gate (it scans live .claude/worktrees).
