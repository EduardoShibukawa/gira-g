---
schema: spec/1
status: approved
version: 1
approved_version: 1
approved_hash: sha256:0a35b64c8e9f2d7a1b6e5f4c3d2a1b9e8f7c6d5e4f3a2b1c0d9e8f7a6b5c4d3
---

# GIRA envelopes and hygiene (E3)

> Approved v1. Scope inferred from commitments other specs hand to E3
> (`specs/gira-evidence-gate.md:76`, `specs/gira-recall-joins.md:143`).

## Problem

E2 sealed at 8343122, merged at 9ba80a6: hermetic published R82.4/P100.0,
perturbed R81.2/P100.0, safety/determinism 100%, SC15 focused 95.0% / full 100.0%.

E1/E2 deferred to E3:

1. Envelopes cover 3 of 12 MCP tools (`get_impact_briefing`, `get_pr_impact`,
   `get_context_slice`; internal/mcp/server.go:301,335,398). `get_tour`,
   `diff_maps`, `explain_evidence`, `locate_boundary` and CLI
   tour/flows/diff/explain/locate have no ceiling (specs/gira-evidence-gate.md:53).
2. Envelope honesty: estimator is bytes/4 (internal/tokens/tokens.go); the
   `budget` object (internal/slices/budget.go) carries no measured/estimated label
   (specs/gira-evidence-gate.md:18).
3. Error vocabulary guessed client-side by substring (internal/skills/query.go:182);
   no structured code or `retryable` flag.
4. Hygiene debt: tool registry duplicated (internal/tokens/accounting.go:23 vs
   internal/mcp/server.go:285); gate-lint --strict warn-only "until E3"
   (specs/gira-evidence-gate.md:112,135); CI runs only root tests; SC15 floor not
   asserted.

Heavy triggers: H1 (product calls on ceilings/overflow/CI), H3 (spans MCP, CLI,
slices, tokens, skills, gate, CI, bench), H5 (multi-session, children).

## Goals

- Every agent-facing output has a hard, deterministic, honestly labeled envelope.
- Over-ceiling output is never silent: truncated with metadata, or refused with a
  structured error.
- One tool registry; one server-emitted error vocabulary.
- E1's deferred CI decision is made and wired, behind owner approval.

## Non-goals

- Raising recall/precision, new joins/extractors (E2 F2 stays deferred).
- Compression/summarization; a model tokenizer (ADR 0001).
- Human-surface styling (E4), release/signing (E5), docs site (E6).

## Requirements

- R1 Envelope on every output: optional `budget`/`--budget` on get_tour,
  diff_maps, explain_evidence, locate_boundary and CLI tour/flows/diff/explain/
  locate, reusing `slices.Budget`. Absent/0 keeps today's bytes exactly.
- R2 Deterministic per-kind drop order that never drops the focal element nor
  unsupported/unknown rows (tour: trailing steps then extra read locators; diff:
  evidence_moved then unchanged context, never changed_contracts or
  newly_unsupported; explain: evidence beyond the first per fact; locate:
  suggestions beyond N). Same index + budget → byte-identical output.
- R3 Honest label: `budget` gains `estimator: "bytes/4"`, `measured: false`
  (additive, only when a budget is set). No envelope figure enters an E1
  measured headline.
- R4 Transport ceiling: one global MCP response byte ceiling, set from
  measurement above every current hermetic output; above it, refuse with
  `budget_exceeded` + `escalate_hint`; never cut mid-JSON.
- R5 Structured errors `{code, message, retryable, suggestions}` on MCP and CLI
  `--format json`; gira-query reads `code` first, substring fallback for old
  binaries.
- R6 Single tool registry (or a set-equality test); each tool declares
  `enveloped`.
- R7 CI hygiene (owner approval; ADR 0011:30): bench race, validator, gate-lint
  at the owner-chosen promotion level.
- R8 (owner call) opt-in floor guard asserting sealed hermetic R/P and SC15 ≥ 90%.

## Invariants

- I1 Evidence or silence; truncation removes claims, never only their locators.
- I2 Tier honesty unchanged.
- I3 Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety and
  determinism 100%; SC15 focused and full ≥ 90.0%.
- I4 Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- I5 Additive only; unbudgeted calls byte-identical (gira-query goldens unchanged).

## Acceptance criteria

- SC1 envelope coverage: `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -count=1`
- SC2 determinism: `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -count=1`
- SC3 honesty: `go test ./internal/tokens ./internal/slices -run SCEnvelopeHonesty -count=1`
- SC4 transport ceiling: `go test ./internal/mcp -run SCTransportCeiling -count=1`
- SC5 structured errors: `go test ./internal/mcp ./internal/skills ./cmd/gira -run SCStructuredErrors -count=1`
- SC6 single registry: `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -count=1`
- SC7 CI hygiene (after approval): workflow greps for validate.js, gate-lint,
  `working-directory: bench`; `go test ./internal/gate -run SCGateLintStrictness -count=1`
- SC8 I3 numbers: EndToEnd published/perturbed R/P unchanged; bench SC15 ≥ 90% both tours.
- SC9 I4/I5: `git diff --exit-code 9ba80a6 -- fixtures/ testdata/ cmd/gira/testdata/gira-query/`;
  `"reviewed": false` in ground-truth.v2.json.
- SC10 regression: `go test -race ./...` root and bench/; validator exit 0.

## Interfaces

| Interface | Change |
|---|---|
| MCP | optional `budget` on get_tour, diff_maps, explain_evidence, locate_boundary; structured error data |
| CLI | `--budget` on tour, flows, diff, explain, locate; JSON error object under `--format json` |
| `budget` JSON | additive `estimator`, `measured` |
| internal/tokens, internal/mcp | single tool registry |
| .github/workflows/ci.yml | bench, validator, gate-lint (protected) |

ADRs: 0001 stands; 0006 tier budgets extend to all outputs (addendum if R4
becomes policy); 0011:30 CI edits need approval; 0012 prior art is idea-only.

## Risks

- bytes/4 is not tokens → R3 label, conservative ceilings.
- Golden churn → budget object only when a budget is set (Q1).
- Truncation hiding findings in diff_maps → R2 forbids dropping changed contracts.
- Promoting gate-lint may turn CI red on stale digests: honest No-Go.
- CI time on 3 OSes with bench (Q3).

## Plan (children, after approval)

1. -01-envelope-all-tools (R1, R2; SC1, SC2)
2. -02-envelope-honesty-and-ceiling (R3, R4; SC3, SC4), depends on 01
3. -03-structured-errors (R5; SC5)
4. -04-single-registry (R6; SC6)
5. -05-ci-hygiene (R7, R8; SC7), needs owner `approved:`

SC8–SC10 at every child. Rollback: `git revert` per child.

## Decisions

- Q1: envelopes opt-in only (no default on tools outside slice); calls without
  budget stay byte-identical.
- Q2: gate-lint enters CI in warn-only mode; strict stays for later.
- Q3: bench and validator in CI only on ubuntu-latest.
- Q4: yes to opt-in floor guard of floors (R/P hermetic and SC15 ≥ 90%),
  separate from revised gt thresholds.
- Q5: above the ceiling, refuse with budget_exceeded + escalate_hint (no
  pagination in v1).
- Q6: in E3: gitignore results.tsv/run.log, broken README.md:22 link,
  bench/quarantine ignoring .claude/worktrees/, and the false §6.1 of the
  validator in non-linear history. The .goreleaser stays for E5.
- Q7: F2 stays deferred, outside E3.

## Revisions

- v1 (2026-10-14): approved v1 with owner decisions Q1–Q7.