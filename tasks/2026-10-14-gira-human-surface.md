---
schema: task/1
tier: heavy
status: accepted
spec: gira-human-surface@1
verified_at: "47e79a8"
---

# Human-facing surface (E4)

## Routing and approvals

- tier: heavy (H1: product/preference calls on styling, source excerpts, staleness semantics; H3: spans cmd/gira, internal/slices, internal/domain, go.mod; H5: children, multi-session)
- confirmed: heavy by owner (2026-10-14, "Sim, heavy (Recommended)")
- decided: Q1 add charmbracelet/lipgloss (ADR 0008) with strict NO_COLOR/non-TTY plain fallback; Q2 explain source excerpt on demand, capped ±40, --no-source; Q3 staleness vs indexing commit first, HEAD when resolvable; Q4 view --no-open + CI guard; Q5 golden the plain path byte-for-byte; Q6 network available, go get allowed; Q7 prose minimal, docs site is E6
- approved: spec v1 by owner (2026-10-14, "Aprovar v1 (Recommended)")
- accepted: by owner (2026-10-14, "Mergear e aceitar (Recommended)")

## Objective

Implement ADR 0008 for the Change-maker: a rich terminal surface with a
byte-deterministic plain fallback, on-demand source excerpts in `explain`, the
indexed commit plus honest staleness flags, a safe `view`/`export`, and a
consistent Unsupported Boundaries section — all additive, without changing the
agent (`--format json`) contract.

In scope:
- spec `gira-human-surface` v1;
- the children listed in the Plan;
- integrated verification of SC1–SC7.

Out of scope:
- E5 skills distribution/release, E6 docs site;
- a web server/daemon or a TUI (ADR 0008 rejects both);
- ground-truth changes, `reviewed: true`;
- any model client (ADR 0001), compression/summarization;
- push or release.

## Acceptance criteria

- AC1 = SC1–SC7 of `specs/gira-human-surface.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before children | pass | 21a8092 | E3 merged (spec v2); green baseline |
| B2 | all four children green plus the E4 fix pass | pass | 79fdf1f | E4 setup merge; green before the fix pass |
| SC1 | `go test ./cmd/gira -run SCHumanPlainFallback -count=1` | pass | 47e79a8 | non-TTY == NO_COLOR, no ANSI, golden stable |
| SC2 | `go test ./internal/slices ./cmd/gira -run SCExplainSource -count=1` | pass | 47e79a8 | ±40 window, `--no-source`, unavailable, unrelated checkout rejected, symlink rejected |
| SC3 | `go test ./internal/domain ./cmd/gira -run SCStaleness -count=1` | pass | 47e79a8 | indexed commits listed; `--checkout` HEAD mismatch flags `stale`; no cwd guess |
| SC4 | `go test ./internal/slices ./cmd/gira -run SCViewExport -count=1` | pass | 47e79a8 | `--no-open` does not open; HTML single-file |
| SC5 | `go test ./cmd/gira -run SCBoundariesSection -count=1` | pass | 47e79a8 | Unsupported Boundaries block present |
| SC6 | `go test -race ./...` root and bench/; validator exit 0 | pass | 47e79a8 | root + bench green; validator pass |
| SC7 | EndToEnd published/perturbed R/P unchanged; bench SC15 ≥ 90%; `git diff --exit-code 21a8092 -- fixtures/ testdata/`; `"reviewed": false` | pass | 47e79a8 | 82.4/100.0 and 81.2/100.0; SC15 ok; diff 0; reviewed false |
| AC1 | SC1–SC7 integrated | pass | 47e79a8 | all rows pass at the integrated head |
| R1 | `go test -race ./...` (root) | pass | 47e79a8 | all packages green |
| R2 | `go test -race ./...` (bench/) | pass | 47e79a8 | layer0/1/2 + quarantine green |
| R3 | `node tools/validate/validate.js` | pass | 47e79a8 | manifest-integrity + artifact-records pass |
| V1 | integrated diff read + fresh-context adversarial review | pass | 47e79a8 | findings F1–F9 below, all fixed |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-plain-renderer-and-tty`: F1, F5; SC1, SC5. Styled/plain seam + TTY/NO_COLOR + boundaries block; plain goldens. Needs the lipgloss dependency.
2. `-02-explain-source`: F2; SC2. On-demand excerpt reader + `source unavailable` + `--no-source`. Depends on 01.
3. `-03-staleness`: F3; SC3. Indexed commit + stale banner. Depends on 01.
4. `-04-view-export-polish`: F4; SC4. `view --no-open`, single-file HTML guarantee. Depends on 01.
5. `-05-integrated`: F1–F5; SC6, SC7. Goldens + regression + invariants. Depends on 01–04.

SC6–SC7 at every child. Rollback: `git revert` per child. Branch linear.

## Review

- [F1] a default `--repo .` flagged the cwd as `stale` on the common path -> fixed in 47e79a8 (staleness is `--checkout` only, no default; banner is Markdown-only)
- [F2] `Excerpt` ignored the locator's repo/commit, so a same-named file in an unrelated checkout was shown as evidence -> fixed in 47e79a8 (`--repo` must be a Git checkout that contains the cited commit, else `source unavailable`)
- [F3] the path check was lexical, so a symlink escaped the checkout (arbitrary file read) -> fixed in 47e79a8 (`EvalSymlinks` containment + regular-file requirement)
- [F4] a single-point locator (`EndLine == 0`) showed only the window's first line and dropped the cited line -> fixed in 47e79a8 (`end = StartLine`)
- [F5] the whole file was read and a >1 MB line returned a spurious `source unavailable` -> fixed in 47e79a8 (bounded read, no scanner line cap)
- [F6] only three commands printed the indexed commit -> fixed in 47e79a8 (added export, view, tour, flows, locate, diff)
- [F7] JSON output gained banner lines on stderr -> fixed in 47e79a8 (banner is Markdown-only)
- [F8] the spec interface lists `explain --source/--no-source` but only `--no-source` existed -> fixed in 47e79a8 (added `--source`, default true)
- [F9] child 05 and the parent seal were unrecorded -> fixed in 47e79a8 (child 05 verified; parent sealed with SC1–SC7 evidence)
- note: the per-component stale flag (F3 spec line 55) is realized as the indexed-commit banner plus the existing per-evidence status; an explicit per-node flag is deferred as additive

## Checkpoint

- as of: 2026-10-14, branch e4-human-surface-impl, head 47e79a8, worktree clean
- done: children 01–05 verified; SC1–SC7 + R1–R3 + V1 pass at 47e79a8; review F1–F9 fixed
- next: none — accepted; E4 closed

## Outcome

E4 human-facing surface delivered at spec v1: styled/plain seam with a
byte-identical plain fallback (lipgloss), `explain` source excerpts, indexed
commit + honest staleness, `view --no-open` with a single-file HTML guarantee,
and the Unsupported Boundaries block. SC1–SC7 verified; the fresh-context
adversarial review F1–F9 fixed. Awaiting owner acceptance.
