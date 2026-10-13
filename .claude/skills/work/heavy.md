<!-- Adapted from Matt Pocock's skills at 49dd158: skills/grill-with-docs/SKILL.md, skills/to-spec/SKILL.md, skills/to-tickets/SKILL.md -->
# Heavy Ceremony Procedure

Procedure for `heavy` tasks in G-Ingá.

## Overview

Heavy ceremony is reserved for one-way doors (H2), sensitive behavior (H4), cross-module architecture changes (H3), multi-session work (H5), or major ambiguity (H1).

When both H1 and H5 hold and the destination cannot yet be captured in a single specification, point to the repository decision-map convention (`docs/decision-map.md`) rather than attempting a single giant spec (ADR 0005 item 10).

## Lifecycle

1. **Owner Confirmation:**
   - Propose heavy tier to owner with triggering signals (H1–H6).
   - Await explicit owner confirmation before leaving `todo` or creating decomposed children:
     `confirmed: heavy by owner (<date>, "<ref>")`.

2. **Question Rounds & Spec Authoring:**
   - Investigate technical facts inline.
   - For product or preference decisions, conduct structured question rounds with recommended answers:
     Present the question, options, trade-offs, and recommend one option with clear rationale.
   - Record owner decisions: `decided: <question> → <answer> by owner (<date>, "<ref>")`.
   - Author specification in `specs/<id>.md` (`schema: spec/1`, `status: draft`, `version: 1`).
   - Run validator preflight before approval.
   - Owner formally approves spec:
     - Set `status: approved`, `approved_version: 1`, `approved_hash: sha256:<hash>`.
     - Parent task records: `approved: spec v1 by owner (<date>, "<ref>")`.

3. **Task Decomposition:**
   - Decompose approved spec into ordered child tasks: `tasks/<parent-id>-01-<slug>.md`, `tasks/<parent-id>-02-<slug>.md`.
   - Each child has `tier: standard`, `parent: <parent-id>`, `spec: <spec-id>@<version>`, `depends_on: [...]`.
   - Run validator to verify decomposition invariants (spec approved, valid dependency graph, no cycles).
   - Execute each child following `standard.md`.

4. **Specification Reopening (`.claude/skills/work/references/artifact-contracts.md` §4.3, §4.4):**
   - If requirements change: increment `version` to `v<N+1>`, set `status: draft`, append to `## Revisions`.
   - Reopening immediately blocks active children until owner reapproval.
   - Upon reapproval, update children with `spec-version: v<N> → v<N+1>; rerun: ...` or `...; unaffected: <reason>`.

5. **Integrated Verification & Adversarial Review:**
   - Integrate all children results into parent record.
   - Run end-to-end spec criteria (`SC` rows) and regression test (`R` rows).
   - Run fresh-context adversarial review using the prompt template in `.claude/skills/work/review.md`.
   - Resolve all review findings in `## Review` section.

6. **Owner Acceptance:**
   - Present final integrated evidence to owner.
   - Record explicit acceptance: `accepted: by owner (<date>, "<ref>")`.
   - Set parent task `status: accepted`, `verified_at: "<sha>"`.
