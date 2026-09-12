#!/usr/bin/env bash
set -euo pipefail
repo=${PRESCRIPTION_ROOT:-$(git rev-parse --show-toplevel)}
failed=0
while IFS= read -r -d '' spec; do
  dir=${spec%/*}
  while IFS= read -r sequence; do
    [[ $sequence == *'[retracted]'* ]] && continue
    mapfile -t flags < <(rg -o -- '--[A-Za-z0-9][A-Za-z0-9_=/*.-]*' <<<"$sequence" | sort -u)
    ((${#flags[@]} >= 2)) || continue
    for flag in "${flags[@]}"; do
      rg -q -F "\"$flag\"" "$dir" -g '*.go' -g '!*_test.go' || { echo "prescription-lint: $spec prescribes absent literal $flag" >&2; failed=1; }
    done
  done < <(perl -ne 'while (/`([^`]*--[^`]*--[^`]*)`/g) { print "$1\n" }' "$spec")
done < <(find "$repo/kernel/internal" -name SPEC.md -type f -print0)
exit "$failed"
