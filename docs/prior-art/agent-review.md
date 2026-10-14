# Prior art: agent-review (Gallon methodology)

[agent-review](https://github.com/captivus/agent-review): MIT. Skill v1.0.0 `031841c` (2026-10-11, templates only, no code). Article: <https://gallon.me/i-believe-everything-my-agents-tell-me> (2026-10-13). Premise: never trust the first agent without an independent validation layer.

## How we use it

Methodology to **adopt partially** (pilot on heavy/ground-truth); no model client in `gira`, no default paid fan-out.

## USE / COPY / NORTH STAR

- **USE:** templates as spec (briefs, reconciliation log, `reviews/findings.md`).
- **COPY:** L1 `ARTIFACT-VERIFIED <sha256>` closing line + frozen out-of-write-scope check + 5-point dispatcher self-check; L2 control total with limits + frozen blind A/B + seeded `max(20,5%)` + never-force-match + never-plug/drop; L3 S1 measured/checked/inferred/assumed taxonomy (code-reading=inferred, repro follows the real path) + verified-before-edit findings + independent fix-check + register with coordinator/reviewer severity.
- **NORTH STAR:** hash-pinned S3 fresh-eyes on heavy/external conclusions only; personas as fixed templates (not multi-model); PASS = evidence, not proof; marker test + enforced/instructed label.

## Study 1 — 2026-10-14 (subagent)

- **Fact:** `SKILL.md` + `references/{level-1-dispatch,level-2-brief,level-3-reviewer-prompts,tool-notes}.md`; article cases: 19min/~16M tokens (L1), 56 diffs balanced-to-penny (L2), auto-renewal caught by 2 cold lawyers (L3); author admits "hasn't been measured".
- **Mapping:** L1 covered in spirit (Verification tables, `verified_at`, validator); L2 partial (3-subagent blind re-derivation, locator checker, gate semantics exist — control total, frozen digests, sampling missing); L3 S1/S2-lite exist (`review.md`) — always-on taxonomy, register, fix-check, S3, convergence/hand-back missing.
- **Open:** full-stack detection rate; quarantine/token cost of persona rounds.
