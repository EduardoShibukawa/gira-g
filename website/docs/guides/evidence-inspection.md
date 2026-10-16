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

## Freshness

Nodes/edges cite the commit they were extracted from. Freshness is one of:

- `current` — locator matches the indexing commit
- `stale` — locator commit differs from the indexing commit (or the local
  HEAD, when a checkout is provided)
- `unverified` — the commit could not be resolved

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
