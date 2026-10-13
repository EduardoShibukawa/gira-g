# Quarantined comparative benchmark

GIRA (arm B) against no harness (A), Graphify (C) and Understand Anything (D), under [ADR 0012](../docs/adr/0012-quarantined-comparative-benchmark.md). Specified in `specs/gira-agent-skills-and-eval.md`, part B.

This directory is its own Go module. `go test ./...` at the repository root does not enter it, and nothing outside it may reference it: `quarantine/` checks that statically.

| Path | What |
|---|---|
| `containers/*/Dockerfile` | arm C and D images, built from pinned source commits on base images pinned by digest |
| `phases.json` | build and query phases per arm; every query phase declares `network: none` |
| `scripts/container.sh` | owner-run build and query of one container arm |
| `quarantine/` | the static isolation check (SC12), run with `go test ./...` here |
| `out/` | git-ignored: raw tool output, transcripts, image digests |

Rules that apply to every run:

- Raw outputs stay in `out/`. Only normalized score reports are committed, each naming tool versions, image digests, model IDs and corpus SHAs.
- Paid runs (a frontier model in Layer 1, Understand Anything's LLM indexing) need a separate owner approval with a cost estimate. Mock-model and local Ollama runs need none.
- Results never feed the go/no-go gate.

Run the benchmark's own tests from this directory:

```bash
go test ./...
```
