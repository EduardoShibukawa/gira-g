---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# Envelope on all tools (R1, R2; SC1, SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)

## Objective

Optional `budget`/`--budget` on get_tour, diff_maps, explain_evidence, locate_boundary and CLI tour/flows/diff/explain/locate, reusing `slices.Budget`. Absent/0 keeps today's bytes exactly. Deterministic per-kind drop order that never drops the focal element nor unsupported/unknown rows. Same index + budget → byte-identical output.

In scope: SC1, SC2, test-first. Failing test first: envelope missing on a tool must fail; non-deterministic drop must fail.
Out of scope: honesty labels (02), structured errors (03), single registry (04), CI hygiene (05); ground-truth changes.

## Acceptance criteria

- AC1 = SC1 (envelope coverage)
- AC2 = SC2 (determinism)
- AC3 = SC8–SC10 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | not run | | |
| AC1 | `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -v -count=1` | not run | | |
| AC2 | `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -v -count=1` | not run | | |
| AC3 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | not run | | |
| R1 | `go test ./internal/mcp ./internal/slices ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1–AC3 + spec R1, R2 | not run | | |

## Constraints and budget

- Additive only: new fields on existing tools, new CLI flags; no existing flag/field meaning change.
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012.
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.
- Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety/determinism 100%; SC15 ≥ 90%.
- Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- Unbudgeted calls byte-identical (gira-query goldens unchanged).

## Plan

1. Write failing `TestSCEnvelopeEveryTool` first (missing envelope on get_tour/diff_maps/explain_evidence/locate_boundary + CLI tour/flows/diff/explain/locate must fail).
2. Write failing `TestSCEnvelopeDeterministic` first (same index + budget → non-byte-identical must fail; drop order violates rules must fail).
3. Implement envelope wiring on MCP tools (get_tour, diff_maps, explain_evidence, locate_boundary) reusing `slices.Budget`.
4. Implement `--budget` on CLI tour, flows, diff, explain, locate.
5. Implement deterministic per-kind drop order in `internal/slices` (tour: trailing steps then extra read locators; diff: evidence_moved then unchanged context, never changed_contracts or newly_unsupported; explain: evidence beyond first per fact; locate: suggestions beyond N).
6. Run AC1, AC2, AC3, R1, V1; commit; validator.

## Outcome