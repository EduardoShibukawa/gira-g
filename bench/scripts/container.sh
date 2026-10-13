#!/usr/bin/env bash
# Builds or runs one arm's container phase. Owner-run only; never called from CI.
#   scripts/container.sh build <C|D>   build the image from the pinned recipe and record its digest
#   scripts/container.sh query <C|D>   run the query phase with networking disabled
set -euo pipefail
cd "$(dirname "$0")/.."
phase=${1:?phase}; arm=${2:?arm}
case "$arm" in
  C) tool=graphify ;;
  D) tool=understand-anything ;;
  *) echo "arm must be C or D" >&2; exit 2 ;;
esac
mkdir -p "out/$arm"
image="gira-bench-$tool"
case "$phase" in
  build)
    docker build --pull=false -t "$image" "containers/$tool"
    docker image inspect --format '{{.Id}}' "$image" > "out/$arm/image-digest.txt"
    ;;
  query)
    # Query phase: no network, read-only root, no credentials, corpus copy only.
    exec docker run --rm -i --network none --read-only --cap-drop ALL --security-opt no-new-privileges \
      -v "$PWD/out/corpus:/corpus:ro" -v "$PWD/out/$arm:/out" "$image" "${@:3}"
    ;;
  *) echo "phase must be build or query" >&2; exit 2 ;;
esac
