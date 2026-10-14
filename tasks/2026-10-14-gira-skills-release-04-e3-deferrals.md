---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-skills-release
spec: gira-skills-release@1
depends_on:
  - 2026-10-14-gira-skills-release-01-skills-manifest-and-check
---

# E3 deferrals: gate-lint explicit CI strictness & findings register GC (F5; SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-skills-release; spec v1 approved)

## Objective

Close the two deferred open questions from E3 (`specs/gira-envelopes-hygiene.md`
Open questions) per owner decisions Q5=A and Q6=A:
1. **`gate-lint` CI strictness policy (Q5=A):** Keep the default unconfigured CI
   job warn-only (`TestSCGateLintStrictness` untouched), and add `FailOnCI bool`
   to `gate.LintOptions` plus `--fail-on-ci` on `gira gate-lint` so callers with
   an explicit ledger in CI (`CI=true` / `GIRA_CI=1`) can enforce fail-closed
   `--strict` without `CI` downgrading failures to warnings.
2. **Findings register retention/GC policy (Q6=A):** Add `Register.GC(keepResolved int) Register`
   in `internal/eval/findings.go` and `--gc` + `--keep-resolved <n>` (default `0`)
   in `cmd/gira/review.go` (`gira review --findings-register <path> --gc [--keep-resolved <n>]`):
   every finding with status `open` is preserved unconditionally in order;
   resolved findings (`fixed`, `closed`, `verified`) are compacted to the last
   `keepResolved` entries, and the compacted register is validated and saved.

In scope:
- `internal/gate/gate.go`, `cmd/gira/gate.go`;
- `internal/eval/findings.go`, `cmd/gira/review.go`;
- `TestSCE5Deferrals` in `internal/gate`, `internal/eval`, and `cmd/gira`.

Out of scope:
- changing default warn-only behavior in `.github/workflows/ci.yml`.

## Acceptance criteria

- AC1 = SC5 (`go test ./internal/gate ./internal/eval ./cmd/gira -run SCE5Deferrals -count=1`: `gate-lint --strict --fail-on-ci` fails closed under `CI=true` while default `CI=true` remains warn-only; `Register.GC` and `gira review --gc [--keep-resolved <n>]` keep all `open` findings and compact resolved ones)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before change | not run | | |
| AC1 | `go test ./internal/gate ./internal/eval ./cmd/gira -run SCE5Deferrals -count=1` | not run | | |
| R1 | `go test -race ./internal/gate ./internal/eval ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1 (SC5); verify `TestSCGateLintStrictness` and `TestSCL3TaxonomyRegister` stay green | not run | | |

## Constraints and budget

- Never drop an `open` finding during GC; negative `keepResolved` is treated as `0` (or rejected cleanly).
- Preserve all existing E1/E3 gate and findings tests.

## Plan

1. Write failing `TestSCE5Deferrals` tests in `internal/gate`, `internal/eval`, and `cmd/gira`.
2. Add `FailOnCI` to `gate.LintOptions` and `--fail-on-ci` to `cmd/gira/gate.go`.
3. Add `Register.GC(keepResolved int)` to `internal/eval/findings.go` and `--gc` / `--keep-resolved` to `cmd/gira/review.go`.
4. Run AC1, R1, V1 and validator; commit.
