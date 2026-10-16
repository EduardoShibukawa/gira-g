# Third-party notices

This file attributes third-party works distributed with or used to evaluate
`gira-g` (repository `EduardoShibukawa/gira-g`, MIT; see `LICENSE`).

## Evaluation corpus subset (vendored)

- `testdata/corpus/otel-demo-3.1.0-mini/` is a minimal, unmodified subset of
  the [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-demo),
  release 3.1.0 (commit `dedc0178918e260823323b8d95005a8cb924b007`).
- Copyright The OpenTelemetry Authors. Licensed under the Apache License,
  Version 2.0. The full license text ships with the subset at
  `testdata/corpus/otel-demo-3.1.0-mini/LICENSE`; see also
  `testdata/corpus/otel-demo-3.1.0-mini/NOTICE`.
- It is vendored only as a hermetic evaluation corpus and is not linked into
  the `gira` binary.

## Go module dependencies

Built `gira` binaries include code from these Go modules (see `go.mod` for
pinned versions):

- `github.com/charmbracelet/lipgloss` (MIT)
- `github.com/muesli/termenv` (MIT)
- `gopkg.in/yaml.v3` (MIT)
- `github.com/aymanbagabas/go-osc52/v2` (MIT, indirect)
- `github.com/charmbracelet/colorprofile` (MIT, indirect)
- `github.com/charmbracelet/x/ansi` (MIT, indirect)
- `github.com/charmbracelet/x/cellbuf` (MIT, indirect)
- `github.com/charmbracelet/x/term` (MIT, indirect)
- `github.com/lucasb-eyer/go-colorful` (MIT, indirect)
- `github.com/mattn/go-isatty` (MIT, indirect)
- `github.com/mattn/go-runewidth` (Apache-2.0, indirect)
- `github.com/rivo/uniseg` (MIT, indirect)
- `github.com/xo/terminfo` (MIT, indirect)
- `golang.org/x/sys` (BSD-3-Clause, indirect)

Full license texts for each module are available in the Go module cache and
in each module's source repository.

## Documentation website (build-time only)

The `website/` Docusaurus site uses the following JavaScript packages at
build time only (they are not shipped with the `gira` binary):

- `@docusaurus/core` (MIT)
- `@docusaurus/preset-classic` (MIT)
- `@docusaurus/module-type-aliases` (MIT, dev)
- `@docusaurus/tsconfig` (MIT, dev)
- `@docusaurus/types` (MIT, dev)
- `@easyops-cn/docusaurus-search-local` (MIT)
- `@mdx-js/react` (MIT)
- `clsx` (MIT)
- `prism-react-renderer` (MIT)
- React (MIT)
- `react-dom` (MIT)
- TypeScript (Apache-2.0, dev)

Docusaurus upstream: https://github.com/facebook/docusaurus (MIT, with CC-BY
documentation). Full license texts ship under `website/node_modules/`.
