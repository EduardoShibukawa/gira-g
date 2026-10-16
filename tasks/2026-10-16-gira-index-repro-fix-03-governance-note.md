---
schema: task/1
tier: standard
status: todo
parent: 2026-10-16-gira-index-repro-fix
spec: 2026-10-16-gira-index-repro-fix@1
---

# Index repro fix 03: governance view note

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-16-gira-index-repro-fix)

## Objective

Document the governance filter in `gira slice --view` help
(`cmd/gira/main.go`) and `website/docs/cli/slice.md` (SC3).

In scope: flag help text, slice.md note.
Out of scope: engine behavior change, MCP surface.

## Acceptance criteria

1. AC1: `--view` help states governance keeps only Policy/Task/Skill
   neighbors; `go test ./internal/slices/ ./cmd/gira/` passes.
2. AC2: slice.md note matches help; `mise run docs:build` passes.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | parent B1 (`go test -race`, pre-children) | not run | | cited from parent; rerun in 04 |
| AC1 | help text + package tests | not run | | |
| AC2 | slice.md note + docs build | not run | | |
| V1 | diff read against AC1–AC2 | not run | | |
