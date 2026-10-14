# Spike: next epic after E2 (read-only, main @ 9ba80a6)

Status: proposal to the owner. Nothing implemented. Spec below is DRAFT v1, unapproved.

## Where the queue lives

No map.md/roadmap defines E3–E6; they are one-line names in
specs/gira-evidence-gate.md:76, specs/gira-recall-joins.md:143-144,
tasks/2026-10-14-gira-evidence-gate.md:33, tasks/2026-10-14-gira-recall-joins.md:32.
README.md:22 links to `../../.scratch/system-context-harness/map.md` (broken path).
The meaning of each epic below is inferred from commitments other specs hand to it.

## Comparison

| Epic | Goal (from repo text) | Already built | Depends on | Size |
|---|---|---|---|---|
| E3 envelopes/hygiene | Hard envelope on every output; decide CI strictness (E1 deferrals) | Budgets on 3 of 12 MCP tools; gate-lint warn-only | E1 (done) | M–L (5 children) |
| E4 human surface | Close ADR 0008 gaps (lipgloss/NO_COLOR fallback, Layer 2 moderated study) | explain/locate/export/view/tour/flows/diff exist | stable outputs (E3) | S–M |
| E5 skills distribution | Signed releases + skills (ADR 0010:46-49) | `skills install` verified for 3 hosts | E3 skill contracts; owner-approved release (ADR 0011:30); .goreleaser targets wrong repo; no LICENSE | M |
| E6 docs site | Optional `website/` (docs/prior-art/docusaurus.md) | nothing | E4/E5 content | M |

**Recommendation: E3.** E1 explicitly hands it envelope enforcement and the CI
strictness call; E5 would otherwise ship skill contracts E3 then changes; 9 of 12
tools can still blow an agent's context; it does not touch extractors, so R/P
should stay frozen and that is cheap to verify.

Side findings: the SC15 ≥ 90% floor is logged, never asserted
(bench/layer2/layer2_test.go:61-63); CI runs only root tests
(.github/workflows/ci.yml:39); bench/quarantine scans `.claude/worktrees/` and
fails while agent worktrees exist inside the repo.

---

```markdown
---
schema: spec/1
status: draft
version: 1
approved_version: none
approved_hash: none
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

## Open questions

- Q1 Non-slice envelopes opt-in only (recommended) or default-on with per-tool
  defaults (tour 3500, diff 3500, explain 1500, locate 500)?
- Q2 gate-lint in CI: wire warn-only (recommended), strict on main only, or strict everywhere?
- Q3 bench + validator in CI on all 3 OSes or ubuntu only?
- Q4 Accept R8 floor guards on unreviewed ground truth?
- Q5 Past the ceiling: refuse with budget_exceeded (recommended) or cursors/pagination?
- Q6 Hygiene extras in E3 or deferred: gitignore results.tsv/run.log; fix
  README.md:22 link; bench/quarantine skipping .claude/worktrees; .goreleaser
  target (probably E5)?
- Q7 E2 F2 stays deferred and outside E3 (recommended)?

## Revisions

- v1 (2026-10-14): read-only spike draft, unapproved.
```
