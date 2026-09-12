#!/usr/bin/env bash
set -euo pipefail
repo=${INVARIANT_LINT_ROOT:-$(git rev-parse --show-toplevel)}
failed=0
while IFS= read -r -d '' spec; do
  dir=${spec%/*}
  while IFS= read -r citation; do
    file=${citation%%::*}; test=${citation##*::}
    [[ -f $dir/$file ]] && rg -q "^func[[:space:]]+$test\\(" "$dir/$file" || { echo "invariant-lint: ${spec#"$repo"/} citation does not resolve: $citation" >&2; failed=1; }
  done < <(rg -o '[A-Za-z0-9_]+_test\.go::Test[A-Za-z0-9_]+' "$spec" | sort -u)
done < <(find "$repo/kernel/internal" -name SPEC.md -type f -print0)
exit "$failed"
