# Standard Ceremony Procedure

Procedure for `standard` tasks in G-Ingá on Antigravity.

## Steps

1. **Routing Line:**
   Always record initial entry: `tier: standard (<reason or signal that failed light>)`.

2. **Initialize or Read Task Record (`tasks/<id>.md`):**
   - File naming: `tasks/<YYYY-MM-DD>-<slug>.md` (or decomposed `tasks/<parent-id>-<NN>-<slug>.md`).
   - Schema: `task/1`, as defined in `.agents/skills/work/references/artifact-contracts.md` §3.
   - Frontmatter keys: `schema: task/1`, `tier: standard`, `status: active` (plus `parent`, `spec`, `depends_on` if decomposed).
   - Sections required:
     - `## Routing and approvals`
     - `## Objective`
     - `## Acceptance criteria`
     - `## Verification`
     - `## Plan` (if S2 fails or step order matters, else `skip: <reason>`)
     - `## Outcome` (when reaching completion/verification)

3. **Capture Baseline:**
   - If S6 holds: run relevant checks before editing and record as `B` row in verification table (`B1 | <suite> | pass/fail | <sha> |`).
   - If S6 fails (no existing baseline check): record `B1` row as `not run` with note.

4. **Incremental Implementation:**
   - Make small, focused changes.
   - If blocked or stopping, write `## Checkpoint` with branch, commit, dirty status, completed items, and next steps. Set `status: blocked` or leave `active`.

5. **Verification & Guarded Transition (`.agents/skills/work/references/artifact-contracts.md` §3.4):**
   - Execute each acceptance criterion check.
   - Record results using only `pass`, `fail`, `blocked`, `not run`.
   - Commit code changes so a concrete commit SHA exists for `verified_at`.
   - Update `## Verification` rows with the commit SHA in `Ref`.
   - Run validator: `node tools/validate/validate.js` or `bun tools/validate/validate.ts`.
     - Exit `0`: all checks passed or not run. Transition to `status: verified` with `verified_at: "<sha>"`.
     - Exit `1` (`fail`): invariant or schema violation. Transition forbidden. Fix record/code.
     - Exit `2` (`blocked`): environmental barrier, shallow clone, or prerequisite missing. Transition to `verified` forbidden; record stays `completed` or `blocked`.
   - Commit the verified task record update.

6. **Outcome:**
   - Write `## Outcome` summarizing delivered results.
