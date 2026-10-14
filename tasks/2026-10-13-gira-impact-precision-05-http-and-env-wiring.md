---
schema: task/1
tier: standard
status: verified
verified_at: "5c4aeee"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
depends_on: [2026-10-13-gira-impact-precision-04-grpc-called-clients]
---

# HTTP calls and env-wiring edges (cause 4)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Go HTTP requests on env-bound addresses plus a route, and compose env values whose host names a service, become consumes/accesses edges with locators.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC5
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval | pass | 1716d41 | published recall 77.9%, precision 89.8%; perturbed 76.6%, 89.1% (after child 04) |
| AC1 | `go test -race ./internal/eval -run SC5HTTP` and `go test -race ./internal/extractors` (HTTPCallsAndEnvWiring) | pass | 5c4aeee | failed first on 1716d41: checkout to email and shipping cited only the compose header, no metadata; no agent to frontend edge |
| AC2 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | 5c4aeee | published: recall 77.9% → 82.4%, precision 89.8% → 90.3%; perturbed: recall 76.6% → 81.2%, precision 89.1% → 89.7%; token reduction 78.3% → 76.3% and 80.1% → 78.3% (richer briefings; not an SC7 metric) |
| AC3 | same EndToEnd run, two cold indexes per variant (SC8) | pass | 5c4aeee | safety 100%, determinism 100%; go.mod unchanged |
| AC4 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | 5c4aeee | goldens regenerated, diff explained below |
| AC5 | `node tools/validate/validate.js` (SC10) | pass | 5c4aeee | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | 5c4aeee | full corpus: recall 79.4% → 83.8%, precision 88.5% → 89.1% |
| R2 | `go test ./layer2 -run SC15 -v` in bench/ | pass | 5c4aeee | focused tour assertiveness 5.9% (1/17) → 10.0% (2/20), because env-wiring edges cite the env line, which names the target |
| V1 | read `git diff 1716d41..5c4aeee` against SC5 | pass | 5c4aeee | no fact cites unsupported-language source (SC5 test scans every edge) |

## Outcome

Delivered:
- **`Relationship.metadata`.** An additive field (`omitempty`), deep-copied by the slices engine.
- **Go hints.** `http_call` records a `net/http` request (`NewRequest`, `NewRequestWithContext`, `Get`, `Head`, `Post`, `PostForm`) whose URL is an env-bound field or variable plus a literal route.
- **The HTTP join.** The address variable resolves through compose and `.env` to a compose service. The result is `consumes` (or `accesses`), extracted at the request line, with `protocol: http` and sorted `routes`. Checkout reaches email (`/send_order_confirmation`) and shipping (`/get-quote`, `/ship-order`).
- **Env wiring.** A compose `environment` entry whose resolved value addresses another compose service becomes a declared edge at its line. Such edges merge with the `depends_on` edge of the same pair. agent consumes frontend (`APPLICATION_ENDPOINT`), shipping consumes quote (`QUOTE_ADDR`), and agent consumes mcp (`MCP_ENDPOINT`).
- **Interpretation of "host part".** A value counts as an address only with a scheme, a port, or an address-like variable name (`_ADDR`, `_HOST`, `_ENDPOINT`, `_URL`, `_URI`, `_SERVER`). So `OTEL_SERVICE_NAME=<service>` is not wiring. Self-references are skipped.
- **Golden diffs.** The tour reaches `image-provider` before `cart` at one step: stops are ranked by fan-in, and frontend-proxy's env wiring now consumes image-provider. Some inferred rename scores changed, and `impact-checkout`'s full size grew.

Still open, outside this spec: task 02's `quote` and `image-provider` come only from `depends_on` (start ordering), and agent → product-catalog in task 07 does too.
