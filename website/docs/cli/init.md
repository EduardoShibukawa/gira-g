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
| `--output` | string | — | Path to write an identical byte-for-byte copy in addition to publishing canonically to the store |
| `--offline` | bool | false | Disable network egress; use local cache exclusively |
| `--cache-dir` | string | — | Local repository cache directory |

## Publication and Output Semantics

1. **Deterministic Indexing:** Constructs the Macro Map (Schema 2) with embedded `source_identity` binding raw manifest bytes, extractor versions, and repository commit SHA.
2. **Canonical Versioned Publication:** Publishes the version to `<store_root>/<lineage_id>/versions/<version_id>/` and updates `<store_root>/<lineage_id>/.head` atomically using the D2 store hierarchy (`$XDG_CACHE_HOME/gira/indexes` -> `~/.cache/gira/indexes` -> `<project_root>/.gira/indexes`).
3. **Snapshot Copy (`--output P`):** When `--output P` is passed, writes an identical byte-for-byte copy to `P`. If canonical publication succeeds but the copy fails, a partial publication error is returned.
4. **Manifest Creation (Written Last):** `gira.yaml` is written or updated **last**, only after index publication succeeds. Any failure during build, gate validation, or publication aborts without leaving a broken or uncommitted `gira.yaml` on disk.

If the Git working tree has uncommitted changes, `gira init` prints a warning to `stderr` noting that the index reflects the committed `HEAD` SHA.

## Concurrency and Durability

- **Lock Hierarchy:** Writers acquire native locks in strict sequence: **M** (Manifest) -> **S** (Store) -> **P** (Snapshot) with a 60-second deadline before returning `index_locked`.
- **Durability Reporting:**
  - **Linux / macOS:** Directory metadata is synchronized (`fsync`), reporting `confirmed`. If synchronization fails after version files are written, `init` returns a non-zero exit code, reports `visible_unconfirmed`, and aborts writing `gira.yaml`.
  - **Windows:** Win32 does not provide directory metadata sync; publication succeeds with exit code 0 and reports `visible_unconfirmed` with an informative warning.

## Examples

```bash
# Zero-config initialization and canonical publication to the versioned store
gira init

# Initialize and save a local project snapshot copy in addition to the store
gira init . --output .gira/macro-map.json

# Repin gira.yaml to latest HEAD commit and publish a new version
gira init . --force --output .gira/macro-map.json
```

## Exit codes and Errors

| Code | Meaning |
|---|---|
| `0` | Success (or `visible_unconfirmed` on Windows) |
| `1` | Not a Git repository, `HEAD` without commits, multi-repo `--force` refusal, sync failure, or indexing/lock error |

Errors print structured descriptions to stderr (`manifest_outside_git_root`, `git_root_unavailable`, `index_locked`, `lock_unavailable`, `nondeterministic_build`, etc.).

## See also

- [`gira index`](./index.md) — build an index from an existing single- or multi-repository `gira.yaml`
- [Quickstart](../guides/quickstart.md) — end-to-end onboarding guide
