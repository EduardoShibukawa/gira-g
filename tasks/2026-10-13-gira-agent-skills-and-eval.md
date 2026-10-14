---
schema: task/1
tier: heavy
status: accepted
verified_at: "08665a9"
spec: gira-agent-skills-and-eval@1
---

# GIRA agent and human skills surface and comparative evaluation

## Routing and approvals

- tier: heavy (H1: product calls on skill scope, the human evaluation and comparison policy; H3: spans eval, slices, MCP, CLI and the ADR set; H5: multi-session decomposition)
- confirmed: heavy by owner (2026-10-13, "Heavy no ~/Git/gira-g (Recommended)")
- decided: where the governance artifacts live → standalone repo, validator run with `--root` by owner (2026-10-13, "Heavy no ~/Git/gira-g (Recommended)")
- decided: how to handle the ADR 0007 prior-art ban → new ADR 0012 that amends it by owner (2026-10-13, "Novo ADR 0012 (Recommended)")
- decided: corpus for the end-to-end eval → hermetic mini-corpus, full corpus opt-in by owner (2026-10-13, "Mini-corpus + opt-in (Recommended)")
- decided: what a No-Go on unreviewed ground truth does → report it and regenerate the ground truth; thresholds asserted only on reviewed ground truth, by owner (2026-10-13, "Reportar e regenerar GT (Recommended)")
- approved: spec v1 by owner (2026-10-13, "bora? use o /g-inga:drive e bora fazer tudo")
- approved: ADR 0012 by owner (2026-10-13, "bora? use o /g-inga:drive e bora fazer tudo")
- decided: manifest-integrity blocked → install the G-Ingá bundle in this repo (recommended option, under "bora fazer tudo") by owner (2026-10-13, "bora fazer tudo")
- divergence: spec SC1: the mutation names the Protobuf extractor, but removing it leaves recall unchanged because the Go AST extractor also creates the contract nodes. Child 01 removes Go AST instead (contract recall 0.33 → 0.17). Needs an owner exception or a spec v2.
- approved: exception SC1 — the mutation test removes the Go AST extractor instead of Protobuf, because Go AST also creates the contract nodes; the intent of SC1 (prove the eval depends on extraction) is kept. Resolves the SC1 divergence above. By owner (2026-10-13, "approved: exception SC1 — o teste de mutação remove o extrator Go AST em vez do Protobuf, porque o Go AST também cria os nós de contrato; a intenção do SC1 (provar que o eval depende da extração) está mantida. Isso resolve a divergence: já registrada.")
- approved: run children 02–11 via /g-inga:drive, standard tier each, no paid LLM calls, quarantine per ADR 0012, no push or release, by owner (2026-10-13, "/g-inga:drive --limit 11 --tag gira-agent-skills-and-eval@1")
- divergence: spec SC6: "every task in the 12-task suite". Tasks 05 and 06 target `contract:kafka:orders`, which the extractors never emit (Kafka topic Contracts are an extractor follow-up from child 01, outside this spec). No output exists to budget; the CLI fails with node-not-found and emits nothing (`TestBudgetOnMissingTargetEmitsNothing`). SC6 holds on the other 10 tasks. Options: approve an exception (SC6 measured on tasks whose target exists), or reopen as spec v2 adding Kafka topic extraction.
- divergence: spec Interfaces "no existing flag, field or tool changes meaning" vs SC4. Child 01 made `evidence[].discovered_at` the repository commit time instead of the wall-clock discovery time, because SC4 needs byte-identical cold indexes. Options: approve an exception and document the field's new meaning, or spec v2 adding a separate field.
- approved: exception SC6 — measured on the tasks whose target exists in the index; tasks 05/06 (`contract:kafka:orders`, never extracted) are covered automatically once the extractors emit the topic, and `TestBudgetOnMissingTargetEmitsNothing` pins that a missing target emits nothing. Resolves the SC6 divergence above. By owner (2026-10-13, "ok", answering "aprovo (a) nas duas" as recommended)
- approved: exception Interfaces/SC4 — `evidence[].discovered_at` is the commit time of the cited repository, documented on the field in `internal/domain/models.go`. Resolves the `discovered_at` divergence above. By owner (2026-10-13, "ok", answering "aprovo (a) nas duas" as recommended)
- decided: ground-truth corrections after a blind re-derivation by three subagents (6/12 tasks matched exactly) → task-02 adds service:agent downstream; task-07 drops service:agent downstream and adds service:opamp-server upstream; task-11 replaces service:product-catalog with service:frontend upstream; task-05/06 add the broker service:kafka upstream; task-01 keeps the topic consumers downstream; task-11 keeps env-gated service:mcp; `reviewed` stays false (owner's call) by owner (2026-10-13, "sim")
- accepted: by owner (2026-10-13, "aceito o accepted")

## Objective

Specify GIRA's agent-facing and human-facing skills catalog, and a quarantined comparative benchmark (layers 0, 1 and 2) against Graphify and Understand Anything. Then deliver it through ordered children. The first child removes the circularity in the internal evaluator.

In scope:

- the spec `gira-agent-skills-and-eval` v1;
- ADR 0012;
- decomposing the work into children after the spec is approved;
- integrated verification of SC1–SC16.

Out of scope:

- changes to the G-Ingá harness repository;
- paid benchmark runs without their own approval.

## Acceptance criteria

- AC1 = SC1–SC16 of `specs/gira-agent-skills-and-eval.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 6dacb25 | all 9 packages ok before any edit |
| B2 | `node tools/validate/validate.js` | blocked | 6dacb25 | manifest-integrity blocked (no bundle); bundle installed in e3645d1, the validator passes there |
| AC1 | SC1 to SC16 of spec v1, rows below | pass | 08665a9 | every SC row passes; SC6 with the open divergence on tasks 05/06 |
| SC1 | `go test -race ./internal/eval` with -run EndToEnd, Mutation, RejectsEmbedded | pass | 08665a9 | indexes a materialized git repo through indexer.Build; removing Go AST lowers contract recall 0.30 to 0.15 (owner exception); a fixture with macro_map is rejected |
| SC2 | `go test -race ./internal/eval` with -run Baseline, FailedTask | pass | 08665a9 | no baseline constants; a grown corpus raises the measured baseline while measured harness tokens stay equal; failed tasks add no tokens |
| SC3 | `go test -race ./internal/eval` with -run UnsupportedSafety | pass | 08665a9 | clean index 100; one injected extracted C++ record drops the score and fails the gate through SetCorpusChecks |
| SC4 | `go test -race ./internal/indexer ./internal/eval` with -run Deterministic, EndToEnd | pass | 08665a9 | two cold indexes with separate caches are byte-identical for published and perturbed |
| SC5 | `go test -race ./internal/eval` with -run GateSemantics | pass | 08665a9 | unreviewed No-Go reports; reviewed recall miss fails |
| SC6 | `go test -race ./cmd/gira ./internal/slices` with -run SC6, FocalContract, BudgetOnMissing, NoBudget, Fit, DropOrder | pass | 08665a9 | 10 of 12 tasks x 3 tiers x json, md, MCP within budget, byte-identical, focal kept; tasks 05/06 have no node (divergence above) |
| SC7 | `go test -race ./cmd/gira ./internal/gitdiff` with -run SC7, FromDiff, Parse, Touches, RunRejects | pass | 08665a9 | PaymentService proto edit gives the direct contract Impact Set on CLI, MCP and supplied diff; file-level fallback and unknown paths covered |
| SC8 | `go test -race ./cmd/gira` with -run SC8, Flows, FlowAnchors | pass | 08665a9 | byte-identical tours; every read and why_locator resolves in bounds at the pinned SHA or the why says unknown; why templated; order-flow coverage 81.8% |
| SC9 | `go test -race ./cmd/gira ./internal/archdiff` with -run SC9, ArchDiff, Segments, Ambiguous | pass | 08665a9 | 24 renames (origin inferred) and 2 removed edges exactly as the recipe declares; nothing else structural |
| SC10 | `go test -race ./internal/skills` | pass | 08665a9 | 3 hosts from one catalog; second run unchanged; only markers or the owned JSON key change; every SKILL.md declares trigger, operations, contract, budget |
| SC11 | `go test -race ./cmd/gira` with -run SC11, GiraQuerySkill | pass | 08665a9 | 11 questions incl. pr_impact byte-identical via MCP and CLI and equal to goldens |
| SC12 | `go test -race ./quarantine` inside bench/ | pass | 08665a9 | no guarded file references the bench (bench as a path word, go.work, embedded assets); out/ ignored; query phases network none; script runs by image id with declared mounts; recipes pinned |
| SC13 | `go test -race ./layer0` inside bench/ | pass | 08665a9 | 3 adapter goldens; locator checker rejects out-of-bounds, anchor mismatch, missing file |
| SC14 | `go test -race ./layer1` inside bench/ | pass | 08665a9 | 120 mock records (12 tasks x A, B x 5) valid against the schema; scorer equals the hand-computed fixture |
| SC15 | `go test -race ./layer2` inside bench/ and read of bench/layer2/study-protocol.md | pass | 08665a9 | proxies on real GIRA tours: focused tour assertiveness 5.9% (1/17), coverage 81.8%; protocol has tasks, Latin square, rubric, instruments |
| SC16 | `node tools/validate/validate.js` (and R1) | pass | 08665a9 | overall pass; no fail |
| R1 | `go test -race ./...` at the root and inside bench/ | pass | 08665a9 | all packages ok; existing e2e and dogfood tests unchanged |
| V1 | read `git diff 41fd67b..08665a9` against spec v1, plus the fresh-context review | pass | 08665a9 | all SC rows mapped to code and tests; review findings resolved below |

## Constraints and budget

- External dependency: the validator lives in `~/Git/duh-homelab/tools/validate/` and runs with `--root`. `manifest-integrity` stays `blocked` until the G-Ingá bundle is installed in this repo. That needs an owner decision before the parent can be `verified`.
- Paid LLM runs (Layer 1 frontier, Understand Anything indexing) are protected actions. Each needs its own `approved:` entry with a cost estimate.

## Plan

Children created at `@1` after approval (files `tasks/2026-10-13-gira-agent-skills-and-eval-NN-*.md`), in dependency order:

1. `-01-eval-end-to-end`: shared indexing unit; hermetic mini-corpus and Go perturbation; ground truth v2; measured baseline; real Unsupported Safety; index determinism; gate semantics (SC1–SC5).
2. `-02-ground-truth-rederivation`: re-derive the ground truth against the full pinned corpus (opt-in run); human review sets `reviewed: true` (depends on 01).
3. `-03-hard-budgets`: estimator, deterministic truncation and the `budget` object on CLI and MCP (SC6; depends on 01).
4. `-04-impact-from-diff`: `--from-diff` and `get_pr_impact` (SC7; depends on 03).
5. `-05-tours-and-flows`: `tour`, `flows`, `get_tour` (SC8; depends on 03).
6. `-06-arch-diff`: `diff`, `diff_maps` (SC9; depends on 01).
7. `-07-skills-distribution`: canonical skill sources, `skills install` for 3 host families, the `/work` integration contract (SC10, SC11; depends on 04, 05, 06).
8. `-08-bench-quarantine`: containers, phases, ignored output and the static isolation check (SC12; depends on ADR 0012 accepted).
9. `-09-layer0-normalizer`: NAG, the three adapters and the locator checker (SC13; depends on 08).
10. `-10-layer1-agent-runner`: MCP client loop, mock-model mode, telemetry and the scorer (SC14; depends on 08 and 03).
11. `-11-layer2-human-eval`: tour proxies and the study protocol (SC15; depends on 05).

## Review

Fresh-context adversarial review (`.claude/skills/work/review.md`) of `41fd67b..d92b6d4`, run by a subagent on 2026-10-13. It reported 11 findings: 4 major, 7 minor.

- [F1] major, SC6: a tight budget could drop the focal Contract node from `affected_contracts` (and PR-impact targets) -> fixed in 8db9304
- [F2] major, I1/SC8: tour entry claims ("declares ingress on ports") cited a locator that does not support them; the Layer 2 anchor for `reached` steps was trivially true -> fixed in 8db9304 (steps carry `why_locator`; entry claims say their locator is unknown; Layer 2 checks only `why_locator`)
- [F3] major, I1/A.1: `--from-diff` reported a cited file as "unknown to the index" when no cited range changed -> fixed in 8db9304 (file-level fallback, `file_level_paths`)
- [F4] major, SC12: the static check missed `cd bench`, `working-directory: bench`, `go.work` and embedded assets -> fixed in 8db9304
- [F5] minor, I1: arch diff items had no locator, and renames were not labelled inferred -> fixed in 8db9304
- [F6] minor, ADR 0012 rule 1: the query phase ran by mutable tag, with mounts different from phases.json -> fixed in 8db9304 (runs by the recorded image id with the declared mounts; resolved pip dependencies recorded). Graphify's transitive PyPI dependencies stay resolved at build time and are only recorded: accepted residual risk, since the source itself is pinned by commit.
- [F7] minor, SC6: tasks 05/06 skipped -> rejected: no output exists to budget, because the node is never extracted (out of scope). Raised to the owner as a `divergence:` above. `TestBudgetOnMissingTargetEmitsNothing` pins the behavior.
- [F8] minor, SC2/SC3: the tests were weaker than the criteria -> fixed in 8db9304 (they now run the real evaluator and gate path)
- [F9] minor, spec C: failed tasks inflated token reduction -> fixed in 8db9304 (published token reduction is now 73.4%, previously 75.2%)
- [F10] minor, Interfaces: `discovered_at` changed meaning in child 01 -> rejected: SC4 requires it. Raised to the owner as a `divergence:` above.
- [F11] minor, SC11: no `pr_impact` question in the golden -> fixed in 8db9304

Review focus: (a) the reviewer found no circular path; (b) budgets: F1 was the only hole, now fixed; (c) facts without a locator: F2, F3 and F5, all now fixed.

## Checkpoint

- as of: 2026-10-13, branch main, head 8db9304 plus this record, worktree clean after commit
- done: children 01–11 verified; review findings fixed; SC1–SC16, R1 and V1 pass at 8db9304
- done also: SC1–SC16, R1 and V1 rerun at 08665a9 (corrected ground truth); parent verified there
- done also: accepted by owner
- next: extractor improvements in a new task (prompt handed to the owner); `reviewed: true` stays the owner's call
- done also: both divergences resolved by owner exceptions; parent verified at 8db9304
- open: no push, no release; paid runs (Layer 1 frontier, UA indexing) not performed

## Outcome

Delivered (spec v1): non-circular evaluator (01), opt-in full corpus and anchor checks (02), hard budgets (03), impact from diff (04), tours and flows (05), arch diff (06), skills catalog, installer and gira-query (07), quarantined bench workspace (08), Layer 0 (09), Layer 1 runner (10), Layer 2 proxies and study protocol (11). All CLI and MCP changes are additive. New tools: `get_pr_impact`, `get_tour`, `diff_maps`. New commands: `tour`, `flows`, `diff`, `skills install`, `skills nudge`.

Gate results on unreviewed ground truth (reported, not enforced):

| Variant | Decision | Recall | Precision | Token reduction | Safety | Determinism |
|---|---|---|---|---|---|---|
| Published, hermetic | No-Go | 58.8% | 32.8% | 73.4% | 100% | 100% |
| Perturbed, hermetic | No-Go | 56.2% | 30.5% | 75.8% | 100% | 100% |
| Published, full corpus (opt-in) | No-Go | 60.3% | 30.4% | 99.1% | 100% | 100% |

These figures are against the corrected ground truth (see the `decided:` entry). Before the correction they were 64.6%, 62.3% and 64.6% recall. The drop comes from real facts GIRA does not extract yet: agent → frontend over HTTP, opamp, and the topic broker.

The ADR 0007 gate is not met. This parent does not claim it.

Follow-ups, outside this spec:
- extractors: Kafka topic Contracts; Dockerfile stages and base images emitted as nodes; `service:src` naming; edge evidence at the dependency line rather than the service header (that last one caps Layer 2 assertiveness at 5.9%); a locator for declared ports;
- add `CLAUDE.md` to the corpus memory-leak rule;
- wire arms C and D into the Layer 1 runner after their images are built.
