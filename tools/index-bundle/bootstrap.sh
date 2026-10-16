#!/usr/bin/env bash
# Bootstrap an offline-capable git cache for `gira index`.
#
# Where network allows, this script pre-populates the gira repository
# cache (default $XDG_CACHE_HOME/gira/repos or ~/.cache/gira/repos, or
# --cache-dir) at the pinned commits from the manifest, then runs
# `gira index --offline` and saves the bundle. Copy the cache directory
# plus the emitted macro-map.json to network-blocked machines.
#
# Usage:
#   tools/index-bundle/bootstrap.sh [--manifest gira.yaml]
#     [--cache-dir DIR] [--output macro-map.json] [--skip-index]
set -euo pipefail

MANIFEST="gira.yaml"
CACHE_DIR=""
OUTPUT="macro-map.json"
SKIP_INDEX="0"

usage() {
  sed -n '1,12p' "$0"
  echo "Options: --manifest PATH --cache-dir DIR --output PATH --skip-index --help"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --manifest) MANIFEST="$2"; shift 2 ;;
    --cache-dir) CACHE_DIR="$2"; shift 2 ;;
    --output) OUTPUT="$2"; shift 2 ;;
    --skip-index) SKIP_INDEX="1"; shift 1 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "unknown flag: $1 (see --help)" >&2; exit 1 ;;
  esac
done

CACHE_ARGS=()
if [ -n "$CACHE_DIR" ]; then
  CACHE_ARGS+=(--cache-dir "$CACHE_DIR")
fi

echo "1/2: populating cache from $MANIFEST (needs network once)..."
bin/gira index "$MANIFEST" "${CACHE_ARGS[@]}" --output "$OUTPUT"

if [ "$SKIP_INDEX" = "1" ]; then
  exit 0
fi

echo "2/2: verifying offline reproducibility..."
bin/gira index "$MANIFEST" "${CACHE_ARGS[@]}" --offline --output "$OUTPUT"
echo "offline index ok: $OUTPUT"
