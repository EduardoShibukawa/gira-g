# Target System declaration and repository fetching mechanics

Gira-G (`gira`) declares a Target System through a declarative YAML manifest (`gira.yaml` by default, or specified via `--manifest <path>`). The harness fetches declared repositories into a deterministic local bare cache using the user's host Git configuration and credentials, pins extraction strictly to immutable commit SHAs, operates with a strict network-free `--offline` mode, and surfaces unreachable repositories explicitly as `unreachable/unknown` without heuristic guessing or silent omission.

## Decision Summary

1. **Declarative Manifest Format (`gira.yaml`):**
   - The Target System is defined in a declarative configuration file (`gira.yaml`), discoverable at the repository root or passed via CLI flag `--manifest <path>`.
   - Each repository declaration specifies:
     - `id`: Logical unique identifier within the Target System (e.g. `demo`, `helm-charts`).
     - `url`: Remote Git repository URL (`https://`, `ssh://`, `git@`) or local directory path (`file://` or relative path).
     - `ref`: Human-readable reference (tag or branch, e.g. `v3.1.0`, `main`).
     - `sha`: Mandatory 40-character commit SHA ensuring immutable and deterministic index builds.
     - `path_scope`: Optional relative directory path (e.g. `charts/opentelemetry-demo`). When present, Extractors constrain indexing exclusively to that subtree, scoping Component boundaries accordingly.
   - Example Reference Corpus declaration (`gira.yaml`):
     ```yaml
     version: 1
     target_system:
       name: opentelemetry-demo
       description: "OpenTelemetry Demo Reference Corpus (v3.1.0)"
     repositories:
       - id: demo
         url: "https://github.com/open-telemetry/opentelemetry-demo.git"
         ref: "3.1.0"
         sha: "dedc0178918e260823323b8d95005a8cb924b007"
         path_scope: "."
       - id: helm-charts
         url: "https://github.com/open-telemetry/opentelemetry-helm-charts.git"
         ref: "opentelemetry-demo-0.42.3"
         sha: "a9f2dd0131b933f828049124b113957882b05518"
         path_scope: "charts/opentelemetry-demo"
     ```

2. **Fetch and Local Cache Mechanics:**
   - **Cache Location:** Bare git repositories are cached in `$XDG_CACHE_HOME/gira/repos/<repo_hash>.git` (defaulting to `~/.cache/gira/repos/<repo_hash>.git` on Linux/macOS), where `<repo_hash>` is the SHA256 hash of the normalized repository URL.
   - **Credential Management & Transport:** Harness-initiated cloning and fetching delegates directly to the host's system `git` binary (utilizing existing SSH keys, SSH agent, and Git credential helpers). The harness itself never prompts for, stores, or serializes credentials into the index, logs, or cache metadata (preserving ADR 0001).
   - **Tree Extraction at Pinned SHA:** Content is checked out ephemerally or read via direct Git tree object inspection at the pinned `sha`. If the remote branch/tag does not match the pinned `sha`, the harness aborts with an integrity error to prevent non-deterministic index drift.

3. **Strict `--offline` Mode:**
   - The `--offline` CLI flag completely disables harness network egress (ADR 0001).
   - In offline mode, `gira` verifies that all declared repository caches exist locally and contain the pinned commit SHAs. If any required commit or repository is missing, the harness fails immediately with an error detailing the missing commit, without performing any remote network call.

4. **Handling Unreachable Repositories and Network Failures:**
   - If a declared repository cannot be cloned or fetched (network unavailable, DNS failure, authentication error, deleted upstream), the harness records its status explicitly as `unreachable` or `unknown`.
   - The harness never guesses, hallucinates, or silently skips an unreachable repository.
   - Components mapped from unreachable repositories are recorded in the Macro Map with derivation status `unknown` and diagnostic error evidence, ensuring the Change-maker and agent are informed of the blind spot.

## Considered Options

- **Dynamic Org / SCM Discovery (e.g. querying GitHub/GitLab Org APIs at runtime).**
  - *Rejected:* Dynamic discovery introduces uncontrolled network egress, requires API tokens with broad read permissions, and causes non-deterministic index builds as repositories are added or renamed. A declarative manifest ensures version-controlled reproducibility and zero unauthorized egress.
- **In-process Git (`go-git`) vs Delegated System Git CLI.**
  - *Adopted:* Delegated system `git` CLI for network operations (fetch, clone) behind an internal Git seam, coupled with `go-git` for local tree/blob parsing. System `git` handles complex host authentication (SSH agents, smart cards, Git Credential Manager, enterprise proxies) out of the box without requiring the harness to handle secrets or manage credential lifecycles.
- **Cloning Repositories directly into the User's Current Workspace.**
  - *Rejected:* Dumping auxiliary or consumer repositories into the working directory litters user project trees, triggers unwanted IDE re-indexing, and risks accidental commits into the wrong repository. Caching in `~/.cache/gira/repos/` keeps workspace checkouts clean.
- **Optional or Dynamic Commits (`HEAD`, `latest`, floating branch names).**
  - *Rejected:* Pinned commit SHAs are required for index builds. Floating references make benchmark comparisons invalid and break deterministic reproducibility (ADR 0001, Issue 11). Floating branches are allowed in the declaration as informational hints (`ref`), but `sha` is strictly enforced.

## Consequences

- Target Systems spanning multiple repositories (such as the Reference Corpus OTel Demo code + Helm chart) are expressed cleanly in a single reproducible manifest.
- CI pipelines can pre-populate `~/.cache/gira/repos/` and run `gira build --offline` in fully isolated environments.
- The trust boundary (ADR 0001) is preserved: git fetch with user credentials is the sole default egress, and offline mode is guaranteed.

Source: [How a user declares and fetches a Target System's repos](../../../../.scratch/system-context-harness/issues/23-declare-and-fetch-target-system.md)
