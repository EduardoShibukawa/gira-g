---
schema: task/1
tier: standard
status: verified
verified_at: "40ff6fa"
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
| AC1 | throwaway clone: commit an invalid task record | pass | 40ff6fa | commit refused by validate (status "bogus"); HEAD unchanged |
| AC2 | throwaway clone: commit a misformatted Go file | pass | 40ff6fa | probe in internal/archdiff/archdiff.go committed gofmt-clean, worktree clean |
| AC3 | throwaway clone: `lefthook run pre-push` on head, then with a seeded unused func | pass | 40ff6fa | pre-push rc=0 on head; rc=1 with seeded `gateUnusedProbe` (unused), tests passed |
| AC4 | throwaway clone: `git commit` with lefthook off PATH | pass | 40ff6fa | PATH=/usr/bin:/bin: "Can't find lefthook in PATH", commit aborted |
| AC5 | CLAUDE.md read against the list; `wc -l` | pass | 40ff6fa | 33 lines; each rule checked against the list |
| AC6 | `ls docs/spikes`; `git check-ignore -v` | pass | 40ff6fa | both notes in docs/spikes; `git check-ignore` matches all three |
| R1 | validator; `go test -race ./...` | pass | 40ff6fa | validator pass; `go test -race ./...` pass inside AC3 pre-push run |
| V1 | diff read against AC1–AC6 | pass | 40ff6fa | diff 524f55d..40ff6fa read against AC1–AC6 |

## Plan

1. Complete `lefthook.yml`: fail closed, tools through `mise exec`, validator on commit, `stage_fixed` on gofmt, lint on push with `--new-from-rev`.
2. Write `CLAUDE.md`, update `.gitignore`, copy the spike notes.
3. Verify the gates in a throwaway clone, never in the shared `.git/hooks` another session is using.
4. Fresh-context review (Sonnet subagent), then PR.

## Constraints and budget

- Subagents: 1, for the fresh-context review. The owner asked for orchestration; the work is small and shares one worktree, so it runs inline.
- Hooks are not installed in the owner's checkout before merge: `.git/hooks` is shared with the worktree another session is using, and a fail-closed hook would block its commits.

## Review

Fresh-context review (Sonnet subagent) of 524f55d..40ff6fa: no critical or major findings.

- [F1] validator reads the working tree, not the index; an invalid record staged with an unstaged fix can land -> rejected: validator behavior is out of scope; ordinary `add`/`commit -a` flows are covered and CI enforcement is the follow-up
- [F2] `stage_fixed` re-stages whole Go files, pulling in unstaged hunks of a partially staged file -> rejected: lefthook's documented behavior; partial staging of Go files isn't part of this repo's workflow
- [F3] lint gate needs `origin/main` (fails closed when absent) and lints nothing when pushing `main` itself -> rejected: fails closed by design, and `main` changes only through owner-merged PRs
- [F4] Verification rows still `not run` -> fixed in this record update
- [F5] first log entry says `tier: heavy` while frontmatter says standard -> rejected: §2.3 grammar; the current tier is the target of the last `downgraded:` entry

## Outcome

Delivered: fail-closed lefthook gates (validator and re-staged gofmt on commit; race tests and new-issue lint on push), a 33-line root CLAUDE.md, spike notes under docs/spikes, ignore rules for run outputs and agent worktrees.

Follow-ups:
- After merge, in the owner's checkout: `mise exec -- lefthook install`, then remove the root `spike-*.md` originals (owner go-ahead).
- Run the validator and lint gate in CI so `--no-verify` can't bypass them.
- The 12 lint issues on `main`.
