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
| `mcp` | Required: enables the stdio MCP adapter (ADR 0004) |

## Flags

| Flag | Description |
|---|---|
| `--index` | Path to `macro-map.json` |

## Notes

- `gira` never runs a persistent web server daemon (ADR 0008). The only
  server surface is the stdio MCP adapter used by coding agents.
- Invoke with `serve mcp` only.

## See also

- [ADR 0004: CLI core with stdio MCP adapter](#)
