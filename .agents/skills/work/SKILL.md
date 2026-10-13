---
name: work
description: Primary SDLC harness entry point for G-Ingá on Antigravity. Dispatches work across light, standard, and heavy ceremony tiers and handles resumption.
disable-model-invocation: false
---

# /work — G-Ingá SDLC Harness

The single entry point for changes to deliverables in G-Ingá. Dispatches requests through three ceremony tiers (`light`, `standard`, `heavy`) and handles resumption of paused tasks.

## Invariants

1. Every task execution begins by evaluating routing signals and printing a formatted routing line: `tier: <tier> (<signals>)`.
2. `light` tasks persist ZERO files under `tasks/` or `specs/`. They produce only a concise terminal report with the routing line and verification evidence.
3. `standard` and `heavy` tasks persist contract-compliant artifacts under `tasks/` and `specs/` adhering to the artifact contract, `.agents/skills/work/references/artifact-contracts.md` (delivered with this skill; a generated copy of G-Ingá's canonical contract).
4. Guarded transitions (`verified`, `accepted`, spec approvals) and task record commits require running the bundled validator (`node tools/validate/validate.js` or `bun tools/validate/validate.ts`).
5. Mandatory check failures (`fail` exit 1, `blocked` exit 2) prevent transition to `verified` or `accepted`.

## Routing Rules (ADR 0002)

Evaluate signals S1–S8:

- **S1 (Intent clarity):** Can outcome be stated as checkable criteria without product/preference call?
- **S2 (Approach clarity):** Can diff be described in one sentence following existing pattern?
- **S3 (Footprint):** Inside one module without adding/changing consumed public contracts?
- **S4 (Reversibility):** Fully reversible two-way door?
- **S5 (Sensitive behavior):** Security, auth, permissions, secrets, privacy, money, or harness policy?
- **S6 (Baseline quality):** Relevant check exists and passes before change?
- **S7 (Error detectability):** Mistakes caught quickly by existing check?
- **S8 (Scale):** Fits in one session without decomposition?

### Ceremony Decision

1. **`heavy`** if ANY trigger holds:
   - **H1:** S1 fails (unsettled product/preference call).
   - **H2:** S4 fails (one-way door: migration, deletion, publication).
   - **H3:** S3 fails (contract consumed elsewhere changes, architecture changes, multiple subsystems).
   - **H4:** S5 holds (sensitive behavior changes).
   - **H5:** S8 fails (requires multi-session decomposition).
   - **H6:** S6 fails AND S7 fails (weak baseline and undetected error).
   *Action:* Stop. Read `.agents/skills/work/heavy.md`. Propose heavy to owner and await explicit confirmation.
2. **`light`** if ALL hold: S1, S2, S3, S4, S6 (or non-behavioral prose/formatting), S7, and NOT S5.
   *Action:* Follow light procedure below.
3. **`standard`** otherwise (default when in doubt).
   *Action:* Read `.agents/skills/work/standard.md`.

## Escalation Triggers (ADR 0002 §5)

If new evidence arises during execution, escalate immediately:

- **E1:** New evidence breaks a `light` condition or reveals a `heavy` trigger.
- **E2:** The diff outgrows its one-sentence description or reaches another module.
- **E3:** Verification fails twice without new information.
- **E4:** An affected check fails in baseline and failure cannot be attributed.
- **E5:** Implementation diverges from approved specification.

On escalation: stop editing, output `escalated: <from> → <to> because <E1–E5>: <evidence>`, keep still-valid evidence, complete missing entry steps of new tier. Escalation to `heavy` stops and awaits owner confirmation.

## Resume Command

If invoked with `resume <task-id>`:

- Read `.agents/skills/work/resume.md`.
- Inspect `tasks/<task-id>.md`, check git state against `## Checkpoint`, reconcile discrepancies, and resume.

## Light Tier Procedure

1. **Routing:** Output routing line: `tier: light (S1–S4, S6, S7 hold; S5 doesn't)`.
2. **Inspect:** Check relevant baseline tests or linters.
3. **Execute:** Make minimal required edits.
4. **Verify:** Run cheap relevant check and read diff (`git diff`).
5. **Report Format (`.agents/skills/work/references/artifact-contracts.md` §6):**

   ```text
   tier: light (<deciding signals>)
   Changed: <file>, <description of change> (<N> lines)
   Checks:
   - <Check 1>: pass / fail / blocked / not run (<details>)
   - <Check 2>: pass / fail / blocked / not run (<details>)
   ```

6. **Zero persistence:** Do NOT create or edit any file in `tasks/` or `specs/`.
