---
schema: task/1
tier: standard
status: todo
parent: 2026-10-14-gira-envelopes-hygiene
spec: gira-envelopes-hygiene@1
verified_at: ""
---

# CI hygiene (R7, R8; SC7)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-14-gira-envelopes-hygiene; spec v1 approved)
- depends_on: ["2026-10-14-gira-envelopes-hygiene-01-envelope-all-tools", "2026-10-14-gira-envelopes-hygiene-02-envelope-honesty-and-ceiling", "2026-10-14-gira-envelopes-hygiene-03-structured-errors", "2026-10-14-gira-envelopes-hygiene-04-single-registry"]
- needs owner `approved:` before commit (ADR 0011:30; CI edits are protected)

## Objective

CI hygiene wired: bench race, validator, gate-lint warn-only on ubuntu-latest. Opt-in floor guard asserting sealed hermetic R/P and SC15 ≥ 90%. Hygiene fixes: gitignore results.tsv/run.log, fix README.md:22 broken link, bench/quarantine ignoring .claude/worktrees/, fix validator §6.1 false positive in non-linear history.

In scope: SC7, test-first. Failing test first: CI missing bench/validator/gate-lint must fail; floor guard not asserting must fail; hygiene issues must fail.
Out of scope: envelope features (01–04); ground-truth changes; .goreleaser (E5).

## Acceptance criteria

- AC1 = SC7 (CI hygiene)
- AC2 = SC8–SC10 (invariants + regression guard for the touched paths)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` (root) + `go test ./...` (bench) + validator before change | not run | | |
| AC1 | workflow greps for validate.js, gate-lint, `working-directory: bench`; `go test ./internal/gate -run SCGateLintStrictness -v -count=1` | not run | | |
| AC2 | `go test -race ./...` (root) + bench + validator, no E0 recall/precision dip | not run | | |
| R1 | `go test ./internal/gate ./bench/... -count=1` | not run | | |
| V1 | diff read against AC1–AC2 + spec R7, R8; **show diff and ask for `approved:` before commit** | not run | | |

## Constraints and budget

- CI runs bench + validator only on `ubuntu-latest` (Q3).
- `gate-lint --strict` in CI: warn-only (Q2); strict stays for later.
- Opt-in floor guard (separate from gt thresholds) asserting hermetic R82.4/P100.0, perturbed R81.2/P100.0, SC15 ≥ 90% (Q4).
- `.gitignore`: add `results.tsv`, `run.log` (Q6).
- `README.md:22`: fix broken link to `../../.scratch/system-context-harness/map.md` (Q6).
- `bench/quarantine/check.go`: skip `.claude/worktrees/` in `Guarded` or `References` (Q6).
- `tools/validate/validate.js`: fix false §6.1 positive in non-linear history (lines 7220, 7222, 7224, 7545) (Q6).
- `.goreleaser` stays for E5 (Q6).
- No model client (ADR 0001), no paid fan-out, quarantine ADR 0012.
- COPY = idea in Go, never pasted TS/JS; no ground-truth changes; no push/release.
- Hermetic published R82.4/P100.0, perturbed R81.2/P100.0 exactly; safety/determinism 100%; SC15 ≥ 90%.
- Zero diff in fixtures/ and testdata/; `"reviewed": false` stays.
- Unbudgeted calls byte-identical (gira-query goldens unchanged).
- Branch linear (cherry-pick, no parallel branch merges) until §6.1 fix lands.

## Plan

1. Write failing test for `SCGateLintStrictness` (warn-only in CI, strict fails closed locally).
2. Update `.github/workflows/ci.yml`:
   - Add bench job with `working-directory: bench` on `ubuntu-latest` only.
   - Add validator step (`node tools/validate/validate.js`) on `ubuntu-latest` only.
   - Add `gate-lint` step in warn-only mode.
   - Add opt-in floor guard step (env-gated, asserts hermetic R/P + SC15 ≥ 90%).
3. Fix `.gitignore`: add `results.tsv`, `run.log`.
4. Fix `README.md:22`: correct the broken link.
5. Fix `bench/quarantine/check.go`: skip `.claude/worktrees/` in `Guarded` or `References`.
6. Fix `tools/validate/validate.js`: §6.1 false positive in non-linear history (lines 7220, 7222, 7224, 7545).
7. **Show diff and ask for owner `approved:` before commit.**
8. Run AC1, AC2, R1, V1; commit; validator.

## Outcome