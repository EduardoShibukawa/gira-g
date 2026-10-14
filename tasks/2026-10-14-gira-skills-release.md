---
schema: task/1
tier: heavy
status: active
spec: gira-skills-release@1
---

# Skills distribution & release (E5)

## Routing and approvals

- tier: heavy (H1: product calls Q1–Q6 on licensing, CGO posture, Cosign signing, skills manifest/check, E3 deferrals; H2: release packaging, signing, and CI workflow protected actions; H3: spans .goreleaser.yaml, internal/skills, cmd/gira, internal/gate, internal/eval, .github/workflows/, mise.toml; H5: decomposed children)
- confirmed: heavy by owner (2026-10-14, "Q1. A / Q2. B / Q3. A / Q4. A / Q5. A / Q6. A")
- decided: Q1 CGO_ENABLED=0 in .goreleaser.yaml while pure Go (Option A) by owner (2026-10-14, "Q1. A")
- decided: Q2 root LICENSE (MIT) + THIRD_PARTY_NOTICES.md (Option B) by owner (2026-10-14, "Q2. B")
- decided: Q3 Cosign signs checksums keyless on tag releases, --skip=sign in local snapshot task, hermetic ephemeral-key test (Option A) by owner (2026-10-14, "Q3. A")
- decided: Q4 add gira skills manifest, gira skills check, and skills-manifest.json release artifact (Option A) by owner (2026-10-14, "Q4. A")
- decided: Q5 keep default CI gate-lint warn-only, add --fail-on-ci for explicit ledger enforcement (Option A) by owner (2026-10-14, "Q5. A")
- decided: Q6 close findings register GC with Register.GC and gira review --gc [--keep-resolved <n>] (Option A) by owner (2026-10-14, "Q6. A")
- approved: spec v1 by owner (2026-10-14, "vamos criar a spec e as TASKS sem execucao, sera executado no modelo deep seek v4 pro")
- approved: edits to protected files (.github/workflows/release.yml, mise.toml, .goreleaser.yaml) for E5 by owner (2026-10-14, "vamos criar a spec e as TASKS sem execucao, sera executado no modelo deep seek v4 pro")

## Objective

Deliver E5 (skills distribution + release) per ADR 0010 §5/§6 and ADR 0011 §4:30:
reproducible GoReleaser packaging (`EduardoShibukawa/gira-g`, `CGO_ENABLED=0`),
root `LICENSE` (MIT) and `THIRD_PARTY_NOTICES.md`, `syft` SPDX SBOM + `cosign`
signing contract, `gira skills manifest` + `gira skills check` +
`skills-manifest.json`, tag-triggered `.github/workflows/release.yml`, and closure
of the two E3 deferrals (`gate-lint --fail-on-ci` and findings register `--gc`).

In scope:
- spec `gira-skills-release` v1;
- decomposed children 01–05 after spec v1 approval;
- integrated verification of SC1–SC7.

Out of scope:
- pushing git tags or publishing a live GitHub Release;
- E6 documentation website;
- ground-truth changes, `reviewed: true`;
- any model client (ADR 0001) or touching the quarantined comparative module from product configs (ADR 0012).

## Acceptance criteria

- AC1 = SC1–SC7 of `specs/gira-skills-release.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before E5 children | pass | f404ce0 | E4 merged on main; clean baseline |
| SC1 | `go test ./cmd/gira -run SCGoReleaserConfig -count=1` | not run | | |
| SC2 | `go test ./cmd/gira -run SCSBOMAndCosign -count=1` | not run | | |
| SC3 | `go test ./internal/skills ./cmd/gira -run SCSkillsManifestAndCheck -count=1` | not run | | |
| SC4 | `go test ./internal/mcp ./cmd/gira -run SCReleaseWorkflow -count=1` | not run | | |
| SC5 | `go test ./internal/gate ./internal/eval ./cmd/gira -run SCE5Deferrals -count=1` | not run | | |
| SC6 | `go test -race ./...` root and quarantined module; validator exit 0 | not run | | |
| SC7 | EndToEnd published/perturbed R/P unchanged; SC15 ≥ 90%; `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` | not run | | |
| AC1 | SC1–SC7 integrated | not run | | |
| R1 | `go test -race ./...` (root) | not run | | |
| R2 | `go test -race ./...` (quarantined module) | not run | | |
| R3 | `node tools/validate/validate.js` | not run | | |
| V1 | integrated diff read + fresh-context adversarial review | not run | | |

## Plan

Children are created at `@1` after spec v1 approval, in dependency order. Each one is test-first.

1. `-01-skills-manifest-and-check`: F3; SC3. `gira skills manifest` (JSON/Markdown + `--output`) and `gira skills check` (read-only drift detection across `claude`, `cursor`, `agents`).
2. `-02-goreleaser-sbom-cosign`: F1, F2; SC1, SC2. `.goreleaser.yaml` (`EduardoShibukawa/gira-g`, `CGO_ENABLED=0`, `signs:`, `skills-manifest.json`), root `LICENSE` (MIT) + `THIRD_PARTY_NOTICES.md`, `.gitignore` (`dist/`), `mise.toml` (`--skip=sign` on local snapshot task), and hermetic `syft` + `cosign` tests. Depends on 01.
3. `-03-release-workflow`: F4; SC4. `.github/workflows/release.yml` (tag-triggered + `workflow_dispatch` snapshot dry-run, least-privilege permissions, ADR 0012 quarantine compliance). Depends on 02.
4. `-04-e3-deferrals`: F5; SC5. `gate-lint --fail-on-ci` (`gate.LintOptions.FailOnCI`) and `gira review --gc [--keep-resolved <n>]` (`eval.Register.GC`). Depends on 01.
5. `-05-integrated`: F1–F5; SC6, SC7. Full snapshot dry-run, end-to-end SC1–SC7 verification, regression, and invariants. Depends on 01–04.

SC6–SC7 at every child. Rollback: `git revert` per child.

## Review

skip: runs after child 05 is verified.

## Checkpoint

- as of: 2026-10-14, branch e5-skills-release, head b60aae5, worktree clean after step 2 commit
- done: spike `docs/spikes/2026-10-14-e5-skills-release.md` written; owner confirmed `heavy`, answered Q1–Q6, and approved spec v1 + decomposition without execution; `specs/gira-skills-release.md` approved at v1; children 01–05 decomposed at `status: todo`
- next: execute children 01–05 in dependency order (01 → 02 → 03, 04 → 05), then run integrated verification SC1–SC7 + adversarial review
