#!/usr/bin/env bash
set -euo pipefail
repo=${LESSON_ROOT:-$(git rev-parse --show-toplevel)}
declare -A count=() response=() retired=()
while IFS= read -r line; do
  [[ $line =~ LESSON:[[:space:]]+class=([a-z0-9][a-z0-9-]*) ]] || continue
  class=${BASH_REMATCH[1]}; count[$class]=$(( ${count[$class]:-0} + 1 ))
  [[ $line =~ response=([^[:space:]]+) ]] && response[$class]=${BASH_REMATCH[1]}
  [[ $line =~ retired=([^[:space:]]+) ]] && retired[$class]=${BASH_REMATCH[1]}
done < <(rg --no-heading 'LESSON:' "$repo/notes/journal" -g '*.md' 2>/dev/null || true)
failed=0
for class in "${!count[@]}"; do
  ((${count[$class]} >= 2)) || continue
  if [[ -n ${retired[$class]:-} ]]; then continue; fi
  path=${response[$class]:-}
  [[ -n $path && -e $repo/$path ]] || { echo "lesson-recurrence: class $class recurred without compiled response" >&2; failed=1; }
done
exit "$failed"
