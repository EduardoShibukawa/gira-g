# GIRA ↔ G-Ingá `/work` integration contract

Specified by `specs/gira-agent-skills-and-eval.md` (A.3). This file fixes the contract only. The change that makes `/work` consume it belongs to the G-Ingá repository.

## Inputs `/work` sends

| SDLC moment | Operation | Arguments |
|---|---|---|
| Routing (signal S3, trigger H3) | `get_pr_impact`, or `gira impact --from-diff <base>..<head> --format json` | the working-tree or branch revision range, `budget: 1500` |
| Plan, before the first edit | `get_impact_briefing` / `get_context_slice` | the node ids the plan names, tier budgets 500/1500/3500 |
| Verification row | `diff_maps`, or `gira diff <base-index> <head-index> --format json` | indexes built at the base and the head |

## Outputs `/work` reads

- **S3/H3 evidence:** an `ImpactBriefing` with `diff.targets`. A target Contract in `affected_contracts` whose consumers (`downstream_services`) lie outside the touched Component is H3 evidence ("a contract consumed elsewhere changes"). `diff.unmatched_paths` are unknown to the index: they are neither evidence for nor against S3.
- **Verification:** a `Diff`. Any `changed_contracts`, `removed_edges` or `newly_unsupported` item that the task did not plan is a finding. `evidence_moved` is not a finding.
- **Budget signal:** when `budget.truncated` is true, `/work` either raises the budget to the value in `budget.escalate_hint` or records that the evidence was partial.

## Error vocabulary

`index_missing`, `index_stale`, `node_not_found` (with `suggestions`) and `budget_exceeded`, as returned by the `gira-query` procedure (`internal/skills/query.go`). On any of them, `/work` treats the GIRA signal as unavailable and keeps its other routing signals. A GIRA error is never read as "no impact".
