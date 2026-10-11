# Human-facing surface and evidence inspection model

The human surface of Gira-G (`gira`) in the MVP is designed specifically for the primary human persona: the **Change-maker** (Issue 01). It provides instant, evidence-backed answers through a rich Go terminal CLI and self-contained static exports (Markdown and single-file HTML), without requiring any persistent web server daemon, Node.js runtime, or external services.

## Decision Summary

1. **Dual Human Delivery Surface (CLI + Self-Contained Static Export):**
   - **Rich Terminal CLI:** Pure Go implementation rendering styled tables, tree hierarchies, and token-budgeted Markdown directly to the terminal using `charmbracelet/lipgloss`. Automatically detects non-TTY environments (pipes, CI) or the `NO_COLOR` standard to fall back to plain, clean text.
   - **Self-Contained Static Export (`gira export`, `gira view`):** Generates standalone, single-file Markdown documents or zero-dependency static HTML/SVG files (`--format=html|markdown`). `gira view` generates an artifact into a temporary or configured directory and opens the local default browser (`file:///...`). No background HTTP daemon, no Node.js runtime, and no open listener ports are required.

2. **Core Change-maker Questions Answered Rapidly (<10ms):**
   - *"What breaks if I change this contract or endpoint?"* -> `gira impact --contract <contract-id>` or `gira impact --service <service-id>` emits the immediate Impact Set: affected Components, connecting Contracts, and direct consumers with file/line citations.
   - *"Who consumes this queue, topic, or event?"* -> `gira slice --contract kafka:<topic>` displays upstream publishers and downstream subscriber Components with explicit extraction origins.
   - *"Where is this Component or Contract declared and implemented?"* -> `gira locate <node-id>` and `gira explain <node-id>` identify declaring deployment manifests (Docker Compose, Helm values) and code entry points with exact line locators.

3. **Evidence Inspection (`gira explain`):**
   - Commands `gira explain <node-id>` and `gira explain --edge "<source> -> <target>"` surface the underlying discrete `EvidenceRecord` entries for any node or relationship.
   - Each evidence entry reveals:
     - Exact Git locator: `[repo@commit:path#Lstart-Lend]`
     - Discovery profile: Compose manifest, Helm values, or Code AST
     - Derivation tier: `declared`, `extracted`, or `inferred`
     - Freshness status: `current`, `stale`, or `unverified`
   - In accordance with ADR 0001 (index stores locators, not source bodies), `gira explain` retrieves source code excerpts on-demand from the local git checkout when present, or cleanly reports source unavailable.

4. **Guards Against Unwarranted Authority (Honest Boundaries):**
   - **Discrete Derivation Rubric:** Nodes and edges are labeled explicitly as `declared` (from manifests), `extracted` (from deterministic code parsers), or `inferred` (from naming or environment wiring heuristics).
   - **No Probabilistic Scores:** Adhering to ADR 0003, no continuous floating-point confidence scores (e.g. "85% confidence") are displayed; uncertainty is conveyed via discrete origin tiers.
   - **Freshness & Staleness Flags:** Displays the indexed commit SHA and marks components/edges as `stale` if the local repository HEAD differs from the indexed commit.
   - **Explicit Unsupported Boundaries:** Unhandled languages from Wave 1 (e.g. JavaScript, Python, C# ASTs) and unresolved bindings are visibly highlighted under a dedicated `Unsupported Boundaries & Known Unknowns` section, directing the Change-maker to manual verification targets.

5. **Declarative Human Corrections via Overlays:**
   - The MVP does not implement an interactive mutation system or an opaque mutable database for ad-hoc graph editing.
   - Human corrections are strictly declarative through a `gira.yaml` overlay file (e.g. `overrides: relationships: ...`).
   - Overrides are version-controlled in Git alongside the Target System configuration. When an override takes effect, `gira explain` attributes the edge to `origin: declared:override` with a locator pointing to the line in `gira.yaml`.

## Considered Options

- **Persistent Web UI Daemon & Browser Server (Understand Anything / Graphify style).** Rejected: running a long-running Node.js/Python web server or Go daemon introduces background lifecycle management, port collision risks, node/npm dependency footprints, and security concerns. Self-contained static HTML and rich CLI terminal views fulfill the Change-maker's needs with zero operational overhead.
- **Full-Screen Interactive TUI (Bubbletea modal application).** Rejected for primary workflows: Change-makers require fast, composable command-line output that can be piped to `less`, `grep`, or redirected into PR descriptions and agent scratchpads. Rich CLI output with lipgloss styling delivers visual clarity while remaining fully scriptable and pipe-compatible.
- **Mutable Local Database for Relationship Corrections.** Rejected: allowing developers to interactively edit or correct relationships in a local SQLite or graph database creates unversioned, non-reproducible local state that cannot be shared or validated in CI. Declarative Git-tracked YAML overlays ensure deterministic reproducibility across developers and agents.
- **Probabilistic Confidence Scores.** Rejected: continuous numbers give an illusion of algorithmic authority without auditable grounding. Discrete origin tiers (`declared`, `extracted`, `inferred`) provide transparent, actionable provenance.

## Consequences

- The harness remains a single Go binary with no external runtime dependencies.
- Change-makers obtain instantaneous answers (<10ms) directly in their terminal workflows.
- Evidence is always inspectable down to git commit, file path, and line numbers.
- Generated documentation cannot masquerade as ground truth because derivation tiers, indexing commits, and unsupported boundaries are prominently displayed.
- Human corrections are reproducible, reviewable, and version-controlled via declarative YAML overlays.

Source: [Human-facing surface for the MVP](../../../../.scratch/system-context-harness/issues/16-human-surface.md)
