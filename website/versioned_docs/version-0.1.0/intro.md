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

- **Concepts** — the vocabulary and mental model behind Gira
- **CLI Reference** — every command, flag, and output
- **Guides** — quickstart, workflows, and evidence inspection

## Quick start

```bash
# Index a target system
gira index

# Ask what breaks if you change a target
gira impact <target_id> --index .gira/index.json

# Explore a slice
gira slice <node_id> --tier contracts --view architecture
```

See the [Guides](./guides/quickstart) for a complete walkthrough.
