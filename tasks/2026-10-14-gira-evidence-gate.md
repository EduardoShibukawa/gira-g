---
schema: task/1
tier: heavy
status: active
spec: gira-evidence-gate@1
---

# Evidence & gate hardening (Gallon L1/L2/L3 + token honesty)

## Routing and approvals

- tier: heavy (H1: product calls on sampling rate, taxonomy labels, gate strictness; H3: spans internal/eval, CLI, MCP, bench Layer 2; H5: 5 children, multi-session)
- confirmed: heavy by owner (2026-10-14, "Aprovar heavy E1")
- decided: L2 sampling default → strict max(20,5%) Gallon default by owner (2026-10-14, "max(20,5%) Gallon (Recommended)")
- decided: L3 taxonomy → Gallon S1 verbatim + mapping table to Gira tiers by owner (2026-10-14, "Gallon verbatim + mapa (Recommended)")
- decided: gate-lint --strict → warn-only in CI, fail-closed locally, until E3 by owner (2026-10-14, "Warn-only (Recommended)")
- approved: spec v1 by owner (2026-10-14, "Approve v1 (Recommended)")

## Objective

Make every `gira` claim independently re-derivable or explicitly silent: L1 artifact closing-line, L2 control-total + seeded sampling, L3 taxonomy + findings register + independent fix-check, honest token accounting, gate-lint. No extractor recall change; only the proof gets stronger.

In scope:

- spec `gira-evidence-gate` v1;
- the children listed in the Plan;
- integrated verification of SC1–SC6.

Out of scope:

- E2 joins, E3 envelopes/hygiene, E4 human surface, E5 skills distribution, E6 docs site;
- ground-truth changes, and `reviewed: true` (the owner's call);
- any model client (ADR 0001), any paid fan-out;
- push or release.

## Acceptance criteria

- AC1 = SC1–SC6 of `specs/gira-evidence-gate.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | SC1–SC6 rows | not run | | |
| V1 | read the integrated diff against spec v1, plus the fresh-context review | not run | | |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-l1-artifact-verified`: F1 + frozen scope check + dispatcher self-check (SC1).
2. `-02-l2-control-sample`: control-total + seeded max(20,5%) sampling, never-force-match (SC2). Depends on 01.
3. `-03-l3-taxonomy-register`: taxonomy + register + independent fix-check (SC3). Depends on 02.
4. `-04-token-honesty`: measured/estimated/net, no len/4 (SC4). Depends on 02.
5. `-05-gate-lint-validator`: gate-lint strict + integrated SC1–SC6 + E0 guard (SC5, SC6). Depends on 03, 04.

## Review

skip: runs after the children are integrated (fresh-context adversarial review, `.agents/skills/work/review.md`).
