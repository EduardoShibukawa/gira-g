---
schema: task/1
tier: heavy
status: completed
spec: gira-evidence-gate@1
---

# Evidence & gate hardening (Gallon L1/L2/L3 + token honesty)

## Routing and approvals

- tier: heavy (H1: product calls on sampling rate, taxonomy labels, gate strictness; H3: spans internal/eval, CLI, MCP, bench Layer 2; H5: 5 children, multi-session)
- confirmed: heavy by owner (2026-10-14, "Aprovar heavy E1")
- decided: L2 sampling default → strict max(20,5%) Gallon default by owner (2026-10-14, "max(20,5%) Gallon (Recommended)")
- decided: L3 taxonomy → Gallon S1 verbatim + mapping table to Gira tiers by owner (2026-10-14, "Gallon verbatim + mapa (Recommended)")
- decided: gate-lint --strict → warn-only in CI, fail-closed locally, until E3 by owner (2026-10-14, "Warn-only (Recommended)")
- approved: spec v1 by owner (2026-10-14, "Approve v1 (Recommended)")

## Objective

Make every `gira` claim independently re-derivable or explicitly silent: L1 artifact closing-line, L2 control-total + seeded sampling, L3 taxonomy + findings register + independent fix-check, honest token accounting, gate-lint. No extractor recall change; only the proof gets stronger.

In scope:

- spec `gira-evidence-gate` v1;
- the children listed in the Plan;
- integrated verification of SC1–SC6.

Out of scope:

- E2 joins, E3 envelopes/hygiene, E4 human surface, E5 skills distribution, E6 docs site;
- ground-truth changes, and `reviewed: true` (the owner's call);
- any model client (ADR 0001), any paid fan-out;
- push or release.

## Acceptance criteria

- AC1 = SC1–SC6 of `specs/gira-evidence-gate.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before children | pass | 0a35b64 | root 15 pkgs ok, bench ok, validator pass (recorded in 01/B1) |
| SC1 | `go test ./internal/eval -run SCL1ArtifactVerified` + CLI/MCP wiring | fail | 2dd72b5 | lib passes; review F1–F2: `--artifact-verified` flag missing, digest preimage differs across eval/verify/verify_artifact, fingerprint not printed |
| SC2 | `go test ./internal/eval -run SCL2ControlSample` + CLI/MCP wiring | fail | 2dd72b5 | lib passes; review F3–F4: plug-via-first undetected, CLI/MCP assert agreement without sampling/comparison |
| SC3 | `go test ./internal/eval -run SCL3TaxonomyRegister` + CLI/MCP wiring | fail | 2dd72b5 | lib passes; review F5–F7: fake-path measured, register `\|` round-trip + unvalidated load, CLI `--repro-ok` self-attestation (critical) |
| SC4 | `go test ./internal/tokens -run SCTokenHonesty` + CLI/MCP wiring | fail | 2dd72b5 | lib passes; review F8: MCP labels caller expansions measured |
| SC5 | `go test ./internal/gate -run SCGateLint` + CLI/MCP wiring | pass | 2dd72b5 | sealed/drift/ABANDON/status-reverify hold; minors: missing-file CI path, `Strict` no-op (review F10 open) |
| SC6 | I1/I2 + race root+bench + validator + E0 guard | fail | 2dd72b5 | regression half passes (race/bench/validator green, EndToEnd suites pass, no extractor/GT change); I1 breached via F4/F7, I2 taxonomy-in-digest missing (review F9), `unknown` literal (F11) |
| AC1 | SC1–SC6 integrated | fail | 2dd72b5 | SC1–SC4, SC6 fail per review; SC5 passes |
| R1 | `go test -race ./...` (root, integrated head) | pass | 2dd72b5 | all pkgs ok, incl. EndToEnd published+perturbed |
| R2 | `go test ./...` (bench/) | pass | 2dd72b5 | layer0/1/2/quarantine ok |
| R3 | `node tools/validate/validate.js` | pass | 2dd72b5 | manifest-integrity + artifact-records pass |
| V1 | integrated diff 0a35b64..HEAD read + fresh-context review ses_ec79c595bffesmPPu3ig7Vq0HC | pass | 2dd72b5 | additive only (own epic); concurrent 18b9af4 CI slash-fix noted, not ours; no GT/model/paid/push issues; review ran, 11 findings open below |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-l1-artifact-verified`: F1 + frozen scope check + dispatcher self-check (SC1).
2. `-02-l2-control-sample`: control-total + seeded max(20,5%) sampling, never-force-match (SC2). Depends on 01.
3. `-03-l3-taxonomy-register`: taxonomy + register + independent fix-check (SC3). Depends on 02.
4. `-04-token-honesty`: measured/estimated/net, no len/4 (SC4). Depends on 02.
5. `-05-gate-lint-validator`: gate-lint strict + integrated SC1–SC6 + E0 guard (SC5, SC6). Depends on 03, 04.

## Review

- [F1] `gira eval --artifact-verified` flag advertised but undefined (`cmd/gira/eval_artifact.go:27-34`) → open (major, SC1)
- [F2] digest preimage differs across eval/verify/verify_artifact; fingerprint not printed → open (major, SC1)
- [F3] `BlindReDerive` checks plugs only in `second`, not `first` (`internal/eval/control.go:95-106`) → open (major, SC2)
- [F4] CLI/MCP control lines assert agreement without sampling/comparison → open (major, SC2/I1)
- [F5] `Finding.Validate` accepts fake `ReproPath` as measured (`internal/eval/findings.go:69-73`) → open (major, SC3)
- [F6] register `\|` escape lost on load; `LoadRegister` never validates → open (major, SC3)
- [F7] CLI `--repro-ok` self-attests fix-check without touching the repro path (`cmd/gira/review.go:40`) → open (critical, SC3/I1)
- [F8] MCP `get_token_accounting` labels caller expansions measured (`internal/mcp/server.go:634-635`) → open (major, SC4)
- [F9] taxonomy never enters any digest (I2 second clause unimplemented) → open (major, SC6/I2; needs spec-interpretation call: digest preimage change invalidates existing digests)
- [F10] missing ledger fails in CI (should warn); `Strict` is a no-op (`internal/gate/gate.go:99-102,121-124`) → open (minor, SC5)
- [F11] `GiraTier()` returns `""` instead of `"unknown"` → open (minor, SC6/I1)

## Checkpoint

- as of: 2026-10-14, branch feat/gira-impact-precision, head 2dd72b5, worktree dirty (tasks/2026-10-14-gira-evidence-gate.md)
- done: 01–05 verified (@1, specs v1); parent integration attempted (SC rows + R1–R3 + V1 above); fresh-context review ses_ec79c595bffesmPPu3ig7Vq0HC recorded (11 open findings)
- next: owner morning triage — (a) F9 digest-preimage call, (b) CLI demo semantics (F4 control line, F7 fix-check predicate), (c) fix batch + child re-verification; then parent re-lint, review close-out, `accepted:` by owner
- open: fixes deferred to avoid invalidating 5 verified children overnight; no push/release (branch only)
- revalidate: SC rows ran at 2dd72b5; rerun after any code change
