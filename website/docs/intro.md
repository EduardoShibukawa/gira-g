---
sidebar_position: 1
slug: /
---

# Gira

Gira is a CLI tool that builds and maintains an evidence-backed architectural map
of any software system — from a single repository to many — and serves selected
slices of it to humans and AI coding agents.

## Documentation

This documentation covers:

- **Concepts** — the vocabulary, mental model, and [Evaluation Scorecard](./concepts/scorecard) behind Gira
- **CLI Reference** — every command, flag, and output (`init`, `index`, `impact`, `slice`, `explain`, `tour`, `flows`, `doc`, `diff`, `locate`, `export`, `view`, `skills`, `serve mcp`, and governance tools)
- **Guides** — quickstart, workflows, and evidence inspection

## Quick start

```bash
# Option A: Zero-config single repository (generates gira.yaml pinned to HEAD and indexes)
gira init . --output .gira/macro-map.json

# Option B: Multi-repository manifest (reads gira.yaml; defaults to ~/.cache/gira/indexes/<system>/macro-map.json)
gira index --output .gira/macro-map.json

# Ask what breaks if you change a target (auto-resolves .gira/macro-map.json, macro-map.json, or user cache when --index is omitted)
gira impact <target_id> --index .gira/macro-map.json

# Explore a bounded context slice or guided reading tour
gira slice <node_id> --tier contracts --view architecture
gira tour --format md
```

See the [Quickstart guide](./guides/quickstart) for a complete runnable walkthrough.
