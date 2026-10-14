---
schema: task/1
tier: heavy
status: verified
verified_at: "8343122"
spec: gira-recall-joins@1
---

# Recall joins (E2: Health, depends_on, unsupported posture, tour fan-out)

## Routing and approvals

- tier: heavy (H1: Health denylist scope, depends_on corroboration, unsupported posture, tour breadth are product calls; H3: spans Go/Compose extractors, resolve, slices/tours, eval + bench; H5: 4 causes, multi-session)
- confirmed: heavy by owner (2026-10-14, "Acima + implementar E2")
- approved: spec v1 by owner (2026-10-14, "Acima + implementar E2")

## Objective

Raise honesty first, recall/precision second: denylist stdlib Health
(J1), corroborate `depends_on`-only edges (J2), give every
unsupported-language miss a locator or an explicit `unknown` (J3), fan
tours out to topic consumers (J4). No extractor rewrites; joins only.

In scope:

- spec `gira-recall-joins` v1;
- the children listed in the Plan;
- integrated verification of SC1–SC5.

Out of scope:

- E3 envelopes/hygiene, E4 human surface, E5 skills distribution, E6 docs site;
- new language extractors, ground-truth changes, and `reviewed: true` (the owner's call);
- any model client (ADR 0001), any paid fan-out;
- push or release (each needs its own approval; the overnight branch push is authorized separately).

## Acceptance criteria

- AC1 = SC1–SC5 of `specs/gira-recall-joins.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator + EndToEnd before children | pass | 9050e14 | root 16 pkgs ok, bench ok, validator pass; published R82.4/P90.3, perturbed R81.2/P89.7, safety/determinism 100% |
| AC1 | SC1–SC5 of `specs/gira-recall-joins.md` v1 (rows SC1–SC5 below) | pass | 8343122 | all five SC rows pass at 8343122 |
| SC1 | `go test ./internal/eval -run SC1HealthDenylist` | pass | 8343122 | both variants: no contract:Health in map, task-01/07 unexpected clean; red on 8bbc42d (pre-E2); denylist now stdlib-import only (c48e96a) |
| SC2 | `go test ./internal/eval -run SC2DependsOnCorroboration` | pass | 8343122 | no service→service consumes rests on depends_on alone; corroborated edges cite both lines; task-02/11 clean; red on 8bbc42d. Note: J2's own AC1 filter matched no test, so this is the first end-to-end run |
| SC3 | `go test ./internal/eval -run SC3UnsupportedPosture` | pass | 8343122 | billing/exchange/storefront/opamp-server/ExchangeService (+ published names) each an explicit unknown row with locator; zero silent-missing; gt untouched, reviewed=false |
| SC4 | `go test ./internal/slices -run SCTourFanout` + bench Layer 2 SC15 | pass | 8343122 | order-flow 100% (accounting + fraud-detection); assertiveness focused 95.0% (19/20), full 100.0% (20/20) ≥ 90.0%; bench anchor `-`/`_` normalization per owner decision 2026-10-14 |
| SC5 | I1/I2 + race root+bench + validator + E0 guard | pass | 8343122 | published R82.4/P100.0 (B1 R82.4/P90.3), perturbed R81.2/P100.0 (B1 R81.2/P89.7); safety/determinism 100%; recall not lowered; inferred edges stay inferred |
| R1 | `go test -race ./...` (root, integrated head) | pass | 8343122 | all packages ok |
| R2 | `go test -race ./...` (bench/) | pass | 8343122 | all packages ok |
| R3 | `node tools/validate/validate.js` | pass | 8343122 | Overall: pass |
| V1 | integrated diff read + fresh-context review | pass | 8343122 | review of 9050e14..a2b7b29 returned fail on F1 (transitive fan-out via broker) → fixed 0a47433; F3 (Health not import-scoped) → fixed c48e96a; F2/F4–F7 non-blocking, recorded as follow-ups in ## Review |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-health-denylist`: J1 stdlib-Health denylist (SC1).
2. `-02-depends-on-corroboration`: J2 corroboration rule (SC2). Depends on 01.
3. `-03-unsupported-posture`: J3 join-where-wired else explicit unknown (SC3). Depends on 02.
4. `-04-tour-fanout`: J4 consumer fan-out (SC4). Depends on 03.

SC5 is checked at every child and integrated on the parent.

## Review

Fresh-context review (V1) of `9050e14..a2b7b29`, 2026-10-14. Verdict:
fail on F1. Every finding and what was done with it:

- F1 (high, blocking): the tour fanned out transitively through the broker
  (Infrastructure provider of every topic) → fixed in 0a47433 (only flow
  Services not reached via a topic fan out; `TestTourFanoutBrokerDoesNotFanOut`).
- F3 (medium): the Health denylist was by name only → fixed in c48e96a (requires
  the `google.golang.org/grpc/health` import; `TestGoASTExtractor_HealthDenylistIsStdlibOnly`).
- F2 (medium, follow-up): `chatbot → opamp-server` (env `OPAMP_SERVER_ENDPOINT`
  + `.env OPAMP_SERVER_HOST`) is pruned because env wiring needs a declared
  target node; metric-neutral (task-07 already missed opamp-server). J3 posture
  holds (explicit unknown); a join for undeclared env targets is later work.
- F4 (low, follow-up): `depends_on` through YAML alias/merge keys yields no
  hint, so such ordering-only edges survive (precision only).
- F5 (low, follow-up): corroboration is same-edge/same-verb; contract-reference
  corroboration (spec J2) is not implemented; no connectivity loss today.
- F6 (low, follow-up): `TestSC3UnsupportedPosture` matches rows by directory,
  not by the row's `unknown_joins` naming the id; not vacuous (8 fails on 9050e14).
- F7 (low, follow-up): a parsed Go file with an unsupported construct appears in
  `mm.Unsupported` (export/LocateBoundary count it as a boundary).

No assertion weakened; zero diff in `fixtures/` and `testdata/`; `reviewed`
stays false.

## Checkpoint

- as of: 2026-10-14, branch overnight/oct15, head 8343122, worktree clean (untracked results.tsv, run.log, spike-e2-joins.md stay uncommitted)
- done: children 01–04 verified; SC1–SC5, R1–R3, V1 pass at 8343122
- next: PR overnight/oct15 → main and merge (owner-authorized in chat, 2026-10-14)

## Outcome

E2 delivered: J1 stdlib Health denylist, J2 depends_on corroboration, J3
explicit-unknown posture for unsupported-language misses, J4 one-hop topic
fan-out. Hermetic precision 90.3→100.0 (published) / 89.7→100.0 (perturbed),
recall unchanged by design (honesty first); tour assertiveness 95.0/100.0,
order-flow 100%. Follow-ups F2, F4–F7 in ## Review.
