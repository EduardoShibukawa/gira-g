---
sidebar_position: 8
---

# `gira diff`

Compare two Macro Map indexes and report what changed.

## Synopsis

```text
gira diff <index-a> <index-b> [--format json|md] [--checkout <dir>]
```

## Arguments

| Argument | Description |
|---|---|
| `index-a` | Path to the baseline `macro-map.json` |
| `index-b` | Path to the comparison `macro-map.json` |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--format` | string | `json` | `json` \| `md` |
| `--checkout` | string | — | Local checkout for staleness comparison |

## Output

An architecture diff: renames, added/removed nodes and edges, contract
changes, and unsupported paths, each with evidence locators where applicable.

## Example

```bash
gira diff .gira/base.json .gira/index.json --format md
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Impact](./impact.md)
