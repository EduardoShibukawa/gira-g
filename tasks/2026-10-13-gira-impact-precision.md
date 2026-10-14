---
schema: task/1
tier: heavy
status: active
spec: gira-impact-precision@1
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
| AC1 | SC1–SC10 rows | not run | | |
| V1 | read the integrated diff against spec v1, plus the fresh-context review | not run | | |

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

skip: runs after the children are integrated (fresh-context adversarial review, `.claude/skills/work/review.md`).
