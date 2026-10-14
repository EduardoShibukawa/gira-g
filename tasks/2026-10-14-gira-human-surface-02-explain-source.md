---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: ""
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
| B1 | `go test -race ./...` (root) + bench + validator before change | not run | | |
| AC1 | `go test ./internal/slices ./cmd/gira -run SCExplainSource -count=1` | not run | | |
| R1 | `go test ./internal/slices ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1; show diff before commit | not run | | |

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
