---
sidebar_position: 2
---

# `gira setup`

One-shot setup that initializes the Gira-G architectural index and installs agent skills and MCP registration.

## Synopsis

```text
gira setup [<dir>] [--output <path>]
```

## Description

`gira setup` combines index creation and agent skills installation into a single deterministic command. It:
1. Initializes the architectural index (`macro-map.json`).
2. Configures agent skills and MCP server integration for the workspace.

## Examples

```bash
gira setup .
```
