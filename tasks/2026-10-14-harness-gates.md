---
schema: task/1
tier: standard
status: active
---

# Local commit gates and agent guidance for gira-g

## Routing and approvals

- tier: heavy (H4: changes harness policy, the commit gates and agent directives)
- confirmed: heavy by owner (2026-10-14, "Heavy em worktree (recommended)")
- decided: gate split → fast on commit, heavy on push by owner (2026-10-14, "Rápido no commit, pesado no push (recommended)")
- decided: /work bundle copy → keep and document by owner (2026-10-14, "Manter e sincronizar (recommended)")
- decided: loose root files → spikes to docs/spikes, rest ignored by owner (2026-10-14, "Spikes → docs/spikes, resto no .gitignore (recommended)")
- approved: heavy downgrade by owner (2026-10-14, "vai afetar somente esse repo correto? se sim nao precisariamos de spec correto? nao afeta o prodiuto em si")
- downgraded: heavy → standard because the change touches only this repo's local dev workflow, not the product, CI or the /work bundle; the unpushed draft spec commit was dropped

## Objective

Make the local Git hooks actually gate commits and pushes, give agents the non-derivable project rules, and clear the untracked spike leftovers from the repo root.

In scope: `lefthook.yml`, root `CLAUDE.md`, `.gitignore`, spike notes under `docs/spikes/`.
Out of scope: validator and lint in CI; the 12 lint issues already on `main`; the /work bundle contents; installing the hooks in the shared checkout (done after merge, see Outcome).

## Acceptance criteria

1. AC1: a commit whose task records fail the validator is refused and leaves no commit.
2. AC2: a commit of an unformatted Go file lands with the file `gofmt`-clean.
3. AC3: pre-push runs `go test -race ./...` and `golangci-lint` limited to issues new since `origin/main`; it passes on this branch and refuses a seeded new lint issue.
4. AC4 [inferred]: when `lefthook` can't be found, the hook refuses the commit instead of letting it through.
5. AC5: root `CLAUDE.md` is under 40 lines and holds only non-derivable rules: no piped or chained checks before commit; auto-merge disabled; subagent model tiers; `mise trust` in new worktrees; short Portuguese approvals; G-Ingá project-agnostic; /work is the bundle copy.
6. AC6: `docs/spikes/` holds both spike notes; `results.tsv`, `run.log` and `.claude/worktrees/` are ignored.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | validator; `go test -race ./...` | pass | 524f55d | |
| B2 | `golangci-lint run` | fail | 524f55d | 12 pre-existing issues: unused 5, ineffassign 5, errcheck 1, staticcheck 1 |
| B3 | hooks installed in the checkout | fail | 524f55d | no hooks in `.git/hooks`: lefthook.yml was never installed |
| AC1 | throwaway clone: commit an invalid task record | not run | | |
| AC2 | throwaway clone: commit a misformatted Go file | not run | | |
| AC3 | throwaway clone: `lefthook run pre-push` on head, then with a seeded unused func | not run | | |
| AC4 | throwaway clone: `git commit` with lefthook off PATH | not run | | |
| AC5 | CLAUDE.md read against the list; `wc -l` | not run | | |
| AC6 | `ls docs/spikes`; `git check-ignore -v` | not run | | |
| R1 | validator; `go test -race ./...` | not run | | |
| V1 | diff read against AC1–AC6 | not run | | |

## Plan

1. Complete `lefthook.yml`: fail closed, tools through `mise exec`, validator on commit, `stage_fixed` on gofmt, lint on push with `--new-from-rev`.
2. Write `CLAUDE.md`, update `.gitignore`, copy the spike notes.
3. Verify the gates in a throwaway clone, never in the shared `.git/hooks` another session is using.
4. Fresh-context review (Sonnet subagent), then PR.

## Constraints and budget

- Subagents: 1, for the fresh-context review. The owner asked for orchestration; the work is small and shares one worktree, so it runs inline.
- Hooks are not installed in the owner's checkout before merge: `.git/hooks` is shared with the worktree another session is using, and a fail-closed hook would block its commits.
