---
schema: task/1
tier: standard
status: verified
verified_at: "9c324a5"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Impact from diff (SC7)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira impact --from-diff` and MCP `get_pr_impact`.

In scope: spec §A items covered by SC7.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC7

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 0301d36 | all packages ok before any edit |
| AC1 | `go test -race ./cmd/gira ./internal/gitdiff -run SC7 and -run Unmatched and -run Parse and -run Touches` | pass | 9c324a5 | proto PaymentService edit: CLI, MCP (git and supplied diff) impact set equals direct contract target; unmatched paths reported |
| R1 | `go test -race ./...` | pass | 9c324a5 | all packages ok; e2e unchanged |
| R2 | `node tools/validate/validate.js` | pass | 9c324a5 | overall pass |
| V1 | read the diff against the ACs | pass | 9c324a5 | scope: gitdiff package, slices GetPRImpact and diff field, CLI --from-diff, MCP get_pr_impact, tools-list test relaxed to superset |

## Plan

skip: the diff reader, the engine mapping, then the CLI and MCP wiring; the order is fixed by the dependency.

## Outcome

Delivered:
- `internal/gitdiff`: reads `git diff -U0 <base>..<head>` into base-side line hunks. An insertion counts as inside a range only when both neighbouring lines are. Option-like ranges are rejected;
- `Engine.GetPRImpact`: a node is a target when one of its evidence locators cites a changed range. A relationship whose evidence does makes its source a target. The briefings of all targets are merged. A changed file that no locator cites is listed in `diff.unmatched_paths` and never guessed. The briefing gains an optional `diff` object (`range`, `changed_paths`, `targets`, `unmatched_paths`), shown in Markdown too;
- CLI `gira impact --from-diff <base>..<head> [--repo <dir>] [--repo-id <id>]`, which composes with `--budget` and `--format`;
- MCP `get_pr_impact` (`range`, optional `diff` text, `repo_path`, `repo_id`, `depth`, `budget`). It emits the same JSON as the CLI.

SC7 evidence: adding `rpc Refund` inside `service PaymentService` gives the targets `[contract:PaymentService]`, and its Impact Set equals that of `gira impact contract:PaymentService`, through the CLI and through MCP, whether git runs or the diff is supplied.

One existing unit test changed: `TestServer_ToolsList` asserted exactly 4 MCP tools. It now asserts that the original 4 are still listed, because new tools are additive by spec. The e2e tests are unchanged.

