---
sidebar_position: 7
---

# `gira docs`

Generate a complete, modern project documentation site and interactive visual architecture explorer without external dependencies or heavy JavaScript runtimes.

## Synopsis

```text
gira docs [--output <dir>] [--index <path>] [--root <dir>] [--title <string>] [--serve <port>]
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--output` | Target directory to export static documentation site | `build` |
| `--index` | Path to `macro-map.json` for graph generation | Auto-detected |
| `--root` | Repository root directory | Git top-level or `.` |
| `--title` | Documentation portal title | `Gira-G Architecture Harness` |
| `--serve` | Optional local HTTP server port to preview site immediately (e.g. `8080`) | None |

## Overview

The `gira docs` command compiles all repository documentation into a modern, static, zero-bloat documentation site ready for publication to GitHub Pages or internal static hosting.

### Key Features

1. **Spring Reference Living Architecture**: Compiles high-density architectural overviews, community clustering, centrality rankings, and Mermaid sequence flows.
2. **Interactive Visual Explorer (Cytoscape)**: Generates `architecture/graph.html` featuring interactive codebase exploration, community color-coding, centrality-weighted sizing, and a slide-over node inspector.
3. **Zero-Latency Client-Side Search**: Automatically builds `search-index.json` and embeds a modal fuzzy search (`⌘K` / `Ctrl+K`) indexing all pages, headings, and symbols.
4. **Clean Static Generation**: Native Go compiler with `goldmark` and GFM extensions. Requires zero Node.js/npm dependencies, replacing heavyweight generators like Docusaurus.

## Examples

Build documentation into `build/` directory:

```bash
gira docs --output build
```

Build and preview locally with the built-in HTTP server:

```bash
gira docs --output build --serve 8080
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Parse or build error |
