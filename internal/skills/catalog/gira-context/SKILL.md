---
name: gira-context
description: Fetch the architectural context of one node (service, contract, infrastructure) progressively, starting at the summary tier and escalating to contracts or deep only when the budget's truncation signal or the task requires it. Use when you need what a component exposes, consumes or depends on.
operations: ["gira slice <node> --tier summary --budget 500 --format json", "gira slice <node> --tier contracts --budget 1500 --format json", "gira slice <node> --tier deep --budget 3500 --format json", "MCP get_context_slice"]
output_contract: ContextSlice JSON (focal_node, tier, neighbors, relationships, unsupported_boundaries, budget)
budget: summary 500, contracts 1500, deep 3500
---
# gira-context

Progressive disclosure over a context slice.

## When to use

When a task needs a component's neighbours, Contracts or evidence, and an impact briefing is more than you need.

## Operations

1. Start at `summary`: MCP `get_context_slice {node_id, tier: "summary", budget: 500}` or `gira slice <node> --tier summary --budget 500 --format json`.
2. Escalate to `contracts` (budget 1500) when you need RPC methods and messages, and to `deep` (budget 3500) when you need every evidence record or the unsupported boundaries.
3. If `budget.truncated` is true, either rerun with the budget in `budget.escalate_hint` or narrow the slice with `--view architecture`.

## Output contract

A `ContextSlice` JSON object: `focal_node`, `tier`, `neighbors`, `relationships`, `unsupported_boundaries` (deep tier), and `budget {limit, estimated, truncated, dropped {nodes, relationships, evidence}, escalate_hint}`. Every response states its tier, its budget, the estimated tokens used and what was dropped.

## Budget

ADR 0006 tiers: summary 500, contracts 1500, deep 3500 estimated tokens. The focal node, the unsupported boundaries and the budget marker are never dropped.

## Rules

- Evidence or silence: keep the locator of every fact you repeat; mark anything else `unknown`.
- Narrate, don't infer: label non-literal sentences `narrated` and cite their locators.
