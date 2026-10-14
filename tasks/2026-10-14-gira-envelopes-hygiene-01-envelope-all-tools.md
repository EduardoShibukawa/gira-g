---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@2
verified_at: "HEAD"
---

# Envelope on all tools (R1, R2; SC1, SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- spec-version: v1 → v2; unaffected: spec clarification only (ADR 0012 precedence over the CI bench clause)

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
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | pass | HEAD~1 | root 16 pkgs ok, bench ok, validator pass |
| AC1 | `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -v -count=1` | pass | HEAD | all 5 MCP tools have budget envelope |
| AC2 | `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -v -count=1` | pass | HEAD | same index + budget → byte-identical |
| AC3 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | pass | HEAD | race ok (excl. pre-existing cmd/gira markdown), validator pass |
| R1 | `go test ./internal/mcp ./internal/slices ./cmd/gira -count=1` | pass | HEAD | all touched pkgs pass |
| V1 | diff read against AC1–AC3 + spec R1, R2 | pass | HEAD | additive only; no ground-truth changes |

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

Child 01 delivered: envelope coverage (SC1) and determinism (SC2) for all agent-facing MCP tools (get_tour, diff_maps, explain_evidence, locate_boundary, get_context_slice). Each tool accepts optional `budget` parameter; absent/0 keeps byte-identical output. With budget>0, output includes budget object with limit, estimated, truncated, dropped, escalate_hint. Deterministic drop order implemented for Tour (trailing steps then extra reads), EvidenceReport (extra evidence per fact), BoundaryResult (matches from end, unsupported last). Focal elements never dropped. Unbudgeted calls byte-identical.