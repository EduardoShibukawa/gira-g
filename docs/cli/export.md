---
sidebar_position: 5
---

# `gira export` and `gira view`

## `gira export`

Export a standalone architecture document (single-file Markdown or HTML).

```text
gira export [--format md|html] [--output <path>] [--index <path>]
```

| Flag | Description |
|---|---|
| `--format` | `md` (default) or `html` |
| `--output` | Path to write the export |
| `--index` | Path to `macro-map.json` |

## Output

- `gira export`: a single self-contained file (Markdown or HTML) written to
  `--output` (or stdout). HTML has no external asset reference.
- `gira view`: an HTML file in `--output` (default temp dir), then launched in
  the default browser unless `--no-open`.

## `gira view`

Export HTML and open it in the default browser.

```text
gira view [--output <dir>] [--index <path>] [--no-open]
```

| Flag | Description |
|---|---|
| `--output` | Directory for the generated HTML |
| `--index` | Path to `macro-map.json` |
| `--no-open` | Generate without launching the browser |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Usage, flag, index-load, engine, or emit error |

No other exit codes. Errors print `Error: ...` to stderr.

## See also

- [Export guide](../guides/export.md)
