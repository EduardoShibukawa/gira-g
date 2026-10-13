---
name: gira-arch-diff
description: Compare two GIRA indexes (two refs, two releases, base vs head of a PR) and report the architectural changes - nodes, edges and Contracts added, removed, renamed or changed, newly unsupported paths, and moved evidence - as a human changelog or JSON for CI. Use at review time to catch unintended Contract changes.
operations: ["gira diff <index-a> <index-b> --format md|json", "MCP diff_maps"]
output_contract: Diff JSON (renamed, added_nodes, removed_nodes, added_edges, removed_edges, changed_contracts, newly_unsupported, no_longer_unsupported, evidence_moved, token_renames)
budget: none (proportional to the change)
---
# gira-arch-diff

Architectural evolution between two indexes.

## Operations

1. Index both sides (`gira index ... --output base.json`, `... --output head.json`).
2. MCP `diff_maps {index_a, index_b}` or `gira diff base.json head.json --format md`.

## Output contract

`Diff` JSON: `renamed`, `added_nodes`, `removed_nodes`, `added_edges`, `removed_edges` (removed items use the first index's IDs), `changed_contracts`, `newly_unsupported`, `no_longer_unsupported`, `evidence_moved` (informational), `token_renames`. Renames are detected from structure; an ambiguous one is reported as a removal plus additions, never guessed.

## Budget

None: the output is proportional to the change.

## Rules

- In review, flag every `changed_contracts`, `removed_edges` and `newly_unsupported` item; cite the IDs as reported.
- Don't present `evidence_moved` as an architectural change.
