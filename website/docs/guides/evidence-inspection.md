---
sidebar_position: 3
---

# Evidence inspection

Every fact in a Macro Map carries an Evidence Record with a locator and a
derivation tier.

## Derivation tiers

- **declared** — from a manifest or overlay (e.g. `gira.yaml`)
- **extracted** — from deterministic code parsers (Extractors)
- **inferred** — from naming or env-wiring heuristics, surfaced honestly

## Staleness

Nodes/edges cite the commit they were extracted from. If that commit differs
from the indexing commit, the item is flagged `stale`. When a local checkout is
provided, the local HEAD is also compared.

## `gira explain`

```bash
gira explain --node <node-id> --index .gira/index.json
gira explain --edge "<source> -> <target>" --index .gira/index.json
```

Surface the `EvidenceRecord` entries for a node or edge: git locator, discovery
profile, derivation tier, freshness status. `explain` fetches a capped source
excerpt (default ±40 lines) on demand; `--no-source` disables it.

## Related

- [Concepts: Evidence Record](../concepts/evidence-record.md)
- [Export guide](./export.md)
