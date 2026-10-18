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

Fix finding 1 from the Macro Map consumer audit: `cloneMacroMap` must preserve `UnresolvedRefs` and `Conflicts`, including their evidence, without aliasing the source slices.

## Scope

- Copy `UnresolvedRefs` and `Conflicts` in `internal/indexer/incremental.go`.
- Deep-copy the nested `Evidence` slices on both record types.
- Add regression tests for all diagnostic fields, evidence independence, and nil-slice behavior.

## Out of scope

- Schema 2, `source_identity`, publication/store changes, and unrelated copy paths.
- Changes to the versioned-publication spec or parent task status.

## Acceptance criteria

- AC1: cloned MacroMap preserves every field in each unresolved-reference and conflict record.
- AC2: nested evidence slices are independent; mutating the clone does not mutate the source.
- AC3: nil diagnostic slices remain nil when the source slices are nil.
- AC4: focused tests and the repository CI pass on Ubuntu, macOS, and Windows.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | regression test for unresolved refs and conflicts | not run | | Await CI |
| AC2 | nested evidence non-aliasing assertions | not run | | Await CI |
| AC3 | nil-slice regression test | not run | | Await CI |
| AC4 | CI tests on Ubuntu/macOS/Windows, validator, floor guard | not run | | Await CI |
| V1 | self-review of diff against AC1–AC4 | not run | | Await final head SHA |

## Plan

1. Preserve diagnostic slices in `cloneMacroMap`, copying each nested evidence slice.
2. Add focused tests that fail on the old implementation.
3. Run repository CI and review the integrated diff.
4. Update verification only against the final tested commit SHA.

## Blockers

- CI for the current pull request is still running; local Go tests were not available in the execution environment.
- This fix does not authorize or implement the versioned-publication proposal.
