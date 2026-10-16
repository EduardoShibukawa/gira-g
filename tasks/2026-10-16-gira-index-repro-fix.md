---
schema: task/1
tier: heavy
status: active
spec: 2026-10-16-gira-index-repro-fix@1
---

# Gira index reproducibility fix

## Routing and approvals

- tier: heavy (H3: S3 fails, manifest+cache+docs span subsystems; H5: S8 fails, fetch network-blocked, multi-session)
- confirmed: heavy by owner (2026-10-16, "Confirmar heavy")
- decided: bundle strategy → bootstrap script + docs, no vendored binaries, by owner (2026-10-16, "Bootstrap script")
- decided: governance note → CLI `--view` help plus docs site, by owner (2026-10-16, "Help + docs")
- approved: spec v1 by owner (2026-10-16, question-round answers folded pre-approval)

## Objective

Make `bin/gira index` reproducible per `specs/2026-10-16-gira-index-repro-fix.md`.

In scope: spec approval, children, integrated verification.
Out of scope: new extractors, remote repinning without owner call.

## Acceptance criteria

SC1–SC3 from the spec.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | `go test -race ./...` before any child started | pass | a82e44b | exit 0, all pkgs ok (cached, tree clean at proposal commit) |
| SC1 | offline index path reproducible where bundle exists | not run | | |
| SC2 | unsupported posture declared | not run | | |
| SC3 | governance behavior documented | not run | | |
| R1 | full suite + lint on integrated head | not run | | |
| V1 | integrated diff read against SC1–SC3 | not run | | |

## Plan

1. 2026-10-16-gira-index-repro-fix-01-offline-bundle-bootstrap: bootstrap script + index.md (todo)
2. 2026-10-16-gira-index-repro-fix-02-unsupported-declaration: extractor.md (todo)
3. 2026-10-16-gira-index-repro-fix-03-governance-note: slice help + slice.md (todo)
4. 2026-10-16-gira-index-repro-fix-04-integrated: SC1–SC3 + R1 (todo, depends on 01–03)

## Review

skip: runs after children are verified.
