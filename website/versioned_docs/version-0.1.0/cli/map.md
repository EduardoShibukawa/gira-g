---
sidebar_position: 0
---

# `gira map`

> Note: `gira` does not ship a standalone `gira map` command. The Macro Map
> is produced by [`gira index`](./index.md). This page is a redirect-style
> placeholder retained for vocabulary consistency with the documentation
> spec — see [`gira index`](./index.md).

## What "the map" is

`gira index` builds the Macro Map (`macro-map.json`) for a Target System. The
Map is the central artifact every other command reads (`--index`).

## Equivalent

```bash
gira index --output .gira/index.json
```

## See also

- [`gira index`](./index.md)
- [Concepts: Macro Map](../concepts/macro-map.md)

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (`gira index`) |
| `1` | Usage, flag, index-load, engine, or emit error — including `gira map` itself, which is an unknown command |

`gira` has no other exit codes. Errors print `Error: ...` to stderr.
