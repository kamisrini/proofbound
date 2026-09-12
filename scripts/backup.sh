#!/usr/bin/env bash
set -euo pipefail

repo=${BACKUP_ROOT:-$(git rev-parse --show-toplevel)}
dest=${PROOFBOUND_BACKUP_DIR:-}
[[ -n $dest ]] || { echo 'backup: set PROOFBOUND_BACKUP_DIR to an explicit off-worktree directory' >&2; exit 1; }
repo_abs=$(cd "$repo" && pwd -P)
mkdir -p "$dest"
dest_abs=$(cd "$dest" && pwd -P)
case $dest_abs/ in "$repo_abs"/*) echo 'backup: destination must be outside the worktree' >&2; exit 1;; esac
name=$(basename "$repo_abs")
stamp=$(date -u +%Y%m%dT%H%M%SZ)
output=$dest_abs/$name-$stamp.bundle
tmp=$output.tmp
trap 'rm -f "$tmp"' EXIT
git -C "$repo_abs" bundle create "$tmp" --all
git bundle verify "$tmp" >/dev/null
mv "$tmp" "$output"
trap - EXIT
printf '%s\n' "$output"
