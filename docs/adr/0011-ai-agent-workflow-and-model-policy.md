# AI agent workflow, guardrails, and model policy

AI coding agents working in this repository operate under vendor-neutral guidelines anchored by a portable instructions layer, deterministic repository-level guardrails, a structured Wayfinder-to-SDLC delivery flow, clear unattended action boundaries, and an explicit model/effort policy.

## Decision Summary

1. **Portable Instruction Layer (`AGENTS.md` and `SKILL.md`):**
   - Universal root: `AGENTS.md` at the repository root is the sole authoritative, vendor-neutral layer of repository instructions, operational boundaries, and model policies across all agent harnesses (Claude Code, OpenAI Codex, GitHub Copilot, Cursor, Antigravity, etc.).
   - Thin Vendor Adapters: Tool-specific configuration files (e.g. `CLAUDE.md`, `.cursorrules`, `.copilot-instructions.md`) must only act as thin adapters pointing back to `AGENTS.md` (e.g., a `CLAUDE.md` containing `@AGENTS.md`), preventing instruction divergence and vendor lock-in.
   - Nested Scopes: Domain/lab subdirectories (such as `labs/gira-g/`) may declare contextual guidance that strictly augments—never overrides or relaxes—root guardrails.
   - Modular Skills: Reusable procedures, tool patterns, and workflows follow the open Agent Skills specification (`SKILL.md`). Directory symlinking/mapping (`.agents/skills/` and `.claude/skills/`) is maintained to ensure portability across different agent runners.

2. **Dual-Layer Guardrail Enforcement:**
   - Universal & Deterministic (Layer 1): Primary guardrails are strictly agent-agnostic and enforced deterministically via Git hooks managed by `lefthook` (pre-commit, pre-push) running `mise run check` (Go tests, linters, TypeScript typechecks, golden tests, license safeguards).
   - Absolute CI Verification: Because local git hooks can be bypassed (e.g. `--no-verify`) or omitted in headless environments, GitHub Actions CI executes the identical `mise run check` suite on every PR and branch. CI is the absolute barrier of truth.
   - Convenience Adapters (Layer 2): Agent-specific pre-tool or pre-action hooks (such as Claude Code hooks or Copilot permissions) serve solely as interactive convenience filters (e.g. catching accidental destructive file edits or unauthorized terminal calls early), but never substitute for deterministic git/CI gates.

3. **Wayfinder → Spec → Tickets Delivery Flow:**
   - Architectural Mapping: Exploratory architectural design, trade-offs, and design questions originate and evolve in Wayfinder maps (`.scratch/<effort>/map.md`) and grilling/research tickets (`.scratch/<effort>/issues/*.md`).
   - Architectural Decision Records (ADRs): Settled decisions are codified into formal ADRs in `labs/gira-g/docs/adr/`.
   - Formal Specification: Once an effort is decision-complete, the Wayfinder map transitions to formal MVP functional/technical specifications via `/to-spec`.
   - Execution Ceremony: Implementation proceeds strictly under SDLC execution ceremonies (`/work`, tracked in `tasks/`), with deliverables submitted through verifiable unit/integration tests and automated quality gates.

4. **Unattended Execution Boundaries:**
   - Allowed Unattended: Agents may autonomously explore code, read documentation, execute read-only queries, run local build/lint/test verification tasks (`mise run check`), draft design proposals, and prepare working tree modifications.
   - Forbidden Unattended: Agents are strictly prohibited from:
     - Committing or pushing changes without 100% pass on all local checks (`mise run check`) and license compliance.
     - Modifying `mise.toml` or `mise.lock` (toolchain dependencies and hashes) without explicit human sign-off.
     - Bypassing hooks via `--no-verify`.
     - Publishing releases, creating tags, or modifying CI/CD workflows autonomously.

5. **Model and Effort Policy:**
   - Standardized Tiers: Repository tasks are classified into vendor-neutral tiers with clear cost, latency, and reasoning profiles:
     - **Tier 1 (Frontier / High Reasoning):** Architectural grilling, complex synthesis, design dispute resolution, and security-critical reviews. Deployed on Frontier models (e.g. Claude 3.7 Sonnet / Claude Opus 5.5 / GPT-4.5 / Gemini 2.5 Pro) with high reasoning effort.
     - **Tier 2 (Smart & Fast / Balanced Execution):** Research subagents, deep code analysis, component implementation, and refactoring. Deployed on balanced reasoning models (e.g. Claude 3.5/3.7 Sonnet / GPT-4o / Gemini Flash 2.0 Thinking) with medium reasoning effort.
     - **Tier 3 (Light & Fast / Mechanical):** Mechanical search, syntax formatting, repetitive boilerplate, and data extraction. Deployed on lightweight models (e.g. Claude Haiku 3.5/5.5 / GPT-4o-mini / Gemini Flash Lite) with low reasoning effort.
   - Enforcement: The policy is declared canonically in `AGENTS.md` and explicitly enforced via subagent invocation manifests (e.g. `model: inherit | flash | pro`, `effort: high | medium | low`) and agent task templates, avoiding accidental frontier model sprawl on mechanical tasks.

## Considered Options

- **Vendor-locked instructions (`CLAUDE.md`-first or Cursor-first):**
  - *Rejected:* Breaches repository portability and introduces instruction divergence across developers and automated agents. `AGENTS.md` is universally respected or adaptable via thin single-line imports.
- **Relying solely on agent pre-action hooks (e.g. Claude Code PreToolUse):**
  - *Rejected:* Highly fragile and vendor-dependent. Pre-tool hooks do not exist across all CLI tools, cloud agents, or IDE extensions. Deterministic git hooks (`lefthook`) + CI matrix gates provide unbreakable enforcement.
- **Permitting autonomous dependency / lockfile updates:**
  - *Rejected:* Modifying toolchains and dependencies can introduce supply chain risks or toolchain drift. Any change to `mise.toml` or `mise.lock` requires explicit human approval.
- **Unrestricted model usage (running all tasks on Frontier/Opus models):**
  - *Rejected:* Extremely wasteful in cost and latency, as demonstrated when charting subagents ran on Opus 5.5 for routine factual searches. Explicit tier assignment and sweep-effort policies optimize token economics and latency.

## Consequences

- Clean interoperability across diverse agent tools and developer setups.
- Guaranteed safety and quality gate parity between human contributors and AI agents.
- Predictable, budget-efficient token consumption across SDLC phases.
- Rigorous traceability from Wayfinder discovery to code implementation.

Source: [How AI agents work in this repo](../../../../.scratch/system-context-harness/issues/21-ai-workflow-in-repo.md)
