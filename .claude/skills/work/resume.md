# Resume Procedure

Procedure for resuming paused or inter-session tasks in G-Ingá.

## Invocation

Called via `/work resume <task-id>` or when resuming work referenced in `tasks/<task-id>.md`.

## Steps

1. **Locate Task Record:**
   - Read `tasks/<task-id>.md`.
   - Inspect YAML frontmatter (`status`, `tier`, `parent`, `spec`, `depends_on`).

2. **Reconcile Git State:**
   - Read the latest entry in `## Checkpoint`.
   - Compare recorded branch, commit SHA, and modified files with actual working tree (`git status`, `git log -n 1`).
   - If working tree differs from checkpoint, reconcile and note differences before continuing.
   - Any verification rows whose `Ref` commit is not an ancestor of the current HEAD must be flagged as stale and re-verified.

3. **Check Dependencies:**
   - If `depends_on` lists tasks, verify that each dependency has `status: verified` or `status: accepted`.
   - If any dependency is incomplete, stop and report `blocked`.

4. **Verify Spec Alignment (Heavy Children):**
   - If task has `spec: <spec-id>@<version>`, inspect `specs/<spec-id>.md`.
   - Verify that spec is `status: approved` at that version.
   - If spec has been reopened or superseded, handle per `.claude/skills/work/references/artifact-contracts.md` §4.4.

5. **Resume Execution:**
   - Remove or update blocker in `## Blockers`.
   - Transition `status: active`.
   - Continue execution under `standard.md` or `heavy.md`.
