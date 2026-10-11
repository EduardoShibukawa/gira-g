# Context slice shape and delivery model

Gira-G (`gira`) serves context slices as deterministic, token-budgeted, provenance-backed subgraphs rendered in two canonical formats: structured JSON (for agent/programmatic callers over stdio MCP and CLI `--json`) and compact Markdown text (for human terminal consumption and agent scratchpads). Slices follow a "concise by default, depth on demand" progressive disclosure model.

## Decision Summary

1. **Progressive Disclosure via 3 Depth Tiers:**
   - **Tier 1 (Default: `summary` / ~250–500 tokens):** Emits the immediate Impact Set (the focal Component/Contract, directly connected Components, connecting Contracts, and discrete derivation status). Locators provide file paths and line ranges, without AST syntax snippets or contract bodies.
   - **Tier 2 (`contracts` / ~500–1,500 tokens):** Expands the slice to include full Contract boundary definitions (gRPC Protobuf RPC method signatures, input/output message schemas, HTTP route verbs/paths, event schemas).
   - **Tier 3 (`deep` / ~1,500–3,500 tokens):** Includes specific extracted callsites, environment variable wiring mappings, and discrete `EvidenceRecord` entries for every supporting discovery across deployment profiles (Compose, Helm, Code).

2. **Dual Canonical Representation:**
   - **Structured JSON (`SliceResponse`):** Machine-readable canonical schema carrying discrete node objects, directional relation edges, 1:N discrete evidence locators, and an explicit `unsupported` inventory list.
   - **Compact Markdown (`--format=md`):** Token-optimized text rendering designed for direct injection into agent prompt caches and human terminal readability, using bulleted node hierarchies, explicit ASCII edge arrows (`-provides->`, `<-consumes-`), and locator citations (`[repo@commit:path#Lstart-Lend]`).

3. **Discrete Uncertainty, Staleness, and Unsupported Status:**
   - No floating-point confidence scores. Every relation and node carries a discrete origin (`declared`, `extracted`, `inferred`) and status (`current`, `stale`, `unverified`).
   - If a Component uses an unhandled language or missing manifest, it is surfaced as an explicit `unsupported` or `unknown` notice with actionable remediation pointers (e.g. which manifest or contract to inspect manually).

## Considered Options

- **Single large raw graph dump (Graphify default).** Rejected: blows up token limits (150k+ tokens), invalidates agent context windows, and forces the consuming LLM to perform graph parsing that the Go harness can execute in <5ms.
- **Natural language LLM-generated summaries inside the harness.** Rejected: violates ADR 0001 and ADR 0004 (100% deterministic, zero model client in the harness). Summaries are produced by structured template renderers in Go.
- **Continuous numeric confidence scores (0.0 to 1.0).** Rejected: ADR 0003 established that continuous scores give an illusion of precision and cause models to hallucinate thresholds. Discrete origins (`declared`, `extracted`, `inferred`) are unambiguous.

## Consequences

- Agents receive byte-stable, token-efficient context slices that fit easily within prompt caching boundaries.
- Humans and agents query the exact same underlying slice engine through CLI (`gira impact`, `gira slice`) and MCP tools (`get_impact_briefing`, `get_context_slice`).

Source: [What should a context slice look like?](../../../../.scratch/system-context-harness/issues/15-context-slice-shape.md)
