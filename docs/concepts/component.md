---
sidebar_position: 2
---

# Component

## Definition

The unit of the Macro Map: a deployable Service, or a module/package boundary
declared in a manifest or workspace. Never an arbitrary directory, class, or
function.

**Avoid**: module, node, unit. Not a C4 Component (an in-process part) — a
Service maps to a C4 Container.

## Kinds

- **Service** — independently deployable
- **Library** — a declared module/package boundary, consumed in-process
- **Infrastructure** — shared stateful store, broker, or middleware

## Related

- [Service](./service.md)
- [Library](./library.md)
- [Infrastructure](./infrastructure.md)
- [Macro Map](./macro-map.md)
- [Contract](./contract.md)
