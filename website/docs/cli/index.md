---
sidebar_position: 1
---

# `gira index`

Build a Macro Map index from a target system manifest.

## Synopsis

```text
gira index [manifest] [--output <path>] [--offline] [--cache-dir <dir>]
```

## Arguments

| Argument | Description |
|---|---|
| `manifest` | Path to the manifest (defaults to `gira.yaml` in the current directory) |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--output` | string | — | Path to write a detached snapshot without modifying the canonical store or `.head` |
| `--offline` | bool | false | Disable network egress; use local cache exclusively |
| `--cache-dir` | string | — | Local repository cache directory |
| `--manifest` | string | — | Path to the manifest (alternative to positional arg) |

## Publication and Output

- **Canonical store (default):** When `--output` is omitted, `gira index` publishes an immutable version to the canonical versioned store resolved via hierarchy D2 (`$XDG_CACHE_HOME/gira/indexes` -> `~/.cache/gira/indexes` -> `<project_root>/.gira/indexes`). The version directory `<store_root>/<lineage_id>/versions/<version_id>/` contains `macro-map.json` and publication `meta.json`, and `.head` is updated atomically.
- **Detached snapshot (`--output P`):** Writes a standalone Macro Map directly to path `P`. It does **not** touch the store, does not update `.head`, and does not alter `gira.yaml`. A notice is emitted to stderr indicating `P` is a detached snapshot and not the active index.
- **Schema 2 and Provenance:** Both canonical publications and detached snapshots use integer `schema_version: 2` and embed a self-contained `source_identity` object (`manifest_sha256`, extractor versions, and `sources` with repository hashes).

## Examples

```bash
# Index using gira.yaml and publish canonically to the versioned store
gira index

# Generate a detached snapshot for offline transfer or PR review diffing
gira index --output .gira/snapshot.json --offline
```

## Offline bundle

`--offline` needs the pinned commits already cached (default
`$XDG_CACHE_HOME/gira/repos` or `~/.cache/gira/repos`, override with
`--cache-dir`). Where network allows, bootstrap once, then copy the
cache directory plus the emitted `macro-map.json` to blocked machines:

```bash
# Needs network once: populates cache, then re-indexes offline
tools/index-bundle/bootstrap.sh --output /tmp/dogfood-bundle/macro-map.json

# On the blocked machine (bundle copied over)
gira index --offline --cache-dir /tmp/dogfood-bundle/cache \
  --output .gira/snapshot.json
```

Without a warm cache, offline fails with `commit <sha> missing from
... cache`; online without access fails fetching the manifest remotes
(e.g. the `gira.yaml` dogfood pins). Pass `--skip-index` to only warm
the cache without re-indexing.

## Notes

- Run before any other command that needs `--index` or relies on store discovery from `gira.yaml`.
- Re-indexing with identical `version_id` requires deterministic byte equivalence; non-deterministic output fails closed with `nondeterministic_build`.
- Inter-process concurrency is guarded by native M -> S -> P file locks with a 60-second timeout (`index_locked`).

## Exit codes and Errors

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, missing/corrupt index, unsupported schema, lock failure, or indexing error |

Errors print structured descriptions to stderr, including stable codes such as `index_missing`, `index_corrupt`, `index_schema_unsupported`, `index_tool_mismatch`, `nondeterministic_build`, `index_locked`, or `lock_unavailable`.

## See also

- [`gira init`](./init.md) — zero-config repository onboarding
- [Impact](./impact.md) — analyze change impact from an index
- [Slice](./slice.md) — extract a bounded context slice
