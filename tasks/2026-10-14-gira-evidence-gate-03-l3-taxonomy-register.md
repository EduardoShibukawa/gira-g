---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
depends_on: [2026-10-14-gira-evidence-gate-02-l2-control-sample]
verified_at: "a389d21"
---

# L3 taxonomy and findings register (F3)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Every finding carries a Gallon S1 verbatim label (measured/checked/inferred/assumed) mapped to Gira tiers, plus coordinator/reviewer severity. Code-reading without a real-path repro is `inferred` at most. Additive `reviews/findings.md` register (pointer only, outside slices); fix-check re-runs the real path independently.

In scope: SC3, test-first. Failing test first: an unlabeled finding, and a fix-check that reads the fixer's log, must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC3
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` + bench + validator before change | pass | 16e9da0 | 02 verified; green at 16e9da0 |
| AC1 | `go test ./internal/eval -run SCL3TaxonomyRegister -v` | pass | a389d21 | re-verified after F5/F6/F7/F11 fix batch: repro-path os.Stat, register \|\| escape + load-validate, --expect-content real-path fix-check, unknown→unknown |
| AC2 | `go test -race ./...` + bench + validator, no E0 dip | pass | a389d21 | root/bench green, validator pass; R82.4/P90.3 unchanged |
| R1 | `go test ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | a389d21 | touched pkgs pass |
| V1 | diff read against AC1–AC2 + spec F3 | pass | a389d21 | register validated-on-load, fix-check always reads path; pointer-only preserved |

## Constraints and budget

- Additive only: new `internal/eval/findings.go`, `gira review --findings-register --fix-check`, MCP `get_findings_register`; register pointer-only outside slices.
- No model client, no paid, quarantine ADR 0012, COPY idea in Go, no ground-truth changes, no push/release.

## Plan

1. Failing `TestSCL3TaxonomyRegister` first.
2. Implement taxonomy+mapping+severity, register, independent fix-check.
3. Wire CLI/MCP.
4. Run AC1/R1/AC2/V1; commit; validator.

## Outcome

F3 delivered + F5/F6/F7/F11 sealed: Gallon S1 verbatim taxonomy mapped to Gira tiers (unknown→unknown), dual severity, `reviews/findings.md` register round-trip with hostile cells, validated-on-load, independent fix-check via --expect-content real-path read. Re-verified at a389d21; 05 still waits on 04.
