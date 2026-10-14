# gira-g — agent guidance

Only what the code can't tell you. Everything else: read the repo.

## Commit and push gates

- Git hooks come from `lefthook.yml` and fail closed. Commit and push through mise so the hook finds its tools: `mise exec -- git commit …`, `mise exec -- git push …`.
- Never pipe a validator, linter or test into `tail`/`head`/`grep`, or chain it with `&&`, in the same command as `git commit` or `git push`. The pipe's exit code hides the real one. Run the check alone, read its exit code, then commit.
- Never commit after a failed check, and never use `--no-verify`.
- `golangci-lint` has pre-existing debt on `main`; the gate blocks only issues new since `origin/main`. Don't fix unrelated lint in a feature diff.

## Pull requests

- Auto-merge is disabled. Open the PR, wait for CI to go green, then tell the owner it's ready. The owner merges.
- After the owner says they merged, `git fetch origin` before reading `main`.
- Before ending a session, report branches and worktrees not yet on `origin/main`.

## Agents and cost

- Subagents default to Sonnet, or Haiku for plain search. Opus only for spec writing, ADRs and adversarial review, or when the owner asks.
- In a new worktree, run `mise trust` before any other `mise` command.

## Working with the owner

- The owner writes in Portuguese. Short replies such as "s", "ok" or "Bora" mean approval.
- Treat the owner's brief as the spec; don't ask what it already answers.
- G-Ingá (the harness) is project-agnostic. Never frame it around one target system.
- For Claude Code plugin, marketplace or settings questions, check the official docs first.

## Harness

- `/work` here is the generated G-Ingá bundle copy in `.claude/skills/work`, integrity-checked against `manifest.json` by `node tools/validate/validate.js`. Don't edit it in place; change G-Ingá and regenerate.
- Task records and specs follow `.claude/skills/work/references/artifact-contracts.md`.
