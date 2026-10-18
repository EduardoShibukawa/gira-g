# Gira-G Macro Map schema and consumer compatibility audit

**Audit date:** 2026-10-18 (UTC timestamps from GitHub; user-facing timezone America/Sao_Paulo)  
**Scope:** Read-only inspection of the current remote default branches and their published CI evidence. No production code, tests, fixtures, specs, ADRs, task states, or benchmark inputs were changed.

## Executive result

**Verdict: the schema-2/versioned-publication work is not ready for spec approval or implementation.** The newly merged G-Ingá fallback validation is verified and correctly rejects integer schema `2`; the producer/reader contract in Gira-G is still the old string-schema implementation. The schema-consumer audit task is therefore not complete, and no acceptance claim is made for schema 2.

| Repository | Remote `main` SHA inspected | Current evidence |
|---|---|---|
| [Gira-G](https://github.com/EduardoShibukawa/gira-g) | `5a24a361841fdbe0ef85ce41987d2a17ba87b45b` | Post-merge CI run [39090339780](https://github.com/EduardoShibukawa/gira-g/actions/runs/39090339780) completed successfully on Ubuntu, macOS and Windows, and the validator, floor guard and warn-only gate-lint also succeeded. Documentation run [39090339761](https://github.com/EduardoShibukawa/gira-g/actions/runs/39090339761) succeeded. These green checks validate the current existing code and docs, not the proposed schema-2 implementation. |
| [G-Ingá](https://github.com/EduardoShibukawa/G-Inga) | `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0` | Main CI run [39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098) succeeded for both `test` and `pre-commit`/validation. The strict fallback is in main via [PR #108](https://github.com/EduardoShibukawa/G-Inga/pull/108). |

### Local-baseline limitation

The audit runtime contained no Git checkout (`git status` returned “not a git repository”). A fresh `git ls-remote` attempt failed because this runtime could not resolve `github.com`. Consequently, **the developer's local branch, HEAD, and worktree cleanliness were not observable**. This report distinguishes remote GitHub evidence from local execution; it does not claim a local test run or a clean user worktree. The local-baseline part of AC1 remains unverified.

## Verified findings in Gira-G

### 1. The model and persistence layer still implement the old contract

- `internal/domain/models.go:145-159` defines `MacroMap` with the eight existing fields. It has no `source_identity` field. `NewMacroMap` still emits `schema_version: "1.0.0"` (`models.go:196-204`). [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/internal/domain/models.go#L145-L204)
- `MacroMap.Validate()` only requires a non-empty schema-version string, target system, and structurally valid graph; it does not enforce an explicit schema allowlist. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b/internal/domain/models.go#L206-L253)
- `CanonicalJSON()` reconstructs the serialized model field by field (`models.go:286-347`). It serializes the current known fields but cannot preserve `source_identity`, because the model has no such field. Future fields added to `MacroMap` must also be deliberately propagated through this reconstruction. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b/internal/domain/models.go#L286-L347)
- `internal/store/store.go:83-224` publishes one JSON file by temporary-file write, file sync, rename and parent-directory sync; `LoadFile()` JSON-decodes into the current struct and invokes `Validate()`. There is no versioned immutable store/head protocol or strict schema reader here. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b/internal/store/store.go#L83-L224)

**Compatibility implication (inference from the inspected implementation):** an integer `schema_version: 2` fails while JSON is being decoded into the string field and is classified as corrupt/invalid data, not the required stable `index_schema_unsupported`. An unknown but non-empty string version can pass the current schema check when the remaining graph is valid. Neither behavior satisfies the proposed R2 reader contract.

### 2. Incremental copying demonstrably drops two existing fields

`internal/indexer/incremental.go:706-754`, `cloneMacroMap()`, copies `SchemaVersion`, `TargetSystem`, nodes, relationships, unsupported rows and `UnsupportedConstructs`, but does not copy `UnresolvedRefs` or `Conflicts`. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/internal/indexer/incremental.go#L706-L754)

This is a confirmed field-propagation gap in current code. The future `source_identity` field would also need to be copied once introduced. Tests must explicitly prove preservation of all eight current fields plus the proposed field through clone and canonical serialization; a field-to-field checklist or equivalent regression should detect future omissions.

### 3. Existing reader and error paths do not implement the proposed schema policy

- The CLI's `loadIndex()` routes explicit `--index` reads through `store.LoadFile()` (`cmd/gira/main.go:300-310`). Project discovery still searches the current legacy index locations (`main.go:312 onward`). [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/cmd/gira/main.go#L300-L340)
- MCP budget/diff loading calls `LoadFile()` and maps load failures to the generic `index_missing` structured error (`internal/mcp/server.go:624-632`), which does not distinguish unsupported schema from a missing file or corruption. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b/internal/mcp/server.go#L624-L632)
- `internal/skills/install.go` stores an explicit index path to pass to `gira serve mcp --index`; it does not itself provide schema validation. Further end-to-end coverage is needed for `skills install/check`, MCP, bench, `tools/index-bundle`, `internal/benchtest`, goldens, fixtures, scorecard/eval paths and documentation as required by SC9.

### 4. Git path parsing remains fragile for unusual names

`internal/gitcache/cache.go:146-192` invokes `git ls-tree -r --full-name` without `-z` and parses newline-separated output with tab splitting. That is not a NUL-delimited pathname protocol and does not meet the proposed adversarial coverage for tabs/newlines and Git quoting. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/internal/gitcache/cache.go#L145-L192)

`ExportTree()` starts `tar`, then returns immediately if `archiveCmd.Run()` fails before waiting for `tarCmd` (`cache.go:226-253`). Cleanup/reaping behavior on this failure path remains a code-level risk to fix and test. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/internal/gitcache/cache.go#L226-L253)

### 5. Git environment isolation and `init` ordering are not fixed

- `gitToplevel()` and `gitDirty()` execute Git commands with the inherited environment (`cmd/gira/main.go:569-582`). The required test matrix for `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, `GIT_INDEX_FILE`, object-directory variables and related redirection is not evidenced. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/cmd/gira/main.go#L569-L582)
- `runInit()` writes or repins `gira.yaml` before calling the shared indexing path (`main.go:529-566` and `writeLocalManifest()` at `main.go:585-614`). A later build/index/gate failure can therefore leave a new or repinned manifest behind. This contradicts the proposed SC8 rule that manifest publication happens last. [Source](https://github.com/EduardoShibukawa/gira-g/blob/5a24a361841fdbe0ef85ce41987d2a17ba87b45b/cmd/gira/main.go#L529-L566)

### 6. Versioned publication is not present in the current implementation

The current store writes a mutable destination file; the inspected model/store/CLI paths do not implement `source_key`, `tool_id`, immutable `version_id` directories, lineage heads, `source_identity`, strict `meta.json`, the M→S→P lock protocol, or the stable errors/state specified in SC3–SC8. `syncDirectory()` returns success without syncing on Windows, but the current writer does not surface the proposed `visible_unconfirmed` state/warning contract. These are implementation gaps, not tests that can be declared passed merely because the pre-existing `go test` suite passes.

The existing `runInit()` manifest-before-index behavior and Windows path/durability concerns are also called out by the current spec and decision records; they remain open until changed and tested.

## Cross-repository consumer result

G-Ingá's `tools/eval/evaluate_girag_slicing.ts:266-320` now explicitly accepts only string `schema_version === "1.0.0"`, requires arrays for `nodes` and `relationships`, validates node IDs and relationship endpoints, and reports unsupported schema separately from malformed JSON/structure. Its test file contains regression cases for integer schema 2, other version representations, malformed data and the same rejection after either a missing CLI or failed CLI invocation. [Implementation](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/tools/eval/evaluate_girag_slicing.ts#L266-L320) · [Tests](https://github.com/EduardoShibukawa/G-Inga/blob/e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0/tools/eval/evaluate_girag_slicing.test.ts#L188-L275)

This consumer is intentionally **v1-only**. It does not prove schema-2 support and must not be widened until the Gira-G producer contract and the cross-consumer matrix are approved. Main's `test` and `pre-commit`/validation jobs passed at `e13d49d` (run [39089480098](https://github.com/EduardoShibukawa/G-Inga/actions/runs/39089480098)).

## Compatibility matrix (current remote source inspection)

| Path | Observed acceptance/behavior | Result against proposed schema-2/R2 contract |
|---|---|---|
| Gira-G producer/model | String `"1.0.0"`; no `source_identity` | **Not implemented** |
| `CanonicalJSON` | Hand-built projection of current model fields | **New field will not persist until explicitly represented and serialized** |
| `cloneMacroMap` | Drops `UnresolvedRefs` and `Conflicts` today | **Confirmed field-loss bug; source identity also needs propagation** |
| Gira-G `Store.LoadFile` | JSON into string-typed model + structural validation; no version allowlist | **Integer 2 becomes decode/corruption path; unknown strings are not rejected by schema allowlist** |
| Gira-G MCP reader | Any load error can be reported as `index_missing` | **Error taxonomy insufficient for `index_schema_unsupported`** |
| G-Ingá slicing-evaluator fallback | Accepts only string `"1.0.0"`; rejects integer 2 and invalid structures | **Correct for today's approved/verified v1 consumer contract; schema 2 intentionally unsupported** |
| Remaining readers: bench, bundle, skills/check, scorecard/eval, fixtures/goldens/docs and dynamic callers | Full call-chain execution was not possible; individual direct/indirect readers need inventory and compatibility cases | **Open/unknown; do not infer compatibility** |

## Acceptance-criteria status for the audit task

| ID | Status | Evidence / limitation |
|---|---|---|
| B1 / AC1: local branch, HEAD and worktree | **blocked / not observable** | No checkout existed in this runtime; GitHub DNS failure prevented creating a fresh clone. Remote SHAs are recorded above; they are not a statement about the developer's local worktree. |
| AC2: producer/serializer/clone/store/CLI/MCP/skills/bench/bundle/eval/goldens/docs inventory | **partial** | Central paths and some readers traced above; bench/bundle/scorecard and dynamic consumers still need full call-chain inventory. |
| AC3: eight current fields and proposed `source_identity` | **partial, key losses verified** | `CanonicalJSON` carries current explicit fields; `cloneMacroMap` omits `UnresolvedRefs` and `Conflicts`; `source_identity` is absent from the model and so is not preserved today. |
| AC4: strict version and old-reader compatibility matrix | **partial** | Current Gira-G/G-Ingá behavior inspected, but the full consumer inventory and downgrade/no-overwrite tests do not exist for schema 2. |
| AC5: separate tests defined, CI and tests run in this audit | **pass for evidence separation** | No local tests were run in this audit. Existing remote CI evidence is linked above; SC1–SC10 remain `not run`. |
| AC6: evidence, facts/inference and minimal sequence | **pass in this report** | Source links, line references and SHAs included; implementation remains unauthorized. |
| AC7: G-Ingá fallback on current main | **pass** | Revalidated at `e13d49d`; code requires string `"1.0.0"`, rejects integer 2; main CI passed. |
| AC8: only named audit report modified | **pass in this audit scope** | Read-only inspection only; no source, tests, fixtures, spec, ADR or task record edited. |
| V1: final self-review against AC1–AC8 | **partial** | Report checked against known criteria, but AC1–AC4 are not fully satisfied; do not mark the task verified or change its `todo` status. |

## Remaining gates, in priority order

1. Finish the full read-only consumer inventory on an accessible clean checkout; record local branch/HEAD/worktree if one is made available. Identify unknown or dynamic consumers rather than assuming they are safe.
2. Resolve the confirmed propagation defects in the authorized implementation stage: carry `UnresolvedRefs` and `Conflicts` through incremental clones; add `source_identity` to the model, clone and serializer; add regression coverage for all fields.
3. Implement the explicit schema reader policy and error taxonomy (`index_schema_unsupported` vs `index_corrupt` vs `index_missing`), including fail-closed downgrade behavior and tests against old/new binaries.
4. Implement and test the manifest-root policy, environment sanitization, deterministic NUL-delimited Git-tree parsing, `tool_id` rules and evaluator's no-fabricated-manifest behavior.
5. Add test coverage for `git archive | tar -x` with tab-containing names on Windows, plus justified skips for symlink cases unavailable to a standard user.
6. Only after Gates 1–4 have reviewed evidence, ask the owner to approve the spec and explicitly authorize the heavy implementation. Do not create implementation children or mark `confirmed: heavy` by inference. Gate 5 durability/failure-publication tests are mandatory acceptance evidence for the first authorized implementation and must pass on Linux, macOS and Windows before that implementation is accepted.
7. Keep SC1–SC10, integrated CI, cache-intact check and integrated diff review marked `not run` until actual implementation tests execute on the exact integrated head.

## Final assessment

G-Ingá's strict fallback fix is complete and its main CI is green. Gira-G's new audit task and regression matrix are now on main as planning artifacts, but neither is proof that the proposed producer contract works. Two concrete current data-loss paths, version-reader behavior, Git path/environment safety, `init` transaction ordering, Windows-specific behavior and every SC1–SC10 implementation test remain unresolved. The spec must remain `draft`, the heavy parent must remain `todo`, and no schema-2 support should be advertised yet.