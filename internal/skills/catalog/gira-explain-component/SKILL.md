---
name: gira-explain-component
description: Build a contextual dossier for one subsystem (what it is, Contracts provided and consumed, who depends on it, infrastructure, unsupported boundaries, evidence) plus its deterministic flows from an entry through consumes edges to infrastructure. Use when a human asks how a component works or how a request flows through it.
operations: ["gira explain <node> --format json", "gira slice <node> --view architecture --tier contracts --format json", "gira flows <entry> --format md|json"]
output_contract: EvidenceReport JSON + ContextSlice JSON + Flows JSON (entry, flows[{steps[{node_id, via, verb, locator}]}])
budget: contracts tier 1500 for the slice
---
# gira-explain-component

A dossier and its flows, from evidence only.

## Operations

1. `gira explain <node> --format json` for the node's evidence (origins, deployment profiles, freshness).
2. `gira slice <node> --view architecture --tier contracts --budget 1500 --format json` for its Contracts and neighbours.
3. `gira flows <node> --format md` for chains like HTTP route → frontend → CartService → cart → valkey. Each hop cites the edge's locator.

## Output contract

`EvidenceReport`, `ContextSlice` and `Flows` JSON. Flows hops are `{node_id, via, verb, locator}`.

## Budget

The slice uses the contracts budget, 1500 estimated tokens.

## Rules

- GIRA does not extract business rules. To explain what code does, point at the cited ranges and label your explanation `narrated`, citing those locators.
- Unsupported boundaries are `unknown`, not guesses.
