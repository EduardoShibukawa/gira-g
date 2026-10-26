---
sidebar_position: 6
---

# `gira architecture`

Generate high-density, living architecture reference documentation following the Spring Framework Reference Overview standard.

```text
gira architecture [--format md|html|json] [--output <path>] [--index <path>]
```

| Flag | Description |
|---|---|
| `--format` | `md` (default), `html`, or `json` |
| `--output` | Path to write the output document (e.g. `ARCHITECTURE.md`) |
| `--index` | Path to `macro-map.json` (optional; falls back to local repository AST scan) |
| `--checkout` | Local checkout path for git freshness comparison |

## Overview

The `gira architecture` command synthesizes living architectural documentation directly from the codebase AST and MacroMap topology:

1. **Executive Overview & Core Philosophy**: Invariants, append-only Merkle hashing, zero-token fast pass, and deterministic AST manipulation.
2. **Layered Architecture Block Diagram**: Interactive Mermaid diagram showing all 6 decoupled architectural layers.
3. **Architectural Communities & Modules Breakdown**: AST nodes clustered by domain (Core Engine, AST Manipulator, Extractors, CLI/MCP Entrypoints, Governance/Gate, Storage) with Centrality scores and Core Abstraction vs Utility classifications.
4. **Critical Lifecycles & Intent Flows**: End-to-end execution paths with Mermaid sequence diagrams (Surgical AST Edit, Impact Calculation, Extraction/Indexing, Governance Gate).
5. **API Contracts & Module Catalog**: Direct clickable file citations (`file://...#L...`) to every core type, struct, and exported function.

## Output Formats

- **`--format md`**: Standalone, GitHub-flavored Markdown optimized for human engineers and repository documentation.
- **`--format json`**: Compact, token-efficient JSON topology for AI agents (Hermes, Claude Code, OpenCode).
- **`--format html`**: Single-file, zero-dependency standalone HTML document with embedded styling.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Parse or write error |
