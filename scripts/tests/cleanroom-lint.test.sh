#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q; git -C "$tmp" config user.email x@y; git -C "$tmp" config user.name x
printf 'safe\n' >"$tmp/a"; git -C "$tmp" add a; git -C "$tmp" commit -qm x
CLEANROOM_ROOT=$tmp scripts/cleanroom-lint.sh 2>&1 | rg -q INERT
printf 'forbidden-value\n' >"$tmp/patterns"; CLEANROOM_ROOT=$tmp PROOFBOUND_CLEANROOM_PATTERNS=$tmp/patterns scripts/cleanroom-lint.sh
printf 'forbidden-value\n' >>"$tmp/a"; git -C "$tmp" add a
if CLEANROOM_ROOT=$tmp PROOFBOUND_CLEANROOM_PATTERNS=$tmp/patterns scripts/cleanroom-lint.sh >/dev/null 2>&1; then exit 1; fi
