---
schema: task/1
tier: standard
status: verified
parent: 2026-10-14-gira-recall-joins
spec: gira-recall-joins@1
verified_at: "3a39829"
---

# F7: unsupported construct vs unsupported file

## Routing and approvals

- tier: standard (review follow-up F7 of parent 2026-10-14-gira-recall-joins; touches domain + extractors pipeline + slices + eval safety, additive index field)

## Objective

A Go file that WAS parsed but contains an unsupported construct (J3 OpAMP
row) landed in `mm.Unsupported` and was counted by export and
`LocateBoundary` as an Unsupported Boundary, as if the whole file were in an
unsupported language. Separate "unsupported construct inside a supported
file" from "unsupported file": the parsed file must not count as an
Unsupported Boundary in export / `LocateBoundary`, while the briefing keeps
its explicit unknown row (J3: never a silent miss).

Out of scope: `layer2_compose*.go`, `resolve*.go`, `internal/eval/posture_test.go`,
fixtures/testdata/ground truth (unchanged; `reviewed` stays false).

## Acceptance criteria

- AC1: a parsed Go file with an unsupported construct is not an Unsupported Boundary in export (md + html) or `LocateBoundary`; it is still surfaced (export section "Unsupported Constructs in Parsed Files") and the impact briefing keeps its unknown row with locator.
- AC2: no regression: EndToEnd published ≥ R82.4/P100.0, perturbed ≥ R81.2/P100.0 (regression guards: R1 SC15 assertiveness ≥ 90%, R2 race suites green, R3 validator exit 0).

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | validator before change | pass | 9ba80a6 | exit 0 |
| RED | `go test ./internal/slices -run TestUnsupportedConstructIsNotUnsupportedFile` before fix | fail | 9ba80a6 | md header "Unsupported Boundaries: 2" (want 1), md/html section lists `src/catalog/main.go`, `LocateBoundary` catalog/main.go/OpAMP returned it as `unsupported`; briefing row assertion already held |
| AC1 | new tests `TestUnsupportedConstructIsNotUnsupportedFile` (slices), `TestPipelineSeparatesConstructFromFile` (extractors, incl. canonical JSON round trip), `TestUnsupportedSafetyIgnoresConstructRows` (eval) + `go test ./internal/eval -run SC3UnsupportedPosture -v` | pass | 3a39829 | SC3: opamp-server unknown at `src/product-catalog/main.go` / `src/inventory/main.go`, all 14 watched rows present |
| AC2 | `go test ./internal/eval -run EndToEnd -v` | pass | 3a39829 | published R82.4/P100.0/TR78.4, perturbed R81.2/P100.0/TR80.2; safety/determinism 100% |
| R1 | `(cd bench && go test ./layer2 -run SC15 -v)` | pass | 3a39829 | assertiveness 95.0% (19/20, --focus service:frontend) and 100.0% (20/20) |
| R2 | `go test -race ./...` root + bench/ | pass | 3a39829 | root all ok (gira-query goldens unchanged); bench all ok — bench/layer0 needs gitignored `testdata/gira/macro-map.json`, absent in the worktree, copied from main checkout for the run and removed (untracked, not committed) |
| R3 | `node tools/validate/validate.js; echo $?` | pass | 3a39829 | exit 0 |
| V1 | diff read against AC1–AC2 + spec J3 | pass | 3a39829 | only in-scope files touched; no fixtures/testdata/gt change; briefing rows unchanged (goldens byte-identical); no weakened assertion |

## Plan

1. Red test (export + LocateBoundary + briefing) on real pipeline output.
2. Additive `MacroMap.UnsupportedConstructs` set by the pipeline for rows an Extractor emits; `UnsupportedFiles()` / `UnsupportedConstructRows()` helpers.
3. Export, `LocateBoundary`, `UnsupportedSafety` use unsupported files only; briefing unchanged.

## Outcome

- `domain.MacroMap.UnsupportedConstructs` (additive, omitempty, canonical-sorted) + `IsUnsupportedConstruct`, `UnsupportedFiles`, `UnsupportedConstructRows`. Rows stay in `Unsupported` so existing consumers and the briefing JSON (gira-query goldens) are byte-identical.
- Pipeline marks every boundary returned by an Extractor (file parsed) as a construct.
- Export md/html count and list only unsupported files; constructs get a separate "Unsupported Constructs in Parsed Files" section (explicit unknown, not silent).
- `LocateBoundary` and `UnsupportedSafety` ignore construct rows.
- Not changed: `archdiff` still diffs all rows by path (out of scope).

- Linearized 2026-10-14 by the orchestrator: originally verified at d457a2a (on 9ba80a6) in an isolated worktree; cherry-picked on top of F4+F6 as 3a39829 and re-verified there by the orchestrator in the main checkout: validator exit 0, race root ok, bench layer0/1/2 race ok, EndToEnd + SC1–SC3 pass (published R82.4/P100.0, perturbed R81.2/P100.0), SC15 95.0%/100.0%. Scope note: internal/slices/export.go was outside the declared scope (export lives there); accepted, no collision with other workers. bench/quarantine re-run at the final gate.
