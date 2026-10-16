---
sidebar_position: 4
---

# `gira serve`

Start the stdio MCP protocol adapter for AI agent clients.

## Synopsis

```text
gira serve mcp [--index <path>]
```

## Arguments

| Argument | Description |
|---|---|
| `mcp` | Required: enables the stdio MCP adapter |

## Flags

| Flag | Description |
|---|---|
| `--index` | Path to `macro-map.json` |

## Output

The adapter speaks MCP over stdio. On stdout it emits protocol frames; logs go
to stderr. It exposes `impact`, `slice`, `locate`, and `explain`-style tools
derived from the index.

## Example

```bash
gira serve mcp --index .gira/index.json
```

## Notes

- `gira` never runs a persistent web server daemon (ADR 0008). The only
  server surface is the stdio MCP adapter used by coding agents.
- Invoke with `serve mcp` only.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (adapter running until stdin closes) |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Slice](./slice.md)
