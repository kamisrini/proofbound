#!/usr/bin/env bash
set -euo pipefail

repo=${SPEC_LINT_ROOT:-$(git rev-parse --show-toplevel)}
lock=${INVARIANTS_LOCK_FILE:-$repo/docs/invariants.lock}
[ -s "$lock" ] || { echo 'invariants lock is missing or empty' >&2; exit 1; }
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
SPEC_LINT_ROOT="$repo" INVARIANTS_LOCK_OUTPUT="$tmp" "$repo/scripts/gen-invariants-lock.sh"
cmp -s "$tmp" "$lock" || { echo 'invariant numbering drift; run make invariants-lock only for a deliberate append/retitle' >&2; exit 1; }
