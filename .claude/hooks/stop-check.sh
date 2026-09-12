#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
if [[ -n $(git -C "$root" status --porcelain --untracked-files=all) ]]; then echo 'stop-check: worktree is dirty; commit durable work before ending' >&2; fi
if ! STATE_FRESHNESS_ROOT=$root "$root/scripts/state-freshness.sh"; then echo 'stop-check: notes/state.md is stale' >&2; fi
exit 0
