---
schema: spec/1
status: draft
version: 1
---

# GIRA skills distribution and release (E5)

> Implements ADR 0010 (`docs/adr/0010-repo-development-workflow-and-toolchain.md`)
> §5/§6, ADR 0011 (`docs/adr/0011-ai-agent-workflow-and-model-policy.md`) §4:30,
> and closes the E3 deferrals (`specs/gira-envelopes-hygiene.md` Q6 and Open
> questions).

## Problem

E4 sealed at 47e79a8, merged at f404ce0: hermetic published R82.4/P100.0,
perturbed R81.2/P100.0, safety/determinism 100%, SC15 focused 95.0% / full
100.0%. Five release and distribution gaps remain for E5:

1. **GoReleaser target and multi-platform build posture.** `.goreleaser.yaml`
   targets `duh-org/duh-homelab` instead of `EduardoShibukawa/gira-g`, sets
   `CGO_ENABLED=1` across all five targets even though the codebase has zero
   `import "C"` (causing `goreleaser release --snapshot --clean` to fail when
   cross-compiling `arm64`/`darwin`/`windows` on a single Linux `amd64` host),
   and references `LICENSE*` when neither root `LICENSE` nor
   `THIRD_PARTY_NOTICES.md` (ADR 0010 §6.5) exists.
2. **SBOM and Cosign signing contract.** `syft` SPDX SBOM generation is wired
   in `.goreleaser.yaml`, and `cosign` is pinned in `mise.toml`, but there is
   no `signs:` stanza in `.goreleaser.yaml` and no hermetic verification test
   covering `syft` SPDX JSON validity and `cosign` blob signing/verification.
3. **Skills catalog manifest and drift check.** `gira skills install --host
   claude|cursor|agents` installs the 6 canonical skills idempotently, but
   there is no `gira skills manifest` (emitting version, skill metadata, and
   per-skill SHA-256 digests), no non-mutating `gira skills check` to detect
   drift in installed host files, and no `skills-manifest.json` release
   artifact.
4. **Tag-triggered release workflow.** Only `.github/workflows/ci.yml` exists;
   there is no `.github/workflows/release.yml` to run preflight checks, build
   release artifacts, generate SBOMs, sign `checksums.txt` keyless via GitHub
   OIDC (`cosign`), and publish on `v*.*.*` tags (protected by ADR 0011:30).
5. **E3 deferrals ("E4+ decides").** Two open questions from E3 remain open:
   (a) the exact `gate-lint --strict` CI vs explicit-ledger policy, and
   (b) a deterministic retention/GC policy (`--gc`) for the L3 findings
   register (`reviews/findings.md`).

Heavy triggers: H1 (product decisions Q1–Q6 on licensing, CGO posture, signing,
skills surface, and E3 deferrals), H2 (release/signing/workflow one-way doors
and protected actions), H3 (spans `.goreleaser.yaml`, `internal/skills`,
`cmd/gira`, `internal/gate`, `internal/eval`, `.github/workflows/`,
`mise.toml`), H5 (decomposed children).

## Expected behavior

One epic, five additive features.

### F1 — GoReleaser packaging & licensing hygiene

`.goreleaser.yaml` targets `owner: EduardoShibukawa`, `name: gira-g` and sets
`CGO_ENABLED=0` across all 5 targets (`linux`/`darwin` × `amd64`/`arm64`,
`windows/amd64`, ignoring `windows/arm64`) while the codebase is pure Go (the
tree-sitter seam from ADR 0002 stays intact). Root `LICENSE` (MIT) and
`THIRD_PARTY_NOTICES.md` (attributing the Apache-2.0 OpenTelemetry Demo subset
in `testdata/corpus/otel-demo-3.1.0-mini/` and Go module dependencies) are
added at the repository root and packaged into every release archive alongside
`README.md` and `skills-manifest.json`. `.gitignore` ignores `dist/`.

### F2 — SBOM (`syft`) and signing (`cosign`) contract

`.goreleaser.yaml` keeps `syft` SPDX JSON SBOM generation (`${artifact}.spdx.sbom.json`
for every archive) and adds a `signs:` stanza signing `checksums` via `cosign
sign-blob --yes --output-signature=${signature} --output-certificate=${certificate}
${artifact}` (keyless OIDC on tag releases). In `mise.toml`, the local
`release` task runs `goreleaser release --snapshot --clean --skip=sign` so
unauthenticated local/CI snapshot builds produce all archives, SBOMs, and
`checksums.txt` without requiring a live OIDC token. Hermetic Go tests verify:
(1) `goreleaser check` and snapshot packaging, (2) valid SPDX JSON output from
`syft`, and (3) `cosign generate-key-pair` + `cosign sign-blob` + `cosign
verify-blob` round-trip in `t.TempDir()`.

### F3 — Skills catalog manifest, drift check & release artifact

`internal/skills` and `cmd/gira` gain two additive subcommands under `gira skills`:
- `gira skills manifest [--format json|md] [--output <path>]`: emits a
  deterministic `Manifest` document containing the `gira` version, catalog
  digest (`sha256:<hex>`), and the sorted list of canonical skills (`name`,
  `description`, `operations`, `output_contract`, `budget`, `sha256` of the
  canonical `SKILL.md` bytes). A GoReleaser `before.hooks` step (or build hook)
  writes `dist/skills-manifest.json` (or includes `skills-manifest.json` in
  `extra_files` / archive files) so every release ships the verifiable manifest.
- `gira skills check --host claude|cursor|agents [--dir <project>] [--index <path>] [--command <cmd>] [--hook]`:
  computes the expected host file mutations in memory without touching disk and
  reports `{host, in_sync, missing, drifted, ok}` (JSON on stdout), returning a
  non-zero exit code when any owned file or region is missing or drifted.

### F4 — Tag-triggered GitHub Actions release workflow

`.github/workflows/release.yml` (protected by ADR 0011:30):
- Triggers on `push: tags: ['v*.*.*']` and `workflow_dispatch` (which runs a
  non-publishing `--snapshot --clean --skip=sign` dry-run).
- Top-level `permissions: contents: read`.
- Runs preflight (`go test -race ./...` and `node tools/validate/validate.js`),
  and on tag push runs GoReleaser with `permissions: contents: write, id-token: write`
  to publish the release, SBOMs, `skills-manifest.json`, and keyless Cosign
  signatures.
- Complies strictly with ADR 0012 rule 3 (zero references to the quarantined
  comparative module).

### F5 — E3 deferrals: `gate-lint` policy & findings register GC

1. **`gate-lint` policy (Q5=A):** Default CI job in `.github/workflows/ci.yml`
   remains warn-only (`continue-on-error: true`) because `.gira/` is gitignored
   in clean checkouts. `gate.LintOptions` and `gira gate-lint` gain
   `--fail-on-ci` (or `FailOnCI bool`) so callers supplying an explicit
   committed/generated `--state <path>` in CI can enforce fail-closed `--strict`
   even when `CI=true` is set in the environment, while keeping the default
   unconfigured CI job warn-only (`TestSCGateLintStrictness` preserved).
2. **Findings register retention/GC (Q6=A):** `eval.Register.GC(keepResolved int)`
   and `gira review --findings-register <path> --gc [--keep-resolved <n>]`
   deterministically compact the register: all findings with status `open` are
   kept unconditionally in original order; resolved findings (`fixed`,
   `closed`, `verified`) are pruned down to the last `keepResolved` entries
   (default `0`), and remaining entries are validated and saved atomically.

## Scope

**In scope**

- F1–F5 above, all additive.
- Root `LICENSE` (MIT) and `THIRD_PARTY_NOTICES.md`.
- `.goreleaser.yaml`, `mise.toml` (`release` task `--skip=sign`), `.gitignore`
  (`dist/`), `.github/workflows/release.yml`.
- `gira skills manifest` and `gira skills check`.
- `gira gate-lint --fail-on-ci` and `gira review --gc [--keep-resolved <n>]`.

**Out of scope**

- Pushing git tags or publishing a live GitHub Release (never run without a
  separate explicit owner command).
- E6 documentation website.
- Ground-truth changes, `reviewed: true`.
- Any model client (ADR 0001), paid runs, or touching the quarantined module
  from product configs (ADR 0012).

**Non-goals**

- Changing recall/precision, envelopes (E3), or human-surface rendering (E4).
- Introducing CGO dependencies in E5 (the tree-sitter seam from ADR 0002 stays
  ready for future use).

## Acceptance criteria

- **SC1 — GoReleaser & licensing hygiene:** `go test ./cmd/gira -run SCGoReleaserConfig -count=1`
  verifies `.goreleaser.yaml` targets `EduardoShibukawa/gira-g`, sets
  `CGO_ENABLED=0`, configures `sboms:` (`syft`) and `signs:` (`cosign`),
  `LICENSE` (MIT) and `THIRD_PARTY_NOTICES.md` exist and are non-empty, and
  `goreleaser check` exits 0.
- **SC2 — SBOM & Cosign round-trip:** `go test ./cmd/gira -run SCSBOMAndCosign -count=1`
  verifies `syft` produces valid SPDX JSON (`spdxVersion`, `packages`) on a
  built `gira` archive/binary and `cosign` signs and verifies a checksums blob
  using an ephemeral keypair in `t.TempDir()`.
- **SC3 — Skills manifest & drift check:** `go test ./internal/skills ./cmd/gira -run SCSkillsManifestAndCheck -count=1`
  verifies `gira skills manifest` is deterministic with valid SHA-256 digests
  for all 6 skills, `gira skills check` passes right after `gira skills install`
  for all 3 hosts (`claude`, `cursor`, `agents`), and `gira skills check`
  reports drift (non-zero exit) when an owned region or key is tampered with or
  missing, without modifying any file on disk.
- **SC4 — Release workflow & quarantine guard:** `go test ./internal/mcp ./cmd/gira -run SCReleaseWorkflow -count=1`
  verifies `.github/workflows/release.yml` triggers on `v*.*.*` tags and
  `workflow_dispatch`, uses least-privilege permissions (`contents: read` at
  root; `contents: write` and `id-token: write` on the release job), and passes
  the ADR 0012 quarantine static check.
- **SC5 — E3 deferrals (`gate-lint` & findings GC):**
  `go test ./internal/gate ./internal/eval ./cmd/gira -run SCE5Deferrals -count=1`
  verifies `gate-lint --strict --fail-on-ci` fails closed even under `CI=true`
  while default CI warn-only (`TestSCGateLintStrictness`) stays green, and
  `gira review --gc [--keep-resolved <n>]` preserves all `open` findings while
  compacting resolved ones.
- **SC6 — Regression & governance:** `go test -race ./...` passes at root and
  in the quarantined module; `node tools/validate/validate.js` exits 0.
- **SC7 — I3/I4/I5 invariants:** EndToEnd hermetic published `R82.4 / P100.0`
  and perturbed `R81.2 / P100.0` unchanged; SC15 ≥ 90% on both tours;
  `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` in
  `fixtures/eval/ground-truth.v2.json`.

## Interfaces and dependencies

| Interface | Change |
|---|---|
| CLI `skills` | `gira skills manifest [--format json\|md] [--output <path>]`; `gira skills check --host claude\|cursor\|agents [--dir <project>] [--index <path>] [--command <cmd>] [--hook]` |
| CLI `gate-lint` | `--fail-on-ci` flag (`gate.LintOptions.FailOnCI`) |
| CLI `review` | `--gc` and `--keep-resolved <n>` flags (`eval.Register.GC`) |
| `.goreleaser.yaml` | `EduardoShibukawa/gira-g`, `CGO_ENABLED=0`, `signs:` (`cosign`), `LICENSE` + `THIRD_PARTY_NOTICES.md` + `skills-manifest.json` |
| `mise.toml` | `tasks.release` adds `--skip=sign` for local snapshot builds |
| `.github/workflows/release.yml` | tag-triggered (`v*.*.*`) + `workflow_dispatch` snapshot dry-run |
| Root files | `LICENSE` (MIT), `THIRD_PARTY_NOTICES.md`, `.gitignore` (`dist/`) |

## Decisions

- **Q1 → Option A (by owner, 2026-10-14):** Set `CGO_ENABLED=0` in
  `.goreleaser.yaml` while the codebase is pure Go (keeping the ADR 0002 seam),
  so multi-target snapshot builds succeed deterministically on a single host.
- **Q2 → Option B (by owner, 2026-10-14):** Root `LICENSE` is MIT, plus
  `THIRD_PARTY_NOTICES.md` attributing the Apache-2.0 OpenTelemetry Demo
  mini-corpus and Go dependencies.
- **Q3 → Option A (by owner, 2026-10-14):** Wire `signs:` in `.goreleaser.yaml`
  for keyless Cosign signing of `checksums.txt` on tag releases, pass
  `--skip=sign` in the local `mise.toml` snapshot task, and test `cosign` +
  `syft` hermetically in Go tests with an ephemeral keypair in `t.TempDir()`.
- **Q4 → Option A (by owner, 2026-10-14):** Add `gira skills manifest`,
  `gira skills check`, and include `skills-manifest.json` in release artifacts.
- **Q5 → Option A (by owner, 2026-10-14):** Keep the default CI `gate-lint` job
  warn-only when no ledger is committed, add `--fail-on-ci` for explicit ledger
  enforcement in CI, and close the E3 open question.
- **Q6 → Option A (by owner, 2026-10-14):** Close the E3 findings register GC
  open question with `eval.Register.GC(keepResolved)` and `gira review --gc
  [--keep-resolved <n>]`.

## Risks and constraints

- **Protected files (ADR 0011:28-30):** Editing `.github/workflows/release.yml`
  and `mise.toml` requires explicit owner approval recorded in the parent task;
  no tag push or live release publication is ever executed by the agent.
- **Quarantine (ADR 0012 rule 3):** Neither `.github/workflows/release.yml`,
  `mise.toml`, `.goreleaser.yaml`, nor any Go package may reference the
  quarantined comparative module.
- **Snapshot determinism:** `goreleaser release --snapshot --clean --skip=sign`
  writes to `dist/`; tests clean up `dist/` or use isolated temp output dirs so
  the working tree stays clean.

## Verification

- Per child + integrated: B1 `go test -race ./...` (root); B2 `go test -race ./...`
  (quarantined module); B3 `go test ./internal/eval -run EndToEnd -v`
  (published/perturbed R/P); B4 SC1–SC7 commands above; B5
  `node tools/validate/validate.js`.
- **Review focus:** Does `gira skills check` ever mutate disk? Does `gira skills
  manifest` hash the exact canonical `SKILL.md` bytes? Does `.goreleaser.yaml`
  pass `goreleaser check` and build snapshot archives cleanly? Does
  `release.yml` obey least-privilege permissions and ADR 0012 quarantine? Does
  `Register.GC` ever drop an `open` finding?

## Rollback

`git revert` per child. Because no git tag is pushed and no GitHub Release is
published during E5 verification, all changes are purely repository-local and
fully reversible via `git revert`.

## Open questions

None (Q1–Q6 resolved by owner on 2026-10-14).

## Revisions

- v1 (2026-10-14): draft v1 incorporating owner decisions Q1–Q6.
