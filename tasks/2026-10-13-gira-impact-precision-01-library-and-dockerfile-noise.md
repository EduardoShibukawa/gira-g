---
schema: task/1
tier: standard
status: verified
verified_at: "6ee2c85"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
---

# Library and Dockerfile noise out of the Impact Set (cause 1)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Libraries reached through imports move to an additive `upstream_libraries` field; Dockerfile stages and base images become Service metadata instead of nodes.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC1
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval | pass | 44310f1 | published recall 58.8%, precision 32.8%; perturbed 56.2%, 30.5%; safety and determinism 100% |
| AC1 | `go test -race ./internal/eval -run SC1Impact` and `./internal/slices -run Libraries` | pass | 6ee2c85 | failed first on 44310f1 (180 findings: `service:stage:*` nodes, base images as `library:` nodes; `upstream_libraries` empty) |
| AC2 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | 6ee2c85 | published: recall 58.8% → 58.8%, precision 32.8% → 83.3%; perturbed: recall 56.2% → 56.2%, precision 30.5% → 81.8% |
| AC3 | same EndToEnd run, which indexes each variant twice from cold caches (SC8) | pass | 6ee2c85 | safety 100%, determinism 100% in both variants; go.mod unchanged, so no model client |
| AC4 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | 6ee2c85 | goldens regenerated, diff explained in the Outcome; no assertion weakened |
| AC5 | `node tools/validate/validate.js` (SC10) | pass | 6ee2c85 | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | 6ee2c85 | full corpus: recall 60.3% → 60.3%, precision 30.4% → 78.8% (reported only) |
| R2 | `go test ./layer2 -run SC15 -v` in bench/ | pass | 6ee2c85 | assertiveness unchanged: focused 5.9% (1/17), full 0/18 |
| V1 | read `git diff 44310f1..6ee2c85` against SC1 | pass | 6ee2c85 | scope: slices engine/budget/render/primpact, Dockerfile extractor and its tests, goldens |

## Outcome

Delivered:
- `upstream_libraries`, an additive field with `omitempty`, so outputs with no libraries stay byte-identical. It is on the impact JSON and Markdown, MCP and PR impact. Budget truncation drops it before any other node.
- A Dockerfile emits one Service, named by the directory rule and never by a stage name. Its `stages` and `base_images` are metadata entries, each with a `path#Lstart` locator.
- The summary tier strips `stages` and `base_images`, as it already strips contract methods. Otherwise checkout's focal node broke SC6's 500-token summary budget.

Golden diffs (`cmd/gira/testdata/gira-query`), all explained:
- **Build metadata.** Services carry their build metadata. The currency Dockerfile's final stage `AS release` used to be a separate `service:release` node. It now merges into `service:currency`.
- **One unsupported boundary fewer.** `impact-checkout` no longer lists `src/accounting/Consumer.cs`. It matched only through the old shared `service:stage:builder` node, which merged every Dockerfile's "builder" stage, accounting's included.
- **Rename scores.** The inferred rename scores in `diff-published-perturbed` changed, because the nodes now carry different evidence.
