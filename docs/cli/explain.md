---
sidebar_position: 6
---

# `gira explain`

Explain the discrete Evidence Records behind a node or an edge.

## Synopsis

```text
gira explain (--node <id> | --edge <id>) [--index <path>] [--format json|md] [--repo <dir>] [--checkout <dir>] [--source | --no-source]
```

A bare positional id is treated as `--node`.

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--node` | string | — | Node id to explain |
| `--edge` | string | — | Edge id to explain |
| `--index` | string | — | Path to `macro-map.json` |
| `--format` | string | `json` | `json` \| `md` |
| `--repo` | string | `.` | Local checkout the locators are read from (source excerpts) |
| `--checkout` | string | — | Local checkout whose HEAD is compared against the indexed commits (staleness) |
| `--source` | bool | true | Fetch source excerpts from `--repo` (md output) |
| `--no-source` | bool | false | Do not fetch source excerpts (locators only) |

## Output

An Evidence Report: the matching Evidence Records with locators
(`[repo@commit:path#Lstart-Lend]`) and derivation tiers. Markdown output
appends a Source Excerpts section unless `--no-source` is given; an unreadable
locator renders the "source unavailable" marker.

## Example

```bash
gira explain --node service:checkout --index .gira/index.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Locate](./locate.md)
- [Slice](./slice.md)
