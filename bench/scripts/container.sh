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
    docker run --rm --entrypoint sh "$(cat "out/$arm/image-digest.txt")" -c 'cat /pip-freeze.txt 2>/dev/null || true' > "out/$arm/resolved-deps.txt"
    ;;
  query)
    # Query phase: run the image by the digest recorded at build time, never by tag;
    # no network, read-only root, no credentials, and only the mounts phases.json declares.
    digest_file="out/$arm/image-digest.txt"
    [ -s "$digest_file" ] || { echo "build arm $arm first: $digest_file is missing" >&2; exit 2; }
    image_id=$(cat "$digest_file")
    case "$arm" in
      C) mounts=(-v "$PWD/out/corpus:/corpus:ro" -v "$PWD/out/C:/out") ;;
      D) mounts=(-v "$PWD/out/D:/out:ro") ;;
    esac
    exec docker run --rm -i --network none --read-only --cap-drop ALL --security-opt no-new-privileges \
      "${mounts[@]}" "$image_id" "${@:3}"
    ;;
  *) echo "phase must be build or query" >&2; exit 2 ;;
esac
