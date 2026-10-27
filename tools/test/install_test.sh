#!/usr/bin/env bash
# Teste: `mise run install` gera binário versionado em $GIRA_INSTALL_DIR.
set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
cd "$root"
GIRA_INSTALL_DIR="$tmp" mise run install >/dev/null 2>&1 || { echo "FAIL: mise run install"; exit 1; }
[ -x "$tmp/gira" ] || { echo "FAIL: $tmp/gira ausente"; exit 1; }
out="$("$tmp/gira" --version)"
want="$(git describe --tags)"
case "$out" in *0.1.0-dev*) echo "FAIL: versão dev: $out"; exit 1;; esac
case "$out" in *"$want"*) ;; *) echo "FAIL: '$out' sem '$want'"; exit 1;; esac
grep -q 'mise run install' README.md || { echo "FAIL: README sem 'mise run install'"; exit 1; }
echo "PASS: $out"
