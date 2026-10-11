# Go core, with the official tree-sitter binding behind a seam

The harness core is written in Go. Extractors are compiled into the binary behind one internal interface whose output is a data contract (facts with evidence), not loaded at runtime. Programming-language parsing uses the official CGO binding of tree-sitter, pinned to a commit, behind an internal parse-tree seam so a pure-Go runtime can replace it later. We chose Go because the inputs that define Components — deployment YAML, Helm, Kubernetes, HCL, OpenAPI, git — have reference-grade Go libraries, and Go's one gap (tree-sitter) is closable inside Go, whereas Rust's gaps (no Helm renderer, deprecated and unsound YAML crates, pre-1.0 `gix`) are not closable inside Rust.

## Considered Options

- **Rust.** Rejected despite its upstream tree-sitter binding and runtime Wasm grammars: with Extractors compiled in, Wasm loading buys nothing, and its weaknesses fall on the Macro Map's primary evidence (deployment and contract files).
- **TypeScript compiled with Bun.** Rejected for the core: largest binary, Apple-system SQLite on macOS, no native HCL or Helm. Still a candidate for repo scripts and CI.
- **Pure-Go `gotreesitter`, no C at all.** Deferred, not rejected: created 2026-02, one dominant author, 119 hand-ported external scanners without independent parity checks. A silently diverging scanner yields a wrong Impact Set with nothing to flag it. The seam lets it replace the CGO binding once parity on the MVP languages is proven.
- **Runtime-loaded Extractors (Wasm or plug-ins).** Rejected for the MVP: no outside contributors, a public plug-in API would freeze before the schema settles, and installed plug-in versions would make the index non-deterministic. The data-contract output keeps an out-of-process path open.
- **tree-sitter in Rust, the rest in Go.** Rejected: both bindings wrap the same C runtime, so parse output is identical; it adds a second toolchain and an FFI bridge to fix only the stalled Go wrapper.
- **Rust's `tree-sitter-loader` (compile grammars on the user's machine).** Rejected: needs a C compiler wherever the binary runs, and is runtime loading.
- **C tree-sitter compiled to Wasm, run in Go (e.g. `wazero`).** Not evaluated; a candidate replacement behind the seam that would keep the real C scanners and drop CGO.

## Consequences

The build needs a C toolchain (CGO cross-compiled with `zig cc`); CGO-free SQLite is no longer a requirement. The official Go binding has stalled since 2025-11, so this repo may have to maintain a fork of its thin wrapper.

Source: [Choose the implementation language](../../../../.scratch/system-context-harness/issues/09-choose-implementation-language.md)
