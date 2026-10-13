---
schema: task/1
tier: standard
status: verified
verified_at: "2248e07"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-08-bench-quarantine]
---

# Layer0 normalizer (SC13)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

NAG schema, adapters for GIRA, Graphify and UA, locator checker.

In scope: spec §B items covered by SC13.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC13

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | b1d2af9 | all packages ok before any edit |
| AC1 | `go test -race ./...` inside `bench/` (layer0 goldens, locator checker; quarantine still holds) | pass | 2248e07 | 3 adapter goldens byte-identical and deterministic; locator checker accepts valid and rejects out-of-bounds, anchor mismatch and missing file; quarantine holds |
| R1 | `go test -race ./...` at the repository root | pass | 2248e07 | all root packages ok |
| R2 | `node tools/validate/validate.js` | pass | 2248e07 | overall pass |
| V1 | read the diff against the ACs | pass | 2248e07 | scope: bench/layer0, ground-truth component_paths, eval.GroundTruth field |

## Plan

skip: the NAG and the lifter, then the three adapters, then the locator checker and the goldens.

## Outcome

Delivered `bench/layer0`:
- the NAG: Components, Contracts, and edges `{source, target, verb, evidence[{path, start, end}]}` whose verb is one of provides, consumes, imports, accesses, publishes, subscribes or unknown, plus `projected`, which reports when a tool's graph was lifted;
- `Lifter`: maps files to Components by longest path prefix, using the ground truth's new `component_paths` table. That table is shared by every tool and renamed by `Perturbed()`;
- `FromGIRA`: Service and Infrastructure nodes are Components. A consumer of Contract X consumes each provider of X, citing the consumer's evidence. Library nodes are not Components;
- `FromGraphify` (documented `graph.json`): link order gives the direction. `imports`/`imports_from` map to imports and everything else to unknown. Evidence is `source_file` plus the `L<n>` line;
- `FromUA` (documented `knowledge-graph.json`): `direction` is honoured. Only `imports` maps to a verb. Edges have no evidence, and the adapter never borrows a node's location for one;
- `CheckLocator`: the file must exist at the pinned SHA, the range must be in bounds, and the range must contain the ground-truth anchor.

SC13 evidence: three hand-written goldens (`testdata/*/expected.nag.json`) match the adapters' output byte for byte, and each adapter is deterministic. The locator table test accepts 2 valid locators and rejects 3 out-of-bounds locators, 1 locator that misses its anchor and 1 missing file.

Root change: the ground truth gains `component_paths`, and `eval.GroundTruth` gets the matching field, which the internal evaluator ignores. The bench reads the fixture; nothing in the root references the bench.

