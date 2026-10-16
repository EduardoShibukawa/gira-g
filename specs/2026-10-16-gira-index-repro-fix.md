---
schema: spec/1
status: approved
version: 1
approved_version: 1
approved_hash: sha256:aad83ab298c401fe0663d06265a12ec56f6798230576aa40c1e59b4a3190f451
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

Bootstrap script + docs over vendoring now (fetch network-blocked here;
owner decided 2026-10-16). Governance note in CLI `--view` help plus
docs site (owner decided 2026-10-16). Repin to other remotes deferred
until network allows; additive, safe. Rejected: vendoring repos in this
session (cannot fetch), new extractors for site sources (H3, out of
scope).

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

None — bundle strategy, governance placement and repin deferral decided
(see parent `decided:` entries, 2026-10-16).

## Revisions

- v1: initial; question-round answers folded pre-approval.
