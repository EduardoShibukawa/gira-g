# Gira-G Repository — Autonomous Agent Governance

All autonomous agents working in `gira-g` must adhere to repository guidelines and P-Ingá verification standards.

## Core Rules

1. **Verification & Gates:**
   - Commit and push through mise: `mise exec -- git commit ...`.
   - Never use `--no-verify`.
   - Run tests before committing: `go test ./...`.
2. **P-Ingá Formal Contracts:**
   - For structured tasks, author a `tasks/T<N>/PLAN.md` (<= 25 lines) and verify with `pinga verify tasks/T<N>/PLAN.md`.
   - Use `/pinga` and `/red-audit` skills for invariant compliance and adversarial reviews.
3. **Surgical Modifications:**
   - Use `gira outline` and `gira read` to isolate symbols before modifying Go code.
