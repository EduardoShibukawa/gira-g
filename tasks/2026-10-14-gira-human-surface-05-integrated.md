---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: "HEAD"
---

# Integrated verification (F1–F5; SC6, SC7)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-human-surface; spec v1 approved)
- depends_on: ["2026-10-14-gira-human-surface-01-plain-renderer-and-tty", "2026-10-14-gira-human-surface-02-explain-source", "2026-10-14-gira-human-surface-03-staleness", "2026-10-14-gira-human-surface-04-view-export-polish"]

## Objective

Integrate the four features and run the end-to-end regression: race root + bench,
the validator, the plain goldens, and the E3 invariants (R/P, SC15, frozen
ground-truth/corpus, `reviewed: false`).

## Acceptance criteria

- AC1 = SC6 (regression: race root+bench, validator exit 0, plain goldens stable)
- AC2 = SC7 (I3 invariants: published/perturbed R/P unchanged, SC15 ≥ 90%, zero fixture diff, `reviewed: false`)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before the integrated run | pass | HEAD~1 | child 04 merged; all green |
| AC1 | `go test -race ./...` root and bench/; `node tools/validate/validate.js` | pass | HEAD | root + bench green; validator pass |
| AC2 | `go test ./internal/eval -run EndToEnd -v`; bench SC15; `git diff --exit-code 21a8092 -- fixtures/ testdata/`; `"reviewed": false` | pass | HEAD | published 82.4/100.0, perturbed 81.2/100.0; SC15 ok; diff 0; reviewed false |
| R1 | `go test ./internal/slices ./internal/domain ./internal/human ./cmd/gira -count=1` | pass | HEAD | E4 packages green |
| V1 | fresh-context adversarial review (review.md); resolve findings in the parent ## Review | pass | HEAD | 9 findings F1–F9; fixed |

## Constraints and budget

- No change to `--format json`, envelope objects, or E3 behavior.
- Determinism 100%; zero fixture diff; no ground-truth change.
- README/docs changes minimal (docs site is E6).

## Plan

1. Integrate F1–F5; fix cross-feature issues.
2. Run AC1/AC2/R1 with fresh-context review.
3. Report integrated evidence for the parent seal.

## Outcome

Integrated: SC6 (race root+bench, validator) and SC7 (R/P unchanged, SC15,
frozen ground-truth/corpus, reviewed false) pass. The fresh-context adversarial
review produced F1–F9; all were fixed on this branch (see the parent `## Review`).
