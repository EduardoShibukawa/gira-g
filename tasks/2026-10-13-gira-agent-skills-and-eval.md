---
schema: task/1
tier: heavy
status: active
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
| B2 | `node ~/Git/duh-homelab/tools/validate/validate.js --root .` | blocked | 6dacb25 | manifest-integrity blocked: the repo has no G-Ingá bundle `manifest.json`; artifact-records not run (no records yet) |
| SC1–SC16 | see spec §Verification | not run | | runs after the children are verified |
| R1 | `go test -race ./...` on the integrated head | not run | | |
| V1 | read the integrated diff against spec v1 | not run | | |

## Constraints and budget

- External dependency: the validator lives in `~/Git/duh-homelab/tools/validate/` and runs with `--root`. `manifest-integrity` stays `blocked` until the G-Ingá bundle is installed in this repo. That needs an owner decision before the parent can be `verified`.
- Paid LLM runs (Layer 1 frontier, Understand Anything indexing) are protected actions. Each needs its own `approved:` entry with a cost estimate.

## Plan

Children are created only after the spec is approved, at `@1`, in dependency order:

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

skip: the fresh-context adversarial review runs after the children are integrated (spec §Verification).

## Checkpoint

- as of: 2026-10-13, branch main, head 6dacb25, worktree dirty (specs/, tasks/, docs/adr/0012)
- done: research and inspection; owner question round 1 (4 decisions); spec v1 drafted; ADR 0012 proposed
- next: owner reviews spec v1 and ADR 0012 → approval → set `approved_hash`, then decompose into children; start child 01
- open: the validator's `manifest-integrity` is blocked in this repo (see Constraints)
- revalidate: B1 ran on 6dacb25
