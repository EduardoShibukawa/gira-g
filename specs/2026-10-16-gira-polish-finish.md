---
schema: spec/1
status: draft
version: 1
---

# Gira polish finish (docs correctness + identity + CI hygiene)

> Follow-up to E6 docs site (`specs/2026-10-15-gira-docs-site.md`, accepted).
> E6 shipped scaffold, functionality-first content, ordered sidebar, local
> search, build pipeline, quarantine gate, licenses. Subagent audits
> (2026-10-16) found ~30 residual gaps: wrong CLI flag examples, missing
> exit codes, partial CLI surface, Export nav duplication, placeholder
> site identity, unversioned docs, narrow quarantine grep, non-frozen
> docs install, THIRD_PARTY gaps.

## Problem

The docs site builds clean but misleads on real CLI usage and leaves
CI/quarantine gaps:

- `website/docs/intro.md` shows `gira map` (no such command; builder is
  `gira index`), `gira impact --service` (no `--service`; positional
  `target_id`), `gira slice --contract` (no `--contract`; positional
  `node_id` + `--tier/--view`). `website/docs/cli/impact.md` repeats a
  `--contract` example. No CLI page documents exit codes (`0` ok,
  `1` usage/load/emit/flag error; `skills check` out-of-sync `1`,
  `skills nudge` always `0`).
- Real commands `explain`, `tour`, `diff`, `flows`, `locate`, `skills`
  (and `view`, folded into export) have no `website/docs/cli/*.md`
  reference page; only guides cover them. Spec SC2 "every CLI command
  has a reference page" fails on the extended surface.
- `website/sidebars.ts` lists `guides/export` inside Guides plus a
  top-level `Export` pointing at `cli/export`: duplicate Export label.
- `website/docusaurus.config.ts` keeps placeholder `url`
  `https://gira.local`, `organizationName: 'gira-g'`,
  `projectName: 'gira-docs'` (remote is `EduardoShibukawa/gira-g`);
  local-search plugin lacks `indexBlog: false` tuning.
- No `versions.json` / `versioned_docs/` snapshot for released
  `v0.1.0`; no `website/static/.nojekyll`.
- `.github/workflows/docs.yml` quarantine grep is case-sensitive,
  misses `gira-bench`, `egonex-ai`, standalone `bench`,
  `understand_anything` / `understand anything`, and scans only a
  subset of `website/`. Canonical pattern lives in
  `bench/quarantine/check.go` (case-insensitive).
- `mise.toml` `docs:build` runs bare `bun install` (no
  `--frozen-lockfile`); `THIRD_PARTY_NOTICES.md` misses `clsx`,
  `react-dom`, `typescript`, `@docusaurus/*` type packages.

Heavy triggers: H1 (hosting URL, versioning timing, CLI scope,
org/project identity), H3 (website + CI + notices touch three areas),
H5 (three children across sessions).

## Expected behavior

### F1 — Docs correctness + full CLI surface

- `intro.md` examples use real invocations: `gira index`,
  `gira impact <target_id>`, `gira slice <node_id> [--tier …] [--view …]`.
- `cli/impact.md` `--contract` example becomes positional.
- New reference pages: `cli/explain.md`, `cli/tour.md`, `cli/diff.md`,
  `cli/flows.md`, `cli/locate.md`, `cli/skills.md` (view stays folded
  into `cli/export.md` + `guides/export.md` cross-link kept).
- Every `cli/*.md` documents exit codes (`0`/`1` + special cases).
- `cli/serve.md` dead `[ADR 0004](#)` link removed (ADRs stay in repo).
- `cli/map.md` stays as explicit alias placeholder (builder is
  `gira index`; `gira map` exits `1` unknown command).
- Sidebar: single top-level `Export` (`cli/export`); `guides/export`
  removed from Guides items (file kept, linked from `cli/export.md`).
- Search plugin sets `indexBlog: false` (blog disabled).

### F2 — Site identity + version snapshot

- `url` stays `https://gira.local` (artifact-only, host decision later).
- `organizationName` → `EduardoShibukawa`, `projectName` → `gira-g`.
- `website/static/.nojekyll` added. `CNAME` stays deferred.
- `docusaurus docs:version 0.1.0` snapshot: `versions.json` +
  `versioned_docs/0.1.0/`; `current` = main, `latest` = `0.1.0`.

### F3 — CI hygiene + licenses

- `docs.yml` quarantine grep uses canonical case-insensitive pattern
  covering `bench`, `gira-bench`, `graphify`,
  `understand[-_ ]?anything`, `egonex-ai`, scoped to full `website/`
  inputs (`docs/`, `src/`, `static/`, `*.ts`, `package.json`).
- `mise.toml` `docs:build` uses `bun install --frozen-lockfile`.
- `THIRD_PARTY_NOTICES.md` adds `clsx`, `react-dom`, `typescript`,
  `@docusaurus/module-type-aliases`, `@docusaurus/tsconfig`,
  `@docusaurus/types`.

## Scope

**In scope**

- F1–F3 above.
- `website/docs/**`, `website/sidebars.ts`,
  `website/docusaurus.config.ts`, `website/static/.nojekyll`,
  `versions.json` + `versioned_docs/0.1.0/`.
- `.github/workflows/docs.yml` grep, `mise.toml` `docs:build`,
  `THIRD_PARTY_NOTICES.md`.

**Out of scope**

- Hosting deploy / GitHub Pages action / `CNAME` (artifact only).
- Real brand assets beyond landed SVG placeholders.
- Go behavior changes (`cmd/`, `internal/`, `go.mod`).
- `.goreleaser.yaml`, release workflows, GoReleaser config.
- i18n, Algolia, blog, semantic search.

**Non-goals**

- Interactive playground, live CLI demo.
- Backfilling versions older than `0.1.0`.

## Acceptance criteria

- SC1 docs-correctness: `intro.md` + `cli/impact.md` use real flags;
  `grep -r -- --service website/docs/cli website/docs/intro.md`
  and `--contract` return empty; `serve.md` has no `ADR 0004` link.
- SC2 full-surface: `cli/explain.md`, `cli/tour.md`, `cli/diff.md`,
  `cli/flows.md`, `cli/locate.md`, `cli/skills.md` exist with flags,
  inputs, output shape, one example, exit codes; every `cli/*.md`
  documents exit codes.
- SC3 navigation: single `Export` entry in sidebar; `mise run
  docs:build` exits 0 with `onBrokenLinks: throw`; search plugin has
  `indexBlog: false`.
- SC4 identity+version: `organizationName`/`projectName` match real
  coordinates; `url` stays placeholder; `website/static/.nojekyll`
  ships in `build/`; `versions.json` contains `0.1.0`.
- SC5 CI-hygiene: `docs.yml` grep is case-insensitive, matches
  canonical alternatives, scans full `website/`; `docs:build` uses
  `--frozen-lockfile`; `THIRD_PARTY_NOTICES.md` lists the six missing
  packages.
- SC6 regression: `go test -race ./...` unchanged (no Go files
  touched); `node tools/validate/validate.js` exit 0; website
  quarantine grep returns empty.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| `website/docs/cli/*` (6 new pages, 3 fixed) | real flags + exit codes |
| `website/sidebars.ts` | dedup Export, register 6 new pages |
| `website/docusaurus.config.ts` | org/project align, search tuning |
| `website/static/.nojekyll`, `versions.json`, `versioned_docs/` | new |
| `.github/workflows/docs.yml` | widened quarantine grep |
| `mise.toml` | `docs:build` frozen lockfile |
| `THIRD_PARTY_NOTICES.md` | six entries added |

No Go module, `cmd/gira`, `internal/`, `.goreleaser.yaml`, or
release workflow changes.

## Decisions

- Q0: tier → **heavy fatiado** (owner, 2026-10-16, "Heavy fatiado
  (Recomendado)"). H1+H3+H5 hold; spec + 3 standard children.
- Q1: hosting URL → **manter `https://gira.local`** (owner,
  2026-10-16, "Manter placeholder (Recomendado)"). Artifact-only;
  avoids infra one-way door; coherent with E6 Q5.
- Q2: org/project → **alinhar real** (`EduardoShibukawa` /
  `gira-g`) (owner, 2026-10-16, "Alinhar real (Recomendado)").
  Avoids future baseUrl breakage.
- Q3: versionamento → **snapshot `0.1.0` agora** (owner,
  2026-10-16, "Snapshot v0.1.0 (Recomendado)"). Fulfills E6 Q6;
  reversible via revert.
- Q4: escopo CLI → **full surface** (owner, 2026-10-16,
  "Full surface (Recomendado)"). Closes SC2 on the extended surface.
- Q5: extras → **`.nojekyll` agora, `CNAME` deferred** (owner,
  2026-10-16, ".nojekyll agora (Recomendado)").
- Rejected: single giant commit; rejected: github.io URL now (needs
  deploy decision); rejected: keeping 5-command scope (leaves SC2 gap).

## Risks and constraints

- `docusaurus docs:version` rewrites `sidebars`/`docs` into
  `versioned_*`; keep diff reviewable, run full build per child.
- Version snapshot is additive and revertable; no Pages deploy, no
  external state.
- Quarantine widening must not flag benign `benchmarks` prose
  (`reference-corpus.md`); canonical `bench([^a-z0-9_-]|$)` excludes it.
- `bun install --frozen-lockfile` fails if `bun.lock` drifts; update
  lockfile explicitly if deps change (none planned).
- Windows CI: no shell builtins in mise tasks.

## Verification

- Per child + integrated:
  - B1 `node tools/validate/validate.js` before edits (exit 0).
  - B2 `mise run docs:build` before edits (exit 0).
  - SC1–SC2: grep checks + per-page flag/exit-code review.
  - SC3: build with `onBrokenLinks: throw` + sidebar read.
  - SC4: `versions.json` + `build/.nojekyll` presence.
  - SC5: `grep` pattern review + `THIRD_PARTY` entry check.
  - SC6: `go test -race ./...` + validator + quarantine grep empty.
- AC rows map SC1–SC6; `not run` forbidden at parent seal.
- **Review focus:** Do examples match `cmd/gira/main.go` flags? Any
  invented flag? Any quarantined token in `website/`? Does the
  versioned build resolve all links?

## Rollback

`git revert` per child. `website/` + `versions.json` +
`versioned_docs/` are additive; CI/notices edits are text-only. No
migration, no publication, no external state.

## Open questions

None. Hosting target + `CNAME` + real brand assets deferred (safe:
artifact-only, placeholders ship).

## Revisions

- v1 (2026-10-16): initial draft with owner decisions Q0–Q5.
