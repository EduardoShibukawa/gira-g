---
sidebar_position: 4
---

# `gira serve`

Start the stdio MCP protocol adapter for AI agent clients.

## Synopsis

```text
gira serve mcp [--index <path>]
```

## Arguments

| Argument | Description |
|---|---|
| `mcp` | Required: enables the stdio MCP adapter |

## Flags

| Flag | Description |
|---|---|
| `--index` | Path to `macro-map.json` (auto-resolves `.gira/macro-map.json`, `macro-map.json`, or `~/.cache/gira/indexes/<system>/macro-map.json` when omitted) |

## Exposed MCP Tools (13 tools)

The MCP server speaks JSON-RPC 2.0 over stdio (supporting both newline-delimited JSON and `Content-Length` framing) and exposes 13 deterministic tools:

### Architecture & Impact Tools

| Tool Name | Required Parameters | Optional Parameters | Description |
|---|---|---|---|
| `get_impact_briefing` | `target_id` | `depth`, `budget` | Analyzes upstream services, downstream consumers, mediating contracts, and unsupported boundaries affected by a change to `target_id`. |
| `get_pr_impact` | `range` | `diff`, `repo_path`, `repo_id`, `depth`, `budget` | Maps changed line ranges of a Git revision range (`<base>..<head>`) or unified diff to nodes whose evidence they touch, merging their impact briefings. |
| `get_tour` | — | `focus`, `max_steps`, `budget` | Deterministic guided reading tour walking `consumes`/`provides`/`accesses` edges breadth-first from entry points or `focus`. |
| `get_doc` | `service_id` | `budget` | Single-service onboarding document: entrypoints, symbols, contracts, imports, calls, manifests, change dirs, tour, flows, and known unknowns (`gira doc --format json`). |
| `diff_maps` | `index_a`, `index_b` | `budget` | Architecture diff between two index files: structural renames, added/removed nodes and edges, changed contracts, and newly unsupported paths. |
| `get_context_slice` | `node_id` | `tier` (`summary`\|`contracts`\|`deep`), `view` (`full`\|`architecture`\|`governance`), `budget` | Retrieves a bounded context slice around `node_id` across disclosure tiers and views. |
| `explain_evidence` | — | `node_id`, `edge_id`, `budget` | Inspects provenance locators, derivation tier, deployment profile, and freshness status for a node or relationship. |
| `locate_boundary` | `query` | `budget` | Finds architectural boundary components, contract RPC/HTTP methods, or unsupported files matching `query`. |

### Governance & Verification Tools (F1–F5)

| Tool Name | Parameters | Description |
|---|---|---|
| `verify_artifact` | `check`, `expect`, `cwd`, `exit_code`, `fingerprint`, `taxonomy` | Recomputes the L1 artifact digest and returns `{digest, closing_line}` (F1). |
| `get_control_total` | `total`, `matched`, `mismatched`, `seed` | Validates `control:{total,matched,mismatched,sampled_n,seed}` against the seeded `max(20, 5%)` sampling rule (F2). |
| `get_findings_register` | `register` | Returns the L3 findings register entries outside context slices (F3). |
| `get_token_accounting` | `text`, `expansions` | Returns measured token accounting `{gross, expansions, net, measured_bool}` (F4). |
| `gate_status` | `state` | Non-executing gate status read for a gate ledger file (F5). |

## Example

```bash
gira serve mcp --index .gira/macro-map.json
```

## Notes

- `gira` never runs a persistent web server daemon (ADR 0008). The only server surface is the stdio MCP adapter used by coding agents.
- Invoke with `gira serve mcp` (the `mcp` subcommand is required).
- Use [`gira skills install`](./skills.md) to automatically register `gira serve mcp` in Claude Code, Cursor, or `.agents` configurations.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (adapter running until stdin closes) |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [`gira skills`](./skills.md)
- [`gira doc`](./doc.md)
- [`gira slice`](./slice.md)
- [Governance & Verification](./governance.md)
