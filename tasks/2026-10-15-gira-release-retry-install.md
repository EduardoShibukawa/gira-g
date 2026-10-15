---
schema: task/1
tier: standard
status: completed
---

# Release retry v0.1.0 + install.sh via curl

## Routing and approvals

- tier: standard (surgical fix + additive install script; owner directed: "script igual os outros lugares fazem com um CURL", retry "Re-usa v0.1.0", install scope delegated: "pense nisso")
- decided: retry re-uses tag v0.1.0 (delete + re-push after fix lands on main; nothing was published) by owner (2026-10-15, "Re-usa v0.1.0")
- decided: install.sh primary via curl with sha256 verify; mise/ubi deferred to follow-up after first release with assets by agent (owner: "Nao sei pense nisso, pensei em podef azer por mise haha")

## Objective

Unblock the failed `v0.1.0` release and provide one-line machine install.

In scope:
- `.goreleaser.yaml` signs fix + regression test tightening;
- root `install.sh` (POSIX sh, curl, sha256 verify) + README section.

Out of scope:
- pushing/deleting git tags or publishing a live GitHub Release (owner commands after merge);
- cosign verify inside install.sh (follow-up: exact .sig/.pem artifact names unconfirmed until first green release);
- touching `.github/workflows/ci.yml`.

## Acceptance criteria

- AC1 = `goreleaser release` signing step passes on the `v*.*.*` retry (artifacts enum valid).
- AC2 = `curl .../install.sh | sh` installs a working `gira` with checksum fail-closed on tamper.

## Verification

| ID | Check | Result | Ref | Note |
|---|---|---|---|---|
| B1 | release job `38780553092` log (before fix) | pass | 6392407 | `signing artifacts` fails: `invalid list of artifacts to sign: checksums`; root cause `artifacts: checksums` (plural) vs GoReleaser enum `checksum` (singular, confirmed via `goreleaser jsonschema`) |
| B2 | release job `38807111633` log (bundle fix) | pass | b44ac84 | signs com enum ok, mas cosign 3.1.3 rejeita flags v2: `must specify --bundle with --new-bundle-format`; padrão oficial GoReleaser é `--bundle` com `signature` `.sigstore.json` |
| AC1a | `go test ./cmd/gira -run 'TestSCSBOMAndCosign\|TestSCGoReleaserConfig' -count=1` | pass | local | tightened assertion to exact `checksum`; `goreleaser check` still exit 0 |
| AC2a | `sh -n install.sh` | pass | local | POSIX syntax |
| AC2b | simulated install flow (local tarball + checksums.txt, asset name `gira_0.1.0_linux_amd64.tar.gz` as in failed release log) | pass | local | checksum match → extract → `gira --version` ok; tampered tarball rejected |
| AC1b | `go test ./cmd/gira -run TestSCGoReleaserConfig -count=1` após padrão bundle | pass | local | exige `signature` `${artifact}.sigstore.json`, `--bundle` e rejeita flags v2 removidas |
| AC2c | `sh -n install.sh` + `--help` após flag `--cosign` | pass | local | sintaxe ok; verificação live vira AC2 abaixo |
| AC1 | release retry on re-pushed `v0.1.0` | pass | run `38819865583` | preflight + release success no commit `f5dad05`; `v0.1.0` publicado com 12 assets incl bundle |
| AC2 | live curl install against published `v0.1.0` assets | pass | local | via gh: sha256 ok + `gira 0.1.0` roda; via python3 com `--cosign`: `Verified OK` + sha256 ok + roda |
| R1 | `node tools/validate/validate.js` | pass | local | exit 0 |
| V1 | diff read (fix + test + install.sh + README) | pass | local | no workflow/quarantine tokens; `ci.yml` untouched |

## Outcome

Delivered: bundle-signing fix, `install.sh` (gh + python3 paths, sha256 fail-closed, `--cosign` keyless), README. Verified live against `v0.1.0`. Pending: merge install-token PR; then flip to `verified`.
