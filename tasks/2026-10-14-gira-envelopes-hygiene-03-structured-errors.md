---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# Structured errors (R5; SC5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- depends_on: ["2026-10-14-gira-envelopes-hygiene-01-envelope-all-tools"]

## Objective

Structured errors `{code, message, retryable, suggestions}` on MCP and CLI `--format json`; gira-query reads `code` first, substring fallback for old binaries.

In scope: SC5, test-first. Failing test first: unstructured error on MCP/CLI must fail; gira-query not reading `code` first must fail.
Out of scope: envelope coverage/determinism (01), honesty/ceiling (02), single registry (04), CI hygiene (05); ground-truth changes.

## Acceptance criteria

- AC1 = SC5 (structured errors)
- AC2 = SC8–SC10 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | not run | | |
| AC1 | `go test ./internal/mcp ./internal/skills ./cmd/gira -run SCStructuredErrors -v -count=1` | not run | | |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | not run | | |
| R1 | `go test ./internal/mcp ./internal/skills ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1–AC2 + spec R5 | not run | | |

## Constraints and budget

- Additive only: new error structure; substring fallback for old binaries.
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012.
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.
- Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety/determinism 100%; SC15 ≥ 90%.
- Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- Unbudgeted calls byte-identical (gira-query goldens unchanged).
- Error codes: `budget_exceeded`, `envelope_error`, `tool_not_found`, `invalid_budget`, `internal_error` (extensible).

## Plan

1. Write failing `TestSCStructuredErrors` first (MCP error without `{code,message,retryable,suggestions}` must fail; CLI `--format json` error without structure must fail; gira-query substring fallback must fail).
2. Define error code vocabulary in `internal/mcp/errors.go` (or similar).
3. Update MCP tools to return structured errors.
4. Update CLI commands (tour, flows, diff, explain, locate) to emit structured errors under `--format json`.
5. Update `internal/skills/query.go` to read `code` first, substring fallback for old binaries.
6. Run AC1, AC2, R1, V1; commit; validator.

## Outcome