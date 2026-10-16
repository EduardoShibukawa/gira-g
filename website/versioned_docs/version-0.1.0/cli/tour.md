---
sidebar_position: 7
---

# `gira tour`

Render a deterministic guided reading tour from the entry points, or from a
focus node.

## Synopsis

```text
gira tour [--focus <node>] [--max-steps <n>] [--index <path>] [--format json|md] [--checkout <dir>]
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--focus` | string | — | Start the tour at this node instead of the entry points |
| `--max-steps` | int | (default tour steps) | Maximum number of tour steps |
| `--index` | string | — | Path to `macro-map.json` |
| `--format` | string | `json` | `json` \| `md` |
| `--checkout` | string | — | Local checkout for staleness comparison |

## Output

An ordered list of tour steps, each with the node, its evidence locators, and
the reason it was selected. Tours are deterministic: repeated runs over the
same index produce byte-identical output.

## Example

```bash
gira tour --focus service:checkout --index .gira/index.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Flows](./flows.md)
- [Slice](./slice.md)
