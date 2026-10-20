---
sidebar_position: 2
---

# `gira impact`

Compute the Impact Set (affected components, contracts, consumers) for a target
Component, contract, or a git diff.

## Synopsis

```text
gira impact <target_id> [--depth <int>] [--budget <tokens>] [--index <path>]
gira impact --from-diff <base>..<head> [--repo <dir>] [--repo-id <id>]
```

## Global flags (see [`gira index`](./index.md) context)

| Flag | Description |
|---|---|
| `--index` | Path to `macro-map.json` |
| `--format` | `json`, `md` or `caveman` (default `json`) |

## Local flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--depth` | int | 1 | Graph traversal depth when expanding the Impact Set |
| `--budget` | int | 1500 | Hard ceiling on estimated output tokens |
| `--checkout` | string | — | Local checkout whose HEAD is compared against indexed commits (staleness) |
| `--from-diff` | string | — | Start from a git revision range `<base>..<head>` instead of a target id |
| `--repo` | string | `.` | Repository the `--from-diff` range is read from |
| `--repo-id` | string | — | Only match locators of this repository id (`--from-diff`) |

## Output (JSON)

An Impact Briefing containing the affected Components, connecting Contracts,
and direct consumers, each with evidence locators (`[repo@commit:path#Lstart-Lend]`)
and derivation tiers (`declared`/`extracted`/`inferred`). The indexed commit
SHA is printed; locators pointing at other commits are flagged `stale`.

## Examples

```bash
# Impact of a contract
gira impact kafka:orders.created --index .gira/index.json

# Impact of all nodes a diff touches
gira impact --from-diff HEAD~1..HEAD --repo .
```

## See also

- [Slice](./slice.md)
- [Concepts: Impact Set](../concepts/impact-set.md)

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.
