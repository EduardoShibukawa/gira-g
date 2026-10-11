# Macro Map schema and evidence model

The Macro Map represents a Target System at the Component level using a discrete, deterministic property graph with explicit provenance on every node and relationship. The serialized schema defines four concrete node kinds (`Service`, `Library`, `Infrastructure`, `Contract`) and four relationship verbs (`provides`, `consumes`, `imports`, `accesses`). Inter-service RPC and messaging are mediated through `Contract` nodes (`Service A -provides-> Contract <-consumes- Service B`), enabling direct O(1) inbound lookup of consumers for any contract change without guessing. Every node and relationship carries one or more `EvidenceRecord` entries containing locators (repo, commit SHA, path, line range), deployment profile tags (e.g. Compose vs Helm), discrete origin tiers (`declared`, `extracted`, `inferred`), and status flags (`current`, `stale`, `unverified`).

## Considered Options

- **Keep `Component` as the sole concrete node type in the schema.** Rejected: creates ambiguity across export targets (C4 Containers vs Components, Backstage services vs libraries) and forces extractors to inspect ad-hoc metadata fields. `Component` remains the conceptual domain umbrella in `CONTEXT.md`, while concrete nodes in the schema are `Service` and `Library`.
- **Direct `calls` edges between services without `Contract` nodes.** Rejected: loses contract provenance, creates dangling edges when consumers are extracted before providers (or in asynchronous broker topologies), and makes contract-impact analysis expensive. Derived `calls` edges can be computed virtually on demand.
- **Probabilistic confidence score (0.0 to 1.0).** Rejected: the MVP contains no LLM step (ADR 0001); numeric confidence gives an illusion of precision without auditable grounding. Replaced by a discrete derivation rubric (`declared`, `extracted`, `inferred`).
- **Single evidence record per edge (overwriting on subsequent discoveries).** Rejected: drops critical provenance when declarations (e.g. Helm values) and source code disagree, or when multiple callsites invoke the same contract.

Source: [Minimum Macro Map schema and evidence model](../../../../.scratch/system-context-harness/issues/12-macro-map-schema.md)
