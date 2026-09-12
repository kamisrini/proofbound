#!/usr/bin/env bash
set -euo pipefail

repo=${WRAP_VERIFY_ROOT:-$(git rev-parse --show-toplevel)}
[[ -z $(git -C "$repo" status --porcelain --untracked-files=all) ]] || { echo 'wrap-verify: worktree is not clean' >&2; exit 1; }
git -C "$repo" diff-tree --root --no-commit-id --name-only -r HEAD | rg -qx 'notes/state.md' || { echo 'wrap-verify: HEAD did not update notes/state.md' >&2; exit 1; }
