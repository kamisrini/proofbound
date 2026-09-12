#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q
git -C "$tmp" config user.email fixture@example.invalid
git -C "$tmp" config user.name fixture
mkdir -p "$tmp/notes"
printf 'state\n' >"$tmp/notes/state.md"; git -C "$tmp" add .; git -C "$tmp" commit -qm state
for n in 1 2 3; do printf '%s\n' "$n" >"$tmp/file"; git -C "$tmp" add file; git -C "$tmp" commit -qm "$n"; done
STATE_FRESHNESS_ROOT=$tmp STATE_FRESHNESS_COMMITS=3 scripts/state-freshness.sh
printf '4\n' >"$tmp/file"; git -C "$tmp" add file; git -C "$tmp" commit -qm 4
if STATE_FRESHNESS_ROOT=$tmp STATE_FRESHNESS_COMMITS=3 scripts/state-freshness.sh >/dev/null 2>&1; then echo 'stale state passed' >&2; exit 1; fi
printf 'refreshing\n' >>"$tmp/notes/state.md"
STATE_FRESHNESS_ROOT=$tmp STATE_FRESHNESS_COMMITS=3 scripts/state-freshness.sh
if STATE_FRESHNESS_ROOT=$tmp STATE_FRESHNESS_COMMITS=bad scripts/state-freshness.sh >/dev/null 2>&1; then echo 'bad limit passed' >&2; exit 1; fi
