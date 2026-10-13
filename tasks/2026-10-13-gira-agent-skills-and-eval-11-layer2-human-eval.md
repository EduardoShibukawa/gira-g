---
schema: task/1
tier: standard
status: completed
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-05-tours-and-flows]
---

# Layer2 human eval (SC15)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Tour assertiveness and coverage proxies; moderated study protocol document.

In scope: spec §B items covered by SC15.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC15

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | b30c60f | all packages ok before any edit |
| AC1 | `go test -race ./...` inside `bench/` (layer2: proxies on GIRA tours from the real binary, hand-checked proxies, UA adapter, protocol document; quarantine still holds) | not run | | |
| R1 | `go test -race ./...` at the repository root | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| V1 | read the diff against the ACs | not run | | |

## Plan

skip: the proxies and adapters, then the tests and the protocol document. The shared bench test helper moved into `bench/internal/benchtest` so that layer1 and layer2 both use it.

## Outcome

Delivered:
- `bench/layer2`: a normalized tour (steps with node ids, locators and claims) and `Measure`, which computes:
  - assertiveness: the share of claims whose anchor appears in a cited range that resolves at the pinned SHA;
  - critical-path coverage against a ground-truth flow;
  - step count and words per step.
- Two tour adapters:
  - `FromGIRA`: one claim per step, anchored on the identifier its `why` template names;
  - `FromUA`: UA's documented `TourStep`. UA cites no line ranges, so a claim verifies only when a symbol it names exists in the step's files.
- `bench/layer2/study-protocol.md`: the moderated study. It covers the goal; a within-subject design over 3 tools × 3 corpus areas with a 3×3 Latin square; two tasks per area (including the spec's two questions); a blind 0–3 rubric; the instruments (time to first useful understanding, correctness, 5-point clarity, raw NASA-TLX); the procedure; data handling; and an analysis that makes no significance claims. Recruiting is still deferred, per the spec.

SC15 evidence, from GIRA tours produced by the real binary over the hermetic corpus:

| Tour | Steps | Assertiveness | Order-flow coverage | Words/step |
|---|---|---|---|---|
| `--focus service:frontend` | 17 | 100% (17/17) | 81.8% | 9.0 |
| entry points (no focus) | 18 | 61.1% (11/18) | 81.8% | 10.6 |

Finding, raised for the integrated review: all 7 failed claims are the entry-point steps (`why_template: entry`). Their claim "declares ingress on `${X_PORT}`" cites the node's first evidence, usually its `compose.yaml` line, and not the Dockerfile line that declares the port. Under I1 that claim lacks a supporting locator. This is a defect in child 05's tour, not in the proxy. One entry point is also `service:src`, the cart Dockerfile named after its `src/` directory, an extractor naming artifact.

