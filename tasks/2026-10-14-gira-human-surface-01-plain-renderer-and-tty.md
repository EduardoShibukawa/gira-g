---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: ""
---

# Plain renderer, TTY and boundaries (F1, F5; SC1, SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-human-surface; spec v1 approved)
- adds `charmbracelet/lipgloss` (owner Q1 approved); record the version

## Objective

Add a styled/plain rendering seam to the human commands: `lipgloss` on a TTY,
byte-deterministic plain text when stdout is not a TTY or `NO_COLOR` is set; add
the Unsupported Boundaries & Known Unknowns block. The agent `--format json`
contract is untouched.

## Acceptance criteria

- AC1 = SC1 (`SCHumanPlainFallback`: non-TTY and `NO_COLOR` byte-identical, no ANSI)
- AC2 = SC5 (`SCBoundariesSection`)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before change | not run | | |
| AC1 | `go test ./cmd/gira -run SCHumanPlainFallback -count=1` | not run | | |
| AC2 | `go test ./cmd/gira -run SCBoundariesSection -count=1` | not run | | |
| R1 | `go test ./internal/slices ./cmd/gira -count=1`; `--format json` goldens unchanged | not run | | |
| V1 | diff read against AC1–AC2; show diff before commit | not run | | |

## Constraints and budget

- Plain path byte-identical (goldens); no ANSI when `NO_COLOR` or non-TTY.
- `--format json` and E3 envelopes untouched.
- go.mod gains lipgloss; deterministic output; zero fixture diff.

## Plan

1. Failing goldens/tests for the plain fallback and the boundaries block.
2. Add lipgloss; implement the styled/plain seam + TTY/`NO_COLOR` detection.
3. Boundaries section in the human renderers.
4. Run AC1/AC2/R1/V1; commit.
