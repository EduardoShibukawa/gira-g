---
sidebar_position: 9
---

# Extractor

## Definition

The part of the harness that reads one kind of artifact — a data format, a
tool's files (a CI workflow, a deployment manifest), or a programming
language — and emits facts with evidence for the Macro Map. Anything no
Extractor covers is reported as unsupported, never guessed.

**Avoid**: parser, plugin, scanner, analyzer

## Notes

Extractors are deterministic and LLM-free (ADR 0001). They run against the
Target System's pinned commit and emit facts with full evidence locators.

## Related

- [Target System](./target-system.md)
- [Macro Map](./macro-map.md)
- [Evidence Record](./evidence-record.md)
