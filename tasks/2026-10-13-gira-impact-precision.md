---
schema: task/1
tier: heavy
status: verified
spec: gira-impact-precision@1
verified_at: "99bb529"
---

# Impact Set precision and recall from the Wave 1 extractors

## Routing and approvals

- tier: heavy (H1: Library in the Impact Set, HTTP modelling and the source of env values are product calls; H3: spans the Go AST, Compose and Dockerfile extractors, the pipeline, slices, tours and the eval corpus; H5: six causes, each test-first, decomposed into children)
- confirmed: heavy by owner (2026-10-13, "Heavy (Recommended)")
- decided: do Library nodes stay in the Impact Set → new additive field `upstream_libraries`; Dockerfile stages and base images become Service metadata by owner (2026-10-13, "New additive field (Recommended)")
- decided: how HTTP calls are modelled → service-to-service consumes edge with the route in metadata, no `contract:http:*` by owner (2026-10-13, "Service edge + route (Recommended)")
- decided: where the topic value and APPLICATION_ENDPOINT come from → vendor the upstream `.env` byte-identical, resolve as declared by owner (2026-10-13, "Vendor upstream .env (Recommended)")
- approved: spec v1 by owner (2026-10-13, "Approve both (Recommended)")
- approved: ADR 0013 by owner (2026-10-13, "Approve both (Recommended)")

## Objective

Raise recall and precision of Impact Sets on the published and perturbed hermetic variants by fixing the six extractor causes the non-circular evaluator exposed. Keep determinism and Unsupported Safety at 100%, with no model client (ADR 0001).

In scope:

- spec `gira-impact-precision` v1 and ADR 0013;
- the children listed in the Plan;
- integrated verification of SC1–SC10.

Out of scope:

- ground-truth changes, and `reviewed: true` (the owner's call);
- push or release.

## Acceptance criteria

- AC1 = SC1–SC10 of `specs/gira-impact-precision.md`, at the approved version.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test ./internal/eval -run EndToEnd -v` | pass | fd8be30 | published: recall 58.8%, precision 32.8%; perturbed: 56.2%, 30.5%; safety 100%, determinism 100% |
| B2 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | fd8be30 | full corpus: recall 60.3%, precision 30.4% |
| B3 | `go test ./layer2 -run SC15 -v` inside bench/ | pass | fd8be30 | focused tour assertiveness 5.9% (1/17); full tour 0.0% (0/18) |
| AC1 | SC1–SC10 integrated below | pass | 99bb529 | all SC rows pass on the integrated head; children 01–06 verified @1 |
| SC1 | `go test ./internal/eval -run SC1Impact -count=1` | pass | 99bb529 | no `service:stage:*`, no base-image nodes; briefings hold only Service/Infrastructure; `upstream_libraries` byte-identical CLI/MCP, dropped first by budget |
| SC2 | `go test ./internal/eval -run SC2Dockerfiles -count=1` plus extractor unit tests | pass | 99bb529 | `src/cart/src/Dockerfile` → `service:cart` (published) / `service:basket` (perturbed); no `service:src`; directory rule kept (unit test) |
| SC3 | `go test ./internal/eval -run SC3Kafka -count=1` | pass | 99bb529 | `contract:kafka:orders` and `contract:kafka:order-events` exist; producer extracted, kafka declared, accounting/fraud inferred; tasks 05+06 brief; SC6 12/12 |
| SC4 | `go test ./internal/eval -run SC4GRPC -count=1` plus GRPCConsumption/ClientConsumer/Layer3Mediation | pass | 99bb529 | checkout drops EmailService/ShippingService; keeps Cart, ProductCatalog, Payment, Currency (published); 4 unit cases pass |
| SC5 | `go test ./internal/eval -run SC5HTTP -count=1` plus HTTPCallsAndEnvWiring | pass | 99bb529 | checkout→email `/send_order_confirmation`, →shipping `/get-quote,/ship-order` extracted with `protocol:http,routes`; agent→frontend via APPLICATION_ENDPOINT declared; no unsupported-language source |
| SC6 | `go test ./internal/eval -run SC6Locators -count=1` and `go test ./layer2 -run SC15 -v` in bench/ | pass | 99bb529 | compose edges cite depends_on/env line; ports carry locators; entry steps cite `why_locator`; focused 5.9% → 90.0% (18/20), full 0.0% → 90.0% (18/20) |
| SC7 | `go test ./internal/eval -run EndToEnd -v` | pass | 99bb529 | published 58.8%/32.8% → 82.4%/90.3%; perturbed 56.2%/30.5% → 81.2%/89.7%; each child has before/after rows; chain monotone, no drops, no divergence |
| SC8 | same EndToEnd run (two cold indexes per variant) + `go.mod`/`go.sum` diff + import scan | pass | 99bb529 | safety 100%, determinism 100% both variants at every child and integrated; two cold indexes byte-identical; `go.mod`/`go.sum` unchanged; no LLM SDK, no new network client in indexing path |
| SC9 | `go test -race ./...` at root and `go test -race -count=1 ./...` in bench/ + ground-truth check | pass | 99bb529 | both suites pass; `budget_test` tightening only (skip→error on missing target, retargeted missing-topic fixture); ground-truth.v2.json unchanged, `reviewed:false`; only in-scope recipe edit (`.env` KAFKA_TOPIC) |
| SC10 | `node tools/validate/validate.js` | pass | 99bb529 | no `fail` |
| R1 | `go test -race ./...` at root | pass | 99bb529 | all packages pass (cmd/gira 16s, eval 8s) |
| R2 | `go test -race -count=1 ./...` in bench/ | pass | 99bb529 | layer0, layer1, layer2, quarantine pass; quarantine ADR 0012 respected, no paid runs |
| R3 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` (reported only) | pass | 99bb529 | full corpus: recall 83.8%, precision 89.1%, safety 100%, determinism 100% |
| R4 | `go test ./layer2 -run SC15 -v` in bench/ | pass | 99bb529 | assertiveness 90.0% (18/20) both tours; order-flow 81.8% (accounting, fraud-detection uncovered — tour traversal follow-up, outside spec) |
| V1 | read `git diff fd8be30..99bb529` against spec v1 + fresh-context adversarial review | pass | 99bb529 | diff covers only spec scope + 1-line review fix; prior-art docs (8be32dc) docs-only, no code impact; review finding F1 fixed in 99bb529 |

## Plan

Children are created at `@1` after approval, in dependency order. Each one is test-first and records metrics before and after.

1. `-01-library-and-dockerfile-noise`: cause 1 (SC1).
2. `-02-cross-file-resolution`: the resolution step (ADR 0013), the vendored `.env`, and Dockerfile attribution by compose service, which is cause 6 (SC2). Depends on 01.
3. `-03-kafka-topic-contracts`: cause 2 (SC3), plus the recipe edit for the topic value. Depends on 02.
4. `-04-grpc-called-clients`: cause 3 (SC4). Depends on 03.
5. `-05-http-and-env-wiring`: cause 4 (SC5). Depends on 04.
6. `-06-locator-precision`: cause 5, measured by Layer 2 (SC6). Depends on 05.

SC7–SC10 are checked at every child and integrated on the parent.

## Review

- F1 (minor, rel-ID convention): `renameNode` recomputed a moved edge id with `strings.ReplaceAll(rel.ID, from, to)` where `from`/`to` carry the `service:` prefix but rel ids use bare names (`rel:cart-consumes-email`), so the replace was a no-op and the dedupe-merge could never fire. Latent (Dockerfile nodes emit no rels; no corpus rel references `service:src`; all SC pass) -> fixed in 99bb529 (`moved.ID = relID(moved.SourceID, moved.Verb, moved.TargetID)`; extractors + SC1–SC6 + EndToEnd rerun, metrics unchanged).
- Fresh-context review focus rechecked on 99bb529: no unresolved fact named (refusal paths unit-tested); no fact cites unsupported-language source (SC5 scan); inferred (topic consumers) vs declared (broker, env wiring) vs extracted (producer, calls, HTTP) labels correct; no assertion weakened (budget_test Logf→Errorf tightening for SC3, missing-topic retarget after `contract:kafka:orders` started existing).
- No open findings.

## Outcome

Delivered the six extractor causes plus the integrated head at 99bb529:
- Published hermetic: recall 58.8% → 82.4%, precision 32.8% → 90.3%. Perturbed hermetic: 56.2% → 81.2%, 30.5% → 89.7%. Safety and determinism 100% throughout. Full corpus (opt-in, reported only): 83.8% recall, 89.1% precision. Layer 2 focused/full assertiveness: 5.9%/0.0% → 90.0%/90.0% (18/20).
- Children 01–06 verified @1 (6ee2c85, fa772f7, ee7b029, e9163a7, 5c4aeee, 117e781); chain monotone with per-child before/after rows; integrated SC1–SC10, R1–R4, V1 pass at 99bb529; validator clean.
- Ground truth unchanged (`reviewed:false`); only the spec-authorized perturbation edit (`.env` KAFKA_TOPIC) plus the byte-identical vendored `.env` (blob 3b80f698). No push, no release, no paid runs; bench quarantine (ADR 0012) respected.

Follow-ups outside this spec (owner calls, recorded in children): `contract:Health` unexpected in task 01; `depends_on`-only edges (frontend→quote/image-provider, agent→product-catalog); tour traversal to topic consumers (accounting, fraud-detection; order-flow 81.8%); Layer 2 upper-case env misses (2/20).
