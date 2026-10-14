---
schema: task/1
tier: heavy
status: verified
spec: gira-evidence-gate@1
verified_at: "0c5f8b6"
---

# Evidence & gate hardening (Gallon L1/L2/L3 + token honesty)

## Routing and approvals

- tier: heavy (H1: product calls on sampling rate, taxonomy labels, gate strictness; H3: spans internal/eval, CLI, MCP, bench Layer 2; H5: 5 children, multi-session)
- confirmed: heavy by owner (2026-10-14, "Aprovar heavy E1")
- decided: L2 sampling default → strict max(20,5%) Gallon default by owner (2026-10-14, "max(20,5%) Gallon (Recommended)")
- decided: L3 taxonomy → Gallon S1 verbatim + mapping table to Gira tiers by owner (2026-10-14, "Gallon verbatim + mapa (Recommended)")
- decided: gate-lint --strict → warn-only in CI, fail-closed locally, until E3 by owner (2026-10-14, "Warn-only (Recommended)")
- approved: spec v1 by owner (2026-10-14, "Approve v1 (Recommended)")
- decided: fix-all → owner ordered the full F1–F11 fix batch, autoresearch-grounded, by owner (2026-10-14, "faca o fix de tudo")

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
| SC1 | `go test ./internal/eval -run SCL1ArtifactVerified` + CLI/MCP wiring | pass | 0c5f8b6 | canonical 6-field preimage (taxonomy in digest), `--artifact-verified` flag, fingerprint printed, stdout re-derives (TestCLI_EvalStdoutRederives) |
| SC2 | `go test ./internal/eval -run SCL2ControlSample` + honesty + CLI/MCP wiring | pass | 0c5f8b6 | plug rejected both sides, ControlTotal.Validate, CLI runs real seeded sample + blind round, MCP rejects inconsistent lines |
| SC3 | `go test ./internal/eval -run SCL3TaxonomyRegister` + honesty + CLI/MCP wiring | pass | 0c5f8b6 | repro path must exist, register round-trips hostile cells + validates on load, `--expect-content` fix-check reads the path |
| SC4 | `go test ./internal/tokens -run SCTokenHonesty` + MCP wiring | pass | 0c5f8b6 | net math, quarantine, no len/4 as measured; MCP mixed reports labeled unmeasured |
| SC5 | `go test ./internal/gate -run SCGateLint` + strictness + CLI/MCP wiring | pass | 0c5f8b6 | sealed/drift/ABANDON/status-reverify hold; CI warn-only incl. missing ledger; non-strict warns, strict fails closed locally |
| SC6 | I1/I2 + race root+bench + validator + E0 guard | pass | 0c5f8b6 | race/bench/validator green, EndToEnd published+perturbed pass, R82.4/P90.3/S100/D100 unchanged, no extractor/GT change; taxonomy in digest (I2), unknown→silent (I1) |
| AC1 | SC1–SC6 integrated | pass | 0c5f8b6 | all SC rows pass at the integrated head |
| R1 | `go test -race ./...` (root, integrated head) | pass | 0c5f8b6 | all pkgs ok, incl. EndToEnd published+perturbed |
| R2 | `go test ./...` (bench/) | pass | 0c5f8b6 | layer0/1/2/quarantine ok |
| R3 | `node tools/validate/validate.js` | pass | 0c5f8b6 | manifest-integrity + artifact-records pass |
| V1 | integrated diff 0a35b64..HEAD read + fresh-context review ses_ec79c595bffesmPPu3ig7Vq0HC + fix re-lint | pass | 0c5f8b6 | additive only (own epic + 18b9af4 CI slash-fix, not ours); no GT/model/paid/push issues; 11 findings fixed below |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-l1-artifact-verified`: F1 + frozen scope check + dispatcher self-check (SC1).
2. `-02-l2-control-sample`: control-total + seeded max(20,5%) sampling, never-force-match (SC2). Depends on 01.
3. `-03-l3-taxonomy-register`: taxonomy + register + independent fix-check (SC3). Depends on 02.
4. `-04-token-honesty`: measured/estimated/net, no len/4 (SC4). Depends on 02.
5. `-05-gate-lint-validator`: gate-lint strict + integrated SC1–SC6 + E0 guard (SC5, SC6). Depends on 03, 04.

## Review

- [F1] `gira eval --artifact-verified` flag advertised but undefined → fixed in a389d21 (explicit flag, default on; TestCLI_EvalStdoutRederives)
- [F2] digest preimage differs across entry points; fingerprint not printed → fixed in a389d21 (canonical 6-field constructor shared by eval/verify/verify_artifact; fingerprint line; stdout re-derives)
- [F3] `BlindReDerive` plug check only in `second` → fixed in a389d21 (both sides; TestSCL2ControlHonesty)
- [F4] control lines assert agreement without a round → fixed in a389d21 (CLI runs SeededSample+BlindReDerive+Validate; MCP rejects inconsistent lines)
- [F5] `Finding.Validate` accepts fake `ReproPath` → fixed in a389d21 (os.Stat existence for measured/checked)
- [F6] register `\|` escape lost; load never validates → fixed in a389d21 (escape-aware split/unescape; load validates every row)
- [F7] CLI `--repro-ok` self-attestation → fixed in a389d21 (`--expect-content`; rerun always reads the path; TestCLI_ReviewFixCheckReadsPath)
- [F8] MCP labels caller expansions measured → fixed in a389d21 (mixed reports unmeasured; TestMCP_TokenAccountingHonesty)
- [F9] taxonomy not in any digest → fixed in a389d21 (TAXONOMY line in preimage = spec I2; no persisted digests existed, so nothing invalidated in practice)
- [F10] missing ledger fails in CI; `Strict` no-op → fixed in a389d21 (CI warn-only everywhere; non-strict warns, strict fails closed locally; TestSCGateLintStrictness)
- [F11] `GiraTier()` returns `""` → fixed in a389d21 (`"unknown"`; TestSCL3RegisterHonesty)

## Checkpoint

- as of: 2026-10-14, branch overnight/oct15, head 0c5f8b6, worktree dirty (tasks/2026-10-14-gira-evidence-gate.md)
- done: 01–05 verified @a389d21; parent re-lint green (SC1–SC6 + R1–R3 + V1 at 0c5f8b6); review findings F1–F11 closed above
- next: `accepted:` by owner — nothing else pending; no push/release (branch only)
- revalidate: rows ran at 0c5f8b6; rerun after any code change

## Outcome

Epic delivered: every `gira` claim re-derivable or silent — L1 canonical closing line, L2 blind control-total, L3 taxonomy register + independent fix-check, honest token accounting, strict gate-lint. E0 untouched (R82.4/P90.3/S100/D100).
