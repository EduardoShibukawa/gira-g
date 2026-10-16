---
sidebar_position: 9
---

# `gira flows`

List the deterministic chains from an entry node through consumes edges to
infrastructure.

## Synopsis

```text
gira flows <entry> [--index <path>] [--format json|md] [--checkout <dir>]
```

## Arguments

| Argument | Description |
|---|---|
| `entry` | Entry node id the flows start from |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--index` | string | — | Path to `macro-map.json` |
| `--format` | string | `json` | `json` \| `md` |
| `--checkout` | string | — | Local checkout for staleness comparison |

## Output

The flows reachable from the entry, each an ordered chain of nodes with
evidence locators. Output is deterministic for a given index.

## Example

```bash
gira flows service:checkout --index .gira/index.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Tour](./tour.md)
- [Slice](./slice.md)
