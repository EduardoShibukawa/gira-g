---
sidebar_position: 1
---

# Quickstart

Get from a clean checkout to your first Impact Briefing in three steps.

## 1. Install

Public releases do not require authentication or tokens.

```bash
# Option A: Install via mise (native github backend)
mise use -g github:EduardoShibukawa/gira-g@0.1.0

# Option B: Install via install.sh (SHA-256 verified; pass --cosign for Sigstore bundle check)
curl -fsSL https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh

# Option C: Build from source
go build ./cmd/gira

# Verify the installed binary
gira version
```

> **Note (`GITHUB_TOKEN` optional):** Exporting `GITHUB_TOKEN="$(gh auth token)"` is optional and supported for authenticated GitHub API rate limits or private repository forks.

## 2. Build an index (`gira init` vs `gira index`)

Gira supports both zero-config single-repository onboarding and manifest-driven multi-repository indexing:

- **Single repository (`gira init`):** Generates `gira.yaml` pinned to the local Git `HEAD` (when absent) and builds the Macro Map index immediately.
- **Multi-repository (`gira index`):** Reads an existing `gira.yaml` manifest declaring one or more repositories and pinned commit SHAs.

### Where the index is stored and resolved

- **Default user cache:** When `--output` is omitted, both `gira init` and `gira index` persist the canonical index to `~/.cache/gira/indexes/<system>/macro-map.json` (falling back to `.gira/indexes/<system>/macro-map.json`).
- **Automatic `--index` resolution:** When `--index` is omitted on query commands (`impact`, `slice`, `explain`, `tour`, `flows`, `doc`, `locate`, `export`, `view`, `serve mcp`), Gira automatically searches in order:
  1. `.gira/macro-map.json`
  2. `macro-map.json`
  3. `~/.cache/gira/indexes/<system>/macro-map.json`
  4. `.gira/indexes/<system>/macro-map.json`
- **Explicit path:** Pass `--output .gira/macro-map.json` (or `.gira/index.json`) when indexing and `--index .gira/macro-map.json` when querying to pin a specific file path in your workspace.

```bash
# Zero-config in any local Git repository
cd /path/to/target-repo
gira init . --output .gira/macro-map.json

# Or from an existing gira.yaml manifest
gira index --output .gira/macro-map.json
```

### Runnable in-repo example (`testdata/dogfood/services`)

You can run an end-to-end walkthrough immediately using the public fixture included in the repository:

```bash
cp -r testdata/dogfood/services /tmp/gira-demo
git -C /tmp/gira-demo init -q
git -C /tmp/gira-demo add -A
git -C /tmp/gira-demo -c user.name=demo -c user.email=demo@local commit -q -m "init"

# Zero-config init + index
gira init /tmp/gira-demo --output /tmp/gira-demo/.gira/macro-map.json

# Inspect tour, service doc, and impact briefing
gira tour --format md --index /tmp/gira-demo/.gira/macro-map.json
gira doc service:order --format md --index /tmp/gira-demo/.gira/macro-map.json
gira impact service:order --format md --index /tmp/gira-demo/.gira/macro-map.json
```

## 3. Ask architectural questions

```bash
# What breaks if you change a Component?
gira impact <component-id> --index .gira/macro-map.json

# Or analyze a git revision range diff
gira impact --from-diff HEAD~1..HEAD --index .gira/macro-map.json

# Extract a bounded context slice
gira slice <component-id> --tier summary --index .gira/macro-map.json

# Generate a single-service onboarding document
gira doc <service-id> --format md --index .gira/macro-map.json
```

## Next steps

- [Change-maker workflow](./change-maker-workflow.md)
- [Evidence inspection](./evidence-inspection.md)
- [Concepts](../concepts/target-system.md)
