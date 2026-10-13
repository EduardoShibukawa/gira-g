# Layer 2 moderated study protocol

Human comprehension evaluator, periodic and manual (spec `gira-agent-skills-and-eval`, part B, Layer 2). It complements the automated proxies in this package and never runs in CI. Participant recruiting is deferred (spec Open questions); this document fixes everything else so a session can be run as soon as participants exist.

## Goal

Compare how quickly and how correctly a newcomer understands an unfamiliar system with each tool's human surface:

- **T1, GIRA:** `gira tour`, `gira explain`, `gira flows`, rendered as Markdown;
- **T2, Understand Anything:** its dashboard tour and explanations, built in quarantine (its LLM indexing is a paid run that needs owner approval);
- **T3, no tool:** the source tree and a code search.

## Design

- Within-subject: every participant uses all three tools, each on a **different corpus area**, so that nobody learns an area with one tool and re-uses that knowledge with another.
- Corpus: the OpenTelemetry Demo at the pinned commit in `phases.json`, with `README*`, `AGENTS.md` and `docs/` removed (the ADR 0007 memory-leak rule).
- Areas, chosen to be comparable in size and in number of hops:
  - **Area 1, ordering:** frontend → checkout → cart, payment, shipping, email, kafka;
  - **Area 2, catalog and recommendations:** frontend → product-catalog, recommendation, ad, currency;
  - **Area 3, async processing:** checkout → kafka → accounting, fraud-detection.
- Sessions are moderated, one participant at a time, 75 minutes at most.

## Counterbalancing

A 3×3 Latin square assigns tool × area × order. Participants are assigned to rows in rotation (participant *k* gets row *k* mod 3), so every tool meets every area and every position equally often per block of 3 participants.

| Row | First | Second | Third |
|---|---|---|---|
| R1 | T1 on Area 1 | T2 on Area 2 | T3 on Area 3 |
| R2 | T2 on Area 3 | T3 on Area 1 | T1 on Area 2 |
| R3 | T3 on Area 2 | T1 on Area 3 | T2 on Area 1 |

## Tasks

Each block asks the participant to answer two questions about its area, in this order, aloud and in writing:

1. A **flow** question. Area 1: "how does an order flow from the cart to shipping?" Area 2: "how does the storefront decide which products and ads to show?" Area 3: "what happens to an order after it is placed, asynchronously?"
2. An **impact** question. Area 1: "who must be notified if PaymentService changes?" Area 2: "who must be notified if ProductCatalogService changes?" Area 3: "who must be notified if the orders topic schema changes?"

The ground truth for every question is fixed in advance from `fixtures/evaluation/reference-corpus/ground-truth.v2.json` (the `order-flow` flow and the task Impact Sets), extended for areas 2 and 3 by the same evidence rule, and reviewed by the owner before the first session.

## Rubric

Answers are graded blind: the grader sees the answer text without the tool or the participant.

| Score | Flow question | Impact question |
|---|---|---|
| 0 | wrong or no answer | no correct Component named |
| 1 | at most half of the ground-truth flow steps, in any order | at most half of the expected Impact Set |
| 2 | more than half of the steps, mostly in order | more than half of the Impact Set, at most one wrong Component |
| 3 | every step in order, no invented step | the whole Impact Set, no wrong Component |

A Component or step that is not in the corpus counts as invented and caps the score at 1.

## Instruments

- **Time to first useful understanding:** from the moment the question is read to the first answer graded at least 2, measured by the moderator's stopwatch. Participants may revise; each revision is timestamped.
- **Correctness:** the rubric score of the final answer.
- **Self-reported clarity:** after each block, "How clear was the system to you with this tool?" on a 5-point scale (1 very unclear, 5 very clear).
- **Workload:** raw NASA-TLX after each block: mental demand, physical demand, temporal demand, performance, effort and frustration, each 0–100 in steps of 5, averaged without weighting.

## Procedure

1. Consent and a 5-minute warm-up on a toy repository with each tool (not the study corpus).
2. Three blocks in the participant's Latin-square row. Each block is 20 minutes at most: two questions, then the clarity item and NASA-TLX.
3. Debrief: a short open question on what helped and what misled.

## Data handling

- Recorded per block: participant code, row, tool, area, timestamps, the written answers, the rubric scores, the clarity rating and the NASA-TLX sub-scores.
- No names, voices or screen recordings are stored. Raw sheets stay in the git-ignored `out/` directory; only an aggregated report is committed, naming tool versions, image digests and the corpus SHA.

## Analysis

- Report per tool: the median and interquartile range of time to first useful understanding, the rubric score distribution, the clarity median and the raw NASA-TLX mean.
- Report the participant count next to every number. Small samples are reported as such: no significance tests, no p-values and no claims of superiority. Reports show per-participant values alongside the summaries.
- Put the automated proxies (assertiveness, coverage, steps, words per step) for the same tours in the same report, so the study can show whether they predict human outcomes.
