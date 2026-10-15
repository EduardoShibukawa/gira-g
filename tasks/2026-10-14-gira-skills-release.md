---
schema: task/1
tier: heavy
status: accepted
verified_at: "d39a0ce"
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
- accepted: by owner (2026-10-15, "Aceito E5")

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
| SC1 | `go test ./cmd/gira -run SCGoReleaserConfig -count=1` | pass | d39a0ce | child 02 verified; 05 rerun pass (EduardoShibukawa/gira-g, CGO_ENABLED=0, sboms+signs, LICENSE+NOTICES, check exit 0) |
| SC2 | `go test ./cmd/gira -run SCSBOMAndCosign -count=1` | pass | d39a0ce | child 02 verified; 05 rerun pass (syft SPDX-2.3 17 pkgs; cosign ephemeral round-trip + tamper rejection) |
| SC3 | `go test ./internal/skills ./cmd/gira -run SCSkillsManifestAndCheck -count=1` | pass | d39a0ce | child 01 verified; 05 rerun pass (deterministic manifest, 6 digests; read-only check, non-zero on drift) |
| SC4 | `go test ./internal/mcp ./cmd/gira -run SCReleaseWorkflow -count=1` | pass | d39a0ce | child 03 verified; 05 rerun pass (v*.*.* + dispatch triggers, least-privilege perms, quarantine clean) |
| SC5 | `go test ./internal/gate ./internal/eval ./cmd/gira -run SCE5Deferrals -count=1` | pass | d39a0ce | child 04 verified; 05 rerun pass (--fail-on-ci fail-closed under CI, default warn-only; GC keeps open) |
| SC6 | `go test -race ./...` root and quarantined module; validator exit 0 | pass | d39a0ce | root race `-count=1` exit 0 (17 ok); bench race `-count=1` exit 0; validator exit 0 |
| SC7 | EndToEnd published/perturbed R/P unchanged; SC15 ≥ 90%; `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` | pass | d39a0ce | published R82.4/P100.0, perturbed R81.2/P100.0; SC15 focused 95.0/full 100.0; fixture diff exit 0; `reviewed: false` |
| AC1 | SC1–SC7 integrated | pass | d39a0ce | all SC rows pass on integrated head; snapshot dry-run exit 0 + cleanup (child 05 R1) |
| R1 | `go test -race ./...` (root) | pass | d39a0ce | `-count=1` exit 0, 17 ok 0 fail (one mid-run env collision with concurrent snapshot dry-run; clean rerun) |
| R2 | `go test -race ./...` (quarantined module) | pass | d39a0ce | bench `-count=1` exit 0 (layer0/1/2/quarantine ok) |
| R3 | `node tools/validate/validate.js` | pass | d39a0ce | exit 0 (manifest-integrity + artifact-records pass) |
| V1 | integrated diff read + fresh-context adversarial review | pass | d39a0ce | diff f404ce0..d39a0ce vs spec; 1 minor observation rejected with reason; zero open findings |

## Plan

Children are created at `@1` after spec v1 approval, in dependency order. Each one is test-first.

1. `-01-skills-manifest-and-check`: F3; SC3. `gira skills manifest` (JSON/Markdown + `--output`) and `gira skills check` (read-only drift detection across `claude`, `cursor`, `agents`).
2. `-02-goreleaser-sbom-cosign`: F1, F2; SC1, SC2. `.goreleaser.yaml` (`EduardoShibukawa/gira-g`, `CGO_ENABLED=0`, `signs:`, `skills-manifest.json`), root `LICENSE` (MIT) + `THIRD_PARTY_NOTICES.md`, `.gitignore` (`dist/`), `mise.toml` (`--skip=sign` on local snapshot task), and hermetic `syft` + `cosign` tests. Depends on 01.
3. `-03-release-workflow`: F4; SC4. `.github/workflows/release.yml` (tag-triggered + `workflow_dispatch` snapshot dry-run, least-privilege permissions, ADR 0012 quarantine compliance). Depends on 02.
4. `-04-e3-deferrals`: F5; SC5. `gate-lint --fail-on-ci` (`gate.LintOptions.FailOnCI`) and `gira review --gc [--keep-resolved <n>]` (`eval.Register.GC`). Depends on 01.
5. `-05-integrated`: F1–F5; SC6, SC7. Full snapshot dry-run, end-to-end SC1–SC7 verification, regression, and invariants. Depends on 01–04.

SC6–SC7 at every child. Rollback: `git revert` per child.

## Review

- [O1] `Register.Save` single-write vs crash-atomic (spec F5 "saved atomically") -> rejected: pre-existing shared `Save` semantics, no new risk from GC, rework out of scope (see child 05 ## Review).
- [C1–C6] Child 05 adversarial review: `skills check` strictly read-only; manifest hashes exact embedded `SKILL.md` bytes; snapshot build verified live; `release.yml` least-privilege + quarantine clean; `GC` never drops `open`; `FailOnCI` truth table matches Q5=A. All clean, no fixes needed.

No open findings. All divergences: none.

## Outcome

Delivered E5 (F1–F5; SC1–SC7) at spec v1: GoReleaser packaging (`EduardoShibukawa/gira-g`, `CGO_ENABLED=0`, `signs:` + `syft` SBOMs, MIT `LICENSE` + `THIRD_PARTY_NOTICES.md`), `gira skills manifest` + read-only `gira skills check` + `skills-manifest.json` artifact, tag-triggered `release.yml`, and both E3 deferrals (`gate-lint --fail-on-ci`, findings `--gc`). Integrated evidence at `d39a0ce`: SC1–SC7 pass, root + quarantined race suites pass, validator exit 0, invariants unchanged. Accepted by owner 2026-10-15. Pending (require separate explicit owner commands): any tag push (`v0.1.0` planejada, só após merge no `main`) or live GitHub Release publish.

## Checkpoint

- as of: 2026-10-14, branch e5-skills-release, head b60aae5, worktree clean after step 2 commit
- done: spike `docs/spikes/2026-10-14-e5-skills-release.md` written; owner confirmed `heavy`, answered Q1–Q6, and approved spec v1 + decomposition without execution; `specs/gira-skills-release.md` approved at v1; children 01–05 decomposed at `status: todo`
- next: execute children 01–05 in dependency order (01 → 02 → 03, 04 → 05), then run integrated verification SC1–SC7 + adversarial review
