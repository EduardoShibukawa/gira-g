---
schema: task/1
tier: standard
status: completed
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-08-bench-quarantine, 2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Layer1 agent runner (SC14)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Agent runner with an MCP client, mock-model mode, telemetry schema, deterministic scorer.

In scope: spec §B items covered by SC14.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC14

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | adc6215 | all packages ok before any edit |
| AC1 | `go test -race ./...` inside `bench/` (layer1: SC14 mock run, scorer fixture, schema validation, endpoint guard; quarantine still holds) | not run | | |
| R1 | `go test -race ./...` at the repository root | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| V1 | read the diff against the ACs | not run | | |

## Constraints and budget

- No paid model call was made. Local Ollama is not installed on this machine (`localhost:11434` does not answer), so only mock mode ran.

## Plan

1. Write the model interface, the MCP client and the base tools.
2. Write the answer schema, the scorer, the telemetry schema and its validator.
3. Write the runner loop, the mock model and the OpenAI-compatible client with its paid-endpoint guard.
4. Write the SC14 test and the owner-run command.

## Outcome

Delivered `bench/layer1` and `bench/cmd/layer1`:
- `Runner.RunTask`: the fixed prompt plus the JSON answer schema, then a tool loop of at most 12 turns. Arm A gets read-only `grep`, `glob` and `read`, all confined to the corpus copy. Arm B adds GIRA's MCP tools through `MCPClient`, a stdio JSON-RPC client driving `gira serve mcp`. Transcripts go only to the git-ignored output dir.
- `MockModel` (`mock-policy/1`): a deterministic offline policy through the real loop. Arm B calls `get_impact_briefing` (budget 1500) and answers from it; arm A greps the target name and answers from the hits. Usage is the 4-bytes-per-token estimate.
- `OpenAIModel`: an OpenAI-compatible chat client for a local Ollama. It refuses any non-loopback endpoint unless `--approved-paid-run` names an owner approval. Model IDs are pinned with `--model-id` and recorded.
- `telemetry.schema.json`, plus a validator for the subset of JSON Schema it uses. Records carry `tokens_input`, `tokens_output`, `cache_read_tokens`, `tool_calls`, `tool_calls_by_tool`, `turns`, `wall_latency_ms`, `precision`, `recall`, `f1`, `locator_validity`, `hallucinated_nodes`, and the run metadata. No prompt text is recorded.
- `ScoreAnswer`: a deterministic set comparison against the ground truth, with the ignore list applied. Locators are checked with the Layer 0 rule. Hallucinated nodes are IDs outside the ground-truth universe.
- `Summarize`: median and IQR per cell. `cmd/layer1` writes `telemetry.jsonl` and `summary.json` and requires at least 5 repetitions.

SC14 evidence: 120 records (12 tasks × arms A and B × 5 repetitions) in mock mode, offline, against a `gira` binary built from this repository and an index built with `gira index`. Every record validates against the schema, every run answers, and the IQR is 0 across repetitions. The scorer reproduces the hand-computed fixture (P 0.6, R 0.75, F1 0.667, locator validity 0.25, 1 hallucinated id).

Limits:
- The mock's F1 values (A 0.55, B 0.38) test the loop, not a model. They are not a comparison.
- Arms C and D need their container query phases, so `cmd/layer1` refuses them for now.
- A read-only shell is not offered as a base tool, because the runner cannot sandbox one. Reports state this.

