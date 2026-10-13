---
schema: task/1
tier: standard
status: completed
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
---

# Bench quarantine (SC12)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Benchmark workspace outside the Go module: container recipes, two-phase runs, ignored output, static isolation check.

In scope: spec §B items covered by SC12.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC12

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | fd35d79 | all packages ok before any edit |
| AC1 | `go test -race ./...` run inside `bench/` (quarantine package) | not run | | |
| R1 | `go test -race ./...` at the repository root; the bench module is not entered | not run | | |
| R2 | `node tools/validate/validate.js` | not run | | |
| V1 | read the diff against the ACs | not run | | |

## Constraints and budget

- No image was built or run, and no tool was installed. Pins were resolved with read-only metadata queries: `git ls-remote` for the Graphify tag, and registry `HEAD` requests for image digests.

## Plan

skip: the workspace files, then the static check and its tests.

## Outcome

Delivered `bench/`, a separate Go module (`github.com/EduardoShibukawa/gira-g/bench`). The root `go test ./...` and CI never enter it.
- `containers/graphify/Dockerfile` builds Graphify v0.9.82 from commit `5b74d7d7…`. `containers/understand-anything/Dockerfile` builds Understand Anything from commit `ffa2f0a8…` (plugin 2.9.7). Base images and the Dockerfile frontend are pinned by digest, and sources by full commit through `ADD <git-url>#<commit>`. Remaining exposure: Graphify's Python dependencies resolve from PyPI at build time, and `pip freeze` is recorded in the image for each report. UA installs from its frozen pnpm lockfile.
- `phases.json` declares each arm's build and query phases. Every query phase is `network: none`. Containers mount only the read-only corpus copy and `out/<arm>`. Arm D is marked `paid_indexing`.
- `scripts/container.sh` builds and queries a container arm, run by the owner only. The query phase is `docker run --network none --read-only --cap-drop ALL`.
- `.gitignore` ignores `out/`.
- `quarantine/` holds the static check, also run as tests. It checks that no guarded file (Go module, `mise.toml`, CI, git hooks, agent configuration) references `bench/`, `gira-bench`, Graphify or Understand Anything; that `out/` is ignored; that query phases are `none` and mounts are limited; and that recipes are pinned. `TestCheckCatchesViolations` proves it catches each case.

Not done, by design: building the images or indexing with the external tools. Graphify's build command and output path come from the prior-art study, and `phases.json` says to confirm them on the first build. UA's indexing is a paid action that needs its own owner approval.

