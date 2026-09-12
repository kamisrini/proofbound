#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

bash scripts/p6-task2-close.sh --check
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
cp docs/verification/p6-connector-reality.md "$tmp"
sed -i '/| C4-009 | sessions |/d' "$tmp"
if P6_TASK2_ARTIFACT="$tmp" bash scripts/p6-task2-close.sh --check >/dev/null 2>&1; then
  echo 'missing connector classification was accepted' >&2
  exit 1
fi
