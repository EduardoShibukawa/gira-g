---
schema: spec/1
status: draft
version: 1
---

# Gira index reproducibility fix

## Problem

`bin/gira index` against `gira.yaml` dogfood pins fails in this
environment both offline (commit `ec0f0c4` absent from cache) and online
(`duh-homelab/gira-dogfood-services` unreachable, exit 128), diagnosed
in `tasks/2026-10-16-gira-index-reproducibility-diagnosis.md`. Fixing it
touches manifest docs, cache bootstrap/CI and user docs (H3), and the
vendor fetch cannot run where there is no network, so it likely spans
sessions (H5).

## Expected behavior

`bin/gira index` succeeds offline with a documented bundle and online
where network allows, deterministically; website/docs sources are
declared unsupported by design instead of extractor noise.

## Scope

In: offline-bundle docs/bootstrap, unsupported-by-design declaration,
governance focal-node note. Out: new extractors, manifest repinning to
other remotes, vendored repo contents in this session.

## Acceptance criteria

- SC1: offline index path documented and reproducible where bundle exists.
- SC2: unsupported posture declared; locate noise rationale recorded.
- SC3: governance empty-neighborhood behavior documented.

## Interfaces and dependencies

`gira.yaml` pins (consumers: `gira index`, evaluator, CI); gitcache
offline mode; docs site. No contract change to CLI envelopes.

## Decisions

Bundle-docs over vendoring now (network blocked here); repin rejected
without owner call. Details after question round.

## Risks and constraints

Network-blocked fetch; no secrets/auth changes; bundle size unknown
until fetchable. Protected actions: none planned.

## Verification

SC1–SC3 end to end on integrated head; `go test -race ./...`,
`golangci-lint --new-from-rev=origin/main`, validator; review focused
on manifest/cache instructions accuracy.

## Rollback

skip: docs-only plus additive bootstrap; revert the commits.

## Open questions

- Bundle format/location (cache dir tarball vs checked-in bootstrap script)?
- Governance note placement (CLI help vs docs site)?
- Defer repin decision until network allows (safe: additive)?

## Revisions

- v1: initial.
