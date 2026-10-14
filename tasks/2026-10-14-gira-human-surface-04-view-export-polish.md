---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: "HEAD"
---

# View/export polish (F4; SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-human-surface; spec v1 approved)
- depends_on: ["2026-10-14-gira-human-surface-01-plain-renderer-and-tty"]

## Objective

`gira view` gains `--no-open` (and skips the browser in CI/headless);
`gira export --format html` is guaranteed a single self-contained file with no
external asset reference and no server; `md` and `html` carry the same content,
including the boundaries section.

## Acceptance criteria

- AC1 = SC4 (`SCViewExport`: `--no-open` writes but does not open; HTML has no external asset reference)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before change | pass | HEAD~1 | child 03 merged; all green |
| AC1 | `go test ./internal/slices ./cmd/gira -run SCViewExport -count=1` | pass | HEAD | `--no-open` opens nothing; HTML is single-file; a fake xdg-open proves the open path |
| R1 | `go test ./internal/slices ./cmd/gira -count=1` | pass | HEAD | no regression |
| V1 | diff read against AC1 | pass | HEAD | `openInBrowser(path, noOpen)`; export unchanged |

## Constraints and budget

- `--no-open` must not invoke the browser; CI/headless guard.
- HTML export has no external URLs/assets (single file).
- No behavior change for `export --format md` content.
- Deterministic output; zero fixture diff.

## Plan

1. Failing tests for `--no-open` and the single-file guarantee.
2. Implement the flag + env guard; assert no external references in the HTML.
3. Run AC1/R1/V1; commit.

## Outcome

`gira view` gains `--no-open` (and still honors `CI`/`NO_BROWSER`/headless):
`openInBrowser(path, noOpen)` never launches a browser with the flag. The
single-file guarantee is asserted for both `export` and `view` HTML (no
`<script src>`, `<link>`, external `src`/`href`, `@import` or `url(http)`).
