---
sidebar_position: 1
---

# `gira init`

Generate a `gira.yaml` manifest from a local Git checkout (pinned to `HEAD`) and build the Macro Map index in one step.

## Synopsis

```text
gira init [dir] [--force] [--output <path>] [--offline] [--cache-dir <dir>]
```

## Arguments

| Argument | Description |
|---|---|
| `dir` | Target Git repository directory (defaults to `.`) |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--force` | bool | false | Repin `gira.yaml` to the local `HEAD` SHA when a single-repository manifest already exists |
| `--output` | string | — | Path to write `macro-map.json` (defaults to `~/.cache/gira/indexes/<system>/macro-map.json`) |
| `--offline` | bool | false | Disable network egress; use local cache exclusively |
| `--cache-dir` | string | — | Local repository cache directory |

## Output

1. When `gira.yaml` is absent in `dir`, writes a single-repository manifest (`version: 1`) with `target_system.name` derived from the repository root directory and `repositories[0]` pinned to `file://<toplevel>` at the local `HEAD` SHA.
2. When `gira.yaml` already exists and `--force` is set, updates `repositories[0].sha` to the local `HEAD` SHA (refuses multi-repository manifests so hand-authored manifests are never clobbered).
3. Runs the deterministic indexing pipeline and writes `macro-map.json` to `--output` (or `~/.cache/gira/indexes/<system>/macro-map.json` by default).

If the Git working tree has uncommitted changes, `gira init` prints a warning to `stderr` noting that the index reflects the committed `HEAD` SHA.

## Examples

```bash
# Zero-config initialization and indexing in the current Git repository
gira init

# Write the index to a project-local path
gira init . --output .gira/macro-map.json

# Repin gira.yaml to the latest local HEAD commit and re-index
gira init . --force --output .gira/macro-map.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Directory is not a Git repository, `HEAD` has no commits, multi-repo `--force` refusal, or indexing error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [`gira index`](./index.md) — build an index from an existing single- or multi-repository `gira.yaml`
- [Quickstart](../guides/quickstart.md) — end-to-end onboarding guide
