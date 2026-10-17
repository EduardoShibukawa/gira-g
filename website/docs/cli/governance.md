---
sidebar_position: 12
---

# Governance & Verification (`eval`, `verify`, `review`, `gate`, `gate-lint`)

Gira includes five deterministic governance and verification commands (F1–F5) for auditing evaluation artifacts, control totals, token accounting, L3 findings registers, and gate ledgers.

## `gira eval`

Run the deterministic evaluation harness stub and emit the canonical `ARTIFACT-VERIFIED` closing line (F1), optional seeded control-total blind re-derivation (F2), and measured token accounting (F4).

```text
gira eval [--check <label>] [--expect <marker>] [--cwd <dir>] [--taxonomy measured|checked|inferred|assumed] [--control-total] [--sample max20-5pct] [--seed <uint>] [--token-report measured] [--artifact-verified] [--no-artifact-verified]
```

| Flag | Default | Description |
|---|---|---|
| `--check` | `gira eval` | `CHECK:` command label recorded in the digest |
| `--expect` | `ok` | `EXPECT:` success marker recorded in the digest |
| `--cwd` | current dir | Working directory recorded in the digest |
| `--taxonomy` | `inferred` | L3 claim taxonomy (`measured` \| `checked` \| `inferred` \| `assumed`) |
| `--control-total` | `false` | Emit `control:{total,matched,mismatched,sampled_n,seed}` line via seeded blind re-derivation (F2) |
| `--sample` | `max20-5pct` | Sampling rule (`max(20, 5%)`) |
| `--seed` | `0` | Seed for the deterministic sampler |
| `--token-report` | — | Set to `measured` to emit `tokens:{gross,expansions,net,measured}` (F4) |
| `--artifact-verified` | `true` | Emit the `ARTIFACT-VERIFIED` closing line |
| `--no-artifact-verified` | `false` | Debug-only: suppress the closing line |

## `gira verify`

Recompute and print the `ARTIFACT-VERIFIED` closing line from frozen inputs without running an evaluation (F1).

```text
gira verify --check <label> --expect <marker> [--cwd <dir>] [--fingerprint <hex>] [--taxonomy measured|checked|inferred|assumed] [--exit <code>]
```

| Flag | Default | Description |
|---|---|---|
| `--check` | `""` | `CHECK:` command label |
| `--expect` | `""` | `EXPECT:` success marker |
| `--cwd` | `.` | Working directory recorded in the digest |
| `--fingerprint` | `""` | Frozen input fingerprint hex (when empty, computed from `CHECK`, `EXPECT`, and `gira.yaml` if present) |
| `--taxonomy` | `inferred` | L3 claim taxonomy (`measured` \| `checked` \| `inferred` \| `assumed`) |
| `--exit` | `0` | Exit code recorded in the digest |

## `gira review`

Inspect the L3 findings register, independently re-verify a finding's reproduction path (`--fix-check`), or compact resolved findings (`--gc`) (F3, F5).

```text
gira review [--findings-register <path>] [--fix-check --finding <id> --expect-content <str>] [--gc [--keep-resolved <n>]]
```

| Flag | Default | Description |
|---|---|---|
| `--findings-register` | `reviews/findings.md` | Path to the findings register |
| `--fix-check` | `false` | Independently read the finding's recorded repro path and compare with `--expect-content` |
| `--finding` | `""` | Finding ID to verify when `--fix-check` is set |
| `--expect-content` | `""` | Expected bytes at the finding's repro path |
| `--gc` | `false` | Compact the register: keep all `open` findings in order and prune resolved (`fixed`/`closed`/`verified`) entries |
| `--keep-resolved` | `0` | Number of most-recent resolved findings to retain during `--gc` |

## `gira gate`

Separate non-executing gate status inspection (`--status`) from active digest re-verification (`--reverify`) (F5).

```text
gira gate --status | --reverify [--state <path>] [--digest <hex>]
```

| Flag | Default | Description |
|---|---|---|
| `--status` | `false` | Non-executing read of the gate ledger |
| `--reverify` | `false` | Re-run verification against `--digest` |
| `--state` | `.gira/gate.json` | Gate ledger file path |
| `--digest` | `""` | Expected digest hex for `--reverify` |

## `gira gate-lint`

Validate the gate ledger state and digest (F5). By default, CI runs (`CI=1` or `GIRA_CI=1`) are warn-only unless `--fail-on-ci` is passed; local runs with `--strict` fail closed.

```text
gira gate-lint [--strict] [--fail-on-ci] [--state <path>] [--digest <hex>]
```

| Flag | Default | Description |
|---|---|---|
| `--strict` | `false` | Enable strict gate lint (fail-closed locally, warn-only in CI by default) |
| `--fail-on-ci` | `false` | Enforce fail-closed `--strict` even under CI |
| `--state` | `.gira/gate.json` | Gate ledger file path |
| `--digest` | `""` | Expected digest hex |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Verification, review, or gate check passed |
| `1` | Invalid flags, tampered digest, failed fix-check, or strict gate-lint failure |

## See also

- [`gira serve`](./serve.md) — MCP tools `verify_artifact`, `get_control_total`, `get_findings_register`, `get_token_accounting`, and `gate_status`
- [Evaluation Scorecard](../concepts/scorecard.md)
