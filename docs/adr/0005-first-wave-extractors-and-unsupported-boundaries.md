# First wave of extractors and explicit unsupported boundaries

In the MVP of Gira-G (`gira`), Extractors are compiled directly into the Go binary behind a single internal data contract interface (`Extractor` -> `[]Fact` with `EvidenceRecord`). Extraction is structured strictly into three layers (ADR 0002):
1. **Layer 1: Data Formats** (generic Go parsers for YAML, JSON, TOML, protobuf, HCL).
2. **Layer 2: Tool Meaning / Declarations** (semantic domain rules on top of data formats).
3. **Layer 3: Programming Languages** (code AST parsing via tree-sitter CGO binding behind an internal parse-tree seam).

To guarantee that the Reference Corpus (OTel Demo 3.1.0) and dogfooding (Gira-G repository) are mapped deterministically without hallucination, the MVP implements a concrete set of extractors across these layers, while explicitly emitting `unsupported` status descriptors for any unhandled languages, files, or constructs.

## Considered Options

- **Full Helm Template Evaluation in-process via Helm v4 SDK vs Direct YAML Values & Chart Metadata Ingestion.**
  - *Adopted:* Direct ingestion of `Chart.yaml`, `values.yaml`, and values overlays via Layer 2 YAML extractor; full in-process rendering of complex Go templates (`helm.sh/helm/v4`) is deferred beyond MVP.
  - *Rationale:* In the Reference Corpus (`charts/opentelemetry-demo`), service names, dependencies, port allocations, and service-to-service endpoints (`*_ADDR`, `KAFKA_TOPIC`) are explicitly declared in `values.yaml` and `Chart.yaml`. Direct structured values parsing extracts 100% of the corpus's deployment evidence without dragging in Kubernetes template rendering complexities or simulated cluster states.
- **Support all 12 OTel Demo languages via tree-sitter in MVP.**
  - *Rejected:* Dragging in 12 tree-sitter C grammars in the MVP significantly increases CGO build complexity, binary size, and grammar divergence risks.
  - *Adopted:* Focused code extraction on **Go** (for dogfooding and OTel Demo Go services like `checkout`, `product-catalog`) and **gRPC Protobuf** (`.proto` files for the core contracts). All remaining languages (Java, TypeScript, Python, C#, Rust, Ruby, PHP, Elixir, C++, Kotlin, Swift) are explicitly recorded as `unsupported` with discrete `status: unsupported` markers.
- **Dynamic or LLM-based Fallback for Unknown Languages.**
  - *Rejected:* ADR 0001 and ADR 0004 establish that the harness has zero model inference and is 100% deterministic. Any file without an active compiled-in Extractor is reported as `unsupported`, never guessed or heuristically summarized.

## Consequences

- The build remains clean and deterministic with CGO tree-sitter limited to Go and Protobuf native parser.
- Services in unsupported languages within the Reference Corpus still exist in the Macro Map as `Service` nodes because they are discovered via Layer 2 deployment manifests (Docker Compose and Helm/K8s values), but their internal code calls are marked `status: unsupported`.
- Gira-G dogfooding is fully covered (Go codebase, `go.mod`, GitHub Actions workflows, `mise.toml`).

Source: [First wave of extractors](../../../../.scratch/system-context-harness/issues/13-first-extractor-wave.md)
