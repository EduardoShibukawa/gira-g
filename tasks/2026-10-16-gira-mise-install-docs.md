---
schema: task/1
tier: standard
status: verified
verified_at: "bd47f59"
---

# Mise installation path + README and Quickstart alignment

## Routing and approvals

- tier: standard (S1 settled via owner Q1–Q3; S3 spans README.md + website/docs/guides/quickstart.md + versioned_docs/version-0.1.0/guides/quickstart.md; S2, S4, S6, S7, S8 hold, !S5)
- decided: focus and tier → Foco (a) em tier standard (instalação via mise + alinhar README/Quickstart, mantendo Pages e brand deferidos) by owner (2026-10-16, "(Recommended) Foco (a) em tier standard: instalação via mise (backend github:EduardoShibukawa/gira-g) + alinhar README/Quickstart, mantendo Pages e brand deferidos")
- decided: mise backend → usar apenas github:EduardoShibukawa/gira-g (ubi: depreciado no mise) by owner (2026-10-16, "(Recommended) Usar apenas `github:EduardoShibukawa/gira-g` (nativo do mise; `ubi:` emite warning de depreciação para remoção no mise 2027.1.0)")
- decided: file scope → atualizar README.md (Instalação + trocar gira map residual em CLI) + website/docs/guides/quickstart.md + snapshot version-0.1.0/guides/quickstart.md by owner (2026-10-16, "(Recommended) Atualizar README.md (seção Instalação + trocar `gira map` residual por `gira index` em CLI) + website/docs/guides/quickstart.md + snapshot version-0.1.0/guides/quickstart.md")

## Objective

Complete the deferred `mise` installation path unlocked by `v0.1.0` release assets and align `README.md` and the Quickstart guides (current + `0.1.0` snapshot) with real installation and CLI invocations.

In scope:
- Verify and document `mise` installation via `github:EduardoShibukawa/gira-g@0.1.0` (with `GITHUB_TOKEN`) alongside `install.sh` in `README.md`, `website/docs/guides/quickstart.md`, and `website/versioned_docs/version-0.1.0/guides/quickstart.md`;
- Remove the residual `gira map` line from `README.md` `## CLI`.

Out of scope:
- GitHub Pages deployment, `CNAME`, or changing `https://gira.local`;
- Real brand SVG assets;
- Go code, `.goreleaser.yaml`, or CI workflow modifications.

## Acceptance criteria

- AC1 = `mise exec github:EduardoShibukawa/gira-g@0.1.0 -- gira version` installs from the published `v0.1.0` release asset and outputs `gira 0.1.0`.
- AC2 = `README.md` documents `mise` (`github:EduardoShibukawa/gira-g@0.1.0`) alongside `install.sh` and lists only real CLI commands in `## CLI` (no `gira map`).
- AC3 = `website/docs/guides/quickstart.md` and `website/versioned_docs/version-0.1.0/guides/quickstart.md` document `mise` and `install.sh` installation, and `mise run docs:build` plus website quarantine grep pass.

## Plan

skip: single atomic documentation update across three markdown files verified against live v0.1.0 release assets and docs build.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `node tools/validate/validate.js` (baseline) | pass | 418bcb6 | exit 0 before edits |
| B2 | `mise run docs:build` (baseline) | pass | 418bcb6 | exit 0 before edits |
| AC1 | isolated `mise exec github:EduardoShibukawa/gira-g@0.1.0 -- gira version` | pass | bd47f59 | downloaded `gira_0.1.0_linux_amd64.tar.gz` and printed `gira 0.1.0 (commit: f5dad05...)` |
| AC2 | `README.md` inspection (`github:EduardoShibukawa/gira-g`, no `gira map`) | pass | bd47f59 | `mise use -g` + `mise.toml` documented; `grep "gira map" README.md` empty (exit 1) |
| AC3 | `mise run docs:build` + quarantine grep on `website/` | pass | bd47f59 | `docs:build` exit 0; canonical quarantine grep exit 1 (no matches) |
| R1 | `node tools/validate/validate.js` | pass | bd47f59 | exit 0 on working tree before verified commit |
| V1 | `git diff` review against AC1–AC3 | pass | bd47f59 | all 3 criteria covered; no unintended files or quarantined tokens |

## Outcome

Delivered and verified at `bd47f59`:
- Verified `mise` native `github:` backend installation against published `v0.1.0` release assets (`github:EduardoShibukawa/gira-g@0.1.0`, avoiding the deprecated `ubi:` backend);
- Updated `README.md`, `website/docs/guides/quickstart.md`, and `website/versioned_docs/version-0.1.0/guides/quickstart.md` with `mise` and `install.sh` instructions;
- Replaced the residual `gira map` entry in `README.md` `## CLI` with real commands (`gira index`, `gira impact`, `gira slice`, `gira explain`, `gira serve`).
