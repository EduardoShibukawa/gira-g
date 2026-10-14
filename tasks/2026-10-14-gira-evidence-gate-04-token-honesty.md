---
schema: task/1
tier: standard
status: active
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-02-l2-control-sample]
verified_at: "f48e56b"
---

# Honest token accounting (F4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

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
| AC1 | `go test ./internal/tokens -run SCTokenHonesty -v` | pass | f48e56b | failed first (undefined symbols), passes after accounting.go; len/4-as-measured fails, quarantine/net/lift hold |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | f48e56b | root/bench green, validator pass; Estimate untouched so E0 unaffected |
| R1 | `go test ./internal/tokens ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | f48e56b | touched pkgs pass; `gira eval --token-report=measured` emits measured tokens line |
| V1 | diff read against AC1–AC2 + spec F4 | pass | f48e56b | additive only: accounting.go, eval flag, get_token_accounting; no envelopes, no existing change |

## Constraints and budget

- Additive only: new `internal/tokens/accounting.go`, extend `gira eval --token-report`, MCP `get_token_accounting`; existing `Estimate` untouched (estimated, quarantined).
- No model client, no paid, quarantine ADR 0012, COPY idea in Go, no ground-truth changes, no push/release. Measurement only, no envelopes.

## Plan

1. Failing `TestSCTokenHonesty` first (len/4 as measured must fail).
2. Implement report+net, quarantine, lift-before-measure, single-source defs.
3. Wire CLI/MCP.
4. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

F4 delivered: `{gross,expansions,net,measured_bool}`, quarantine keeps estimated out of headlines, lift-before-measure, single-source defs. 05 unblocked (03+04 now verified).
