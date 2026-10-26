---
sidebar_position: 8
---

# Macro Map

## Definition

The default view of a Target System: its Components, their interfaces, and the
relationships between them.

**Avoid**: graph, code graph, architecture diagram

## Schema 2 and Provenance (`source_identity`)

Canonical Macro Maps use `schema_version: 2` (integer). Both canonical store versions and detached snapshots carry an embedded `source_identity` object binding the index to its source provenance:

```json
{
  "schema_version": 2,
  "target_system": "my-system",
  "source_identity": {
    "schema_version": 2,
    "manifest_sha256": "<64-hex SHA-256 of raw manifest bytes>",
    "extractors": [
      { "id": "go", "version": "1.0.0" }
    ],
    "sources": [
      {
        "repo_id": "main-repo",
        "sha": "<40-hex commit SHA>",
        "path_scope": ".",
        "url_sha256": "<repo-hash>"
      }
    ]
  },
  "nodes": [...],
  "relationships": [...]
}
```

- **Manifest Binding:** `source_key` is computed directly from raw manifest bytes, preventing drift across line endings or formatting.
- **Tool Identity:** `tool_id` cryptographically captures binary version, commit, Go version, and dependencies, failing closed with `index_tool_mismatch` if the binary changes.
- **Version Addressing:** Each version is uniquely identified by `version_id = SHA-256(source_key || tool_id)[:16]`.

## Versioned Store and Detached Snapshots

- **Store Hierarchy (D2):** Resolved across `$XDG_CACHE_HOME/gira/indexes` -> `~/.cache/gira/indexes` -> `<project_root>/.gira/indexes`.
- **Layout:** Stored under `<store_root>/<lineage_id>/versions/<version_id>/macro-map.json` alongside strict publication `meta.json`. An atomic `.head` file points to the active version.
- **Detached Snapshots (`--output P`):** Portable, self-contained copies created via `gira index --output P` that do not touch the active store or `.head`.

## Reader Policy R2 and Legacy Isolation

- **Allowlist Enforcement:** Readers accept only explicit schemas (`2` integer, and `"1.0.0"` string under legacy constraints). Unknown schemas fail closed with `index_schema_unsupported`.
- **Legacy Isolation:** `"1.0.0"` indexes are never loaded implicitly when `gira.yaml` is present. They are accessible read-only via explicit `--index <path>` with a visible warning and are never upgraded in-place.

## Related

- [Target System](./target-system.md)
- [Component](./component.md)
- [Contract](./contract.md)
- [Evidence Record](./evidence-record.md)
- [Extractor](./extractor.md)
- [`gira index`](../cli/index.md)
- [`gira init`](../cli/init.md)
