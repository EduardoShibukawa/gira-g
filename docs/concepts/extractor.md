---
sidebar_position: 9
---

# Extractor

## Definition

The part of the harness that reads one kind of artifact — a data format, a
tool's files (a CI workflow, a deployment manifest), or a programming
language — and emits facts with evidence for the Macro Map. Anything no
Extractor covers is reported as unsupported, never guessed.


## Notes

Extractors are deterministic and LLM-free (ADR 0001). They run against the
Target System's pinned commit and emit facts with full evidence locators.

## Unsupported by design

Website and docs sources are intentionally out of extractor scope —
never guessed, reported as unsupported:

- `website/docs/**` and `website/versioned_docs/**` (`.md` prose)
- `website/docusaurus.config.ts`, `website/sidebars.ts` (TypeScript config)
- `website/src/css/custom.css`, `website/static/img/*` (CSS/SVG identity)
- repo skill docs (`.agents/skills/**`, `.claude/skills/**`)

`bin/gira locate <query>` surfaces these as `UNSUPPORTED` matches with
reasons, so map consumers can tell "known-absent" from "not indexed".
Adding an extractor is an architecture change (new facts, new evidence),
not hygiene.

## Related

- [Target System](./target-system.md)
- [Macro Map](./macro-map.md)
- [Evidence Record](./evidence-record.md)
