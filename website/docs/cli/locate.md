---
sidebar_position: 10
---

# `gira locate`

Search component boundaries, methods, and unhandled files for a query string.

## Synopsis

```text
gira locate <query> [--index <path>] [--format json|md] [--checkout <dir>]
```

## Arguments

| Argument | Description |
|---|---|
| `query` | Free-text query matched against boundaries, methods, and unhandled files |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--index` | string | — | Path to `macro-map.json` |
| `--format` | string | `json` | `json` \| `md` |
| `--checkout` | string | — | Local checkout for staleness comparison |

## Output

Boundary matches for the query, each with evidence locators. Unknown nodes
reported by other commands carry these suggestions in `node_not_found` errors.

## Example

```bash
gira locate currency --index .gira/index.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (including zero matches) |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Explain](./explain.md)
