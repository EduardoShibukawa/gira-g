---
name: gira-query
description: Ask GIRA a structured architecture question (impact, slice, explain, locate, tour, PR impact, diff) and get the same JSON whether the GIRA MCP server is registered or only the gira CLI is installed. Use whenever you need a single GIRA fact during implementation.
operations: ["MCP get_impact_briefing, get_context_slice, explain_evidence, locate_boundary, get_tour, get_pr_impact, diff_maps", "gira impact|slice|explain|locate|tour|diff ... --format json"]
output_contract: 'the JSON of the operation called, or an error object {"error": {"code": "index_missing|index_stale|node_not_found|budget_exceeded", "message": "...", "suggestions": [...]}}'
budget: per call (pass the operation's budget)
---
# gira-query

Structured lookup with graceful fallback.

## Operations

1. If GIRA MCP tools are registered (`get_impact_briefing` and friends), call the tool.
2. Otherwise run the equivalent CLI command with `--format json`:

| Question | MCP tool | CLI |
|---|---|---|
| impact | `get_impact_briefing {target_id, depth, budget}` | `gira impact <target_id> --depth <d> --budget <b> --format json` |
| slice | `get_context_slice {node_id, tier, view, budget}` | `gira slice <node_id> --tier <t> --view <v> --budget <b> --format json` |
| explain | `explain_evidence {node_id, edge_id}` | `gira explain --node <id> --edge <id> --format json` |
| locate | `locate_boundary {query}` | `gira locate <query> --format json` |
| tour | `get_tour {focus, max_steps}` | `gira tour --focus <id> --max-steps <n> --format json` |
| PR impact | `get_pr_impact {range, budget}` | `gira impact --from-diff <range> --budget <b> --format json` |
| diff | `diff_maps {index_a, index_b}` | `gira diff <index_a> <index_b> --format json` |

Both paths return the same JSON document for the same question.

## Output contract

The operation's JSON, or an error object `{"error": {"code", "message", "suggestions"}}` with one of these codes:

- `index_missing`: no index was found. Run `gira index` first.
- `index_stale`: the index cites a commit that is not the repository's HEAD. Re-index before trusting it.
- `node_not_found`: the ID is not in the index. `suggestions` lists `locate_boundary` matches for it.
- `budget_exceeded`: even the parts that are never dropped exceed the budget. Raise it.

## Budget

Per call: pass the budget of the skill you are serving (500/1500/3500 for slices, 1500 for impact).

## Rules

- Evidence or silence: never restate a GIRA fact without its locator; on `node_not_found` say so instead of guessing.
