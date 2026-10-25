---
sidebar_position: 14
---

# `gira outline`

Extract symbol outlines, function and method signatures, and line ranges for a source file using Layer 3 AST extractors.

## Synopsis

```text
gira outline <file> [--format text|json]
```

## Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--format` | string | `text` | `text` \| `json` |

## Description

`gira outline` extracts classes, interfaces, types, functions, methods, and their exact line boundaries from source files using Layer 3 language AST extractors (Go, Python, TypeScript). It allows autonomous agents and developers to understand the structure of a file in minimal tokens without reading the entire file.

## Examples

### Text Format (Agent-friendly outline)

```bash
gira outline src/cli.ts
```

### JSON Format

```bash
gira outline src/cli.ts --format json
```
