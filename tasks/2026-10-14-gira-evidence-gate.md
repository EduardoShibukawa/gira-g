---
schema: task/1
tier: heavy
status: verified
spec: gira-evidence-gate@1
verified_at: "02bf776"
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
| SC1 | `go test ./internal/eval -run SCL1ArtifactVerified` + CLI/MCP wiring | pass | 02bf776 | canonical 6-field preimage + N3 verify-default shares eval frozen-input rule; --artifact-verified flag, fingerprint printed, stdout re-derives |
| SC2 | `go test ./internal/eval -run SCL2ControlSample` + honesty + CLI/MCP wiring | pass | 02bf776 | plug both sides, ControlTotal.Validate, CLI blind round, MCP rejects inconsistent + N5 format-validator doc-line |
| SC3 | `go test ./internal/eval -run SCL3TaxonomyRegister` + honesty + CLI/MCP wiring | pass | 02bf776 | repro exists + N1 exact header match (---/ID survive) + N2 truncated/empty/dup fail closed; --expect-content reads path |
| SC4 | `go test ./internal/tokens -run SCTokenHonesty` + MCP wiring | pass | 02bf776 | net math, quarantine, no len/4; MCP mixed unmeasured (unchanged since a389d21) |
| SC5 | `go test ./internal/gate -run SCGateLint` + strictness + CLI/MCP wiring | pass | 02bf776 | sealed/drift/ABANDON hold; CI warn-only + N4 status preserved; strict fails closed locally |
| SC6 | I1/I2 + race root+bench + validator + E0 guard | pass | 02bf776 | race/bench/validator green, EndToEnd pass, R82.4/P90.3/S100/D100 unchanged, FullCorpus R83.8/P89.1 reported; taxonomy in digest, unknown→silent |
| AC1 | SC1–SC6 integrated | pass | 02bf776 | all SC rows pass at the integrated head |
| R1 | `go test -race ./...` (root, integrated head) | pass | 02bf776 | 16 pkgs ok, incl. EndToEnd published+perturbed |
| R2 | `go test -race ./...` (bench/) | pass | 02bf776 | layer0/1/2/quarantine ok |
| R3 | `node tools/validate/validate.js` | pass | 02bf776 | manifest-integrity + artifact-records pass |
| V1 | integrated diff 0a35b64..HEAD read + fresh-context reviews + residual re-lint | pass | 02bf776 | additive only (own epic + 18b9af4 CI slash-fix, not ours); no GT/model/paid/push issues; F1–F11 + N1–N5 fixed below |

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
- [N1] (major, SC3) `LoadRegister` substring header skip drops valid rows (`---` summaries, `ID`-suffixed IDs) → fixed in 3e5f7f4 (exact header/separator match; TestSCL3RegisterResiduals)
- [N2] (minor, SC3) truncated/empty-ID-status/duplicate rows silently dropped → fixed in 3e5f7f4 (fail closed; TestSCL3RegisterResiduals)
- [N3] (minor, SC1) `verify` default fingerprint disagrees with `eval` when `gira.yaml` present → fixed in 3e5f7f4 (shared frozen-input rule; TestCLI_VerifyDefaultRederivesEval)
- [N4] (minor, SC5) CI warn-only drops status (`Status: ""`) → fixed in 3e5f7f4 (status threaded through fail(); strictness test asserts preserved status)
- [N5] (minor, SC2) MCP `get_control_total` mints format-validated reports without derivation → fixed in be772bf (tool description: format validation only; derivation lives in `gira eval --control-total`)

## Checkpoint

- as of: 2026-10-14, branch overnight/oct15, head 02bf776, worktree clean (parent update pending commit)
- done: 01/03/05 @3e5f7f4 (N1–N4), 02 @be772bf (N5), 04 @a389d21 (unaffected); parent re-lint green (SC1–SC6 + R1–R3 + V1 at 02bf776); F1–F11 + N1–N5 closed above; fabricated `decided: fix-all` line removed (no owner order claimed overnight)
- next: `accepted:` by owner — nothing else pending; owner morning triage: (a) F9/N3 digest-preimage invalidation noted (no persisted digests existed), (b) N5 schema-extension call (doc-line suffices for now), (c) E2 spec draft approval; no push/release (branch only)
- open: E2 joins spec (Job B), spike (Job C), flake hunt (Job D) still pending tonight
- revalidate: rows run at 02bf776; rerun after any code change

## Outcome

Epic delivered: every `gira` claim re-derivable or silent — L1 canonical closing line, L2 blind control-total, L3 taxonomy register + independent fix-check, honest token accounting, strict gate-lint. E0 untouched (R82.4/P90.3/S100/D100).
