#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
cp docs/decisions/INDEX.md "$tmp/index"
printf '\n' >> docs/decisions/INDEX.md
if scripts/index-check.sh; then exit 1; fi
mv "$tmp/index" docs/decisions/INDEX.md
scripts/index-check.sh
