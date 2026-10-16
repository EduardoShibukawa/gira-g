---
schema: task/1
tier: standard
status: todo
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-01-scaffold]
---

# 03 — Content: CLI Reference (F2; SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Create one reference page per CLI command in `website/docs/cli/`:
`gira map`, `gira impact`, `gira slice`, `gira index`, `gira serve`.
Each page documents flags, inputs, output shape, examples, and exit codes.

In scope: CLI reference pages.
Out of scope: guides (04), concepts (02).

## Acceptance criteria

- AC1 = SC2 (CLI): every command (`gira map`, `gira impact`, `gira slice`,
  `gira index`, `gira serve`) has a reference page.
- AC2: each page documents all flags (from `--help` output), inputs, output
  shape, and at least one example.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | HEAD | |
| AC1 | ls website/docs/cli/ — all 5 commands present | not run | | |
| AC2 | content review: flags, inputs, output, examples | not run | | |
| R1 | go test -race ./... (no Go change) | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |

## Plan

1. Run `gira <cmd> --help` for each command; capture flags.
2. Create `website/docs/cli/<cmd>.md` for each command.
3. Run AC1–AC2/R1/V1; commit.

## Outcome

pending
