---
schema: spec/1
status: draft
version: 1
---

# GIRA recall joins (E2)

## Problem

E1 sealed the proof (parent `2026-10-14-gira-evidence-gate`, verified):
every `gira` claim is now re-derivable or silent. The numbers underneath did
not move, by design (E1 non-goal):

| Variant | Recall | Precision | Safety | Determinism |
|---|---|---|---|---|
| Published, hermetic | 82.4% | 90.3% | 100% | 100% |
| Perturbed, hermetic | 81.2% | 89.7% | 100% | 100% |
| Published, full corpus (opt-in) | 83.8% | 89.1% | 100% | 100% |

The per-task `missing` / `unexpected` report at the sealed head points to
four join causes (measured, not guessed). All anchors below were mapped
read-only on 2026-10-14 (see `spike-e2-joins.md`, untracked):

1. **`contract:Health` noise (precision).** `task-01-checkout-placeorder`
   lists `unexpected=[contract:Health]`; `task-07` adds
   `unexpected=[contract:Health service:agent]`; `task-11` adds
   `unexpected=[service:inventory]`. Source: `GoASTExtractor` turns every
   `Register*Server` call into a `provides` edge
   (`internal/extractors/layer3_go.go`, `grpcServerRegisterRegex`), with no
   `health/grpc_health` denylist. Anchors:
   `src/checkout/main.go#L271-L272` (`health.NewServer();
   healthpb.RegisterHealthServer`), `src/product-catalog/main.go#L207-L208`.
   Ground truth (`ground-truth.v2.json:task-01`) expects only
   `[Cart,Checkout,Currency,Payment,ProductCatalog,kafka:orders]`.
2. **`depends_on`-only edges (precision).** `task-02` excludes
   `quote/image-provider` (`notes`: "only in frontend's compose depends_on
   (start ordering)"); `task-11` upstream `[frontend,mcp]` excludes the
   `agent→product-catalog` edge (`notes`: "depends_on on product-catalog is
   start ordering only"). Source: `ComposeExtractor` emits
   `VerbConsumes` for every `depends_on` entry
   (`internal/extractors/layer2_compose.go:depList`), never requiring
   call-site or env corroboration. Anchors: `compose.yaml#L295-L315`
   (`frontend:depends_on: quote#L306, image-provider#L314`),
   `compose.agent.yaml#L53-L57` (`agent:depends_on: product-catalog#L56`).
3. **Unsupported-language joins (recall, full corpus).** Full corpus still
   misses `billing` (payment→billing, `src/payment/index.js#L42`),
   `exchange` (currency→exchange, `src/currency/src/server.cpp#L129`),
   `storefront` (`task-08 missing=[service:storefront]`,
   `frontend/gateways/rpc/*.gateway.ts`), `opamp-server`
   (`task-07 missing=[service:opamp-server]`,
   `src/product-catalog/opamp.go` + `compose.observability.yaml`, both
   absent from the hermetic mini), `contract:ExchangeService`
   (`task-12 missing=[contract:ExchangeService]`). Source:
   `Layer3GoAST.CanExtract` covers `*.go` only; kafka/grpc/HTTP matchers
   cover sarama/`grpc.NewClient`/URL args only — no JS/TS/C++/C#/Kotlin,
   no `opamp-go`. The hermetic mini intentionally subsets upstream (10
   `src/` dirs; no `observability/agent` compose), so these misses persist
   full-corpus by extractor coverage, not by fixture.
4. **Tour fan-out to topic consumers (recall, traversal).** `order-flow`
   81.8%: `accounting` + `fraud-detection` uncovered. Source:
   `slices/tour.go:GetTour` expands a `Contract` node to its providers
   only, never to its consumers, so
   `frontend→checkout→kafka:orders` never enqueues the two inferred
   consumers (`resolve_kafka.go`, `compose.full.yaml KAFKA_TOPIC`).
   Anchors: `src/checkout/kafka/producer.go:return "orders"`,
   `src/accounting/Consumer.cs`,
   `src/fraud-detection/src/main/kotlin/frauddetection/main.kt`.

Heavy triggers:

- **H1:** product calls on Health denylist scope, `depends_on` corroboration
  rule, unsupported-language posture (declare-and-skip vs extract), tour
  fan-out direction.
- **H3:** spans Go/Compose extractors, the resolve step, slices/tours,
  eval corpus + bench Layer 2.
- **H5:** four causes, each test-first, decomposed into children.

## Expected behavior

One epic, four additive join fixes. No extractor rewrites; joins only.
Anything no Extractor covers stays reported as unsupported, never guessed
(CONTEXT.md: Extractor).

### J1 — Health denylist (precision)

`health` / `grpc_health` registrations never become `provides` edges.
`checkout provides contract:Health` disappears from depth-1 briefings;
`task-01` and `task-07` `unexpected` lose `contract:Health` with no recall
loss. Denylist is a fixed literal set (stdlib health only), not a pattern
that could swallow business contracts.

- Ground-truth change: none (`reviewed` stays false).

### J2 — `depends_on` corroboration (precision)

A `depends_on`-only edge is emitted only with corroboration: a call-site,
an env wiring, or a contract reference between the same pair. Bare ordering
edges (`frontend→quote`, `frontend→image-provider`,
`agent→product-catalog`) disappear; corroborated edges keep their locator,
now citing both the `depends_on` line and the corroborating line.

- `task-02` and `task-11` lose the ordering-only `unexpected`/upstream
  entries with no recall loss.

### J3 — unsupported-language posture (recall honesty)

No new language extractors in E2. Instead: each unsupported call-site gets
a declared-or-inferred join with an explicit `unsupported` locator, or an
explicit unknown — never a named guess. `billing` / `exchange` /
`storefront` / `opamp-server` / `ExchangeService` move from silent-missing
to either joined (where compose/env wiring suffices) or explicitly
`unknown` with a locator to the unsupported file. Full-corpus recall may
rise only where wiring alone proves the edge; the headline is honesty, not
a number.

### J4 — tour fan-out to consumers (recall)

`GetTour` enqueues consumers of a visited `contract:kafka:T` alongside
providers. `order-flow` covers `accounting` + `fraud-detection`;
focused/full Layer 2 assertiveness does not regress from 90.0% (18/20).

### Invariants (all joins)

- **I1 — Evidence or silence.** No claim without a locator-backed
  `EvidenceRecord`; unresolved → `unknown`, never named.
- **I2 — Tier honesty.** `declared/extracted/inferred` never mislabeled;
  inferred consumer edges stay `inferred`.

## Scope

**In scope**

- J1–J4 above, each test-first, all additive (no flag/field meaning
  changes; budgets/truncation untouched).
- Deterministic join rules in the resolve step and tour engine (not model
  calls).

**Out of scope**

- E3 envelopes/hygiene, E4 human surface, E5 skills distribution, E6 docs
  site.
- New language extractors (E2 declares posture only; extractors are later
  work).
- Ground-truth changes, `reviewed: true`, push/release.
- Any model client (ADR 0001), any paid multi-model fan-out.

**Non-goals**

- A recall/precision target. E2 raises honesty first; numbers are reported,
  never chased by weakening assertions.

## Acceptance criteria

- **SC1 (Health).** `task-01` and `task-07` report no `contract:Health`
  in `unexpected`; hermetic recall/precision do not drop. Test:
  `go test ./internal/eval -run SC1HealthDenylist`.
- **SC2 (depends_on).** `task-02` and `task-11` report no ordering-only
  edges; corroborated edges cite both lines. Test:
  `go test ./internal/eval -run SC2DependsOnCorroboration`.
- **SC3 (unsupported posture).** Every `billing/exchange/storefront/
  opamp-server/ExchangeService` miss is either joined with a locator or
  explicitly `unknown` with an `unsupported` locator; zero silent-missing.
  Test: `go test ./internal/eval -run SC3UnsupportedPosture`.
- **SC4 (tour fan-out).** `order-flow` covers `accounting` +
  `fraud-detection`; bench Layer 2 assertiveness stays ≥ 90.0%.
  Test: `go test ./internal/slices -run SCTourFanout` plus
  `go test ./layer2 -run SC15 -v` in `bench/`.
- **SC5 (invariants + regression).** I1/I2 hold;
  `go test -race ./...` passes at root and in `bench/`; validator reports
  no `fail`; determinism 100%, safety 100%; hermetic recall/precision not
  lowered (any dip needs a `divergence:` owner entry).

## Interfaces and dependencies

All additive. No existing flag, field or tool changes meaning.

| Interface | Change |
|---|---|
| Resolve step | Health denylist, `depends_on` corroboration rule (new deterministic rules) |
| Slices/tours | consumer fan-out on visited topic contracts |
| Eval corpus | new SC1–SC4 tests; `spike-e2-joins.md` anchors (untracked, never committed) |
| CLI/MCP | none required; if new flags appear they are additive `--dry-*` style |

**ADRs.** ADR 0001 stands (no model client). ADR 0012 stands (quarantine;
prior-art repos run by hand in throwaway envs only, output never
committed).

## Decisions

- None yet — draft. Owner question rounds will cover: Health denylist
  scope (stdlib-only vs configurable); corroboration minimum for
  `depends_on`; unsupported posture (join-where-wired vs always-unknown);
  tour fan-out breadth (one hop vs transitive).

## Risks and constraints

- **Over-denylisting.** A literal health set cannot swallow business
  contracts; any broader pattern needs owner approval.
- **Corroboration strictness.** Too strict drops real edges (recall);
  too loose keeps ordering noise (precision). Each rule change ships with
  before/after per-task rows.
- **Unsupported honesty vs recall.** Explicit `unknown` may read as a
  miss; that is the honest state and is reported as such.
- **Compatibility.** All changes additive; `git revert` per child
  suffices.
- **Honest No-Go.** A stricter join may flip sealed runs. Expected and
  acceptable: fix the join, not the gate.

## Verification

- Per child + integrated: B1 `go test -race ./...` (root); B2
  `go test -race ./...` (bench/); B3
  `go test ./internal/eval -run EndToEnd -v` (both hermetic variants,
  report recall/precision/safety/determinism); B4 opt-in full-corpus run
  (reported only); B5 `node tools/validate/validate.js`.
- AC rows map SC1–SC5 to the `go test` names above; `not run` forbidden
  at parent seal.
- **Review focus:** does the Health denylist touch only stdlib health? Is
  `depends_on` corroboration real (two locators)? Is every unsupported
  miss either joined or explicitly unknown (no silent-missing)? Does tour
  fan-out stop at consumers (no transitive explosion)? Any weakened
  assertions or named guesses?

## Rollback

`git revert` per child. No migration, publication or external state.

## Open questions

- deferred: **Health denylist home** (extractor vs resolve step). Safe:
  either is additive with the same tests.
- deferred: **unsupported join depth** (wiring-only vs manifest hints).
  Safe: posture first, depth later.
- deferred: **tour fan-out breadth**. Safe: one hop closes order-flow;
  transitivity is a separate call.

## Revisions

- v1 (2026-10-14, overnight draft, unapproved): first draft from sealed-head measured misses.
