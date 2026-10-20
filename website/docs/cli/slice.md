---
sidebar_position: 3
---

# `gira slice`

Extract a bounded context slice around a Component, Contract, or node.

## Synopsis

```text
gira slice <node_id> [--tier <tier>] [--view <view>] [--budget <tokens>] [--format json|md|caveman] [--index <path>]
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--tier` | string | `summary` | `summary` \| `contracts` \| `deep` |
| `--view` | string | `full` | `full` \| `architecture` \| `governance` |
| `--budget` | int | (tier default) | Hard ceiling on estimated output tokens (500/1500/3500 per tier) |
| `--checkout` | string | — | Local checkout for staleness comparison |
| `--format` | string | `json` | `json` \| `md` \| `caveman` (alias `llm`) |
| `--index` | string | — | Path to `macro-map.json` |

## Output (JSON)

A bounded context slice containing the selected Components, Contracts, and
relationships at the requested tier and view, with evidence locators and
derived token estimate.

## Example

```bash
gira slice service:checkout --tier contracts --view architecture --index .gira/index.json
```

### Agent format (`caveman`)

`--format=caveman` renders one fact per line for LLM context budgets:

```text
Component: Service[service:checkout] @ src/checkout/main.go:L12
Contracts:
  - Contract[contract:Checkout] @ pb/checkout.proto:L5
    - Checkout(CheckoutRequest) -> CheckoutResponse
Dependencies:
  - imports library:payments @ src/checkout/main.go:L4
Impact:
  - consumed by Service[service:frontend] @ web/src/api.ts:L3
Budget: limit=400 estimated=212 truncated=false
```

The budget fitter drops whole facts (farthest neighbors first), never a line
in the middle. Any other value for `--format` fails with
`json|md|caveman` instead of silently falling back to JSON.

## Notes

- `--tier deep` renders the widest context; `summary` is the Change-maker
  default for a quick scan.
- `--view governance` keeps only `Policy`, `Task` and `Skill` neighbors
  (see `isNodeAllowedInView` in `internal/slices/engine.go`). A focal
  `Service`/`Contract` with no such neighbors renders with an empty
  neighborhood — that is the filter working, not missing data. `full`
  shows everything; `architecture` keeps Service/Contract/Infrastructure/
  Library.

## See also

- [Impact](./impact.md)
- [Quickstart guide](../guides/quickstart.md)

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.
