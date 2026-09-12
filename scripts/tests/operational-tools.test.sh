#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
repo=$tmp/repo; backups=$tmp/backups
git init -q "$repo"; git -C "$repo" config user.email fixture@example.invalid; git -C "$repo" config user.name fixture
mkdir -p "$repo/notes"; printf 'state\n' >"$repo/notes/state.md"; git -C "$repo" add .; git -C "$repo" commit -qm state
STATE_ROOT=$repo scripts/gen-state.sh | rg -q '^state_distance=0$'
META_TAX_ROOT=$repo scripts/meta-tax.sh | rg -q 'total=1'
bundle=$(BACKUP_ROOT=$repo PROOFBOUND_BACKUP_DIR=$backups scripts/backup.sh)
git bundle verify "$bundle" >/dev/null
if BACKUP_ROOT=$repo PROOFBOUND_BACKUP_DIR=$repo/Backups scripts/backup.sh >/dev/null 2>&1; then echo 'in-worktree backup passed' >&2; exit 1; fi
WRAP_VERIFY_ROOT=$repo scripts/wrap-verify.sh
printf 'dirty\n' >"$repo/dirty"
if WRAP_VERIFY_ROOT=$repo scripts/wrap-verify.sh >/dev/null 2>&1; then echo 'dirty wrap passed' >&2; exit 1; fi
