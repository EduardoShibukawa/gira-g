---
schema: spec/1
status: draft
version: 1
---

# GIRA envelopes and hygiene (E3)

> DRAFT v1, unapproved. Scope inferred from commitments other specs hand to E3
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

## Expected behavior

One epic, five additive features.

### F1 — Envelope on every output

Optional `budget`/`--budget` on get_tour, diff_maps, explain_evidence, locate_boundary and CLI tour/flows/diff/explain/locate, reusing `slices.Budget`. Absent/0 keeps today's bytes exactly. Deterministic per-kind drop order that never drops the focal element nor unsupported/unknown rows (tour: trailing steps then extra read locators; diff: evidence_moved then unchanged context, never changed_contracts or newly_unsupported; explain: evidence beyond the first per fact; locate: suggestions beyond N). Same index + budget → byte-identical output.

### F2 — Honest envelope and transport ceiling

`budget` gains `estimator: "bytes/4"`, `measured: false` (additive, only when a budget is set). No envelope figure enters an E1 measured headline. One global MCP response byte ceiling, set from measurement above every current hermetic output; above it, refuse with `budget_exceeded` + `escalate_hint`; never cut mid-JSON.

### F3 — Structured errors

Structured errors `{code, message, retryable, suggestions}` on MCP and CLI `--format json`; gira-query reads `code` first, substring fallback for old binaries.

### F4 — Single tool registry

Single tool registry (or a set-equality test); each tool declares `enveloped`.

### F5 — CI hygiene

Bench race, validator, gate-lint at owner-chosen promotion level on ubuntu-latest. Opt-in floor guard asserting sealed hermetic R/P and SC15 ≥ 90%. Hygiene fixes: gitignore results.tsv/run.log, fix README.md:22 link, bench/quarantine ignoring .claude/worktrees/, fix validator §6.1 false positive in non-linear history.

## Scope

**In scope**

- F1–F5 above, each additive (no flag/field meaning changes; budgets/truncation untouched except where explicitly required).
- Seeded floor guard config (not model calls).
- Warn-only CI wiring for gate-lint (strict enforced locally).
- Hygiene fixes listed in F5.

**Out of scope**

- E2 joins, E4 human surface, E5 skills distribution, E6 docs site.
- Ground-truth changes, `reviewed: true`.
- Any model client (ADR 0001), any paid multi-model fan-out, any semantic harvest.
- Compression/summarization; a model tokenizer (ADR 0001).
- .goreleaser target (E5).

**Non-goals**

- Raising recall/precision, new joins/extractors (E2 F2 stays deferred).
- Human-surface styling (E4), release/signing (E5), docs site (E6).

## Acceptance criteria

- SC1 envelope coverage: `go test ./internal/mcp ./cmd/gira -run SCEnvelopeEveryTool -count=1`
- SC2 determinism: `go test ./internal/slices ./cmd/gira -run SCEnvelopeDeterministic -count=1`
- SC3 honesty: `go test ./internal/tokens ./internal/slices -run SCEnvelopeHonesty -count=1`
- SC4 transport ceiling: `go test ./internal/mcp -run SCTransportCeiling -count=1`
- SC5 structured errors: `go test ./internal/mcp ./internal/skills ./cmd/gira -run SCStructuredErrors -count=1`
- SC6 single registry: `go test ./internal/mcp ./internal/tokens -run SCToolRegistrySingleSource -count=1`
- SC7 CI hygiene (after approval): workflow greps for validate.js, gate-lint, `working-directory: bench`; `go test ./internal/gate -run SCGateLintStrictness -count=1`
- SC8 I3 numbers: EndToEnd published/perturbed R/P unchanged; bench SC15 ≥ 90% both tours.
- SC9 I4/I5: `git diff --exit-code 9ba80a6 -- fixtures/ testdata/ cmd/gira/testdata/gira-query/`; `"reviewed": false` in ground-truth.v2.json.
- SC10 regression: `go test -race ./...` root and bench/; validator exit 0.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| MCP | optional `budget` on get_tour, diff_maps, explain_evidence, locate_boundary; structured error data |
| CLI | `--budget` on tour, flows, diff, explain, locate; JSON error object under `--format json` |
| `budget` JSON | additive `estimator`, `measured` |
| internal/tokens, internal/mcp | single tool registry |
| .github/workflows/ci.yml | bench, validator, gate-lint (protected) |

ADRs: 0001 stands; 0006 tier budgets extend to all outputs (addendum if R4 becomes policy); 0011:30 CI edits need approval; 0012 prior art is idea-only.

## Decisions

- Q1: envelopes opt-in only (no default on tools outside slice); calls without budget stay byte-identical.
- Q2: gate-lint enters CI in warn-only mode; strict stays for later.
- Q3: bench and validator in CI only on ubuntu-latest.
- Q4: yes to opt-in floor guard of floors (R/P hermetic and SC15 ≥ 90%), separate from revised gt thresholds.
- Q5: above the ceiling, refuse with budget_exceeded + escalate_hint (no pagination in v1).
- Q6: in E3: gitignore results.tsv/run.log, broken README.md:22 link, bench/quarantine ignoring .claude/worktrees/, and the false §6.1 of the validator in non-linear history. The .goreleaser stays for E5.
- Q7: F2 stays deferred, outside E3.

## Risks and constraints

- bytes/4 is not tokens → R3 label, conservative ceilings.
- Golden churn → budget object only when a budget is set (Q1).
- Truncation hiding findings in diff_maps → R2 forbids dropping changed contracts.
- Promoting gate-lint may turn CI red on stale digests: honest No-Go.
- CI time on 3 OSes with bench (Q3).

## Verification

- Per child + integrated: B1 `go test -race ./...` (root); B2 `go test -race ./...` (bench/); B3 `go test ./internal/gate -run EndToEnd -v` (both hermetic variants, report recall/precision/safety/determinism); B4 opt-in full-corpus run (reported only); B5 `node tools/validate/validate.js`.
- AC rows map SC1–SC10 to the `go test` names above; `not run` forbidden at parent seal.
- **Review focus:** does the ceiling actually cover all hermetic outputs? Is truncation deterministic and honest? Is the error vocabulary complete? Does the floor guard opt-in correctly? Does CI run bench+validator only on ubuntu-latest? Any weakened assertions or named guesses?

## Rollback

`git revert` per child. No migration, publication or external state. Hygiene fixes are additive; reverting restores pre-E3 behavior.

## Open questions

- deferred: **exact CI promotion point** for strict mode (E4+ decides). Safe: warn-only needs no interface change later.
- deferred: **register retention/GC policy**. Safe: pointer-only files, additive.

## Revisions

- v1 (2026-10-14): draft v1 with owner decisions Q1–Q7.