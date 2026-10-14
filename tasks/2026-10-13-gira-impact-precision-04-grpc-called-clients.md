---
schema: task/1
tier: standard
status: verified
verified_at: "e9163a7"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
depends_on: [2026-10-13-gira-impact-precision-03-kafka-topic-contracts]
---

# gRPC consumption only on method calls (cause 3)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

A service consumes a gRPC Contract only when its Go package invokes a method on a client bound to it.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC4
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval | pass | 711b70d | published recall 77.9%, precision 86.9%; perturbed 76.6%, 86.0% (after child 03) |
| AC1 | `go test -race ./internal/eval -run SC4GRPC` and `go test -race ./internal/extractors` (GRPCConsumption, ClientConsumer, Layer3Mediation) | pass | e9163a7 | failed first on 711b70d: checkout consumed EmailService and ShippingService (ordering: NotifyService, DispatchService) from field and constructor lines |
| AC2 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | e9163a7 | published: recall 77.9% → 77.9%, precision 86.9% → 89.8%; perturbed: recall 76.6% → 76.6%, precision 86.0% → 89.1% |
| AC3 | same EndToEnd run, two cold indexes per variant (SC8) | pass | e9163a7 | safety 100%, determinism 100%; go.mod unchanged |
| AC4 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | e9163a7 | the old SC1 mutation test (removing Go AST lowers contract recall) still passes |
| AC5 | `node tools/validate/validate.js` (SC10) | pass | e9163a7 | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | e9163a7 | full corpus: recall 79.4% → 79.4%, precision 83.1% → 88.5% |
| R2 | `go test ./layer2 -run SC15 -v` in bench/ | pass | e9163a7 | unchanged: 5.9% (1/17) |
| V1 | read `git diff 711b70d..e9163a7` against SC4 | pass | e9163a7 | |

## Outcome

Delivered:
- **Per-file extraction.** A `New<X>Client` constructor or a `pkg.<X>Client` field now only creates the Contract node.
- **Go hints.**
  - `grpc_client`: a field or variable bound to a client.
  - `go_alias`: a variable assigned from a field or variable.
  - `go_method_call`: an exported method called on a field or variable.
- **The resolution step.** It emits `service consumes contract:<X>` for each call on a bound client, following aliases, anywhere in the service's Go files. The evidence is extracted, at the call line, with profile `grpc`.
- **Unit tests.** Cases: a field called from another file, a local alias of the field, a variable assigned `New<X>Client`, and two clients never called. The end-to-end mediation test's checkout now calls `Charge`, since constructing a client no longer counts.
- **Golden diffs.** The PaymentService consumption cites the `paymentService.Charge` call instead of the field and constructor lines. Checkout's briefing loses EmailService and ShippingService. The checkout → ordering rename score changed (3 → 3.029).

Not changed: `contract:Health`, from the gRPC health server registration, remains the one unexpected id in task 01 (a follow-up outside this spec).
