#!/usr/bin/env bash
# Quarantine compliance check for the docs site: fail if the published site
# references the quarantined comparative-benchmark module or a benchmarked
# tool. The forbidden patterns live here, under website/ (unguarded), and not
# in .github/workflows/docs.yml: CI workflows are guarded files under ADR 0012
# rule 3, so listing the patterns there trips the bench quarantine guard.
set -euo pipefail
cd "$(dirname "$0")/.."
if grep -rEi "(^|[^a-z0-9_-])\.?/?bench([^a-z0-9_-]|$)|gira-bench|graphify|understand[-_ ]?anything|egonex-ai" . --exclude-dir=build --exclude-dir=node_modules --exclude-dir=.docusaurus --exclude=check-quarantine.sh; then
  echo "Quarantined comparative module reference found in docs site" >&2
  exit 1
fi
