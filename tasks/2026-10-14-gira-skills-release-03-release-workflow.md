---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-skills-release
spec: gira-skills-release@1
depends_on:
  - 2026-10-14-gira-skills-release-02-goreleaser-sbom-cosign
---

# Tag-triggered GitHub Actions release workflow (F4; SC4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-skills-release; spec v1 approved)
- approved: create `.github/workflows/release.yml` (ADR 0011:30) by owner in parent task (2026-10-14)

## Objective

Create `.github/workflows/release.yml` to automate tag-triggered releases (`v*.*.*`)
and snapshot dry-runs (`workflow_dispatch`) with least-privilege permissions and
strict ADR 0012 quarantine isolation:
1. Trigger on `push: tags: ['v*.*.*']` and `workflow_dispatch`.
2. Top-level `permissions: contents: read`.
3. Preflight steps: checkout (`fetch-depth: 0`), `jdx/mise-action@v5`, Node 22 +
   `node tools/validate/validate.js`, `go test -race ./...`.
4. Snapshot dry-run on `workflow_dispatch`: `goreleaser release --snapshot --clean --skip=sign`
   plus `gira skills manifest` check.
5. Tag release job on `startsWith(github.ref, 'refs/tags/v')` with job-scoped
   `permissions: contents: write, id-token: write`, running `goreleaser release --clean`
   with `GITHUB_TOKEN` and keyless OIDC for `cosign`.

In scope:
- `.github/workflows/release.yml`;
- `TestSCReleaseWorkflow` in `internal/mcp` and `cmd/gira`.

Out of scope:
- modifying `.github/workflows/ci.yml`;
- pushing any git tag.

## Acceptance criteria

- AC1 = SC4 (`go test ./internal/mcp ./cmd/gira -run SCReleaseWorkflow -count=1`: verifies `.github/workflows/release.yml` triggers on `v*.*.*` and `workflow_dispatch`, enforces top-level `contents: read` and job-level `contents: write` + `id-token: write`, runs preflight + GoReleaser, and passes the ADR 0012 quarantine check)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before change | not run | | |
| AC1 | `go test ./internal/mcp ./cmd/gira -run SCReleaseWorkflow -count=1` | not run | | |
| R1 | `go test -race ./internal/mcp ./cmd/gira -count=1` + quarantined module static check | not run | | |
| V1 | diff read against AC1 (SC4) and ADR 0012 rule 3 (zero forbidden tokens in `.github/workflows/`) | not run | | |

## Constraints and budget

- Strictly forbid the quarantined module token anywhere in `.github/workflows/release.yml` (ADR 0012 rule 3).
- Keep `.github/workflows/ci.yml` untouched so E3 `TestSCGateLintStrictness` and `TestCICIConfiguration` stay green.

## Plan

1. Write failing test `TestSCReleaseWorkflow` in `internal/mcp` and `cmd/gira`.
2. Author `.github/workflows/release.yml`.
3. Run AC1, R1 (including the quarantined module's static check), V1 and validator; commit.
