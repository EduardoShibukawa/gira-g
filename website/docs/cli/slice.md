---
sidebar_position: 3
---

# `gira slice`

Extract a bounded context slice around a Component, Contract, or node.

## Synopsis

```text
gira slice <node_id> [--tier <tier>] [--view <view>] [--budget <tokens>] [--index <path>]
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--tier` | string | `summary` | `summary` \| `contracts` \| `deep` |
| `--view` | string | `full` | `full` \| `architecture` \| `governance` |
| `--budget` | int | (tier default) | Hard ceiling on estimated output tokens (500/1500/3500 per tier) |
| `--checkout` | string | — | Local checkout for staleness comparison |
| `--format` | string | `json` | `json` \| `md` |
| `--index` | string | — | Path to `macro-map.json` |

## Notes

- `--tier deep` renders the widest context; `summary` is the Change-maker
  default for a quick scan.
- `--view governance` keeps only security/owner-relevant items.

## See also

- [Impact](./impact.md)
- [Quickstart guide](../guides/quickstart.md)
