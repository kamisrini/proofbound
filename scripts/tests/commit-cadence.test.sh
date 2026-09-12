#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q
git -C "$tmp" config user.email fixture@example.invalid
git -C "$tmp" config user.name fixture
printf 'base\n' >"$tmp/file"
git -C "$tmp" add file
GIT_AUTHOR_DATE='2000-01-01T00:00:00Z' GIT_COMMITTER_DATE='2000-01-01T00:00:00Z' git -C "$tmp" commit -qm base
COMMIT_CADENCE_ROOT=$tmp COMMIT_CADENCE_NOW_EPOCH=946684800 scripts/commit-cadence.sh
printf 'dirty\n' >>"$tmp/file"
COMMIT_CADENCE_ROOT=$tmp COMMIT_CADENCE_NOW_EPOCH=946684860 scripts/commit-cadence.sh
if COMMIT_CADENCE_ROOT=$tmp COMMIT_CADENCE_NOW_EPOCH=946690201 scripts/commit-cadence.sh >/dev/null 2>&1; then echo 'stale dirty work passed' >&2; exit 1; fi
if COMMIT_CADENCE_ROOT=$tmp COMMIT_CADENCE_MINUTES=bad scripts/commit-cadence.sh >/dev/null 2>&1; then echo 'bad limit passed' >&2; exit 1; fi
