---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-evidence-gate
spec: gira-evidence-gate@2
verified_at: "3e5f7f4"
---

# L1 artifact-verified closing line (F1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-evidence-gate; spec v1 approved)
- spec-version: v1 → v2; unaffected: policy definition only (Decisions/Open questions, no code, gate, or criteria change)

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
| AC1 | `go test ./internal/eval -run SCL1ArtifactVerified -v` | pass | 3e5f7f4 | re-verified after N3 residual: verify-default shares eval frozen-input rule; canonical preimage, fingerprint, --artifact-verified hold |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | pass | 3e5f7f4 | root 16 pkgs ok, bench ok, validator pass; R82.4/P90.3/S100/D100 unchanged |
| R1 | `go test ./internal/eval ./internal/mcp ./cmd/gira -count=1` | pass | 3e5f7f4 | all touched pkgs pass; verify-default re-derives eval closing line |
| V1 | diff read against AC1–AC2 + spec F1 | pass | 3e5f7f4 | N3 one-spot fix (gira.yaml in verify-default); no other meaning change |

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

F1 delivered + F1/F2/F9 sealed + N3 residual: `internal/eval/artifact.go` (canonical 6-field digest with taxonomy, fingerprint 1 MiB cap, closing line, write-scope fail-closed, 5-point self-check), `gira eval/verify` additive CLI (`--artifact-verified`, `--taxonomy`, fingerprint line; verify-default shares eval frozen-input rule), `verify_artifact` MCP tool (taxonomy param). Re-verified at 3e5f7f4; 02 unblocked.
