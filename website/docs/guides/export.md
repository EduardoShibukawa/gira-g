---
sidebar_position: 4
---

# Export

Gira produces standalone documentation artifacts without a server or daemon.

## `gira export`

Generate a single-file Markdown or HTML document.

```bash
gira export --format html --output arch.html --index .gira/index.json
gira export --format md --output arch.md --index .gira/index.json
```

## `gira view`

Export HTML and open it in the default browser.

```bash
gira view --index .gira/index.json
gira view --no-open --output /tmp/arch-html --index .gira/index.json
```

## Limitations

- No interactive web server / daemon (ADR 0008).
- The HTML export is a single self-contained file with no external assets.
- Markdown and HTML share the same content.

## See also

- [CLI: export](../cli/export.md)
