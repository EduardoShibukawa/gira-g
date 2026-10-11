# Evaluation design, benchmark tasks, and go/no-go thresholds

Gira-G (`gira`) validates its architectural extraction and context slice delivery through an automated, repeatable evaluation harness against the pinned Reference Corpus (OpenTelemetry Demo `3.1.0` + Helm chart `0.42.3`) and internal dogfooding. To guard against model memorization, the definitive go/no-go gate is measured strictly against the **Perturbed Variant**.

## Decision Summary

1. **No-Harness Baseline:**
   - **Agent Baseline:** An unassisted coding agent equipped with standard workspace tools (grep, glob, file read, bash) prompted with the exact same change request to identify the Impact Set.
   - **Human Baseline:** A developer manually inspecting the repository tree, manifests, and search tools to compile the Impact Set, recording elapsed time-to-answer.
   - **Prior Art Isolation:** Neither Graphify nor Understand Anything is used in the evaluation pipeline (per ADR 0001 and Ticket 10), preserving clean dependency isolation.

2. **Benchmark Suite (12 Representative Tasks on Reference Corpus):**
   - **Scenario 6 (Endpoint Change Pre-Context):**
     - Task 1: Modifying Go gRPC handler in `checkoutservice` (`PlaceOrder`).
     - Task 2: Modifying HTTP frontend handler (`/cart`).
   - **Scenario 3 (Contract Change Impact):**
     - Task 3: Modifying `PaymentService` method in `demo.proto` -> identifies consumers.
     - Task 4: Modifying `CurrencyService` method in `demo.proto` -> identifies `frontend` and `checkoutservice`.
   - **Scenario 4 (Event Streaming & Freshness):**
     - Task 5: Modifying schema on Kafka topic `orders` -> publisher `checkoutservice`, consumers `accountingservice` & `frauddetectionservice`.
     - Task 6: Audit topic freshness and commit pin alignment (`dedc017`).
   - **Scenario 1 (API, Event, and File Ownership):**
     - Task 7: Ownership and contract inventory for `productcatalogservice` (Go).
     - Task 8: Ownership and contract linkage of shared contract `pb/demo.proto`.
   - **Scenario 8 (Traps, Conflicts & Unsupported Boundaries):**
     - Task 9: `EmailService` trap (proto gRPC vs actual HTTP POST callsite conflict).
     - Task 10: `ShippingService` trap (proto gRPC vs actual HTTP POST callsite conflict).
     - Task 11: Deployment divergence trap (Compose excludes `agent`/`mcp`/`chatbot`; Helm includes them).
     - Task 12: Language boundary audit (`currencyservice` in C++, `paymentservice` in JS) -> must report `status: unsupported`.

3. **Evaluation Metrics:**
   - **Impact Set Precision & Recall:** Scored component-by-component and edge-by-edge against ground truth.
   - **Evidence Locator Correctness:** Percentage of evidence locators (`repo@commit:path#Lstart-Lend`) pointing to exact, verifiable source lines.
   - **Token Consumption:** Input and output tokens consumed by the agent to reach the final Impact Set.
   - **Unsupported Language Safety (Zero Hallucination):** Rate of unhandled language ASTs falsely reported as resolved (must be strictly 0%).
   - **Harness Query Latency & Build Time:** Time to generate a slice (`<50ms`) and time for cold index creation (`<15s`).
   - **Human Time-to-Answer:** Measured reduction in time required by a developer to produce an accurate briefing.

4. **Model Comparison Grid:**
   - **Local / Small:** Llama 3.1 8B, Qwen 2.5 7B (quantized on local 32GB developer machine).
   - **Frontier:** Claude 3.5 Sonnet, GPT-4o (remote APIs).
   - *Note:* The harness core has zero model client (ADR 0001); models are evaluated strictly as consuming agents receiving context slices vs raw grep.

5. **Perturbed Variant & Ground-Truth Storage:**
   - **Perturbed Variant:** An automated script (`scripts/perturb-corpus.sh`) deterministically renames services, Kafka topics, env vars, proto packages, and alters selected topological edges.
   - **Memory Leak Protection:** Evaluated agents are strictly prohibited from reading `AGENTS.md`, `README.md`, or `docs/` within the corpus to prevent learning pre-existing architectural documentation.
   - **Ground-Truth Storage:** Stored as declarative JSON fixtures in `fixtures/evaluation/reference-corpus/{ground-truth.published.json, ground-truth.perturbed.json}`, version-pinned to commit SHAs.

6. **Go/No-Go Acceptance Thresholds (Evaluated on Perturbed Variant):**
   - **Impact Set Recall:** $\ge 90\%$ on Perturbed Variant.
   - **Impact Set Precision:** $\ge 80\%$ on Perturbed Variant.
   - **Token Reduction:** $\ge 60\%$ reduction compared to no-harness baseline.
   - **Network Isolation:** $100\%$ compliance (0 network egress beyond declared git fetch, verified in an offline sandbox).
   - **Unsupported AST Hallucination:** $0\%$ (100% of non-Wave-1 languages reported as `status: unsupported`).
   - **Index Determinism:** $100\%$ identical index output across multiple runs from the same commit.

## Considered Options

- **Evaluate go/no-go on the Published Variant alone.** Rejected: frontier LLMs have the OpenTelemetry Demo in their training pre-training weights, obscuring whether success is due to harness context slices or model memorization.
- **Include Graphify or Understand Anything in the automated benchmark pipeline.** Rejected: violates ADR 0001 and Ticket 10 by introducing external non-deterministic runtime dependencies into CI.
- **Allow consuming agents to read reference corpus documentation/READMEs.** Rejected: creates a memory-leak vector where agents read high-level architecture docs instead of testing harness evidence extraction.
- **Single monolithic macro F1 score.** Rejected: hides false negatives on critical consumer boundaries and masks hallucinated edges on unsupported languages.

## Consequences

- The Perturbed Variant serves as the uncheatable arbiter of harness utility.
- Small/local models (Llama 8B, Qwen 7B) gain a fair benchmark demonstrating whether deterministic context slices enable them to match frontier model performance on multi-service tasks.
- CI includes a network-isolated verification test enforcing zero egress.

Source: [Evaluation design and go/no-go thresholds](../../../../.scratch/system-context-harness/issues/17-evaluation-design.md)
