# Spike: E4 — Human-facing surface (read-only, unapproved)

> Read-only spike. No spec, no task records, no code. Scope inferred from
> ADR 0008 (`docs/adr/0008-human-facing-surface-and-inspection.md`) and the
> commits other specs hand to E4 (`specs/gira-envelopes-hygiene.md:73,82`).

## Problem

ADR 0008 decides the human surface for the Change-maker: a rich terminal CLI
plus self-contained static exports, evidence inspection down to git locators,
honest boundaries, and declarative overlays. What exists today is a subset:

1. **Rich terminal CLI is not there.** The CLI prints plain text via
   `fmt.Fprintf`; there is **no** `charmbracelet/lipgloss` (ADR 0008 §1) and no
   TTY/`NO_COLOR` detection. Tables, trees and token-budgeted Markdown are
   plain.
2. **`gira explain` does not retrieve source.** ADR 0008 §3 says it fetches
   source excerpts on demand from the local git checkout, or reports "source
   unavailable". Today `runExplain` renders only the `EvidenceReport`
   (origin/profile/status/locator).
3. **Staleness is not surfaced at the index level.** ADR 0008 §4 wants the
   indexed commit SHA shown and components/edges marked `stale` when the local
   HEAD differs. Today only per-evidence `status` is rendered; the CLI never
   prints the indexed commit nor a stale banner.
4. **`gira view`/`export` gaps.** `view` writes `index.html` and opens a
   browser unconditionally (`openInBrowser`), with no `--no-open` for
   CI/headless. `export --format html|md` exists; single-file/zero-dependency
   guarantees and boundaries visibility are only implicit.
5. **Unsupported boundaries** are rendered in slice/export Markdown but there is
   no dedicated "Unsupported Boundaries & Known Unknowns" section across the
   human commands, as ADR 0008 §4 promises.

Heavy triggers: H1 (product/preference calls: styling dependency, source
excerpts, staleness semantics), H3 (spans `cmd/gira`, `internal/slices`,
`internal/domain`, new dependency), H5 (multi-session, children).

## Expected behavior (proposed F1–F5)

### F1 — Rich terminal with a plain fallback
Styled tables/trees/Markdown for TTY output; a **byte-deterministic plain
fallback** when stdout is not a TTY or `NO_COLOR` is set (pipable, CI-safe).

### F2 — `explain` source excerpts
`gira explain <node|--edge>` gains an opt-in source excerpt: read the cited
`[repo@commit:path#Lstart-Lend]` from the local checkout, show a capped window,
and report `source unavailable` cleanly (ADR 0001: locators only in the index).

### F3 — Indexed commit and staleness
Every human command prints the indexed commit SHA; components/edges whose
locators cite a commit other than the indexing commit (or whose local HEAD
differs) are flagged `stale`, with a one-line banner.

### F4 — Export/view polish
`view` gains `--no-open`; `export --format html` stays a single self-contained
file (no external assets, no server); both include the boundaries section.

### F5 — Boundaries section
A consistent "Unsupported Boundaries & Known Unknowns" block in the human
commands that can carry it (slice, impact, export, view).

## Scope

**In scope:** F1–F5 above, additive; deterministic plain output; golden files
for the human surface.

**Out of scope:** E5 skills distribution/release, E6 docs site; a web server or
daemon (ADR 0008 rejects it); a TUI (ADR 0008 rejects it); ground-truth changes;
`reviewed: true`; any model client (ADR 0001).

**Non-goals:** raising recall/precision (E2), envelopes (E3), release/signing
(E5), documentation site (E6).

## Interfaces and dependencies

| Interface | Change |
|---|---|
| CLI | `explain --source/--no-source`, `view --no-open`, styling with plain fallback |
| internal/slices | plain renderer + (optional) styled renderer; source excerpt reader |
| internal/domain | indexed-commit accessor / staleness helper |
| go.mod | **new**: `charmbracelet/lipgloss` (ADR 0008) — network to fetch |

## Open questions (owner)

- **Q1 Rich terminal dependency.** ADR 0008 names `charmbracelet/lipgloss`. Add
  it now (pulls transitive deps; needs network), or keep the CLI
  dependency-free plain text and defer styling? *Recommend: add lipgloss with a
  strict `NO_COLOR`/non-TTY plain fallback — ADR 0008 already decided it.*
- **Q2 Source excerpts.** Fetch on demand from the local checkout, capped
  window (default ±40 lines), `--no-source` to disable? *Recommend: yes, capped,
  opt-out.*
- **Q3 Staleness semantics.** Compare locators against the indexing commit only,
  or also against the local repo HEAD when a repo path is known? *Recommend:
  indexing commit first; HEAD comparison only when the repo is resolvable.*
- **Q4 `view` browser.** Keep auto-open but add `--no-open`, and skip open when
  `CI`/headless? *Recommend: yes.*
- **Q5 Determinism.** Golden-test the plain (non-TTY) human output byte for
  byte? *Recommend: yes; styling must not leak into the plain path.*
- **Q6 Network.** Is `go get charmbracelet/lipgloss` allowed in this
  environment/repo policy? If not, Q1 collapses to plain-only for E4.
- **Q7 Boundary vs E6.** Keep styling hints/linkouts minimal so E6 (docs site)
  owns prose docs? *Recommend: yes.*

## Risks

- New dependency vs the "single binary, no runtime deps" spirit (ADR 0008
  allows compile-time deps) and supply chain (ADR 0012 quarantine is unrelated).
- Styling churn → plain path must stay byte-identical (goldens).
- Source excerpts read the local checkout → ensure no source bodies enter the
  index (ADR 0001) and excerpts never persist.
- `view` auto-open surprises CI → `--no-open` + env guard.

## Plan (children, after spec approval)

1. `-01-plain-renderer-and-tty` (F1, F5): plain/styled seam + TTY/NO_COLOR +
   boundaries section, goldens.
2. `-02-explain-source` (F2): on-demand excerpts + `source unavailable`.
3. `-03-staleness` (F3): indexed commit + stale banner.
4. `-04-view-export-polish` (F4): `--no-open`, single-file guarantees.
5. `-05-integrated` (F1–F5): goldens + regression.

## Revisions

- v1 (2026-10-14): read-only spike draft, unapproved.
