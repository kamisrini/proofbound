#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# Baseline: the committed artifact is fresh and there are no pending routed-dir changes.
bash scripts/p6-route-matrix.sh --check

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
cp docs/verification/p6-event-route-matrix.md "$tmp"
sed -i '/| C5-052 |/d' "$tmp"
if P6_ROUTE_ARTIFACT="$tmp" bash scripts/p6-route-matrix.sh --check >/dev/null 2>&1; then
  echo 'incomplete route matrix was accepted' >&2
  exit 1
fi

# Staged-aware guard (DEF-ROUTE-MATRIX-PRECOMMIT): when a routed kernel dir has pending changes,
# the git-log-derived pin will be stale after commit, so --check MUST fail closed even though the
# committed artifact still matches the committed history. The P6_ROUTE_PENDING_OVERRIDE seam
# simulates `git status --porcelain` output without mutating the working tree.
if P6_ROUTE_PENDING_OVERRIDE=' M kernel/internal/gates/gates.go' bash scripts/p6-route-matrix.sh --check >/dev/null 2>&1; then
  echo 'pre-commit route-matrix blind spot: pending routed-dir change was accepted' >&2
  exit 1
fi

# Neutral control: an explicitly-empty override (no pending routed changes) still passes.
if ! P6_ROUTE_PENDING_OVERRIDE='' bash scripts/p6-route-matrix.sh --check >/dev/null 2>&1; then
  echo 'route-matrix --check wrongly failed with no pending routed changes' >&2
  exit 1
fi
