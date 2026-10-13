---
name: gira-impact-briefing
description: Before modifying code in a service, contract or file you have not mapped yet, or before opening a PR, get the deterministic blast radius from GIRA (upstream dependencies, downstream consumers, affected Contracts, mediation edges, unsupported boundaries). Use for a node ID, a file path you are about to edit, or a git revision range.
operations: ["gira impact <target> --format json --budget 1500", "gira impact --from-diff <base>..<head> --format json --budget 1500", "MCP get_impact_briefing", "MCP get_pr_impact"]
output_contract: ImpactBriefing JSON (target_id, upstream_services, downstream_services, affected_contracts, mediation_edges, unsupported_boundaries, optional diff, budget)
budget: 1500
---
# gira-impact-briefing

Pre-coding context: the blast radius of a change, from the evidence-backed Macro Map.

## When to use

- Before the first edit to code mapped to a Component or Contract.
- When routing a change: a changed Contract with consumers outside the touched Component is evidence that the change crosses modules.
- Before a PR: run it on the revision range.

## Operations

1. Prefer MCP: `get_impact_briefing {target_id, depth, budget: 1500}`; for a change set, `get_pr_impact {range: "<base>..<head>", budget: 1500}`.
2. Without GIRA MCP tools, run the CLI: `gira impact <target_id> --format json --budget 1500`, or `gira impact --from-diff <base>..<head> --format json --budget 1500`.
3. If you only have a file path, run `gira locate <path-or-name> --format json` first, or use the diff mode on the working tree change.

## Output contract

An `ImpactBriefing` JSON object: `target_id`, `upstream_services`, `downstream_services`, `affected_contracts`, `mediation_edges`, `unsupported_boundaries`, and when budgeted a `budget` object `{limit, estimated, truncated, dropped, escalate_hint}`. Diff mode adds `diff {range, changed_paths, targets, unmatched_paths}`.

## Budget

1500 estimated tokens by default. If `budget.truncated` is true and you need the dropped nodes, rerun with the budget named in `budget.escalate_hint`.

## Rules

- Evidence or silence: every node and edge you repeat must keep its locator (`repo@sha:path#Lstart-Lend`). If a path is listed in `unsupported_boundaries` or `diff.unmatched_paths`, say `unknown`; never infer what it does.
- Narrate, don't infer: you may summarize the briefing, but label any sentence that is not a direct rendering of a GIRA fact as `narrated` and cite the locators it paraphrases.
