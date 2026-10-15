---
schema: task/1
tier: standard
status: verified
verified_at: "d6caa31"
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
| AC1a | `go test ./cmd/gira -run 'TestSCSBOMAndCosign\|TestSCGoReleaserConfig' -count=1` | pass | d6caa31 | ran on working tree, re-pass on merged main; tightened assertion to exact `checksum` |
| AC2a | `sh -n install.sh` | pass | d6caa31 | ran on working tree, re-pass on merged main; POSIX syntax |
| AC2b | simulated install flow (local tarball + checksums.txt) | pass | d6caa31 | ran on working tree; checksum match, extract and run ok; tampered tarball rejected |
| AC1b | `go test ./cmd/gira -run TestSCGoReleaserConfig -count=1` após padrão bundle | pass | d6caa31 | ran on working tree, re-pass on merged main; exige bundle `.sigstore.json`, rejeita flags v2 |
| AC2c | `sh -n install.sh` + `--help` após flag `--cosign` | pass | d6caa31 | ran on working tree, re-pass on merged main |
| AC1 | release retry on re-pushed `v0.1.0` | pass | d6caa31 | run `38819865583` success on `f5dad05`; `v0.1.0` publicado com 12 assets incl bundle |
| AC2 | live curl install against published `v0.1.0` assets | pass | d6caa31 | ran against live assets with files identical to merged main (see V1); gh path and python3 plus `--cosign` (`Verified OK`) both install a working `gira 0.1.0` |
| R1 | `node tools/validate/validate.js` | pass | d6caa31 | exit 0 on merged main |
| V1 | re-run on merged main + content equality with live-tested files | pass | d6caa31 | `sh -n`, SC pair and validator re-pass on merged main; `git diff f514c01 origin/main -- install.sh README.md` empty; no workflow or quarantine tokens; `ci.yml` untouched |

## Outcome

Delivered and verified at `d6caa31`: bundle-signing fix, `install.sh` (gh + python3 paths, sha256 fail-closed, `--cosign` keyless), README. Live `v0.1.0` install proven on both paths. Release backlog selado.
