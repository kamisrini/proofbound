#!/usr/bin/env bash
set -euo pipefail

repo=${STATE_FRESHNESS_ROOT:-$(git rev-parse --show-toplevel)}
limit=${STATE_FRESHNESS_COMMITS:-3}
[[ $limit =~ ^[0-9]+$ ]] || { echo 'state-freshness: STATE_FRESHNESS_COMMITS must be a nonnegative integer' >&2; exit 1; }
git -C "$repo" rev-parse --verify HEAD >/dev/null 2>&1 || { echo 'state-freshness: HEAD is required' >&2; exit 1; }
if ! git -C "$repo" diff --quiet -- notes/state.md || ! git -C "$repo" diff --cached --quiet -- notes/state.md; then
  exit 0
fi
last=$(git -C "$repo" log -1 --format=%H -- notes/state.md)
[[ -n $last ]] || { echo 'state-freshness: notes/state.md has no committed history' >&2; exit 1; }
distance=$(git -C "$repo" rev-list --count "$last..HEAD")
((distance <= limit)) || { echo "state-freshness: notes/state.md is $distance commits behind HEAD (limit $limit)" >&2; exit 1; }
