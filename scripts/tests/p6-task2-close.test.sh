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

# Staged-aware guard (DEF-ROUTE-MATRIX-PRECOMMIT, peer artifact): a pending change to a pinned path
# must fail --check closed, since the git-log pin will be stale after commit. Seam-driven so it needs
# no working-tree mutation.
if P6_TASK2_PENDING_OVERRIDE=' M kernel/internal/connector/reviews/reviews.go' bash scripts/p6-task2-close.sh --check >/dev/null 2>&1; then
  echo 'p6-task2-close pre-commit blind spot: pending pinned-path change was accepted' >&2
  exit 1
fi
if ! P6_TASK2_PENDING_OVERRIDE='' bash scripts/p6-task2-close.sh --check >/dev/null 2>&1; then
  echo 'p6-task2-close --check wrongly failed with no pending pinned-path changes' >&2
  exit 1
fi
