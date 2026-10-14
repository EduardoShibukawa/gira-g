---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-human-surface
spec: gira-human-surface@1
verified_at: ""
---

# Indexed commit and staleness (F3; SC3)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-human-surface; spec v1 approved)
- depends_on: ["2026-10-14-gira-human-surface-01-plain-renderer-and-tty"]

## Objective

Every human command prints the indexed commit SHA and flags a component/edge
`stale` when its locators cite a commit other than the indexing commit; when the
target repo path resolves, the local HEAD is compared too. A one-line banner
states the comparison. No false `stale`/`current`.

## Acceptance criteria

- AC1 = SC3 (`SCStaleness`: indexed commit printed; locator/HEAD mismatch flags `stale`)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + bench + validator before change | not run | | |
| AC1 | `go test ./internal/domain ./cmd/gira -run SCStaleness -count=1` | not run | | |
| R1 | `go test ./internal/domain ./cmd/gira -count=1` | not run | | |
| V1 | diff read against AC1; show diff before commit | not run | | |

## Constraints and budget

- Staleness is honest: no false positive/negative on the indexed commit.
- HEAD comparison only when the repo path resolves (skip silently otherwise).
- Additive to the human output; `--format json` unchanged.
- Deterministic output; zero fixture diff.

## Plan

1. Failing tests for the indexed-commit accessor and the stale flag.
2. Implement the domain helper + banner + per-node/edge flag.
3. Wire the styled/plain renderer from child 01.
4. Run AC1/R1/V1; commit.
