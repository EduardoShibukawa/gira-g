---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@2
depends_on: [2026-10-14-gira-evidence-gate-03-l3-taxonomy-register, 2026-10-14-gira-evidence-gate-04-token-honesty]
verified_at: "3e5f7f4"
---

# gate-lint and validator integration (F5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)
- spec-version: v1 → v2; unaffected: policy definition only (Decisions/Open questions, no code, gate, or criteria change)

## Objective

Port of the unlazy gate contract (idea, Go reimplementation): success-only marker + digest, stale-on-drift fail-closed, `--status`/`--reverify` split, `ABANDON` as qualified handoff that never reports `ALL MET`. Warn-only in CI until E3; fail-closed locally. Integrated SC1–SC6 plus the E0 recall/precision guard.

In scope: SC5 and SC6, test-first.
Out of scope: ground-truth changes.

## Acceptance criteria

- AC1 = SC5
- AC2 = SC6

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` + bench + validator before change | pass | 2888806 | 03+04 verified; green at 2888806 |
| AC1 | `go test ./internal/gate -run SCGateLint -v` | pass | 3e5f7f4 | re-verified after N4 residual: CI warn-only preserves status; missing/non-strict/strict paths hold, ABANDON≠ALL MET |
| AC2 | integrated SC1–SC6 + E0 guard + `go test -race ./...` root+bench + validator | pass | 3e5f7f4 | SCL1/2/3 + residuals + SCTokenHonesty + SCGateLint(+Strictness) pass; race root+bench green; validator pass; R82.4/P90.3 unchanged |
| R1 | `go test ./internal/gate ./internal/eval ./internal/tokens ./internal/mcp ./cmd/gira -count=1` | pass | 3e5f7f4 | touched pkgs pass |
| V1 | diff read against AC1–AC2 + spec F5 | pass | 3e5f7f4 | N4 threads status through fail(); CI warn-only / local fail-closed preserved |

## Constraints and budget

- Additive only: new `internal/gate` pkg, `gira gate-lint --strict`, `gira gate --status\|--reverify`, MCP `gate_status`; CI warn-only, local fail-closed.
- No model client, no paid, quarantine ADR 0012, COPY idea in Go, no ground-truth changes, no push/release.

## Plan

1. Failing `TestSCGateLint` first (stale digest must fail, ABANDON must not yield ALL MET).
2. Implement gate contract + lint + status/reverify split.
3. Wire CLI/MCP.
4. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

F5 delivered + F10 sealed + N4 residual: success-only marker+digest, stale-on-drift fail-closed, status/reverify split, ABANDON qualified handoff, CI warn-only with status preserved / local fail-closed on every path incl. missing ledger. Re-verified at 3e5f7f4. Parent integration unblocked.
