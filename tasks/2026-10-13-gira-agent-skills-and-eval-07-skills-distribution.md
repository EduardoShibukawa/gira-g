---
schema: task/1
tier: standard
status: verified
verified_at: "ddc4920"
parent: 2026-10-13-gira-agent-skills-and-eval
spec: gira-agent-skills-and-eval@1
depends_on: [2026-10-13-gira-agent-skills-and-eval-04-impact-from-diff, 2026-10-13-gira-agent-skills-and-eval-05-tours-and-flows, 2026-10-13-gira-agent-skills-and-eval-06-arch-diff]
---

# Skills distribution (SC10, SC11)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-agent-skills-and-eval; spec v1 approved)

## Objective

Canonical skill sources; `gira skills install` for Claude Code, Cursor and agents hosts; MCP/CLI fallback contract.

In scope: spec §A items covered by SC10, SC11.
Out of scope: other children's criteria.

## Acceptance criteria

- AC1 = SC10
- AC2 = SC11

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` | pass | bace0dc | all packages ok before any edit |
| AC1 | `go test -race ./internal/skills -v` (catalog contract, install idempotent and marker-scoped for claude, cursor, agents; opt-in hook) | pass | ddc4920 | 6 skills declare trigger, operations, contract, budget; 3 hosts generated; second run unchanged; outside-marker and other JSON bytes kept; hook opt-in and append-only |
| AC2 | `go test -race ./cmd/gira -run SC11 and -run GiraQuerySkill -v` (golden, MCP path and CLI path) | pass | ddc4920 | 10 questions byte-identical via MCP and CLI and equal to goldens; error vocabulary covered |
| R1 | `go test -race ./...` | pass | ddc4920 | all packages ok |
| R2 | `node tools/validate/validate.js` | pass | ddc4920 | overall pass |
| V1 | read the diff against the ACs | pass | ddc4920 | scope: internal/skills, CLI skills install and nudge, cmd/gira goldens, docs/skills/work-integration.md, go.mod yaml direct |

## Plan

1. Write the canonical skill sources and the parser that checks their contract.
2. Write the marker regions and the byte-preserving JSON key and array edits.
3. Write the installer for the three host families, then the CLI.
4. Write the `gira-query` procedure and its golden test on both paths.
5. Write the opt-in nudge hook.

## Outcome

Delivered:
- Six canonical skills embedded in the binary (`internal/skills/catalog/*/SKILL.md`): `gira-impact-briefing`, `gira-context`, `gira-query`, `gira-tour`, `gira-explain-component` and `gira-arch-diff`. Each declares `description` (its trigger), `operations`, `output_contract` and `budget`, and has `## Operations`, `## Output contract` and `## Budget` sections. `skills.Parse` rejects a skill that lacks any of them. The bodies carry rule I1 (evidence or silence) and rule I2 (label `narrated` sentences).
- `gira skills install --host claude|cursor|agents [--dir] [--index] [--command] [--hook]`:
  - claude: `.claude/skills/gira-*/SKILL.md`, plus `mcpServers.gira` in `.mcp.json`;
  - cursor: `.cursor/rules/gira.mdc`, plus `mcpServers.gira` in `.cursor/mcp.json`;
  - agents: `.agents/skills/gira-*/SKILL.md`.
- Markdown files carry `gira:begin <id>`/`gira:end <id>` regions: YAML comments inside the frontmatter, HTML comments around the body. In JSON files, the owned key `mcpServers.gira` stands in for the markers. It is edited by byte splicing, so the rest of the file keeps its bytes, indentation included. A second run reports every file `unchanged`.
- Opt-in `--hook` (claude host): appends one `PreToolUse` entry running `gira skills nudge` to `.claude/settings.json`. User entries are kept and the append is idempotent. The nudge emits `additionalContext` naming the mapped node ids. It never sets a permission decision and always exits 0.
- `skills.Querier`, the `gira-query` procedure. It uses MCP when registered and the CLI with `--format json` otherwise. Output is normalized JSON. The shared error vocabulary is `index_missing`, `index_stale` (HEAD not cited by the index), `node_not_found` (with deduplicated `locate` suggestions) and `budget_exceeded`.
- `docs/skills/work-integration.md`: the `/work` contract (inputs, outputs and error handling). The G-Ingá change itself is out of scope.
- `go.mod`: `go mod tidy` marks `gopkg.in/yaml.v3` as a direct dependency. It was already required, and its version is unchanged.

SC11 evidence: 10 questions (impact, budgeted slices, explain, locate, tour, diff, and the `node_not_found` and `budget_exceeded` errors) give byte-identical JSON through MCP and through the CLI. Both match the goldens in `cmd/gira/testdata/gira-query/`; regenerate them with `-update-golden`. `index_missing` is tested on the CLI path only, since an MCP server cannot start without its index.

Note: the nudge relies on Claude Code reading `additionalContext` from a `PreToolUse` hook. If a host version ignores that field, the hook does nothing, which is still safe.

