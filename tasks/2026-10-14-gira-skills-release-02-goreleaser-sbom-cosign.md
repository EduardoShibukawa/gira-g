---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-skills-release
spec: gira-skills-release@1
depends_on:
  - 2026-10-14-gira-skills-release-01-skills-manifest-and-check
---

# GoReleaser packaging, licensing, SBOM & Cosign contract (F1, F2; SC1, SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-skills-release; spec v1 approved)
- approved: edit `.goreleaser.yaml` and `mise.toml` (`tasks.release` `--skip=sign`) by owner in parent task (2026-10-14)

## Objective

Fix `.goreleaser.yaml` and repository licensing so `goreleaser check` and local
snapshot builds (`mise run release`) succeed across all 5 targets, and verify
the `syft` SPDX SBOM and `cosign` blob signing/verification contracts
hermetically:
1. Update `.goreleaser.yaml`: `owner: EduardoShibukawa`, `name: gira-g`,
   `CGO_ENABLED=0` across all 5 targets (`linux`/`darwin` × `amd64`/`arm64`,
   `windows/amd64`), generate `skills-manifest.json` via `before.hooks` and
   include `README.md`, `LICENSE*`, `THIRD_PARTY_NOTICES.md`, and
   `skills-manifest.json` in archives/release extra files, keep `sboms:` (`syft`
   SPDX JSON), and add `signs:` (`cosign sign-blob --yes --output-signature=${signature} --output-certificate=${certificate} ${artifact}` on `checksums`).
2. Add root `LICENSE` (MIT) and `THIRD_PARTY_NOTICES.md` (attributing the
   Apache-2.0 OpenTelemetry Demo subset in `testdata/corpus/otel-demo-3.1.0-mini/`
   and Go module dependencies).
3. Add `dist/` and `skills-manifest.json` (if generated at root during hooks) to
   `.gitignore`, and update `mise.toml` `tasks.release` to
   `goreleaser release --snapshot --clean --skip=sign`.

In scope:
- `.goreleaser.yaml`, `LICENSE`, `THIRD_PARTY_NOTICES.md`, `.gitignore`, `mise.toml`;
- `TestSCGoReleaserConfig` and `TestSCSBOMAndCosign` in `cmd/gira`.

Out of scope:
- `.github/workflows/release.yml` (child 03);
- pushing tags or publishing a live release.

## Acceptance criteria

- AC1 = SC1 (`go test ./cmd/gira -run SCGoReleaserConfig -count=1`: `.goreleaser.yaml` targets `EduardoShibukawa/gira-g`, `CGO_ENABLED=0`, `sboms:` and `signs:` configured, `LICENSE` and `THIRD_PARTY_NOTICES.md` present, `goreleaser check` exits 0)
- AC2 = SC2 (`go test ./cmd/gira -run SCSBOMAndCosign -count=1`: `syft` emits valid SPDX JSON on built artifact and `cosign` signs/verifies checksums with an ephemeral keypair in `t.TempDir()`)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before change | not run | | |
| AC1 | `go test ./cmd/gira -run SCGoReleaserConfig -count=1` | not run | | |
| AC2 | `go test ./cmd/gira -run SCSBOMAndCosign -count=1` | not run | | |
| R1 | `mise exec -- goreleaser check` and `go test -race ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1–AC2; verify ADR 0012 quarantine compliance | not run | | |

## Constraints and budget

- Never push tags or publish a release; snapshot tests clean up any temporary `dist/` artifacts.
- Never write the forbidden quarantine token in `mise.toml`, `.goreleaser.yaml`, or `cmd/` (ADR 0012 rule 3).
- Zero diff in `fixtures/` and `testdata/`.

## Plan

1. Write failing tests `TestSCGoReleaserConfig` and `TestSCSBOMAndCosign` in `cmd/gira`.
2. Create root `LICENSE` (MIT) and `THIRD_PARTY_NOTICES.md`; update `.gitignore`, `.goreleaser.yaml`, and `mise.toml`.
3. Run AC1, AC2, R1, V1 and validator; ensure working tree is clean (`dist/` removed); commit.
