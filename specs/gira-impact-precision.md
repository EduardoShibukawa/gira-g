---
schema: spec/1
status: approved
version: 1
approved_version: 1
approved_hash: sha256:59aff522efa827b65d6bfb3a98f9b06667278ac51776ff4ce41d2b9fe9c15010
---

# Impact Set precision and recall from the Wave 1 extractors

## Problem

The evaluator is no longer circular (spec `gira-agent-skills-and-eval` v1), so it now shows the real state of the extractors on the unreviewed ground truth:

| Variant | Recall | Precision | Unsupported Safety | Determinism |
|---|---|---|---|---|
| Published, hermetic | 58.8% | 32.8% | 100% | 100% |
| Perturbed, hermetic | 56.2% | 30.5% | 100% | 100% |
| Published, full corpus (opt-in) | 60.3% | 30.4% | 100% | 100% |

The Layer 2 proxy for focused tours scores 5.9% assertiveness (1 of 17 claims verifiable).

The per-task report (`missing` and `unexpected`) points to six causes:

1. **Noise upstream.** Dockerfile build stages become `service:stage:*` nodes, base images become `library:<image>` nodes, and every `library:` reached through `imports` is listed in `upstream_services`. These account for about three quarters of all false positives (tasks 01, 02, 07, 09, 10).
2. **No Kafka topic Contracts.** `contract:kafka:orders` is never emitted. Tasks 05 and 06 fail with "target node not found", and task 01 misses the topic and its consumers (accounting, fraud-detection).
3. **gRPC clients that are built but never called.** Checkout builds `EmailService` and `ShippingService` clients and never invokes them. It reaches email and shipping over HTTP instead. Both contracts are still listed as consumed: this is the trap in tasks 09 and 10.
4. **No HTTP or env-wiring edges.** Go HTTP calls to an address bound from `<SVC>_ADDR` are not extracted. Compose environment wiring that names another service, such as `APPLICATION_ENDPOINT` resolving to the frontend, is not used. So `service:agent` is missing in tasks 02 and 11.
5. **Imprecise locators.** Compose edges cite the service header line instead of the `depends_on` entry or the environment variable. Declared ports have no locator of their own, which caps Layer 2 assertiveness.
6. **Wrong names.** `src/cart/src/Dockerfile` becomes `service:src` instead of the compose service `cart`.

Two facts the causes depend on are absent from the hermetic corpus. The topic name (`KAFKA_TOPIC=orders`) and `APPLICATION_ENDPOINT=${FRONTEND_HOST}:${FRONTEND_PORT}` live in the upstream `.env`, which is not vendored. Compose files reference those variables only by name.

Heavy triggers:

- **H1:** product calls on whether Libraries belong in the Impact Set, how HTTP calls are modelled, and where environment values come from. All three are decided below.
- **H3:** the change spans the Go AST, Compose and Dockerfile extractors, the pipeline, the slices engine, the tours and the evaluation corpus.
- **H5:** six causes, each test-first, decomposed into children.

## Expected behavior

### Impact Set shape

- `upstream_services` and `downstream_services` hold only Service and Infrastructure nodes.
- Library nodes reached through `imports` move to a new additive field, `upstream_libraries`. It appears on the CLI (JSON and Markdown) and on MCP (`get_impact_briefing`, `get_pr_impact`). Budget truncation drops it before the Service lists.
- Dockerfile build stages and base images are never nodes. They are recorded as metadata on the Service the Dockerfile builds (`stages`, `base_images`), each with its own locator in the evidence.

### Cross-file resolution

Extractors stay per-file (ADR 0005). A deterministic resolution step runs after per-file extraction and joins facts across files. It never guesses: anything it cannot resolve stays unresolved and is reported as unknown, never named.

- **Environment values.** A compose `environment` entry gets its value from the inline value, or else from the project `.env`, with `${VAR}` interpolation. The vendored `.env` is the upstream file, byte for byte.
- **Dockerfile attribution.** A Dockerfile that a compose service builds (`build.dockerfile` and `build.context`, after interpolation) belongs to that compose service. A Dockerfile no compose service builds keeps the current directory rule.

### Kafka topic Contracts

- A Go service that builds a producer message whose topic resolves to a name `T` provides `contract:kafka:T`. The evidence is extracted, at the line that sets the topic.
- The topic resolves from a string literal, or else from the environment variable the expression reads. That covers a package-level variable whose initializer returns `os.Getenv("X")` with a literal fallback. The value comes from that service's compose wiring, or else from the literal fallback.
- The broker that the producer's broker address resolves to also provides the topic. That is `service:kafka`, through `KAFKA_ADDR`. The evidence is declared, at the env line.
- Every other compose service wired to the same topic consumes it. A service is wired to a topic when its `KAFKA_TOPIC` resolves to `T`. Its code is not read, so the direction is inferred: origin `inferred`, locator at its `KAFKA_TOPIC` line.

### gRPC consumption

A service consumes a gRPC Contract only when its Go package invokes a method on a client bound to that Contract. A binding is a struct field or variable typed as the client, or assigned from `New<Contract>Client`, including a local alias of such a field. Building a client or declaring the field is no longer enough. Providing (`Register<Contract>Server`) does not change.

### HTTP and env wiring

- **Go HTTP calls.** A Go HTTP request whose URL is an address bound from environment variable `X`, plus a literal route, produces `consumes` from the caller to the compose service that `X`'s resolved host names. The address is bound through `os.Getenv("X")`, or through a call that passes `&field` and the literal `"X"`. The evidence is extracted, at the request line. The edge metadata holds `protocol: http` and the sorted `routes`. No HTTP Contract node is created.
- **Env wiring.** A compose environment entry whose resolved value has a host part naming another compose service produces a declared edge from the declaring service to that service. The edge is `consumes`, or `accesses` for Infrastructure, at the env line.
- **Unsupported code.** No fact is ever extracted from unsupported-language source (ADR 0005). Services in those languages get edges only from declared wiring.

### Locators

- Every compose-derived edge cites the line of its `depends_on` entry or of its environment variable. It never cites the service header.
- Declared ports (Dockerfile `EXPOSE`, compose `ports`) carry a locator on their own line.
- Tour entry steps cite that port locator as `why_locator` instead of saying the locator is unknown.

### Metrics

Each child reports recall, precision, Unsupported Safety and determinism before and after its change, for both hermetic variants. When the full corpus is fetched, it is reported too. Layer 2 assertiveness is reported by the locator child, and by any child that changes it.

## Scope

In scope:

- the six causes above, each delivered test-first: a failing test on the hermetic corpus comes first;
- vendoring the upstream `.env` into the hermetic mini-corpus;
- the perturbation recipe edits needed to keep the Perturbed Variant coherent: the topic value in `.env`, declared like the existing edits;
- regenerating the `gira-query` golden files whose content changes, with the diff explained in the child record;
- ADR 0013 for the cross-file resolution step.

Out of scope:

- changing any ground-truth file, or setting `reviewed: true` (the owner's call);
- HTTP Contract nodes (`contract:http:*`);
- any model client (ADR 0001);
- extracting facts from non-Go source;
- dropping `depends_on`-only edges (start ordering, such as frontend → quote): a recorded follow-up;
- `contract:Health`, the gRPC health service: a recorded follow-up;
- push or release.

Non-goals: reaching the ADR 0007 gate thresholds (90% recall, 80% precision). This spec raises the measured numbers. It does not promise the gate.

## Acceptance criteria

- **SC1 (cause 1).** In both hermetic variants:
  - no node id starts with `service:stage:`, and no Dockerfile base image is a node;
  - every briefing's `upstream_services` and `downstream_services` hold only Service and Infrastructure nodes;
  - Libraries reached through `imports` appear in `upstream_libraries`, byte-identical between CLI and MCP;
  - budget truncation drops `upstream_libraries` before any Service list.
- **SC2 (cause 6).** `src/cart/src/Dockerfile` is attributed to `service:cart` (published) and `service:basket` (perturbed). No `service:src` exists. A Dockerfile that no compose service builds keeps the directory rule (unit test).
- **SC3 (cause 2).** In both hermetic variants:
  - the topic Contract exists (`contract:kafka:orders`, `contract:kafka:order-events`);
  - the producing service provides it, extracted, at a Go line;
  - `service:kafka` provides it, declared;
  - accounting (ledger) and fraud-detection consume it, inferred, at their `KAFKA_TOPIC` line;
  - tasks 05 and 06 produce a briefing;
  - SC6 of the previous spec holds on 12 of 12 tasks.
- **SC4 (cause 3).** Checkout no longer consumes `contract:EmailService` or `contract:ShippingService`. It still consumes Cart, ProductCatalog, Payment and, in the published variant, Currency. Unit tests cover:
  - a constructed but uncalled client;
  - a called field;
  - a called local alias;
  - a call in another file of the same package.
- **SC5 (cause 4).** Checkout consumes email (`/send_order_confirmation`) and shipping (`/get-quote`, `/ship-order`). Each edge is extracted, at the request line, with `protocol: http` and `routes`. `service:agent` consumes `service:frontend` through `APPLICATION_ENDPOINT`, declared, at the env line. No fact cites a path in an unsupported language.
- **SC6 (cause 5).**
  - Every compose-derived edge cites a `depends_on` entry or an environment variable line.
  - Every declared port has its own locator.
  - Tour entry steps cite the port locator as `why_locator`.
  - The Layer 2 focused-tour assertiveness is strictly above 5.9%; the number is reported.
- **SC7 (metrics).** After all children:
  - recall and precision are each strictly above the baseline in the table above, for both hermetic variants;
  - each child's record has a before and after row for both hermetic variants;
  - no child lowers recall or precision of either hermetic variant, unless the owner accepts the drop through a `divergence:` entry.
- **SC8 (invariants).**
  - Unsupported Safety and determinism are 100% in both variants, at every child.
  - Two cold indexes are byte-identical.
  - No model client is added: no LLM SDK in `go.mod`, and no new network client in the indexing path.
- **SC9 (regression).**
  - `go test -race ./...` passes at the root and inside `bench/`.
  - Assertions of existing end-to-end tests are not weakened. They change only where a criterion above tightens them.
  - Ground-truth files are unchanged, and `reviewed` stays false.
- **SC10.** The validator reports no `fail`.

## Interfaces and dependencies

- **Impact briefing (CLI JSON and Markdown, MCP `get_impact_briefing`, `get_pr_impact`).**
  - New field `upstream_libraries`. This is additive.
  - Library nodes leave `upstream_services`. That restores the field's documented meaning: it is named for Services. No other field changes meaning.
- **Index (`macro-map.json`).**
  - Removed: `service:stage:*` nodes and base-image `library:` nodes.
  - New nodes: `contract:kafka:*` Contracts.
  - New edges: HTTP and env-wiring edges.
  - New metadata: `stages`, `base_images`, `protocol`, `routes`.
  - Compose edge locators move to the `depends_on` or env line.
  - Consumers that diff indexes (`gira diff`) see these as changes between an old and a new index, which is expected.
- **Tours.** Entry steps gain a `why_locator`. The template text says the port's locator instead of "locator unknown".
- **CLI flags and MCP tool names and parameters.** None change.
- **Evaluation corpus.**
  - The hermetic mini-corpus gains the upstream `.env`, byte-identical at `upstream_sha`.
  - The perturbation recipe gains one edit: the `KAFKA_TOPIC` value.
  - The measured grep baseline changes slightly because the corpus grows. This is reported.
- **`gira-query` goldens.** Regenerated where the content changes.
- **External systems.** None at index time. Fetching the full corpus stays opt-in.

## Decisions

- **Libraries in the Impact Set.** Moved to an additive `upstream_libraries` field. Rejected:
  - removing them entirely, because agents lose the signal;
  - having the evaluator ignore `library:`, because that games the metric.

  Decided by the owner (2026-10-13).
- **HTTP calls.** A direct service-to-service `consumes` edge, with the route in metadata. No `contract:http:*` nodes, so the unreviewed ground truth stays valid (task 09 expects no Contract). Decided by the owner (2026-10-13).
- **Environment values.** The upstream `.env` is vendored byte-identical and resolved as declared evidence. Inferring values from variable names alone was rejected as weaker evidence. Decided by the owner (2026-10-13).
- **Cross-file resolution.** A resolution step after per-file extraction, recorded in ADR 0013 as an amendment to ADR 0005. Rejected: making every extractor read sibling files, because it duplicates parsing and breaks the per-file seam.
- **Topic direction for unsupported-language consumers.** Origin `inferred`, because their code is not read. Labelling them `declared` would claim a direction no source states.
- **The broker as topic provider.** `service:kafka` provides the topic, so a Contract-targeted briefing lists it upstream, as the ground truth expects. The evidence is the producer's broker env wiring.
- **Child order.** The children run in dependency order, not in the order of expected impact: 1, 6, 2, 3, 4, 5. Causes 6, 2 and 4 need the cross-file step, which child 02 introduces together with the naming fix.

## Risks and constraints

- **Over-matching env values.** A value such as `localhost` or a URL to an external host could match a service name. The rule matches only the host part, exactly, against declared compose service names. Self-references are skipped.
- **Perturbation coherence.** The `.env` content is renamed together with the corpus. The topic value needs its own recipe edit, or the published and perturbed topics would disagree.
- **Golden churn.** Goldens change by design. Each child explains its golden diff, so an accidental change is visible.
- **Determinism.** The resolution step iterates in sorted order. Determinism is measured at every child (SC8).
- **No protected actions.** Nothing is pushed, released or sent to a paid run.

## Verification

- Per child: a failing test first on the hermetic corpus, then the fix. The child's `AC` rows cite these criteria. Its record has before and after metric rows for both hermetic variants.
- Integrated: SC1–SC10 on the final head, plus `go test -race ./...` at the root and in `bench/`, and the opt-in full-corpus run, reported only.
- Review focus:
  - no unresolved fact is named (no guessing);
  - no fact cites unsupported-language source;
  - inferred versus declared versus extracted labels are correct;
  - no test assertion was weakened.

## Rollback

skip: every change is a code or fixture change in this repository, revertible with `git revert`. No migration, publication or external state.

## Open questions

None.

## Revisions

- v1 (2026-10-13): initial version, from the owner's request and the four decisions above.
