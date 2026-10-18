---
sidebar_position: 8
---

# `gira doc`

Render a single-service onboarding document from the Macro Map: Markdown for humans or budget-enveloped JSON for AI coding agents, backed by the exact same extracted facts.

## Synopsis

```text
gira doc <service-id> [--format md|json|html] [--index <path>] [--checkout <dir>] [--budget <tokens>]
```

## Arguments

| Argument | Description |
|---|---|
| `service-id` | Canonical ID of the service node to document (for example `service:order` or `service:checkout`) |

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--format` | string | `md` | Output format: `md` (Markdown), `json` (agent envelope), or `html` (self-contained single-file HTML) |
| `--index` | string | — | Path to `macro-map.json` (auto-resolves `.gira/macro-map.json`, `macro-map.json`, or user cache when omitted) |
| `--checkout` | string | — | Local checkout directory whose `HEAD` is compared against indexed commits for staleness warnings on `stderr` |
| `--budget` | int | `0` | Hard ceiling on estimated output tokens (`0` = no budget ceiling) |

## Output

Both `--format md` and `--format json` emit the same underlying service onboarding structure:
- **Entrypoints & ports** — declared listening ports, main packages, and runtime entry locators
- **Provided & consumed contracts** — Protobuf/gRPC, OpenAPI, AsyncAPI, and GraphQL boundaries with evidence locators
- **Infrastructure & data accesses** — databases, caches, and topics accessed by the service
- **Symbols, imports & calls** — extracted package/module symbols and call-graph edges
- **Run & test manifests** — detected build, compose, and workflow files
- **Change directories** — primary directory scopes touched when modifying this service
- **Guided tour & flows** — deterministic reading order and request-to-storage flows rooted at the service
- **Known unknowns** — unsupported files or dynamic boundaries adjacent to the service

## Examples

```bash
# Render a human-readable Markdown onboarding guide for service:order
gira doc service:order --format md --index .gira/macro-map.json

# Emit a token-budgeted JSON document for an AI coding agent
gira doc service:order --format json --budget 1500 --index .gira/macro-map.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Missing `--format`, unsupported format (`html`), unknown service ID, or index load error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [`gira tour`](./tour.md) — guided reading tour across the system
- [`gira flows`](./flows.md) — deterministic chains from entry points to infrastructure
- [`gira serve`](./serve.md) — exposes `get_doc` over MCP stdio
