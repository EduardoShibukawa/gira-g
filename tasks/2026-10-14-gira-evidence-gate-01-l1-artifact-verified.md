---
schema: task/1
tier: standard
status: active
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@1
verified_at: "559f408"
---

# L1 artifact-verified closing line (F1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)

## Objective

Every `eval`/`verify` run emits `ARTIFACT-VERIFIED sha256:<hex>` as its final line; the digest covers command/expect/cwd + exit + frozen input fingerprint (1 MiB cap). Writes outside the frozen scope fail closed. Dispatcher 5-point self-check.

In scope: SC1, test-first. Failing test first: a tampered digest must fail.
Out of scope: the other children's features; ground-truth changes.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC6 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | pass | 0a35b64 | root ok (15 pkgs), bench ok (layer0/1/2/quarantine), validator pass |
| AC1 | `go test ./internal/eval -run SCL1ArtifactVerified -v` | pass | 559f408 | failed first (undefined symbols), passes after artifact.go; tampered digest/OOW/substitution all fail as required |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | pass | 559f408 | root 15 pkgs ok, bench ok, validator pass; no extractor/slice change so E0 unaffected |
| R1 | `go test ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | 559f408 | all touched pkgs pass; `gira eval` emits closing line as final line |
| V1 | diff read against AC1–AC2 + spec F1 | pass | 559f408 | additive only: new eval/verify cmds, verify_artifact tool; no existing meaning change |

## Constraints and budget

- Additive only: new file `internal/eval/artifact.go`, new flags/commands, new MCP tool; no existing flag/field meaning change.
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012 (prior-art by hand only, never committed).
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.

## Plan

1. Write failing `TestSCL1ArtifactVerified` first (tampered digest must fail, OOW write must fail, self-check 5 points).
2. Implement `internal/eval/artifact.go` (fingerprint 1 MiB cap, digest, closing line, write-scope, self-check).
3. Wire additive CLI `gira eval --artifact-verified/--no-artifact-verified` + `verify_artifact` MCP tool.
4. Run AC1, R1, AC2, V1; commit; validator.

## Outcome

F1 delivered: `internal/eval/artifact.go` (fingerprint 1 MiB cap, digest, closing line, write-scope fail-closed, 5-point self-check), `gira eval/verify` additive CLI, `verify_artifact` MCP tool. No follow-ups; 02 unblocked.
