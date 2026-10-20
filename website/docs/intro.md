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
# Option A: Zero-config single repository (publishes canonically to versioned store, writes gira.yaml last)
gira init .

# Option B: Multi-repository manifest (indexes gira.yaml into immutable versioned store)
gira index

# Ask what breaks if you change a target (auto-resolves from versioned store using gira.yaml)
gira impact <target_id>

# Or generate a detached snapshot for isolated review or offline use
gira index --output .gira/snapshot.json
gira impact <target_id> --index .gira/snapshot.json

# Explore a bounded context slice or guided reading tour
gira slice <node_id> --tier contracts --view architecture
gira tour --format md
```

See the [Quickstart guide](./guides/quickstart) for a complete runnable walkthrough.
