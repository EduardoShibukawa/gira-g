# Cross-file resolution and declared environment wiring

**Status:** accepted (2026-10-13, by owner). This ADR amends [ADR 0005](0005-first-wave-extractors-and-unsupported-boundaries.md): extractors stay per-file, and one deterministic resolution step is added after them. Specified in `specs/gira-impact-precision.md`.

ADR 0005 puts every Extractor behind one contract, `Extract(file) -> facts`. Some architectural facts can't be read from one file:

- The value of a compose `environment` entry often lives in the project `.env`.
- The Service a Dockerfile builds is named by the compose service that references it, not by its directory.
- A Go address field is bound to an environment variable in one place and used in an HTTP request somewhere else. The variable's value, which names the target Service, is in compose or `.env`.
- A Kafka topic name comes from the producer's code default or from its environment, and the consumers are known only through their own wiring.

## Decision

1. **Extractors stay per-file.** They keep parsing one file and emitting nodes, edges and unsupported boundaries with evidence. They may also emit *resolution hints*: typed, internal records carrying a locator. Examples are "this Go package binds field F to env var X", "this request uses the address in F plus the route R", and "this compose service declares env var X with the value V at line L". Hints never reach the Macro Map directly.
2. **One resolution step runs after every file is extracted.** It reads the hints and the per-file facts, in sorted order. It emits the joined facts: env values resolved through `.env` interpolation, Dockerfile attribution to compose services, HTTP and env-wiring edges, and topic Contracts with their producers, broker and consumers. It is pure and has no I/O.
3. **No guessing.** A value the step can't resolve stays unresolved, and no fact is emitted from it. A fact whose direction no source states, such as a topic consumer in an unsupported language, is emitted only with origin `inferred`.
4. **Origins.**
   - `declared`: wiring read from compose or `.env`.
   - `extracted`: facts read from Go source.
   - `inferred`: facts the step concludes without a source that states them.

   Each joined fact keeps the locator of the line it rests on.
5. **Unsupported boundaries are untouched.** The step reads only hints from supported extractors. Unsupported-language source still produces only `unsupported` markers.

## Considered options

- **Every extractor reads its siblings, for example compose reading `.env`.** Rejected: it duplicates parsing, it couples extractors to repository layout, and it breaks the per-file seam that keeps them testable.
- **Resolve in the slices engine at query time.** Rejected: the index would no longer hold the facts. `gira diff`, tours and exports would each need the same logic, and the evidence would not be in `macro-map.json`.

## Consequences

- The index gains HTTP and env-wiring edges and Kafka topic Contracts, each with a locator.
- Dockerfile-built Services take their compose names.
- Determinism is preserved: the step is a pure function of sorted inputs, and index determinism is measured on every change (ADR 0007).
