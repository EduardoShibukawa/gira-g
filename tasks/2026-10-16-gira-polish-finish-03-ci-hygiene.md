---
schema: task/1
tier: standard
status: verified
verified_at: "8742187"
parent: 2026-10-16-gira-polish-finish
spec: 2026-10-16-gira-polish-finish@1
depends_on: [2026-10-16-gira-polish-finish-02-identity-version]
---

# Polish 03: CI hygiene + licenses

## Routing and approvals

- tier: standard (decomposed from heavy parent)

## Objective

Close CI/quarantine/license gaps: canonical case-insensitive
quarantine grep over full `website/`, frozen docs install, six
missing `THIRD_PARTY_NOTICES.md` entries.

In scope: `.github/workflows/docs.yml`, `mise.toml` `docs:build`,
`THIRD_PARTY_NOTICES.md`.
Out of scope: docs prose, versioning, Go code.

## Acceptance criteria

1. AC1 = SC5a: grep covers canonical alternatives
   (bench, gira-bench, graphify, understand variants, egonex-ai),
   case-insensitive, full `website/`; benign `benchmarks` not flagged.
2. AC2 = SC5b: `docs:build` uses `--frozen-lockfile`; notices list
   `clsx`, `react-dom`, `typescript`, three `@docusaurus/*` packages.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 8419f77 | pre-edit, exit 0 |
| B2 | mise run docs:build (baseline) | pass | 8419f77 | pre-edit |
| AC1 | grep pattern review + empty run | pass | 8742187 | canonical alternatives, exit 1 empty; sanity matches all tokens |
| AC2 | mise task read + notices entries | pass | 8742187 | frozen-lockfile, six entries added |
| R1 | validator + quarantine grep empty | pass | 8742187 | both pass |
| V1 | diff read against AC1–AC2 | pass | 8742187 | 3 files, no Go/docs touched |

## Plan

skip: two-file mechanical change, order fixed by AC1→AC2

## Outcome

Delivered at `8742187`: canonical quarantine grep, frozen docs
install, six missing license entries. No Go/docs changes.
