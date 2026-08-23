#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp); trap 'rm -f "$tmp"' EXIT
GEN_INDEX_OUTPUT="$tmp" scripts/gen-index.sh
cmp -s "$tmp" docs/decisions/INDEX.md || { echo 'index stale; run make index' >&2; exit 1; }
