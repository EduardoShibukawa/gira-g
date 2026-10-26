---
sidebar_position: 11
---

# `gira skills`

Generate, emit, or check the GIRA skills and MCP registration for an agent
host.

## Synopsis

```text
gira skills install --host <host> [--dir <project>] [--index <path>] [--command <bin>] [--hook]
gira skills manifest [--format json|md] [--output <path>]
gira skills check --host <host> [--dir <project>] [--index <path>] [--command <bin>] [--hook]
```

The hidden `gira skills nudge` subcommand is always exit 0 and never blocks.

## Subcommands and flags

| Subcommand | Flags |
|---|---|
| `install` | `--host` (claude\|cursor\|agents), `--dir` (default `.`), `--index`, `--command` (default `gira`), `--hook` (claude opt-in PreToolUse nudge, never blocks) |
| `manifest` | `--format` (`json`\|`md`, default `json`), `--output` (write to path instead of stdout) |
| `check` | `--host`, `--dir` (default `.`), `--index`, `--command` (default `gira`), `--hook` (also check the opt-in nudge) |

## Output

- `install`: a JSON report of the host files written.
- `manifest`: the deterministic skills catalog manifest (`json` or `md`).
- `check`: a non-mutating drift report for the installed host files.

## Example

```bash
gira skills install --host claude --dir . --index .gira/index.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (`nudge` always exits 0) |
| `1` | Usage, flag, index-load, engine, or emit error (`check` reports drift as an error) |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Serve](./serve.md)
