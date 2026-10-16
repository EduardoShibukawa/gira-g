---
schema: spec/1
status: approved
version: 2
approved_version: 2
approved_hash: sha256:931ba41b6c8aaae9c99045419fbff21fac70d91a8538a085db39c199b07d1ea3
---

# GIRA Evidence & Gate Hardening

## Problem

E0 sealed impact-precision at recall 82.4% / precision 90.3% (parent `2026-10-13-gira-impact-precision`, verified at `99bb529`), but the numbers are still *claims*, not independently re-derivable *evidence*. Four gaps, mapped from `docs/prior-art/agent-review.md`, `token-optimizer.md` and `unlazy.md`:

1. **No artifact closing-line (Gallon L1).** Eval/verify runs don't emit a frozen digest, so a "pass" can't be re-checked against its inputs. G-Ingá has Verification tables and `verified_at` pins, but no dispatch contract (named file + byte floor + frozen check outside worker write-scope + hash-before/after) and no mandatory closing line.
2. **No control-total / sampling (Gallon L2).** Ground truth was blind re-derived once (3 subagents), but there is no standing control-total doctrine, no frozen A/B digests before reconciliation, and no seeded agreement-sampling rule. Divergences have no never-force-match / never-plug-drop invariant.
3. **No claim taxonomy / register (Gallon L3).** Fresh-context review (`review.md`) exists, but findings carry no measured/checked/inferred/assumed labels, no `reviews/findings.md` register, no independent fix-checker, no fresh-eyes-on-final round. Code-reading can pass as verified.
4. **Token accounting mixes measured with estimated.** Harness tokens use the budget estimator; `len/4`-class heuristics are banned nowhere by test. Per `token-optimizer.md`, `len/4` errs +130.4% on whitespace-heavy source and −75% on base64. Without `net = gross − expansions` and quarantine of modeled claims, E3's hard envelopes would encode lies.

Heavy triggers:

- **H1:** product calls on sampling rate, taxonomy labels, gate strictness, enforcement point.
- **H3:** spans `internal/eval`, CLI, MCP surface, `tools/validate`, bench Layer 2.
- **H5:** multi-session, 5 children.

## Expected behavior

One epic, five additive features. COPY = idea only, reimplemented in Go; never paste TS/JS.

### F1 — L1 artifact line

Every `eval`/`verify` run emits `ARTIFACT-VERIFIED sha256:<hex>` as its final line. The digest covers `CHECK:` (command) / `EXPECT:` (success marker) / `CWD:` + exit code + frozen input fingerprint (1 MiB cap, unlazy contract). Writes outside the frozen scope fail closed. Dispatcher self-check (5 points: existence, harness digest, own run, hash recompute, read-for-substitute).

- CLI: `gira eval --artifact-verified` (default on; `--no-artifact-verified` debug-only, additive). MCP: `verify_artifact` → `{digest, closing_line}`.
- Budget: +0 tokens to slices.

### F2 — L2 control-total + sampling

Each run emits `control:{total, matched, mismatched, sampled_n, seed}`. Seeded sampling rule `max(20, 5%)` (Gallon default, owner decision 2026-10-14). Blind re-derivation on the sample must match 100% or report a diff — never plug, never drop, never force-match. Frozen digests compared before reconciliation.

- CLI: `gira eval --control-total --sample max20-5pct --seed <u64>`. MCP: `get_control_total`.

### F3 — L3 taxonomy + register

Every finding carries a taxonomy label (`measured/checked/inferred/assumed`, Gallon S1 verbatim) mapped to Gira tiers (`declared/extracted/inferred`), plus coordinator/reviewer severity. Code-reading without a real-path repro is `inferred` at most. `reviews/findings.md` register is ID-paired with status; fix-check re-runs the real path independently (never reads the fixer's log).

- CLI: `gira review --findings-register reviews/findings.md --fix-check`. MCP: `get_findings_register`. Register lives outside token-budgeted slices (pointer only).

### F4 — Token honesty

Token report shows `{gross, expansions, net, measured_bool}` with `net = gross − expansions` (may be negative). Estimated-only claims are quarantined with an `estimated` label and never enter headline numbers. No `len/4`. `liftTextPart`-equivalent (escape-tax lift) runs before measuring. Single-source tool defs for counting.

- CLI: `gira eval --token-report=measured`. MCP: `get_token_accounting`. Measurement only — envelope enforcement stays E3's scope.

### F5 — gate-lint

Port of the unlazy gate contract (not its JS): success-only marker + digest, stale-on-drift fail-closed, `--status` (non-executing) / `--reverify` split, `ABANDON` as qualified handoff that never reports `ALL MET`. CI mode is warn-only until E3 (owner decision 2026-10-14); local runs fail closed.

- CLI: `gira gate-lint --strict`, `gira gate --status | --reverify`. MCP: `gate_status`.

### Invariants (all features)

- **I1 — Evidence or silence.** No claim without a locator-backed `EvidenceRecord`; unresolved → `unknown`, never named.
- **I2 — Tier honesty.** `declared/extracted/inferred` never mislabeled; L3 taxonomy part of the digest.

## Scope

**In scope**

- F1–F5 above, each exposed on CLI and MCP, all additive (no flag/field meaning changes; budgets/truncation untouched).
- Seeded sampler + control-total in the deterministic scorer config (not model calls).
- Warn-only CI wiring for `gate-lint` (strict enforced locally).

**Out of scope**

- E2 joins, E3 envelopes/hygiene, E4 human surface, E5 skills distribution, E6 docs site.
- Ground-truth changes, `reviewed: true`, push/release.
- Any model client (ADR 0001), any paid multi-model fan-out, any semantic harvest.
- Extracting business rules.

**Non-goals**

- Raising recall/precision (E2's job) or cutting tokens (E3's job). This epic only strengthens the proof.

## Acceptance criteria

- **SC1 (L1).** Every `eval`/`verify` run emits `ARTIFACT-VERIFIED sha256:<hex>` as its final line; the digest covers command/expect/cwd + exit + frozen input fingerprint; out-of-write-scope writes fail closed. Test: `go test ./internal/eval -run SCL1ArtifactVerified`.
- **SC2 (L2).** Control-total emitted per run; blind re-derivation on the seeded `max(20,5%)` sample matches 100% or reports an unplugged diff; frozen digests compared. Test: `go test ./internal/eval -run SCL2ControlSample`.
- **SC3 (L3).** Every finding carries taxonomy + severity; the register round-trips; fix-check re-runs the real path independently. Test: `go test ./internal/eval -run SCL3TaxonomyRegister`.
- **SC4 (tokens).** Token report shows `measured/estimated/net` with `net=gross−expansions`; estimated-only claims quarantined; no `len/4`. Test: `go test ./internal/tokens -run SCTokenHonesty`.
- **SC5 (gate).** `gate-lint --strict` passes on sealed runs and fails closed on stale digest/drift; `ABANDON` never yields `ALL MET`. Test: `go test ./internal/gate -run SCGateLint`.
- **SC6 (invariants + regression).** I1/I2 hold; `go test -race ./...` passes at root and in `bench/`; validator reports no `fail`; determinism 100%, safety 100%; E0 recall/precision not lowered (any dip needs a `divergence:` owner entry).

## Interfaces and dependencies

All additive. No existing flag, field or tool changes meaning.

| Interface | Change |
|---|---|
| CLI `eval` | `--artifact-verified` (default on), `--no-artifact-verified` (debug), `--control-total`, `--sample`, `--seed`, `--token-report` (new flags) |
| CLI | `review --findings-register --fix-check`, `gate-lint --strict`, `gate --status \| --reverify` (new) |
| MCP | `verify_artifact`, `get_control_total`, `get_findings_register`, `get_token_accounting`, `gate_status` (new tools) |
| `reviews/findings.md` | new additive register file (pointer only, outside slices) |

**ADRs.** ADR 0001 stands (no model client). ADR 0012 stands (quarantine; prior-art repos run by hand in throwaway envs only, output never committed). L3 personas are fixed templates, not models.

## Decisions

- `confirmed` heavy by owner (2026-10-14, question round "E1 heavy?" → "Aprovar heavy E1").
- `decided` sampling default → strict `max(20,5%)` Gallon default by owner (2026-10-14).
- `decided` taxonomy → Gallon S1 verbatim + mapping table to Gira tiers by owner (2026-10-14).
- `decided` `gate-lint --strict` → warn-only in CI, fail-closed locally, until E3 by owner (2026-10-14).
- `decided` strict-mode promotion point → warn-only remains the default CI posture; promotion to fail-closed happens only via the explicit `--fail-on-ci` switch (skills-release 04) after a clean `gira gate-lint --strict` dry-run reports zero stale digests on `main` and `SCGateLint`/`SCGateLintStrictness` pass. Checks must land before the switch, never in the same commit.
- `decided` register retention/GC policy → registers are append-only with indefinite pointer-only retention by default; compaction is always explicit (`gira review --gc [--keep-resolved <n>]`, default 0), preserving every `open`/`inferred`/`assumed` finding and trimming only `fixed`/`closed`/`verified` to the last N, keeping ID+status tombstones and a pre-compaction digest. No time-based or automatic purge.
- **COPY = idea, never literal code.** Reimplement in Go; verbatim snippets only as rare exceptions with license header + `THIRD_PARTY_NOTICES`.

## Risks and constraints

- **Gate too strict.** Mitigation: warn-only CI; thresholds tunable without interface change.
- **Cost.** Sampling is bounded (never full re-derivation by default); L3 rounds are template-driven, no paid fan-out (protected action needs its own approval).
- **Non-determinism.** Personas are fixed templates; scorer/checkers stay the arbiters; reviewer verdicts advisory with coordinator-owned severity.
- **Compatibility.** All CLI/MCP changes additive; register files additive; `git revert` per child suffices.
- **Honest No-Go.** A stricter gate may flip sealed runs to fail-closed on stale digests. Expected and acceptable: fix the digest, not the gate.

## Verification

- Per child + integrated: B1 `go test -race ./...` (root); B2 `go test -race ./...` (bench/); B3 `go test ./internal/eval -run EndToEnd -v` (both hermetic variants, report recall/precision/safety/determinism); B4 opt-in full-corpus run (reported only); B5 `node tools/validate/validate.js`.
- AC rows map SC1–SC6 to the `go test` names above; `not run` forbidden at parent seal.
- **Review focus:** does the closing-line digest actually cover inputs (not just exit 0)? Is sampling seeded + blind (no ground-truth leak)? Is taxonomy uninflated (measured requires real-path repro)? Is `net` math over both SHA-256 payloads? Any weakened assertions or named guesses?

## Rollback

`git revert` per child. No migration, publication or external state. Register files are additive; deleting them restores pre-gate behavior.

## Open questions

- resolved: **exact CI promotion point** for strict mode (decision above + envelopes spec v3 Q8): warn-only stays the default; promotion runs through the explicit `--fail-on-ci` switch and requires a clean strict dry-run with zero stale digests plus green `SCGateLint` before it is flipped.
- resolved: **register retention/GC policy** (decision above + envelopes spec v3 Q9): append-only indefinite retention by default; compaction is explicit, `--gc`-driven, open/inferred/assumed findings are never dropped, and compaction keeps tombstones plus a pre-compaction digest.

## Revisions

- v1 (2026-10-14): first draft.
- v2 (2026-10-16): closes the two deferred questions of v1 via the decided promotion point and register retention/GC policy, drawing on the mechanisms shipped in gira-skills-release-04 (FailOnCI, --fail-on-ci, Register.GC, --gc). No code, gate, or criteria changes; Decisions and Open questions updated only. Approved by owner on 2026-10-16 ("Confirm heavy ...").
