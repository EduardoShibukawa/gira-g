# Versioned index publication and contract

**Status:** accepted (2026-10-18, by owner). Specified in [`specs/2026-10-17-gira-versioned-publication.md`](../../specs/2026-10-17-gira-versioned-publication.md). Consolidates architectural decisions N1–N4, D1–D9c, and technical Gates 1–5. Amends and aligns with [ADR 0001](0001-index-stores-locators-no-model-client.md) (locators and trust boundaries), [ADR 0003](0003-macro-map-schema-and-evidence-model.md) (discrete Macro Map graph schema), [ADR 0007](0007-evaluation-design-and-thresholds.md) (100% determinism), and [ADR 0009](0009-target-system-declaration-and-repo-fetching.md) (Git configuration).

## Context

In the legacy design (v1), the generated Macro Map index was published to an unversioned single file path identified only by target system name (`$CACHE/gira/indexes/<system>/macro-map.json`). This model introduced severe correctness and operational flaws:
- **No cryptographic binding:** Nothing tied the stored index to the exact manifest bytes, the extractor versions, or the binary that produced it. An agent probe (scenario S2) demonstrated that a stale index was served as current, returning empty impact sets (`[]`) with an exit code of 0.
- **Flawed `--output` semantics:** Flags `index --output P` and `init --output P` bypassed the store and created inconsistencies with project discovery; `init` wrote `gira.yaml` before index construction, leaving broken manifests on disk if builds failed.
- **Environment and path leakage:** Override locators stored paths as typed on the CLI (yielding divergent hashes between relative and absolute invocations); repository caching respected `XDG_CACHE_HOME` while index storage ignored it; and Git reads inherited user configuration without `-z` safety.
- **Concurrency and platform hazards:** Multi-process invocations lacked mutual exclusion, and durability guarantees on Windows were unconfirmed due to lack of directory metadata synchronization.

## Nomenclature and Roadmap Positioning (N1, N2, N4)

1. **Etapa 1.2 nomenclature (N1):** This versioned publication contract was referenced during exploratory planning as "Etapa 2" (sub-stages 2a–2i). By explicit owner decision N1, it is formally labeled **Etapa 1.2**, positioned as a direct continuation and completion of Item 1 of the approved evolution roadmap. Item 2 of the roadmap remains reserved for end-to-end freshness and multi-repo synchronization.
2. **Prior stage acceptance (N2):** Baseline Etapa 1/1.1 (at commit `61b58c2`) was formally accepted following adversarial review of commits `4c2204c` and `0d5f248`.
3. **Store as derived cache (N4):** The index store is strictly a local, regenerable, unshared cache. Git remains the sole authority for shared, versioned knowledge (manifests and checked-in snapshots).

## Decision

### 1. Schema 2 and Embedded `source_identity` (N3, D4, D5, Gate 1, Gate 3)

- **Integer schema generation (Gate 1, Option B):** The Macro Map advances to `schema_version: 2` represented as an integer (not a SemVer string). This ensures older binaries expecting a string fail closed on `json.Unmarshal` rather than silently ignoring embedded provenance or serving invalid data.
- **Embedded `source_identity` (N3, Gate 1):** Both canonical published indexes and detached snapshots embed a self-contained `source_identity` object directly within the Macro Map. Sidecar metadata files are rejected for portable snapshots.
  ```json
  "source_identity": {
    "schema_version": 2,
    "manifest_sha256": "<64-hex SHA-256 of raw manifest bytes>",
    "extractors": [
      { "id": "<extractor-id>", "version": "<semver>" }
    ],
    "sources": [
      {
        "repo_id": "<repo-id>",
        "sha": "<40-hex commit SHA>",
        "path_scope": "<path>",
        "url_sha256": "<gitcache RepoHash>"
      }
    ]
  }
  ```
- **Raw manifest bytes for `source_key` (D4):** `source_key` is computed strictly over the raw, unparsed bytes of the manifest. Variations in comments, formatting, BOM, or CRLF/LF line endings produce distinct keys, forcing re-indexing to ensure key and parser observe identical content.
- **Source list provenance:** `sources` enumerates all repositories declared in `gira.yaml` in deterministic order, even when no extracted evidence is found. Raw clone URLs are never embedded (preventing credential or path leakage); `url_sha256` is used instead.
- **Build identity `tool_id` (D5, Gate 3, Option C):** `tool_id` binds the binary identity: `main.version`, `main.commit`, Go version, dependencies, and explicit semantic compiler flags. The entire `-ldflags` setting is excluded (Option C), making `main.date` purely informational and preventing spurious invalidation across identical release builds. For untrusted development builds (where `commit == "none"` or `vcs.modified == true`), `tool_id` includes the executable binary's SHA-256 hash.
- **Version ID:** Each index version is uniquely and deterministically addressed by:
  $$\text{version\_id} = \text{SHA-256}(\text{source\_key} \parallel \text{tool\_id})$$
- **Internal model integrity:** Functions `cloneMacroMap` and `CanonicalJSON` must preserve all fields (`source_identity`, `UnresolvedRefs`, `Conflicts`) without data loss.

### 2. Strict Reader Policy R2 and Error Taxonomy (D9, D9b, D9c, Gate 1, Gate 2)

- **Strict schema allowlist (Policy R2):** The reader accepts only explicit schema versions (`2` as integer, and `"1.0.0"` string under legacy constraints). Unknown schema versions fail closed with `index_schema_unsupported`, which is distinct from file corruption (`index_corrupt`).
- **Legacy index isolation (D9):** Legacy `"1.0.0"` indexes are never loaded implicitly when `gira.yaml` is present. They are permitted solely in read-only mode via explicit `--index <path>` flags, accompanied by a visible deprecation warning. Legacy indexes are never modified or upgraded in-place.
- **Downgrade prevention:** The publisher strictly rejects schema downgrades; a legacy index cannot overwrite or supersede a schema 2 version.
- **Stable error taxonomy:**
  - `index_missing`: Required index version or lineage does not exist (replaces empty sets `[]` with exit 0).
  - `index_tool_mismatch`: Cached index exists for the manifest but was built with an incompatible tool/binary version; fails closed without serving stale data.
  - `index_schema_unsupported`: Schema version is unknown, incompatible, or improperly typed.
  - `index_corrupt`: Stored file is truncated, contains malformed JSON, or fails checksum verification.
  - `index_ambiguous`: Discovery without a manifest yields ambiguous candidate lineages or requires missing `tool_id` (D9b).
  - `nondeterministic_build`: Re-indexing an identical `version_id` produces differing bytes; publication fails closed without overwrite.
  - `manifest_outside_git_root`: Manifest with overrides is outside a Git worktree (Gate 2, Option O1).
  - `git_root_unavailable`: Git discovery probe fails or returns ambiguous ownership.
  - `manifest_path_required`: Evaluation requires an override locator but no manifest path was provided (Gate 2, removing silent `"gira.yaml"` fallback).
  - `index_locked`: Concurrency timeout (60s) reached while waiting for index locks.
  - `lock_unavailable`: Platform locking primitives fail or are unsupported on the underlying filesystem.
- **Skill CLI normalization (D9c):** `skills install --index` normalizes relative index paths to absolute paths and emits warnings when targeting legacy formats.

### 3. Versioned Publication and Storage Architecture (D1, D2, D7)

- **Store location hierarchy (D2):** The unified resolver searches in order:
  1. `$XDG_CACHE_HOME/gira/indexes`
  2. `~/.cache/gira/indexes`
  3. `<project_root>/.gira/indexes` (fallback)
- **Immutable version layout:**
  ```
  <store_root>/
    <lineage_id>/
      .head                # Atomically references active version_id
      versions/
        <version_id>/
          macro-map.json   # Immutable versioned Macro Map (schema 2)
          meta.json        # Strict publication metadata
  ```
- **Atomic publication sequence:**
  1. Acquire locks in strict order **M -> S**.
  2. Write canonical version artifacts to a temporary directory on the same filesystem.
  3. Sychronize version files (`fsync`) and atomically rename directory to `<version_id>`.
  4. Atomically point `<lineage_id>/.head` to the new `<version_id>` (write temporary head, fsync, rename).
  5. Sychronize parent directory metadata.
- **CLI output semantics (D1):**
  - `gira index --output P`: Produces a detached snapshot written directly to path `P`. It does not touch the store, create heads, or modify `gira.yaml`. A prominent CLI message warns that `P` is a detached snapshot and not the active index.
  - `gira init --output P`: Publishes canonically to the store, updates `.head`, and writes an identical byte-for-byte copy to `P`. If canonical publication succeeds but `P` copy fails, partial success is reported. Canonical publication occurs first; `gira.yaml` is written **last**. Any build or gate failure aborts without creating or modifying `gira.yaml`.
- **Retention (D7):** Automatic garbage collection is omitted. Past versions remain immutable until manual removal. Quarantine isolation occurs only under explicit `--recover-corrupt`.

### 4. Cross-Platform Concurrency and Locking (D8, Gate 4)

- **Global lock acquisition hierarchy:** Writers must acquire locks in strictly ordered sequence:
  $$\mathbf{M} \text{ (Manifest Lock)} \longrightarrow \mathbf{S} \text{ (Store Lock)} \longrightarrow \mathbf{P} \text{ (Snapshot Lock)}$$
- **Platform locking primitives:**
  - Unix / macOS: Native advisory file locking via `flock(fd, LOCK_EX)`.
  - Windows: Native file range locking via `LockFileEx(LOCKFILE_EXCLUSIVE_LOCK)`.
- **Lock lifecycle & timeouts:** A uniform 60-second deadline governs lock acquisition before returning `index_locked`. Lock handles are bound to OS file table entries; termination of the owning process or child immediately releases the lock.
- **Canonical lock identity:** Locks bind to physical file identity (`os.SameFile` / file IDs), never to textual file paths. This eliminates alias evasion through symlinks, hardlinks, 8.3 short names, directory junctions, case variants, or `\\?\` prefixes.
- **Network filesystems:** Remote filesystems (NFS, SMB, FUSE) lacking robust POSIX/Win32 lock semantics fail closed with `lock_unavailable`.

### 5. Platform Boundaries, Durability, and Normalized Locators (D3, D6, Gate 2, Gate 4, Gate 5)

- **Git-relative locator normalization (D3, Gate 2 - O1):** Locators inside the Macro Map are normalized relative to the Git repository root using forward slashes (`/`). If a manifest with overrides sits outside a Git repository, publication fails closed with `manifest_outside_git_root`. Git environment variables (`GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`, etc.) are sanitized during probes and initialization to prevent hijacking.
- **Git for Windows extraction limits (Gate 4):** Paths valid in Git objects but illegal in Win32 file systems (e.g. filenames containing tabs `\t`, colons, trailing dots/spaces) fail closed immediately. Partial extractions are never treated as successful builds.
- **Durability semantics (D6, Gate 5):**
  - **Unix / POSIX:** Metadata and parent directory `fsync` are enforced. If synchronization fails after data becomes visible, the CLI emits non-zero exit status, reports `visible_unconfirmed`, and aborts writing `gira.yaml` in `init`.
  - **Windows:** Win32 does not provide reliable directory metadata flushing; publication reports `visible_unconfirmed` with exit code 0 and an explicit warning.

## Considered Options

- **Option A: SemVer string `"2.0.0"` for schema version.** Rejected: Legacy binaries unmarshal string schema versions without error and silently drop unknown fields such as `source_identity`, permitting silent downgrades and cache corruption.
- **Sidecar metadata files (`P.meta.json`) for snapshots.** Rejected: Portable snapshots shared across machines or committed to Git must be completely self-contained (N3).
- **Automatic retention (K=3) and automatic garbage collection.** Rejected: Dropping versions automatically can break concurrent readers, long-running agent sessions, or external tools referencing specific heads.
- **Unsynchronized `gira.yaml` creation during `init`.** Rejected: Generating `gira.yaml` prior to build leaves incomplete manifests on disk when extraction or gates fail.
- **Text-based lock identification.** Rejected: Fails to detect aliased file paths (symlinks, junctions, 8.3 short names), causing race conditions and lock bypass.
- **Silent fallback to `"gira.yaml"` in evaluation.** Rejected: Distorts provenance and breaks reproducibility; explicit manifest paths are required when overrides are evaluated.

## Consequences

- Published indexes achieve complete, auditable provenance cryptographically bound to raw manifest bytes, source commits, and build tool identities.
- Readers and publishers operate under a fail-closed regime, eliminating silent downgrade, stale index delivery, and ambiguous discovery.
- Cross-platform atomic publication guarantees consistency across Linux, macOS, and Windows, with verified lock exclusion and predictable durability reporting.
- Downstream tools (`gira diff`, `skills`, MCP adapters, and CI harnesses) consume deterministic, immutably versioned snapshots.
