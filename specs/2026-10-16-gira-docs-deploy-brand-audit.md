---
schema: spec/1
status: draft
version: 1
---

# Gira docs hosting deploy, real brand visual identity, and integrated Gira NFR audit

> Follow-up to `2026-10-16-gira-polish-finish` and
> `2026-10-16-gira-mise-install-docs`. Closes the remaining deferred
> items from E6/polish: (b) real GitHub Pages URL/baseUrl and
> deployment workflow for the Docusaurus site, (c) definitive vector
> brand assets (*Girassol* = *Gira* + musical note *G* / Sun +
> architectural graph) and cohesive theme palette, and (d) integrated
> NFR + dogfooding audit using `gira` itself (`index`, `locate`,
> `slice`, `explain`, `impact --from-diff`, `diff`, `skills check`)
> plus validator and quarantine gates.

## Problem

1. **Hosting & Deploy (frente b):** `website/docusaurus.config.ts`
   retains the placeholder `url: 'https://gira.local'` and
   `baseUrl: '/'`. `.github/workflows/docs.yml` only uploads a raw CI
   artifact (`actions/upload-artifact@v4`) without configuring GitHub
   Pages (`actions/configure-pages`, `actions/upload-pages-artifact`,
   `actions/deploy-pages`). `README.md` (`## Documentação`) does not
   link to the published documentation site.
2. **Brand & Visual Identity (frente c):** `website/static/img/logo.svg`,
   `website/static/img/favicon.svg`, and
   `website/static/img/gira-social-card.svg` are minimal placeholders
   that rely on system `<text font-family="sans-serif">` elements and a
   mismatched crimson/navy palette (`#e94560` / `#1a1a2e`), while
   `website/src/css/custom.css` still uses the stock Docusaurus Infima
   green (`#2e8550` / `#25c2a0`).
3. **Maintenance, NFR & Integrated Audit (frente d):** As the docs, CI,
   and brand assets reach production readiness, we must verify zero
   regression across `gira skills check` (embedded catalog in
   `internal/skills/catalog`), `mise run docs:build`,
   `mise exec -- go test -race ./...`, the canonical quarantine grep,
   and audit the repository's own architectural map and diff impact
   using `gira` (`index`, `locate`, `slice`, `explain`,
   `impact --from-diff`, `diff`).

Heavy triggers: H1 (hosting target URL/Pages workflow and visual brand
direction/palette), H2 (external publication configuration to GitHub
Pages), H3 (spans `website/`, `.github/workflows/`, `README.md`, and
repo-wide audit), H5 (decomposed into 3 `standard` child tasks).

## Expected behavior

### F1 — Docs site hosting & GitHub Pages deployment (frente b)

- `website/docusaurus.config.ts`:
  - `url` set to `'https://eduardoshibukawa.github.io'`
  - `baseUrl` set to `'/gira-g/'`
  - `organizationName: 'EduardoShibukawa'` and `projectName: 'gira-g'`
    preserved; no `website/static/CNAME` file created.
- `.github/workflows/docs.yml`:
  - Triggers: `push` (`branches: [main]`), `pull_request`
    (`branches: [main]`), and `workflow_dispatch`.
  - Job `docs-build`: minimal permissions (`contents: read`), runs
    `mise run docs:build`, runs the canonical case-insensitive
    quarantine compliance grep on `website/`, and uploads the Pages
    artifact via `actions/upload-pages-artifact@v3` (`path: website/build/`).
  - Job `docs-deploy`: `needs: docs-build`, runs on
    `if: github.ref == 'refs/heads/main' && (github.event_name == 'push' || github.event_name == 'workflow_dispatch')`,
    sets `permissions: { pages: write, id-token: write }`,
    `concurrency: { group: pages, cancel-in-progress: false }`,
    `environment: { name: github-pages, url: ${{ steps.deployment.outputs.page_url }} }`,
    and executes `actions/deploy-pages@v4`.
- `README.md`:
  - Adds the published documentation site URL
    (`https://eduardoshibukawa.github.io/gira-g/`) at the top of
    `## Documentação`.

### F2 — Real vector brand identity & Solar Graph palette (frente c)

- Replace placeholder SVGs with pure vector (`<path>`, `<circle>`,
  `<line>`, `<g>`, `<defs>` — zero external font/network dependencies,
  no `<text>` in `logo.svg` or `favicon.svg`, compact size `< 8 KB`
  each) implementing the **Helio-Graph** direction (*Girassol* = *Gira*
  + musical note *G* / Sun + architectural graph):
  - `website/static/img/logo.svg`: radial crown of geometric sunflower
    petals (`#f59e0b` / `#fbbf24`) around a dark core (`#0f172a`) where
    architectural graph nodes and edges form the letter **G**.
  - `website/static/img/favicon.svg`: optically simplified 16x16
    variant of the Helio-Graph mark (thicker strokes, high-contrast
    petals and central **G** node graph on a rounded dark badge) for
    crisp browser tab rendering.
  - `website/static/img/gira-social-card.svg` (`1200x630`): cohesive
    Helio-Graph illustration + vector/geometric lockup and architectural
    graph motif on `#0f172a` background with solar gold (`#f59e0b` /
    `#fbbf24`) and sky-cyan evidence accents (`#38bdf8`), zero external
    assets.
- `website/src/css/custom.css`:
  - Replace stock Infima green with the **Solar Graph** palette:
    - `:root` (light mode): accessible warm amber/solar gold primary
      scale (`--ifm-color-primary: #b45309`, shades `#92400e`,
      `#78350f`, `#d97706`, `#f59e0b`).
    - `[data-theme='dark']`: vibrant sunflower gold primary scale
      (`--ifm-color-primary: #f59e0b`, shades `#d97706`, `#b45309`,
      `#fbbf24`, `#fcd34d`).

### F3 — Integrated NFR verification & Gira dogfooding audit (frente d)

- Verify skill catalog integrity and host sync checks using `bin/gira`:
  - `bin/gira skills manifest --format json` exits `0`;
  - `bin/gira skills check --host agents`, `--host claude`, and
    `--host cursor` against installed/generated host directories or
    verified catalog contracts.
- Execute full `gira` dogfooding audit over the repository and its
  canonical dogfood target:
  - `bin/gira index`, `bin/gira locate`, `bin/gira slice --tier summary`,
    `bin/gira explain`, `bin/gira impact --from-diff`, and `bin/gira diff`
    between baseline (`origin/main`) and the integrated HEAD index,
    confirming valid JSON/Markdown envelopes and zero broken contracts.
- Ensure all NFR gates pass:
  - `mise run docs:build` exits `0` with `baseUrl: '/gira-g/'` and
    `onBrokenLinks: 'throw'`;
  - `mise exec -- go test -race ./...` exits `0`;
  - Canonical quarantine grep across `website/` returns empty (exit `0`);
  - `node tools/validate/validate.js` exits `0`.

## Scope

**In scope**

- F1–F3 above.
- `website/docusaurus.config.ts`, `.github/workflows/docs.yml`,
  `README.md`.
- `website/static/img/logo.svg`, `website/static/img/favicon.svg`,
  `website/static/img/gira-social-card.svg`,
  `website/src/css/custom.css`.
- Verification and dogfooding audit artefacts recorded in the child and
  parent task records.

**Out of scope**

- Custom DNS domain / `website/static/CNAME`.
- Changes to Go production packages under `cmd/` or `internal/`.
- Changes to `.claude/skills/work` or `.agents/skills/work` (protected
  by `manifest.json`).

**Non-goals**

- Raster image generation (`.png`, `.ico`, `.webp`) or external CDN
  fonts.

## Acceptance criteria

- SC1 hosting-config: `website/docusaurus.config.ts` sets
  `url: 'https://eduardoshibukawa.github.io'` and
  `baseUrl: '/gira-g/'`; no `https://gira.local` remains in
  `website/docusaurus.config.ts`; `README.md` links to
  `https://eduardoshibukawa.github.io/gira-g/`.
- SC2 pages-workflow: `.github/workflows/docs.yml` defines `docs-build`
  (`contents: read`, `mise run docs:build`, canonical quarantine grep,
  `actions/upload-pages-artifact@v3`) and `docs-deploy` (`needs:
  docs-build`, `pages: write`, `id-token: write`, `environment:
  github-pages`, `actions/deploy-pages@v4` on `main` push /
  `workflow_dispatch`).
- SC3 brand-assets: `website/static/img/logo.svg`,
  `website/static/img/favicon.svg`, and
  `website/static/img/gira-social-card.svg` are valid SVGs implementing
  the Helio-Graph visual identity without external dependencies, each
  `< 8 KB`, with `logo.svg` and `favicon.svg` free of `<text>` elements
  and distinct (16x16-optimized `favicon.svg`).
- SC4 theme-palette: `website/src/css/custom.css` replaces stock Infima
  green (`#2e8550` / `#25c2a0`) with the Solar Graph amber/gold palette
  (`#b45309` light / `#f59e0b` dark).
- SC5 gira-dogfood-audit: `bin/gira` successfully runs `index`,
  `locate`, `slice --tier summary`, `explain`,
  `impact --from-diff origin/main..HEAD`, `diff`, and
  `skills manifest` / `skills check` without envelope or contract
  regressions.
- SC6 nfr-gates: `mise run docs:build` exits `0` (producing
  `website/build/index.html` and `website/build/.nojekyll`),
  `mise exec -- go test -race ./...` exits `0`, canonical quarantine
  grep on `website/` returns empty, and
  `node tools/validate/validate.js` exits `0`.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| `website/docusaurus.config.ts` | `url: 'https://eduardoshibukawa.github.io'`, `baseUrl: '/gira-g/'` |
| `.github/workflows/docs.yml` | `docs-build` + `docs-deploy` GitHub Pages workflow |
| `README.md` | Published Docs Site link in `## Documentação` |
| `website/static/img/{logo,favicon,gira-social-card}.svg` | Definitive Helio-Graph vector brand assets |
| `website/src/css/custom.css` | Solar Graph primary color palette (light & dark) |

No changes to Go source files (`cmd/`, `internal/`), `go.mod`, or
protected harness skills.

## Decisions

- Q0: ceremony tier & decomposition → **heavy with 3 standard children
  and subagents enabled** (owner, 2026-10-16, "(Recommended) Confirmar
  tier: heavy com decomposição em 3 tasks filhas standard
  (01-docs-hosting, 02-brand-identity, 03-nfr-gira-audit) e subagents
  habilitados no budget").
- Q1: target URL, `baseUrl`, `CNAME`, and `README.md` →
  **`https://eduardoshibukawa.github.io` + `/gira-g/`, no `CNAME`, link
  in `README.md`** (owner, 2026-10-16, "(Recommended) GitHub Pages
  padrão do projeto: url 'https://eduardoshibukawa.github.io', baseUrl
  '/gira-g/', sem CNAME, adicionando link em README.md").
- Q2: `.github/workflows/docs.yml` deployment strategy → **two jobs
  (`docs-build` + `docs-deploy` on `main` push / `workflow_dispatch`
  with `actions/deploy-pages@v4`)** (owner, 2026-10-16, "(Recommended)
  Dois jobs (docs-build + docs-deploy em push na main /
  workflow_dispatch, com deploy-pages@v4 incondicional na main)").
- Q3: visual direction for SVGs → **Helio-Graph pure `<path>` vector
  mark with dedicated 16x16 `favicon.svg`** (owner, 2026-10-16,
  "(Recommended) Helio-Graph (coroa de pétalas de girassol + núcleo em
  grafo formando G em <path> puro) com favicon.svg dedicado otimizado
  para 16x16").
- Q4: theme color palette in `website/src/css/custom.css` → **Solar
  Graph palette (`#b45309` light / `#f59e0b` dark)** (owner,
  2026-10-16, "(Recommended) Paleta Solar Graph: atualizar custom.css
  do verde padrão Infima para Âmbar/Ouro Girassol acessível (#b45309
  light / #f59e0b dark) coeso com os SVGs").

## Risks and constraints

- Changing `baseUrl` from `'/'` to `'/gira-g/'` requires all internal
  Docusaurus links and asset paths to resolve under `/gira-g/`;
  `onBrokenLinks: 'throw'` in `mise run docs:build` enforces this at
  build time.
- GitHub Pages deployment on `main` requires GitHub Pages (`Source:
  GitHub Actions`) enabled in repository settings.
- SVGs must not contain any token matching the canonical quarantine
  regex (`bench`, `gira-bench`, `graphify`, `understand[-_ ]?anything`,
  `egonex-ai`).

## Verification

- Baseline checks before implementation:
  - B1: `node tools/validate/validate.js` (exit 0)
  - B2: `mise run docs:build` (exit 0)
  - B3: `mise exec -- go test -race ./...` (exit 0)
- Per-child and integrated verification covering SC1–SC6, R1 regression
  suite, V1 self-review, and fresh-context adversarial review per
  `.agents/skills/work/review.md`.

## Rollback

`git revert` of the child commits restores `https://gira.local`,
artifact-only `docs.yml`, and placeholder SVGs/CSS without data loss or
schema migration.

## Open questions

None.

## Revisions

- v1 (2026-10-16): initial specification incorporating owner decisions
  Q0–Q4.
