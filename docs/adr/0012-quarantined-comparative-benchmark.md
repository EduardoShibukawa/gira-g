# Quarantined comparative benchmark against prior art

**Status:** accepted (2026-10-13, by owner). This ADR amends [ADR 0007](0007-evaluation-design-and-thresholds.md) §1, "Prior Art Isolation", and the matching rejected option. Specified in `specs/gira-agent-skills-and-eval.md`.

GIRA's product claim is comparative: agent efficiency like Graphify, human clarity like Understand Anything, and determinism on top. ADR 0007 bans both tools from every evaluation pipeline, so the claim can't be tested. We keep the ban where it protects GIRA, and lift it inside a quarantine where it does not:

- **The go/no-go gate stays GIRA-only.** It stays deterministic and offline, and it runs in CI, exactly as ADR 0007 says.
- **The comparison is a separate benchmark.** It runs Graphify (pinned version) and Understand Anything (pinned commit) next to GIRA and a no-harness baseline, on the same corpus: the OpenTelemetry Demo, as published and as the Perturbed Variant.
- **The benchmark runs only in quarantine.** Its guards are listed below.

## Quarantine rules

1. **Containers.** Each external tool runs in a disposable container. The container is built from pinned source and referenced by image digest, never by a floating tag. It runs without credentials and mounts only a copy of the corpus.
2. **Two phases.**
   - In the build phase, the tool indexes the corpus.
   - In the query phase, networking is disabled, except for the model endpoint of the arm under test.
3. **No coupling.** No benchmark asset is referenced from the Go module, `mise.toml`, CI workflows, git hooks, `AGENTS.md` or any agent configuration in this repository. A static check enforces this.
4. **Outputs.**
   - Raw tool output and transcripts stay in a git-ignored directory.
   - Only normalized score reports are committed.
   - Each committed report records tool versions, image digests, model IDs and corpus SHAs.
   
   This keeps the prior-art policy: their output is never committed.
5. **Paid runs.** Any run that spends money, such as a frontier API call or Understand Anything's LLM indexing pass, is a protected action. It needs a separate owner approval with a cost estimate.
6. **No influence on the gate.** Benchmark results never feed the ADR 0007 go/no-go gate.

## Considered Options

- **Keep the ban (ADR 0007 as written).** Rejected: the comparative claim would stay untested and become marketing, not evidence.
- **Add the tools to CI.** Rejected for the original reasons in ADR 0007. They are non-deterministic, they have LLM dependencies, and running them would add network and supply-chain exposure.
- **Edit ADR 0007 in place.** Rejected: an amending ADR keeps the original decision and its reasoning auditable.

## Consequences

- We can publish side-by-side numbers: the four arms A (no harness), B (GIRA), C (Graphify) and D (Understand Anything), at three layers:
  - structural graph;
  - agent tasks: tokens, tool calls, latency and accuracy;
  - human comprehension.
  
  Each report states the versions behind its numbers.
- The Graphify and Understand Anything prior-art docs gain dated benchmark sections, which link to the committed reports.
- Maintenance cost: the tool images and adapters must be bumped by hand when a new prior-art version is studied.
