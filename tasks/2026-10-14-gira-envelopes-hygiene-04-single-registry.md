---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@3
verified_at: "HEAD"
---

# Single tool registry (R6; SC6)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- spec-version: v1 → v2; unaffected: spec clarification only (ADR 0012 precedence over the CI bench clause)
- spec-version: v2 → v3; unaffected: policy definition only (Decisions/Open questions, no code, gate, or criteria change)
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
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | pass | HEAD~1 | root 16 pkgs ok, bench ok, validator pass |
| AC1 | `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -v -count=1` | pass | HEAD | MCP tools match registry, enveloped field present |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | pass | HEAD | race ok, validator pass |
| R1 | `go test ./internal/mcp ./internal/tokens -count=1` | pass | HEAD | all touched pkgs pass |
| V1 | diff read against AC1–AC2 + spec R6 | pass | HEAD | additive only; no ground-truth changes |

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

Child 04 delivered: single tool registry (SC6). Created `internal/tools/registry.go` as the single source of truth for all tool definitions. Each tool declares `enveloped: bool`. Updated `internal/tokens/accounting.go` to use the registry via `ListTools()`. Updated MCP server `getToolDefinitions()` to include `Enveloped: true` for all 12 tools. Set-equality test ensures both sides stay in sync. All SC6 tests pass.