# Prior art: autoresearch

[autoresearch](https://github.com/karpathy/autoresearch): MIT. AI agents running
research on single-GPU nanochat training automatically. Studied 2026-10-14 (read
only: README + `program.md`; never cloned, never run — quarantine per ADR 0012).

## How we use it

Studied; port the **contract**, never the code. Directly shaped the review-fix
batch of `gira-evidence-gate` (parent `2026-10-14-gira-evidence-gate`).

## USE / COPY / NORTH STAR

- **USE:** the fixed-harness doctrine (`prepare.py` read-only, `evaluate_bpb` as
  the single ground-truth metric) as the model for our canonical digest
  preimage: one constructor serves CLI eval/verify and MCP verify_artifact.
- **COPY:** keep-or-discard by measured result (never by assertion) → control
  lines must validate internally and come from a real seeded sample + blind
  round; fix-checks must touch the real path (`--expect-content` replaces the
  self-attesting `--repro-ok`); caller-supplied debits are never labeled
  measured (quarantine); taxonomy is part of the digest (experimental
  conditions are logged with the result).
- **NORTH STAR:** the morning `results.tsv` (commit + metric + keep/discard +
  crash, untracked ledger) as the shape of honest automation: our gate ledger
  and findings register follow the same append-only, validated-on-load,
  warn-vs-fail discipline; simplicity criterion (a 0.001 gain with 20 hacky
  lines is not worth it) gates fix scope.

## Study 1 — 2026-10-14 (read-only)

- **Fact:** three files matter: `prepare.py` (fixed, do-not-modify), `train.py`
  (the only file the agent edits), `program.md` (human-edited agent
  instructions = "research org code").
- **Fact:** fixed 5-minute wall-clock budget makes runs comparable across
  changes; metric is val_bpb (lower better, vocab-independent).
- **Fact:** experiment loop: commit → run with output to `run.log` (never tee
  into context) → grep metric → log TSV → keep (advance) on improvement else
  `git reset`; crashes logged as crash; loop never stops to ask.
- **Open:** detection rate of silent-incompleteness fixes vs overhead, same open
  question as our gate work; their answer (fixed budget + single metric) is
  what we copied.
