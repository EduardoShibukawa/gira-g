---
sidebar_position: 1
---

# Quickstart

Get from a clean checkout to your first Impact Briefing in three steps.

## 1. Install

```bash
# Download the gira binary (or `go build ./cmd/gira`)
gira version
```

## 2. Build an index

```bash
cd /path/to/target-system
gira index --output .gira/index.json
```

This runs the Extractors against your manifest and dependencies, writing
`macro-map.json`.

## 3. Ask the first question

```bash
# What breaks if you change a Component?
gira impact <component-id> --index .gira/index.json

# Or analyze a diff
gira impact --from-diff HEAD~1..HEAD --index .gira/index.json

# Extract a slice
gira slice <component-id> --tier summary --index .gira/index.json
```

## Next steps

- [Change-maker workflow](./change-maker-workflow.md)
- [Evidence inspection](./evidence-inspection.md)
- [Concepts](../concepts/target-system.md)
