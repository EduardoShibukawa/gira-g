---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@2
depends_on: [2026-10-14-gira-evidence-gate-02-l2-control-sample]
verified_at: "a389d21"
---

# Honest token accounting (F4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)
- spec-version: v1 → v2; unaffected: policy definition only (Decisions/Open questions, no code, gate, or criteria change)

## Objective

Token report shows `{gross, expansions, net, measured_bool}` with `net = gross − expansions`. Estimated-only claims are quarantined with an `estimated` label and never enter headline numbers. No `len/4`. Escape-tax lift before measuring. Measurement only — envelopes stay E3's scope.

In scope: SC4, test-first. Failing test first: a `len/4` estimate passing as measured must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC4
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` + bench + validator before change | pass | 0aa3c1e | 03 verified; green at 0aa3c1e |
| AC1 | `go test ./internal/tokens -run SCTokenHonesty -v` | pass | a389d21 | re-verified after F8 fix batch: caller expansions force measured=false, net=gross-expansions, quarantine holds |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | a389d21 | root/bench green, validator pass; R82.4/P90.3 unchanged |
| R1 | `go test ./internal/tokens ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | a389d21 | touched pkgs pass; `gira eval --token-report=measured` emits measured tokens line |
| V1 | diff read against AC1–AC2 + spec F4 | pass | a389d21 | mixed reports quarantined-not-rejected; pure measurements stay measured |

## Constraints and budget

- Additive only: new `internal/tokens/accounting.go`, extend `gira eval --token-report`, MCP `get_token_accounting`; existing `Estimate` untouched (estimated, quarantined).
- No model client, no paid, quarantine ADR 0012, COPY idea in Go, no ground-truth changes, no push/release. Measurement only, no envelopes.

## Plan

1. Failing `TestSCTokenHonesty` first (len/4 as measured must fail).
2. Implement report+net, quarantine, lift-before-measure, single-source defs.
3. Wire CLI/MCP.
4. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

F4 delivered + F8 sealed: `{gross,expansions,net,measured_bool}`, quarantine keeps estimated out of headlines, lift-before-measure, single-source defs; MCP mixed reports labeled unmeasured. Re-verified at a389d21; 05 unblocked (03+04 now verified).
