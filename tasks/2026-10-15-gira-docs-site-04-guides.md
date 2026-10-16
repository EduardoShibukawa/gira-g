---
schema: task/1
tier: standard
status: todo
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-02-concepts, 2026-10-15-gira-docs-site-03-cli-reference]
---

# 04 — Content: Guides (F2; SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Create guide pages in `website/docs/guides/`: Quickstart (install, first
index, first impact), Change-maker workflow (brief → edit → verify),
Evidence inspection (`gira explain`, derivation tiers, source excerpts,
staleness), and Export (static HTML/Markdown export surface).

In scope: guide pages.
Out of scope: concepts (02), CLI reference (03).

## Acceptance criteria

- AC1 = SC2 (guides): `website/docs/guides/` has quickstart, workflow,
  evidence, and export pages.
- AC2: quickstart includes install steps and a first `gira map` + `gira impact`
  example; evidence guide covers `gira explain`, tiers, and staleness.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | HEAD | |
| AC1 | ls website/docs/guides/ — 4 guides present | not run | | |
| AC2 | content review: quickstart, workflow, evidence, export | not run | | |
| R1 | go test -race ./... (no Go change) | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |

## Plan

1. Write `website/docs/guides/quickstart.md`.
2. Write `website/docs/guides/change-maker-workflow.md`.
3. Write `website/docs/guides/evidence-inspection.md`.
4. Write `website/docs/guides/export.md`.
5. Run AC1–AC2/R1/V1; commit.

## Outcome

pending
