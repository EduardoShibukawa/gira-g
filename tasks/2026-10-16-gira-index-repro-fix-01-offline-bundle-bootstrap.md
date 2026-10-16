---
schema: task/1
tier: standard
status: verified
verified_at: "59b8a5f"
parent: 2026-10-16-gira-index-repro-fix
spec: 2026-10-16-gira-index-repro-fix@1
---

# Index repro fix 01: offline bundle bootstrap

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-index-repro-fix)

## Objective

Ship `tools/index-bundle/bootstrap.sh` plus an "Offline bundle" section
in `website/docs/cli/index.md` (SC1).

In scope: script, index.md section.
Out of scope: vendored repo contents, repinning.

## Acceptance criteria

1. AC1: `bash -n tools/index-bundle/bootstrap.sh` exits 0 and `--help`
   explains cache-dir/output/manifest flow.
2. AC2: index.md documents offline bundle flow and dogfood pins.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | parent B1 (`go test -race`, pre-children) | not run | a82e44b | cited from parent; rerun in 04 |
| AC1 | `bash -n` + `--help` | pass | 59b8a5f | `bash -n` ok; `--help` prints usage + options |
| AC2 | index.md section present + `mise run docs:build` | pass | 59b8a5f | Offline bundle section live; build SUCCESS (index.html + .nojekyll) |
| V1 | diff read against AC1–AC2 | pass | 59b8a5f | 2 files, 71 insertions, no other touches |

## Outcome

Bootstrap script executable (`tools/index-bundle/bootstrap.sh`) with
`--manifest/--cache-dir/--output/--skip-index/--help`; index.md gains
the Offline bundle flow. No follow-ups.
