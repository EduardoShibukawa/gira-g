# Prior art: PostHog

[PostHog](https://github.com/posthog/posthog): MIT (mostly) + proprietary `ee/`. Django+React+Node+Rust+Go monorepo (~67k commits, 40.3k stars, 2026-10-14, SHA to pin).

## How we use it

Studied, never depended on (ADR 0012). By hand only, outside the toolchain; its output is never committed.

## USE / COPY / NORTH STAR

- **USE:** as a **hard extraction fixture** (manifest/routes/OpenAPI/compose/topics/tach) in quarantine; not as a dependency.
- **COPY:** facade + frozen dataclasses as boundary; `manifest.tsx`→`products.json` + `register_routes` for declared Libraries and route extractor; `mcp/tools.yaml` + OpenAPI→Orval→MCP as codegen template; CLI `exec` + `features/tools/readonly` filters for budgets; `db_routing.yaml`/`topics.txt`/Caddy routing as Infra fixtures; `owners.yaml`/crossings baselines for evidence + staleness.
- **NORTH STAR:** Product≈Component, facade≈Contract; LLM-everywhere as counter-example (reinforces ADR 0001, Evidence-or-silence).

## Study 1 — 2026-10-14 (subagent)

- **Fact:** `products/<p>/{backend/frontend/manifest.tsx, mcp/tools.yaml}` + `backend/facade/{api.py,contracts.py}`; cross-product contact only via facade, enforced by `tach` + `import-linter` + `hogli product:lint` (`docs/internal/monorepo-layout.md`, `products/architecture.md`, `AGENTS.md`).
- **Fact:** Infra = per-product PG (`products/db_routing.yaml`), ClickHouse, Kafka (`docker/kafka/topics.txt`), Redis/Valkey, Temporal (2 MiB cap → pass-by-ref), SeaweedFS, etcd, OTel; Rust `capture/feature-flags/personhog` (gRPC `:50052`) behind Caddy (`docker-compose.dev.yml`).
- **Fact:** hosted MCP `mcp.posthog.com/mcp` with 1100+ OpenAPI-generated tools + `tools.yaml`; `?readonly/features/tools`, CLI mode, MCP Analytics (`$mcp_tool_call`); skills in `.agents/skills/`; evals in `products/posthog_ai/evals/`.
- **Inference:** closest at-scale precedent for MCP+skills+budgets+eval, but inverted (serves product-data, not arch-map).
- **Open:** exact HEAD SHA; AsyncAPI/proto coverage; vestigial `proto/`?; Helm still present?; hard budget numbers; eval-harness isolation.
