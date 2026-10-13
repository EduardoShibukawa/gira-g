---
schema: spec/1
status: approved
version: 1
approved_version: 1
approved_hash: sha256:dd4caed0ae65e14b16fbc86b3044ed58e2d256e644ab6fdbdeabacb75787d2c5
---

# GIRA Agent & Human Skills Surface and Comparative Evaluation Framework

## Problem

GIRA (`gira`) already gives deterministic, evidence-backed answers. The CLI has `index`, `impact`, `slice`, `explain`, `locate`, `export` and `view`. The stdio MCP server has four tools: `get_impact_briefing`, `get_context_slice`, `explain_evidence` and `locate_boundary`. These answers are delivered in three tiers (`summary`, `contracts`, `deep`) and three views (`full`, `architecture`, `governance`). There are four gaps.

1. **There are no packaged workflows (skills).** An agent or a human has to know which command to run, when to run it, and how to escalate between tiers. Prior art shows that the workflow matters as much as the engine:
   - Graphify succeeds with agents through explicit token budgets (`query --budget`), surgical impact tools (`affected`, `get_pr_impact`) and an always-on instruction block plus hooks. Its graph has no Component concept, its node IDs change when unrelated files are added, and its evidence is a single line number (docs/prior-art/graphify.md, Study 1).
   - Understand Anything succeeds with humans through guided tours (`TourStep`), domain flows (`/understand-domain`), on-demand explanations and a dashboard. Every edge except `imports` is emitted by an LLM, with no evidence, provenance or confidence (docs/prior-art/understand-anything.md, Study 1).
   - GIRA has neither surface. Change-makers and coding agents get raw commands but no guided entry points.
2. **The tiers are not budgeted.** ADR 0006 promises about 500, 1,500 and 3,500 tokens for `summary`, `contracts` and `deep`. Nothing enforces or reports these limits, and there is no truncation signal that would tell an agent to escalate.
3. **The internal evaluator is circular, so today's Go decision proves nothing about GIRA:**
   - The `MacroMap` under test is read from a hand-authored ground-truth fixture. It is never produced by the extractors. The extraction pipeline is not exercised at all.
   - Baseline token counts are per-task constants (`BaselineGrepTokens`), with a further 8,000 fallback. Harness tokens are `len(json)/4`, so token reduction compares a measured number against an invented one.
   - The Unsupported Safety metric is a loop with an empty body. It always returns 100%.
   - Determinism is measured by querying the same in-memory engine twice. Index determinism is never measured.
   - The ground truth has not been checked against the pinned corpus. Locators use placeholder ranges (`#L1-L100`) and paths that don't match the pinned OpenTelemetry Demo, such as a Go `main.go` for the TypeScript frontend.
4. **There is no comparative evidence.** ADR 0007 §1 excludes Graphify and Understand Anything from every evaluation pipeline. So the claim "agent efficiency like Graphify, human clarity like Understand Anything, plus determinism" cannot be tested.

Heavy triggers:

- **H1:** product calls on skill scope, how humans are evaluated, and when comparisons run.
- **H3:** the work touches eval, slices, MCP, CLI and the ADR set across subsystems.
- **H5:** it needs several sessions and decomposes into children.

## Expected behavior

The deliverable has three parts:

- **(A)** a skills catalog served on top of the deterministic engine;
- **(B)** a quarantined comparative benchmark against Graphify and Understand Anything;
- **(C)** a non-circular internal evaluator that runs the real pipeline end to end.

### A. Skills catalog

A **Skill** is a host-neutral instruction pack (a `SKILL.md` with a trigger description) bound to:

- one or more deterministic GIRA operations (CLI command and the equivalent MCP tool);
- a typed output contract;
- a token budget.

Skills obey three invariants.

- **I1 — Evidence or silence.** Every factual claim a skill emits carries a locator (`repo@sha:path#Lstart-Lend`) or an explicit `unsupported` / `unknown` marker. Skills never present an inferred relationship as extracted.
- **I2 — Narrate, don't infer.** The host agent's LLM may narrate over GIRA output (rephrase it, order it, empathize). Any narrated sentence that is not a direct rendering of a GIRA fact is labeled `narrated` and cites the locators it paraphrases. GIRA itself gains no model client (ADR 0001).
- **I3 — Same engine for both audiences.** Agent skills and human skills call the same operations. Only rendering, budget and ordering differ.

#### A.1 Agent-facing skills (Graphify-style efficiency)

| Skill | Purpose | Operations | Default budget |
|---|---|---|---|
| `gira-impact-briefing` | Pre-coding context. Before an agent edits code mapped to a Component or Contract, it gets the blast radius: upstream dependencies, downstream consumers, affected Contracts, mediation edges (broker/topic) and unsupported boundaries. It accepts a node ID, a file path or a git revision range (PR / working-tree diff). | `gira impact <target>`; `gira impact --from-diff <base>..<head>` (new); MCP `get_impact_briefing`, `get_pr_impact` (new) | 1,500 tokens |
| `gira-context` | Progressive disclosure. Fetches a slice at `summary`, then escalates to `contracts` or `deep` only when the truncation signal or the task requires it. Every response states its tier, its budget, the estimated tokens it used, and what was dropped. | `gira slice <node> --tier <t> --budget <n>` (budget is new); MCP `get_context_slice` with `budget` (new) | summary 500 / contracts 1,500 / deep 3,500 (ADR 0006) |
| `gira-query` | Structured lookup with graceful fallback. Tries the MCP server first. If no GIRA MCP tools are registered, it shells out to the CLI with `--format json`. Both paths return the same JSON. It defines the error vocabulary agents branch on: `index_missing`, `index_stale`, `node_not_found` (with `locate_boundary` suggestions) and `budget_exceeded`. | MCP tools, or the equivalent `gira <cmd> --format json` | per call |

Budget semantics, a new behavior shared by CLI and MCP:

- A budget is a **hard ceiling** on the estimated tokens of the rendered output. GIRA uses a documented, deterministic estimator based on bytes and structure, not a model tokenizer, so no model dependency is added (ADR 0001). The estimator is the same one the evaluator uses.
- When the full slice does not fit, GIRA truncates in a fixed, documented priority order:
  1. evidence records beyond the first per fact;
  2. neighbors by graph distance, then by ID;
  3. relationships to dropped neighbors.
  
  The focal node, its unsupported boundaries and the truncation marker are never dropped.
- Output carries `budget: {limit, estimated, truncated: bool, dropped: {nodes, relationships, evidence}, escalate_hint}`. Truncation is deterministic: the same index and the same budget give byte-identical output.

#### A.2 Human-facing skills (Understand-Anything-style clarity)

| Skill | Purpose | Operations |
|---|---|---|
| `gira-tour` | A guided architectural tour and reading roadmap for a newcomer, or for a Change-maker heading to Component X. Steps are computed deterministically: start from entry points (Components with declared ingress: exposed ports, HTTP routes), walk `consumes`/`provides` edges breadth-first, and rank ties by fan-in, then by ID. Each step has `{order, title, node_ids, read: [locators], why}`. `why` is a template rendered from the edges that justify the step, never free text. The host may add a narrated layer under I2. | `gira tour [--focus <node>] [--max-steps n] [--format md\|json]` (new); MCP `get_tour` (new) |
| `gira-explain-component` | A contextual dossier for one subsystem: what it is (kind, language, deployment profiles), Contracts provided and consumed, who depends on it, infrastructure, unsupported boundaries, and evidence. It also renders **deterministic flows**: chains from an entry Contract through `consumes` edges to infrastructure (for example HTTP route → frontend → CartService → cart → valkey). Business rules are not extracted. The skill points the host at the cited code ranges and labels any explanation of them as `narrated`. | `gira explain <node>` plus `gira slice --view architecture`; `gira flows <entry>` (new) |
| `gira-arch-diff` | Architectural evolution between two refs or releases: nodes, edges and Contracts added, removed or changed, evidence that moved, and newly unsupported paths. Rendered as a human changelog (Markdown) and as JSON for agents and CI. | `gira diff <index-a> <index-b>` (new); MCP `diff_maps` (new) |

#### A.3 Activation triggers and SDLC integration

| SDLC moment | Skill | How it activates |
|---|---|---|
| Routing a request (G-Ingá `/work`, signal S3 "footprint inside one module / no consumed contract change") | `gira-impact-briefing` (diff mode) | `/work` routing consults the Impact Set as deterministic evidence for S3/H3. A changed Contract with consumers outside the touched Component is H3 evidence. |
| Planning / before the first edit (`standard` plan step, `/drive` wave partitioning) | `gira-impact-briefing`, `gira-context` | skill description trigger ("before modifying code in an unfamiliar service…") |
| Implementation | `gira-query` | on demand, MCP-first |
| Verification / review (pre-commit, PR) | `gira-arch-diff` (base index vs head index) | `/work` verification row and an optional CI job that flags unintended Contract changes |
| Onboarding / reading | `gira-tour`, `gira-explain-component` | explicit invocation by a human, or a host slash command |

Host distribution comes from one canonical source per skill. `gira skills install --host <host>` generates the host files idempotently and edits only content inside its own markers:

- **Claude Code:** `.claude/skills/gira-*/SKILL.md`, plus MCP registration of `gira serve` in project MCP config. An opt-in `PreToolUse` nudge suggests `gira-impact-briefing` before edits to mapped paths. It nudges and never blocks (unlike Graphify's `hook-guard`).
- **Cursor:** a rules file plus MCP config.
- **Antigravity and other `AGENTS.md`-style hosts:** an `.agents/skills/gira-*` mirror, the same layout G-Ingá ships.

The contract between `/work` and GIRA is specified here. The change to the G-Ingá harness itself belongs to the G-Ingá repository and is out of scope (see Open questions).

### B. Comparative benchmark (quarantined)

Governed by ADR 0012, which amends ADR 0007 §1.

**Arms.**

- **A** — no harness: read-only grep/glob/read/bash.
- **B** — GIRA: MCP, with CLI fallback.
- **C** — Graphify, at a pinned version, through its MCP server.
- **D** — Understand Anything, at a pinned commit. The agent reads its `knowledge-graph.json` as its skills instruct.

Every arm gets the same base tools. Arms B to D add their tool's surface.

**Corpus.** OpenTelemetry Demo pinned at the ADR 0007 commit, plus the Helm chart, in two forms: as published and as the Perturbed Variant. Before any tool indexes the corpus, `README*`, `AGENTS.md` and `docs/` are removed for all arms (the ADR 0007 rule against memory leaks).

**Quarantine.**

- Each external tool runs in a disposable container built from pinned source and referenced by digest.
- Indexing happens in a build phase.
- The query phase runs with networking disabled, except for the model endpoint of the arm under test.
- Raw tool output and agent transcripts stay in an ignored output directory outside version control.
- Only normalized score reports are committed. Each report records tool versions, image digests, model IDs and corpus SHAs.
- Nothing in the benchmark is referenced from the Go module, `mise.toml`, CI, git hooks or agent configuration.
- Results never feed the go/no-go gate.

**Layer 0 — structural graph evaluator.**

- One adapter per tool maps its output into a **Normalized Architecture Graph (NAG)**:
  - Component nodes;
  - edges with a verb (`provides`, `consumes`, `imports`, `accesses`, `publishes`, `subscribes`), or `unknown` when the tool has no verb;
  - evidence `{path, start, end}`.
- Graphify and Understand Anything have no Component concept. The evaluator, not the tool, lifts file and symbol nodes to Components. It uses a path→Component table that is part of the ground truth and the same for every tool. Edges between files in different Components become Component edges.
- Metrics:
  - edge precision, recall and F1, both verb-aware and verb-agnostic;
  - Contract coverage;
  - locator accuracy: the cited file exists at the pinned SHA, the range is in bounds, and the range contains the anchor string recorded in the ground truth;
  - lists of false positives and false negatives;
  - unsupported honesty: facts claimed from files in languages the tool cannot parse deterministically;
  - index determinism: two cold indexes give byte-identical NAGs;
  - index cost: wall time, plus LLM tokens for any LLM-assisted indexing pass.

**Layer 1 — agent task benchmark.**

- The 12 ADR 0007 tasks run for each arm (A/B/C/D) and each model.
- Each task has:
  - a fixed prompt;
  - a required JSON answer schema (the Impact Set plus locators);
  - a deterministic scorer: set comparison against the ground truth. No LLM judge decides primary accuracy.
- Telemetry per run:
  - `tokens_input`, `tokens_output`, `cache_read_tokens`;
  - tool calls, by tool;
  - turns;
  - wall latency;
  - Impact Set precision, recall and F1;
  - locator validity in the answer;
  - hallucinated nodes: IDs absent from the corpus ground truth.
- Each task × arm × model cell runs at least 5 times, at temperature 0 where the model supports it. Reports give the median and the interquartile range.
- The model roster has two tiers:
  - **local:** about 8B, through an OpenAI-compatible Ollama endpoint;
  - **frontier:** a current model.
  
  Exact model IDs are pinned at run time and recorded. The roster in ADR 0007 is out of date. The question the roster answers is whether deterministic slices let a local model approach frontier accuracy.
- The runner speaks MCP as a client. It has a mock-model mode, so its loop, telemetry and scoring can be tested offline.

**Layer 2 — human comprehension evaluator.**

- **Automated proxies**, repeatable and run for every tool that produces tours or explanations:
  - **tour assertiveness:** the percentage of tour-step claims that can be verified at their cited locators;
  - **critical-path coverage:** the share of ground-truth flow steps the tour visits;
  - step count;
  - tokens or words per step.
- **Periodic moderated study**, manual and not in CI:
  - It is within-subject. Participants get a different corpus area per tool, counterbalanced with a Latin square to avoid learning effects.
  - Tasks include "how does an order flow from the cart to shipping?" and "who must be notified if PaymentService changes?".
  - Metrics: **time to first useful understanding** (the first answer graded correct); correctness, graded blind against the ground truth with a rubric; self-reported clarity (5-point scale); and raw NASA-TLX workload.
  - Small samples are reported as such, with no significance claims.

### C. Non-circular internal evaluator

**Pipeline under test.**

1. Load the corpus.
2. Index it through the same code path as `gira index`. That path is extracted from the CLI into a reusable indexing unit, which both the CLI and the evaluator call.
3. Emit briefings and slices through the engine.
4. Score them against the ground truth.

The evaluator never reads a pre-built `MacroMap`.

**Corpora.**

- **Hermetic, always on.** A minimal, vendored, license-attributed subset of the pinned OpenTelemetry Demo. It includes the shared proto, Compose and Helm membership, the Go services, and the Dockerfiles plus one source file per non-Go service, so that unsupported boundaries trigger. It is materialized into a deterministic local bare git repository and indexed offline. The Perturbed Variant is generated from it by a deterministic perturbation (a rename table plus altered edges), so it does not need to be vendored separately.
- **Full, opt-in.** The full pinned corpus, fetched through the git cache. It runs only when an explicit environment variable or build tag is set. Otherwise it skips with a visible reason.

**Ground truth v2.**

- The ground truth holds only expected Impact Sets, plus the locator anchors that Layer 0 uses. It has no `macro_map`.
- It carries `reviewed: true|false`. The flag becomes `true` only after a human has checked the ground truth against the pinned SHAs.
- The current fixtures are re-derived against the real corpus.

**Metrics fixed:**

- **Harness tokens** use the same estimator as the budget feature, applied to the bytes actually emitted.
- **Baseline tokens** are measured, not declared. A deterministic grep-baseline simulator runs a fixed search strategy over the actual corpus: search the target's identifiers, read the matching files, follow one hop of identifiers. Its cost is the estimated tokens of everything it reads. Changing the corpus changes the baseline. The real agent baseline is Layer 1, arm A.
- **Unsupported Safety:** for each expected unsupported path, there must be an unsupported record under that path, and there must be no `extracted` evidence whose locator cites a source file in a non-Wave-1 language under that path. Score = 100 × (1 − violations / checks).
- **Determinism:** two independent cold indexes, with separate cache directories, give byte-identical canonical JSON, and repeated briefings are byte-identical.

**Gate semantics** (decided by the owner on 2026-10-13):

- The test always asserts:
  - the pipeline runs end to end;
  - the report is valid against its schema;
  - Determinism = 100;
  - Unsupported Safety = 100.
  
  GIRA guarantees these properties by construction.
- Recall, precision and token-reduction thresholds are asserted only on ground truth marked `reviewed: true`. On unreviewed ground truth the full `GateReport` is written out and logged, and those thresholds do not fail the build.

## Scope

**In scope**

- The skill catalog in A.1 and A.2, generated for three host families. It covers the canonical sources, the install command and the activation contract with G-Ingá.
- New deterministic engine capabilities that the skills need:
  - hard token budgets with truncation metadata;
  - impact from a git diff;
  - tours;
  - deterministic flows;
  - architectural diff.
  
  Each is exposed on both the CLI and MCP, and all are additive.
- ADR 0012, which amends ADR 0007 §1.
- Benchmark quarantine scaffolding, Layer 0 adapters and scorer, the Layer 1 runner and telemetry, Layer 2 proxies, and the Layer 2 study protocol.
- Refactoring the internal evaluator as in part C: the hermetic mini-corpus, deterministic perturbation, ground truth v2, the measured baseline, real Unsupported Safety, and index determinism.

**Out of scope**

- Any model client inside `gira`, and any LLM-generated fact in the Macro Map.
- Extracting business rules.
- Changes to the G-Ingá harness repository. Only the integration contract is defined here.
- A web dashboard or HTTP daemon (ADR 0008 stands).
- Running Graphify or Understand Anything in CI. Committing their raw outputs.

**Non-goals**

- Beating Graphify's symbol-level coverage, or Understand Anything's prose richness. The target is Component-level correctness with evidence at a lower token cost.
- Claiming statistical significance from the human study.

## Acceptance criteria

- **SC1 — Non-circular evaluator.**
  - The evaluator obtains its `MacroMap` only by indexing a git repository through the shared indexing unit. No code path in the evaluator deserializes a pre-built `MacroMap`.
  - A mutation test proves the dependency: with the Protobuf extractor removed from the pipeline, recall on the Contract tasks drops measurably.
- **SC2 — No invented token constants.**
  - No per-task baseline constant and no fallback constant remain.
  - A test shows that adding files that match the target identifiers increases the measured baseline, while the harness tokens stay the same.
- **SC3 — Unsupported Safety is real.** On the hermetic corpus the score is 100. A test that injects one `extracted` evidence record citing a C++ source under an unsupported path scores below 100 and fails the gate.
- **SC4 — Index determinism.** Two cold indexes with separate cache directories produce byte-identical canonical JSON for both the published variant and the perturbed one.
- **SC5 — Gate semantics.** On unreviewed ground truth, a No-Go result writes the report and does not fail the build. On ground truth marked reviewed, a threshold miss fails the build. Both behaviors are covered by tests.
- **SC6 — Hard budgets.** For every task in the 12-task suite and every tier, the estimated tokens of the CLI and MCP output are ≤ the budget. Outputs that were truncated carry `truncated: true` and the `dropped` counts. A repeated call is byte-identical. The focal node is never dropped.
- **SC7 — Impact from diff.** On a fixture change that edits the shared proto service definition, `--from-diff` and `get_pr_impact` return the same Impact Set as the direct Contract target.
- **SC8 — Tours are deterministic and verifiable.**
  - Tour output is byte-identical across runs.
  - Every step locator resolves at the pinned SHA, and is in bounds.
  - Every `why` sentence is produced by the template renderer.
  - Tour critical-path coverage of the order flow in the ground truth is at least 80%.
- **SC9 — Architectural diff.** Diffing the published and perturbed indexes reports exactly the renames and altered edges that the perturbation table declares, with no extra changes.
- **SC10 — Skills distribution.**
  - `gira skills install` for each of the three host families generates files from the single canonical source.
  - A second run makes no changes.
  - Only content inside the install markers is modified.
  - Each `SKILL.md` declares a trigger description, the operations it uses, its output contract and its budget.
- **SC11 — Graceful fallback.** The `gira-query` procedure returns the same JSON for the same question whether the MCP server is registered or only the CLI is available. This is verified by a golden test covering both paths.
- **SC12 — Quarantine holds.**
  - A static check proves that no benchmark asset is referenced from the Go module, `mise.toml`, CI workflows, git hooks or agent configuration.
  - The output directory is ignored by git.
  - The container query phase declares that networking is disabled.
- **SC13 — Layer 0.** For each of GIRA, Graphify and Understand Anything, the adapter maps a small synthetic sample, written to the tool's documented schema, to the expected NAG (golden tests). The locator checker accepts valid locators and rejects out-of-bounds locators and locators whose anchor does not match.
- **SC14 — Layer 1.** In mock-model mode, the runner executes all 12 tasks for arms A and B offline. It emits telemetry records that are valid against the schema, with every field listed in part B present, and the scorer's precision and recall match a hand-computed fixture.
- **SC15 — Layer 2.** The automated proxies (assertiveness, coverage) run on GIRA tours in tests. The moderated-study protocol, with tasks, counterbalancing, rubric and instruments, is committed as a document.
- **SC16 — Regression and governance.** `go test -race ./...` passes. The G-Ingá validator reports no `fail` for the artifact records.

## Interfaces and dependencies

**Changed or new public contracts.** All are additive. No existing flag, field or tool changes meaning.

| Interface | Change |
|---|---|
| CLI `impact`, `slice` | new `--budget`; `impact` adds `--from-diff <base>..<head>` |
| CLI | new `tour`, `flows`, `diff`, `skills install` |
| MCP | `get_impact_briefing` and `get_context_slice` accept `budget`; new tools `get_pr_impact`, `get_tour`, `diff_maps` |
| Slice/briefing JSON | new optional `budget` object |
| Ground-truth fixture format | v2: no `macro_map`, adds `reviewed` and locator anchors. Its only consumer is the internal evaluator. |
| Internal `eval` API | reworked. There are no external consumers (it is an internal package). |

**Internal refactor.** The repo-loading and indexing sequence moves out of the CLI command into a shared unit. The CLI's behavior does not change, and the existing CLI e2e tests guard that.

**ADRs.**

- ADR 0012 (new) amends ADR 0007 §1 ("Prior Art Isolation") and the matching rejected option.
- ADR 0006's tier budgets become enforced behavior.
- ADR 0001 stands unchanged: no model client, and locators instead of source.

**External systems.**

- The OpenTelemetry Demo and Helm chart (Apache-2.0) are vendored as a minimal attributed subset with a NOTICE file.
- Graphify (Apache-2.0) and Understand Anything (MIT) run only inside quarantine containers. They are never dependencies.
- Model endpoints (a local Ollama endpoint and a frontier API) are used only by the Layer 1 runner, outside `gira`.

## Decisions

- `decided` on 2026-10-13 by the owner: artifacts live in the standalone repository, and the validator runs with `--root`.
- `decided` on 2026-10-13 by the owner: a new ADR 0012 amends ADR 0007, rather than editing ADR 0007 in place or keeping the ban.
- `decided` on 2026-10-13 by the owner: a hermetic mini-corpus is always on, and the full corpus is opt-in.
- `decided` on 2026-10-13 by the owner: No-Go results are reported honestly, and the ground truth is re-derived. Threshold assertions apply only to reviewed ground truth.
- **Token estimator, not a tokenizer.** A model tokenizer would add a dependency tied to a model vendor and would differ between hosts. A documented byte and structure estimator is deterministic and the same on every host. Layer 1 records real token counts from each model's API. *Alternative rejected:* bundling a tokenizer.
- **Tours are computed, not generated.** We borrow Understand Anything's `TourStep` shape, but not its LLM tour builder. UA ships a deterministic `generateHeuristicTour` that its pipeline never calls. That is evidence the heuristic approach is viable but under-used. *Alternative rejected:* letting the host LLM order the steps.
- **The evaluator lifts other tools' graphs to Components.** Lifting with one shared table keeps the comparison fair and keeps the work out of each tool's adapter. The report states that Graphify and UA are measured on a projection. *Alternative rejected:* comparing at file or symbol level, where GIRA has no claim to make.
- **Perturbation runs in Go, not a shell script.** A Go perturbation is deterministic, testable, and shared by the internal evaluator and the benchmark. It replaces the `scripts/perturb-corpus.sh` planned in ADR 0007.
- **The Layer 1 primary scorer is deterministic.** An LLM judge would reintroduce the non-determinism this framework exists to measure. Narrative quality is judged only in Layer 2, by humans.

## Risks and constraints

- **Money (protected action).** Layer 1 runs against frontier APIs, and Understand Anything's LLM indexing pass, cost money. Each paid run needs a separate owner `approved:` entry with a cost estimate, recording runs × tasks × models × the estimated tokens. Mock-model and Ollama runs need no approval.
- **Supply chain.** External tools are built from pinned source commits into images referenced by digest. Images are never pulled by a floating tag. Containers get no repository mounts beyond the corpus copy, run without credentials, and have networking disabled during the query phase.
- **Licensing.** The vendored corpus subset keeps its Apache-2.0 headers and gets a NOTICE entry. Nothing from Graphify or UA is copied (prior-art policy). The adapters are written from the tools' documented schemas.
- **Honest No-Go.** Once the evaluator stops being circular, the published Go result will likely flip to No-Go until the extractors and the ground truth catch up. This is expected and acceptable (owner decision). The parent task must not claim the ADR 0007 gate is met.
- **Fairness.** Lifting other tools' graphs to Components can under-credit symbol-level tools. Reports state this, and include verb-agnostic edge metrics.
- **Privacy.** Telemetry follows the ADR 0006 sanitization rule referenced by the harness: no prompts, secrets or repository metadata outside the public corpus are persisted. Transcripts stay in the ignored output directory.
- **Compatibility.** All CLI and MCP changes are additive. Existing MCP tool schemas only gain optional fields.
- **Performance.** Budget truncation must keep the latency targets in ADR 0007: a slice in under 50 ms, a cold index of the hermetic corpus in under 15 s.

## Verification

- **Integrated regression:** `go test -race ./...` on the integrated head, plus the existing CLI e2e and dogfood tests, unchanged.
- **End-to-end checks:**
  - SC1 to SC5, SC7 and SC9 run as Go tests over the hermetic corpus.
  - SC6 and SC8 run as Go tests across the 12 tasks.
  - SC10 and SC11 are golden tests.
  - SC12 is a static check script, also run as a test.
  - SC13 and SC14 are tests of the benchmark tooling, inside its own quarantined workspace, so they never run in the Go module's CI.
  - SC15 combines a test with a document review.
  - SC16 is the validator run with `--root`.
- **Opt-in:** the full-corpus evaluation is run once by hand, and its report is attached to the parent task. The run is free, since it involves no models.
- **Review focus:** a fresh-context adversarial review on three questions:
  1. Does any remaining path let the evaluator see ground truth before indexing? (circularity)
  2. Are budgets truly hard on every output format?
  3. Can any skill output state a fact without a locator?

## Rollback

skip: no one-way door. Every change is additive, or lives in internal packages and fixtures, and can be reverted with git. The only protected actions (paid benchmark runs) are gated individually under Risks.

## Open questions

- deferred: **the exact model roster** for Layer 1. It is safe to proceed because the runner pins and records model IDs at run time, and the roster does not change any interface.
- deferred: **recruiting participants** for the Layer 2 moderated study. It is safe because the automated proxies cover the acceptance criteria, and the study is periodic and manual.
- deferred: **how G-Ingá `/work` consumes the impact briefing** in its own routing code. It is safe because this spec fixes only the integration contract (inputs, outputs, error vocabulary). The change belongs to the G-Ingá repository.
- deferred: **the default budgets** for `gira-impact-briefing` (1,500). It is safe because the budget is a flag and can be tuned against Layer 1 data without an interface change.

## Revisions

- v1 (2026-10-13): first draft.
