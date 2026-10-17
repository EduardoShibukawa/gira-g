---
sidebar_position: 12
---

# Evaluation Scorecard

Gira's extraction accuracy, determinism, and safety guarantees are continuously verified against a pinned multi-service [Reference Corpus](./reference-corpus.md) (`otel-demo-3.1.0-mini`) and its [Perturbed Variant](./perturbed-variant.md) (`otel-demo-3.1.0-perturbed`).

## Reference Corpus & Perturbed Variant Scorecard

| Metric | Published Reference Corpus (`otel-demo-3.1.0-mini`) | Perturbed Variant (`otel-demo-3.1.0-perturbed`) | Enforcement |
|---|---|---|---|
| **Precision** | **100.0%** | **100.0%** | Zero false-positive nodes or relationships |
| **Recall** | **82.4%** | **81.2%** | Ground-truth boundary & relationship coverage |
| **Determinism** | **100.0%** | **100.0%** | Byte-identical `macro-map.json` across repeated runs |
| **Unsupported Safety** | **100.0%** | **100.0%** | Out-of-scope files explicitly surfaced, never guessed |
| **Dangling Edges** | **0** | **0** | Every relationship endpoint resolves to a verified node |

Why evaluate both corpora?
- **Published Reference Corpus (`otel-demo-3.1.0-mini`):** Measures extraction fidelity on realistic polyglot microservice boundaries, contracts, and compose deployments.
- **Perturbed Variant (`otel-demo-3.1.0-perturbed`):** Systematically renames services, contracts, and edges so that any reliance on LLM training memory or hardcoded heuristics fails immediately.

## Extractor Support Matrix

Every fact in the [Macro Map](./macro-map.md) is produced by a deterministic, LLM-free [Extractor](./extractor.md) backed by an [Evidence Record](./evidence-record.md):

| Artifact / Ecosystem | Extracted Facts | Derivation Tier |
|---|---|---|
| **Protobuf / gRPC (`.proto`)** | Service contracts, RPC methods, message schemas | `explicit_contract` |
| **OpenAPI / Swagger (`.yaml`, `.json`)** | HTTP REST contracts, paths, operations | `explicit_contract` |
| **AsyncAPI & GraphQL** | Event-driven channels/topics, GraphQL schemas & operations | `explicit_contract` |
| **Go (`go.mod`, `.go` AST)** | Services, libraries, imports, gRPC/HTTP bindings, call edges | `static_ast` |
| **Python / TypeScript / Java / Rust / C# / C++ / Ruby / PHP / Kotlin** | Service entrypoints, package manifests, intra-service call & import graphs | `static_ast` |
| **Docker Compose & Kubernetes / Helm** | Runtime services, declared ports, environment wiring, infrastructure dependencies | `deployment_wiring` |
| **GitHub Actions & CI Workflows** | Governance tasks, build/release verification pipelines | `deployment_wiring` |
| **Out-of-scope files (`.md`, `.css`, `.svg`, unknown formats)** | Recorded as `UNSUPPORTED` with explicit reason (`gira locate`) | Never guessed |

## Adversarial & Hygiene Gates

Gira enforces fail-closed adversarial gates in CI and release preflight:

1. **Zero Hallucinated Edges:** No relationship can be emitted without a concrete file-and-line `Locator` pinned to an indexed commit SHA.
2. **Unsupported Boundary Honesty:** Files not covered by a deterministic extractor are listed in `unsupported` records rather than heuristically inferred.
3. **Byte-Identical Canonical Serialization:** `MacroMap.CanonicalJSON()` sorts all nodes, relationships, contracts, and evidence records deterministically.
4. **Artifact Digest & Gate Verification (`F1–F5`):** `gira eval`, `gira verify`, `gira review --fix-check`, and `gira gate --reverify` ensure evaluation outputs and L3 findings cannot be self-attested.

## Related

- [Reference Corpus](./reference-corpus.md)
- [Perturbed Variant](./perturbed-variant.md)
- [Extractor](./extractor.md)
- [Evidence Record](./evidence-record.md)
- [CLI: Governance & Verification](../cli/governance.md)
