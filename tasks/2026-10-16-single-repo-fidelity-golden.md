---
schema: task/1
tier: standard
status: verified
verified_at: "bee2b27"
---

# Single-repo fidelity: intra-module marking + self-index golden

## Routing and approvals

- tier: standard (S2 holds: mark via pipeline post-pass; golden via committed snapshot; S1, S3, S4, S6, S7 hold; S5 absent; S8 holds)
- decided: internal pkgs as Library (already true, 18 nodes) by owner (2026-10-16, "Library interna")
- decided: CLI contracts deferred to follow-up by owner (2026-10-16, "Depois")
- decided: golden as committed snapshot with UPDATE_SNAPSHOTS=1 by owner (2026-10-16, "Snapshot committed")

## Objective

Make single-repo maps first-class without schema change: mark
intra-module libraries, add a self-index golden snapshot test, and
record follow-ups (CLI contracts, minting discipline).

In scope: `layer3_go`-adjacent pipeline post-pass, golden test +
snapshot, follow-up pointers.
Out of scope: new NodeType, CLI contract extraction, repinning,
minting-discipline fix (H3: changes node sets consumed by eval).

## Acceptance criteria

1. AC1: intra-module `library:` nodes carry
   `metadata.origin_detail: intra-module`; externals unchanged.
2. AC2: golden test green against committed snapshot; regen via
   `UPDATE_SNAPSHOTS=1`.
3. AC3: follow-ups recorded (CLI contracts, minting discipline).

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test ./internal/extractors/` before edits | pass | 11645f9 | ok pre-edit |
| AC1 | unit test + package suite | pass | bee2b27 | TestMarkIntraModuleLibraries green; own marked, external/service untouched |
| AC2 | golden green + snapshot committed | pass | bee2b27 | TestSelfFidelityGolden green; `_note` records service:internal noise |
| AC3 | follow-ups in `## Outcome` | pass | bee2b27 | CLI contracts + minting discipline listed |
| R1 | `go test -race ./...` + lint + validator | pass | bee2b27 | suite 0 (17 ok); lint 0 new; validator pass; corpus golden accepted (+41 tokens estimate, reviewed) |
| V1 | diff read against AC1–AC3 | pass | bee2b27 | 5 files, additive only (post-pass + 2 tests + snapshot + corpus golden) |

## Outcome

Delivered: post-pass `markIntraModuleLibraries` (no interface change),
unit + golden tests, committed snapshot. Follow-ups: CLI commands as
`contract:` nodes; Service-minting discipline for non-main packages
(`service:internal`/`service:bench` noise, H3: node-set change).
Incident: first draft of this work was lost uncommitted when a
concurrent agent commit landed (11645f9, no file overlap); recreated
verbatim and committed as adb8113 — commit early when sharing a checkout.
