---
schema: task/1
tier: standard
status: verified
verified_at: "9de69b2"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-01-eval-end-to-end]
---

# Hard budgets (SC6)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Deterministic token estimator; truncation in priority order; `budget` object on the CLI `--budget` flag and on MCP `budget`.

In scope: spec §A items covered by SC6.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC6

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 4d197c0 | clean worktree at 4d197c0; taken after the first edits had started, on the unedited commit |
| AC1 | `go test -race ./cmd/gira -run SC6 -v` and `go test -race ./internal/slices -run Fit` | pass | 9de69b2 | 10 of 12 tasks x 3 tiers x json/md/mcp within budget; repeat byte-identical; focal kept; tasks 05/06 have no node (see Outcome) |
| R1 | `go test -race ./...` | pass | 9de69b2 | all packages ok; existing e2e unchanged |
| R2 | `node tools/validate/validate.js` | pass | 9de69b2 | overall pass |
| V1 | read the diff against the ACs | pass | 9de69b2 | scope: slices budget and render, CLI impact/slice flags, MCP budget arg, eval.Hermetic helper |

## Plan

skip: one shared fitter in `internal/slices`, then the CLI and MCP wiring; the order is fixed by the dependency.

## Outcome

Delivered:
- `slices.Budget` (`limit`, `estimated`, `truncated`, `dropped{nodes, relationships, evidence}`, `escalate_hint`), an optional `budget` field on `ContextSlice` and `ImpactBriefing`, and a Markdown budget marker;
- `ContextSlice.Emit` and `ImpactBriefing.Emit`: they render the exact bytes the CLI writes (JSON, or Markdown plus its newline) and fit them to the ceiling in the spec's order: extra evidence; then nodes, farthest first (hops over the briefing's mediation edges) and then by descending ID, together with their relationships. The focal node or target, the unsupported boundaries and the marker are never dropped. `estimated` settles to `tokens.Estimate` of the emitted bytes, the same estimator the evaluator uses. If the parts that are never dropped don't fit, the result is a `budget_exceeded` error;
- CLI `--budget` on `impact` and `slice`; MCP `budget` on `get_impact_briefing` and `get_context_slice`;
- default budgets as constants (summary 500, contracts 1,500, deep 3,500, briefing 1,500).

Budgets are opt-in. Without `--budget` or `budget`, the output is byte-identical to before (test `TestNoBudgetKeepsOutputUnchanged`). That keeps the change additive.

Coverage note for SC6: tasks 05 and 06 target `contract:kafka:orders`. The extractors never emit that node (child 01's known gap), so both commands return node-not-found and there is no output to budget. The other 10 tasks are checked in every tier, in JSON and Markdown, on both CLI and MCP.

