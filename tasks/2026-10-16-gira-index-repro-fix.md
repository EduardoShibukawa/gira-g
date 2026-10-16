---
schema: task/1
tier: heavy
status: verified
verified_at: "46821c1"
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
| B1 | `go test -race ./...` before any child started | pass | a82e44b | exit 0, pre-children baseline |
| SC1 | offline index path reproducible where bundle exists | pass | 46821c1 | bootstrap `--help` ok; index.md section live |
| SC2 | unsupported posture declared | pass | 46821c1 | extractor.md section live |
| SC3 | governance behavior documented | pass | 46821c1 | slice help + slice.md:42 live |
| R1 | integrated regression (suite + lint + validator) | pass | 46821c1 | suite 0; lint 0 new; validator pass |
| V1 | integrated diff read against SC1–SC3 | pass | 46821c1 | +124/−16 vs eec886f; help-only CLI change, no envelope break |

## Plan

1. 2026-10-16-gira-index-repro-fix-01-offline-bundle-bootstrap: bootstrap script + index.md (verified 59b8a5f)
2. 2026-10-16-gira-index-repro-fix-02-unsupported-declaration: extractor.md (verified 5b5deab)
3. 2026-10-16-gira-index-repro-fix-03-governance-note: slice help + slice.md (verified 267388d)
4. 2026-10-16-gira-index-repro-fix-04-integrated: SC1–SC3 + R1 (verified 26a28e5)

## Review

Fresh-context review of integrated head against spec v1 (aad83ab2):

- F1: does `--view` help change alter any CLI contract? No —
  description string only, no flag added/renamed; suite + lint green.
  Resolution: no action.
- F2: does the bootstrap script promise fetching it cannot do here?
  No — step 1 states "needs network once"; offline path verified by
  flag semantics. Resolution: no action.
- F3: any child diverging from spec @1? No — all four reference
  `@1`, no `divergence:` entries. Resolution: no action.

No open findings.

## Outcome

SC1–SC3 delivered: bootstrap script, unsupported declaration,
governance note. Awaiting owner `accepted:`; deploy/publish needs
no protected action beyond the normal push (done per commit).
