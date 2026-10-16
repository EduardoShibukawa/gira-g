---
schema: task/1
tier: standard
status: verified
verified_at: "b29904b"
parent: 2026-10-15-gira-docs-site
spec: 2026-10-15-gira-docs-site@1
depends_on: [2026-10-15-gira-docs-site-01-scaffold]
---

# 02 — Content: Concepts (F2; SC2)

## Routing and approvals

- tier: standard (decomposed from heavy parent 2026-10-15-gira-docs-site; spec v1 approved)

## Objective

Migrate the vocabulary from `CONTEXT.md` "Language" section into
`website/docs/concepts/`: one page per term group (Target System, Component,
Service, Library, Infrastructure, Contract, Evidence Record, Macro Map,
Extractor, Evaluation corpora, Change impact, Change-maker, Impact Briefing,
Impact Set).

In scope: concept pages derived from CONTEXT.md terms.
Out of scope: CLI reference (03), guides (04), ADRs/decisions (out of spec).

## Acceptance criteria

- AC1 = SC2 (concepts): every term in `CONTEXT.md` "Language" section has a
  corresponding page in `website/docs/concepts/`.
- AC2: each concept page has a definition, _Avoid_ list (where present in
  CONTEXT.md), and cross-references to related concepts.

## Verification

| ID | Check | Result | Ref | Note |
|----|-------|--------|-----|------|
| B1 | node tools/validate/validate.js (baseline) | pass | 3b01049 | repo clean before child |
| AC1 | grep CONTEXT.md terms vs website/docs/concepts/ | pass | b29904b | 14 pages cover every term in CONTEXT.md Language |
| AC2 | content review: definitions, avoid lists, xrefs | pass | b29904b | each page has definition, avoid list, related xrefs |
| R1 | go test -race ./... (no Go change) | pass | b29904b | no Go touched |
| V1 | diff read against AC1–AC2 | pass | b29904b | only concept pages added |

## Plan

1. Read `CONTEXT.md` "Language" section; enumerate all terms.
2. Create `website/docs/concepts/` with one .mdx per term group.
3. Run AC1–AC2/R1/V1; commit.

## Outcome

Added 14 concept pages in `website/docs/concepts/` covering every term in
`CONTEXT.md` "Language" (Target System, Component, Service, Library,
Infrastructure, Contract, Evidence Record, Macro Map, Extractor, Reference
Corpus, Perturbed Variant, Change-maker, Impact Briefing, Impact Set). Each
page has the definition, avoid list, and cross-references. Build passes;
AC1/AC2/R1/V1 all pass at b29904b.
