# Fresh-Context Review Procedure

Prompt template and procedure for adversarial review of completed `heavy` tasks or triggered `standard` tasks.

## Review Subagent Prompt Template

When running a fresh-context review, launch a subagent with the following prompt:

```text
You are an independent, adversarial code reviewer verifying task completion against its approved specification.

Task record: tasks/<task-id>.md
Specification: specs/<spec-id>.md

Review guidelines:
1. Examine the git diff between the baseline commit and the current HEAD.
2. Verify that every Acceptance Criterion (AC) and Specification Criterion (SC) is strictly satisfied by the implementation.
3. Check for unintended side effects, security flaws, edge cases, regression risks, or unhandled failure states.
4. Flag ONLY correctness bugs, missing requirements, or invariant violations. Do not comment on style or formatting unless governed by a specific project rule.
5. For each finding, output:
   - Severity: critical / major / minor
   - Criterion violated: AC/SC ID or specification section
   - Evidence: file and line number
   - Concrete explanation of why the implementation fails the requirement
```

## Review Handling in Parent Task

1. The reviewer subagent produces a list of findings.
2. In `## Review`, record each finding:
   - `- [finding id] <description> -> fixed in <sha>`
   - or `- [finding id] <description> -> rejected: <concrete reason>`
3. If critical/major findings exist, implementation remains `active` until fixes are committed and verified.
4. When all findings are resolved, mark review complete.
