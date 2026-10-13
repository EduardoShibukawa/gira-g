<!-- Generated copy of docs/artifact-contracts.md from G-Ingá, delivered in the /work bundle. Do not edit: change the canonical file and run `bun tools/validate/contract-copy.mjs`. Transform: this comment is added and relative link targets are prefixed with ../../../../docs/, which resolves only in a G-Ingá checkout. -->
# Artifact contracts

**Status:** accepted (decision D3, approved by the owner on 2026-10-10, [ADR 0003](../../../../docs/decisions/0003-minimum-artifact-schemas.md)). Nothing here is implemented. No validator exists yet, so every "deterministic check" below is a specification for M1, not an enforced rule.

This document is the field-level contract. The rationale and the alternatives are in ADR 0003. The tiers and routing rules come from [ADR 0002](../../../../docs/decisions/0002-ceremony-tiers-and-routing.md) (`accepted`) and aren't restated here.

## 1. What each tier persists

| | `light` | `standard` | `heavy` |
|---|---|---|---|
| Routing record | routing line in the report | `## Routing and approvals` in the task record | `## Routing and approvals` in the parent task record |
| Task record | none | one file | one parent file (`tier: heavy`) plus one file per decomposed task (each `tier: standard`) |
| Spec | none | none | one file, approved before any child task starts |
| Checkpoint | none | `## Checkpoint` section, only when the work stops or crosses sessions | same, in each task that stops or crosses sessions |
| Evidence | in the report | `## Verification` table in the task record | each child's table, plus the parent's integrated table and `## Review` |
| ADR | none | none | when architectural, in `docs/decisions/`, linked from the spec |

A typical `standard` record is 30–60 lines. It's written once, updated in place, and re-read only on resume or review.

```mermaid
flowchart TD
    R["Request"] --> RT{"Routing, ADR 0002"}
    RT -->|"light"| LR["Report: routing line and evidence, no file"]
    RT -->|"standard"| T["Task record, tier standard"]
    RT -->|"heavy, owner confirms"| P["Parent task record, tier heavy"]
    LR -.->|"escalation creates"| T
    T -.->|"escalation to heavy creates"| P
    P -->|"spec: id@version"| S["Spec: draft, approved, superseded"]
    S -.->|"links when architectural"| A["ADR"]
    P -->|"decomposes into"| C["Child task records, tier standard"]
    C -->|"spec: id@version"| S
    C -->|"verified children"| P
    P --> E["Integrated evidence: SC and R rows, Review, owner acceptance"]
    T --> CK["Checkpoint section when work stops or crosses sessions"]
    C --> CK
```

## 2. Shared conventions

### 2.1 Files and identifiers

- Task records live in `tasks/<id>.md` and specs in `specs/<id>.md` ([SDLC brief](../../../../docs/briefs/sdlc-harness-brief.md) §11). Where they live in a target repository is D5's call.
- The id is the file name without `.md`: `<YYYY-MM-DD>-<slug>`, for example `2026-10-12-pagination-off-by-one`. A date plus a slug doesn't collide across branches the way a sequence number would.
- A decomposed task's id is `<parent-id>-<NN>-<slug>`, with `NN` in dependency order.
- The title is the file's H1. There is no `id` or `title` key in the frontmatter, so they can't drift from the file name and heading.

### 2.2 Frontmatter and sections

Each file is Markdown with a short YAML frontmatter block. The frontmatter holds only what a script must read: the schema version, the tier, the state and the references to other files. Everything else is a `##` section with a fixed name.

Commit SHAs in the frontmatter are always quoted. YAML reads an unquoted all-digit SHA such as `0123456` as a number (checked with PyYAML: it becomes `42798`).

A conditional section that doesn't apply is either left out or written as `skip: <reason>` when the omission should stay visible (ADR 0002 §6). Unknown keys and sections are allowed and preserved.

### 2.3 Routing and approvals log

One append-only list, shared by every tier. In `light` it's the first lines of the report. In `standard` and `heavy` it's the `## Routing and approvals` section. Each entry is one bullet that starts with a keyword:

| Keyword | Format | When |
|---|---|---|
| `tier:` | `tier: <tier> (<deciding signals>)` | first entry, exactly once |
| `escalated:` | `escalated: <from> → <to> because <E1–E5>: <evidence>` | ADR 0002 §5 |
| `downgraded:` | `downgraded: <from> → <to> because <reason>` | `standard → light` once before edits; from `heavy` only with an `approved:` entry |
| `override:` | `override: owner ordered <tier>; risk: <one line>` | owner orders a lower tier |
| `confirmed:` | `confirmed: heavy by owner (<date>, <ref>)` | entering `heavy` |
| `approved:` | `approved: <subject> by owner (<date>, <ref>)` | spec version, ADR, exception, `heavy` downgrade, protected action |
| `decided:` | `decided: <question> → <answer> by owner (<date>, <ref>)` | a preference call during the work |
| `divergence:` | `divergence: <spec section>: <finding>` | E5 |
| `accepted:` | `accepted: by owner (<date>, <ref>)` | final acceptance |
| `spec-version:` | `spec-version: v<N> → v<N+1>; rerun: <row ids>` or `…; unaffected: <reason>` | a child moves to a reapproved spec version (§4.4) |

`<ref>` is the owner's own words quoted, or a link to the commit or comment where the owner approved. The agent writes the entry, but only after the owner's explicit message. The file can't prove who approved: that's a trust boundary, not a check (see §7).

`→` and `->` are both accepted. The current tier is the target of the last `tier:`, `escalated:`, `downgraded:` or `override:` entry. The one exception is a record closed by a downgrade to `light` (§6.1).

### 2.4 Results and provenance

- Check results use the brief's vocabulary only: `pass`, `fail`, `blocked`, `not run`. A row nobody has run yet is `not run`.
- A result is valid for the commit it ran on, never for "the code" in general.
- An acceptance criterion the agent derived without a source in the request or the spec is tagged `[inferred]`. This applies [ADR 0001](../../../../docs/decisions/0001-product-boundary.md) §3 (provenance) to the one place where an inference could silently become a requirement.

## 3. Task record (`schema: task/1`)

Used by `standard` tasks, by the `heavy` parent and by the `heavy` children.

### 3.1 Frontmatter

| Key | Required | Values | Notes |
|---|---|---|---|
| `schema` | always | `task/1` | contract version (§8) |
| `tier` | always | `standard`, `heavy` | `light` never has a file. Must equal the current tier in the log |
| `status` | always | see §3.3 | |
| `parent` | decomposed tasks | parent task id | |
| `spec` | `heavy` parent and its children | `<spec-id>@<version>` | the spec version this task implements |
| `depends_on` | when non-empty | list of task ids | tasks that must be `verified` first |
| `verified_at` | `verified`, `accepted` | commit SHA, quoted | the commit every `AC`, `SC`, `R` and `V` row ran on (§3.4). `B` rows keep their own, earlier commits |

### 3.2 Sections

| Section | Required when | Content |
|---|---|---|
| `## Routing and approvals` | always | the log (§2.3) |
| `## Objective` | always | the outcome in one to three sentences, then `In scope:` and `Out of scope:` bullets |
| `## Acceptance criteria` | always | numbered `AC1`, `AC2`, … Each is checkable. Children may cite spec criteria (`AC1 = SC2`) instead of copying them |
| `## Verification` | always | table `ID \| Check \| Result \| Ref \| Note` (§3.4) |
| `## Constraints and budget` | when something differs from the defaults (§3.5) | budget overrides, external dependencies, applicable limits |
| `## Plan` | S2 fails or step order matters (ADR 0002 §4); always for a `heavy` parent | ordered steps. A `heavy` parent lists its children here. Otherwise `skip: <reason>` or absent |
| `## Blockers` | `blocked`, `partial` | what blocks, what's needed, from whom. Options with trade-offs when it's a decision |
| `## Checkpoint` | the work stops or crosses sessions | §3.6 |
| `## Review` | `heavy` parent; `standard` after a review trigger (ADR 0002 §4) | each finding with its resolution: `fixed in <sha>`, `rejected: <reason>` or `open` |
| `## Outcome` | `verified`, `accepted`, `partial`, `cancelled` | what was delivered, what wasn't, follow-ups |

### 3.3 States

The brief requires `completed`, `verified` and `accepted` to be distinct, with partial and blocked outcomes.

| State | Meaning | Who sets it |
|---|---|---|
| `todo` | recorded, not started: a proposed `heavy` parent, or a child decomposed after the spec's approval | agent |
| `active` | work in progress. For a `heavy` parent this includes the question rounds and drafting the spec | agent |
| `blocked` | can't continue without a decision, an external event, a permission or a spec approval | agent |
| `completed` | implementation finished. Verification isn't all `pass` yet, for example a check is `blocked` or `not run` | agent |
| `verified` | every row that must run at `verified_at` meets its requirement (§3.4) | agent, guarded by checks |
| `accepted` | the owner accepted the result | owner (agent records the `accepted:` entry) |
| `partial` | stopped with some criteria verified and the rest recorded as not delivered | agent proposes, owner may resume |
| `cancelled` | dropped | owner. The agent sets it only to close a record after a `standard → light` downgrade (§6.1) |

```mermaid
stateDiagram-v2
    [*] --> todo: heavy proposed, or child decomposed after spec approval
    [*] --> active: standard starts
    todo --> active: start guard holds
    todo --> blocked: spec reopened
    active --> blocked: decision, external event, permission or spec reopened
    blocked --> active: blocker resolved
    active --> completed: implementation finished
    completed --> blocked: spec reopened
    completed --> active: verification fails, fix needed
    completed --> verified: verification rule holds at verified_at
    verified --> active: new change or new spec version invalidates evidence
    verified --> accepted: owner accepts
    active --> partial: stop with remainder recorded
    blocked --> partial: stop with remainder recorded
    completed --> partial: stop with remainder recorded
    partial --> active: owner resumes
    todo --> cancelled: owner drops, or downgrade to light
    active --> cancelled: owner drops, or downgrade to light
    blocked --> cancelled: owner drops
    verified --> [*]: standard ends here unless the owner reviews
    accepted --> [*]: heavy parent ends here
    cancelled --> [*]
```

The owner can cancel from any state that isn't terminal. The diagram shows only the common cases.

Guards. "det" marks what a script can check; the rest is agent judgment.

**Start (`todo → active`), by role:**

| Role | Guard |
|---|---|
| `standard`, no parent | every `depends_on` task is `verified` or `accepted` (det) |
| `heavy` parent | a `confirmed:` entry exists (det). The spec may be `draft`: the parent is `active` while it runs the question rounds and drafts the spec (ADR 0002 §4) |
| `heavy` child | every `depends_on` task is `verified` or `accepted`, and the spec is `approved` with `version` equal to the child's `@N` (det) |

**Spec gate (det, holds at every commit, not only on a transition):**

- Children exist only after the spec's first approval. A child whose spec has no `approved_version`, or whose `@N` is greater than the spec's `version`, fails.
- While the spec is `draft`, no child is `todo`, `active` or `completed`. Each one is `blocked`, and `## Blockers` names the spec version awaiting approval. Children that are `verified`, `accepted`, `partial` or `cancelled` keep their state.
- While the spec is `approved` at version `V`, a child can be `todo`, `active` or `completed` only with `@V`. The full rules for children on an older version are in §4.4.
- The parent's own `spec:` always equals the spec's current `version`, `draft` or not.

**Other guards:**

- `active → completed`: every `AC`, and for a `heavy` parent every spec `SC`, has at least one row in `## Verification` (det). The agent believes the implementation is done (judgment).
- `completed → verified`: the verification rule in §3.4 holds (det, except attribution). For a `heavy` parent, also:
  - every child that isn't `cancelled` is `verified` or `accepted` with `@V`, where `V` is the spec's current approved version;
  - the spec is `approved` and passes the history check (§7.1);
  - no `## Review` finding is `open`;
  - every `divergence:` entry has a later `approved:` entry resolving it.
- `verified → accepted`: an `accepted:` entry exists (det for presence; only the owner makes it true).
- `blocked` and `partial`: `## Blockers` is non-empty (det). `partial` also needs `## Outcome` naming the criteria not delivered (det for presence).
- A child task never becomes `heavy`. If it finds a `heavy` trigger the spec didn't cover, that's a `divergence:` on the parent (E5).

### 3.4 Verification table

One table plans the checks before editing and records the results afterward, so criteria, checks and evidence live in one place. Columns: `ID | Check | Result | Ref | Note`.

| ID prefix | Row kind | Runs on | Result needed for `verified` |
|---|---|---|---|
| `B<n>` | baseline: the relevant checks before the first edit (ADR 0002 §4) | its own commit, before the change | any. It records the starting state and is never rerun to reach `verified` |
| `AC<n>` | an acceptance criterion and the check that exercises it. A bug's `AC` starts with a failing repro | `verified_at` | `pass` |
| `SC<n>` | a spec criterion, end to end (parent only) | `verified_at` | `pass` |
| `R<n>` | regression or integrated run | `verified_at` | `pass`, or an attributed `fail` (below) |
| `V<n>` | self-review, for example "diff read against AC1–AC3" | `verified_at` | `pass` |

**Verification rule.** A record is `verified` only if all of these hold (det):

1. Every `AC`, `SC`, `R` and `V` row has `Ref` equal to `verified_at`.
2. Every `AC`, `SC` and `V` row is `pass`.
3. Every `R` row is `pass` or an attributed `fail`.
4. Every `B` row has its own `Ref`, and that commit is a strict ancestor of `verified_at` (`git merge-base --is-ancestor`, and not the same commit). A `B` row keeps the commit it ran on; it isn't refreshed.
5. A `standard` task and a `heavy` parent have at least one `B` row and one `V` row (ADR 0002 §4). When no relevant check exists (S6 fails), the `B` row says so and is `not run`.

**Attribution isn't a pass.** An `R` row that fails at `verified_at` keeps `Result: fail`. Its `Note` says `baseline: B<n>` and names the failing checks.

- **Script (det):** the cited `B` row exists and is `fail`. When both rows list failing checks, the `R` list is a subset of the `B` list.
- **Agent (judgment):** confirms it's the same failure.
- **Report:** lists the row as `fail (pre-existing, B<n>)`. It never folds it into a pass count.
- `AC`, `SC` and `V` rows can't be attributed.

**Flaky.** A flaky check is a `fail` with `Note: flaky: <evidence>`. It never counts as `pass`, and it can be attributed only if a `B` row shows the same failure.

`Ref` is the short commit SHA the row ran on, plus a log path when a log was kept. Logs aren't stored by default: the command and the commit make the row reproducible. A repro's first failure can go in `Note` (`failed first on <sha>`).

### 3.5 Budget defaults

ADR 0002 sets no numeric limits and D4 hasn't decided what can be measured, so every limit carries an enforcement label: `hard` (a platform mechanism enforces it) or `guideline` (the agent follows it, nothing enforces it).

| Limit | Default | Enforcement today |
|---|---|---|
| Verification attempts without new information | 2, then escalate (E3) | guideline |
| Subagents | 0 (brief §3: don't delegate by default) | guideline |
| Investigation for a routing signal | bounded, then record a blocker or question (ADR 0002 §2) | guideline |
| Tokens, time | none until D4 provides a signal | not set |
| Permissions and protected actions | Claude Code permission settings | hard (platform) |

A task records only what differs from these defaults, in `## Constraints and budget`. Expanding a `guideline` limit because progress justifies it is recorded there with the reason (brief §3). A `hard` limit is never exceeded. No tier and no budget relaxes approved acceptance criteria, security requirements or protected-action approval.

### 3.6 Checkpoint section

Concise operational state, not a transcript (brief §6). The latest checkpoint replaces the previous one; Git keeps the history.

```markdown
## Checkpoint

- as of: 2026-10-12, branch fix/pagination, head 4f2a9c1, worktree dirty (src/paging.ts)
- done: AC1 repro written and failing (see Verification)
- next: change the offset calculation in src/paging.ts, then rerun AC1 and R1
- open: hypothesis (inferred) that pageSize=0 is unreachable from the UI
- revalidate: B1 ran on 9d0e311; rerun it if main moved
```

On resume, before any edit: compare the current branch, head and worktree with `as of`. If they differ, reconcile first. Treat any row whose `Ref` isn't the current head as stale until rerun or explicitly judged unaffected.

## 4. Spec (`schema: spec/1`, `heavy` only)

### 4.1 Frontmatter

| Key | Required | Values | Notes |
|---|---|---|---|
| `schema` | always | `spec/1` | |
| `status` | always | `draft`, `approved`, `superseded` | the brief's three states |
| `version` | always | integer, starts at 1 | goes up by one each time an approved spec is reopened |
| `approved_version` | after the first approval | integer | the last version the owner approved |
| `approved_hash` | after the first approval | `sha256:<hex>` | hash of the body at that approval (§4.3) |
| `superseded_by` | `superseded` | spec id | |

The approval itself is an `approved: spec v<N>` entry in the parent task's log. It isn't in the spec, because the spec body is hashed and an entry there would change the hash it records.

### 4.2 Sections

| Section | Required | Content |
|---|---|---|
| `## Problem` | always | current behavior, who is affected, which `heavy` triggers apply |
| `## Expected behavior` | always | observable behavior after the change |
| `## Scope` | always | in scope, out of scope, non-goals |
| `## Acceptance criteria` | always | `SC1`, `SC2`, …, checkable end to end |
| `## Interfaces and dependencies` | always | contracts that change, their consumers, data and external systems. `none` needs a reason |
| `## Decisions` | always | alternatives considered and the choice, briefly. Link the ADR when architectural |
| `## Risks and constraints` | always | security, data, compatibility, performance. Protected actions the work will need |
| `## Verification` | always | integrated regression scope, end-to-end checks for the `SC` rows, the review's focus |
| `## Rollback` | when H2 holds or a protected action is planned; else `skip:` | how to undo, or why it can't be undone and what guards the step |
| `## Open questions` | always | empty, or each one marked `deferred: <why it's safe to proceed>` before approval |
| `## Revisions` | always | one line per version: what changed and why, with the `divergence:` it resolves |

Following the references (A2 in the [D3 evidence](../../../../docs/research/d3-artifact-evidence.md)), a spec avoids file paths and code unless they're the contract itself.

### 4.3 Approval and versions

The approved text must not change without the owner seeing it (brief §3).

1. **Draft.** The parent is `active` while the spec is `draft` (§3.3). No child exists yet.
2. **Approve.** The owner approves version `N`. The agent sets `status: approved`, `approved_version: N` and `approved_hash`, and adds `approved: spec vN by owner (…)` to the parent log. The parent then decomposes into children at `@N`.
3. **Freeze.** While `status: approved`, the body must hash to `approved_hash` (det). Any edit fails the check, even a typo fix. Editing the body and the hash together is caught by the history check (§7.1).
4. **Reopen.** To change an approved spec, set `status: draft` and `version: N+1`, edit the body, and add a `## Revisions` line. In the same commit, every child in `todo`, `active` or `completed` goes to `blocked` (§3.3 spec gate).
5. **Reapprove.** The owner approves `N+1`. The agent sets `status: approved`, `approved_version: N+1` and the new `approved_hash`, adds the `approved:` entry, and moves the parent to `@N+1`. Children move as §4.4 says.
6. **Exception instead of a change.** The owner can accept a divergence without changing the spec. That's an `approved: exception <what> by owner (…)` entry in the parent log, and the spec stays as it is.
7. **Supersede.** When the work is replaced wholesale, the old spec becomes `superseded` with `superseded_by`. Its body doesn't change in that commit, and the file is never deleted.

The hash covers every byte after the closing `---` of the frontmatter, with LF line endings. Frontmatter edits, such as a schema migration (§8), don't change it. One definition, for illustration:

```sh
awk 'n >= 2 { print } /^---$/ { n++ }' specs/<id>.md | sha256sum
```

### 4.4 Children and spec versions

V1 blocks every open child when the spec is reopened. It doesn't try to sort children into affected and unaffected. That sorting would compare against a draft that can still change, and the validator couldn't check it.

A child moves from `@N` to `@N+1` only after reapproval, in one edit: update `spec:` and add a `spec-version:` entry to its log.

| Child state when the spec is reapproved | What happens |
|---|---|
| `blocked` | moves to `@N+1`, then resumes (`blocked → active`). The `spec-version:` entry lists the rows to rerun |
| `verified` or `accepted` | either goes `verified → active` and reruns (entry `rerun: …`), or keeps its state with `unaffected: <reason>` (judgment, recorded). `verified_at` doesn't change in the second case |
| `partial` or `cancelled` | may stay at `@N`. It doesn't count toward the parent |

What the validator does (det), using only the current files. It reads no old spec text:

| Spec now | Child's `@` | Child states allowed |
|---|---|---|
| never approved | any | none: a child must not exist |
| `draft`, reopened (`version = approved_version + 1`) | `@approved_version` | `blocked`, `verified`, `accepted`, `partial`, `cancelled` |
| `approved` at `V` | `@V` | any |
| `approved` at `V` | lower than `V` | `blocked`, `verified`, `accepted`, `partial`, `cancelled`. The parent can't be `verified` while a counted child is below `@V` |
| `superseded` | any | `blocked`, `verified`, `accepted`, `partial`, `cancelled`. The parent can't be `verified` against a superseded spec |
| any | higher than `version` | none: fail |

Old spec text is still in Git for people who want to read it. The validator never needs it to resolve a version. The only history it reads is the transition walk in §7.1.

## 5. Heavy evidence

ADR 0002 requires an evidence record for `heavy`. Here it's the parent task record, not a separate file:

- The parent's `## Verification` has `B` rows (the relevant suite before any child started), `SC` rows (each spec criterion end to end on the integrated head) and `R` rows (the integrated regression run).
- The parent's `## Review` holds the fresh-context review against the spec, which flags only correctness or requirement gaps (ADR 0002 §4).
- Each child keeps its own `AC` evidence. The parent links children by id rather than copying their rows.
- Release, deploy, migration or publishing are protected actions. `accepted` doesn't authorize them. Each needs its own `approved:` entry, and `## Outcome` lists any that are still pending.

## 6. Light report

`light` persists nothing (ADR 0002 §4). The report is the record:

```text
tier: light (S1–S4, S6, S7 hold; S5 doesn't)
Changed: README.md, "recieve" → "receive" (1 line, prose only)
Checks:
- V1 diff read: 1 line changed, no other files: pass
- V2 markdown lint on README.md: not run (no linter configured)
```

If a `light` task escalates, the agent creates the task record at that moment. It copies the routing lines and any evidence still valid, then adds the `escalated:` entry.

### 6.1 Downgrade from `standard` to `light`

ADR 0002 allows this once, before any edit. If a task record already exists, the downgrade doesn't delete it silently and doesn't leave a `light` record behind:

1. **The record is untracked** (`git ls-files --error-unmatch` fails), and the agent created it in this same task. Nobody else wrote it, so it isn't user-owned. The agent:
   - copies its routing lines, the new `downgraded:` line and any `B` rows into the `light` report;
   - deletes the file;
   - says so in the report: `task record tasks/<id>.md removed (untracked, created in this task)`.
2. **Otherwise** (the record is committed, or the agent can't be sure it's the only author), the agent keeps the file and closes it:
   - appends `downgraded: standard → light because <reason>`;
   - sets `status: cancelled`;
   - adds `## Outcome: continued as light; see <commit or report>`.

   The frontmatter keeps `tier: standard`. A closed `standard` record whose last routing entry is a downgrade to `light` is exempt from the current-tier rule (det). It's history, not a `light` artifact.

Either way, the report starts with the same routing lines the record had.

### 6.2 Interruption

A `light` task has no resume contract. It fits one session by definition (S8). If it's interrupted anyway, the next session reroutes from the request and the Git diff, which is cheap because the change is small.

## 7. Validation

Nothing below exists yet. M1 can implement the deterministic column as one small script.

| Check | Deterministic script | Agent judgment | Owner |
|---|---|---|---|
| Frontmatter parses, `schema` known, required keys present, enum values valid | ✓ | | |
| File name is a valid id, ids are unique, `parent`, `depends_on` and `spec` resolve, no dependency cycles | ✓ | | |
| Required sections present for the tier and status; `skip:` has a reason | ✓ | | |
| Log grammar; frontmatter `tier` equals the current tier, except the downgrade closure (§6.1); `heavy` parent has `confirmed:` before leaving `todo` | ✓ | | |
| `AC`/`SC` ids unique; every criterion has a verification row; `B` and `V` rows present; result values valid | ✓ | | |
| Start guards and spec gate (§3.3); child version table (§4.4) | ✓ | | |
| Verification rule (§3.4): rows at `verified_at`, `B` rows strict ancestors, attribution citations | ✓ (needs Git) | same-failure judgment | |
| Approved spec body matches `approved_hash` | ✓ | | |
| History walk for specs and task records (§7.1) | ✓ (needs full history) | | |
| Retired-schema records are frozen (§8) | ✓ (needs full history) | | |
| Internal links resolve | ✓ | | |
| Obvious secrets in records (brief §6) | ✓ (pattern scan, incomplete) | | |
| Routing signals answered correctly | | ✓ | can override |
| Criteria are checkable and cover the objective | | ✓ | approves them in `heavy` |
| The check actually exercises its criterion | | ✓ | |
| Evidence still applies after later changes; a child `unaffected` by a new spec version | | ✓ | |
| Implementation diverges from the spec | | ✓ (raises it) | decides |
| Review findings: real or noise | | ✓ | final call in `heavy` |
| `heavy` confirmation, spec approval, exceptions, acceptance, overrides, protected actions | | records them | ✓ |

The script can prove that an `approved:` line exists, not that the owner wrote it. That limit holds until approvals come from a channel the agent can't write, and no such channel is in scope.

### 7.1 History walk

For each spec and task record, the script walks the commits that touched the file, oldest first (`git log --reverse --format=%H -- <file>`). The working tree counts as a final state. It checks each consecutive pair of states, `prev` and `cur`; `h(x)` is the body hash of a state.

Spec transitions:

| `prev → cur` | Allowed only if |
|---|---|
| none → `draft` | `version: 1`, no `approved_*` keys |
| `draft → draft` | `version`, `approved_version` and `approved_hash` unchanged |
| `draft → approved` | `approved_version = version`. `approved_hash = h(cur)`. The parent record at `cur`'s commit has an `approved: spec v<version>` entry that it didn't have at `prev`'s commit |
| `approved → approved`, same `version` | `version`, `approved_version` and `approved_hash` unchanged, and `h(cur) = h(prev)`. Only frontmatter changed, for example a migration |
| `approved → draft` | `version = prev.version + 1`; `approved_version` and `approved_hash` unchanged; `## Revisions` gained a line |
| `approved → approved`, `version + 1` | a reopen and a reapproval squashed into one commit: both rows above hold, combined |
| `draft` or `approved` → `superseded` | `h(cur) = h(prev)`; `superseded_by` resolves |
| anything else | fail. That includes body and hash changed together without a new version and a new `approved:` entry, a lower `version`, any change after `superseded`, and deleting the file |

Every `approved` state must also satisfy `h = approved_hash` at its own commit.

Task records follow the same walk. `cur.status` must be reachable from `prev.status` in the §3.3 diagram, and the guards of `cur.status` must hold at `cur`. A commit can cover several steps, so reachability is checked, not single edges.

Limits:

- A shallow clone or missing commits make the walk `blocked`, never `pass`.
- The walk sees only the history that exists. A rewritten history (rebase, force-push) or a squash merge shows up only as its end states, which still have to satisfy the table.
- It can't tell who made a commit. Approvals stay a trust boundary (§7).

### 7.2 Validator test cases (for M1)

These fixtures specify the validator. None has run, because there's no validator yet.

| ID | Fixture | Expected |
|---|---|---|
| T1 | `heavy` parent `active`, `confirmed:` present, spec `draft` v1, no children | pass |
| T2 | `heavy` parent `active` without `confirmed:` | fail |
| T3 | child exists while the spec was never approved | fail |
| T4 | spec reopened (`draft` v2); child `active` at `@1` | fail |
| T5 | spec reopened (`draft` v2); child `blocked` at `@1`; another child `verified` at `@1` | pass |
| T6 | spec `approved` v2; child `active` at `@1` | fail |
| T7 | spec `approved` v2; child `verified` at `@1`; parent `active` / parent `verified` | pass / fail |
| T8 | child moved to `@2` with a `spec-version: v1 → v2; unaffected: …` entry; parent `verified` | pass |
| T9 | approved spec body edited, `approved_hash` unchanged | fail |
| T10 | one commit changes an approved spec's body and `approved_hash`, same `version` | fail (history) |
| T11 | same as T10, but `version + 1`, a new `## Revisions` line and a new `approved: spec v2` entry in the parent | pass (history) |
| T12 | history check run in a shallow clone | blocked |
| T13 | `verified` with an `R` row `fail`, `Note: baseline: B1`, and `B1` `fail` | pass |
| T14 | same, but `B1` is `pass`; or the attributed row is an `AC` | fail |
| T15 | `verified` with a `B` row whose `Ref` equals `verified_at`, or isn't its ancestor | fail |
| T16 | `verified` with an `AC` row whose `Ref` differs from `verified_at` | fail |
| T17 | `cancelled` record, `tier: standard`, last entry `downgraded: standard → light` | pass |
| T18 | retired-major record (`task/1` once `task/3` is current), `accepted`, and unchanged since its status last changed / changed later / `active` | frozen / fail / fail |
| T19 | commit that changes `schema:` and also `status` | fail |

## 8. Versioning and evolution

- `schema: task/1` and `schema: spec/1` carry a major version only.
- **No bump:** adding an optional key or section, or adding a log keyword. Readers ignore what they don't know and keep it.
- **Bump:** a new required key or section, a renamed or removed key, or a changed meaning or enum.
- **Supported majors** (current and previous): the validator checks a record's content fully. It writes only the current major.
- **Retired majors** (older than previous): the validator keeps a registry with one line per schema ever released: its name and its terminal states (`verified`, `accepted`, `partial`, `cancelled` for `task/1`; `superseded` for `spec/1`). Lines are never removed. A record on a retired major is checked for freezing, not content (det):
  1. its `schema` is in the registry;
  2. its status is terminal for that schema;
  3. no commit after the one where its status last changed has touched the file, and the working tree matches `HEAD`.

  The validator reports such a record as `frozen`, not `pass`: it was valid under the rules of its time, and Git shows it hasn't changed since. A retired-major record that isn't terminal, or that changed after freezing, fails. It has to be migrated before any work resumes on it.
- **Migration** is an explicit, separate commit the owner can review. The agent never migrates in passing while doing other work. The migration commit changes `schema:` and the keys the migration names. It doesn't change `status` or any `## Verification` row (det, T19). A frontmatter-only migration doesn't touch an approved spec's hash. A body migration does, so the owner has to reapprove the spec through §4.3.
- **Unknown or missing `schema`:** the file is treated as unstructured prose and reported as `blocked` for validation. The agent doesn't guess the version.
- **Missing required value:** the validator fails. The agent doesn't invent values: a missing approval stays missing until the owner gives one, and a missing result is `not run`.
- **Invalid state**, for example `verified` with a `not run` row: the fix is to move the status back with a reason, never to edit the evidence.

## 9. Examples

### 9.1 `light`: fix a README typo (ADR 0002 R01)

See §6. No file is created.

### 9.2 `standard`: off-by-one pagination bug (ADR 0002 R04)

`tasks/2026-10-12-pagination-off-by-one.md`:

```markdown
---
schema: task/1
tier: standard
status: verified
verified_at: "7c1e0b2"
---

# Owners list skips one owner between pages

## Routing and approvals

- tier: standard (S2 fails: cause unknown; S1, S3–S7 hold)

## Objective

Page 2 of the owners list starts at the 12th owner instead of the 11th. Fix it so pages are contiguous.

In scope: the paging offset in the owners query.
Out of scope: page-size options, sorting.

## Acceptance criteria

1. AC1: with 25 owners and page size 10, pages 1–3 return owners 1–10, 11–20 and 21–25.
2. AC2: page 1 is unchanged.
3. AC3 [inferred]: a page past the end returns an empty list, not an error.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | owners module tests before the change | pass | 9d0e311 | |
| AC1 | new test `paginates contiguously` | pass | 7c1e0b2 | failed first on 3b8f4d0 |
| AC2 | existing test `first page` | pass | 7c1e0b2 | |
| AC3 | new test `page past end is empty` | pass | 7c1e0b2 | |
| R1 | owners module tests | pass | 7c1e0b2 | |
| V1 | diff read against AC1–AC3 | pass | 7c1e0b2 | |

## Plan

1. Write the AC1 repro and confirm it fails.
2. Find where the offset is computed and fix it.
3. Add AC3, then run R1.

## Outcome

Offset used `page * size + 1`; now `(page - 1) * size`. No follow-ups.
```

### 9.3 `heavy`: rename a response field the client consumes (ADR 0002 R07)

Before the spec's approval, the parent was already `active` with `spec: 2026-10-14-owner-phone-field@1`. The spec was `draft` v1 and no child file existed (§4.3 step 1, test T1). Its log then had only the `tier:` and `confirmed:` entries.

Parent `tasks/2026-10-14-owner-phone-field.md`, after approval, while the second child is in progress:

```markdown
---
schema: task/1
tier: heavy
status: active
spec: 2026-10-14-owner-phone-field@1
---

# Rename Owner.telephone to phone across the REST API and the Angular client

## Routing and approvals

- tier: heavy (H3: changes a response field the Angular client consumes)
- confirmed: heavy by owner (2026-10-14, "pode seguir como heavy")
- approved: spec v1 by owner (2026-10-14, "spec aprovada")

## Objective

Owners expose `phone` instead of `telephone`, without breaking the client at any commit. See the spec.

In scope: the REST response and request bodies, the Angular client.
Out of scope: the database column name.

## Acceptance criteria

SC1–SC3 from the spec.

## Plan

1. 2026-10-14-owner-phone-field-01-expand: serve both fields (verified)
2. 2026-10-14-owner-phone-field-02-client: client uses `phone` (active, depends on 01)
3. 2026-10-14-owner-phone-field-03-contract: remove `telephone` (todo, depends on 02)

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | REST and Angular test suites before 01 | pass | a11f0e4 | |
| SC1 | end-to-end: create and edit an owner's phone from the UI | not run | | |
| SC2 | API response has `phone` and no `telephone` | not run | | |
| SC3 | API rejects a request body that uses `telephone` with a clear error | not run | | |
| R1 | REST and Angular suites on the integrated head | not run | | |
| V1 | integrated diff read against SC1–SC3 | not run | | |

## Review

skip: runs after child 03 is verified

## Checkpoint

- as of: 2026-10-15, branch feat/owner-phone, head c93d2aa, worktree clean
- done: 01 verified (both fields served)
- next: finish 02, then 03, then SC1–SC3 and R1, then the review
- revalidate: B1 ran on a11f0e4; rerun before R1 if main moved
```

Spec `specs/2026-10-14-owner-phone-field.md`, abridged:

```markdown
---
schema: spec/1
status: approved
version: 1
approved_version: 1
approved_hash: sha256:<hash of the body at approval>
---

# Owner phone field

## Problem

The API calls the field `telephone`, the UI calls it `phone`. Renaming it changes a contract the Angular client consumes (H3).

## Expected behavior

Responses use `phone`. Requests that send `telephone` get a 400 with a message naming `phone`.

## Scope

In: owner endpoints, the Angular owner forms. Out: the database column; other entities.

## Acceptance criteria

- SC1: a user can create and edit an owner's phone from the UI.
- SC2: owner responses contain `phone` and not `telephone`.
- SC3: a request body with `telephone` returns 400 naming `phone`.

## Interfaces and dependencies

Owner JSON contract (consumer: Angular client in this system). No external consumers known (inferred; checked by searching both repositories).

## Decisions

Expand–contract: serve both fields, move the client, then remove `telephone`. Rejected: a single breaking commit, because the client would break between deploys.

## Risks and constraints

Unknown external consumers would break at the contract step. No data migration.

## Verification

Both suites at each child; end to end for SC1–SC3 on the integrated head; review focused on the contract step.

## Rollback

skip: no one-way door; reverting 03 restores `telephone`.

## Open questions

None.

## Revisions

- v1: initial.
```

Child `tasks/2026-10-14-owner-phone-field-02-client.md` has the same shape as §9.2, with `parent: 2026-10-14-owner-phone-field`, `spec: 2026-10-14-owner-phone-field@1`, `depends_on: [2026-10-14-owner-phone-field-01-expand]`, and criteria like `AC1 = SC1 (client side)`. Its routing log starts with `tier: standard (decomposed from heavy parent)`. If the spec were later reopened as v2, this child would go to `blocked` in the same commit, and after reapproval it would move to `@2` with a `spec-version:` entry (§4.4).

### 9.4 What they share

| Concept | `light` | `standard` | `heavy` |
|---|---|---|---|
| Routing log grammar | in the report | in the task record | in the parent; children say "decomposed from heavy parent" |
| Results vocabulary | ✓ | ✓ | ✓ |
| Criteria → check → result | rows `V`, optional `AC` in the report | `## Verification` | the same table at two levels |
| State machine | none | `active` → … → `verified` | children as `standard`; parent ends at `accepted` |
| Checkpoint | none | only when stopped or crossing sessions | per task, same format |
| Owner gates | permissions only | `decided:` when a call arises | `confirmed:`, `approved:` spec, `accepted:` |
