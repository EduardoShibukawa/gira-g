---
name: gira-tour
description: Give a newcomer, or a Change-maker heading to one component, a guided architectural tour and reading roadmap computed deterministically from the Macro Map. Use when someone asks "where do I start", "walk me through this system" or "what do I read before touching X".
operations: ["gira tour [--focus <node>] [--max-steps <n>] --format md|json", "MCP get_tour"]
output_contract: Tour JSON (focus, entry_points, steps[{order, title, node_ids, read[locators], why, why_template, why_args}], truncated)
budget: max 20 steps by default (--max-steps)
---
# gira-tour

A computed reading roadmap.

## Operations

1. MCP `get_tour {focus?, max_steps?}`, or `gira tour [--focus <node>] [--max-steps n] --format md`.
2. Without a focus, the tour starts at the entry points (Components with declared ports). With a focus, it starts there.

## Output contract

`Tour` JSON: `entry_points` and `steps`, each `{order, title, node_ids, read, why}`. `read` lists the locators to open, in order. `why` is rendered from a template (`why_template`, `why_args`) and never written freely.

## Budget

20 steps unless `--max-steps` says otherwise; `truncated` is true when the walk was cut.

## Rules

- Present steps in `order`, and give the reader the `read` locators verbatim.
- You may add a friendlier explanation per step, but label it `narrated` and cite the step's locators. Never add a step GIRA did not produce.
