---
schema: task/1
tier: standard
status: completed
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-03-hard-budgets]
---

# Tours and flows (SC8)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira tour`, `gira flows`, MCP `get_tour`.

In scope: spec §A items covered by SC8.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC8

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 3051f5b | all packages ok before any edit |
| AC1 | `go test -race ./cmd/gira -run SC8 and -run Flows and -run FlowAnchors -v` | not run | | |
| R1 | `go test -race ./...` | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| V1 | read the diff against the ACs | not run | | |

## Plan

skip: the engine (tour, flows), then the ground-truth flow and the coverage proxy, then the CLI and MCP wiring.

## Outcome

Delivered:
- `Engine.GetTour(focus, maxSteps)`. The walk starts from `--focus`, or from the entry points: Services with declared `ports`. It is breadth-first over outgoing `consumes`, `provides` and `accesses` edges; a Contract leads on to its providers. Within a level, stops are ranked by fan-in, then by ID. Stops are runtime Components only. Libraries are skipped, and so are Services carrying a `stage` key (Dockerfile build stages). Each step has `order`, `title`, `node_ids` (the stop, plus the Contracts it consumes or provides), `read` (locators) and `why`. `why` is only ever `RenderWhy(why_template, why_args)`; both fields are emitted so the rendering can be checked;
- `Engine.GetFlows(entry)`: simple paths from the entry through `consumes`/`accesses`, and through providers when a hop is a Contract, ending at nodes with nowhere further to go. They are depth-capped at 6 and count-capped at 50. Every hop cites its edge's locator;
- CLI `gira tour [--focus] [--max-steps] [--format]` and `gira flows <entry> [--format]`; MCP `get_tour`;
- ground truth `flows`: `order-flow` with 11 steps, anchored in checkout's `PlaceOrder`. `Perturbed()` renames it. `eval.TourCoverage` is the critical-path proxy that Layer 2 reuses.

SC8 evidence on the hermetic index: tours with and without focus are byte-identical across runs (JSON and Markdown, CLI equals MCP). Every `read` locator cites the pinned commit and resolves in bounds. Every `why` re-renders from its template. Coverage of `order-flow` from `service:frontend` is 81.8% (9 of 11).

Known limit: the two misses are `service:accounting` and `service:fraud-detection`. They consume the kafka topic that checkout writes to, and a forward walk can't reach a subscriber. Reaching them needs a topic Contract between producer and subscribers, which the extractors don't emit yet (child 01's Kafka follow-up). Until they do, the 80% margin is thin.

