---
sidebar_position: 1
---

# Quickstart

Get from a clean checkout to your first Impact Briefing in three steps.

## 1. Install

Because the repository is private, release downloads require a GitHub token with
`contents: read` access (`export GITHUB_TOKEN="$(gh auth token)"`).

```bash
# Option A: Install via mise (native github backend)
export GITHUB_TOKEN="$(gh auth token)"
mise use -g github:EduardoShibukawa/gira-g@0.1.0

# Option B: Install via install.sh (SHA-256 verified; pass --cosign for Sigstore bundle check)
export GITHUB_TOKEN="$(gh auth token)"
curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://raw.githubusercontent.com/EduardoShibukawa/gira-g/main/install.sh | sh

# Option C: Build from source
go build ./cmd/gira

# Verify the installed binary
gira version
```

## 2. Build an index

```bash
cd /path/to/target-system
gira index --output .gira/index.json
```

This runs the Extractors against your manifest and dependencies, writing
`macro-map.json`.

## 3. Ask the first question

```bash
# What breaks if you change a Component?
gira impact <component-id> --index .gira/index.json

# Or analyze a diff
gira impact --from-diff HEAD~1..HEAD --index .gira/index.json

# Extract a slice
gira slice <component-id> --tier summary --index .gira/index.json
```

## Next steps

- [Change-maker workflow](./change-maker-workflow.md)
- [Evidence inspection](./evidence-inspection.md)
- [Concepts](../concepts/target-system.md)
