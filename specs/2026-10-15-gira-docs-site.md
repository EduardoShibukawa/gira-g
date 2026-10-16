---
schema: spec/1
status: draft
version: 1
---

# GIRA documentation website (E6)

> Implements ADR 0008 (`docs/adr/0008-human-facing-surface-and-inspection.md`)
> static-only delivery constraint (no daemon, no Node runtime in the product),
> ADR 0010 (`docs/adr/0010-repo-development-workflow-and-toolchain.md`)
> Bun/TS-for-repo-scripts convention, and the prior-art decision
> (`docs/prior-art/docusaurus.md`) that Docusaurus is permitted for the
> project's own docs website (`website/`), separate from the Go product.

## Problem

The `gira` CLI (`gira impact`, `gira map`, `gira slice`, `gira index`,
`gira serve`) has no discoverable documentation website. All prose lives in
`CONTEXT.md`, `docs/`, `specs/`, and `README.md` — internal repo files not
rendered for humans. A Change-maker or new contributor must read source to
learn what the tool does.

The product ethos (ADR 0001, prior art NORTH STAR): committed content,
LLM-free viewer, deterministic, no Ask-AI. The docs site is a build artifact
of committed Markdown — it must not introduce a runtime dependency into the
`gira` binary and must not become a daemon (ADR 0008).

Heavy triggers: H1 (product calls on tool, scope, navigation), H3 (new
subsystem: website build, navigation, search), H5 (multi-session
decomposition).

## Expected behavior

One static documentation website, built with Docusaurus, hosted as pure
static files. Content is focused on **functionality**: what `gira` does, how
to run it, what the output means. Not a mirror of internal ADRs/decisions.

### F1 — Docusaurus project scaffold

A `website/` directory with a Docusaurus v3 project: `docusaurus.config.ts`,
`package.json`, `sidebars.ts`, `docs/` source content, static assets. Build
produces a self-contained static bundle (`build/`). No server, no daemon, no
Node runtime required to view — only a static file host.

### F2 — Content migration (functionality-first)

Migrate and rewrite content into `website/docs/`:

- **Concepts** — the vocabulary from `CONTEXT.md` (Macro Map, Component,
  Service, Library, Infrastructure, Contract, Evidence Record, Target System,
  Change-maker, Impact Briefing, Impact Set, derivation tiers
  `declared`/`extracted`/`inferred`, freshness `current`/`stale`/`unverified`).
- **CLI reference** — one page per command: `gira impact`, `gira map`,
  `gira slice`, `gira index`, `gira serve`. Flags, inputs, output shape,
  examples, exit codes.
- **Guides** — quickstart (install, first index, first impact), workflow
  (Change-maker flow: brief → edit → verify), evidence inspection
  (`gira explain`, derivation tiers, source excerpts, staleness).
- **Export** — the static HTML/Markdown export surface and its limitations.

Not migrated (stay in repo, linked at most): `docs/adr/`, `docs/decisions.md`,
`docs/spikes/`, `docs/skills/`, `specs/`, `tasks/`. These are internal
working artifacts, not product documentation.

### F3 — Navigation and search

Single sidebar, ordered categories (prior art: sidebar as ordered ID list):

```
Concepts
CLI Reference
  gira map
  gira impact
  gira slice
  gira index
  gira serve
Guides
  Quickstart
  Change-maker workflow
  Evidence inspection
Export
```

Search: Docusaurus local search (no external Ask-AI, no Algolia dependency —
ADR 0001 NORTH STAR). Versioning: active, driven by Git tags `v*.*.*`
(`current` = main, `latest` = newest tag; `versions.json` + banners).

### F4 — Build and publish pipeline

- `package.json` scripts: `build`, `start` (local dev), `serve` (preview).
- mise task `docs:build` wrapping `bun install && bun run build` in `website/`.
- CI workflow: build the site on push to `main`, upload static artifact.
- No GitHub Pages deploy action in v1 (artifact only; host decision later).
- `THIRD_PARTY_NOTICES.md` updated for Docusaurus license (MIT/CC-BY).

### F5 — Quarantine compliance

Per ADR 0012 rule 3: no reference to the quarantined comparative benchmark
module anywhere in `website/`. A static check (grep) enforces this in CI.

## Scope

**In scope**

- F1–F5 above.
- `website/` Docusaurus project, content, build, mise task, CI artifact.
- License notices update.

**Out of scope**

- Changes to the Go binary (`cmd/gira`, `internal/`, `go.mod`).
- Changes to `.goreleaser.yaml`, release workflows, GoReleaser config.
- Hosting/GitHub Pages deploy (artifact only in v1).
- ADRs, decisions, specs, spikes, tasks, skills content migration.
- Search beyond local (Algolia, Ask-AI, external).
- Versioning (multiple doc versions).
- i18n.

**Non-goals**

- Interactive playground or live CLI demo.
- Blog/news section.
- Full-text semantic search.

## Acceptance criteria

- SC1 scaffold: `website/` builds clean (`bun run build` exits 0, produces
  `build/index.html` with no external asset references).
- SC2 content: every CLI command (`gira impact`, `gira map`, `gira slice`,
  `gira index`, `gira serve`) has a reference page with flags, inputs,
  output shape, and at least one example; every term in `CONTEXT.md`
  "Language" section has a concepts page.
- SC3 navigation: sidebar renders in the ordered category list above; every
  internal link resolves (no 404 in build).
- SC4 search: local search indexes all pages; search returns results for
  CLI command names and concept terms.
- SC5 build pipeline: `mise run docs:build` exits 0 from a clean checkout;
  CI workflow file exists and is valid YAML.
- SC6 quarantine: no file under `website/` references the quarantined
  comparative benchmark (grep for benchmark module identifiers returns empty).
- SC7 regression: `go test -race ./...` unchanged (no Go files touched);
  `node tools/validate/validate.js` exit 0.
- SC8 licenses: `THIRD_PARTY_NOTICES.md` has a Docusaurus entry.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| `website/` (new) | Docusaurus v3 project; Bun-managed; built to static `build/` |
| `mise.toml` | new task `docs:build` |
| `.github/workflows/` | new workflow: build docs artifact on push to main |
| `THIRD_PARTY_NOTICES.md` | new entry for Docusaurus (MIT) |
| `package.json` (new, in `website/`) | `docusaurus`, `@docusaurus/preset-classic`, search plugin |

No changes to Go module, `cmd/gira`, `internal/`, `.goreleaser.yaml`, or
release workflows.

## Decisions

- Q1: tool → **Docusaurus** (owner, 2026-10-15, "ok"). Prior art already
  studied it; patterns extracted (sidebar ordered list, versioning, faceted
  search); Bun already pinned in mise (ADR 0010).
- Q2: content scope → **functionality-first** (owner, 2026-10-15, "focando
  em funcionalidade e nao em ADRs, decisions, etc.."). CLI reference,
  concepts, guides, export. ADRs/decisions/specs/spikes/tasks stay in repo.
- Q3: navigation → **single sidebar, ordered categories** (owner, 2026-10-15,
  "Site único com categorias ordenadas (Recomendado)").
- Q4: search → **Docusaurus local search** (no Algolia, no Ask-AI — ADR 0001
  NORTH STAR).
- Q5: hosting → **CI artifact only in v1** (no GitHub Pages deploy action;
  host decision later — avoids one-way door on infra).
- Q6: versioning → **active, following release tags** (owner, 2026-10-15, "versao agora temos podemos verificar seguindo a tag da release nao?"). Docusaurus versioning plugin driven by Git tags `v*.*.*`; `current` = main, `latest` = newest tag. Matches prior-art pattern (`versions.json`, banners, `<10 active`).
- Q7: quarantine → **no benchmark references in `website/`** (ADR 0012 rule 3;
  static grep check in CI).
- Q8: license → **Docusaurus MIT entry in `THIRD_PARTY_NOTICES.md`** (ADR 0010
  licensing safeguard).

## Risks and constraints

- Docusaurus is a heavy node_modules footprint → acceptable for a build-time
  repo tool (ADR 0010 Bun/TS convention); never a Go runtime dependency.
- Content drift: repo prose and website prose can diverge → website is the
  human surface; repo `CONTEXT.md` stays the source for term definitions and
  website links back to it.
- License compliance → `THIRD_PARTY_NOTICES.md` must be updated before merge.
- Windows CI: Bun shell quirks (ADR 0010 §2) → mise task uses
  `bun install && bun run build` (no shell builtins).
- Quarantine leak → static grep check blocks merge.

## Verification

- Per child + integrated:
  - B1 `mise run docs:build` (clean checkout → exit 0, `build/index.html` exists).
  - B2 content completeness: every CLI command and every `CONTEXT.md` term has
    a page (grep/soncheck).
  - B3 CI workflow: `yamllint` or `actionlint` on the new workflow file.
  - B4 quarantine: `grep -r <benchmark-ids> website/` returns empty.
  - B5 `go test -race ./...` (no Go change — regression guard).
  - B6 `node tools/validate/validate.js` (exit 0).
- AC rows map SC1–SC8 to the checks above; `not run` forbidden at parent seal.
- **Review focus:** Is the build truly static (no server)? Does any page
  reference the quarantined module? Are all internal links resolved? Is the
  license notice correct? Any content that drifted from `CONTEXT.md` terms?

## Rollback

`git revert` per child. The `website/` directory is additive and self-contained;
removing it restores the repo to no-website state. No migration, no publication,
no external state. The CI workflow and mise task are additive and reverted
together with the content.

## Open questions

- deferred: hosting target (GitHub Pages, Netlify, S3) — artifact now, deploy
  later (one-way door avoided).
- deferred: versioning (`versions.json`, `current vs latest`) when the CLI has
  breaking releases.
- deferred:Algolia DocSearch if local search proves insufficient.

## Revisions

- v1 (2026-10-15): draft v1 with owner decisions Q1–Q8.
