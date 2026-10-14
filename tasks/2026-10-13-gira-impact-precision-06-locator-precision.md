---
schema: task/1
tier: standard
status: verified
verified_at: "117e781"
parent: 2026-10-13-gira-impact-precision
spec: gira-impact-precision@1
depends_on: [2026-10-13-gira-impact-precision-05-http-and-env-wiring]
---

# Locator precision for compose edges and ports (cause 5)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-13-gira-impact-precision; spec v1 approved)

## Objective

Compose edges cite their depends_on or env line; declared ports get their own locator and tour entry steps cite it; measured by Layer 2 assertiveness.

In scope: the spec criteria listed below, test-first on the hermetic corpus.
Out of scope: the other children's causes; ground-truth changes.

## Acceptance criteria

- AC1 = SC6
- AC2 = SC7
- AC3 = SC8
- AC4 = SC9
- AC5 = SC10

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | `go test -race ./...` at the root and in bench/; EndToEnd eval; Layer 2 | pass | 39b1f5d | published recall 82.4%, precision 90.3%; perturbed 81.2%, 89.7%; Layer 2 focused 10.0% (2/20), full 0.0% (0/20) |
| AC1 | `go test -race ./internal/eval -run SC6Locators` | pass | 117e781 | failed first on 39b1f5d with 151 findings: compose edges at the service header, ports without a locator, entry steps without `why_locator`. Later it also caught that entry steps lost their locator once the index is read back from JSON (16 findings with that fix removed) |
| AC2 | `go test ./layer2 -run SC15 -v` in bench/ (SC6) | pass | 117e781 | focused tour assertiveness 10.0% (2/20) → 90.0% (18/20); full tour 0.0% (0/20) → 90.0% (18/20) |
| AC3 | `go test ./internal/eval -run EndToEnd -v` (SC7) | pass | 117e781 | published: recall 82.4% → 82.4%, precision 90.3% → 90.3%; perturbed: 81.2% → 81.2%, 89.7% → 89.7%; token reduction 76.3% → 78.0% and 78.3% → 79.8% |
| AC4 | same EndToEnd run, two cold indexes per variant (SC8) | pass | 117e781 | safety 100%, determinism 100%; go.mod unchanged |
| AC5 | `go test -race ./...` at the root and `go test -race -count=1 ./...` in bench/ (SC9) | pass | 117e781 | goldens regenerated, diff explained below |
| AC6 | `node tools/validate/validate.js` (SC10) | pass | 117e781 | |
| R1 | `GIRA_EVAL_FULL_CORPUS=1 go test ./internal/eval -run FullCorpus` | pass | 117e781 | full corpus: recall 83.8%, precision 89.1% (unchanged) |
| V1 | read `git diff 39b1f5d..117e781` against SC6 | pass | 117e781 | |

## Outcome

Delivered:
- **Compose `depends_on` edges** cite the entry line (`      cart:` or `- cart`), read from the YAML node tree. The header line is used only when the entry cannot be located.
- **Declared ports.** Dockerfile `EXPOSE` ports become `exposed_ports` and compose `ports` become `published_ports`. Each is a `{port, locator}` entry on its own line. `ports` is unchanged.
- **Tour entry steps.** When the ports declaration has a locator, they use the new why template `entry_declared` ("declares … at the cited locator") and cite it as `why_locator`. `entry` remains for an unknown locator.
- **Impact briefings** list Services without `stages`, `base_images`, `exposed_ports` and `published_ports`. These details stay in slices (contracts and deep tiers), explain and tours. Without this, checkout's larger node crowded the only consumer out of the 1500-token `pr-impact-paymentservice`. It is back, at 1213 tokens, the same size as before this spec. Summary slices strip the same keys.

Golden diffs:
- `tour-frontend` cites `depends_on` lines.
- Briefing goldens lose the build metadata that child 01 had added.
- `explain-checkout`, `locate-currency` and `slice-cart-contracts` gain the port locators.

Remaining Layer 2 misses (2 of 20): env lines such as `- FRONTEND_PROXY_ADDR` name the target in upper case with underscores, and the proxy's anchor is the hyphenated service name. That is a limit of the proxy, not of the locator.
