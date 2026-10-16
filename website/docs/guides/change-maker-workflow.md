---
sidebar_position: 2
---

# Change-maker workflow

The Change-maker flow is: **brief → edit → verify**.

## 1. Brief

Run `gira impact` (or `gira slice --tier contracts`) on the thing you are
about to change. Read the affected Components, Contracts, and consumers with
their evidence locators.

```bash
gira impact <target-id> --index .gira/index.json
```

## 2. Edit

Make your change. Use `gira slice --tier deep` for the source context you
need; never guess about unseen callers — consult the Impact Set.

## 3. Verify

Re-run the index, then confirm no new consumers appeared:

```bash
gira index --output .gira/index.json
gira impact <target-id> --index .gira/index.json
```

Cross-check with `gira explain` to inspect Evidence Records behind any edge you
doubt.

## See also

- [Quickstart](./quickstart.md)
- [Evidence inspection](./evidence-inspection.md)
- [Concepts: Impact Set](../concepts/impact-set.md)
