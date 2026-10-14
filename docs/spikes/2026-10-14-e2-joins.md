# Spike: E2 recall joins — miss → extractor + fixture + locator (read-only, 2026-10-14)

No code edits. Anchors verified by read on the sealed head
(`go test ./internal/eval -run EndToEnd`: published R82.4/P90.3,
perturbed R81.2/P89.7, full R83.8/P89.1; safety/determinism 100%).

## M1 — `contract:Health` unexpected (task-01; also task-07, task-11 inventory)

- Extractor: `internal/extractors/layer3_go.go` —
  `GoASTExtractor.Extract()`, `grpcServerRegisterRegex:^Register(.+)Server$`
  → `ensureContractNode` + `provides` edge. No health denylist.
- Fixture: `fixtures/evaluation/reference-corpus/ground-truth.v2.json`,
  `task-01-checkout-placeorder`, `expected_affected_contracts`
  `[Cart,Checkout,Currency,Payment,ProductCatalog,kafka:orders]` (no Health).
  E2E `unexpected=[contract:Health]`; task-07 adds
  `unexpected=[contract:Health service:agent]`; task-11 adds
  `unexpected=[service:inventory]`.
- Locator anchors: `src/checkout/main.go` (`health.NewServer();
  healthpb.RegisterHealthServer`, Check/Watch methods);
  `src/product-catalog/main.go` (same health registration).
- Hypothesis: stdlib health registration is syntactically identical to a
  business `provides`; denylist stdlib health literals in resolve (J1).

## M2 — `depends_on`-only edges (task-02 frontend→quote/image-provider; task-11 agent→product-catalog)

- Extractor: `internal/extractors/layer2_compose.go` —
  `ComposeExtractor.Extract()`, `depList` from `depends_on` → `VerbConsumes`
  (`rel:svc-consumes-dep`); `composeLines()` / `findServiceLine()`;
  `internal/extractors/resolve.go:resolve()` leaves per-file facts unfiltered.
- Fixture: ground-truth `task-02` upstream excludes quote/image-provider
  (`notes`: "only in frontend's compose depends_on (start ordering)");
  `task-11` upstream `[frontend,mcp]` (`notes`: "depends_on on
  product-catalog is start ordering only").
- Locator anchors: `compose.yaml` (`frontend:depends_on: quote,
  image-provider`); `compose.agent.yaml` (`agent:depends_on:
  product-catalog`; `mcp:depends_on:frontend`).
- Hypothesis: ordering (`condition:service_started/healthy`) conflated with
  runtime `consumes`; require call-site/env/contract corroboration (J2).

## M3 — unsupported-language joins (full corpus: billing/exchange/storefront/opamp-server/ExchangeService)

- Extractor: `layer3_go.go:CanExtract` (`*.go` only); kafka=sarama only,
  grpc=`NewClient` regex only, HTTP=URL-arg only. No JS/TS/C++/C#/Kotlin,
  no `opamp-go`. `resolve_http.go` / `resolve_kafka.go` cannot fire without
  vendored wiring.
- Fixture: `perturbation.json` (`payment→billing`, `currency→exchange`,
  `frontend→storefront`); `ground-truth.v2.json:component_paths` (10 mini
  prefixes only); task-07 `missing=[service:opamp-server]`; task-08
  `missing=[service:storefront]`; task-12
  `missing=[contract:ExchangeService]`. Full-corpus opt-in
  (`fullcorpus_test.go:FetchFullCorpus + Suite.Run published-full`).
- Locator anchors: `src/payment/index.js` (addService),
  `src/currency/src/server.cpp` (HealthServer),
  `frontend/gateways/rpc/*.gateway.ts`, `src/product-catalog/opamp.go` +
  `compose.observability.yaml` (absent from mini), `.env`
  (`OPAMP_SERVER_*`), `compose.agent.yaml` (OPAMP_ENDPOINT/opamp-server).
  Mini `src/` has 10 dirs only (no billing/exchange/storefront/opamp).
- Hypothesis: mini subsets upstream by design; full corpus has facts but no
  extractor rules → posture: join-where-wired else explicit unknown (J3).

## M4 — tour fan-out to topic consumers (accounting, fraud-detection; order-flow 81.8%)

- Engine: `internal/slices/tour.go:GetTour` BFS over
  `consumes/provides/accesses`; `tourStep()` on `Contract` expands to
  providers only, never consumers.
- Fixture: ground-truth `flows:order-flow` (11 steps incl. kafka,
  accounting, fraud-detection); task-01 downstream; task-05/06
  (`precision_test.go:TestSC3KafkaTopicContracts`, consumers inferred at
  `compose.full.yaml KAFKA_TOPIC`); bench
  `bench/layer2/layer2_test.go:TestSC15ProxiesOnGIRATours` (90.0% both
  tours; order-flow noted uncovered in E0 parent R4).
- Locator anchors: `src/checkout/kafka/producer.go:return "orders"` +
  `compose.full.yaml:KAFKA_ADDR`; consumers `src/accounting/Consumer.cs`,
  `src/fraud-detection/.../main.kt` (unsupported lang, inferred).
- Hypothesis: contract treated as provenance-to-providers; fan out to
  `consumes contract:kafka:orders` prunes BFS → enqueue consumers (J4).

## Next (E2 children, owner approval required — not started)

- J1 Health denylist → `SC1HealthDenylist`; J2 corroboration →
  `SC2DependsOnCorroboration`; J3 posture → `SC3UnsupportedPosture`;
  J4 fan-out → `SCTourFanout` + Layer 2 SC15. Spec:
  `specs/gira-recall-joins.md` v1 draft (unapproved).
