---
sidebar_position: 1
---

# Target System

## Definition

Any software system the harness maps, whether it lives in one repository or spans many.

**Avoid**: the codebase, the project, the estate

## Notes

A Target System may be monolithic (a single repository), multi-service (a
microservice estate), or mixed (application + infrastructure + shared
libraries). Gira never guesses what constitutes a Component — it relies on
declared boundaries (manifests, workspaces) or declared services.

## Related

- [Macro Map](./macro-map.md) — the default view of a Target System
- [Component](./component.md) — the unit of the Macro Map
- [Extractor](./extractor.md) — how facts are obtained from a Target System
