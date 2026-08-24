#!/usr/bin/env bash
set -euo pipefail

repo=${SPEC_LINT_ROOT:-$(git rev-parse --show-toplevel)}
failed=0
while IFS= read -r -d '' spec; do
  while IFS= read -r line; do
    [[ $line == \|* ]] || continue
    IFS='|' read -r _ id _ citation _ <<< "$line"
    id=${id#"${id%%[![:space:]]*}"}
    id=${id%"${id##*[![:space:]]}"}
    citation=${citation#"${citation%%[![:space:]]*}"}
    citation=${citation%"${citation##*[![:space:]]}"}
    [[ $id =~ ^((G-)?INV-[0-9]+[a-z]?|F[0-9]+)$ ]] || continue
    [ "$citation" = '—' ] && continue
    if [[ ! $citation =~ ^[A-Za-z0-9_]+_test\.go::Test[A-Za-z0-9_]+$ ]]; then
      echo "$spec: invariant $id has malformed proving-test citation: $citation" >&2
      failed=1
    fi
  done < "$spec"
done < <(find "$repo/kernel" -name SPEC.md -type f -print0)
exit "$failed"
