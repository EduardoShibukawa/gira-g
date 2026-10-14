---
schema: task/1
tier: heavy
status: active
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
| SC1 | `go test ./internal/eval -run SC1HealthDenylist` | not run | | |
| SC2 | `go test ./internal/eval -run SC2DependsOnCorroboration` | not run | | |
| SC3 | `go test ./internal/eval -run SC3UnsupportedPosture` | not run | | |
| SC4 | `go test ./internal/slices -run SCTourFanout` + bench Layer 2 SC15 | not run | | |
| SC5 | I1/I2 + race root+bench + validator + E0 guard | not run | | |
| R1 | `go test -race ./...` (root, integrated head) | not run | | |
| R2 | `go test -race ./...` (bench/) | not run | | |
| R3 | `node tools/validate/validate.js` | not run | | |
| V1 | integrated diff read + fresh-context review | not run | | |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-health-denylist`: J1 stdlib-Health denylist (SC1).
2. `-02-depends-on-corroboration`: J2 corroboration rule (SC2). Depends on 01.
3. `-03-unsupported-posture`: J3 join-where-wired else explicit unknown (SC3). Depends on 02.
4. `-04-tour-fanout`: J4 consumer fan-out (SC4). Depends on 03.

SC5 is checked at every child and integrated on the parent.

## Review

skip: runs after child 04 is verified

## Checkpoint

- as of: 2026-10-14, branch overnight/oct15, head 9050e14, worktree clean (parent newly created, spec approval pending commit)
- done: E1 accepted; E2 parent created with confirmed + approved entries; B1 green
- next: approve spec v1 (hash below), decompose 01–04, implement J1–J4 test-first
- open: push of overnight/oct15 authorized — do after spec approval commit
- revalidate: B1 ran on 9050e14
