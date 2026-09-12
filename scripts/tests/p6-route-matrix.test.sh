#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

bash scripts/p6-route-matrix.sh --check
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
cp docs/verification/p6-event-route-matrix.md "$tmp"
sed -i '/| C5-052 |/d' "$tmp"
if P6_ROUTE_ARTIFACT="$tmp" bash scripts/p6-route-matrix.sh --check >/dev/null 2>&1; then
  echo 'incomplete route matrix was accepted' >&2
  exit 1
fi
