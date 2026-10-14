---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: "HEAD"
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
| B1 | `go test -race ./...` (root) + bench + validator before change | pass | HEAD~1 | E4 setup merged; all green |
| AC1 | `go test ./cmd/gira -run SCHumanPlainFallback -count=1` | pass | HEAD | non-TTY and NO_COLOR byte-identical, no ANSI, matches `testdata/human/impact-checkout.md` |
| AC2 | `go test ./cmd/gira -run SCBoundariesSection -count=1` | pass | HEAD | `slice service:checkout --tier deep --format md` carries the Unsupported Boundaries block |
| R1 | `go test ./cmd/gira ./internal/human -count=1`; `--format json` unchanged | pass | HEAD | cmd/gira green; `emitHuman` only decorates Markdown on a TTY |
| V1 | diff read against AC1–AC2 | pass | HEAD | additive seam in `internal/human`; md sites routed through `emitHuman`/`emitMarkdownLine` |

## Constraints and budget

- Plain path byte-identical (goldens); no ANSI when `NO_COLOR` or non-TTY.
- `--format json` and E3 envelopes untouched.
- go.mod gains lipgloss; deterministic output; zero fixture diff.

## Plan

1. Failing goldens/tests for the plain fallback and the boundaries block.
2. Add lipgloss; implement the styled/plain seam + TTY/`NO_COLOR` detection.
3. Boundaries section in the human renderers.
4. Run AC1/AC2/R1/V1; commit.

## Outcome

Added `charmbracelet/lipgloss` v1.1.0 and the `internal/human` seam:
`Styled(w)` (false on `NO_COLOR` or non-TTY) and `Decorate` (bold headings,
yellow unsupported warnings; identity when not styled). Routed every Markdown
human site through `emitHuman`/`emitMarkdownLine`: impact, slice, explain,
locate, tour, flows, diff. The plain path is byte-identical (golden
`cmd/gira/testdata/human/impact-checkout.md`); JSON is untouched. The
Unsupported Boundaries block is asserted via `slice service:checkout --tier deep`.
