---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: "HEAD"
---

# Explain source excerpts (F2; SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-human-surface; spec v1 approved)
- depends_on: ["2026-10-14-gira-human-surface-01-plain-renderer-and-tty"]

## Objective

`gira explain <node|--edge>` reads the cited `[repo@commit:path#Lstart-Lend]`
from the local checkout on demand, shows a capped window (default ±40 lines),
and reports `source unavailable` cleanly when the checkout or range is missing.
`--no-source` disables the excerpt. The index keeps locators only (ADR 0001);
excerpts are read at render time and never persisted.

## Acceptance criteria

- AC1 = SC2 (`SCExplainSource`: window honored, `--no-source`, `source unavailable`, no bodies in the index)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before change | pass | HEAD~1 | child 01 merged; all green |
| AC1 | `go test ./internal/slices ./cmd/gira -run SCExplainSource -count=1` | pass | HEAD | window ±40 honored, `--no-source`, `source unavailable`, path escape rejected |
| R1 | `go test ./internal/slices ./cmd/gira -count=1` | pass | HEAD | no regression; JSON stays locator-only |
| V1 | diff read against AC1 | pass | HEAD | `slices.Excerpt` reads at render time; never persists (ADR 0001) |

## Constraints and budget

- Excerpts read only; never written to the index or disk (ADR 0001).
- Missing checkout/range → `source unavailable`, never an error exit.
- `--format json` stays locator-only (no bodies).
- Deterministic plain output; zero fixture diff.

## Plan

1. Failing test for the excerpt window and the `source unavailable` path.
2. Implement the on-demand reader (git checkout) + cap + `--no-source`.
3. Wire the styled/plain renderer from child 01.
4. Run AC1/R1/V1; commit.

## Outcome

`gira explain` gains `--repo` (default `.`) and `--no-source`, and appends a
`## Source Excerpts` section: `slices.Excerpt` reads the cited lines ±40 from the
local checkout, one block per distinct locator, with `source unavailable` when
the file/range is missing or the path escapes the checkout. The index keeps
locators only (ADR 0001); JSON output is unchanged.
