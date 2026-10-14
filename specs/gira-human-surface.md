---
schema: spec/1
status: draft
version: 1
---

# GIRA human-facing surface (E4)

> Implements ADR 0008 (`docs/adr/0008-human-facing-surface-and-inspection.md`)
> for the Change-maker. Scope inferred from commitments other specs hand to E4
> (`specs/gira-envelopes-hygiene.md:73,82`).

## Problem

E3 sealed at 21a8092: envelopes on all agent-facing output, honest budgets,
structured errors, one tool registry, CI hygiene. The human surface ADR 0008
decided is only partially built:

1. The CLI prints plain text; there is no `charmbracelet/lipgloss` styling and
   no TTY/`NO_COLOR` detection (ADR 0008 §1).
2. `gira explain` renders the `EvidenceReport` but never fetches the source
   excerpt its locator points at (ADR 0008 §3).
3. The indexed commit is never shown and nothing is flagged `stale` at the
   command level (ADR 0008 §4).
4. `gira view` opens a browser unconditionally; there is no `--no-open`, and
   the single-file guarantee of the HTML export is implicit (ADR 0008 §1/§4).
5. No dedicated "Unsupported Boundaries & Known Unknowns" section across the
   human commands (ADR 0008 §4).

Heavy triggers: H1 (product calls on styling, excerpts, staleness semantics),
H3 (spans cmd/gira, internal/slices, internal/domain, go.mod), H5 (children).

## Expected behavior

One epic, five additive features.

### F1 — Rich terminal with a plain fallback

Human commands render with `lipgloss` on a TTY, and fall back to a
byte-deterministic **plain** rendering when stdout is not a TTY or `NO_COLOR`
is set (pipes, CI, `grep`). Styling never leaks into the plain path.

### F2 — `explain` source excerpts

`gira explain <node|--edge>` gains an on-demand source excerpt: read the cited
`[repo@commit:path#Lstart-Lend]` from the local checkout, show a capped window
(default ±40 lines), and report `source unavailable` cleanly when the checkout
or range is missing. `--no-source` disables it. Per ADR 0001 the index keeps
locators only; excerpts are read at render time and never persisted.

### F3 — Indexed commit and staleness

Every human command prints the indexed commit SHA; components/edges whose
locators cite a commit other than the indexing commit are flagged `stale`, with
a one-line banner. When the target repo path is resolvable, the local HEAD is
compared too.

### F4 — Export/view polish

`gira view` gains `--no-open` (and skips the browser in CI/headless);
`gira export --format html` is guaranteed a single self-contained file (no
external assets, no server); `md` and `html` share the same content.

### F5 — Boundaries section

A consistent "Unsupported Boundaries & Known Unknowns" block in the human
commands that can carry it (slice, impact, export, view).

## Scope

**In scope**

- F1–F5 above, each additive (no agent-output contract changes; envelopes and
  `--format json` untouched).
- Golden files for the plain human surface.

**Out of scope**

- E5 skills distribution / release, E6 docs site.
- A web server/daemon or a full-screen TUI (ADR 0008 rejects both).
- Ground-truth changes, `reviewed: true`.
- Any model client (ADR 0001), compression, semantic harvest.

**Non-goals**

- Raising recall/precision (E2), envelopes (E3), release/signing (E5),
  documentation prose (E6).

## Acceptance criteria

- SC1 plain fallback: `go test ./cmd/gira -run SCHumanPlainFallback -count=1`
  (non-TTY and `NO_COLOR` outputs are byte-identical to each other and to a
  golden; no ANSI escapes).
- SC2 explain source: `go test ./internal/slices ./cmd/gira -run SCExplainSource -count=1`
  (excerpt window honored, `--no-source`), and `source unavailable` on a missing
  checkout; the index carries no source bodies (ADR 0001).
- SC3 staleness: `go test ./internal/domain ./cmd/gira -run SCStaleness -count=1`
  (indexed commit printed; locator/HEAD mismatch flags `stale`).
- SC4 view/export: `go test ./internal/slices ./cmd/gira -run SCViewExport -count=1`
  (`--no-open` writes but does not open; HTML has no external asset reference).
- SC5 boundaries: `go test ./cmd/gira -run SCBoundariesSection -count=1`
  (the block appears with the unsupported entries).
- SC6 regression: `go test -race ./...` root and bench/; validator exit 0;
  plain goldens stable.
- SC7 I3 invariants: EndToEnd published/perturbed R/P unchanged; bench SC15 ≥
  90%; `git diff --exit-code 21a8092 -- fixtures/ testdata/`; `"reviewed": false`.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| CLI | `explain --source/--no-source`, `view --no-open`; styled TTY + plain fallback |
| internal/slices | plain renderer (existing) + styled renderer seam; source excerpt reader |
| internal/domain | indexed-commit accessor and staleness helper |
| go.mod | new: `github.com/charmbracelet/lipgloss` |

## Decisions

- Q1: add `charmbracelet/lipgloss` (ADR 0008), with a strict `NO_COLOR`/non-TTY plain fallback.
- Q2: `explain` fetches a capped source excerpt on demand (default ±40 lines), `--no-source` to disable.
- Q3: staleness compares locators against the indexing commit first; the local HEAD only when the repo resolves.
- Q4: `view` keeps auto-open but gains `--no-open` and skips the browser in CI/headless.
- Q5: the plain (non-TTY) human output is golden-tested byte for byte; styling must not leak into it.
- Q6: network is available; `go get` lipgloss is allowed.
- Q7: prose stays minimal; the docs site is E6.

## Risks and constraints

- New compile-time dependency vs the single-binary spirit (ADR 0008 allows it); supply-chain review to be recorded.
- Styling churn → the plain path must stay byte-identical (goldens).
- Source excerpts must not enter the index (ADR 0001) and must never persist.
- `view` auto-open in CI → `--no-open` + env guard.

## Verification

- Per child + integrated: B1 `go test -race ./...` (root); B2 `go test -race ./...` (bench/); B3 `go test ./internal/eval -run EndToEnd -v` (published/perturbed R/P); B4 the SC commands above; B5 `node tools/validate/validate.js`.
- AC rows map SC1–SC7 to the `go test` names above; `not run` forbidden at parent seal.
- **Review focus:** is the plain path byte-identical (no ANSI)? Does the excerpt read only locators and never persist? Is staleness honest (no false `stale`/`current`)? Does `--no-open` really not open? Any weakened assertions or named guesses?

## Rollback

`git revert` per child. No migration, publication or external state. Additive;
reverting restores the plain-only behavior.

## Open questions

- deferred: adopt the styled output in the MCP/agent JSON? No — agent output keeps `--format json`; styling is human-only (safe to decide later).
- deferred: richer staleness policy (per-locator commit vs index commit) if drift is common; additive.

## Revisions

- v1 (2026-10-14): draft v1 with owner decisions Q1–Q7.
