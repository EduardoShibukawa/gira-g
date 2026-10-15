# Spike: E5 — Skills distribution + release (read-only, unapproved)

> Read-only spike on `main` @ `f404ce0`. No spec, no task records, no product or
> CI edits. Scope derived from ADR 0010 (`docs/adr/0010-repo-development-workflow-and-toolchain.md`),
> ADR 0011 (`docs/adr/0011-ai-agent-workflow-and-model-policy.md`),
> `specs/gira-agent-skills-and-eval.md` (part A.3), and the E3 deferrals
> (`specs/gira-envelopes-hygiene.md` Q6 and Open questions).

## Routing

`tier: heavy (H1, H2, H3, H5)`

- **H1 (product / preference calls):** repository owner/name in `.goreleaser.yaml`, root `LICENSE`, CGO vs pure-Go build posture in GoReleaser, signing mode (`cosign` keyless OIDC vs ephemeral/local key in dry-runs), skills catalog version/manifest metadata, and whether the two E3 deferrals (`gate-lint --strict` in CI; findings register GC/retention) enter E5.
- **H2 (one-way door / protected action):** release packaging, cryptographic signing (`cosign`), SBOM publication (`syft`), and tag-triggered GitHub Actions workflows (`.github/workflows/`, governed by ADR 0011 §4:30).
- **H3 (cross-module footprint):** touches `.goreleaser.yaml`, `internal/skills/`, `cmd/gira/`, `.github/workflows/`, licensing files, and (if included) `internal/gate/` + `internal/eval/`.
- **H5 (multi-session / decomposed children):** decomposes cleanly into ordered `standard` children.

---

## Baseline inspection (read-only on `main` @ `f404ce0`)

### 1. `.goreleaser.yaml` & snapshot build
- `mise exec -- goreleaser check` **passes** (`1 configuration file(s) validated`).
- `mise exec -- goreleaser build --snapshot --clean --single-target` **succeeds** on `linux_amd64_v1` in ~2s and injects `-X main.version`, `-X main.commit`, `-X main.date` (`gira 0.0.0-SNAPSHOT-f404ce0`).
- **Four concrete gaps found in `.goreleaser.yaml`:**
  1. **Wrong GitHub target (`.goreleaser.yaml:9-10`):** points to `owner: duh-org`, `name: duh-homelab`, whereas `git remote -v` and `go.mod` are `EduardoShibukawa/gira-g`.
  2. **Multi-target `goreleaser release --snapshot --clean` fails on single host:** `.goreleaser.yaml:19` sets `CGO_ENABLED=1` across `linux`/`darwin`/`windows` × `amd64`/`arm64`. Running `mise run release` on an `x86_64` Linux host fails when compiling `linux_arm64` (`runtime/cgo: gcc_arm64.S: Error: no such instruction`) because the host `gcc` is `x86_64` and no cross C compiler or split-build flag is configured. Furthermore, the codebase currently has **zero `import "C"`** (Layer 3 Go AST uses `go/ast`, and tree-sitter CGO has not been wired in yet), OR if `CGO_ENABLED=1` is kept per ADR 0010 §5, single-host snapshot runs need either GoReleaser split/single-target semantics or `CGO_ENABLED=0` until a CGO grammar is actually linked.
  3. **Missing root `LICENSE` and `THIRD_PARTY_NOTICES.md`:** `.goreleaser.yaml:46` archives `LICENSE*`, which matches nothing at the repo root (only `testdata/corpus/otel-demo-3.1.0-mini/LICENSE` exists). ADR 0010 §6.5 also specifies `THIRD_PARTY_NOTICES.md`.
  4. **No `signs:` stanza in `.goreleaser.yaml`:** `cosign = "3.1.3"` is pinned in `mise.toml`, and `sboms:` (`syft` SPDX JSON) is configured in `.goreleaser.yaml:48-57`, but there is no `signs:` block for `checksums.txt` / archives.

### 2. SBOM (`syft`) and signing/provenance (`cosign`)
- Tools available via `mise`: `goreleaser v2.7.0`, `syft v1.54.1`, `cosign v3.1.3`.
- **SBOM (`syft`):** `.goreleaser.yaml` already defines `sboms:` on `artifacts: archive` producing `${artifact}.spdx.sbom.json`. Works offline in `--snapshot` once the archive step succeeds.
- **Signing (`cosign`):**
  - **Tag release (GitHub Actions):** ADR 0010 §5 specifies keyless signing via GitHub OIDC (`cosign sign-blob --yes` with `id-token: write` permission, outputting `.sig` / `.pem` bundle) and/or `actions/attest-build-provenance`.
  - **Local / CI `--snapshot` dry-run:** Keyless OIDC cannot run offline without an OIDC token. We can either:
    - **(a)** Gate the GoReleaser `signs:` stanza with an env/condition or skip signing on `--snapshot` (`--skip=sign`), while testing `cosign sign-blob` / `cosign verify-blob` in Go tests using an ephemeral keypair generated in `t.TempDir()`; or
    - **(b)** Keep `cosign` in the release workflow aggregator job and test the signing/verification contract in a dedicated hermetic test.

### 3. Skills catalog distribution (`internal/skills` + `cmd/gira`)
- **What exists:**
  - 6 canonical skills embedded via `//go:embed catalog/*/SKILL.md` (`gira-arch-diff`, `gira-context`, `gira-explain-component`, `gira-impact-briefing`, `gira-query`, `gira-tour`).
  - `gira skills install --host claude|cursor|agents [--dir <project>] [--index <path>] [--command <cmd>] [--hook]` is implemented and tested (`SC10`, `TestInstallIsIdempotentAndMarkerScoped`, `TestHookIsOptInAppendOnlyAndIdempotent`, `TestNudgeNeverBlocks`).
- **Gaps for E5 distribution & release:**
  1. **No catalog version / content digest / manifest:** Installed `SKILL.md` and `.cursor/rules/gira.mdc` blocks carry `gira:begin` / `gira:end` markers, but do not record the `gira` version or a catalog SHA-256 digest, nor is there a `gira skills list` / `gira skills status` (or `--check` flag) to inspect installed skills, detect drift, or emit a deterministic skills manifest (`skills-manifest.json` with per-skill SHA-256) in release artifacts.
  2. **Standalone skills bundle in release artifacts:** Users who want to inspect or vendor the canonical `SKILL.md` catalog (or verify checksums before installing) have no standalone `gira-skills_<version>.tar.gz` or `skills-manifest.json` attached to the release alongside the binaries.
  3. **End-to-end release-binary smoke test:** No test builds/invokes the packaged `gira` binary + `gira skills install` + `gira version` + `syft` SBOM schema check as an integrated release gate.

### 4. GitHub Actions release workflow (protected by ADR 0011 §4:30)
- Today only `.github/workflows/ci.yml` exists (`test` on 3 OSes, `validate`, `gate-lint`, `floor-guard` on `ubuntu-latest`).
- There is no `.github/workflows/release.yml`.
- Per ADR 0010 §5 and ADR 0011 §4:30:
  - Must only be authored with explicit owner approval (`approved:` in the task record).
  - Triggered on `push: tags: ['v*.*.*']` (never publishes on normal branches/PRs; optional `workflow_dispatch` dry-run in `--snapshot` mode).
  - Minimal permissions: `contents: read` at workflow level; `contents: write` and `id-token: write` (for keyless `cosign` / provenance) scoped strictly to the release/publish job.
  - ADR 0012 rule 3: must never contain the forbidden quarantine token.

### 5. E3 deferred items ("E4+ decides")
1. **Promote `gate-lint` from warn-only to `--strict` in CI (`.github/workflows/ci.yml` + `cmd/gira/gate.go`):**
   - Currently `ci.yml:74-77` runs `go run ./cmd/gira gate-lint || true` with `continue-on-error: true`, and `internal/gate/gate.go:106` downgrades `Strict: true` to a warning whenever `opts.CI` is true (`CI=true` is set by GitHub Actions).
   - Also note: `.gira/` is in `.gitignore` (`.gitignore:15`), so `.gira/gate.json` is not committed in the repo; if `gate-lint --strict` ran fail-closed in CI without a committed ledger (or a release-time generated ledger), CI would fail immediately on missing `.gira/gate.json`.
2. **Retention / GC policy for the L3 findings register (`reviews/findings.md`, `internal/eval/findings.go`):**
   - Currently `eval.Register` loads/saves all rows in `reviews/findings.md` and `gira review --findings-register` lists or runs `--fix-check`. There is no `--prune` / `--gc` / retention policy for closed/verified findings (or cap on resolved rows).

---

## Proposed E5 Feature Set (F1–F5)

- **F1 — Reproducible GoReleaser packaging & licensing hygiene:**
  - Fix `.goreleaser.yaml` repo target (`EduardoShibukawa/gira-g`), align build configuration so `mise run release-check` and local snapshot builds succeed cleanly, and add root `LICENSE` + `THIRD_PARTY_NOTICES.md` (covering the Apache-2.0 OpenTelemetry Demo mini-corpus in `testdata/corpus/otel-demo-3.1.0-mini/` and Go dependencies) included in release archives.
- **F2 — SBOM & Cosign signing contract:**
  - Wire `syft` SPDX SBOM generation and `cosign` checksum signing into the release configuration (`signs:` stanza for `checksums.txt`), with `--snapshot` compatibility (`--skip=sign` for local unauthenticated snapshots, plus hermetic Go test verifying `cosign` blob signing/verification with an ephemeral key and `syft` SPDX JSON validity).
- **F3 — Skills catalog manifest, version stamping & verification (`gira skills`):**
  - Deterministic catalog manifest (`skills.Manifest()` / `gira skills manifest` or `gira skills list --format json|md`) with per-skill SHA-256 digests and version stamp, plus `gira skills check` (or `install --check`) to verify installed host files are in sync with the binary's embedded catalog without mutating the working tree.
- **F4 — Tag-triggered GitHub Actions release workflow (`.github/workflows/release.yml`):**
  - Protected workflow (ADR 0011:30) triggered on `v*.*.*` tags (plus optional snapshot dry-run), minimal permissions, running `mise` toolchain, tests/validator preflight, GoReleaser build + `syft` SBOM + `cosign` keyless signing, without ever pushing tags or publishing releases during automated tests.
- **F5 — E3 deferrals (subject to owner decision in Q5/Q6):**
  - Optional: findings register GC/pruning (`gira review --gc` / `eval.Register.Compact`) and/or `gate-lint` strictness policy.

---

## Open Questions for Owner (Q1–Q6)

- **Q1 (GoReleaser CGO & snapshot strategy):** Today the codebase has zero `import "C"` (pure Go), while `.goreleaser.yaml` sets `CGO_ENABLED=1` across all 5 OS/arch targets — which causes `mise run release` (`goreleaser release --snapshot --clean`) to fail on a single Linux x86_64 host when cross-compiling `linux/arm64`, `darwin/*`, and `windows/amd64`. How should we configure `.goreleaser.yaml` and `.github/workflows/release.yml`?
  - *Option A (Recommended):* Set `CGO_ENABLED=0` in `.goreleaser.yaml` while the codebase is pure Go (keeping the tree-sitter seam from ADR 0002), so `mise run release` (`goreleaser release --snapshot --clean`) builds all 5 OS/arch archives + SBOMs + `checksums.txt` deterministically on any single runner/workstation, AND run a 3-OS smoke test of the built binaries in CI. (If CGO is later wired in behind the seam, switch to GoReleaser Pro split builds or `zig cc`.)
  - *Option B:* Keep `CGO_ENABLED=1`, change `mise.toml` `release` task to `--single-target`, and in `.github/workflows/release.yml` build per-OS archives on the 3-OS matrix (`ubuntu-latest`, `macos-latest`, `windows-latest`) via `goreleaser build --single-target` followed by a merge/release job.
- **Q2 (Root `LICENSE` & `THIRD_PARTY_NOTICES.md`):** `.goreleaser.yaml` packages `LICENSE*`, which currently does not exist at the repo root, and ADR 0010 §6.5 requires `THIRD_PARTY_NOTICES.md`. Which license should the root `LICENSE` use alongside `THIRD_PARTY_NOTICES.md`?
  - *Option A (Recommended):* Apache-2.0 at root `LICENSE` (matching the vendored OpenTelemetry Demo mini-corpus license) + `THIRD_PARTY_NOTICES.md` attributing the OpenTelemetry Demo subset (`testdata/corpus/otel-demo-3.1.0-mini`) and Go module dependencies.
  - *Option B:* MIT at root `LICENSE` + `THIRD_PARTY_NOTICES.md`.
  - *Option C:* Do not add a root `LICENSE`; remove `LICENSE*` from `.goreleaser.yaml` and only ship `THIRD_PARTY_NOTICES.md`.
- **Q3 (Cosign signing mode):** How should `cosign` signing be wired in `.goreleaser.yaml` and tested?
  - *Option A (Recommended):* Configure `signs:` in `.goreleaser.yaml` to sign `checksums` with `cosign sign-blob --yes --output-signature=${signature} --output-certificate=${certificate} ${artifact}` (keyless OIDC on tag releases), pass `--skip=sign` in the local `mise.toml` snapshot task (`goreleaser release --snapshot --clean --skip=sign` so local snapshots need no OIDC token), and test `cosign` sign+verify end-to-end in `go test` using an ephemeral local keypair in `t.TempDir()`.
  - *Option B:* Sign individual archives in addition to `checksums.txt`.
- **Q4 (Skills catalog distribution & CLI surface):** Beyond the existing `gira skills install --host claude|cursor|agents`, what should E5 add for skills distribution and release?
  - *Option A (Recommended):* Add `gira skills manifest` (or `gira skills list --format json|md`) emitting the deterministic catalog manifest (skill names, operations, budgets, SHA-256 of each canonical `SKILL.md`, and binary version), add `gira skills check --host <host> [--dir <project>]` (non-mutating drift check that exits non-zero if installed host files differ from the embedded catalog), and ship `skills-manifest.json` as an extra release artifact in GoReleaser.
  - *Option B:* Keep `gira skills install` as the only subcommand and only add release packaging tests.
- **Q5 (E3 deferral 1 — `gate-lint --strict` in CI):** Should E5 promote `gate-lint` in `.github/workflows/ci.yml` from warn-only to fail-closed `--strict`, or keep it warn-only in CI?
  - *Option A (Recommended):* Keep `gate-lint` warn-only in `.github/workflows/ci.yml` (because `.gira/` is gitignored and no static `.gira/gate.json` is committed to the repo, so fail-closed on missing ledger in CI would break every PR unless a committed ledger is maintained on every commit), OR promote `--strict` only when an explicit `--state` ledger is present/supplied, and document this as the settled policy (closing the E3 open question).
  - *Option B:* Commit a canonical `.gira/gate.json` ledger (un-ignoring `.gira/gate.json`), remove the `opts.CI` warn-only bypass in `internal/gate/gate.go` when `--strict` is set, and make `gate-lint --strict` fail-closed in `.github/workflows/ci.yml`.
- **Q6 (E3 deferral 2 — Findings register retention/GC policy):** Should E5 close the findings register retention/GC deferral from E3?
  - *Option A (Recommended):* Yes, close it in E5 with a small deterministic GC rule in `internal/eval/findings.go` + `gira review --gc [--keep-resolved <n>]`: keep all non-resolved (`open`) findings unconditionally, and compact `resolved`/`closed`/`verified` findings (or prune those whose repro path was temporary/cleaned up when `--gc` is invoked), while preserving fail-closed validation on load.
  - *Option B:* Leave findings register GC out of E5 and defer to E6+.

## Revisions

- v1 (2026-10-14): read-only spike and owner question round Q1–Q6.
