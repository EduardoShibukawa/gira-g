# Index stores locators, not source; no model client in the MVP

The index and logs hold structured facts, locators (repo, commit, path, line range) and identifiers, never source bodies; excerpts are rendered at query time from a local checkout or reported unavailable. The MVP ships no model client: every Impact Set claim comes from deterministic extraction or declared input, and by default the harness makes no network calls beyond fetching declared repos with the user's own git credentials. We chose this because a deferred CI-built shared index must not carry source-level sensitivity, because a required LLM step would tie the go/no-go to a model and blur what the harness itself contributes, and because "never sends your code anywhere" should hold by construction, verified by a network-isolated acceptance test.

## Considered Options

- **Store code excerpts in the index as evidence.** Rejected: makes the index as sensitive as every repo it covers.
- **Ship an optional LLM step (e.g. Component summaries) in the MVP.** Rejected: no measured value yet; any future LLM step must be off by default, use an explicitly configured vendor-neutral endpoint, be opted into per Target System, and record each request in a local egress log.
- **Restrict what agents receive (a "remote-safe" slice mode).** Rejected: the harness cannot know which model sits behind the agent, so it would weaken Impact Sets without enforcing anything. The trust boundary covers harness-initiated traffic only.

Source: [Trust boundary: what may leave the machine in the MVP?](../../../../.scratch/system-context-harness/issues/02-trust-boundary.md)
