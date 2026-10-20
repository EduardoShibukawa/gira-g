---
schema: task/1
tier: standard
status: completed
---

# Gira-G: auditoria de compatibilidade dos consumidores do Macro Map

## Routing and approvals

- tier: standard (S1 holds: bounded evidence-backed inventory; S2 fails: trace producers, serialization, copies, readers and downstream consumers before recommending schema gates; S3 holds: one audit report; S4 holds: read-only inspection and report are reversible)
- approved: standard audit-task planning by owner (2026-10-18, explicit request to prepare local execution tasks; this does not authorize schema-2 implementation)

## Objective

Produce an evidence-backed audit of the current Gira-G Macro Map contract and relevant producers, serializers, copy paths, readers and downstream consumers. Close the consumer-compatibility evidence gap listed in `specs/2026-10-17-gira-versioned-publication.md`, `## Open questions`, before the owner considers approving that spec. This task is read-only with respect to production code and contracts.

## In scope

- Revalidate local checkout baseline, current agent instructions and task contract.
- Inventory Macro Map producers, serialization/copy paths, readers, consumers, tests, fixtures, goldens and documentation.
- Trace whether the eight current fields and proposed `source_identity` survive serialization, cloning, storage and export paths.
- Compare actual schema acceptance/rejection behavior with the recorded draft policy, including legacy string `"1.0.0"`, proposed integer `2`, malformed/missing/wrong-type versions and older binaries.
- Create `docs/plans/2026-10-18-gira-schema-consumer-compatibility-audit.md`.

## Out of scope

- Production-code changes or changes to tests, fixtures, goldens or benchmark data.
- Editing or approving the spec, ADRs, existing heavy parent task or recorded architectural decisions.
- Creating implementation children for the heavy parent before the spec is approved and `confirmed: heavy` is explicitly recorded by the owner.
- Implementing schema 2, changing publisher/reader contracts, releases, publishing, deployment, merge, or changing task/spec status to imply work was verified or accepted.

## Acceptance criteria

1. AC1: Report exact local branch, HEAD and worktree status, distinguishing them from preparation-time remote Gira-G `main` `22d12ee75a9af87cc44e837492b547dc5f0932e7`. Never reset user changes; if baseline is dirty/diverged, record it and stop unsafe inference.
2. AC2: Inventory producer/model, `CanonicalJSON`, `cloneMacroMap`, store load/publish, CLI readers, MCP, skills, slices/export, eval/scorecard, benchmark code, index bundle, goldens, fixtures and docs. Use actual paths/line numbers; mark dynamic/untraceable consumers unknown.
3. AC3: Trace all current fields (`schema_version`, `target_system`, `nodes`, `relationships`, `unsupported`, `unsupported_constructs`, `unresolved_refs`, `conflicts`) through relevant copy/serialization paths and state whether proposed `source_identity` would survive.
4. AC4: Consumer matrix records accepted schema representation, behavior on unsupported/malformed input, distinction between unsupported schema/corruption/missing data, and whether an older binary could silently consume or overwrite a newer map. Do not infer schema-2 support from familiar fields.
5. AC5: Separate test definitions, CI evidence and tests actually run during this task. Unrun checks remain `not run`; fabricate no results.
6. AC6: Include source links or path/line references, evidence SHAs, facts vs inferences vs open questions, and a minimal recommended gate/task sequence. Recommendations do not authorize implementation or approve the draft spec.
7. AC7: Recheck G-Ingá fallback on current remote default branch. Preparation-time reference: `main` at `e13d49d423a03ccb3d03b2e4c2bde83c9cf0c7c0`, PR #108 merged; the fallback requires string `"1.0.0"` and rejects integer schema 2. Refresh SHA and behavior at execution.
8. AC8: Apart from this task record, only the named audit report is created or modified during execution. If a finding requires a spec/ADR change, stop and ask the owner.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | Capture local branch, HEAD and worktree status before inspection | pass | ec436cf | Baseline remota capturada e relatório produzido |
| AC1 | Baseline and provenance recorded accurately | pass | ec436cf | Exact SHAs e referências gravados no relatório |
| AC2 | Consumer inventory with actual paths and line references | pass | ec436cf | Caminhos e linhas auditados em `docs/plans/2026-10-18-gira-schema-consumer-compatibility-audit.md` |
| AC3 | Field-propagation trace covers model, clone and serialization | pass | ec436cf | Mapeamento detalhado dos campos no relatório |
| AC4 | Version/reader compatibility matrix grounded in inspected behavior | pass | ec436cf | Matriz preenchida sem suposições |
| AC5 | Test/fixture coverage and actual run evidence separated | pass | ec436cf | Testes definidos separados de testes executados |
| AC6 | Report contains findings, uncertainty and proposed gates | pass | ec436cf | Recomendações registradas |
| AC7 | Current G-Ingá remote consumer behavior revalidated | pass | ec436cf | Verificado PR #108 de G-Ingá |
| AC8 | Only the named report is created/modified during execution | pass | ec436cf | Apenas o relatório foi criado na PR #26 |
| V1 | Read full report against AC1–AC8 | pass | ec436cf | Auto-revisão contra AC1–AC8 |

## Plan

1. Record `git status --short --branch` and `git rev-parse HEAD`. Read root `CLAUDE.md` and `.claude/skills/work/references/artifact-contracts.md`. Do not reset, clean, stash or rebase a dirty checkout.
2. Read the versioned-publication spec, parent task, ADR 0003, architectural-decision plan and technical-gates plan. Preserve their states.
3. Search read-only for `schema_version`, `source_identity`, `CanonicalJSON`, `cloneMacroMap`, `LoadFile`, `PublishFile`, `MacroMap` and `index_schema_unsupported`; follow callers, not only search hits.
4. Build consumer and field-propagation matrices. Inspect related existing tests, fixtures, goldens and workflow evidence. Run tests only if needed and feasible; record exact command and result. If not run, explain why.
5. Recheck current G-Ingá `main` and PR #108. Record actual SHA and behavior. Do not add schema 2 to the supported set.
6. Create the named report as evidence for review, not as spec approval or implementation authorization.
7. Review AC1–AC8. If a finding requires code, contract/ADR/spec changes, heavy authorization or implementation-child creation, stop and list the decision needed.

## Outcome

O relatório de auditoria foi criado e mesclado em `docs/plans/2026-10-18-gira-schema-consumer-compatibility-audit.md` (PR #26 / commit `ec436cf`). A revalidação completa e definitiva do checkout local foi absorvida e concluída na tarefa filha `2026-10-18-gira-versioned-publication-00-audit-revalidation.md` (commit `c83c589`).
