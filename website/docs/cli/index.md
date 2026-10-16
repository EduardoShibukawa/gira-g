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
| `--output` | string | — | Path to write `macro-map.json` |
| `--offline` | bool | false | Disable network egress; use local cache exclusively |
| `--cache-dir` | string | — | Local repository cache directory |
| `--manifest` | string | — | Path to the manifest (alternative to positional arg) |

## Output

A JSON `macro-map.json` (or the file given by `--output`) containing the
extracted Macro Map: Components, Contracts, relationships, and Evidence
Records. Foreign files under `.git/`, caches, and dependencies are skipped.

## Examples

```bash
# Index the current directory using gira.yaml
gira index

# Write the index to a specific file, offline
gira index --output .gira/index.json --offline
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
  --output .gira/index.json
```

Without a warm cache, offline fails with `commit <sha> missing from
... cache`; online without access fails fetching the manifest remotes
(e.g. the `gira.yaml` dogfood pins). Pass `--skip-index` to only warm
the cache without re-indexing.

## Notes

- Run before any other command that needs `--index`.
- Re-run after changing extractors or overlays to keep the index current.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Impact](./impact.md) — analyze change impact from an index
- [Slice](./slice.md) — extract a bounded context slice
