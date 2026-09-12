#!/usr/bin/env bash
set -euo pipefail

repo=${STATE_ROOT:-$(git rev-parse --show-toplevel)}
head=$(git -C "$repo" rev-parse HEAD)
branch=$(git -C "$repo" symbolic-ref --quiet --short HEAD || echo detached)
last=$(git -C "$repo" log -1 --format=%H -- notes/state.md)
distance=unknown
[[ -n $last ]] && distance=$(git -C "$repo" rev-list --count "$last..HEAD")
dirty=no
[[ -n $(git -C "$repo" status --porcelain --untracked-files=all) ]] && dirty=yes
printf 'head=%s\nbranch=%s\ndirty=%s\nstate_commit=%s\nstate_distance=%s\n' "$head" "$branch" "$dirty" "${last:-none}" "$distance"
