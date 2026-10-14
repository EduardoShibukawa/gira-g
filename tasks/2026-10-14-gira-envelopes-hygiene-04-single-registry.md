---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# Single tool registry (R6; SC6)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- depends_on: ["2026-10-14-gira-envelopes-hygiene-01-envelope-all-tools"]

## Objective

Single tool registry (or a set-equality test); each tool declares `enveloped`.

In scope: SC6, test-first. Failing test first: duplicate registry must fail; tool missing `enveloped` declaration must fail.
Out of scope: envelope coverage/determinism (01), honesty/ceiling (02), structured errors (03), CI hygiene (05); ground-truth changes.

## Acceptance criteria

- AC1 = SC6 (single registry)
- AC2 = SC8–SC10 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | not run | | |
| AC1 | `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -v -count=1` | not run | | |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | not run | | |
| R1 | `go test ./internal/mcp ./internal/tokens -count=1` | not run | | |
| V1 | diff read against AC1–AC2 + spec R6 | not run | | |

## Constraints and budget

- Additive only: single source of truth for tool definitions; `enveloped` boolean per tool.
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012.
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.
- Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety/determinism 100%; SC15 ≥ 90%.
- Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- Unbudgeted calls byte-identical (gira-query goldens unchanged).
- Duplicated registry at `internal/tokens/accounting.go:23` vs `internal/mcp/server.go:285` must be unified.

## Plan

1. Write failing `TestSCToolRegistrySingleSource` first (two registries with different tool sets must fail; tool without `enveloped` declaration must fail).
2. Create single source of truth for tool definitions (e.g., `internal/mcp/tools.go` or `internal/tools/registry.go`).
3. Each tool declares `enveloped: bool`.
4. Update `internal/tokens/accounting.go` and `internal/mcp/server.go` to use the single registry.
5. Add set-equality test to ensure both sides stay in sync.
6. Run AC1, AC2, R1, V1; commit; validator.

## Outcome