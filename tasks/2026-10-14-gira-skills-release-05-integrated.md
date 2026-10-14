---
schema: task/1
tier: standard
status: todo
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
| B1 | `go test -race ./...` (root) + quarantined suite + validator before integrated run | not run | | |
| AC1 | `go test -race ./...` root and quarantined module; `node tools/validate/validate.js` | not run | | |
| AC2 | `go test ./internal/eval -run EndToEnd -v`; SC15 ≥ 90%; `git diff --exit-code f404ce0 -- fixtures/ testdata/`; `"reviewed": false` | not run | | |
| R1 | SC1–SC5 test commands + `mise exec -- goreleaser release --snapshot --clean --skip=sign` (and clean `dist/`) | not run | | |
| V1 | fresh-context adversarial review (`.agents/skills/work/review.md`); resolve all findings in parent `## Review` | not run | | |

## Constraints and budget

- Worktree must be completely clean (`dist/` removed) before committing.
- Zero fixture/corpus diff; `"reviewed": false` unchanged; no tag push or release publish.

## Plan

1. Run SC1–SC5 and a full `goreleaser release --snapshot --clean --skip=sign` dry-run; inspect generated archives, SBOMs, `checksums.txt`, and `skills-manifest.json`; remove `dist/`.
2. Run AC1 and AC2 (root + quarantined race suites, EndToEnd R/P, SC15, fixture diff check, validator).
3. Conduct fresh-context adversarial review against `specs/gira-skills-release.md`, fix any findings, record evidence in child 05 and parent `tasks/2026-10-14-gira-skills-release.md`, and seal parent as `verified`.
