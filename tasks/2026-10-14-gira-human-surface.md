---
schema: task/1
tier: heavy
status: active
spec: gira-human-surface@1
verified_at: ""
---

# Human-facing surface (E4)

## Routing and approvals

- tier: heavy (H1: product/preference calls on styling, source excerpts, staleness semantics; H3: spans cmd/gira, internal/slices, internal/domain, go.mod; H5: children, multi-session)
- confirmed: heavy by owner (2026-10-14, "Sim, heavy (Recommended)")
- decided: Q1 add charmbracelet/lipgloss (ADR 0008) with strict NO_COLOR/non-TTY plain fallback; Q2 explain source excerpt on demand, capped ±40, --no-source; Q3 staleness vs indexing commit first, HEAD when resolvable; Q4 view --no-open + CI guard; Q5 golden the plain path byte-for-byte; Q6 network available, go get allowed; Q7 prose minimal, docs site is E6
- approved: spec v1 by owner (2026-10-14, "Aprovar v1 (Recommended)")

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
| B1 | `go test -race ./...` (root) + bench + validator before children | not run | | |
| SC1 | `go test ./cmd/gira -run SCHumanPlainFallback -count=1` | not run | | |
| SC2 | `go test ./internal/slices ./cmd/gira -run SCExplainSource -count=1` | not run | | |
| SC3 | `go test ./internal/domain ./cmd/gira -run SCStaleness -count=1` | not run | | |
| SC4 | `go test ./internal/slices ./cmd/gira -run SCViewExport -count=1` | not run | | |
| SC5 | `go test ./cmd/gira -run SCBoundariesSection -count=1` | not run | | |
| SC6 | `go test -race ./...` root and bench/; validator exit 0 | not run | | |
| SC7 | EndToEnd published/perturbed R/P unchanged; bench SC15 ≥ 90%; `git diff --exit-code 21a8092 -- fixtures/ testdata/`; `"reviewed": false` | not run | | |
| AC1 | SC1–SC7 integrated | not run | | |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first.

1. `-01-plain-renderer-and-tty`: F1, F5; SC1, SC5. Styled/plain seam + TTY/NO_COLOR + boundaries block; plain goldens. Needs the lipgloss dependency.
2. `-02-explain-source`: F2; SC2. On-demand excerpt reader + `source unavailable` + `--no-source`. Depends on 01.
3. `-03-staleness`: F3; SC3. Indexed commit + stale banner. Depends on 01.
4. `-04-view-export-polish`: F4; SC4. `view --no-open`, single-file HTML guarantee. Depends on 01.
5. `-05-integrated`: F1–F5; SC6, SC7. Goldens + regression + invariants. Depends on 01–04.

SC6–SC7 at every child. Rollback: `git revert` per child. Branch linear.

## Review

## Checkpoint

- as of: 2026-10-14, branch main, head 21a8092, worktree has the E4 spike/spec draft
- done: spike + Q1–Q7 decisions; spec v1 draft written; parent todo
- next: owner spec approval, then children

## Outcome

Not started.
