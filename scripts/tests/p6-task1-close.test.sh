#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/repo"
cp -a .claude README.md Makefile docs gates kernel scripts "$tmp/repo/"

P6_TASK1_ROOT="$tmp/repo" bash "$tmp/repo/scripts/p6-task1-close.sh" --write
P6_TASK1_ROOT="$tmp/repo" bash "$tmp/repo/scripts/p6-task1-close.sh" --check

sed -i '/make gates-canary/d' "$tmp/repo/README.md"
if P6_TASK1_ROOT="$tmp/repo" bash "$tmp/repo/scripts/p6-task1-close.sh" --check >/dev/null 2>&1; then
  echo 'missing documented target was accepted' >&2
  exit 1
fi
cp README.md "$tmp/repo/README.md"

sed -i '/gates\/intent-reference-integrity.yaml/d' "$tmp/repo/docs/gates.md"
if P6_TASK1_ROOT="$tmp/repo" bash "$tmp/repo/scripts/p6-task1-close.sh" --check >/dev/null 2>&1; then
  echo 'unregistered gate path was accepted' >&2
  exit 1
fi
cp docs/gates.md "$tmp/repo/docs/gates.md"

sed -i 's/2026-10-16/2026-01-01/' "$tmp/repo/docs/gates.md"
if P6_TASK1_ROOT="$tmp/repo" bash "$tmp/repo/scripts/p6-task1-close.sh" --check >/dev/null 2>&1; then
  echo 'expired advisory was accepted' >&2
  exit 1
fi
