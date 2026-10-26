---
sidebar_position: 1
---

# `gira ask`

Query your codebase architecture, blast radius, and workflows in plain natural language with local zero-token fallback and on-demand Mermaid diagrams.

## Synopsis

```text
gira ask "<question>" [--diagram] [--json] [--index <path>]
```

## Description

`gira ask` (aliased as `gira query`) bridges the gap between human questions and deterministic code intelligence. It evaluates developer questions against the indexed Macro Map, architectural decisions (ADRs), and contract definitions without requiring expensive cloud LLM subscriptions or leaking proprietary code.

### Dual-Mode Execution:
1. **Local Deterministic Fallback (Zero-Token):** Operates 100% offline and air-gapped, matching queries against symbols, callers, and entrypoints in <5ms.
2. **Mermaid Diagram Generation (`--diagram`):** Automatically synthesizes syntactically valid Mermaid flowcharts visualizing caller/callee relationships or onboarding flows.
3. **Structured JSON Output (`--json`):** Emits standardized responses suitable for AI agent context consumption.

## Options

- `--diagram`: Include a syntax-valid Mermaid flow diagram in the output.
- `--json`: Output answer as structured JSON.
- `--index <path>`: Path to a specific Macro Map index file (defaults to `.gira/macromap.json`).

## Examples

### 1. Project Purpose & Onboarding
```bash
gira ask "Qual é o objetivo principal deste projeto?"
```

### 2. Quickstart with Architecture Flow Diagram
```bash
gira ask "Como eu inicializo e utilizo o sistema?" --diagram
```

### 3. Blast Radius & Impact Analysis
```bash
gira ask "O que quebra se eu alterar o pacote de persistência?" --diagram
```
