---
schema: task/1
tier: standard
status: todo
---

# Gira-G: preservar diagnósticos ao clonar MacroMap

## Routing and approvals

- tier: standard (S1 holds: one bounded data-loss bug; S2 holds: direct copy-path fix; S3 holds: one implementation plus regression tests; S4 holds: reversible and contained)
- approved: standard fix by owner (2026-10-18, "faça o fix do 1")

## Objective

Fix the bounded data-loss bug in `cloneMacroMap` and close adjacent, independently testable Git path/process/environment safety findings from the same audit.

## Scope

- Copy `UnresolvedRefs` and `Conflicts` in `internal/indexer/incremental.go`, including independent nested `Evidence` slices.
- Parse `git ls-tree` output using NUL-delimited records and preserve unusual pathnames.
- Always reap the `tar` child when `git archive` fails.
- Sanitize Git repository-redirection environment variables for local CLI repository probes.
- Add focused regression tests for diagnostic copies, unusual pathnames, and Git environment isolation.

## Out of scope

- Schema 2, `source_identity`, versioned publication/store protocol, strict schema error taxonomy, and `gira init` manifest publication ordering. These remain gated by the draft spec and separate acceptance plan.
- Changes to the versioned-publication spec or parent task status.

## Acceptance criteria

- AC1: cloned MacroMap preserves every field in each unresolved-reference and conflict record.
- AC2: nested evidence slices are independent; mutating the clone does not mutate the source.
- AC3: nil diagnostic slices remain nil when the source slices are nil.
- AC4: `ListTree` preserves filenames containing tabs, newlines, and backslashes.
- AC5: `ExportTree` reaps the tar child even when `git archive` fails.
- AC6: local Git probes ignore repository-redirection environment variables.
- AC7: focused tests and repository CI pass on Ubuntu, macOS, and Windows.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | regression test for unresolved refs and conflicts | not run | | Await CI |
| AC2 | nested evidence non-aliasing assertions | not run | | Await CI |
| AC3 | nil-slice regression test | not run | | Await CI |
| AC4 | unusual-path regression for `ListTree` | not run | | Await CI |
| AC5 | `ExportTree` child cleanup reviewed; failure injection not yet covered | partial | | Await CI and follow-up failure-path test |
| AC6 | Git environment isolation regression | not run | | Await CI |
| AC7 | CI tests on Ubuntu/macOS/Windows, validator, floor guard | not run | | Await latest head CI |
| V1 | self-review of diff against AC1–AC4 | not run | | Await final head SHA |

## Plan

1. Preserve diagnostic slices in `cloneMacroMap`, copying each nested evidence slice.
2. Switch Git tree parsing to NUL-delimited records and always reap the extraction process.
3. Sanitize inherited repository-redirection variables for local Git probes.
4. Run repository CI and review the integrated diff.
5. Update verification only against the final tested commit SHA.

## Blockers

- CI for the updated pull request is pending; local Go tests were not available in the execution environment.
- `ExportTree` still needs a deterministic archive-failure regression test.
- This fix does not implement the versioned-publication proposal or change its draft/todo governance.
