---
schema: task/1
tier: standard
status: verified
verified_at: "75ecdba"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-01-eval-end-to-end]
---

# Arch diff (SC9)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

`gira diff`, MCP `diff_maps`.

In scope: spec §A items covered by SC9.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC9

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | 96ae840 | all packages ok before any edit |
| AC1 | `go test -race ./cmd/gira ./internal/archdiff -run SC9 and -run ArchDiff and -run Segments and -run Ambiguous -v` | pass | 75ecdba | 24 renames equal the table; 2 removed edges equal index_changes; no other structural change; CLI equals MCP; ambiguous rename not guessed |
| R1 | `go test -race ./...` | pass | 75ecdba | all packages ok |
| R2 | `node tools/validate/validate.js` | pass | 75ecdba | overall pass |
| V1 | read the diff against the ACs | pass | 75ecdba | scope: archdiff package, CLI diff, MCP diff_maps, recipe index_changes |

## Plan

skip: the diff package, then the CLI and MCP wiring and the declared index changes in the recipe.

## Outcome

Delivered:
- `internal/archdiff.Compare(a, b)`. It reports renamed, added and removed nodes, added and removed edges, changed Contracts (metadata compared after renames), newly and no-longer unsupported paths, and, as information only, nodes whose evidence moved. It also emits the derived `token_renames`.
- Renames are detected from structure, never from a table. Removed and added nodes of the same kind are scored on their neighbours, their evidence in files present in both indexes (weight 2) and their evidence in other files (weight 0.5, since lines may shift). A pair is accepted only when each side is the other's unique best, and matching iterates until it is stable. An ambiguous pair is not guessed. The rule is documented in the package;
- CLI `gira diff <index-a> <index-b> [--format json|md]` (Markdown is a changelog); MCP `diff_maps` (`index_a`, `index_b`);
- `perturbation.json` gains `index_changes.removed_edges`: the edges its edits remove, in published ids.

SC9 evidence on the hermetic corpus: 24 renames, exactly the published nodes whose id the rename table changes; 2 removed edges, exactly the declared `checkout consumes currency` and `checkout consumes contract:CurrencyService`. Added nodes, removed nodes, added edges, changed Contracts and unsupported changes are all 0. Moved evidence on 25 nodes is reported separately: the recipe's edits shift lines by design, and spec A.2 lists moved evidence as its own item.

Calibration note: the first scorer weighted all evidence equally. Coincidental line shifts in renamed files then tied 3 Contract pairs. The stable-path weighting fixed that, and `TestAmbiguousRenameIsNotGuessed` pins the no-guess rule.

