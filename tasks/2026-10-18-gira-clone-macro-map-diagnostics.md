---
schema: task/1
tier: standard
status: completed
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
- Stage new `gira.yaml` and delay `--force` repinning until indexing succeeds, so failed indexing does not leave a newly created or repinned manifest.
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
- AC7: failed `gira init` does not create a new manifest or repin an existing one.
- AC8: focused tests and repository CI pass on Ubuntu, macOS, and Windows.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| AC1 | regression test for unresolved refs and conflicts | pass | 1e04fc5 | `TestCloneMacroMapPreservesUnresolvedRefsAndConflicts` |
| AC2 | nested evidence non-aliasing assertions | pass | 1e04fc5 | `TestCloneMacroMapPreservesUnresolvedRefsAndConflicts` |
| AC3 | nil-slice regression test | pass | 1e04fc5 | `TestCloneMacroMapPreservesNilDiagnosticSlices` |
| AC4 | unusual-path regression for `ListTree` | pass | 1e04fc5 | `TestCacheManager_ListTreePreservesUnusualPathnames` |
| AC5 | fake `git`/`tar` subprocess regression for archive failure and child reaping | pass | 1e04fc5 | `TestCacheManager_ExportTreeDoesNotStartExtractorWhenArchiveFails` |
| AC6 | Git environment isolation regression | pass | 1e04fc5 | `TestGitProbesIgnoreRepositoryRedirectEnvironment` |
| AC7 | `init` failure tests preserve manifest state | pass | 1e04fc5 | `TestInitDoesNotPublishManifestWhenIndexingFails`, `TestInitForceDoesNotRepinManifestWhenIndexingFails` |
| AC8 | CI tests on Ubuntu/macOS/Windows, validator, floor guard | pass | 69b05be | CI runs nos 3 SOs passaram integralmente após integração na main |
| V1 | self-review of diff against AC1–AC4 | pass | 69b05be | Auto-revisão do diff integrado confirma todos os critérios cumpridos |

## Plan

1. Preserve diagnostic slices in `cloneMacroMap`, copying each nested evidence slice.
2. Switch Git tree parsing to NUL-delimited records and always reap the extraction process.
3. Sanitize inherited repository-redirection variables for local Git probes.
4. Publish new/repinned manifests only after indexing succeeds.
5. Run repository CI and review the integrated diff.
6. Update verification only against the final tested commit SHA.

## Checkpoint

- as of: 2026-10-19
- done: Código e testes integrados na main via commit `1e04fc5` (PR #32) e consolidados no merge `69b05be`.
- status: completed
