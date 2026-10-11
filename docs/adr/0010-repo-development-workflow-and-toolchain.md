# Repository development workflow, toolchain pinning, CI/CD, and quality gates

Gira-G (`gira`) standardizes its development toolchain on **mise** (`mise.toml` pinned via `mise.lock`), uses **mise tasks** as the canonical task runner, executes core product verification natively via Go tooling (`go test`, `golangci-lint`), and confines TypeScript/Bun strictly to auxiliary repository automation scripts behind mise task facades. CI/CD runs on **GitHub Actions** across a native 3-OS matrix (Ubuntu, macOS, Windows) with `jdx/mise-action`. Binary distribution uses **GoReleaser** with split builds across native matrix runners (solving CGO cross-compilation for the official tree-sitter binding), coupled with Syft SBOM generation, SHA-256 checksums, and Cosign keyless signatures. Every change must pass 100% of quality gates locally and in CI, including automated licensing safeguards for prior-art code (`THIRD_PARTY_NOTICES`).

## Decision Summary

1. **Toolchain Pinning (`mise.toml` and `mise.lock`):**
   - `mise` is the universal, authoritative toolchain manager across developer workstations and CI runners.
   - Pinned tools in `mise.toml`:
     - `go`: Exact version (e.g. `1.24.x`).
     - `bun`: Exact version (e.g. `1.4.x`) for repo scripts only.
     - `golangci-lint`: Exact version for Go code quality.
     - `goreleaser`: Exact version for release builds.
     - `syft` & `cosign`: Exact versions for SBOM generation and artifact signing.
     - `lefthook`: Exact version for Git hooks.
     - `zig`: Pinned optional utility for local CGO cross-compilation.
   - Checksums for every platform are pinned in `mise.lock`.
   - CI uses `jdx/mise-action` pinned to an immutable release tag/hash, avoiding dynamic auto-adoption lags.

2. **Task Runner and Repository Automation (mise + Bun):**
   - **Canonical Entry Point:** All developer commands run through mise tasks (`mise run <task>`). Neither developers nor CI call underlying raw scripts directly.
   - **Inline Task Portability:** Because mise executes inline tasks under `cmd /c` on Windows and `sh` on Unix, all task definitions in `mise.toml` are restricted to single-line cross-platform invocations (e.g. `go test ./...` or calling a dedicated script runner).
   - **Repository Scripts in TypeScript (Bun):** Repository automations, benchmark data generation, and evaluation harness orchestration are written in TypeScript and executed via Bun (`bun run scripts/<task>.ts`).
   - **Bun 1.4 Protection & Script Discipline:**
     - To prevent Windows Bun Shell panics (`#43604`) and OS shell incompatibilities (Bun shell has only ~18 builtins without `cp`, `rm`, `sed`, `grep`), scripts **must not** use Bun Shell `$` template literals for OS commands.
     - Scripts must strictly use portable Node/Bun standard libraries (`node:fs`, `node:path`, `node:child_process` with argument arrays, or native `Bun.file` / `Bun.write`).
     - Because the Bun runtime does not type-check, a mandatory `tsc --noEmit` quality gate runs on all TypeScript scripts.
   - **Core Tooling Separation:** Bun/TypeScript is strictly forbidden inside the core product codebase (`labs/gira-g/`). The `gira` binary itself has zero runtime or build dependency on Bun or Node.

3. **Core Language Tooling (Native Go):**
   - The `gira` core product utilizes Go's native tooling directly:
     - `go test -race -v ./...` for unit and integration testing.
     - `golangci-lint run` for static analysis and linting.
     - `go vet ./...` and `gofmt` for formatting and correctness checks.
   - All tests must pass with zero flakiness.

4. **CI Provider & Architecture (GitHub Actions Matrix):**
   - GitHub Actions provides automated gate enforcement on pull requests and commits to `main`.
   - **Runner Matrix:** Full cross-platform verification across Tier-1 platforms:
     - `ubuntu-latest` (Linux x86_64)
     - `macos-latest` (macOS Apple Silicon / ARM64)
     - `windows-latest` (Windows x86_64)
   - Every runner initializes the toolchain via `jdx/mise-action`.
   - **Complete Gate Parity:** CI executes the exact same suite (`mise run check`) as local environments. Zero failures tolerated; since AI coding agents write code and the user does not conduct manual line-by-line reviews, CI carries the full verification burden.

5. **Release Pipeline (GoReleaser with CGO):**
   - Triggered on semver git tags (`v*.*.*`).
   - **CGO Handling for Tree-Sitter:** ADR 0002 mandates the official CGO tree-sitter binding. To avoid cross-compilation failures on CGO (specifically MSVC/MinGW on Windows and Mach-O on macOS), the release pipeline uses **GoReleaser Split Builds** across GitHub Actions native matrix runners (`ubuntu-latest`, `macos-latest`, `windows-latest`). Each runner compiles natively with its native C compiler.
   - A subsequent aggregator job merges release archives, computes unified `checksums.txt`, generates SBOMs using `syft`, signs binaries with `cosign` (keyless via GitHub OIDC / `actions/attest@v4`), and publishes the GitHub Release.
   - Outputs a single, self-contained binary `gira` per target architecture:
     - `gira-linux-amd64`, `gira-linux-arm64`
     - `gira-darwin-amd64`, `gira-darwin-arm64` (and universal binary)
     - `gira-windows-amd64.exe`

6. **Quality Gates and Prior-Art Licensing Guardrails:**
   - **Local Hooks:** Managed via `lefthook`:
     - `pre-commit`: Fast checks (`gofmt`, fast linting, `tsc --noEmit`).
     - `pre-push`: Complete check suite (`mise run check`).
   - **The Quality Suite (`mise run check`):**
     1. *Formatting:* `gofmt` / `goimports` for Go, `prettier` for TypeScript, Markdown, JSON, YAML.
     2. *Type & Static Checking:* `golangci-lint run`, `tsc --noEmit`.
     3. *Test Suite:* 100% pass rate on `go test -race ./...`.
     4. *Contract & Schema Validation:* Golden test verification on Macro Map schema and Evidence Record contracts.
     5. *Licensing & Prior-Art Safeguards (`mise run check:licenses`):*
        - Confirms neither `graphify` nor `understand-anything` is present in `mise.toml`, CI configs, git hooks, or package dependencies (ADR 0002, Issue 10).
        - Verifies that any verbatim third-party code (from Apache-2.0 or MIT prior art) maintains its upstream license header and has a valid entry in `THIRD_PARTY_NOTICES.md`.
        - Enforces the "reimplement, don't copy" mandate.

## Considered Options

- **Bash / Shell scripts vs Bun + TypeScript for repo scripts:**
  - *Adopted:* TypeScript on Bun for automation scripts, wrapped behind mise tasks. Delivers type-safe, cross-platform scripting on developer machines, while strict API guidelines (`node:fs`, `node:path`) neutralize Bun 1.4 Windows shell interpreter bugs.
  - *Rejected:* Bash/sh scripts. Unreliable on Windows without WSL or MSYS2/Git Bash, causing constant cross-platform breakage.
- **Go-only repo scripts (`go run scripts/...`):**
  - *Rejected for repo tasks:* Writing small automation scripts, JSON/YAML data mangling, and test runners in Go is verbose and slow to iterate. Bun + TypeScript offers superior developer ergonomics for repo chores while keeping the Go core untouched.
- **Single-runner cross-compilation with `zig cc` for all release targets:**
  - *Deferred to local builds:* While `zig cc` works well for Linux and macOS cross-compilation, Windows CGO support under `zig cc` frequently encounters runtime CRT and ABI discrepancies with official tree-sitter C bindings. GoReleaser native runner matrix (split builds) guarantees 100% ABI compliance and zero cross-compilation headaches.
- **Node.js + `tsx` or Deno instead of Bun:**
  - *Rejected:* Adds unnecessary toolchain sprawl. Bun provides single-binary speed, instant TypeScript execution without configuration, and fast package installation.
- **Manual PR Reviews vs CI Quality Gates:**
  - *Rejected:* Human manual review is excluded by project design. CI carries the full verification burden; changes with any gate failure are rejected unconditionally.

## Consequences

- Consistent, deterministic environment across Linux, macOS, and Windows workstations and CI via `mise.toml`.
- Single-command entry point for developers and agents: `mise run check` and `mise run build`.
- Fast, typed repository scripting in TypeScript without contaminating the standalone Go product.
- Fully automated, secure, and reproducible release pipeline generating verified, signed single-binary releases for all major platforms.
- Absolute isolation of study subjects (`graphify`, `understand-anything`) with automated license compliance.

Source: [This repo's development workflow](../../../../.scratch/system-context-harness/issues/20-repo-dev-workflow.md)
