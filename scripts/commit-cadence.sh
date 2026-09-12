#!/usr/bin/env bash
set -euo pipefail

repo=${COMMIT_CADENCE_ROOT:-$(git rev-parse --show-toplevel)}
minutes=${COMMIT_CADENCE_MINUTES:-90}
now=${COMMIT_CADENCE_NOW_EPOCH:-$(date +%s)}
[[ $minutes =~ ^[1-9][0-9]*$ ]] || { echo 'commit-cadence: COMMIT_CADENCE_MINUTES must be a positive integer' >&2; exit 1; }
[[ $now =~ ^[0-9]+$ ]] || { echo 'commit-cadence: invalid current epoch' >&2; exit 1; }
git -C "$repo" rev-parse --verify HEAD >/dev/null 2>&1 || { echo 'commit-cadence: HEAD is required' >&2; exit 1; }
[[ -z $(git -C "$repo" status --porcelain --untracked-files=all) ]] && exit 0
head_time=$(git -C "$repo" show -s --format=%ct HEAD)
age=$((now - head_time))
((age >= 0)) || { echo 'commit-cadence: HEAD timestamp is in the future' >&2; exit 1; }
limit=$((minutes * 60))
((age <= limit)) || { echo "commit-cadence: dirty worktree outlived ${minutes}m commit window" >&2; exit 1; }
