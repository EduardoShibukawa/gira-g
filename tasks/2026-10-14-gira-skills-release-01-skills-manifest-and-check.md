---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-skills-release
spec: gira-skills-release@1
---

# Skills catalog manifest & drift check (F3; SC3)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-skills-release; spec v1 approved)

## Objective

Extend `internal/skills` and `cmd/gira` (`gira skills`) with:
1. `gira skills manifest [--format json|md] [--output <path>]`: deterministic
   catalog manifest containing the `gira` binary version, a catalog-level
   `sha256:<hex>` digest, and the sorted list of canonical skills (`name`,
   `description`, `operations`, `output_contract`, `budget`, `sha256` of the
   canonical `SKILL.md` bytes).
2. `gira skills check --host claude|cursor|agents [--dir <project>] [--index <path>] [--command <cmd>] [--hook]`:
   non-mutating verification that computes the expected host files in memory,
   compares them against disk, prints `{host, in_sync, missing, drifted, ok}`
   (JSON on stdout), and exits non-zero when any owned file or marker region is
   missing or drifted, without writing or touching any file on disk.

In scope:
- `internal/skills` manifest and check helpers;
- `cmd/gira` `skills manifest` and `skills check` subcommands;
- `TestSCSkillsManifestAndCheck` in `internal/skills` and `cmd/gira`.

Out of scope:
- changing `gira skills install` behavior or canonical `SKILL.md` contents;
- GoReleaser / CI workflow wiring (children 02 and 03).

## Acceptance criteria

- AC1 = SC3 (`go test ./internal/skills ./cmd/gira -run SCSkillsManifestAndCheck -count=1`: deterministic manifest with valid SHA-256 digests for all 6 skills; `skills check` passes right after `skills install` for `claude`, `cursor`, `agents` including `--hook`; `skills check` exits non-zero on missing/drifted owned regions without modifying disk)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before change | not run | | |
| AC1 | `go test ./internal/skills ./cmd/gira -run SCSkillsManifestAndCheck -count=1` | not run | | |
| R1 | `go test -race ./internal/skills ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1 (SC3); verify `skills check` never writes to disk | not run | | |

## Constraints and budget

- `gira skills install` idempotence and marker-scoping must remain 100% intact.
- `gira skills check` is strictly read-only (must never create directories or modify files).
- Zero fixture diff (`fixtures/`, `testdata/`); ADR 0001 and ADR 0012 invariants hold.

## Plan

1. Write failing tests `TestSCSkillsManifestAndCheck` in `internal/skills` and `cmd/gira`.
2. Implement `skills.BuildManifest(version string)` and `skills.Check(host string, opts Options)` in `internal/skills`.
3. Wire `gira skills manifest` and `gira skills check` in `cmd/gira/main.go`.
4. Run AC1, R1, V1 and validator; commit.
