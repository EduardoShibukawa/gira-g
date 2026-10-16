---
schema: task/1
tier: standard
status: verified
verified_at: "7d46ad9"
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-01-scaffold]
---

# 06 — Build pipeline (F4; SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Add the `docs:build` mise task (wrapping `bun install && bun run build` in
`website/`) and a CI workflow that builds the docs site on push to `main` and
uploads the static artifact.

In scope: mise task, CI workflow YAML.
Out of scope: hosting/deploy (artifact only in v1), content (02–05).

## Acceptance criteria

- AC1 = SC5: `mise run docs:build` exits 0 from a clean checkout.
- AC2 = SC5: CI workflow file exists in `.github/workflows/` and is valid YAML;
  it builds the site and uploads the artifact on push to main.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 3b01049 | |
| AC1 | mise run docs:build (clean checkout) | pass | 7d46ad9 | build success |
| AC2 | YAML parse of docs.yml | pass | 7d46ad9 | valid YAML |
| R1 | go test -race ./... (no Go change) | pass | 7d46ad9 | no Go touched |
| V1 | diff read against AC1–AC2 | pass | 7d46ad9 | only mise.toml + workflow |

## Plan

1. Add `docs:build` task to `mise.toml`.
2. Create `.github/workflows/docs.yml`.
3. Validate workflow YAML.
4. Run AC1–AC2/R1/V1; commit.

## Outcome

Added `docs:build` mise task (`cd website && bun install && bun run build`) and
`.github/workflows/docs.yml` with quarantine grep step and artifact upload.
mise task runs green; workflow YAML parses. AC1–AC2, R1, V1 pass at 7d46ad9.
