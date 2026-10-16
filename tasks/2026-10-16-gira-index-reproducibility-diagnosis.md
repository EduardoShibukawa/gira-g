---
schema: task/1
tier: standard
status: verified
verified_at: "843ccfc"
---

# Gira index reproducibility diagnosis

## Routing and approvals

- tier: standard (S2 fails: fix approach open; S6 fails: no passing index-repro baseline; H1–H6 absent for diagnostic scope)
- decided: scope → diagnostic + fix proposal, no product-code edits, by owner (2026-10-16, "Diagnóstico + spec")
- decided: index source → dogfood + offline bundle, by owner (2026-10-16, "Dogfood + bundle offline")
- decided: unsupported posture → declare .md/TS/CSS/SVG by design, by owner (2026-10-16, "Declarar by design")
- decided: governance anomaly → investigate root cause in this task, by owner (2026-10-16, "Investigar")

## Objective

Diagnose why `bin/gira index` fails in this environment and record a
reversible fix proposal without changing product code.

In scope: reproduce offline/online failures, quantify index staleness
and unsupported posture, root-cause the governance-view anomaly.
Out of scope: manifest rewrites, new extractors, cache vendoring.

## Acceptance criteria

1. AC1: offline and online `index` failures reproduced with exact errors.
2. AC2: staleness (indexed commit vs HEAD) and unsupported breakdown recorded.
3. AC3: governance-view 0-rels anomaly root-caused to code location.
4. AC4: dogfood + offline-bundle proposal recorded, no `specs/` file (standard).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | index-repro baseline before diagnosis | not run | f2f8943 | S6 fails: no existing passing check for `bin/gira index`; established on worktree at 843ccfc |
| AC1 | `bin/gira index --offline` (exit 1, cache-missing) + online `--output /tmp/dogfood-fresh.json` (exit 1, unreachable) | pass | 843ccfc | offline: commit ec0f0c4 missing from cache; online: duh-homelab fetch exit 128, no device/auth |
| AC2 | `.gira/index.json` staleness + unsupported census via python3 | pass | 843ccfc | indexed 4aef216 vs HEAD 843ccfc (3 behind); 307 nodes, 733 rels, 127 unsupported (94 .md, 112 non-corpus, 1 OpAMP construct) |
| AC3 | governance filter root-cause (`internal/slices/engine.go`) | pass | 843ccfc | `isNodeAllowedInView` (engine.go:536-545): governance allows only Policy/Task/Skill; checkout neighbors are Service/Contract/Library/Infra, so 0 rels by design |
| AC4 | fix proposal in `## Outcome` | pass | 843ccfc | dogfood + offline bundle + by-design declaration, no `specs/` file (standard) |
| V1 | diff read (task record only, no product edits) | pass | 843ccfc | `git status` shows only the new untracked task file; zero product files touched |

## Plan

skip: diagnostic order fixed in AC1–AC4; no step-order risk.

## Outcome

`bin/gira index` (gira.yaml dogfood pins) is irreproducible here:
offline needs cache for `ec0f0c4` (absent), online cannot fetch
`duh-homelab/gira-dogfood-services` (exit 128). Prior NFR audit
(`4aef216`) claimed index exit 0 — discrepancy recorded, likely
cache/network drift. `.gira/index.json` is 3 commits stale
(4aef216 vs 843ccfc). Governance 0-rels for checkout is by design
(engine.go:540-541 filters to Policy/Task/Skill). Proposal: vendor or
document an offline cache bundle for the two dogfood repos; declare
website/docs `.md`/TS/CSS/SVG unsupported by design instead of new
extractors; note governance focal-node-with-empty-neighborhood in docs.
No product code touched. Follow-up fix (manifest/cache/docs) is a
separate task and may escalate to heavy (H3/H5) by owner call.
