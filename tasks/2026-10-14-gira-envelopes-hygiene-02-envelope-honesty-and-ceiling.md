---
schema: task/1
tier: standard
status: pending
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# Envelope honesty and ceiling (R3, R4; SC3, SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- depends on: 2026-10-14-gira-envelopes-hygiene-01-envelope-all-tools

## Objective

Honest label: `budget` gains `estimator: "bytes/4"`, `measured: false` (additive, only when a budget is set). No envelope figure enters an E1 measured headline. Transport ceiling: one global MCP response byte ceiling, set from measurement above every current hermetic output; above it, refuse with `budget_exceeded` + `escalate_hint`; never cut mid-JSON.

In scope: SC3, SC4, test-first. Failing test first: missing honesty labels must fail; ceiling exceeded without structured refusal must fail.
Out of scope: envelope coverage/determinism (01), structured errors (03), single registry (04), CI hygiene (05); ground-truth changes.

## Acceptance criteria

- AC1 = SC3 (honesty)
- AC2 = SC4 (transport ceiling)
- AC3 = SC8–SC10 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | pending | | |
| AC1 | `go test ./internal/tokens ./internal/slices -run SCEnvelopeHonesty -v -count=1` | pending | | |
| AC2 | `go test ./internal/mcp -run SCTransportCeiling -v -count=1` | pending | | |
| AC3 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | pending | | |
| R1 | `go test ./internal/mcp ./internal/tokens ./internal/slices -count=1` | pending | | |
| V1 | diff read against AC1–AC3 + spec R3, R4 | pending | | |

## Constraints and budget

- Additive only: `estimator` and `measured` fields only when budget is set; global ceiling constant; refusal error.
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012.
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.
- Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety/determinism 100%; SC15 ≥ 90%.
- Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- Unbudgeted calls byte-identical (gira-query goldens unchanged).
- Ceiling measured from current hermetic outputs + margin; `budget_exceeded` + `escalate_hint`; no pagination in v1.

## Plan

1. Write failing `TestSCEnvelopeHonesty` first (budget without `estimator`/`measured` must fail; envelope figure in measured headline must fail).
2. Write failing `TestSCTransportCeiling` first (response over ceiling without `budget_exceeded` + `escalate_hint` must fail; mid-JSON cut must fail).
3. Add `estimator: "bytes/4"` and `measured: false` to `budget` JSON (only when budget is set).
4. Implement global MCP response byte ceiling in `internal/mcp/server.go` (measured from current hermetic outputs + margin).
5. Implement refusal with `budget_exceeded` code + `escalate_hint` suggestion; ensure never cut mid-JSON.
6. Run AC1, AC2, AC3, R1, V1; commit; validator.

## Outcome