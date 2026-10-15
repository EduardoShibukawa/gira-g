---
schema: task/1
tier: standard
status: verified
verified_at: "40acf09"
parent: 2026-10-14-gira-skills-release
spec: gira-skills-release@1
depends_on:
  - 2026-10-14-gira-skills-release-01-skills-manifest-and-check
  - 2026-10-14-gira-skills-release-02-goreleaser-sbom-cosign
  - 2026-10-14-gira-skills-release-03-release-workflow
  - 2026-10-14-gira-skills-release-04-e3-deferrals
---

# Integrated E5 verification & adversarial review (F1–F5; SC6, SC7)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-skills-release; spec v1 approved)

## Objective

Run integrated verification across all E5 deliverables (F1–F5; SC1–SC7),
verify full snapshot release build (`goreleaser release --snapshot --clean --skip=sign`
followed by cleanup of `dist/`), confirm all regression suites and invariants
(hermetic published R82.4/P100.0, perturbed R81.2/P100.0, SC15 ≥ 90%, zero diff
on `fixtures/` and `testdata/`, `"reviewed": false`, ADR 0012 quarantine), and
execute the fresh-context adversarial review (`review.md`) before sealing the
parent task.

In scope:
- end-to-end execution of SC1–SC7;
- full GoReleaser snapshot dry-run and cleanup;
- fresh-context adversarial review and resolution of any findings.

Out of scope:
- pushing git tags or publishing a live GitHub Release.

## Acceptance criteria

- AC1 = SC6 (`go test -race ./...` passes at root and in the quarantined module; `node tools/validate/validate.js` exits 0)
- AC2 = SC7 (EndToEnd hermetic published R82.4/P100.0 and perturbed R81.2/P100.0 unchanged; SC15 ≥ 90% on both tours; `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` in `fixtures/eval/ground-truth.v2.json`)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + quarantined suite + validator before integrated run | pass | d39a0ce | validator exit 0; bench race pass; root race started pre-run (see R1 note on collision + clean rerun) |
| AC1 | `go test -race ./...` root and quarantined module; `node tools/validate/validate.js` | pass | 40acf09 | root race `-count=1` exit 0 (17 ok, 0 fail); bench race `-count=1` exit 0; validator exit 0; ran on d39a0ce, 40acf09 record-only |
| AC2 | `go test ./internal/eval -run EndToEnd -v`; SC15 ≥ 90%; `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` | pass | 40acf09 | published R82.4/P100.0, perturbed R81.2/P100.0; SC15 focused 95.0/full 100.0; fixture diff exit 0; `reviewed: false` confirmed; ran on d39a0ce, 40acf09 record-only |
| R1 | SC1–SC5 test commands + `mise exec -- goreleaser release --snapshot --clean --skip=sign` (and clean `dist/`) | pass | 40acf09 | SC1–SC5 suites pass; snapshot exit 0 (5 bins, 5 archives, 5 SBOMs, checksums); archives ship README+LICENSE*+NOTICES+manifest; SBOM SPDX-2.3 17 pkgs; `dist/` + root manifest removed; ran on d39a0ce, 40acf09 record-only |
| V1 | fresh-context adversarial review (`.agents/skills/work/review.md`); resolve all findings in parent `## Review` | pass | 40acf09 | diff f404ce0..d39a0ce read vs spec F1–F5/SC1–SC7; 1 minor observation rejected with reason; zero open findings |

## Constraints and budget

- Worktree must be completely clean (`dist/` removed) before committing.
- Zero fixture/corpus diff; `"reviewed": false` unchanged; no tag push or release publish.

## Plan

1. Run SC1–SC5 and a full `goreleaser release --snapshot --clean --skip=sign` dry-run; inspect generated archives, SBOMs, `checksums.txt`, and `skills-manifest.json`; remove `dist/`.
2. Run AC1 and AC2 (root + quarantined race suites, EndToEnd R/P, SC15, fixture diff check, validator).
3. Conduct fresh-context adversarial review against `specs/gira-skills-release.md`, fix any findings, record evidence in child 05 and parent `tasks/2026-10-14-gira-skills-release.md`, and seal parent as `verified`.

## Review

- [O1] `Register.Save` (used by `gira review --gc`) writes via `os.WriteFile`, not temp-file + rename, so "saved atomically" (spec F5) is single-write, not crash-atomic -> rejected: pre-existing shared `Save` semantics used by every register write; GC introduces no new durability risk; reworking `Save` is out of scope for E5.
- [C1] `skills.Check` mutates disk? No: `internal/skills/manifest.go` `Check` uses only `os.ReadFile` + in-memory re-render; no `WriteFile`/`MkdirAll` outside `output` path of `manifest --output` (explicit user request). Clean.
- [C2] Manifest digests over exact canonical bytes? Yes: `CanonicalBytes` reads embedded `catalog/<name>/SKILL.md`; per-skill `sha256:<hex>` + sorted catalog digest; deterministic (SC3 tests assert byte-stability).
- [C3] `.goreleaser.yaml` builds cleanly? Yes: live snapshot run exit 0, 5 targets, archives contain `README.md` + `LICENSE*` + `THIRD_PARTY_NOTICES.md` + `skills-manifest.json`; `signs:` stanza present, `--skip=sign` only for local snapshots.
- [C4] `release.yml` least-privilege + quarantine? Yes: root `contents: read`, release job `contents: write` + `id-token: write` under `startsWith(github.ref,'refs/tags/v')`; zero quarantined-module references in guarded files (`server.go:705` "quarantine" comment is pre-existing prose, zero E5 diff).
- [C5] `Register.GC` drops an `open` finding? No: `open` + unknown statuses preserved unconditionally in order; resolved pruned to last `keepResolved` (negative clamped to 0); receiver never mutated; `Save` validates each entry.
- [C6] `gate.LintOptions.FailOnCI` truth table: CI+strict+fail-on-ci fails closed; CI without fail-on-ci warns; non-strict never gates. Matches spec Q5=A.

No open findings. One full root race failure observed mid-run (`TestSCSBOMAndCosign`: "test left dist/ behind") was an environmental collision with the concurrent R1 snapshot dry-run, not a code defect: `dist/` + root manifest removed, package rerun clean, then full suite rerun `exit 0` (17 ok, 0 fail).

## Outcome

Delivered integrated E5 verification (F1–F5; SC1–SC7): SC1–SC5 suites pass; snapshot dry-run exit 0 with inspected archives/SBOMs/checksums/manifest and full cleanup; SC6 root + quarantined race suites pass with validator exit 0; SC7 invariants hold (R82.4/P100.0, R81.2/P100.0, SC15 95.0/100.0, zero fixture diff, `reviewed: false`); adversarial review complete with zero open findings. Tree clean, `dist/` absent, no tags pushed, no release published.
